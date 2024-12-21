package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/getsavvyinc/savvy-cli/config"
	"github.com/getsavvyinc/savvy-cli/llm"
	"github.com/getsavvyinc/savvy-cli/model"
)

type Service interface {
	GenerateRunbook(ctx context.Context, commands []model.RecordedCommand) (*llm.Runbook, error)
	Ask(ctx context.Context, question *model.QuestionInfo) (*llm.Runbook, error)
	Explain(ctx context.Context, code *model.CodeInfo) (<-chan string, error)
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

type streamData struct {
	Data string `json:"data"`
}

func (d *defaultLLM) Explain(ctx context.Context, code *model.CodeInfo) (<-chan string, error) {
	cl := d.cl
	bs, err := json.Marshal(code)
	if err != nil {
		return nil, err
	}

	apiURL := genAPIURL(d.apiHost, "/api/v1/public/explain")
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
