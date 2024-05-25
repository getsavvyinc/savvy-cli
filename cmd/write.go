package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// writeCmd represents the write command
var writeCmd = &cobra.Command{
	Use:   "write",
	Short: "Write writes a file to your file system",
	Example: `
  savvy write --step-id=stp-id
  `,
	Long: `
  Write writes a file to your file system.

  Write is used while running a runbook to write important data to your filesystem.
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("write called")
	},
}

func init() {
	rootCmd.AddCommand(writeCmd)
}
