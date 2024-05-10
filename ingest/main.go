package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/getsavvyinc/savvy-cli/ingest/cheatsheet"
	"github.com/getsavvyinc/savvy-cli/ingest/cheatsheet/parser"
	"github.com/getsavvyinc/savvy-cli/ingest/llm"
)

const tldrPath = "tldr/pages/"

const maxConcurrency = 500

func main() {
	logger := slog.Default()
	ctx := context.Background()

	authToken := os.Getenv("OPENAI_API_KEY")
	if authToken == "" {
		logger.Error("OPENAI_API_KEY is required")
		return
	}

	llmClient := llm.NewOpenAIClient(authToken)
	cheatsheetParser := parser.New(cheatsheet.TLDR)

	var g errgroup.Group
	g.SetLimit(maxConcurrency)

	var mu sync.Mutex
	var cheatsheets []*cheatsheet.CheatSheet

	err := filepath.Walk(tldrPath, func(path string, info fs.FileInfo, err error) error {
		logger := logger.With("path", path)
		if err != nil {
			logger.Error(err.Error())
			return nil
		}

		if info.IsDir() {
			logger.Info("skipping directory")
			return nil
		}

		g.Go(func() error {
			cheatSheet, err := cheatsheetParser.Parse(path)
			if err != nil && !errors.Is(err, parser.ErrRequiredMdFile) {
				err = fmt.Errorf("failed to parse file: %w", err)
				logger.Error(err.Error(), "provider", cheatsheetParser.Provider())
				return err
			}
			if cheatSheet == nil {
				logger.Warn("cheat sheet is nil")
				return fmt.Errorf("cheat sheet is nil: %s", path)
			}
			mu.Lock()
			cheatsheets = append(cheatsheets, cheatSheet)
			mu.Unlock()
			return nil
		})

		return nil
	})

	if err != nil {
		logger.Error(err.Error())
	}

	if err := g.Wait(); err != nil {
		logger.Error(err.Error(), "component", "errgroup")
	}

	for i, cs := range cheatsheets {
		if i > 10 {
			logger.Info("breaking early")
			return
		}

		prefix := cs.CommonEmbeddingPrefix()
		for _, example := range cs.Examples {
			command := example.Command
			explanation := example.Explanation
			if embedding, err := llmClient.CreateEmbeddings(ctx, strings.Join([]string{prefix, explanation}, " ")); err != nil {
				logger.Error(err.Error(), "component", "llmClient.embed.explanation")
				continue
			} else {
				logger.Info("embedding created", "embedding", embedding[:10])
			}

			if embedding, err := llmClient.CreateEmbeddings(ctx, strings.Join([]string{command}, " ")); err != nil {
				logger.Error(err.Error(), "component", "llmClient.embed.command")
				continue
			} else {
				logger.Info("embedding created", "embedding", embedding[:10])
			}
		}
	}
}
