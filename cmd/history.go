package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"sort"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	huhSpinner "github.com/charmbracelet/huh/spinner"
	"github.com/creack/pty"
	"github.com/getsavvyinc/savvy-cli/client"
	"github.com/getsavvyinc/savvy-cli/cmd/component"
	"github.com/getsavvyinc/savvy-cli/display"
	"github.com/getsavvyinc/savvy-cli/server"
	"github.com/getsavvyinc/savvy-cli/shell"
	"github.com/muesli/cancelreader"
	"github.com/spf13/cobra"
)

// historyCmd represents the history command
var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Create a runbook from your shell history",
	Long: `Create a runbook from  a selection of the last 100 commands in your shell history.
  Savvy can expand all aliases used in your shell history without running the commands.`,
	Hidden: true,
	Run:    recordHistory,
}

func init() {
	recordCmd.AddCommand(historyCmd)
}

func recordHistory(cmd *cobra.Command, _ []string) {
	ctx := cmd.Context()
	logger := loggerFromCtx(ctx).With("command", "history")
	cl, err := client.New()
	if err != nil && errors.Is(err, client.ErrInvalidClient) {
		display.Error(errors.New("You must be logged in to record a runbook. Please run `savvy login`"))
		os.Exit(1)
	}

	commandProcessedChan := make(chan bool, 1)
	defer close(commandProcessedChan)

	hook := func(cmd string) {
		logger.Debug("command recorded", "command", cmd)
		commandProcessedChan <- true
	}

	ctx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx()

	ss, err := server.NewUnixSocketServerWithDefaultPath(server.WithCommandRecordedHook(hook))
	if errors.Is(err, server.ErrAbortRecording) {
		display.Info("Recording aborted")
		return
	}

	if err != nil {
		display.FatalErrWithSupportCTA(err)
		return
	}
	defer ss.Close()

	go func() {
		ss.ListenAndServe()
		// kill b/g shell if we exit early
		cancelCtx()
		os.Exit(1)
	}()

	if err != nil {
		display.FatalErrWithSupportCTA(err)
	}

	sh := shell.New("/tmp/savvy-socket")
	lines, err := sh.TailHistory(ctx)
	if err != nil {
		display.FatalErrWithSupportCTA(err)
	}

	selectedHistory := allowUserToSelectCommands(lines)
	if len(selectedHistory) == 0 {
		display.Error(errors.New("No commands were selected"))
		return
	}

	var commands []*server.RecordedCommand
	if err := huhSpinner.New().Title("Processing selected commands").Action(func() {
		var err error

		commands, err = expandHistory(ctx, logger, ss, sh, selectedHistory, commandProcessedChan)
		if err != nil {
			display.FatalErrWithSupportCTA(err)
		}

		if len(commands) == 0 {
			display.Error(errors.New("No commands were recorded"))
			return
		}
	}).Run(); err != nil {
		logger.Debug("failed to run spinner", "error", err.Error())
	}

	gctx, cancel := context.WithCancel(ctx)
	gm := component.NewGenerateRunbookModel(commands, cl)
	p := tea.NewProgram(gm, tea.WithOutput(programOutput), tea.WithContext(gctx))
	if _, err := p.Run(); err != nil {
		err = fmt.Errorf("failed to generate runbook: %w", err)
		display.ErrorWithSupportCTA(err)
		os.Exit(1)
	}

	// ensure the bubble tea program is finished before we start the next one
	logger.Debug("wait for bubbletea program", "component", "generate runbook", "status", "running")
	cancel()
	p.Wait()
	logger.Debug("wait for bubbletea program", "component", "generate runbook", "status", "finished")

	runbook := <-gm.RunbookCh()
	m, err := newDisplayCommandsModel(runbook)
	if err != nil {
		display.FatalErrWithSupportCTA(err)
	}

	p = tea.NewProgram(m, tea.WithOutput(programOutput), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		logger.Debug("failed to display runbook", "error", err.Error())
		display.Info("View and edit your runbook online at: " + runbook.URL)
		os.Exit(1)
	}
	if runbook.URL != "" {
		display.Success("View and edit your runbook online at: " + runbook.URL)
	}
}

type selectableCommand struct {
	Key     int
	Command string
}

func allowUserToSelectCommands(history []string) []string {
	var options []huh.Option[selectableCommand]
	var selectedOptions []selectableCommand
	for i, cmd := range history {
		options = append(options, huh.NewOption(fmt.Sprintf("%d %s", i+1, cmd), selectableCommand{Key: i, Command: cmd}))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[selectableCommand]().
				Title("Savvy History").
				Description("Press x to include/exclude commands in your Runbook. Selected Commands will NOT be executed.").
				Value(&selectedOptions).
				Height(33).
				Options(options...),
		),
	)

	if err := form.Run(); err != nil {
		log.Fatal(err)
	}

	sort.Slice(selectedOptions, func(i, j int) bool {
		// older commands should be at the top
		return selectedOptions[i].Key > selectedOptions[j].Key
	})

	var commands []string
	for _, c := range selectedOptions {
		commands = append(commands, c.Command)
	}
	return commands
}

func expandHistory(ctx context.Context,
	logger *slog.Logger,
	srv *server.UnixSocketServer,
	sh shell.Shell,
	rawCommands []string,
	cmdProcessedCh <-chan bool,
) ([]*server.RecordedCommand, error) {
	logger.Debug("expanding history", "commands", rawCommands)

	ctx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx()

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
	var wg sync.WaitGroup
	wg.Add(1)
	cancelReader, err := cancelreader.NewReader(ptmx)
	if err != nil {
		logger.Debug("failed to create cancel reader", "error", err.Error())
		return nil, err
	}
	defer cancelReader.Close()

	go func() {
		defer wg.Done()
		io.Copy(io.Discard, cancelReader)
	}()

	for i, cmd := range rawCommands {
		if _, err := fmt.Fprintln(ptmx, cmd); err != nil {
			return nil, err
		}
		// Wait for the command to be processed by the server.
		select {
		case <-cmdProcessedCh:
			logger.Debug("command processed", "command", cmd, "index", i, "cmd", "history")
		case <-time.After(5 * time.Second):
			logger.Debug("timeout waiting for command to be processed", "command", cmd, "index", i)
		}
	}
	ptmx.Write([]byte{4}) // End Of Transmission (EOT) == Ctrl-D

	logger.Debug("waiting for wg.Wait()")
	// cancelReader.Cancel() will close the cancelReader and cause io.Copy to return, which will unblock wg.Wait()
	ok := cancelReader.Cancel()
	logger.Debug("cancelReader.Cancel() returned", "ok", ok)
	wg.Wait()
	logger.Debug("wg.Wait() finished")
	logger.Debug("canceling context for psuedo terminal and its associated command")
	cancelCtx()
	logger.Debug("waitng for c.Wait()")
	c.Wait()
	logger.Debug("c.Wait() finished")
	return srv.Commands(), nil
}
