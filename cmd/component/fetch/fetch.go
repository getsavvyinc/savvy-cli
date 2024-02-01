package fetch

import (
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	spinner spinner.Model
	waitMsg string

	done bool
}

var spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))

// New creates a new fetch model.
// waitMsg is the message to display while waiting for the fetch to complete.
func New(waitMsg string) Model {
	m := Model{
		spinner: newSpinner(),
		waitMsg: waitMsg,
		done:    false,
	}
	return m
}

func newSpinner() spinner.Model {
	sp := spinner.New(spinner.WithSpinner(spinner.MiniDot), spinner.WithStyle(spinnerStyle))
	mySpinner := new(spinner.Model)
	*mySpinner = sp
	mySpinner.Spinner.FPS = time.Second / 20
	return *mySpinner
}

type DoneMsg struct{}

func (m Model) Done() DoneMsg {
	return DoneMsg{}
}

func (m Model) View() string {
	if m.done {
		return ""
	}
	return m.spinner.View() + " " + m.waitMsg
}

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	if m.done {
		return nil
	}
	return m.spinner.Tick
}

// Update updates the model.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.(type) {
	case DoneMsg:
		m.done = true
		m.resetSpinner()
		return m, nil
	}
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

// The pointer receiver is important here!
func (m *Model) resetSpinner() {
	m.spinner = newSpinner()
}
