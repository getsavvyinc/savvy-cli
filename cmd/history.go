package cmd

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/charmbracelet/huh"
	"github.com/creack/pty"
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
	sh := shell.New("/tmp/savvy-socket")
	lines, err := sh.TailHistory(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	selectedHistory := allowUserToSelectCommands(lines)
	expandedHistory, err := expandHistory(sh, selectedHistory)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(expandedHistory)
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
