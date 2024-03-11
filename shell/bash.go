package shell

import (
	"bufio"
	"context"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"slices"
	"text/template"
	"time"

	"github.com/getsavvyinc/savvy-cli/tail"
)

type bash struct {
	shellCmd string
	// Exported to use in template
	SocketPath string
}

var _ Shell = (*bash)(nil)

// Adapted from: https://github.com/sbstp/kubie/
const bashrcScript = `
RED=$(tput setaf 1)
RESET=$(tput sgr0)

case "$OSTYPE" in
  solaris*) ;;
  darwin*)  ;;
  linux*)   ;;
  bsd*)     ;;
  msys*)    echo "windows os is not supported" ;;
  cygwin*)  echo "windows os is not supported" ;;
  *)        echo "unknown: $OSTYPE" ;;
esac

SAVVY_INPUT_FILE={{.SocketPath}}

# Reference for loading behavior
# https://shreevatsa.wordpress.com/2008/03/30/zshbash-startup-files-loading-order-bashrc-zshrc-etc/


if shopt -q login_shell; then
    if [[ -f "/etc/profile" ]] ; then
        source "/etc/profile"
    fi

    if [[ -f "$HOME/.bash_profile" ]] ; then
        source "$HOME/.bash_profile"
    elif [[ -f "$HOME/.bash_login" ]] ; then
        source "$HOME/.bash_login"
    elif [[ -f "$HOME/.profile" ]] ; then
        source "$HOME/.profile"
    fi
else
    if [[ -f "/etc/bash.bashrc" ]] ; then
        source "/etc/bash.bashrc"
    fi

    if [[ -f "$HOME/.bashrc" ]] ; then
        source "$HOME/.bashrc"
    fi
fi

if ! type savvy_cmd_pre_exec >/dev/null 2>&1; then
echo "${RED} Your shell is not configured to use Savvy. Please run the following commands: ${RESET}"
echo
echo "${RED}> echo 'eval \"\$(savvy init bash)\"' >> ~/.bashrc${RESET}"
echo "${RED}> source ~/.bashrc${RESET}"
exit 1
fi

echo
echo "Type 'exit' or press 'ctrl+d' to stop recording."
`

var bashTemplate *template.Template

func init() {
	bashTemplate = template.Must(template.New("bash").Parse(bashrcScript))
}

func (b *bash) Spawn(ctx context.Context) (*exec.Cmd, error) {
	// Create a temporary file to store the script
	tmpDir := os.TempDir()
	bashrc, err := os.CreateTemp(tmpDir, "savvy-bashrc-*.bash")
	if err != nil {
		return nil, err
	}
	defer bashrc.Close()

	if err := bashTemplate.Execute(bashrc, b); err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, b.shellCmd, "--rcfile", bashrc.Name())
	cmd.Env = append(os.Environ(), "SAVVY_CONTEXT=1")
	cmd.WaitDelay = 500 * time.Millisecond
	return cmd, nil
}

// bashTimestamp matches the timestamp line in bash history files.
// Example: #1616420000
var bashTimestamp = regexp.MustCompile(`^#(\d){10}$`)

func (b *bash) TailHistory(ctx context.Context) ([]string, error) {
	historyFile := os.Getenv("HISTFILE")
	if historyFile == "" {
		u, err := user.Current()
		if err != nil {
			return nil, err
		}
		historyFile = filepath.Join(u.HomeDir, ".bash_history")
	}

	// Bash history files record timestamp on one line and command on the next line.
	// So to read the last 100 commands, we need to read a maximum of 200 lines.
	rc, err := tail.Tail(historyFile, 200)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	var lines []string
	scanner := bufio.NewScanner(rc)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			lines = append(lines, line)
		}
	}
	// reverse the result
	slices.Reverse(lines)
	// extract the command from the history
	var result []string
	for _, line := range lines {
		// skip timestamps in the history
		if bashTimestamp.MatchString(line) {
			continue
		}
		result = append(result, line)
	}
	return result, nil
}

func (b *bash) SpawnHistoryExpander(ctx context.Context) (*exec.Cmd, error) {
	return nil, nil
}
