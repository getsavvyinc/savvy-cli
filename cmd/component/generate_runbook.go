package component

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/getsavvyinc/savvy-cli/client"
	"github.com/getsavvyinc/savvy-cli/cmd/component/fetch"
)

type GenerateRunbookModel struct {
	fetch.Model
	cl client.Client

	commands  []string
	runbookCh chan *Runbook
	done      bool
}

func (m GenerateRunbookModel) RunbookCh() chan *Runbook {
	return m.runbookCh
}

func NewGenerateRunbookModel(
	commands []string,
	cl client.Client,
) GenerateRunbookModel {
	m := GenerateRunbookModel{
		Model:     fetch.New("Generating runbook..."),
		cl:        cl,
		commands:  commands,
		runbookCh: make(chan *Runbook, 1),
	}
	return m
}

type GenerateRunbookDoneMsg struct {
	GeneratedRunbook *client.GeneratedRunbook
	Err              error
}

func (m *GenerateRunbookModel) IsDone() bool {
	return m.done
}

func (m *GenerateRunbookModel) Generate() tea.Msg {

	generatedRunbook, err := m.cl.GenerateRunbook(context.Background(), m.commands)
	return GenerateRunbookDoneMsg{
		GeneratedRunbook: generatedRunbook,
		Err:              err,
	}
}

func (m GenerateRunbookModel) Init() tea.Cmd {
	return tea.Sequence(m.Model.Init(), m.Generate)
}

func (m GenerateRunbookModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case GenerateRunbookDoneMsg:
		m.runbookCh <- toRunbook(msg.GeneratedRunbook)
		m.done = true
		return m, tea.Quit
	}
	m.Model, cmd = m.Model.Update(msg)
	return m, cmd
}

type Runbook struct {
	Title string
	Steps []RunbookStep
	URL   string
}

type RunbookStep struct {
	Command     string
	Description string
}

func toRunbook(grb *client.GeneratedRunbook) *Runbook {
	if grb == nil {
		return nil
	}

	return &Runbook{
		Title: grb.Runbook.Title,
		Steps: toSteps(grb.Runbook.Steps),
		URL:   grb.URL,
	}
}

func toSteps(steps []client.Step) []RunbookStep {
	var s []RunbookStep
	for _, step := range steps {
		s = append(s, toStep(step))
	}
	return s
}

func toStep(step client.Step) RunbookStep {
	return RunbookStep{
		Command:     step.Command,
		Description: step.Description,
	}
}

func (m GenerateRunbookModel) View() string {
	if m.done {
		return ""
	}
	return m.Model.View()
}
