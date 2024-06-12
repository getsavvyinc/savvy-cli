package run

import (
	"context"
	"encoding/json"
	"net"

	"github.com/getsavvyinc/savvy-cli/server"
)

type Client interface {
	server.ShutdownSender
	NextCommand() (int, error)
	PreviousCommand() error
	CurrentCommand() string
}

func NewDefaultClient(ctx context.Context) (Client, error) {
	return NewClient(ctx, DefaultRunSocketPath)
}

type client struct {
	socketPath string
}

var _ Client = &client{}

func NewClient(ctx context.Context, socketPath string) (Client, error) {
	return &client{
		socketPath: socketPath,
	}, nil
}

func (c *client) SendShutdown() error {
	conn, err := net.Dial("unix", c.socketPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	data := RunCommand{
		Command: shutdownCommand,
	}

	return json.NewEncoder(conn).Encode(data)
}

func (c *client) NextCommand() (int, error) {
	conn, err := net.Dial("unix", c.socketPath)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	data := RunCommand{
		Command: nextCommand,
	}

	if err := json.NewEncoder(conn).Encode(data); err != nil {
		return 0, err
	}

	var response RunCommandIndexResponse
	if err := json.NewDecoder(conn).Decode(&response); err != nil {
		return 0, err
	}

	return response.Index, nil
}

func (c *client) PreviousCommand() error {
	conn, err := net.Dial("unix", c.socketPath)
	if err != nil {
		return err
	}

	data := RunCommand{
		Command: previousCommand,
	}

	return json.NewEncoder(conn).Encode(data)
}

func (c *client) CurrentCommand() string {
	conn, err := net.Dial("unix", c.socketPath)
	if err != nil {
		return err.Error()
	}

	data := RunCommand{
		Command: currentCommand,
	}

	if err := json.NewEncoder(conn).Encode(data); err != nil {
		return err.Error()
	}

	var response RunCommand
	if err := json.NewDecoder(conn).Decode(&response); err != nil {
		return err.Error()
	}

	return response.Command
}
