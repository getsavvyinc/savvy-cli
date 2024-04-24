package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:     "run",
	Short:   "Run takes a runbook ID and runs it",
	Example: "savvy run $r-runbookID",
	Long:    `Run takes a runbook ID and runs it. Run automatically steps though the runbook for you, there's no need manually copy paste individual commands.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("run called")
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
