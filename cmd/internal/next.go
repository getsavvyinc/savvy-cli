package internal

import (
	"context"
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

		if forceNext || state.CommandWithSetParams() == executedCommand {
			updated, err := nextCommand(ctx, cl)
			if err != nil {
				display.ErrorWithSupportCTA(err)
				os.Exit(1)
			}
			fmt.Printf("%d", updated.Index)
			return
		}
		fmt.Printf("%d", state.Index)
	},
}

func nextCommand(ctx context.Context, cl run.Client) (*run.State, error) {
	if err := cl.NextCommand(); err != nil {
		return nil, err
	}

	updatedState, err := cl.CurrentState()
	if err != nil {
		return nil, err
	}
	return updatedState, nil
}

var executedCommand string
var forceNext bool

func init() {
	InternalCmd.AddCommand(nextCmd)

	nextCmd.Flags().StringVarP(&executedCommand, "cmd", "c", "", "previously executed command")
	nextCmd.Flags().BoolVarP(&forceNext, "force", "f", false, "force next command regardless of current state")

}
