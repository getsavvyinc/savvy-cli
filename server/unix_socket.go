package server

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"sync"
	"sync/atomic"
)

type UnixSocketServer struct {
	socketPath string
	listener   net.Listener

	mu       sync.Mutex
	commands []string
	ch       chan bool

	closed atomic.Bool
}

var ErrStartingRecordingSession = errors.New("failed to start recording session")

const defaultSocketPath = "/tmp/savvy-socket"

func NewUnixSocketServerWithDefaultPath(opts ...Option) (*UnixSocketServer, error) {
	return NewUnixSocketServer(defaultSocketPath, opts...)
}

type Option func(*UnixSocketServer)

// NotifyOnCommandProcessed returns an Option that sets the channel to notify when a command is processed.
// NOTE: It is the callers responsibility to close the channel.
func NotifyOnCommandProcessed(ch chan bool) Option {
	return func(s *UnixSocketServer) {
		s.ch = ch
	}
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
		socketPath: socketPath,
		listener:   listener,
	}

	for _, opt := range opts {
		opt(srv)
	}
	return srv, nil
}

func (s *UnixSocketServer) Commands() []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.commands
}

func (s *UnixSocketServer) Close() error {
	if !s.closed.Load() {
		defer s.closed.Store(true)
		if s.ch != nil {
			close(s.ch)
		}
		if s.listener != nil {
			return s.listener.Close()
		}
	}
	// if already closed, return nil
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
			fmt.Printf("Failed to accept connection: %s\n", err)
			continue
		}

		// Handle the connection
		go s.handleConnection(conn)
	}
}

func (s *UnixSocketServer) handleConnection(c net.Conn) {
	defer c.Close()

	slog.Debug("starting to read from connection")
	bs, err := io.ReadAll(c)
	if err != nil {
		fmt.Printf("Failed to read from connection: %s\n", err)
		return
	}
	slog.Debug("read from connection")
	command := string(bs)
	s.appendCommand(command)
	s.notify()
}

func (s *UnixSocketServer) SocketPath() string {
	return s.socketPath
}

func (s *UnixSocketServer) appendCommand(command string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.commands = append(s.commands, command)
}

func (s *UnixSocketServer) notify() {
	if s.closed.Load() {
		slog.Debug("cannot notify waiting channel", "reason", "server is closed")
		return
	}

	if s.ch == nil {
		return
	}

	slog.Debug("notifying waiting channel")
	s.ch <- true
}
