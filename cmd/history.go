package cmd

import (
	"context"
	"fmt"
	"log"

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
	for _, line := range lines {
		fmt.Println(line)
	}
}
