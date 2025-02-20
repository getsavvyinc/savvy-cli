package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"strings"

	"github.com/getsavvyinc/savvy-cli/authz"
	"github.com/getsavvyinc/savvy-cli/config"
	"github.com/getsavvyinc/savvy-cli/extension"
	"github.com/getsavvyinc/savvy-cli/llm"
	"github.com/getsavvyinc/savvy-cli/llm/service"
	"github.com/getsavvyinc/savvy-cli/model"
)

type RunbookClient interface {
	RunbookByID(ctx context.Context, id string) (*Runbook, error)
	Runbooks(ctx context.Context, opts RunbooksOpt) ([]RunbookInfo, error)
}

type RunbookSaver interface {
	SaveRunbook(ctx context.Context, runbook *Runbook) (*GeneratedRunbook, error)
}

type Client interface {
	RunbookClient
	RunbookSaver
	WhoAmI(ctx context.Context) (string, error)
	GenerateRunbookV2(ctx context.Context, commands []model.RecordedCommand, links []extension.HistoryItem) (*GeneratedRunbook, error)
	// Deprecated. Use GenerateRunbookV2 instead
	GenerateRunbook(ctx context.Context, commands []string) (*GeneratedRunbook, error)
	Ask(ctx context.Context, question *model.QuestionInfo) (*Runbook, error)
	Explain(ctx context.Context, code *model.CodeInfo) (<-chan string, error)
	StepContentByStepID(ctx context.Context, stepID string) (*StepContent, error)
}

type StepContent struct {
	Content []byte      `json:"content"`
	Mode    fs.FileMode `json:"mode"`
	Name    string      `json:"name"`
	DirPath string      `json:"dir_path"`
}

// UnmarshalJSON is a custom unmarshaler for StepContent that handles the mode as a float64.
func (sc *StepContent) UnmarshalJSON(data []byte) error {
	// Create a temporary structure that mirrors StepContent but with Mode as float64.
	type Alias StepContent
	aux := &struct {
		Mode float64 `json:"mode"`
		*Alias
	}{
		Alias: (*Alias)(sc),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Convert the float64 mode to fs.FileMode and assign it to StepContent.Mode.
	sc.Mode = fs.FileMode(int64(aux.Mode))
	return nil
}

type FileInfo struct {
	Mode    fs.FileMode `json:"mode,omitempty"`
	Content []byte      `json:"content,omitempty"`
	Path    string      `json:"path,omitempty"`
}

type client struct {
	cl      *http.Client
	llmSvc  service.Service
	apiHost string
}

var _ Client = (*client)(nil)

var ErrInvalidClient = errors.New("invalid client")

func New() (Client, error) {
	cfg, err := config.LoadFromFile()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidClient, err)
	}

	cl := &http.Client{
		Transport: authz.NewRoundTripper(cfg.Token, config.Version(), ErrInvalidClient),
	}

	c := &client{
		cl:      cl,
		llmSvc:  service.New(cfg, cl),
		apiHost: config.APIHost(),
	}

	// validate token as early as possible
	if _, err := c.WhoAmI(context.Background()); err != nil && errors.Is(err, ErrInvalidClient) {
		return nil, err
	}
	return c, nil
}

// apiURL returns the full url to the api endpoint
// path must start with a slash. e.g. /api/v1/whoami
// apiURL will add a slash if it's missing
func (c *client) apiURL(path string) string {
	return genAPIURL(c.apiHost, path)
}

func genAPIURL(host, path string) string {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return host + path
}

func (c *client) WhoAmI(ctx context.Context) (string, error) {
	cl := c.cl
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiURL("/api/v1/whoami"), nil)
	if err != nil {
		return "", err
	}
	resp, err := cl.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	whoami, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(whoami), nil
}

type GeneratedRunbook struct {
	Runbook Runbook `json:"runbook"`
	URL     string  `json:"url"`
}

type Runbook struct {
	RunbookID string                  `json:"runbook_id"`
	Title     string                  `json:"title"`
	Steps     []Step                  `json:"steps"`
	Links     []extension.HistoryItem `json:"links"`
}

type RunbookInfo struct {
	RunbookID string `json:"runbook_id"`
	Title     string `json:"title"`
}

type StepTypeEnum string

const (
	StepTypeCode StepTypeEnum = "code"
	StepTypeFile StepTypeEnum = "file"
)

type Step struct {
	Type        StepTypeEnum `json:"type"`
	Description string       `json:"description"`
	Command     string       `json:"command"`
}

func (rb *Runbook) Commands() []string {
	var commands []string
	for _, step := range rb.Steps {
		commands = append(commands, step.Command)
	}
	return commands
}

func (c *client) GenerateRunbookV2(ctx context.Context, commands []model.RecordedCommand, links []extension.HistoryItem) (*GeneratedRunbook, error) {
	generatedRunbook, err := c.llmSvc.GenerateRunbook(ctx, commands)
	if err != nil {
		return nil, err
	}

	// Save the generated Runbook
	clientRunbook := toClientRunbook(generatedRunbook)
	if len(links) > 0 {
		clientRunbook.Links = links
	}

	savedRunbook, err := c.SaveRunbook(ctx, clientRunbook)
	if err != nil {
		return nil, err
	}
	return savedRunbook, nil
}

func toClientRunbook(rb *llm.Runbook) *Runbook {
	clientSteps := make([]Step, len(rb.Steps))
	for i, step := range rb.Steps {
		clientSteps[i] = Step{
			Type:        StepTypeCode,
			Description: step.Description,
			Command:     step.Command,
		}
	}
	return &Runbook{
		Title: rb.Title,
		Steps: clientSteps,
	}
}

func (c *client) SaveRunbook(ctx context.Context, runbook *Runbook) (*GeneratedRunbook, error) {
	cl := c.cl
	bs, err := json.Marshal(runbook)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiURL("/api/v1/runbook"), bytes.NewReader(bs))
	if err != nil {
		return nil, err
	}
	resp, err := cl.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var generatedRunbook GeneratedRunbook
	if err := json.NewDecoder(resp.Body).Decode(&generatedRunbook); err != nil {
		return nil, err
	}
	return &generatedRunbook, nil
}

func (c *client) GenerateRunbook(ctx context.Context, commands []string) (*GeneratedRunbook, error) {
	cl := c.cl
	bs, err := json.Marshal(struct{ Commands []string }{commands})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiURL("/api/v1/generate_runbook"), bytes.NewReader(bs))
	if err != nil {
		return nil, err
	}
	resp, err := cl.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var generatedRunbook GeneratedRunbook
	if err := json.NewDecoder(resp.Body).Decode(&generatedRunbook); err != nil {
		return nil, err
	}
	return &generatedRunbook, nil
}

func (c *client) RunbookByID(ctx context.Context, id string) (*Runbook, error) {
	cl := c.cl
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiURL("/api/v1/runbook"), nil)
	if err != nil {
		return nil, err
	}

	qp := req.URL.Query()
	qp.Set("runbook_id", id)
	req.URL.RawQuery = qp.Encode()

	resp, err := cl.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var runbook Runbook
	if err := json.NewDecoder(resp.Body).Decode(&runbook); err != nil {
		return nil, err
	}
	return &runbook, nil
}

type RunbooksOpt struct {
	ExcludeTeamRunbooks bool
}

func (c *client) Runbooks(ctx context.Context, opts RunbooksOpt) ([]RunbookInfo, error) {
	cl := c.cl
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiURL("/api/v1/list_runbooks/all"), nil)
	if opts.ExcludeTeamRunbooks {
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, c.apiURL("/api/v1/list_runbooks"), nil)
	}
	if err != nil {
		return nil, err
	}
	resp, err := cl.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var runbooks []RunbookInfo
	if err := json.NewDecoder(resp.Body).Decode(&runbooks); err != nil {
		return nil, err
	}
	return runbooks, nil
}

func (c *client) Ask(ctx context.Context, question *model.QuestionInfo) (*Runbook, error) {
	answer, err := c.llmSvc.Ask(ctx, question)
	if err != nil {
		return nil, err
	}
	return toClientRunbook(answer), nil
}

func (c *client) StepContentByStepID(ctx context.Context, stepID string) (*StepContent, error) {
	cl := c.cl
	apiPath := fmt.Sprintf("/api/v1/step/content/%s", stepID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiURL(apiPath), nil)
	if err != nil {
		return nil, err
	}

	resp, err := cl.Do(req)
	if err != nil {
		return nil, err
	}

	var stepContent StepContent

	if err := json.NewDecoder(resp.Body).Decode(&stepContent); err != nil {
		return nil, err
	}
	return &stepContent, nil
}

func (c *client) Explain(ctx context.Context, code *model.CodeInfo) (<-chan string, error) {
	return c.llmSvc.Explain(ctx, code)
}

func VerifyLogin() error {
	_, err := New()
	return err
}
