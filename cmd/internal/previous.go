package internal

import (
	"context"
	"fmt"
	"os"

	"github.com/getsavvyinc/savvy-cli/display"
	"github.com/getsavvyinc/savvy-cli/server/run"
	"github.com/spf13/cobra"
)

// previousCommand represents the next command
var previousCmd = &cobra.Command{
	Use:    "previous",
	Hidden: true,
	Short:  "Update runbook state to the previous step",
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

		if forcePrevious {
			updated, err := previousCommand(ctx, cl)
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

func previousCommand(ctx context.Context, cl run.Client) (*run.State, error) {
	if err := cl.PreviousCommand(); err != nil {
		return nil, err
	}

	updatedState, err := cl.CurrentState()
	if err != nil {
		return nil, err
	}
	return updatedState, nil
}

var forcePrevious bool

func init() {
	InternalCmd.AddCommand(previousCmd)
	previousCmd.Flags().BoolVarP(&forcePrevious, "force", "f", false, "force previous command regardless of current state")
}
