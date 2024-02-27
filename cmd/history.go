package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/charmbracelet/huh"
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
	sh := shell.New("/tmp/savvy.sock")
	lines, err := sh.TailHistory(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	selectedHistory := presentHistory(lines)
	fmt.Println(selectedHistory)
}

func presentHistory(history []string) (selectedHistory []string) {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Savvy History").
				Description("Press x to include/exclude commands in your Runbook").
				Value(&selectedHistory).
				Height(33).
				Options(huh.NewOptions(history...)...),
		),
	)

	if err := form.Run(); err != nil {
		log.Fatal(err)
	}
	return
}
