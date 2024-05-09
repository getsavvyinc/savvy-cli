package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"path/filepath"

	"golang.org/x/sync/errgroup"

	"github.com/getsavvyinc/savvy-cli/ingest/parser"
)

const tldrPath = "tldr/pages/"

const maxConcurrency = 500

func main() {
	logger := slog.Default()
	cheatsheetParser := parser.New(parser.TLDR)

	var g errgroup.Group
	g.SetLimit(maxConcurrency)

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
}
