package internal

import (
	"fmt"

	"github.com/getsavvyinc/savvy-cli/display"
	"github.com/getsavvyinc/savvy-cli/server/run"
	"github.com/spf13/cobra"
)

// nextCmd represents the next command
var nextCmd = &cobra.Command{
	Use:    "next",
	Hidden: true,
	Short:  "Update runbook state to next step",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		cl, err := run.NewDefaultClient(ctx)
		if err != nil {
			display.ErrorWithSupportCTA(err)
			return
		}

		curr := cl.CurrentCommand()
		if curr != executedCommand {
			return
		}

		idx, err := cl.NextCommand()
		if err != nil {
			display.ErrorWithSupportCTA(err)
			return
		}

		fmt.Printf("%d", idx)
	},
}

var executedCommand string

func init() {
	InternalCmd.AddCommand(nextCmd)

	nextCmd.Flags().StringVarP(&executedCommand, "cmd", "c", "", "previously executed command")
}
