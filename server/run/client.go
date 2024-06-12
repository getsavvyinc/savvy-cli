package run

import (
	"context"
	"encoding/json"
	"net"

	"github.com/getsavvyinc/savvy-cli/server"
)

type Client interface {
	server.ShutdownSender
	NextCommand() error
	PreviousCommand() error
	FetchCommand() string
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

func (c *client) NextCommand() error {
	conn, err := net.Dial("unix", c.socketPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	data := RunCommand{
		Command: nextCommand,
	}

	return json.NewEncoder(conn).Encode(data)
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

func (c *client) FetchCommand() string {
	conn, err := net.Dial("unix", c.socketPath)
	if err != nil {
		return err.Error()
	}

	data := RunCommand{
		Command: fetchCommand,
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
