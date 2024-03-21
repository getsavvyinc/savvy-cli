package server

import (
	"context"
	"encoding/json"
	"net"

	"github.com/getsavvyinc/savvy-cli/idgen"
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

	data := RecordedData{
		Command: msg,
		StepID:  idgen.New(idgen.FilePrefix),
	}

	if err := json.NewEncoder(conn).Encode(data); err != nil {
		return err
	}
	return nil
}
