package shell

import (
	"context"
	"errors"
	"os/exec"

	"github.com/getsavvyinc/savvy-cli/client"
	"github.com/getsavvyinc/savvy-cli/shell/internal/detect"
	"github.com/getsavvyinc/savvy-cli/shell/kind"
)

type Shell interface {
	Spawn(ctx context.Context) (*exec.Cmd, error)
	TailHistory(ctx context.Context) ([]string, error)
	SpawnHistoryExpander(ctx context.Context) (*exec.Cmd, error)
	SpawnRunbookRunner(ctx context.Context, runbook *client.Runbook) (*exec.Cmd, error)
}

func New(logTarget string) Shell {
	shell := detect.DetectWithDefault()
	switch shell {
	case kind.Zsh:
		return &zsh{
			shellCmd:   "zsh",
			SocketPath: logTarget,
		}
	case kind.Dash:
		fallthrough
	case kind.Bash:
		return &bash{
			shellCmd:   "bash",
			SocketPath: logTarget,
		}
	default:
		return &todo{}
	}
}

type todo struct{}

func (t *todo) Spawn(ctx context.Context) (*exec.Cmd, error) {
	return nil, errors.New("savvy doesn't support your current shell")
}

func (t *todo) TailHistory(ctx context.Context) ([]string, error) {
	return nil, errors.New("savvy doesn't support your current shell")
}

func (t *todo) SpawnHistoryExpander(ctx context.Context) (*exec.Cmd, error) {
	return nil, errors.New("savvy doesn't support your current shell")
}

func (t *todo) SpawnRunbookRunner(ctx context.Context, runbook *client.Runbook) (*exec.Cmd, error) {
	return nil, errors.New("savvy doesn't support your current shell")
}
