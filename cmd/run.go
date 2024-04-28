package cmd

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/charmbracelet/huh"
	huhSpinner "github.com/charmbracelet/huh/spinner"
	"github.com/creack/pty"
	"github.com/getsavvyinc/savvy-cli/client"
	"github.com/getsavvyinc/savvy-cli/display"
	"github.com/getsavvyinc/savvy-cli/shell"
	"github.com/muesli/cancelreader"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run [runbookID]",
	Short: "Run takes a runbook ID and runs it",
	Example: `
  # Select and run from a list of runbooks you have access to
  savvy run

  # Run a specific runbook
  savvy run rb-runbookID
  `,
	Long: `
  Run allows users to select any runbook and run it.

  Without any arguments, savvy run will display a list of runbooks you have access to and allow you to select one to run.

  If you provide a runbook ID, savvy run will run that specific runbook.

  Run automatically steps though the runbook for you, there's no need manually copy paste individual commands.
  `,
	Run:  savvyRun,
	Args: cobra.MaximumNArgs(1),
}

func init() {
	rootCmd.AddCommand(runCmd)
}

func savvyRun(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	logger := loggerFromCtx(ctx).With("command", "run")

	var cl client.Client
	cl, err := client.New()
	if err != nil {
		logger.Debug("error creating client", "error", err, "message", "falling back to guest client")
		cl = client.NewGuest()
	}

	var runbookID string

	if len(args) == 0 {
		runbookID, err = allowUserToSelectRunbook(ctx, logger, cl)
		if err != nil {
			display.ErrorWithSupportCTA(err)
			os.Exit(1)
		}
	} else {
		runbookID = args[0]
	}

	rb, err := fetchRunbook(ctx, cl, runbookID)
	if err != nil {
		logger.Error("failed to fetch runbook", "runbook_id", runbookID, "error", err)
		return
	}

	if err := runRunbook(ctx, rb); err != nil {
		display.ErrorWithSupportCTA(
			fmt.Errorf("failed to run runbook %s: %w", rb.Title, err),
		)
		return
	}
}

func runRunbook(ctx context.Context, runbook *client.Runbook) error {
	sh := shell.New("/tmp/savvy-socket")
	ctx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx()

	c, err := sh.SpawnRunbookRunner(ctx, runbook)
	if err != nil {
		err := fmt.Errorf("run: failed to spawn shell %w", err)
		return err
	}

	// Start the command with a pty.
	ptmx, err := pty.Start(c)
	if err != nil {
		return err
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
		return fmt.Errorf("failed to set stdin to raw mode: %w", err)
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
	return nil
}

func fetchRunbook(ctx context.Context, cl client.Client, runbookID string) (*client.Runbook, error) {
	var rb *client.Runbook
	var err error
	if serr := huhSpinner.New().Title("Fetching runbook").Action(func() {
		rb, err = cl.RunbookByID(ctx, runbookID)
		if err != nil {
			err = fmt.Errorf("failed to fetch runbook %s: %w", runbookID, err)
			return
		}
	}).Run(); serr != nil {
		return nil, err
	}
	return rb, err
}

type selectableRunbook struct {
	Key       int    `json:"key"`
	RunbookID string `json:"runbook_id"`
}

func allowUserToSelectRunbook(ctx context.Context, logger *slog.Logger, cl client.Client) (string, error) {
	l := logger.With("func", "allowsUserToSelectRunbook")
	runbooks, err := cl.Runbooks(ctx)
	if err != nil {
		l.Debug("failed to fetch runbooks", "error", err)
		return "", err
	}

	var options []huh.Option[selectableRunbook]
	var selectedRunbook selectableRunbook

	for i, rb := range runbooks {
		options = append(options, huh.NewOption(
			fmt.Sprintf("%d. %s", i+1, rb.Title),
			selectableRunbook{Key: i, RunbookID: rb.RunbookID},
		))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[selectableRunbook]().
				Title("Select a runbook").
				Options(options...).
				Description("Press x to select the runbook you want to run").
				Height(33).
				Value(&selectedRunbook),
		),
	)

	if err := form.Run(); err != nil {
		logger.Debug("failed to run form", "error", err)
		return "", err
	}
	return selectedRunbook.RunbookID, nil
}
