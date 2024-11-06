package cmd

import (
	"github.com/getsavvyinc/savvy-cli/client"
	"github.com/getsavvyinc/savvy-cli/storage"
	"github.com/spf13/cobra"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Create a local copy of all your Savvy Artifacts",
	Long: `
  Create a local copy of all your Savvy Artifacts.

  This command will download all your artifacts from the Savvy API and store them in a local directory.
  You can access the artifacts in the local directory even when you are offline using savvy run --local.
  `,
	Run: syncRunbooks,
}

func syncRunbooks(cmd *cobra.Command, args []string) {
	store := map[string]*client.Runbook{}
	ctx := cmd.Context()
	logger := loggerFromCtx(ctx).With("command", "sync")

	var cl client.RunbookClient
	var err error
	cl, err = client.GetLoggedInClient()
	if err != nil {
		logger.Error(err.Error())
		return
	}

	runbookInfo, err := cl.Runbooks(ctx)
	if err != nil {
		logger.Error("failed to fetch runbooks", "error", err)
		return
	}

	for _, rb := range runbookInfo {
		rb, err := cl.RunbookByID(ctx, rb.RunbookID)
		if err != nil {
			logger.Error("failed to fetch runbook", "runbook_id", rb.RunbookID, "error", err)
			return
		}

		store[rb.RunbookID] = rb
	}

	if err := storage.Write(store); err != nil {
		logger.Error("failed to write runbooks to local storage", "error", err)
		return
	}
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
