package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/getsavvyinc/savvy-cli/config"
	"github.com/getsavvyinc/savvy-cli/extension"
	"github.com/getsavvyinc/savvy-cli/llm/service"
	"github.com/getsavvyinc/savvy-cli/login"
	"github.com/getsavvyinc/savvy-cli/model"
)

var _ Client = (*guest)(nil)

func NewGuest() (Client, error) {
	cfg, err := config.LoadFromFile()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidClient, err)
	}

	cl := &http.Client{
		Transport: &GuestRoundTripper{savvyVersion: config.Version()},
	}

	return &guest{
		cl:      cl,
		llmSvc:  service.New(cfg, cl),
		apiHost: config.APIHost(),
	}, nil
}

type GuestRoundTripper struct {
	savvyVersion string
}

type ErrorResponse struct {
	Message string `json:"message"`
}

func (g *GuestRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to ensure thread safety
	clonedReq := req.Clone(req.Context())
	clonedReq.Header.Set("X-Savvy-Version", g.savvyVersion)
	clonedReq.Header.Set("X-Savvy-Guest", "true")

	// Use the embedded Transport to perform the actual request
	return http.DefaultTransport.RoundTrip(clonedReq)
}

type guest struct {
	cl      *http.Client
	llmSvc  service.Service
	apiHost string
}

// apiURL returns the full url to the api endpoint
// path must start with a slash. e.g. /api/v1/whoami
// apiURL will add a slash if it's missing
func (g *guest) apiURL(path string) string {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return g.apiHost + path
}

func (g *guest) WhoAmI(ctx context.Context) (string, error) {
	return "Savvy Guest", nil
}

func (g *guest) GenerateRunbookV2(ctx context.Context, commands []model.RecordedCommand, links []extension.HistoryItem) (*GeneratedRunbook, error) {
	cl, err := getLoggedInClient()
	if err != nil {
		return nil, err
	}
	return cl.GenerateRunbookV2(ctx, commands, links)
}

func (g *guest) GenerateRunbook(ctx context.Context, commands []string) (*GeneratedRunbook, error) {
	cl, err := getLoggedInClient()
	if err != nil {
		return nil, err
	}
	return cl.GenerateRunbook(ctx, commands)
}

func (g *guest) RunbookByID(ctx context.Context, id string) (*Runbook, error) {
	cl := g.cl
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, g.apiURL("/api/v1/public/runbook"), nil)
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

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		cl, err := getLoggedInClient()
		if err != nil {
			return nil, fmt.Errorf("not authorized to view this runbook: %w", err)
		}

		return cl.RunbookByID(ctx, id)
	}

	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("could not parse error message: %w", err)
		}
		return nil, fmt.Errorf("error getting runbook: %s", errResp.Message)
	}

	var runbook Runbook
	if err := json.NewDecoder(resp.Body).Decode(&runbook); err != nil {
		return nil, err
	}
	return &runbook, nil
}

func (g *guest) Runbooks(ctx context.Context, opt RunbooksOpt) ([]RunbookInfo, error) {
	cl, err := getLoggedInClient()
	if err != nil {
		return nil, err
	}
	return cl.Runbooks(ctx, opt)
}

func (g *guest) Ask(ctx context.Context, question *model.QuestionInfo) (*Runbook, error) {
	answer, err := g.llmSvc.Ask(ctx, question)
	if err != nil {
		return nil, err
	}
	return toClientRunbook(answer), nil
}

func (g *guest) Explain(ctx context.Context, code *model.CodeInfo) (<-chan string, error) {
	return g.llmSvc.Explain(ctx, code)
}

func (g *guest) StepContentByStepID(ctx context.Context, stepID string) (*StepContent, error) {
	cl, err := getLoggedInClient()
	if err != nil {
		return nil, err
	}
	return cl.StepContentByStepID(ctx, stepID)
}

func (g *guest) SaveRunbook(ctx context.Context, runbook *Runbook) (*GeneratedRunbook, error) {
	cl, err := getLoggedInClient()
	if err != nil {
		return nil, err
	}
	return cl.SaveRunbook(ctx, runbook)
}

func GetLoggedInClient() (Client, error) {
	return getLoggedInClient()
}

func getLoggedInClient() (Client, error) {
	cl, err := New()
	if err == nil {
		return cl, nil
	}

	login.Run(VerifyLogin)
	cl, err = New()
	if err != nil {
		return nil, err
	}
	return cl, nil
}
