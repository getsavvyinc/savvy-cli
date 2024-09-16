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
`

var fishTemplate *template.Template

func init() {
	fishTemplate = template.Must(template.New("fish").Parse(fishBaseScript))
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
		dataDirs = strings.Join([]string{dataDirs, tmpDir}, ":")
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

	if err := fishTemplate.Execute(fishrc, f); err != nil {
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
