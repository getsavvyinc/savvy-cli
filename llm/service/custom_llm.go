package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/getsavvyinc/savvy-cli/config"
	"github.com/getsavvyinc/savvy-cli/idgen"
	"github.com/getsavvyinc/savvy-cli/llm"
	"github.com/getsavvyinc/savvy-cli/model"
	"github.com/getsavvyinc/savvy-cli/slice"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/sethvargo/go-retry"
)

type customSvc struct {
	cl        *openai.Client
	modelName string
}

func newCustomService(cfg *config.Config) Service {
	baseURL := cfg.LLMBaseURL
	apiKey := cfg.LLMAPIKey

	clientConfig := openai.DefaultConfig(apiKey)
	clientConfig.BaseURL = baseURL

	openaiClient := openai.NewClientWithConfig(clientConfig)

	return &customSvc{
		cl:        openaiClient,
		modelName: cfg.LLMModelName,
	}
}

func hardcodedRunbook(commands []*CommandAndID) *llm.Runbook {
	steps := make([]llm.RunbookStep, len(commands))
	for i, c := range commands {
		steps[i] = llm.RunbookStep{
			Command: c.Command,
		}
	}
	return &llm.Runbook{
		Title: "Savvy Runbook",
		Steps: steps,
	}
}

func (c *customSvc) GenerateRunbook(ctx context.Context, commands []model.RecordedCommand) (*llm.Runbook, error) {

	taggedCommands := slice.Map(commands, func(step model.RecordedCommand) *CommandAndID {
		return &CommandAndID{Command: step.Command, CommandID: idgen.New(idgen.LLMTagPrefix)}
	})

	runbook, err := c.generateRunbookTitleAndDescriptions(ctx, taggedCommands)
	if err != nil {
		err = fmt.Errorf("error generating runbook: %v", err)
		slog.Debug("error generating runbook", "err", err.Error())
		return hardcodedRunbook(taggedCommands), nil
	}

	stepByID := make(map[string]llm.RunbookStep, len(runbook.Steps))
	for _, step := range runbook.Steps {
		stepByID[step.CommandID] = step
	}

	var resultSteps []llm.RunbookStep

	// We tagged each command with an ID, so we can use that to match the command to the steps
	// NOTE: This is required since the LLM may return descriptions out of order or with some commands missing
	// We need to ensure that the descriptions are matched to the correct command
	for _, command := range taggedCommands {
		if step, ok := stepByID[command.CommandID]; ok {
			resultSteps = append(resultSteps, step)
		} else {
			resultSteps = append(resultSteps, llm.RunbookStep{
				Command:   command.Command,
				CommandID: command.CommandID,
			})
		}
	}

	return &llm.Runbook{
		Title: runbook.Title,
		Steps: resultSteps,
	}, nil
}

const (
	generateTitleAndDescriptionPrompt = `
  You are a software engineer who is an experienced oncall engineer. You are tasked with creating a runbook from the following command_id:command pairs:

command_id:command
{{range .Commands}}
  {{.CommandID}}:{{.Command}}
{{end}}

You will generate the Title for the runbook and a meaningful description for each command in the runbook.

The Title must be a short single sentences tha begins with the phrase: "How To". The title must be short and concise and must describe the purpose of the runbook. Do not make the title overly general.

The Description for each command must be short and concise. Use simple words. Limit the description to 1-2 sentences.

Do not include filler words like: "This command is used to" in the description. Get straight to the point.

Take a deep breath, do not rush, and take your time to generate the Title and Descriptions for each command. Do not hallucinate or make things up. Be as accurate as possible.


Generate json output that adheres to the following schema:
{
  title: "describe the purpose and theme of the runbook",
  steps: [
  {
    command: "command,  unchanged from the input prompt",
    command_id: "command_id, unchanged from the input prompt",
    description: "short, conscise, and helpful description of the command."
  }
}
`

	genRunbookTemplateName = "genRunbook"

	generateCommandFromAskPrompt = `Your name is Savvy. You are an expert software engineer with deep knowledge of all shell commands.

You are talking with a software engineer who needs your help generating shell commands for their query.

Query: {{.Query}}

{{if .OS}}
You are generating commands for the following operating system: {{.OS}}
{{end}}

{{if .QueryData }}

{{ if .QueryData.PreviousQuestions }}
Answer the users Query in light of these previous questions they've asked you:
{{range $_, $element := .QueryData.PreviousQuestions}}
 - {{$element}}
{{end}}
{{end}}

{{ if .QueryData.PreviousCommands }}
The user has run these commands before asking you the above Query.
{{range .QueryData.PreviousCommands}}
 - {{.}}
{{end}}

Keep these commands in mind when answering the users query.
{{end}}


{{ if .QueryData.FileData }}
File Data:

{{.QueryData.FileData}}

File Name: {{.QueryData.FileName}}
{{end}}

{{end}}

Generate shell commands to answer the users query.

Follow these guidelines when generating the commands:
- Pay attention to the users query. The commands should be relevant to the query.
{{if .QueryData}}
{{ if .QueryData.FileData }}
- Use the file data and file name to generate semantically relevant commands.
{{end}}
{{ if .QueryData.PreviousQuestions }}
- Use the previous questions to generate semantically relevant commands.
{{end}}
{{end}}
- It is okay to generate just 1 or two commands if that completely answers the users query.
- Decide which shell command or combinations of commands are required to answer the query.
- Read the manual and help pages for each selected command.
- Read relevant stackoverlfow posts, blog posts, and linux mailing lists to understand the command and query.
- Do not include commands that start with "tldr" in the generated commands unless the user's query specifically mentions "tldr".
- Include explanations for each command. The explanation should be short and concise. Use simple words. Limit the explanation to one sentence.
- If you need to add placeholder values to the command, use <placeholder> to indicate where the placeholder should be. Replace placeholder with a user friendly value
- Get straight to the point. Do not use filler words like: "This command is used to" in the explanation.
- Take a deep breath and relax. You got this!
`

	genCommandForQueryTemplateName = "genCommand"
)

var generateRunbookTitleAndDescriptionsPromptTemplate = template.Must(template.New(genRunbookTemplateName).Parse(generateTitleAndDescriptionPrompt))

var (
	GenerateRunbookSchema = jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"title": jsonschema.Definition{
				Type:        jsonschema.String,
				Description: "Title of the runbook",
			},
			"steps": jsonschema.Definition{
				Type:        jsonschema.Array,
				Description: "Steps in the runbook",
				Items: &jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"command": jsonschema.Definition{
							Type:        jsonschema.String,
							Description: "command passed in to the prompt. This should be unchanged from the input prompt.",
						},
						"command_id": jsonschema.Definition{
							Type:        jsonschema.String,
							Description: "ID of the command. This should be unchanged from the input prompt.",
						},
						"description": jsonschema.Definition{
							Type:        jsonschema.String,
							Description: "Short, conscise, and helpful description of the command",
						},
					},
				},
			},
		},
		Required: []string{"title", "steps"},
	}

	GenerateRunbookFromQuerySchema = jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"title": jsonschema.Definition{
				Type:        jsonschema.String,
				Description: "Title of the runbook from the users query",
			},
			"steps": jsonschema.Definition{
				Type:        jsonschema.Array,
				Description: "Commands that answer the users query",
				Items: &jsonschema.Definition{
					Type: jsonschema.Object,
					Properties: map[string]jsonschema.Definition{
						"code": jsonschema.Definition{
							Type:        jsonschema.String,
							Description: "shell command that answers the users query",
						},
						"description": jsonschema.Definition{
							Type:        jsonschema.String,
							Description: "Short, conscise, and helpful description of the command",
						},
					},
				},
			},
		},
		Required: []string{"title", "steps"},
	}
)

// CommandAndID is a struct that holds a command and its corresponding command_id
// CommandID is useful in the prompt to ensure that the llm doesn't change the order or omit/hallucinate any command.
type CommandAndID struct {
	Command   string `json:"command,omitempty"`
	CommandID string `json:"command_id,omitempty"`
}

func (c *customSvc) generateRunbookTitleAndDescriptions(ctx context.Context, commands []*CommandAndID) (*llm.Runbook, error) {
	buf := new(bytes.Buffer)
	if err := generateRunbookTitleAndDescriptionsPromptTemplate.Execute(buf, struct {
		Commands []*CommandAndID
	}{
		Commands: commands,
	}); err != nil {
		return nil, err
	}

	prompt := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: buf.String(),
	}

	b := retry.NewFibonacci(1 * time.Second)
	b = retry.WithMaxRetries(3, b)

	var chatResponse openai.ChatCompletionResponse
	var gerr error

	// retry on bad request errors
	if err := retry.Do(ctx, b, func(ctx context.Context) error {
		chatResponse, gerr = c.cl.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Messages:    []openai.ChatCompletionMessage{prompt},
			Model:       c.modelName,
			MaxTokens:   2500,
			Temperature: 0.1,
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
				JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
					Name:        "generate_runbook_title_and_descriptions",
					Description: "Generate a runbook title and descriptions for each command in the runbook",
					Schema:      &GenerateRunbookSchema,
					Strict:      true,
				},
			},
		})

		var oaiErr *openai.APIError
		// Sometimes, the api returns a 400 error, as it can't force a json response.
		if errors.As(gerr, &oaiErr) && oaiErr.HTTPStatusCode == http.StatusBadRequest {
			log.Printf("retry: bad request to custom llm: %v\n", oaiErr)
			return retry.RetryableError(oaiErr)
		}
		return gerr
	}); err != nil {
		log.Printf("error making request to openai: %v\n", err)
		return nil, err
	}

	if gerr != nil || len(chatResponse.Choices) != 1 {
		return nil, fmt.Errorf("Completion error: err:%v len(choices):%v\n", gerr,
			len(chatResponse.Choices))
	}

	msg := chatResponse.Choices[0].Message.Content
	if len(msg) == 0 {
		return nil, fmt.Errorf("Completion error: len(msg): %v\n", len(msg))
	}

	var runbook llm.Runbook
	if err := json.Unmarshal([]byte(msg), &runbook); err != nil {
		return nil, err
	}
	return &runbook, nil
}

func queryHasOS(query string) (string, bool) {
	if strings.Contains(query, "linux") {
		return "linux", true
	}
	if strings.Contains(query, "ubuntu") {
		return "ubuntu", true
	}
	if strings.Contains(query, "centos") {
		return "centos", true
	}
	if strings.Contains(query, "rhel") {
		return "rhel", true
	}
	if strings.Contains(query, "debian") {
		return "debian", true
	}
	if strings.Contains(query, "macos") {
		return "macos", true
	}
	if strings.Contains(query, "mac") {
		return "macos", true
	}
	if strings.Contains(query, "mac os") {
		return "macos", true
	}

	if strings.Contains(query, "os x") {
		return "macos", true
	}

	if strings.Contains(query, "darwin") {
		return "macos", true
	}

	if strings.Contains(query, "windows") {
		return "windows", true
	}

	return "", false
}

func (c *customSvc) Ask(ctx context.Context, question *model.QuestionInfo) (*llm.Runbook, error) {
	buf := new(bytes.Buffer)
	var osName string
	if question != nil && len(question.Tags) > 0 {
		osName = question.Tags["os"]
	}

	if qos, ok := queryHasOS(question.Question); ok {
		osName = qos
	}

	if err := template.Must(template.New(genCommandForQueryTemplateName).Parse(generateCommandFromAskPrompt)).Execute(buf, struct {
		Query     string
		QueryData *model.QuestionInfo
		OS        string
	}{
		Query:     question.Question,
		QueryData: question,
		OS:        osName,
	}); err != nil {
		return nil, err
	}

	prompt := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: buf.String(),
	}

	resp, err := c.cl.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Messages:    []openai.ChatCompletionMessage{prompt},
		Model:       c.modelName,
		MaxTokens:   2500,
		Temperature: 0.1,
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
			JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
				Name:        "generate_runbook_title_and_descriptions_from_query",
				Description: "Answer the users query with shell commands",
				Schema:      &GenerateRunbookFromQuerySchema,
				Strict:      true,
			},
		},
	})

	if err != nil {
		err = fmt.Errorf("ask: error making request to openai: %w", err)
		return nil, err
	}

	if err != nil || len(resp.Choices) != 1 {
		err = fmt.Errorf("ask: Completion error:%w len(choices):%v", err, len(resp.Choices))
		return nil, err
	}

	msg := resp.Choices[0].Message.Content

	if len(msg) == 0 {
		return nil, fmt.Errorf("Completion error: len(msg): %v\n", len(msg))
	}

	var runbook llm.Runbook
	if err := json.Unmarshal([]byte(msg), &runbook); err != nil {
		return nil, err
	}

	return &runbook, nil
}
