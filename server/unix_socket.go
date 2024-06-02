package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/getsavvyinc/savvy-cli/idgen"
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

type ErrConcurrentRecordingSession struct {
	Path string
}

func (e *ErrConcurrentRecordingSession) Error() string {
	return fmt.Sprintf("%v: concurrent recording session not supported", ErrStartingRecordingSession)
}

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

func RemoveSocket() Option {
	return func(s *UnixSocketServer) {
		if err := os.Remove(s.socketPath); err != nil {
			s.logger.Debug("failed to remove socket file", "error", err.Error())
		}
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
	return newUnixSocketServer(socketPath, opts...)
}

func newUnixSocketServer(socketPath string, opts ...Option) (*UnixSocketServer, error) {
	srv := &UnixSocketServer{
		socketPath:    socketPath,
		logger:        defaultLogger,
		ignoreErrors:  false,
		lookupCommand: make(map[string]*RecordedData),
	}

	for _, opt := range opts {
		opt(srv)
	}

	if fileInfo, _ := os.Stat(socketPath); fileInfo != nil {
		return nil, &ErrConcurrentRecordingSession{Path: socketPath}
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create listener: %w", err)
	}
	srv.listener = listener

	return srv, nil
}

type RecordedCommand struct {
	Command  string    `json:"command"`
	Prompt   string    `json:"prompt,omitempty"`
	FileInfo *FileInfo `json:"file_info,omitempty"`
}

type FileInfo struct {
	Path    string      `json:"path,omitempty"`
	Mode    fs.FileMode `json:"mode,omitempty"`
	Content []byte      `json:"content,omitempty"`
}

func (s *UnixSocketServer) Commands() []*RecordedCommand {
	s.mu.Lock()
	defer s.mu.Unlock()

	var commands []*RecordedCommand

	for _, cmd := range s.commands {
		if s.ignoreErrors && cmd.ExitCode != 0 {
			continue
		}

		if cmd.HasFileData() {
			recordedFile := &RecordedCommand{
				Command: cmd.Command,
				FileInfo: &FileInfo{
					Path:    cmd.Filepath,
					Mode:    cmd.FileMode,
					Content: cmd.FileData,
				},
			}
			commands = append(commands, recordedFile)
			continue
		}

		commands = append(commands, &RecordedCommand{
			Command: cmd.Command,
			Prompt:  cmd.Prompt,
		})
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
	Command  string      `json:"command"`
	StepID   string      `json:"step_id"`
	ExitCode int         `json:"exit_code"`
	Prompt   string      `json:"prompt,omitempty"`
	Filepath string      `json:"filepath,omitempty"`
	FileData []byte      `json:"file_data,omitempty"`
	FileMode fs.FileMode `json:"file_mode,omitempty"`
}

func (rd *RecordedData) HasFileData() bool {
	return strings.HasPrefix(rd.StepID, idgen.FilePrefix)
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

	if data.HasFileData() {
		s.recordFile(data)
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

	// do not record ignore savvy record file commands
	if strings.HasPrefix(cmd, "savvy record file") {
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

func (s *UnixSocketServer) recordFile(data RecordedData) {
	filePath := data.Filepath

	if err := checkFile(filePath); err != nil {
		s.logger.Debug("file checks failed", "error", err.Error())
		return
	}

	fd, err := os.Open(filePath)
	if err != nil {
		s.logger.Debug("failed to open file", "error", err.Error())
		return
	}
	defer fd.Close()

	bs, err := io.ReadAll(fd)
	if err != nil {
		s.logger.Debug("failed to read file", "error", err.Error())
		return
	}

	fi, err := fd.Stat()
	if err != nil {
		s.logger.Debug("failed to get file info", "error", err.Error())
		return
	}

	data.FileData = bs
	data.Filepath = filePath
	data.FileMode = fi.Mode()

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.lookupCommand[data.StepID]; ok {
		return
	}

	s.commands = append(s.commands, &data)
	s.lookupCommand[data.StepID] = &data
	s.logger.Debug("file recorded", "file", data.Filepath)
}
