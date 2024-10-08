package shell

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/getsavvyinc/savvy-cli/client"
	"github.com/getsavvyinc/savvy-cli/tail"
)

const RunbookCommandDelimiter = "COMMA"

type zsh struct {
	shellCmd string
	// Exported to use in template
	SocketPath string
}

var _ Shell = (*zsh)(nil)

const baseScript = `
# Reference for loading behavior
# https://shreevatsa.wordpress.com/2008/03/30/zshbash-startup-files-loading-order-bashrc-zshrc-etc/

RED=$(tput setaf 1)
RESET=$(tput sgr0)

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

`

const recordScript = `
if ! whence -v __savvy_record_pre_exec__ >/dev/null; then
echo "${RED} Your shell is not configured to use Savvy. Please run the following commands: ${RESET}"
echo
echo "${RED}> echo 'eval \"\$(savvy init zsh)\"' >> ~/.zshrc${RESET}"
echo "${RED}> source ~/.zshrc${RESET}"
exit 1
fi

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

	t := template.Must(template.New("zshrc").Parse(baseScript + recordScript))

	if err := t.Execute(zshrc, z); err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, z.shellCmd)
	cmd.Env = append(os.Environ(), "ZDOTDIR="+tmp, "SAVVY_CONTEXT=record")
	cmd.WaitDelay = 500 * time.Millisecond
	return cmd, nil
}

const historyScript = `
function __savvy_history_pre_exec__ {
  local cmd=${3}

  if [[ -n "${cmd}" ]]; then
     # Send the command to the unix socket server
     # Running it as a b/g process is intentional here
     SAVVY_SOCKET_PATH=${SAVVY_INPUT_FILE} savvy send "$cmd" &
  fi
  # This is how we prevent the command from being executed
  exec zsh
}

autoload -Uz add-zsh-hook
add-zsh-hook preexec __savvy_history_pre_exec__
`

func (z *zsh) SpawnHistoryExpander(ctx context.Context) (*exec.Cmd, error) {
	tmp := os.TempDir()
	zshrcPath := filepath.Join(tmp, ".zshrc")
	zshrc, err := os.Create(zshrcPath)
	if err != nil {
		return nil, err
	}
	defer zshrc.Close()

	t := template.Must(template.New("historyZshrc").Parse(baseScript + historyScript))
	if err := t.Execute(zshrc, z); err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, z.shellCmd)
	cmd.Env = append(os.Environ(), "ZDOTDIR="+tmp)
	cmd.WaitDelay = 500 * time.Millisecond
	return cmd, nil
}

func (z *zsh) TailHistory(ctx context.Context) ([]string, error) {
	historyFile := os.Getenv("HISTFILE")
	if historyFile == "" {
		u, err := user.Current()
		if err != nil {
			return nil, err
		}
		historyFile = filepath.Join(u.HomeDir, ".zsh_history")
	}

	rc, err := tail.Tail(historyFile, 100)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	var result []string
	var currentCommand strings.Builder
	scanner := bufio.NewScanner(rc)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, ": ") {
			// New command starts
			if currentCommand.Len() > 0 {
				result = append(result, strings.TrimSpace(currentCommand.String()))
				currentCommand.Reset()
			}
			// Remove timestamp and semicolon
			parts := strings.SplitN(line, ";", 2)
			if len(parts) > 1 {
				currentCommand.WriteString(strings.TrimSpace(parts[1]))
			} else {
				result = append(result, parts[:]...)
			}
		} else {
			// Continuation of a multiline command
			if strings.HasSuffix(currentCommand.String(), "\\\\") {
				// If the command so far ends with an escaped backslash, we know that there was a newline that the scanner removed.
				// So we remove the escaped backslash, so the command is runnable ( this is a zsh specific behavior) and ctrl-r does something similar as well.
				// We also add the newline back at the end of the branch and keep goin.
				escapedCommand := currentCommand.String()[0 : len(currentCommand.String())-1]
				currentCommand.Reset()
				currentCommand.WriteString(escapedCommand)
			}
			currentCommand.WriteString("\n" + line)
		}
	}

	// Add the last command if there is one
	if currentCommand.Len() > 0 {
		result = append(result, strings.TrimSpace(currentCommand.String()))
	}

	if err := scanner.Err(); err != nil {
		err = fmt.Errorf("error reading history file: %w", err)
		return nil, err
	}
	// reverse the result
	slices.Reverse(result)
	// TODO: handle more history formats for zsh here.
	return result, nil
}

const runRunbookScript = `
if ! whence -v __savvy_run_pre_exec__ >/dev/null; then
echo "${RED} Your shell is not configured to use Savvy. Please run the following commands: ${RESET}"
echo
echo "${RED}> echo 'eval \"\$(savvy init zsh)\"' >> ~/.zshrc${RESET}"
echo "${RED}> source ~/.zshrc${RESET}"
exit 1
fi

echo
echo "Type 'exit' or press 'ctrl+d' to stop running."
echo
`

func (z *zsh) SpawnRunbookRunner(ctx context.Context, runbook *client.Runbook) (*exec.Cmd, error) {
	tmp := os.TempDir()
	zshrcPath := filepath.Join(tmp, ".zshrc")
	zshrc, err := os.Create(zshrcPath)
	if err != nil {
		return nil, err
	}
	defer zshrc.Close()

	t := template.Must(template.New("zshrc").Parse(baseScript + runRunbookScript))

	if err := t.Execute(zshrc, z); err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, z.shellCmd)
	cmd.Env = append(os.Environ(), "ZDOTDIR="+tmp)
	cmd.Env = append(cmd.Env, runbookRunMetadata(runbook, z)...)
	cmd.WaitDelay = 500 * time.Millisecond
	return cmd, nil
}

func (z *zsh) DefaultStartingArrayIndex() int {
	return 1
}

func computeRunbookAlias(runbook *client.Runbook) string {
	lc := strings.ToLower(runbook.Title)
	alias := strings.ReplaceAll(lc, " ", "-")
	alias, _ = strings.CutPrefix(alias, "how-to-")
	alias = strings.Trim(alias, "-")
	return alias
}

func nextRunbookStepToRun(sh Shell) int {
	// Inherit the next step from the environment if we are in a subshell
	nextStep := os.Getenv("SAVVY_NEXT_STEP")
	if nextStep == "" {
		return sh.DefaultStartingArrayIndex()
	}

	idx, err := strconv.Atoi(nextStep)
	if err != nil {
		return sh.DefaultStartingArrayIndex()
	}
	return idx
}

func runbookRunMetadata(runbook *client.Runbook, sh Shell) []string {
	runbookCommands := strings.Join(runbook.Commands(), RunbookCommandDelimiter)
	runbookAlias := computeRunbookAlias(runbook)

	return []string{
		"SAVVY_CONTEXT=run",
		fmt.Sprintf("SAVVY_RUNBOOK_COMMANDS=%s", runbookCommands),
		fmt.Sprintf("SAVVY_NEXT_STEP=%d", nextRunbookStepToRun(sh)),
		fmt.Sprintf("SAVVY_RUNBOOK_ALIAS=%s", runbookAlias),
	}
}
