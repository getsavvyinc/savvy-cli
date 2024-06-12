package run

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"sync"
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

	mu        sync.Mutex
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
	return newRunServer(DefaultRunSocketPath, rb, opts...)
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
		go rs.handleConnection(conn)
	}
}

func (rs *RunServer) handleConnection(c net.Conn) {
	defer c.Close()

	bs, err := io.ReadAll(c)
	if err != nil {
		rs.logger.Error("failed to read from connection: %s\n", err)
		return
	}

	var data RunCommand
	if err := json.Unmarshal(bs, &data); err != nil {
		rs.logger.Error("failed to unmarshal data", "error", err.Error(), "component", "run", "input", string(bs))
		return
	}

	if data.IsShutdown() {
		rs.Close()
		return
	}

	cmd := data.Command
	rs.handleCommand(cmd, c)
}

type RunCommandIndexResponse struct {
	Index int `json:"index"`
}

func (rs *RunServer) handleCommand(cmd string, c net.Conn) {
	switch cmd {
	case shutdownCommand:
		rs.Close()
	case nextCommand:
		rs.mu.Lock()
		rs.currIndex++
		if rs.currIndex >= len(rs.commands) {
			rs.currIndex = len(rs.commands) - 1
		}
		rs.mu.Unlock()
	case previousCommand:
		rs.mu.Lock()
		rs.currIndex--
		if rs.currIndex < 0 {
			rs.currIndex = 0
		}
		rs.mu.Unlock()
	case currentCommand:
		var response State
		rs.mu.Lock()
		if rs.currIndex >= len(rs.commands) {
			rs.currIndex = len(rs.commands) - 1
		}
		if rs.currIndex < 0 {
			rs.currIndex = 0
		}
		cmd := rs.commands[rs.currIndex]
		response.Command = cmd.Command
		response.Index = rs.currIndex
		rs.mu.Unlock()
		rs.logger.Debug("fetching command", "command", cmd)
		json.NewEncoder(c).Encode(response)
	default:
		rs.logger.Debug("unknown command", "command", cmd)
	}
}

func (rs *RunServer) SocketPath() string {
	return rs.socketPath
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
