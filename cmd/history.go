package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"sort"
	"sync"
	"syscall"
	"time"

	"github.com/charmbracelet/huh"
	huhSpinner "github.com/charmbracelet/huh/spinner"
	"github.com/creack/pty"
	"github.com/getsavvyinc/savvy-cli/cmd/browser"
	"github.com/getsavvyinc/savvy-cli/display"
	"github.com/getsavvyinc/savvy-cli/export"
	"github.com/getsavvyinc/savvy-cli/extension"
	"github.com/getsavvyinc/savvy-cli/redact"
	"github.com/getsavvyinc/savvy-cli/server"
	"github.com/getsavvyinc/savvy-cli/shell"
	"github.com/getsavvyinc/savvy-cli/theme"
	"github.com/muesli/cancelreader"
	"github.com/spf13/cobra"
)

// historyCmd represents the history command
var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Create an artifact from your shell history",
	Long: `Create an artifact from  a selection of the last 100 commands in your shell history.
  Savvy can expand all aliases used in your shell history without running the commands.`,
	Hidden: false,
	Run:    recordHistory,
}

func init() {
	recordCmd.AddCommand(historyCmd)
}

func recordHistory(cmd *cobra.Command, _ []string) {
	ctx := cmd.Context()
	logger := loggerFromCtx(ctx).With("cmd", "history")

	historyCmds, err := selectAndExpandHistory(ctx, logger)
	if err != nil {
		display.FatalErrWithSupportCTA(err)
		return
	}

	links, err := getLinks(ctx)
	if err != nil {
		fmt.Errorf("failed to get links from Savvy's Chrome Extension: %w", err)
		display.ErrorWithSupportCTA(err)
	}

	if len(historyCmds) == 0 && len(links) == 0 {
		return
	}

	display.Info("Creating artifact...")

	exporter := export.NewExporter(historyCmds, links)
	if err := exporter.Export(ctx); err != nil {
		display.ErrorWithSupportCTA(err)
		os.Exit(1)
	}
}

type selectableCommand struct {
	Key     int
	Command string
}

func getLinks(ctx context.Context) ([]extension.HistoryItem, error) {
	var collectedLinks []extension.HistoryItem

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Create a processor function that collects links
	processor := func(items []extension.HistoryItem) error {
		collectedLinks = append(collectedLinks, items...)
		// cancel the context to stop the extension server and exit the spinner
		cancel()
		return nil
	}

	// Create and start the extension server
	server := extension.New(processor)
	if err := server.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start extension server: %w", err)
	}
	defer server.Close()

	// Present user with form to choose between extension and skip
	var choice string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Share important links from your browser history").
				Description("Select the links you want to share with the team").
				Options(
					huh.NewOption("Open Chrome Extension", "open"),
					huh.NewOption("Skip", "skip"),
				).
				Value(&choice),
		),
	)

	if err := form.WithTheme(theme.New()).Run(); err != nil {
		return nil, err
	}

	if choice == "open" {
		// Open the chrome extension using the browser package
		browser.OpenExtensionSidePanel()

		// Wait for a short duration to collect links
		spinner := huhSpinner.New().
			Title("Waiting for browser history...").
			Action(func() {
				server.Wait()
			})

		if err := spinner.Run(); err != nil {
			return nil, err
		}
	}

	return collectedLinks, nil
}

func allowUserToSelectCommands(logger *slog.Logger, history []string) []string {
	var options []huh.Option[selectableCommand]
	var selectedOptions []selectableCommand
	for i, cmd := range history {
		options = append(options, huh.NewOption(fmt.Sprintf("%d %s", i+1, cmd), selectableCommand{Key: i, Command: cmd}))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[selectableCommand]().
				Title("Select Commands to Create an Artifact").
				Description("Press x to add/remove commands").
				Value(&selectedOptions).
				Height(33).
				Options(options...),
		),
	)

	if err := form.WithTheme(theme.New()).Run(); err != nil {
		logger.Debug("failed to run form", "error", err.Error())
		return nil
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
			logger.Debug("failed to write command to pseudo terminal", "error", err.Error())
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
	logger.Debug("canceling context for pseudo terminal and its associated command")
	cancelCtx()
	logger.Debug("waitng for c.Wait()")
	c.Wait()
	logger.Debug("c.Wait() finished")
	return srv.Commands(), nil
}

func selectAndExpandHistory(ctx context.Context, logger *slog.Logger) ([]*server.RecordedCommand, error) {
	ctx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx()

	sh := shell.New("/tmp/savvy-socket")
	lines, err := sh.TailHistory(ctx)
	if err != nil {
		return nil, err
	}

	selectedHistory := allowUserToSelectCommands(logger, lines)
	if len(selectedHistory) == 0 {
		return nil, nil
	}

	commandProcessedChan := make(chan bool, 1)
	defer close(commandProcessedChan)

	hook := func(cmd string) {
		logger.Debug("command recorded", "command", cmd)
		commandProcessedChan <- true
	}

	ss, err := server.NewUnixSocketServerWithDefaultPath(server.WithCommandRecordedHook(hook))
	if errors.Is(err, server.ErrAbortRecording) {
		display.Info("Recording aborted")
		return nil, nil
	}

	if err != nil {
		return nil, err
	}
	defer ss.Close()

	go func() {
		ss.ListenAndServe()
		// kill b/g shell if we exit early
		cancelCtx()
	}()

	intSigChan := make(chan os.Signal, 1)
	signal.Notify(intSigChan, syscall.SIGINT)
	logger.Debug("listening for interrupt signal")
	go func() {
		<-intSigChan
		ss.Close()
	}()

	var commands []*server.RecordedCommand
	if err := huhSpinner.New().Title("Processing selected commands").Action(func() {
		var err error

		commands, err = expandHistory(ctx, logger, ss, sh, selectedHistory, commandProcessedChan)
		if err != nil {
			display.FatalErrWithSupportCTA(err)
			return
		}

		if len(commands) == 0 {
			display.Error(errors.New("No commands were recorded"))
			return
		}
	}).Run(); err != nil {
		logger.Debug("failed to run spinner", "error", err.Error())
	}

	redacted, err := redact.Commands(commands)
	if err != nil {
		logger.Debug("failed to redact commands", "error", err.Error())
		return nil, err
	}

	return redacted, nil
}
