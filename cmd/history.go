/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// historyCmd represents the history command
var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Create a runbook from your shell history",
	Long: `Create a runbook from  a selection of the last 100 commands in your shell history.
  Savvy can expand all aliases used in your shell history without running the commands.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("history called")
	},
}

func init() {
	recordCmd.AddCommand(historyCmd)
}
