package run

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"os"
	"sync/atomic"

	"github.com/getsavvyinc/savvy-cli/server"
	"github.com/getsavvyinc/savvy-cli/server/cleanup"
	"github.com/getsavvyinc/savvy-cli/server/mode"
)

type RunServer struct {
	socketPath string
	logger     *slog.Logger
	listener   net.Listener
	commands   []*RunCommand

	closed atomic.Bool
}

type RunCommand struct {
	Command string `json:"command,omitempty"`
}

const DefaultRunSocketPath = "/tmp/savvy-run-socket"

var ErrStartingRunSession = errors.New("failed to start run session")

type Option func(s *RunServer)

func WithLogger(logger *slog.Logger) Option {
	return func(s *RunServer) {
		s.logger = logger
	}
}

// cleanupSocket is an internal function.
// It is the callers responsibility to ensure the socketPath exists.
func cleanupSocket(socketPath string) error {
	cl, err := server.NewClient(context.Background(), socketPath)
	if err != nil {
		return err
	}
	cl.SendShutdown()

	if err := os.Remove(socketPath); err != nil {
		return err
	}
	return nil
}

func NewServerWithDefaultSocketPath(commands []*RunCommand, opts ...Option) (*RunServer, error) {
	return newRunServer(DefaultRunSocketPath, commands, opts...)
}

var defaultLogger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

func newRunServer(socketPath string, commands []*RunCommand, opts ...Option) (*RunServer, error) {
	if fileInfo, _ := os.Stat(socketPath); fileInfo != nil {

		cleanup, cerr := cleanup.GetPermission(mode.Run)
		if cerr != nil {
			return nil, cerr
		}

		if !cleanup {
			return nil, server.ErrAbortRecording
		}

		cleanupSocket(socketPath)
	}

	rs := &RunServer{
		socketPath: socketPath,
		logger:     defaultLogger,
		commands:   commands,
	}

	for _, opt := range opts {
		opt(rs)
	}
	return rs, nil
}
