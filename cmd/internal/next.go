package internal

import (
	"fmt"
	"os"

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

		state, err := cl.CurrentState()
		if err != nil {
			display.ErrorWithSupportCTA(err)
			os.Exit(1)
		}

		if state.CommandWithSetParams() != executedCommand {
			fmt.Printf("%d", state.Index)
			return
		}

		if err := cl.NextCommand(); err != nil {
			display.ErrorWithSupportCTA(err)
			os.Exit(1)
		}

		updated, err := cl.CurrentState()
		if err != nil {
			display.ErrorWithSupportCTA(err)
			os.Exit(1)
		}
		fmt.Printf("%d", updated.Index)
	},
}

var executedCommand string

func init() {
	InternalCmd.AddCommand(nextCmd)

	nextCmd.Flags().StringVarP(&executedCommand, "cmd", "c", "", "previously executed command")
}
