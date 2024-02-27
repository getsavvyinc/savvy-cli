package shell

import (
	"context"
	"errors"
	"os/exec"

	"github.com/getsavvyinc/savvy-cli/shell/internal/detect"
	"github.com/getsavvyinc/savvy-cli/shell/kind"
)

type Shell interface {
	Spawn(ctx context.Context) (*exec.Cmd, error)
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
