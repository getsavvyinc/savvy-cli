package cmd

import (
	"context"
	"log/slog"
	"os"
)

type cmdLogger struct{}

var cmdLoggerKey cmdLogger

func loggerFromCtx(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(cmdLoggerKey).(*slog.Logger); ok && logger != nil {
		return logger
	}
	return defaultLogger
}

func ctxWithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, cmdLoggerKey, logger)
}

var defaultLogger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
