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

const fishRecordHistoryScript = `
function savvy_record_history_skip_execution --on-event fish_preexec
  set -l cmd $argv[1]
  SAVVY_SOCKET_PATH={{.SocketPath}} savvy send "$cmd"
  exec fish
end
`

var (
	fishTemplate              *template.Template
	fishRecordHistoryTemplate *template.Template
)

func init() {
	fishTemplate = template.Must(template.New("fish").Parse(fishBaseScript))
}

// injectVendorCode injects the vendor code into the fish shell configuration.
// It returns the parent directory that containsins the fish/vendor_conf.d directory.
// This directory should be added to the XDG_DATA_DIRS environment variable.
func (f *fish) injectVendorCode(vendorCode *template.Template) (string, error) {
	// Create a temporary file to store the script
	tmpDir, err := os.MkdirTemp("", "savvy-fish-*")
	if err != nil {
		return "", err
	}
	fishVendorConfDir := filepath.Join(tmpDir, "fish", "vendor_conf.d")
	if err := os.MkdirAll(fishVendorConfDir, 0755); err != nil {
		return "", err
	}
	fishrc, err := os.CreateTemp(fishVendorConfDir, "savvy-fishrc-*.fish")
	if err != nil {
		return "", err
	}
	defer fishrc.Close()

	if err := vendorCode.Execute(fishrc, f); err != nil {
		return "", err
	}
	return tmpDir, nil
}

func addVendorDirToXDGDataDirPath(vendorDir string) string {
	if dataDirs := os.Getenv("XDG_DATA_DIRS"); len(dataDirs) > 0 {
		return strings.Join([]string{dataDirs, vendorDir}, ":")
	}
	return vendorDir
}

// Spawn starts a fish shell.
func (f *fish) Spawn(ctx context.Context) (*exec.Cmd, error) {
	// Create a temporary file to store the script
	vendorDir, err := f.injectVendorCode(fishTemplate)
	if err != nil {
		return nil, err
	}

	dataDirs := addVendorDirToXDGDataDirPath(vendorDir)

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
	vendorDir, err := f.injectVendorCode(fishTemplate)
	if err != nil {
		return nil, err
	}
	dataDirs := addVendorDirToXDGDataDirPath(vendorDir)

	cmd := exec.CommandContext(ctx, f.shellCmd)
	cmd.Env = append(os.Environ(), "SAVVY_CONTEXT=run", fmt.Sprintf("XDG_DATA_DIRS=%s", dataDirs))
	cmd.Env = append(cmd.Env, runbookRunMetadata(runbook, f)...)
	cmd.WaitDelay = 500 * time.Millisecond
	return cmd, nil
}

func (f *fish) DefaultStartingArrayIndex() int {
	return 1
}
