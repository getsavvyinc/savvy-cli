package cmd

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"

	"github.com/getsavvyinc/savvy-cli/client"
	"github.com/getsavvyinc/savvy-cli/display"
	"github.com/spf13/cobra"
)

var askCmd = &cobra.Command{
	Use:   "ask",
	Short: "Ask Savvy a question and it will generate a command",
	Args:  cobra.MinimumNArgs(1),
	Example: `
  savvy ask "how do I deploy a k8s daemonset?"
  savvy ask "how do I parse a x509 cert"
  savvy ask "how do I find the process id listening on a port?"
  savvy ask "how do I quit vim?"
  `,
	Long: `
  Ask Savvy a question and it will generate a command for you.
  `,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		logger := loggerFromCtx(ctx).With("command", "ask")

		var cl client.Client
		var err error

		cl, err = client.New()
		if err != nil {
			logger.Debug("error creating client", "error", err, "message", "falling back to guest client")
			cl = client.NewGuest()
		}

		// be defensive: users can pass questions as one string or multiple strings
		question := strings.Join(args[:], " ")

		// get info about the os from os pkg: mac/darwin, linux, windows
		os := runtime.GOOS
		if os == "darwin" {
			os = "macos, darwin, osx"
		}

		qi := client.QuestionInfo{
			Question: question,
			Tags: map[string]string{
				"os": os,
			},
		}

		answer, err := cl.Ask(ctx, qi)
		if err != nil {
			display.ErrorWithSupportCTA(fmt.Errorf("error asking savvy: %w", err))
			return
		}

		bs, err := json.Marshal(answer)
		if err != nil {
			display.ErrorWithSupportCTA(fmt.Errorf("error encoding answer: %w", err))
			return
		}
		fmt.Println(string(bs))
	},
}

func init() {
	rootCmd.AddCommand(askCmd)
}
