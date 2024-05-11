/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// askCmd represents the ask command
var askCmd = &cobra.Command{
	Use:   "ask",
	Short: "Ask Savvy a question and it will generate a command",
	Example: `
  savvy ask "how do I deploy a k8s daemonset?"
  savvy ask "how do I find the process id listening on a port?"
  savvy ask "how do I quit vim?"
  `,
	Long: `
  Ask Savvy a question and it will generate a command for you.
  `,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("ask called")
	},
}

func init() {
	rootCmd.AddCommand(askCmd)
}
