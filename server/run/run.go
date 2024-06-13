package run

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"sync/atomic"

	savvy_client "github.com/getsavvyinc/savvy-cli/client"
	"github.com/getsavvyinc/savvy-cli/server/cleanup"
	"github.com/getsavvyinc/savvy-cli/server/mode"
	"github.com/getsavvyinc/savvy-cli/slice"
)

type RunServer struct {
	socketPath string
	logger     *slog.Logger
	listener   net.Listener

	currIndex int
	commands  []*RunCommand

	closed atomic.Bool
}

type RunCommand struct {
	Command string `json:"command,omitempty"`
}

type State struct {
	Command string `json:"command"`
	Index   int    `json:"index"`
}

const DefaultRunSocketPath = "/tmp/savvy-run.sock"

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
	cl, err := NewClient(context.Background(), socketPath)
	if err != nil {
		return err
	}
	cl.SendShutdown()

	if err := os.Remove(socketPath); err != nil {
		return err
	}
	return nil
}

func NewServerWithDefaultSocketPath(rb *savvy_client.Runbook, opts ...Option) (*RunServer, error) {
	return NewServerWithSocketPath(DefaultRunSocketPath, rb, opts...)
}

func NewServerWithSocketPath(socketPath string, rb *savvy_client.Runbook, opts ...Option) (*RunServer, error) {
	return newRunServer(socketPath, rb, opts...)
}

var defaultLogger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

var ErrAbortRun = errors.New("abort running runbook")

func newRunServer(socketPath string, rb *savvy_client.Runbook, opts ...Option) (*RunServer, error) {
	if fileInfo, _ := os.Stat(socketPath); fileInfo != nil {

		cleanupOK, cerr := cleanup.GetPermission(mode.Run)
		if cerr != nil {
			return nil, cerr
		}

		if !cleanupOK {
			return nil, ErrAbortRun
		}

		cleanupSocket(socketPath)
	}
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create listener: %w", err)
	}

	steps := rb.Steps

	cmds := slice.Map(steps, func(step savvy_client.Step) *RunCommand {
		return &RunCommand{
			Command: step.Command,
		}
	})

	rs := &RunServer{
		socketPath: socketPath,
		logger:     defaultLogger,
		commands:   cmds,
		listener:   listener,
	}

	for _, opt := range opts {
		opt(rs)
	}
	return rs, nil
}

func (rs *RunServer) Close() error {
	if rs.closed.Load() {
		return nil
	}
	rs.closed.Store(true)
	return rs.listener.Close()
}

func (rs *RunServer) ListenAndServe() {
	for {
		// Accept new connections
		conn, err := rs.listener.Accept()
		if err != nil {
			if rs.closed.Load() {
				return
			}
			slog.Debug("Failed to accept connection:", "error", err.Error())
			continue
		}

		// Handle the connection
		// Intentionally single threaded
		rs.handleConnection(conn)
	}
}

func (rs *RunServer) handleConnection(c net.Conn) {
	defer c.Close()

	var data RunCommand

	if err := json.NewDecoder(c).Decode(&data); err != nil {
		rs.logger.Error("failed to unmarshal data", "error", err.Error())
		return
	}

	cmd := data.Command
	rs.handleCommand(cmd, c)
}

func (rs *RunServer) handleCommand(cmd string, c net.Conn) {
	switch cmd {
	case shutdownCommand:
		rs.Close()
	case nextCommand:
		rs.currIndex += 1
		// NOTE: we intentionally allow currIndex to = len(rs.commands) that's how we know we're done
		if rs.currIndex > len(rs.commands) {
			rs.currIndex = len(rs.commands)
		}
	case currentCommand:
		response := State{
			Index: rs.currIndex,
		}
		if rs.currIndex < len(rs.commands) {
			cmd := rs.commands[rs.currIndex]
			response.Command = cmd.Command
			rs.logger.Debug("fetching command", "command", cmd)
		}
		json.NewEncoder(c).Encode(response)
	default:
		rs.logger.Debug("unknown command", "command", cmd)
	}
}

func (rs *RunServer) SocketPath() string {
	return rs.socketPath
}

func (rs *RunServer) Commands() []*RunCommand {
	return rs.commands
}

const (
	shutdownCommand = "savvy shutdown"
	nextCommand     = "savvy internal next"
	previousCommand = "savvy internal previous"
	currentCommand  = "savvy internal current"
)

func (rc *RunCommand) IsShutdown() bool {
	return rc.Command == shutdownCommand
}
