package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/getsavvyinc/savvy-cli/config"
)

var _ Client = (*guest)(nil)

func NewGuest() Client {
	return &guest{
		cl: &http.Client{
			Transport: &GuestRoundTripper{savvyVersion: config.Version()},
		},
		apiHost: config.APIHost(),
	}
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

var ErrCannotUseGuest = errors.New("cannot use guest client for this operation")

func (g *guest) GenerateRunbookV2(ctx context.Context, commands []RecordedCommand) (*GeneratedRunbook, error) {
	return nil, fmt.Errorf("%w: %v", ErrCannotUseGuest, "generate runbook")
}

func (g *guest) GenerateRunbook(ctx context.Context, commands []string) (*GeneratedRunbook, error) {
	return nil, fmt.Errorf("%w: %v", ErrCannotUseGuest, "generate runbook")
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

func (g *guest) Runbooks(ctx context.Context) ([]RunbookInfo, error) {
	return nil, fmt.Errorf("%w: %v", ErrCannotUseGuest, "list runbooks")
}
