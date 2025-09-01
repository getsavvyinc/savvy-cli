package shell

import (
	"context"
	"os/exec"

	"github.com/getsavvyinc/savvy-cli/client"
)

var _ Shell = (*nu)(nil)

type nu struct {
	shellCmd string
	// Exported to use in template
	SocketPath string
}

const nuBaseScript = `
$env.SAVVY_INPUT_FILE = "{{.SocketPath}}"
`

// nuRecordHistoryScript is a nushell script that records the command history.
const nuRecordHistoryScript = `
def savvy_record_history_skip_execution [cmd: string] {
  $env.SAVVY_SOCKET_PATH = $env.SAVVY_INPUT_FILE
  savvy send $cmd | ignore
  exec nu
}
`

func (n *nu) Spawn(ctx context.Context) (*exec.Cmd, error) {
	cmd := exec.CommandContext(ctx, n.shellCmd, "-c", nuBaseScript+`
echo "Type 'exit' or press 'ctrl+d' to stop recording."
`)
	cmd.Env = append(cmd.Environ(), "SAVVY_CONTEXT=record")
	return cmd, nil
}

func (n *nu) TailHistory(ctx context.Context) ([]string, error) {
	// For now, return empty slice as Nushell history parsing is complex
	// TODO: Implement proper Nushell history parsing
	return []string{}, nil
}

func (n *nu) SpawnHistoryExpander(ctx context.Context) (*exec.Cmd, error) {
	cmd := exec.CommandContext(ctx, n.shellCmd, "-c", nuBaseScript+nuRecordHistoryScript)
	cmd.Env = cmd.Environ()
	return cmd, nil
}

func (n *nu) SpawnRunbookRunner(ctx context.Context, runbook *client.Runbook) (*exec.Cmd, error) {
	cmd := exec.CommandContext(ctx, n.shellCmd, "-c", nuBaseScript+`
echo "Type 'ctrl+n' to get the next command."
echo "Type 'exit' or press 'ctrl+d' to stop recording."
`)
	cmd.Env = append(cmd.Environ(), runbookRunMetadata(runbook, n)...)
	return cmd, nil
}

func (n *nu) DefaultStartingArrayIndex() int {
	return 0
}
