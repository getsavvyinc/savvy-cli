package service

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/getsavvyinc/savvy-cli/authz"
	"github.com/getsavvyinc/savvy-cli/config"
	"github.com/getsavvyinc/savvy-cli/llm"
)

type Service interface {
	GenerateRunbook(ctx context.Context, commands []llm.RecordedCommand) (*Runbook, error)
}

type Runbook struct {
	Title string
	Steps []RunbookStep
}

type RunbookStep struct {
	Code        string `json:"code,omitempty"`
	Description string `json:"description,omitempty"`
	CodeID      string `json:"code_id,omitempty"`
}

func New(cfg *config.Config) Service {
	if cfg.OpenAIBaseURL == "" {
		return newDefaultService(cfg)
	}
	return &service{}
}

type defaultLLM struct {
	cl      *http.Client
	apiHost string
}

func newDefaultService(cfg *config.Config) Service {
	return &defaultLLM{
		cl: &http.Client{
			Transport: authz.NewRoundTripper(cfg.Token, config.Version()),
		},
		apiHost: config.APIHost(),
	}
}

func (d *defaultLLM) GenerateRunbook(ctx context.Context, commands []llm.RecordedCommand) (*Runbook, error) {
	cl := d.cl
	bs, err := json.Marshal(struct{ Commands []llm.RecordedCommand }{Commands: commands})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, genAPIURL(d.apiHost, "/api/v1/generate_workflow"), bytes.NewReader(bs))
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

func genAPIURL(host, path string) string {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return host + path
}

type service struct{}

func (s *service) GenerateRunbook(ctx context.Context, commands []llm.RecordedCommand) (*Runbook, error) {
	return nil, nil
}
