package internal

import (
	"github.com/spf13/cobra"
)

// internalCmd represents the internal command
var InternalCmd = &cobra.Command{
	Use:    "internal",
	Hidden: true,
	Short:  "Internal commands not meant to be used by end users.",
	Run: func(cmd *cobra.Command, args []string) {
	},
}
