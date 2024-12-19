package client

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/getsavvyinc/savvy-cli/config"
)

type LLMClient interface {
	GenerateRunbook(ctx context.Context, commands []RecordedCommand) (*Runbook, error)
}

func NewLLMClient(cfg *config.Config) LLMClient {
	if cfg.OpenAIBaseURL == "" {
		return &defaultLLMClient{
			apiHost: config.APIHost(),
			cl: &http.Client{
				Transport: &AuthorizedRoundTripper{
					token:        cfg.Token,
					savvyVersion: config.Version(),
				},
			},
		}
	}
	return &customLLMClient{}
}

type defaultLLMClient struct {
	cl      *http.Client
	apiHost string
}

func (dlc *defaultLLMClient) GenerateRunbook(ctx context.Context, commands []RecordedCommand) (*Runbook, error) {
	cl := dlc.cl
	bs, err := json.Marshal(struct{ Commands []RecordedCommand }{Commands: commands})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, genAPIURL(dlc.apiHost, "/api/v1/generate_workflow"), bytes.NewReader(bs))
	if err != nil {
		return nil, err
	}
	resp, err := cl.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var generatedRunbook Runbook
	if err := json.NewDecoder(resp.Body).Decode(&generatedRunbook); err != nil {
		return nil, err
	}
	return &generatedRunbook, nil
}

type customLLMClient struct{}

func (c *customLLMClient) GenerateRunbook(ctx context.Context, commands []RecordedCommand) (*Runbook, error) {
	return nil, nil
}
