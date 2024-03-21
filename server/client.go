package server

import (
	"context"
	"fmt"
	"net"
)

type Client interface {
	Send(msg string) error
}

func NewDefaultClient(ctx context.Context) (Client, error) {
	return &client{
		socketPath: defaultSocketPath,
	}, nil
}

type client struct {
	socketPath string
}

var _ Client = &client{}

func (c *client) Send(msg string) error {
	conn, err := net.Dial("unix", c.socketPath)
	if err != nil {
		return err
	}
	defer conn.Close()
	if len(msg) == 0 {
		return nil
	}

	// TODO: update this to send RecordedData
	if _, err = fmt.Fprintf(conn, "%s\n", msg); err != nil {
		return err
	}
	return nil
}
