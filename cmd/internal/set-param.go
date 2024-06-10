package internal

import (
	"fmt"

	"github.com/spf13/cobra"
)

// subcommandCmd represents the subcommand command
var subcommandCmd = &cobra.Command{
	Use:   "set-param",
	Short: "Prompt the user to set one ore parameters for their runbook",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("subcommand called")
	},
}

func init() {
	InternalCmd.AddCommand(subcommandCmd)
}
