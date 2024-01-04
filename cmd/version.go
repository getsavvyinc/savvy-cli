package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// version is the version of the CLI
// version is set via ldflags at build time
var version string

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Shows the savvy cli version",
	Long:  "Shows the savvy cli version",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println("version:", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
