package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/creack/pty"
	"github.com/getsavvyinc/savvy-cli/client"
	"github.com/getsavvyinc/savvy-cli/cmd/component"
	"github.com/getsavvyinc/savvy-cli/display"
	"github.com/getsavvyinc/savvy-cli/server"
	"github.com/getsavvyinc/savvy-cli/shell"
	"github.com/spf13/cobra"
)

// historyCmd represents the history command
var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Create a runbook from your shell history",
	Long: `Create a runbook from  a selection of the last 100 commands in your shell history.
  Savvy can expand all aliases used in your shell history without running the commands.`,
	Run: recordHistory,
}

func init() {
	recordCmd.AddCommand(historyCmd)
}

func recordHistory(_ *cobra.Command, _ []string) {
	cl, err := client.New()
	if err != nil && errors.Is(err, client.ErrInvalidClient) {
		display.Error(errors.New("You must be logged in to record a runbook. Please run `savvy login`"))
		os.Exit(1)
	}

	sh := shell.New("/tmp/savvy-socket")
	lines, err := sh.TailHistory(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	selectedHistory := allowUserToSelectCommands(lines)
	commands, err := expandHistory(sh, selectedHistory)
	if err != nil {
		display.ErrorWithSupportCTA(err)
		os.Exit(1)
	}
	if len(commands) == 0 {
		display.Error(errors.New("No commands were recorded"))
		return
	}

	ctx := context.Background()
	gctx, cancel := context.WithCancel(ctx)
	gm := component.NewGenerateRunbookModel(commands, cl)
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

func allowUserToSelectCommands(history []string) (selectedHistory []string) {
	var options []huh.Option[string]
	for i, cmd := range history {
		options = append(options, huh.NewOption(fmt.Sprintf("%d %s", i, cmd), cmd))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Savvy History").
				Description("Press x to include/exclude commands in your Runbook").
				Value(&selectedHistory).
				Height(33).
				Options(options...),
		),
	)

	if err := form.Run(); err != nil {
		log.Fatal(err)
	}
	return
}

func expandHistory(sh shell.Shell, rawCommands []string) ([]string, error) {
	socketPath := "/tmp/savvy-socket"
	ss, err := server.NewUnixSocketServer(socketPath)
	if err != nil {
		return nil, err
	}
	go ss.ListenAndServe()
	defer func() {
		ss.Close()
	}()

	ctx, cancelCtx := context.WithCancel(context.Background())
	defer func() {
		cancelCtx()
	}()

	c, err := sh.SpawnHistoryExpander(ctx)
	if err != nil {
		err := fmt.Errorf("failed to start history recording: %w", err)
		return nil, err
	}

	ptmx, err := pty.Start(c)
	if err != nil {
		return nil, err
	}
	// Make sure to close the pty at the end.
	defer ptmx.Close()

	// io.Copy blocks till ptmx is closed.
	go func() {
		io.Copy(io.Discard, ptmx)
	}()

	for _, cmd := range rawCommands {
		if _, err := fmt.Fprintln(ptmx, cmd); err != nil {
			return nil, err
		}
	}
	ptmx.Write([]byte{4}) // EOT

	// time.Sleep(1 * time.Second)
	// pw.Close()
	c.Wait()
	cancelCtx()
	// cancelReader.Cancel()
	// println("waiting for wg")
	// wg.Wait()
	return ss.Commands(), nil
}

// nullWriter implements the io.Writer interface and discards all data written to it.
type nullWriter struct{}

// Write discards the data written to the NullWriter.
func (nw nullWriter) Write(p []byte) (int, error) {
	return len(p), nil
}
