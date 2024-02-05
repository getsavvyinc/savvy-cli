/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/getsavvyinc/savvy-cli/cmd/setup"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:     "init",
	Example: "savvy init zsh",

	Short: "Output shell setup",
	Long:  `Output shell setup`,

	ValidArgs: setup.SupportedShells,
	Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),

	Run: func(cmd *cobra.Command, args []string) {
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.AddCommand(setup.ZshCmd)
	initCmd.AddCommand(setup.BashCmd)
	initCmd.AddCommand(setup.DashCmd)
}
