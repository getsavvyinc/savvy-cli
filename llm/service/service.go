package service

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/getsavvyinc/savvy-cli/config"
	"github.com/getsavvyinc/savvy-cli/llm"
	"github.com/getsavvyinc/savvy-cli/model"
)

type Service interface {
	GenerateRunbook(ctx context.Context, commands []model.RecordedCommand) (*llm.Runbook, error)
	Ask(ctx context.Context, question *model.QuestionInfo) (*llm.Runbook, error)
}

func New(cfg *config.Config, savvyClient *http.Client) Service {
	if cfg.LLMBaseURL == "" {
		return newDefaultService(cfg, savvyClient)
	}
	return newCustomService(cfg)
}

type defaultLLM struct {
	cl      *http.Client
	apiHost string
}

func newDefaultService(cfg *config.Config, savvyClient *http.Client) Service {
	return &defaultLLM{
		cl:      savvyClient,
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

func (d *defaultLLM) Ask(ctx context.Context, question *model.QuestionInfo) (*llm.Runbook, error) {
	apiURL := genAPIURL(d.apiHost, "/api/v1/public/ask")
	return ask(ctx, d.cl, apiURL, question)
}

func ask(ctx context.Context, cl *http.Client, apiURL string, question *model.QuestionInfo) (*llm.Runbook, error) {
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

	var runbook llm.Runbook
	if err := json.NewDecoder(resp.Body).Decode(&runbook); err != nil {
		return nil, err
	}
	return &runbook, nil
}
