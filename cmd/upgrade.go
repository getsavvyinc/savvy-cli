package cmd

import (
	"context"
	"os"

	"github.com/getsavvyinc/savvy-cli/config"
	"github.com/getsavvyinc/savvy-cli/display"
	"github.com/getsavvyinc/upgrade-cli"
	"github.com/getsavvyinc/upgrade-cli/release/asset"
	"github.com/spf13/cobra"
)

const owner = "getsavvyinc"
const repo = "savvy-cli"

// upgradeCmd represents the upgrade command
var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "upgrade savvy to the latest version",
	Long:  `upgrade savvy to the latest version`,
	Run: func(cmd *cobra.Command, args []string) {
		executablePath, err := os.Executable()
		if err != nil {
			display.Error(err)
			os.Exit(1)
		}
		version := config.Version()

		assetDownloader := asset.NewAssetDownloader(executablePath, asset.WithLookupArchFallback(map[string]string{
			"amd64": "x86_64",
			"386":   "i386",
		}))
		upgrader := upgrade.NewUpgrader(owner, repo, executablePath, upgrade.WithAssetDownloader(assetDownloader))

		if ok, err := upgrader.IsNewVersionAvailable(context.Background(), version); err != nil {
			display.Error(err)
			return
		} else if !ok {
			display.Info("Savvy is already up to date")
			return
		}

		display.Info("Upgrading savvy...")
		if err := upgrader.Upgrade(context.Background(), version); err != nil {
			display.Error(err)
			os.Exit(1)
		} else {
			display.Success("Savvy has been upgraded to the latest version")
		}
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
