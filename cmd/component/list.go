package component

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/getsavvyinc/savvy-cli/cmd/browser"
)

var docStyle = lipgloss.NewStyle().Margin(3, 3)

type ListItem struct {
	Command         string
	DescriptionText string
}

var _ list.DefaultItem = ListItem{}

func (i ListItem) Title() string       { return i.Command }
func (i ListItem) Description() string { return i.DescriptionText }
func (i ListItem) FilterValue() string {
	return strings.Join([]string{i.Command, i.DescriptionText}, " ")
}

type ListModel struct {
	list list.Model
	url  string
}

func NewListModel(items []ListItem, title string, url string) ListModel {
	var listItems []list.Item
	for _, i := range items {
		listItems = append(listItems, i)
	}
	m := ListModel{
		list: list.New(listItems, list.NewDefaultDelegate(), 0, 0),
	}
	m.list.Title = title

	editBinding := key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit online"))
	m.list.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{editBinding}
	}

	m.list.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{editBinding}
	}

	m.url = url

	return m
}

func (m ListModel) Init() tea.Cmd {
	return nil
}

// TODO: handle errors by passing in a better error message
func OpenBrowser(url string, onComplete tea.Msg, onErr tea.Msg) tea.Cmd {
	cmd := browser.Open(url)
	if cmd == nil {
		return func() tea.Msg {
			return onErr
		}
	}

	if err := cmd.Start(); err != nil {
		return func() tea.Msg {
			return onErr
		}
	}

	return func() tea.Msg {
		return onComplete
	}
}

type NopMsg struct{}

func (m ListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		if msg.String() == "e" {
			return m, OpenBrowser(m.url, NopMsg{}, NopMsg{})
		}

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	case NopMsg:
		return m, nil
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m ListModel) View() string {
	return docStyle.Render(m.list.View())
}
