package cmd

import (
	"github.com/getsavvyinc/savvy-cli/client"
	"github.com/spf13/cobra"
)

// whoamiCmd represents the whoami command
var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Shows information about the current user",
	Long:  "Shows information about the current user",
	RunE: func(cmd *cobra.Command, args []string) error {
		cl, err := client.New()
		if err != nil {
			return err
		}
		whoami, err := cl.WhoAmI(cmd.Context())
		if err != nil {
			return err
		}
		cmd.Println(whoami)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(whoamiCmd)
}
