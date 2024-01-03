package cmd

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/getsavvyinc/savvy-cli/cmd/browser"
	"github.com/getsavvyinc/savvy-cli/config"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to savvy",
	Long:  `Login allows users to use Google SSO to login to savvy.`,
	Run:   runLogin,
}

var savvyLoginURL string = config.APIHost() + "/login"

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

func runLogin(cmd *cobra.Command, args []string) {
	browserCmd := browser.Open(savvyLoginURL)
	if browserCmd == nil {
		fmt.Printf("\nVisit %s in your browser\n", savvyLoginURL)
	} else {
		if err := browserCmd.Start(); err != nil {
			fmt.Printf("Error opening browser: %v\n", err)
			os.Exit(1)
		}
	}
	p := tea.NewProgram(initialModel())

	m, err := p.Run()
	if err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	} else {
		model := m.(loginModel)
		if model.err != nil {
			fmt.Printf("Error exchanging token: %v\n", model.err)
			os.Exit(1)
		}
		// Handle the token here (e.g., store it)
		cfg := config.Config{Token: model.textInput.Value()}
		if err := cfg.Save(); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Login successful!")
	}
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
