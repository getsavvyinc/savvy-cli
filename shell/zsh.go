package shell

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
	"time"
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
	default:
		return &todo{}
	}
}

type zsh struct {
	shellCmd string
	// Exported to use in template
	SocketPath string
}

const script = `
  # Reference for loading behavior
  # https://shreevatsa.wordpress.com/2008/03/30/zshbash-startup-files-loading-order-bashrc-zshrc-etc/

SAVVY_INPUT_FILE={{.SocketPath}}

if [[ -f "/etc/zshenv" ]] ; then
    source "/etc/zshenv"
elif [[ -f "/etc/zsh/zshenv" ]] ; then
    source "/etc/zsh/zshenv"
fi

if [[ -f "$HOME/.zshenv" ]] ; then
    tmp_ZDOTDIR=$ZDOTDIR
    source "$HOME/.zshenv"
    # If the user has overridden $ZDOTDIR, we save that in $_SAVVY_USER_ZDOTDIR for later reference
    # and reset $ZDOTDIR
    if [[ "$tmp_ZDOTDIR" != "$ZDOTDIR" ]]; then
        _SAVVY_USER_ZDOTDIR=$ZDOTDIR
        ZDOTDIR=$tmp_ZDOTDIR
        unset tmp_ZDOTDIR
    fi
fi

# If a zsh_history file exists, copy it over before zsh initialization so history is maintained
if [[ -f "$HOME/.zsh_history" ]] ; then
    cp $HOME/.zsh_history $ZDOTDIR
fi

SAVVY_LOGIN_SHELL=0

case "$OSTYPE" in
  solaris*) SAVVY_LOGIN_SHELL=1;;
  darwin*)  SAVVY_LOGIN_SHELL=1;;
  linux*)   SAVVY_LOGIN_SHELL=1;;
  bsd*)     SAVVY_LOGIN_SHELL=1;;
  msys*)    echo "windows os is not supported" ;;
  cygwin*)  echo "windows os is not supported" ;;
  *)        echo "unknown: $OSTYPE" ;;
esac

if [[ -f "/etc/zprofile" && "$SAVVY_LOGIN_SHELL" == "1" ]] ; then
    source "/etc/zprofile"
elif [[ -f "/etc/zsh/zprofile" && "$SAVVY_LOGIN_SHELL" == "1" ]] ; then
    source "/etc/zsh/zprofile"
fi

if [[ -f "${_SAVVY_USER_ZDOTDIR:-$HOME}/.zprofile" && "$SAVVY_LOGIN_SHELL" == "1" ]] ; then
    source "${_SAVVY_USER_ZDOTDIR:-$HOME}/.zprofile"
fi

if [[ -f "/etc/zshrc" ]] ; then
    source "/etc/zshrc"
elif [[ -f "/etc/zsh/zshrc" ]] ; then
    source "/etc/zsh/zshrc"
fi

if [[ -f "${_SAVVY_USER_ZDOTDIR:-$HOME}/.zshrc" ]] ; then
    source "${_SAVVY_USER_ZDOTDIR:-$HOME}/.zshrc"
fi

if [[ -f "/etc/zlogin" && "$SAVVY_LOGIN_SHELL" == "1" ]] ; then
    source "/etc/zlogin"
elif [[ -f "/etc/zsh/zlogin" && "$SAVVY_LOGIN_SHELL" == "1" ]] ; then
    source "/etc/zsh/zlogin"
fi

if [[ -f "${_SAVVY_USER_ZDOTDIR:-$HOME}/.zlogin" && "$SAVVY_LOGIN_SHELL" == "1" ]] ; then
    source "${_SAVVY_USER_ZDOTDIR:-$HOME}/.zlogin"
fi

unset _SAVVY_USER_ZDOTDIR

echo
echo "Type 'exit' or press 'ctrl+d' to stop recording."
`

func (z *zsh) Spawn(ctx context.Context) (*exec.Cmd, error) {
	// Referenced: https://github.com/sbstp/kubie/blob/master/src/shell/zsh.rs
	tmp := os.TempDir()
	zshrcPath := filepath.Join(tmp, ".zshrc")
	zshrc, err := os.Create(zshrcPath)
	if err != nil {
		return nil, err
	}
	defer zshrc.Close()

	t := template.Must(template.New("zshrc").Parse(script))

	if err := t.Execute(zshrc, z); err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, z.shellCmd)
	cmd.Env = append(os.Environ(), "ZDOTDIR="+tmp, "SAVVY_CONTEXT=1")
	cmd.WaitDelay = 500 * time.Millisecond
	return cmd, nil
}

type todo struct{}

func (t *todo) Spawn(ctx context.Context) (*exec.Cmd, error) {
	return nil, errors.New("not implemented")
}
