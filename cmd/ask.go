package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	huhSpinner "github.com/charmbracelet/huh/spinner"
	"github.com/getsavvyinc/savvy-cli/client"
	"github.com/getsavvyinc/savvy-cli/cmd/browser"
	"github.com/getsavvyinc/savvy-cli/cmd/component"
	"github.com/getsavvyinc/savvy-cli/cmd/component/list"
	"github.com/getsavvyinc/savvy-cli/display"
	"github.com/getsavvyinc/savvy-cli/server"
	"github.com/getsavvyinc/savvy-cli/slice"
	"github.com/spf13/cobra"
)

var askCmd = &cobra.Command{
	Use:   "ask",
	Short: "Ask Savvy a question and it will generate a command",
	Example: `
  savvy ask # interactive mode
  savvy ask "how do I deploy a k8s daemonset?"
  savvy ask "how do I parse a x509 cert"
  savvy ask "how do I find the process id listening on a port?"
  savvy ask "how do I quit vim?"
  savvy ask "extract filenames from the name key in each line of li_ids.txt" --file /path/to/li_ids.txt
  savvy ask "what's the next command after this?" --history # pass in historical context
  `,
	Long: `
  Ask Savvy a question and it will generate a command for you.

  If a file path is provided, Savvy will use the contents of the file to generate a command.
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

		// get info about the os from os pkg: mac/darwin, linux, windows
		goos := runtime.GOOS
		if goos == "darwin" {
			goos = "macos, darwin, osx"
		}

		fileData, err := fileData(filePath)
		if err != nil {
			display.Error(err)
			os.Exit(1)
		}

		var historyCmds []*server.RecordedCommand
		if useHistory {
			historyCmds, err = selectAndExpandHistory(ctx, logger)
			if err != nil {
				display.FatalErrWithSupportCTA(err)
				return
			}
		}

		var question string
		if len(args) > 0 {
			// be defensive: users can pass questions as one string or multiple strings
			question = strings.Join(args[:], " ")
		}

		params := &AskParams{
			goos:             goos,
			fileData:         fileData,
			filePath:         filePath,
			refine:           false,
			previousCommands: historyCmds,
		}

		var state *runAskTerminalState
		for {
			state = runAsk(ctx, cl, question, params)
			if state == nil || !state.refinePrompt {
				break
			}
			params.refine = true
			question = ""
		}

		selectedCommand := state.selectedCommand
		if selectedCommand != "" {
			if err := clipboard.WriteAll(selectedCommand); err != nil {
				display.Info(selectedCommand)
				return
			}
			display.Info(fmt.Sprintf("Copied to clipboard: %s", selectedCommand))
		}

		if state.createRunbook {
			result, err := createRunbook(ctx, cl, state.runbook)
			if err != nil {
				display.ErrorWithSupportCTA(err)
				os.Exit(1)
			}
			display.Success(fmt.Sprintf("Runbook %s created successfully!", result.Runbook.Title))
			browser.Open(result.URL)
		}

		if state.runSteps {
			if err := runRunbook(ctx, state.runbook); err != nil {
				display.ErrorWithSupportCTA(
					fmt.Errorf("failed to run runbook %s: %w", state.runbook.Title, err),
				)
				return
			}

			var confirmation bool
			confirmCleanup := huh.NewConfirm().
				Title(fmt.Sprintf("Save %q", state.runbook.Title)).
				Affirmative("Save").
				Negative("Cancel").
				Value(&confirmation)
			if err := huh.NewForm(huh.NewGroup(confirmCleanup)).Run(); err != nil {
				display.ErrorWithSupportCTA(err)
				return
			}

			// exit early if user doesn't want to save
			if !confirmation {
				return
			}

			result, err := createRunbook(ctx, cl, state.runbook)
			if err != nil {
				display.ErrorWithSupportCTA(err)
				os.Exit(1)
			}
			display.Successf("Created %q successfully! You can check it out here: %s", result.Runbook.Title, result.URL)
		}
	},
}

func createRunbook(ctx context.Context, cl client.Client, runbook *client.Runbook) (*client.GeneratedRunbook, error) {
	result, err := cl.SaveRunbook(ctx, runbook)
	if err == nil {
		return result, nil
	} else if errors.Is(err, client.ErrCannotUseGuest) {
		return loginAndCreateRunbook(ctx, runbook)
	}
	return result, err
}

func loginAndCreateRunbook(ctx context.Context, runbook *client.Runbook) (*client.GeneratedRunbook, error) {
	runLogin()
	// then create Runbook
	cl, err := client.New()
	if err != nil {
		return nil, err
	}
	return cl.SaveRunbook(ctx, runbook)
}

type AskParams struct {
	goos              string
	fileData          []byte
	filePath          string
	refine            bool
	previousQuestions []string
	previousCommands  []*server.RecordedCommand
}

type runAskTerminalState struct {
	selectedCommand string
	refinePrompt    bool
	createRunbook   bool
	runSteps        bool
	runbook         *client.Runbook
}

func runAsk(ctx context.Context, cl client.Client, question string, askParams *AskParams) *runAskTerminalState {
	logger := loggerFromCtx(ctx).With("command", "ask", "method", "runAsk")
	if len(question) == 0 {
		// interactive mode
		title := "Ask Savvy a question"
		if askParams.refine {
			title = "Refine your question"
		}
		text := huh.NewText().Title(title).Value(&question)
		form := huh.NewForm(huh.NewGroup(text))
		if err := form.Run(); err != nil {
			display.ErrorWithSupportCTA(err)
			os.Exit(1)
		}
	}

	if len(question) == 0 {
		display.Info("Exiting...")
		return nil
	}

	qi := client.QuestionInfo{
		Question: question,
		Tags: map[string]string{
			"os": askParams.goos,
		},
		FileData:          askParams.fileData,
		FileName:          path.Base(askParams.filePath),
		PreviousQuestions: askParams.previousQuestions[:],
		PreviousCommands:  slice.Map(askParams.previousCommands, func(c *server.RecordedCommand) string { return c.Command }),
	}
	askParams.previousQuestions = append(askParams.previousQuestions, question)

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
		return &runAskTerminalState{
			selectedCommand: m.l.SelectedCommand(),
			refinePrompt:    m.refinePrompt,
			createRunbook:   m.saveAsRunbook,
			runSteps:        m.runSteps,
			runbook:         runbook,
		}
	}
	return nil
}

type askCommands struct {
	l             list.Model
	refinePrompt  bool
	saveAsRunbook bool
	runSteps      bool
}

var RefinePromptHelpBinding = list.NewHelpBinding("p", "refine prompt")
var SaveAsRunbookHelpBinding = list.NewHelpBinding("s", "save as runbook")
var ExecuteCommandsHelpBinding = list.NewHelpBinding("r", "run all commands")

func newAskCommandsModel(runbook *component.Runbook) (*askCommands, error) {
	if runbook == nil {
		return nil, errors.New("runbook is empty")
	}

	listItems := toItems(runbook.Steps)

	l := list.NewModelWithDelegate(listItems, runbook.Title, runbook.URL, list.NewAskDelegate(), RefinePromptHelpBinding, SaveAsRunbookHelpBinding, ExecuteCommandsHelpBinding)
	return &askCommands{l: l}, nil
}
func (dc *askCommands) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	dc.l.Init()
	return nil
}

func (dc *askCommands) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case list.RefinePromptMsg:
		dc.refinePrompt = true
		return dc, tea.Quit
	case list.SaveAsRunbookMsg:
		dc.saveAsRunbook = true
		return dc, tea.Quit
	case list.SaveAsRunbookAndExecuteMsg:
		dc.runSteps = true
		return dc, tea.Quit
	}

	m, cmd := dc.l.Update(msg)
	if m, ok := m.(list.Model); ok {
		dc.l = m
	}
	return dc, cmd
}

func (dc *askCommands) View() string {
	return dc.l.View()
}

func fileData(filePath string) ([]byte, error) {
	if filePath == "" {
		return nil, nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	if stat.Size() > 20*1024 {
		return nil, errors.New("file must be less than 20KB")
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return data, nil
}

var filePath string
var useHistory bool

func init() {
	rootCmd.AddCommand(askCmd)
	askCmd.Flags().StringVarP(&filePath, "file", "f", "", "File path for ask to read and use while generating an answer")
	askCmd.Flags().BoolVarP(&useHistory, "history", "", false, "Provide historical context to Savvy's AI model")
}
