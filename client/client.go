package client

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"strings"

	"github.com/getsavvyinc/savvy-cli/config"
)

type Client interface {
	WhoAmI(ctx context.Context) (string, error)
	GenerateRunbookV2(ctx context.Context, commands []RecordedCommand) (*GeneratedRunbook, error)
	// Deprecated. Use GenerateRunbookV2 instead
	GenerateRunbook(ctx context.Context, commands []string) (*GeneratedRunbook, error)
	RunbookByID(ctx context.Context, id string) (*Runbook, error)
	Runbooks(ctx context.Context) ([]RunbookInfo, error)
	Ask(ctx context.Context, question QuestionInfo) (*Runbook, error)
	SaveRunbook(ctx context.Context, runbook *Runbook) (*GeneratedRunbook, error)
	Explain(ctx context.Context, code CodeInfo) (<-chan string, error)
	StepContentByStepID(ctx context.Context, stepID string) (*StepContent, error)
}

type RecordedCommand struct {
	Command  string    `json:"command"`
	Prompt   string    `json:"prompt,omitempty"`
	FileInfo *FileInfo `json:"file_info,omitempty"`
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
	apiHost string
}

var _ Client = (*client)(nil)

var ErrInvalidClient = errors.New("invalid client")

func New() (Client, error) {
	cfg, err := config.LoadFromFile()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidClient, err)
	}

	c := &client{
		cl: &http.Client{
			Transport: &AuthorizedRoundTripper{
				token:        cfg.Token,
				savvyVersion: config.Version(),
			},
		},
		apiHost: config.APIHost(),
	}

	// validate token as early as possible
	if _, err := c.WhoAmI(context.Background()); err != nil && errors.Is(err, ErrInvalidClient) {
		return nil, err
	}
	return c, nil
}

type AuthorizedRoundTripper struct {
	token        string
	savvyVersion string
}

func (a *AuthorizedRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to ensure thread safety
	clonedReq := req.Clone(req.Context())
	clonedReq.Header.Set("Authorization", "Bearer "+a.token)
	clonedReq.Header.Set("X-Savvy-Version", a.savvyVersion)

	// Use the embedded Transport to perform the actual request
	res, err := http.DefaultTransport.RoundTrip(clonedReq)
	if err != nil {
		err = fmt.Errorf("%w: %v", ErrInvalidClient, err)
		return nil, err
	}

	// If we get a 401 Unauthorized, then the token is expired
	// and we need to refresh it
	if res.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("%w: invalid token", ErrInvalidClient)
	}
	return res, err
}

// apiURL returns the full url to the api endpoint
// path must start with a slash. e.g. /api/v1/whoami
// apiURL will add a slash if it's missing
func (c *client) apiURL(path string) string {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return c.apiHost + path
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
	RunbookID string `json:"runbook_id"`
	Title     string `json:"title"`
	Steps     []Step `json:"steps"`
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

func (c *client) GenerateRunbookV2(ctx context.Context, commands []RecordedCommand) (*GeneratedRunbook, error) {
	cl := c.cl
	bs, err := json.Marshal(struct{ Commands []RecordedCommand }{Commands: commands})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiURL("/api/v1/generate_runbookv2"), bytes.NewReader(bs))
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

func (c *client) Runbooks(ctx context.Context) ([]RunbookInfo, error) {
	cl := c.cl
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.apiURL("/api/v1/runbooks/"), nil)
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

type QuestionInfo struct {
	Question          string            `json:"question"`
	Tags              map[string]string `json:"tags,omitempty"`
	FileData          []byte            `json:"file_data,omitempty"`
	FileName          string            `json:"file_name,omitempty"`
	PreviousQuestions []string          `json:"previous_questions,omitempty"`
	PreviousCommands  []string          `json:"previous_commands,omitempty"`
}

func (c *client) Ask(ctx context.Context, question QuestionInfo) (*Runbook, error) {
	return ask(ctx, c.cl, c.apiURL("/api/v1/public/ask"), question)
}

func ask(ctx context.Context, cl *http.Client, apiURL string, question QuestionInfo) (*Runbook, error) {
	bs, err := json.Marshal(question)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(bs))
	if err != nil {
		return nil, err
	}

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

type CodeInfo struct {
	Code     string            `json:"code"`
	Tags     map[string]string `json:"tags,omitempty"`
	FileData []byte            `json:"file_data,omitempty"`
	FileName string            `json:"file_name,omitempty"`
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

func (c *client) Explain(ctx context.Context, code CodeInfo) (<-chan string, error) {
	return explain(ctx, c.cl, c.apiURL("/api/v1/public/explain"), code)
}

func explain(ctx context.Context, cl *http.Client, apiURL string, code CodeInfo) (<-chan string, error) {
	bs, err := json.Marshal(code)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(bs))
	if err != nil {
		return nil, err
	}

	// explain streams the response body.
	stream, err := cl.Do(req)
	if err != nil {
		return nil, err
	}

	if stream.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to explain code: %s", stream.Status)
	}

	resultChan := make(chan string, 1024)
	// Read and print the streamed responses
	scanner := bufio.NewScanner(stream.Body)

	go func(scanner *bufio.Scanner) {
		defer stream.Body.Close()
		defer close(resultChan)

		for scanner.Scan() {
			var data streamData
			line := scanner.Text()
			if len(line) > 6 && line[:6] == "data: " {
				if err := json.Unmarshal([]byte(line[6:]), &data); err != nil {
					// TODO: add debug log stmt here.
					continue
				}
				resultChan <- data.Data
			}
		}

		if err := scanner.Err(); err != nil {
			err = fmt.Errorf("error reading stream: %w", err)
			// display the error message to the user
			resultChan <- err.Error()
			return
		}
	}(scanner)

	return resultChan, nil
}

type streamData struct {
	Data string `json:"data"`
}
