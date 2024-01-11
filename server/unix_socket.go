package server

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"sync/atomic"
)

const socketPath = "/tmp/savvy-socket"

type UnixSocketServer struct {
	socketPath string
	listener   net.Listener

	mu       sync.Mutex
	commands []string

	closed atomic.Bool
}

var ErrStartingRecordingSession = errors.New("failed to start recording session")

func NewUnixSocketServer(socketPath string) (*UnixSocketServer, error) {
	if fileInfo, _ := os.Stat(socketPath); fileInfo != nil {
		return nil, fmt.Errorf("%w: concurrent recording sessions are not supported yet.", ErrStartingRecordingSession)
	}
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create listener: %w", err)
	}

	return &UnixSocketServer{
		socketPath: socketPath,
		listener:   listener,
	}, nil
}

func (s *UnixSocketServer) Commands() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.commands
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
			fmt.Printf("Failed to accept connection: %s\n", err)
			continue
		}

		// Handle the connection
		go s.handleConnection(conn)
	}
}

func (s *UnixSocketServer) handleConnection(c net.Conn) {
	defer c.Close()
	bs, err := io.ReadAll(c)
	if err != nil {
		fmt.Printf("Failed to read from connection: %s\n", err)
		return
	}
	command := string(bs)
	s.mu.Lock()
	s.commands = append(s.commands, command)
	s.mu.Unlock()
}
