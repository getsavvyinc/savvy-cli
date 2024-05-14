package list

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/getsavvyinc/savvy-cli/cmd/browser"
)

var docStyle = lipgloss.NewStyle().Margin(3, 3)

type Item struct {
	Command         string
	DescriptionText string
}

var _ list.DefaultItem = Item{}

func (i Item) Title() string       { return i.Command }
func (i Item) Description() string { return i.DescriptionText }
func (i Item) FilterValue() string {
	return strings.Join([]string{i.Command, i.DescriptionText}, " ")
}

type Model struct {
	list        list.Model
	url         string
	editBinding key.Binding
}

func NewModelWithDelegate(items []Item, title string, url string, delegate list.ItemDelegate) Model {
	var listItems []list.Item
	for _, i := range items {
		listItems = append(listItems, i)
	}

	m := Model{
		list: list.New(listItems, delegate, 0, 0),
	}
	m.list.Title = title

	m.editBinding = key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit online"))
	m.list.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{m.editBinding}
	}

	m.list.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{m.editBinding}
	}

	m.url = url

	return m
}

func NewModel(items []Item, title string, url string) Model {
	return NewModelWithDelegate(items, title, url, list.NewDefaultDelegate())
}

func (m Model) Init() tea.Cmd {
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

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		if msg.String() == "e" && m.list.FilterState() == list.Unfiltered {
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

func (m Model) View() string {
	return docStyle.Render(m.list.View())
}
