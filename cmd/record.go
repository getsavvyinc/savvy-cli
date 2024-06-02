package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/creack/pty"
	"github.com/getsavvyinc/savvy-cli/client"
	"github.com/getsavvyinc/savvy-cli/cmd/component"
	"github.com/getsavvyinc/savvy-cli/cmd/component/list"
	"github.com/getsavvyinc/savvy-cli/display"
	"github.com/getsavvyinc/savvy-cli/shell"
	"github.com/muesli/cancelreader"
	"github.com/muesli/termenv"

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
	PreRun: func(_ *cobra.Command, _ []string) {
		checker := shell.NewSetupChecker()
		if err := checker.CheckSetup(); err != nil {
			display.Error(err)
			os.Exit(1)
		}
	},
	Run: runRecordCmd,
}

var programOutput = termenv.NewOutput(os.Stdout, termenv.WithColorCache(true))

func runRecordCmd(cmd *cobra.Command, _ []string) {
	cl, err := client.New()
	if err != nil && errors.Is(err, client.ErrInvalidClient) {
		display.Error(errors.New("You must be logged in to record a runbook. Please run `savvy login`"))
		os.Exit(1)
	}
	ctx := cmd.Context()

	recordedCommands, err := startRecording(ctx)
	if errors.Is(err, server.ErrAbortRecording) {
		display.Info("Recording aborted")
		return
	}

	if err != nil {
		display.ErrorWithSupportCTA(err)
		os.Exit(1)
	}

	if len(recordedCommands) == 0 {
		display.Error(errors.New("No commands were recorded"))
		return
	}

	gctx, cancel := context.WithCancel(ctx)
	gm := component.NewGenerateRunbookModel(recordedCommands, cl)
	p := tea.NewProgram(gm, tea.WithOutput(programOutput), tea.WithContext(gctx))
	if _, err := p.Run(); err != nil {
		err = fmt.Errorf("failed to generate runbook: %w", err)
		display.ErrorWithSupportCTA(err)
		os.Exit(1)
	}

	// ensure the bubble tea program is finished before we start the next one
	cancel()
	p.Wait()

	runbook := <-gm.RunbookCh()
	m, err := newDisplayCommandsModel(runbook)
	if err != nil {
		display.ErrorWithSupportCTA(err)
		os.Exit(1)
	}

	p = tea.NewProgram(m, tea.WithOutput(programOutput), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		// TODO: fail gracefully and provide users a link to view the runbook
		display.ErrorWithSupportCTA(fmt.Errorf("could not display runbook: %w", err))
		os.Exit(1)
	}
	if runbook.URL != "" {
		display.Success("View and edit your runbook online at: " + runbook.URL)
	}
}

func startRecording(ctx context.Context) ([]*server.RecordedCommand, error) {
	// TODO: Make this unique for each invokation
	ss, err := server.NewUnixSocketServerWithDefaultPath(server.WithIgnoreErrors(ignoreErrors))
	if err != nil {
		return nil, fmt.Errorf("failed to start recording: %w", err)
	}

	ctx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx()

	go func() {
		ss.ListenAndServe()
		// if the server shut down, cancel the shell context
		cancelCtx()
	}()
	defer ss.Close()

	// Create arbitrary command.
	sh := shell.New(ss.SocketPath())

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
	defer ptmx.Close()

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

	// Restore the terminal to its original state when we're done.
	defer func() {
		if err := term.Restore(int(os.Stdin.Fd()), oldState); err != nil {
			// intentionally display the error and continue without exiting
			display.Error(err)
		}
	}()

	// Create a cancelable reader
	// This is used to cancel the reader when the user types 'exit' or 'ctrl+d' to exit the subshell
	// Without this, the io.Copy blocks until the _next_ read that conflicts with bubbletea attempting to read from os.Stdin later on.
	cancelReader, err := cancelreader.NewReader(os.Stdin)
	if err != nil {
		display.Error(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		io.Copy(ptmx, cancelReader)
	}()

	// io.Copy blocks till ptmx is closed.
	io.Copy(os.Stdout, ptmx)

	// cleanup
	//// cancel ctx and wait for the underlying shell command to finish
	cancelCtx()
	c.Wait()

	//// cancel the cancelReader and wait for it's go routine to finish
	cancelReader.Cancel()
	wg.Wait()

	return ss.Commands(), nil
}

type displayCommands struct {
	l list.Model
}

func toItems(steps []component.RunbookStep) []list.Item {
	var items []list.Item
	for _, step := range steps {
		items = append(items, list.Item{
			Command:         step.Command,
			DescriptionText: step.Description,
		})
	}
	return items
}

func newDisplayCommandsModel(runbook *component.Runbook) (*displayCommands, error) {
	if runbook == nil {
		return nil, errors.New("runbook is empty")
	}

	listItems := toItems(runbook.Steps)
	l := list.NewModel(listItems, runbook.Title, runbook.URL)
	return &displayCommands{l: l}, nil
}

func (dc *displayCommands) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	dc.l.Init()
	return nil
}

func (dc *displayCommands) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return dc.l.Update(msg)
}

func (dc *displayCommands) View() string {
	return dc.l.View()
}

var ignoreErrors bool

func init() {
	rootCmd.AddCommand(recordCmd)
	// add a boolean flag
	recordCmd.Flags().BoolVar(&ignoreErrors, "ignore-errors", false, "Ignore commands that return an error when recording commands")
}
