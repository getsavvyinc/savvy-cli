package cmd

import (
	"fmt"

	"github.com/getsavvyinc/savvy-cli/client"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:     "run [runbookID]",
	Short:   "Run takes a runbook ID and runs it",
	Example: "savvy run rb-runbookID",
	Long: `
  Run takes a runbook ID and runs it.

  Run automatically steps though the runbook for you, there's no need manually copy paste individual commands.
  `,
	Run: savvyRun,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("missing: runbookID\n")
		}
		return nil
	},
	// Args: cobra.ExactArgs(1),
}

func init() {
	rootCmd.AddCommand(runCmd)
}

func savvyRun(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()
	logger := loggerFromCtx(ctx).With("command", "run")

	var cl client.Client
	cl, err := client.New()
	if err != nil {
		logger.Debug("error creating client", "error", err, "message", "falling back to guest client")
		cl = client.NewGuest()
	}

	runbookID := args[0]

	rb, err := cl.RunbookByID(ctx, runbookID)
	if err != nil {
		logger.Error("failed to fetch runbook", "runbook_id", runbookID, "error", err)
		return
	}

	fmt.Println("Running runbook", rb.Title)
}
