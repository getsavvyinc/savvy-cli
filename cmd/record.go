package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/creack/pty"
	"github.com/getsavvyinc/savvy-cli/client"
	"github.com/getsavvyinc/savvy-cli/cmd/component"
	"github.com/getsavvyinc/savvy-cli/savvy_errors"
	"github.com/getsavvyinc/savvy-cli/shell"

	"github.com/getsavvyinc/savvy-cli/server"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// recordCmd represents the record command
var recordCmd = &cobra.Command{
	Use:   "record",
	Short: "Record records each terminal command and helps you create a runbook",
	Long: `Record creates a sub shell that records each terminal command and helps you create a runbook.

  Type 'exit' to exit the sub shell and view the runbook.`,
	Run: runRecordCmd,
}

func runRecordCmd(cmd *cobra.Command, args []string) {
	cl, err := client.New()
	if err != nil && errors.Is(err, client.ErrInvalidClient) {
		savvy_errors.Display(errors.New("You must be logged in to record a runbook. Please run `savvy login`"))
		os.Exit(1)
	}

	commands, err := startRecording()
	if err != nil {
		savvy_errors.DisplayWithSupportCTA(err)
		os.Exit(1)
	}

	gm := component.NewGenerateRunbookModel(commands, cl)
	p := tea.NewProgram(gm)
	if _, err := p.Run(); err != nil {
		// TODO: fail gracefully. Provider users either a link to view the runbook or a list of their saved commands
		fmt.Printf("could not run program: %s\n", err)
		os.Exit(1)
	}

	runbook := <-gm.RunbookCh()
	m := newDisplayCommandsModel(runbook)
	p = tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		// TODO: fail gracefully and provide users a link to view the runbook
		fmt.Printf("could not run program: %s\n", err)
		os.Exit(1)
	}
}

func startRecording() ([]string, error) {
	// TODO: Make this unique for each invokation
	socketPath := "/tmp/savvy-socket"
	ss, err := server.NewUnixSocketServer(socketPath)
	if err != nil {
		return nil, err
	}
	// TODO: kill this goroutine when the shell exits
	go ss.ListenAndServe()
	defer ss.Close()
	// Create arbitrary command.
	sh := shell.New(socketPath)
	ctx := context.Background()
	c, err := sh.Spawn(ctx)
	if err != nil {
		err := fmt.Errorf("failed to start recording: %w", err)
		return nil, err
	}

	// Start the command with a pty.
	ptmx, err := pty.Start(c)
	if err != nil {
		return nil, err
	}
	// Make sure to close the pty at the end.
	defer func() { _ = ptmx.Close() }() // Best effort.

	// Handle pty size.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			if err := pty.InheritSize(os.Stdin, ptmx); err != nil {
				log.Printf("error resizing pty: %s", err)
			}
		}
	}()
	ch <- syscall.SIGWINCH                        // Initial resize.
	defer func() { signal.Stop(ch); close(ch) }() // Cleanup signals when done.

	// Set stdin in raw mode.
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return nil, fmt.Errorf("failed to set stdin to raw mode: %w", err)
	}
	defer func() {
		term.Restore(int(os.Stdin.Fd()), oldState)
	}() // Best effort.

	// Copy stdin to the pty and the pty to stdout.
	go func() { _, _ = io.Copy(ptmx, os.Stdin) }()
	// TODO: wrap this in a go-routine and wait for an exit command or signal
	go func() { _, _ = io.Copy(os.Stdout, ptmx) }()

	if err := c.Wait(); err != nil {
		// TODO: inspect the error and determine if we exited due to ctrl-c or exit or something else.
	}
	// close the shell
	if err := term.Restore(int(os.Stdin.Fd()), oldState); err != nil {
		panic(err)
	}

	return ss.Commands(), nil
}

type displayCommands struct {
	l component.ListModel
}

func toListItems(steps []component.RunbookStep) []component.ListItem {
	var items []component.ListItem
	for _, step := range steps {
		items = append(items, component.ListItem{
			Command:         step.Command,
			DescriptionText: step.Description,
		})
	}
	return items
}

func newDisplayCommandsModel(runbook *component.Runbook) displayCommands {
	if runbook == nil {
		panic("runbook cannot be nil")
	}

	l := component.NewListModel(toListItems(runbook.Steps), runbook.Title, runbook.URL)
	return displayCommands{l: l}
}

func (dc displayCommands) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	if err := dc.l.Init(); err != nil {
		fmt.Printf("Error initializing list: %v", err)
		return nil
	}
	return nil
}

func (dc displayCommands) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return dc.l.Update(msg)
}

func (dc displayCommands) View() string {
	return dc.l.View()
}

func init() {
	rootCmd.AddCommand(recordCmd)
}
