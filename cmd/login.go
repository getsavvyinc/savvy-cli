package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/getsavvyinc/savvy-cli/client"
	"github.com/getsavvyinc/savvy-cli/cmd/browser"
	"github.com/getsavvyinc/savvy-cli/config"
	"github.com/getsavvyinc/savvy-cli/display"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to savvy",
	Long:  `Login allows users to use Google SSO to login to savvy.`,
	Run:   runLoginCmd,
}

var savvyLoginURL string = config.DashboardHost() + "/cli_login"

type loginModel struct {
	textInput textinput.Model
	err       error
}

func initialModel() loginModel {
	ti := textinput.New()
	ti.Placeholder = "Paste your login token here"
	ti.Prompt = "🔑 "
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 50
	ti.EchoCharacter = '*'
	ti.EchoMode = textinput.EchoPassword

	return loginModel{textInput: ti}
}

func (m loginModel) Init() tea.Cmd {
	return textinput.Blink
}

func (lm loginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			return lm, tea.Quit
		case tea.KeyCtrlC, tea.KeyEsc:
			return lm, tea.Quit
		}
	}

	var cmd tea.Cmd
	lm.textInput, cmd = lm.textInput.Update(msg)
	return lm, cmd
}

func (lm loginModel) View() string {
	return fmt.Sprintf(
		"Paste login token:\n%s\n\nPress Enter when done, Esc to quit.",
		lm.textInput.View(),
	)
}

func runLoginCmd(cmd *cobra.Command, args []string) {
	force, err := cmd.Flags().GetBool(forceLoginFlag)
	if err != nil {
		display.ErrorWithSupportCTA(fmt.Errorf("error parsing flags: %w", err))
		os.Exit(1)
	}

	if err := verifyLogin(); err == nil && !force {
		display.Info("You are already logged in!")
		display.Info("Run `savvy login --force` to get a new token")
		return
	}

	runLogin()
}

func runLogin() {
	browser.Open(savvyLoginURL)

	p := tea.NewProgram(initialModel())

	m, err := p.Run()
	if err != nil {
		display.ErrorWithSupportCTA(fmt.Errorf("login error: %w\n", err))
		os.Exit(1)
	} else {
		model := m.(loginModel)
		if model.err != nil {
			display.ErrorWithSupportCTA(fmt.Errorf("login failed: %w", model.err))
			os.Exit(1)
		}
		// Handle the token here (e.g., store it)
		tok := model.textInput.Value()
		// Remove quotes and braces and spaces from token
		tok = strings.Trim(tok, "\"{} ")

		defer func() {
			if err := verifyLogin(); err != nil {
				display.ErrorWithSupportCTA(fmt.Errorf("login failed: %w", err))
				os.Exit(1)
			}
			display.Success("Login successful!")
		}()

		cfg := config.Config{Token: tok}
		if err := cfg.Save(); err != nil {
			err = fmt.Errorf("error saving config: %w", err)
			display.ErrorWithSupportCTA(err)
			os.Exit(1)
		}
	}
}

func verifyLogin() error {
	_, err := client.New()
	return err
}

const forceLoginFlag = "force"
const forceLoginFlagShort = "f"

func init() {
	loginCmd.Flags().BoolP(forceLoginFlag, forceLoginFlagShort, false, "Force new login flow")
	rootCmd.AddCommand(loginCmd)
}
