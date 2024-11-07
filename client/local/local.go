package local

import (
	"context"
	"errors"
	"fmt"

	"github.com/getsavvyinc/savvy-cli/client"
	"github.com/getsavvyinc/savvy-cli/storage"
)

func New() client.RunbookClient {
	return &local{}
}

type local struct{}

var ErrNotFound = errors.New("not found")

func (l *local) RunbookByID(ctx context.Context, id string) (*client.Runbook, error) {
	rbs, err := storage.Read()
	if err != nil {
		return nil, err
	}

	rb, ok := rbs[id]
	if !ok {
		err = fmt.Errorf("runbook %s: %w", id, ErrNotFound)
		return nil, err
	}
	return rb, nil
}

// Runbooks returns all runbooks stored in the local  storage
func (l *local) Runbooks(ctx context.Context, _ client.RunbooksOpt) ([]client.RunbookInfo, error) {

	rbs, err := storage.Read()
	if err != nil {
		return nil, err
	}

	var rbis []client.RunbookInfo
	for id, rb := range rbs {
		rbis = append(rbis, client.RunbookInfo{
			RunbookID: id,
			Title:     rb.Title,
		})
	}
	return rbis, nil
}
