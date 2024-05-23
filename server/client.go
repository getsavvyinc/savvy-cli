package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"

	"github.com/getsavvyinc/savvy-cli/idgen"
)

type Client interface {
	// SendFileInfo tells the server to read the file at the given path
	SendFileInfo(filePath string) error
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

func (c *client) SendFileInfo(filePath string) error {
	conn, err := net.Dial("unix", c.socketPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := checkFile(filePath); err != nil {
		return err
	}

	data := RecordedData{
		Command:  fmt.Sprintf("savvy record file %s", filePath),
		Filepath: filePath,
		StepID:   idgen.New(idgen.FilePrefix),
	}

	if err := json.NewEncoder(conn).Encode(data); err != nil {
		return err
	}
	return nil
}

func checkFile(filePath string) error {
	if len(filePath) == 0 {
		return nil
	}

	fi, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	if fi.IsDir() {
		return errors.New("file path provided is a directory")
	}

	if fi.Size() == 0 {
		return errors.New("file provided is empty")
	}

	if fi.Size() > 25*1024 {
		return errors.New("file provided is too large: max size is 25KB")
	}
	return nil
}
