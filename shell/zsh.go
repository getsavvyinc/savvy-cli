package shell

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

type Shell interface {
	Spawn(ctx context.Context) (*exec.Cmd, error)
}

func New(logTarget string) Shell {
	switch shell := os.Getenv("SHELL"); {
	case strings.HasSuffix(shell, "/zsh"):
		return &zsh{
			shellCmd:   shell,
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
if [[ "$OSTYPE" == "darwin"* ]] ; then
    SAVVY_LOGIN_SHELL=1
fi
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

# autoload -Uz add-zsh-hook
#
# This function fixes the prompt via a precmd hook.
# function __savvy_cmd_pre_cmd__() {
#  local status=$?
#  echo "exit_status: ${status}</command>" > $SAVVY_STATUS_FILE
# }

# function __savvy_cmd_pre_exec__() {
#   # $2 is the command with all the aliases expanded
#   local cmd=$3
#   echo "${cmd}" | nc -U $SAVVY_INPUT_FILE
# }
# add-zsh-hook preexec __savvy_cmd_pre_exec__

# When promptinit is activated, a precmd hook which updates PS1 is installed.
# In order to inject the kubie PS1 when promptinit is activated, we must
# also add our own precmd hook which modifies PS1 after promptinit themes.
# add-zsh-hook precmd __savvy_cmd_pre_cmd__
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

	cmd := exec.Command(z.shellCmd)
	cmd.Env = append(os.Environ(), "ZDOTDIR="+tmp, "SAVVY_CONTEXT=1")
	cmd.WaitDelay = 2 * time.Second
	return cmd, nil
}

type todo struct{}

func (t *todo) Spawn(ctx context.Context) (*exec.Cmd, error) {
	return nil, errors.New("Not implemented")
}
