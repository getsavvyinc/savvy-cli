package cmd

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/getsavvyinc/savvy-cli/client"
	"github.com/getsavvyinc/savvy-cli/cmd/component"
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
		goos := runtime.GOOS
		if goos == "darwin" {
			goos = "macos, darwin, osx"
		}

		qi := client.QuestionInfo{
			Question: question,
			Tags: map[string]string{
				"os": goos,
			},
		}

		runbook, err := cl.Ask(ctx, qi)
		if err != nil {
			display.ErrorWithSupportCTA(fmt.Errorf("error asking savvy: %w", err))
			return
		}

		rb := component.NewRunbook(&client.GeneratedRunbook{
			Runbook: *runbook,
		})

		m, err := newDisplayCommandsModel(rb)
		if err != nil {
			display.ErrorWithSupportCTA(err)
			os.Exit(1)
		}

		p := tea.NewProgram(m, tea.WithOutput(programOutput), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			// TODO: fail gracefully and provide users a link to view the runbook
			display.ErrorWithSupportCTA(fmt.Errorf("could not display runbook: %w", err))
			os.Exit(1)
		}
		if rb.URL != "" {
			display.Success("View and edit your runbook online at: " + rb.URL)
		}
		return
	},
}

func init() {
	rootCmd.AddCommand(askCmd)
}
