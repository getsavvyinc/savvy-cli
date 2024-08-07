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
	CurrentState() (*State, error)
	SetParams(params map[string]string) error
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

// NOTE: If we don't wait for the response and if the server has go-routines for each connection, we might have race conditions.
// e.g: c.NextCommand() and c.CurrentState() might be called sequentially, but the server might schedule the go-routines such that c.CurrentState() completes before c.NextCommand()
// Then we might get the wrong state.
// To avoid this, for now, the server is single threaded.
func (c *client) NextCommand() error {
	conn, err := net.Dial("unix", c.socketPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	data := RunCommand{
		Command: nextCommand,
	}

	if err := json.NewEncoder(conn).Encode(data); err != nil {
		return err
	}

	return nil
}

func (c *client) PreviousCommand() error {
	conn, err := net.Dial("unix", c.socketPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	data := RunCommand{
		Command: previousCommand,
	}

	return json.NewEncoder(conn).Encode(data)
}

func (c *client) SetParams(params map[string]string) error {
	conn, err := net.Dial("unix", c.socketPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	data := RunCommand{
		Command: paramCommand,
		Params:  params,
	}

	if err := json.NewEncoder(conn).Encode(data); err != nil {
		return err
	}
	return nil
}

func (c *client) CurrentState() (*State, error) {
	conn, err := net.Dial("unix", c.socketPath)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	data := RunCommand{
		Command: currentCommand,
	}

	if err := json.NewEncoder(conn).Encode(data); err != nil {
		return nil, err
	}

	var response State
	if err := json.NewDecoder(conn).Decode(&response); err != nil {
		return nil, err
	}
	return &response, nil
}
