package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
)

type UnixSocketServer struct {
	socketPath   string
	logger       *slog.Logger
	listener     net.Listener
	ignoreErrors bool

	mu                  sync.Mutex
	commands            []*RecordedData
	lookupCommand       map[string]*RecordedData
	commandRecordedHook func(string)

	closed atomic.Bool
}

var ErrStartingRecordingSession = errors.New("failed to start recording session")

const defaultSocketPath = "/tmp/savvy-socket"

type Option func(*UnixSocketServer)

func WithCommandRecordedHook(hook func(string)) Option {
	return func(s *UnixSocketServer) {
		s.commandRecordedHook = hook
	}
}

func WithLogger(logger *slog.Logger) Option {
	return func(s *UnixSocketServer) {
		s.logger = logger
	}
}

func WithIgnoreErrors(ignoreErrors bool) Option {
	return func(s *UnixSocketServer) {
		s.ignoreErrors = ignoreErrors
	}
}

var defaultLogger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

func NewUnixSocketServerWithDefaultPath(opts ...Option) (*UnixSocketServer, error) {
	return NewUnixSocketServer(defaultSocketPath, opts...)
}

func NewUnixSocketServer(socketPath string, opts ...Option) (*UnixSocketServer, error) {
	if fileInfo, _ := os.Stat(socketPath); fileInfo != nil {
		return nil, fmt.Errorf("%w: concurrent recording sessions are not supported yet", ErrStartingRecordingSession)
	}
	return newUnixSocketServer(socketPath, opts...)
}

func newUnixSocketServer(socketPath string, opts ...Option) (*UnixSocketServer, error) {
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create listener: %w", err)
	}

	srv := &UnixSocketServer{
		socketPath:    socketPath,
		listener:      listener,
		logger:        defaultLogger,
		ignoreErrors:  false,
		lookupCommand: make(map[string]*RecordedData),
	}

	for _, opt := range opts {
		opt(srv)
	}

	return srv, nil
}

func (s *UnixSocketServer) Commands() []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	var commands []string

	for _, cmd := range s.commands {
		if s.ignoreErrors && cmd.ExitCode != 0 {
			continue
		}
		commands = append(commands, cmd.Command)
	}
	return commands
}

func (s *UnixSocketServer) Close() error {
	if s.listener != nil {
		s.closed.Store(true)
		return s.listener.Close()
	}
	return nil
}

func (s *UnixSocketServer) ListenAndServe() {
	for {
		// Accept new connections
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			slog.Debug("Failed to accept connection:", "error", err.Error())
			continue
		}

		// Handle the connection
		go s.handleConnection(conn)
	}
}

type RecordedData struct {
	Command  string `json:"command"`
	StepID   string `json:"step_id"`
	ExitCode int    `json:"exit_code"`
}

func (s *UnixSocketServer) handleConnection(c net.Conn) {
	defer c.Close()

	bs, err := io.ReadAll(c)
	if err != nil {
		fmt.Printf("Failed to read from connection: %s\n", err)
		return
	}

	var data RecordedData
	if err := json.Unmarshal(bs, &data); err != nil {
		s.logger.Debug("Failed to unmarshal data", "error", err.Error(), "component", "server", "input", string(bs))
		return
	}

	if s.maybeAppendData(data) && s.commandRecordedHook != nil {
		s.commandRecordedHook(data.Command)
	}

	if data.StepID != "" && data.ExitCode != 0 {
		s.logger.Debug("command failed", "command", data.Command, "exit_status", data.ExitCode)
		s.updateExitStatus(data.StepID, data.ExitCode)
	}
}

func (s *UnixSocketServer) SocketPath() string {
	return s.socketPath
}

func (s *UnixSocketServer) updateExitStatus(stepID string, exitCode int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if cmd, ok := s.lookupCommand[stepID]; ok {
		cmd.ExitCode = exitCode
	}
}

func (s *UnixSocketServer) maybeAppendData(data RecordedData) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	cmd := strings.TrimSpace(data.Command)

	if cmd == "" {
		return false
	}

	if _, ok := s.lookupCommand[data.StepID]; ok {
		return false
	}

	s.commands = append(s.commands, &data)
	s.lookupCommand[data.StepID] = &data
	s.logger.Debug("command recorded", "command", data.Command)
	return true
}
