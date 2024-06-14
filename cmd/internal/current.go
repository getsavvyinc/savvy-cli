package internal

import (
	"fmt"

	"github.com/getsavvyinc/savvy-cli/display"
	"github.com/getsavvyinc/savvy-cli/server/run"
	"github.com/spf13/cobra"
)

// currentCmd represents the current command
var currentCmd = &cobra.Command{
	Use:   "current",
	Short: "Get the command to run",
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
			return
		}

		fmt.Printf("%s", state.CommandWithSetParams())
	},
}

func init() {
	InternalCmd.AddCommand(currentCmd)
}
