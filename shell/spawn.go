package shell

import (
	"context"
	"errors"
	"os/exec"
)

type Shell interface {
	Spawn(ctx context.Context) (*exec.Cmd, error)
}

func New(logTarget string) Shell {
	shell := detectWithDefault()
	switch shell {
	case Zsh:
		return &zsh{
			shellCmd:   "zsh",
			SocketPath: logTarget,
		}
	case Dash:
		fallthrough
	case Bash:
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
