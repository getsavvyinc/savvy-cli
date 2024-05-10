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
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/getsavvyinc/savvy-cli/ingest/cheatsheet"
	"github.com/getsavvyinc/savvy-cli/ingest/cheatsheet/parser"
	"github.com/getsavvyinc/savvy-cli/ingest/db"
	"github.com/getsavvyinc/savvy-cli/ingest/llm"
)

const tldrPath = "tldr/pages/"

const maxConcurrency = 500
const maxDBConcurrency = 50

func main() {
	logger := slog.Default()
	ctx := context.Background()

	authToken := os.Getenv("OPENAI_API_KEY")
	if authToken == "" {
		logger.Error("OPENAI_API_KEY is required")
		return
	}

	db, err := db.NewDB()
	if err != nil {
		logger.Error(err.Error())
		return
	}
	defer db.Close()

	llmClient := llm.NewOpenAIClient(authToken)
	cheatsheetParser := parser.New(cheatsheet.TLDR)

	var g errgroup.Group
	g.SetLimit(maxConcurrency)

	var mu sync.Mutex
	var cheatsheets []*cheatsheet.CheatSheet

	err = filepath.Walk(tldrPath, func(path string, info fs.FileInfo, err error) error {
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

	var dbGroup errgroup.Group
	dbGroup.SetLimit(maxDBConcurrency)

	items := 0
	lastIndex := 0
	for _, cs := range cheatsheets {
		if items-lastIndex >= 100 && items != 0 {
			// wait a bit to avoid hitting the open ai rate limit
			logger.Info("sleeping for two seconds", "cheatsheet", items)
			time.Sleep(10 * time.Second)
			lastIndex = items
		}

		prefix := cs.CommonEmbeddingPrefix()
		for _, example := range cs.Examples {
			example := example
			prefix := prefix
			dbGroup.Go(processExample(ctx, llmClient, logger, db, example, prefix))
		}
		items += len(cs.Examples)
	}

	if err := dbGroup.Wait(); err != nil {
		logger.Error(err.Error(), "component", "errgroup")
	}

	logger.Info("done", "items", items)
}

func processExample(ctx context.Context, llmClient llm.Client, logger *slog.Logger, db *db.DB, example *cheatsheet.Example, prefix string) func() error {
	return func() error {
		command := example.Command
		explanation := example.Explanation
		explanationEmbedding, err := llmClient.CreateEmbeddings(ctx, strings.Join([]string{prefix, explanation}, " "))
		if err != nil {
			err = fmt.Errorf("failed to create embeddings for explanation:%s %w", explanation, err)
			logger.Error(err.Error())
			return err
		}

		commandEmbedding, err := llmClient.CreateEmbeddings(ctx, strings.Join([]string{command}, " "))
		if err != nil {
			err = fmt.Errorf("failed to create embeddings for command:%s %w", command, err)
			logger.Error(err.Error())
			return err
		}

		if err := db.StoreExampleEmbedding(ctx, example, explanationEmbedding, commandEmbedding); err != nil {
			err = fmt.Errorf("failed to store example embedding: %w", err)
			logger.Error(err.Error())
			return err
		}

		return nil
	}
}
