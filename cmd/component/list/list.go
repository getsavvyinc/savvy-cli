package list

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/getsavvyinc/savvy-cli/cmd/browser"
	"github.com/getsavvyinc/savvy-cli/slice"
)

var docStyle = lipgloss.NewStyle().Margin(3, 3)

var helpStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"})

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
	list            list.Model
	url             string
	helpBindings    []HelpBinding
	helpKeys        []string
	selectedCommand string
}

// TODO: we should be able to specify handlers for key bindings here
type HelpBinding struct {
	Binding key.Binding
}

func NewModelWithDelegate(items []Item, title string, url string, delegate list.ItemDelegate, helpBindings ...HelpBinding) Model {
	var listItems []list.Item
	for _, i := range items {
		listItems = append(listItems, i)
	}

	m := Model{
		list:         list.New(listItems, delegate, 0, 0),
		helpBindings: helpBindings,
	}
	m.list.Title = title
	m.list.Styles.HelpStyle = helpStyle
	m.list.Help.Styles.ShortKey = helpStyle
	m.list.Help.Styles.FullKey = helpStyle
	m.list.Help.Styles.ShortDesc = helpStyle
	m.list.Help.Styles.FullDesc = helpStyle

	keyBindings := slice.Map(helpBindings, func(h HelpBinding) key.Binding {
		return h.Binding
	})

	m.helpKeys = slice.Map(helpBindings, func(h HelpBinding) string {
		if len(h.Binding.Keys()) > 0 {
			return h.Binding.Keys()[0]
		}
		return ""
	})

	if len(keyBindings) > 0 {
		m.list.AdditionalFullHelpKeys = func() []key.Binding {
			return keyBindings
		}

		m.list.AdditionalShortHelpKeys = func() []key.Binding {
			return keyBindings
		}
	}

	m.url = url

	return m
}

var EditOnlineBinding = HelpBinding{
	Binding: key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit online")),
}

func NewHelpBinding(k, description string) HelpBinding {
	return HelpBinding{
		Binding: key.NewBinding(key.WithKeys(k), key.WithHelp(k, description)),
	}
}

func NewModel(items []Item, title string, url string, helpBindings ...HelpBinding) Model {
	return NewModelWithDelegate(items, title, url, list.NewDefaultDelegate(), helpBindings...)
}

func (m Model) Init() tea.Cmd {
	return nil
}

// TODO: handle errors by passing in a better error message
func OpenBrowser(url string, onComplete tea.Msg, onErr tea.Msg) tea.Cmd {
	cmd := browser.OpenCmd(url)
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

type RefinePromptMsg struct{}
type SaveAsRunbookMsg struct{}
type SaveAsRunbookAndExecuteMsg struct{}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		// Handle the "e" key binding
		if msg.String() == "e" && m.list.FilterState() == list.Unfiltered && slice.Has(m.helpKeys, "e") {
			return m, OpenBrowser(m.url, NopMsg{}, NopMsg{})
		}

		if msg.String() == "p" && m.list.FilterState() == list.Unfiltered && slice.Has(m.helpKeys, "p") {
			return m, func() tea.Msg { return RefinePromptMsg{} }
		}

		if msg.String() == "s" && m.list.FilterState() == list.Unfiltered && slice.Has(m.helpKeys, "s") {
			return m, func() tea.Msg { return SaveAsRunbookMsg{} }
		}

		if msg.String() == "r" && m.list.FilterState() == list.Unfiltered && slice.Has(m.helpKeys, "r") {
			return m, func() tea.Msg { return SaveAsRunbookAndExecuteMsg{} }
		}

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	case NopMsg:
		return m, nil
	case SelectedCommandMsg:
		m.selectedCommand = msg.Command
		return m, tea.Quit
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	return docStyle.Render(m.list.View())
}

func (m Model) SelectedCommand() string {
	return m.selectedCommand
}
