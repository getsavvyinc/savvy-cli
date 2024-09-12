package shell

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/getsavvyinc/savvy-cli/client"
)

var _ Shell = (*fish)(nil)

type fish struct {
	shellCmd string
	// Exported to use in template
	SocketPath string
}

const fishBaseScript = `
switch "$OSTYPE"
    case 'solaris*'
    case 'darwin*'
    case 'linux*'
    case 'bsd*'
    case 'msys*'
        echo "windows os is not supported"
    case 'cygwin*'
        echo "windows os is not supported"
    case '*'
        echo "unknown: $OSTYPE"
end

set -g SAVVY_INPUT_FILE {{.SocketPath}}

# Fish doesn't use the same startup files as Bash, but we can approximate the behavior
if status is-login
    if test -f "/etc/fish/config.fish"
        source "/etc/fish/config.fish"
    end
    if test -f "$HOME/.config/fish/config.fish"
        source "$HOME/.config/fish/config.fish"
    end
else
    if test -f "/etc/fish/conf.d/login.fish"
        source "/etc/fish/conf.d/login.fish"
    end
    if test -f "$HOME/.config/fish/conf.d/login.fish"
        source "$HOME/.config/fish/conf.d/login.fish"
    end
end
`
const fishRecordSetup = `
if not functions -q __savvy_record_pre_exec__
    set_color red
    echo -n "Your recording shell is not configured to use Savvy. Please run the following commands: "
    set_color normal
    echo
    set_color red echo "> echo 'savvy init fish | source' >> ~/.config/fish/config.fish" set_color normal
     set_color red echo "> source ~/.config/fish/config.fish" set_color normal
    exit 1
end
echo
echo "Type 'exit' or press 'ctrl+d' to stop recording."
`

var fishTemplate, fishHistoryTemplate, fishRunTemplate *template.Template

func init() {
	fishTemplate = template.Must(template.New("fish").Parse(fishBaseScript + fishRecordSetup))
	fishRunTemplate = template.Must(template.New("fishRun").Parse(fishBaseScript + fishRunRunbookScript))
}

// Spawn starts a fish shell.
func (f *fish) Spawn(ctx context.Context) (*exec.Cmd, error) {
	// Create a temporary file to store the script
	tmpDir, err := os.MkdirTemp("", "savvy-fish-*")
	if err != nil {
		return nil, err
	}
	fishVendorConfDir := filepath.Join(tmpDir, "fish", "vendor_conf.d")
	if err := os.MkdirAll(fishVendorConfDir, 0755); err != nil {
		return nil, err
	}
	fishrc, err := os.CreateTemp(fishVendorConfDir, "savvy-fishrc-*.fish")
	if err != nil {
		return nil, err
	}
	defer fishrc.Close()

	if err := fishTemplate.Execute(fishrc, f); err != nil {
		return nil, err
	}

	dataDirs := os.Getenv("XDG_DATA_DIRS")
	if dataDirs == "" {
		dataDirs = tmpDir
	} else {
		dataDirs = strings.Join([]string{dataDirs, fishVendorConfDir}, ":")
	}
	cmd := exec.CommandContext(ctx, f.shellCmd)
	cmd.Env = append(os.Environ(), "SAVVY_CONTEXT=record", fmt.Sprintf("XDG_DATA_DIRS=%s", dataDirs))
	cmd.WaitDelay = 500 * time.Millisecond
	return cmd, nil
}

func (f *fish) TailHistory(ctx context.Context) ([]string, error) {
	return nil, nil
}

func (f *fish) SpawnHistoryExpander(ctx context.Context) (*exec.Cmd, error) {
	return nil, nil
}

const fishRunRunbookScript = `
if not functions -q __savvy_run_pre_exec__
   or not functions -q __savvy_run_completion__
    set_color red
    echo -n " Your shell is not configured to use Savvy. Please run the following commands: "
    set_color normal
    set_color red
    echo
    echo -n "> echo 'savvy init fish | source' >> ~/.config/fish/config.fish"
    echo -n "> source ~/.config/fish/config.fish"
    set_color normal
    exit 1
end

echo "HERE!!!"

bind \cn '__savvy_run_completion__ "__savvy_run_completion__"'

echo
echo "Type 'exit' or press 'ctrl+d' to stop running."
echo
`

func (f *fish) SpawnRunbookRunner(ctx context.Context, runbook *client.Runbook) (*exec.Cmd, error) {
	tmpDir, err := os.MkdirTemp("", "savvy-fish-*")
	if err != nil {
		return nil, err
	}
	fishVendorConfDir := filepath.Join(tmpDir, "fish", "vendor_conf.d")
	if err := os.MkdirAll(fishVendorConfDir, 0755); err != nil {
		return nil, err
	}
	fishrc, err := os.CreateTemp(fishVendorConfDir, "savvy-fishrc-*.fish")
	if err != nil {
		return nil, err
	}
	defer fishrc.Close()

	if err := fishRunTemplate.Execute(fishrc, f); err != nil {
		return nil, err
	}

	dataDirs := os.Getenv("XDG_DATA_DIRS")
	if dataDirs == "" {
		dataDirs = tmpDir
	} else {
		dataDirs = strings.Join([]string{dataDirs, tmpDir}, ":")
	}
	cmd := exec.CommandContext(ctx, f.shellCmd)
	cmd.Env = append(os.Environ(), "SAVVY_CONTEXT=run", fmt.Sprintf("XDG_DATA_DIRS=%s", dataDirs))
	cmd.Env = append(cmd.Env, runbookRunMetadata(runbook, f)...)
	cmd.WaitDelay = 500 * time.Millisecond
	return cmd, nil
}

func (f *fish) DefaultStartingArrayIndex() int {
	return 1
}
