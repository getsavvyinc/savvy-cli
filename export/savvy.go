package export

import (
	"context"
	"errors"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/getsavvyinc/savvy-cli/client"
	"github.com/getsavvyinc/savvy-cli/cmd/component"
	"github.com/getsavvyinc/savvy-cli/cmd/component/list"
	"github.com/getsavvyinc/savvy-cli/display"
	"github.com/getsavvyinc/savvy-cli/export/markdown"
	"github.com/getsavvyinc/savvy-cli/extension"
	"github.com/getsavvyinc/savvy-cli/server"
	"github.com/getsavvyinc/savvy-cli/slice"
	"github.com/muesli/termenv"
)

type Exporter interface {
	Export(ctx context.Context) error
	toMarkdownFile(ctx context.Context) error
	toSavvyArtifact(ctx context.Context) error
}

const (
	MarkdownFile  = "md"
	SavvyWorkflow = "savvy"
)

func (e *exporter) Export(ctx context.Context) error {
	var exportFormat string
	if err := huh.NewSelect[string]().
		Title("Export Commands").
		Description("Select an export format").
		Options(
			huh.NewOption("Local Markdown File", MarkdownFile),
			huh.NewOption("Savvy Workflow (Recommended)", SavvyWorkflow),
		).Value(&exportFormat).Run(); err != nil {
		return err
	}

	switch exportFormat {
	case MarkdownFile:
		return e.toMarkdownFile(ctx)
	case SavvyWorkflow:
		return e.toSavvyArtifact(ctx)
	default:
		return errors.New("invalid export format")
	}

}

func NewExporter(commands []*server.RecordedCommand, links []extension.HistoryItem) Exporter {
	return &exporter{
		commands: commands,
		links:    links,
		mdSvc:    markdown.NewService(),
	}
}

type exporter struct {
	commands []*server.RecordedCommand
	links    []extension.HistoryItem
	mdSvc    markdown.Service
}

func (e *exporter) toMarkdownFile(ctx context.Context) error {
	commands := slice.Map(e.commands, func(rc *server.RecordedCommand) string {
		return rc.Command
	})

	return e.mdSvc.ToMarkdownFile(ctx, commands, e.links)
}

func (e *exporter) toSavvyArtifact(ctx context.Context) error {
	cl, err := client.GetLoggedInClient()
	if err != nil && errors.Is(err, client.ErrInvalidClient) {
		display.Error(errors.New("You must be logged in to export an Artifact to Savvy. Please run `savvy login`"))
		os.Exit(1)
	} else if err != nil {
		display.ErrorWithSupportCTA(err)
		os.Exit(1)
	}

	gctx, cancel := context.WithCancel(ctx)
	gm := component.NewGenerateRunbookModel(e.commands, e.links, cl)
	var programOutput = termenv.NewOutput(os.Stdout, termenv.WithColorCache(true))
	p := tea.NewProgram(gm, tea.WithOutput(programOutput), tea.WithContext(gctx))
	if _, err := p.Run(); err != nil {
		err = fmt.Errorf("failed to generate runbook: %w", err)
		cancel()
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
	l := list.NewModel(listItems, runbook.Title, runbook.URL, list.EditOnlineBinding, list.ViewAttachedLinksBinding)
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
