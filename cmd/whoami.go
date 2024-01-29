package cmd

import (
	"fmt"

	"github.com/getsavvyinc/savvy-cli/client"
	"github.com/spf13/cobra"
)

// whoamiCmd represents the whoami command
var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Shows information about the current user",
	Long:  "Shows information about the current user",
	Run: func(cmd *cobra.Command, _ []string) {
		cl, err := client.New()
		if err != nil {
			fmt.Println(err)
			return
		}
		whoami, err := cl.WhoAmI(cmd.Context())
		if err != nil {
			cmd.PrintErrln(err)
		}
		cmd.Println(whoami)
	},
}

func init() {
	rootCmd.AddCommand(whoamiCmd)
}
