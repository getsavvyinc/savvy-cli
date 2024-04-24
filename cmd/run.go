package cmd

import (
	"os"

	"github.com/getsavvyinc/savvy-cli/client"
	"github.com/getsavvyinc/savvy-cli/display"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:     "run",
	Short:   "Run takes a runbook ID and runs it",
	Example: "savvy run $rb-runbookID",
	Long:    `Run takes a runbook ID and runs it. Run automatically steps though the runbook for you, there's no need manually copy paste individual commands.`,
	Run:     savvyRun,
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
		cl = client.Guest()
	}

	if len(args) == 0 {
		display.ErrorMsg("missing: runbookID")
		os.Exit(1)
	}

	_ = cl
}
