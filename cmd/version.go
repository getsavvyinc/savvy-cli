package cmd

import (
	"fmt"

	"github.com/getsavvyinc/savvy-cli/config"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Shows the savvy cli version",
	Long:  "Shows the savvy cli version",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println("version:", config.Version())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
