package cmd

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	huhSpinner "github.com/charmbracelet/huh/spinner"
	"github.com/getsavvyinc/savvy-cli/client"
	"github.com/getsavvyinc/savvy-cli/cmd/component"
	"github.com/getsavvyinc/savvy-cli/cmd/component/list"
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

		var runbook *client.Runbook
		if err := huhSpinner.New().Title("Savvy is generating an answer for you").Action(func() {
			var err error

			runbook, err = cl.Ask(ctx, qi)
			if err != nil {
				display.FatalErrWithSupportCTA(err)
				return
			}

			if len(runbook.Steps) == 0 {
				err := errors.New("No commands were generated. Please try again")
				display.FatalErrWithSupportCTA(err)
				return
			}
		}).Run(); err != nil {
			logger.Debug("error asking savvy", "error", err.Error())
			display.FatalErrWithSupportCTA(err)
			os.Exit(1)
		}

		rb := component.NewRunbook(&client.GeneratedRunbook{
			Runbook: *runbook,
		})

		m, err := newAskCommandsModel(rb)
		if err != nil {
			display.ErrorWithSupportCTA(err)
			os.Exit(1)
		}

		p := tea.NewProgram(m, tea.WithOutput(programOutput), tea.WithAltScreen())
		result, err := p.Run()
		if err != nil {
			// TODO: fail gracefully and provide users a link to view the runbook
			display.ErrorWithSupportCTA(fmt.Errorf("could not display runbook: %w", err))
			os.Exit(1)
		}

		if m, ok := result.(*askCommands); ok {
			selectedCommand := m.l.SelectedCommand()
			if selectedCommand == "" {
				return
			}
			if err := clipboard.WriteAll(selectedCommand); err != nil {
				display.Info(selectedCommand)
				return
			}
			display.Info(fmt.Sprintf("Copied to clipboard: %s", selectedCommand))
		}
		return
	},
}

type askCommands struct {
	l list.Model
}

func newAskCommandsModel(runbook *component.Runbook) (*askCommands, error) {
	if runbook == nil {
		return nil, errors.New("runbook is empty")
	}

	listItems := toItems(runbook.Steps)
	l := list.NewModelWithDelegate(listItems, runbook.Title, runbook.URL, list.NewAskDelegate())
	return &askCommands{l: l}, nil
}
func (dc *askCommands) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	dc.l.Init()
	return nil
}

func (dc *askCommands) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m, cmd := dc.l.Update(msg)
	if m, ok := m.(list.Model); ok {
		dc.l = m
	}
	return dc, cmd
}

func (dc *askCommands) View() string {
	return dc.l.View()
}

func init() {
	rootCmd.AddCommand(askCmd)
}
