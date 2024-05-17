package cmd

import (
	"os"
	"path"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/getsavvyinc/savvy-cli/client"
	"github.com/getsavvyinc/savvy-cli/cmd/component/viewport"
	"github.com/getsavvyinc/savvy-cli/display"
	"github.com/spf13/cobra"
)

// explainCmd represents the explain command
var explainCmd = &cobra.Command{
	Use:   "explain",
	Short: "Explain explains shell commands and functions",
	Args:  cobra.MinimumNArgs(1),
	Example: `
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

		// be defensive: users can pass questions as one string or multiple strings
		code := strings.Join(args[:], " ")

		// get info about the os from os pkg: mac/darwin, linux, windows
		goos := runtime.GOOS
		if goos == "darwin" {
			goos = "macos, darwin, osx"
		}

		fileData, err := fileData(filePath)
		if err != nil {
			display.Error(err)
			os.Exit(1)
		}

		ci := client.CodeInfo{
			Code: code,
			Tags: map[string]string{
				"os": goos,
			},
			FileData: fileData,
			FileName: path.Base(filePath),
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
