package cmd

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

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
	Use:     "run [runbookID]",
	Short:   "Run takes a runbook ID and runs it",
	Example: "savvy run rb-runbookID",
	Long: `
  Run takes a runbook ID and runs it.

  Run automatically steps though the runbook for you, there's no need manually copy paste individual commands.
  `,
	Run: savvyRun,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("missing: runbookID\n")
		}
		return nil
	},
	// Args: cobra.ExactArgs(1),
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

	runbookID := args[0]
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

	fmt.Println("next: run the runbook", rb.Title)
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
