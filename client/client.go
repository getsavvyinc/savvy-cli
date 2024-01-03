package client

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/getsavvyinc/savvy-cli/config"
)

type Client interface {
	WhoAmI(ctx context.Context) (string, error)
	GenerateRunbook(ctx context.Context, commands []string) (*GeneratedRunbook, error)
}

type client struct {
	cl      *http.Client
	apiHost string
}

var _ Client = (*client)(nil)

func New() (Client, error) {
	cfg, err := config.LoadFromFile()
	if err != nil {
		return nil, err
	}

	c := &client{
		cl: &http.Client{
			Transport: &AuthorizedRoundTripper{
				token: cfg.Token,
			},
		},
		apiHost: config.APIHost(),
	}
	return c, nil
}

type AuthorizedRoundTripper struct {
	token string
}

func (a *AuthorizedRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to ensure thread safety
	clonedReq := req.Clone(req.Context())
	clonedReq.Header.Set("Authorization", "Bearer "+a.token)

	// Use the embedded Transport to perform the actual request
	return http.DefaultTransport.RoundTrip(clonedReq)
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
	// TODO: remove hardcoded url
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
	Title string `json:"title"`
	Steps []Step `json:"steps"`
}

type Step struct {
	Description string `json:"description"`
	Command     string `json:"command"`
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
