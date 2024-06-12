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

		curr := cl.CurrentCommand()
		fmt.Printf("%s", curr)
	},
}

func init() {
	InternalCmd.AddCommand(currentCmd)
}
