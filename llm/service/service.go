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
	"github.com/getsavvyinc/savvy-cli/model"
)

type Service interface {
	GenerateRunbook(ctx context.Context, commands []model.RecordedCommand) (*llm.Runbook, error)
}

func New(cfg *config.Config) Service {
	if cfg.LLMBaseURL == "" {
		return newDefaultService(cfg)
	}
	return newCustomService(cfg)
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

func (d *defaultLLM) GenerateRunbook(ctx context.Context, commands []model.RecordedCommand) (*llm.Runbook, error) {
	cl := d.cl
	bs, err := json.Marshal(struct{ Commands []model.RecordedCommand }{Commands: commands})
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

	var generatedRunbook llm.Runbook
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

func (s *service) GenerateRunbook(ctx context.Context, commands []model.RecordedCommand) (*llm.Runbook, error) {
	return nil, nil
}
