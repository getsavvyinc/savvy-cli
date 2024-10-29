package export

import (
	"context"
	"errors"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/getsavvyinc/savvy-cli/client"
	"github.com/getsavvyinc/savvy-cli/cmd/component"
	"github.com/getsavvyinc/savvy-cli/cmd/component/list"
	"github.com/getsavvyinc/savvy-cli/display"
	"github.com/getsavvyinc/savvy-cli/server"
	"github.com/muesli/termenv"
)

type Exporter interface {
	ToMarkdownFile(ctx context.Context)
	ToSavvyArtifact(ctx context.Context, cl client.Client) error
}

func NewExporter(commands []*server.RecordedCommand) Exporter {
	return &exporter{
		commands: commands,
	}
}

type exporter struct {
	commands []*server.RecordedCommand
}

func (e *exporter) ToMarkdownFile(ctx context.Context) {
	panic("implement me")
}

func (e *exporter) ToSavvyArtifact(ctx context.Context, cl client.Client) error {
	gctx, cancel := context.WithCancel(ctx)
	gm := component.NewGenerateRunbookModel(e.commands, cl)
	var programOutput = termenv.NewOutput(os.Stdout, termenv.WithColorCache(true))
	p := tea.NewProgram(gm, tea.WithOutput(programOutput), tea.WithContext(gctx))
	if _, err := p.Run(); err != nil {
		err = fmt.Errorf("failed to generate runbook: %w", err)
		return err
	}

	// ensure the bubble tea program is finished before we start the next one
	cancel()
	p.Wait()

	runbook := <-gm.RunbookCh()
	m, err := newDisplayCommandsModel(runbook)
	if err != nil {
		return err
	}

	p = tea.NewProgram(m, tea.WithOutput(programOutput), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		// TODO: fail gracefully and provide users a link to view the runbook
		err = fmt.Errorf("could not display runbook: %w", err)
		return err
	}
	if runbook.URL != "" {
		display.Success("View and edit your runbook online at: " + runbook.URL)
		return nil
	}
	return nil
}

type displayCommands struct {
	l list.Model
}

func newDisplayCommandsModel(runbook *component.Runbook) (*displayCommands, error) {
	if runbook == nil {
		return nil, errors.New("runbook is empty")
	}

	listItems := toItems(runbook.Steps)
	l := list.NewModel(listItems, runbook.Title, runbook.URL, list.EditOnlineBinding)
	return &displayCommands{l: l}, nil
}

func (dc *displayCommands) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	dc.l.Init()
	return nil
}

func (dc *displayCommands) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return dc.l.Update(msg)
}

func toItems(steps []component.RunbookStep) []list.Item {
	var items []list.Item
	for _, step := range steps {
		items = append(items, list.Item{
			Command:         step.Command,
			DescriptionText: step.Description,
		})
	}
	return items
}

func (dc *displayCommands) View() string {
	return dc.l.View()
}
