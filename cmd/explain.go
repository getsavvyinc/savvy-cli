package cmd

import (
	"os"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/getsavvyinc/savvy-cli/client"
	"github.com/getsavvyinc/savvy-cli/cmd/component/viewport"
	"github.com/getsavvyinc/savvy-cli/display"
	"github.com/spf13/cobra"
)

// explainCmd represents the explain command
var explainCmd = &cobra.Command{
	Use:   "explain",
	Short: "Explain explains shell commands and functions",
	Example: `
  savvy explain # interactive mode
  savvy explain 'openssl x509 -text -in ./ca.crt -noout | grep --color=auto -C 2 "Valid"'
  savvy explain cat "file.txt | sort | uniq -c | sort -nr | head -n 10"
  `,
	Long: `
  Explain helps you understand shell commands and functions.

  If a file path is provider, Savvy will use the contents of the file to generate a command.
  `,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		logger := loggerFromCtx(ctx).With("command", "explain")

		var cl client.Client
		var err error

		cl, err = client.New()
		if err != nil {
			logger.Debug("error creating client", "error", err, "message", "falling back to guest client")
			cl = client.NewGuest()
		}

		var code string
		if len(args) == 0 {
			// interactive mode
			text := huh.NewText().Title("Enter the shell command savvy should explain").Value(&code)
			form := huh.NewForm(huh.NewGroup(text))
			if err := form.Run(); err != nil {
				display.ErrorWithSupportCTA(err)
				os.Exit(1)
			}
		}

		// be defensive: users can pass questions as one string or multiple strings
		if len(args) > 0 && len(code) == 0 {
			code = strings.Join(args[:], " ")
		}

		// get info about the os from os pkg: mac/darwin, linux, windows
		goos := runtime.GOOS
		if goos == "darwin" {
			goos = "macos, darwin, osx"
		}

		ci := client.CodeInfo{
			Code: code,
			Tags: map[string]string{
				"os": goos,
			},
		}

		explainCh, err := cl.Explain(ctx, ci)
		if err != nil {
			display.Error(err)
			os.Exit(1)
		}

		m := viewport.NewModel(explainCh)
		p := tea.NewProgram(
			m,
			tea.WithMouseCellMotion(),
		)

		go func() {
			for explanation := range explainCh {
				p.Send(viewport.ContentMsg{Content: explanation})
			}
		}()

		if _, err := p.Run(); err != nil {
			display.Error(err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(explainCmd)
}
