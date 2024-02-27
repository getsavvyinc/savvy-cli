package detect

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/getsavvyinc/savvy-cli/display"
	"github.com/getsavvyinc/savvy-cli/shell/kind"
)

// NOTE: This file is adapted from kubie's shell detection code that can be found here: https://github.com/sbstp/kubie/blob/master/src/shell/detect.rs

func runPS(args []string) (string, error) {
	out, err := exec.Command("ps", args...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("could not run ps: %w", err)
	}
	return string(out), nil
}

// commandOf executes the 'ps' command to get the command line of a process by its PID.
func commandOf(pid string) (string, error) {
	// -o: user-defined format
	// args=: command with all its arguments as a string (no header because of the =)
	// pid: select by PID
	out, err := runPS([]string{"-o", "args=", pid})
	if err != nil {
		return "", fmt.Errorf("could not run ps =%s: %w", pid, err)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) == 0 {
		return "", fmt.Errorf("commandOf: empty output for pid=%s", pid)
	}
	return strings.TrimSpace(lines[0]), nil
}

// parseCommand extracts the command name from a command line string.
func parseCommand(cmd string) string {
	firstSpace := strings.Index(cmd, " ")
	if firstSpace == -1 {
		firstSpace = len(cmd)
	}
	binaryPath := cmd[:firstSpace]
	lastPathSep := strings.LastIndex(binaryPath, string(filepath.Separator))
	// if there is no path separator, binary is the whole path
	binary := binaryPath
	// otherwise:
	if lastPathSep != -1 {
		binary = binaryPath[lastPathSep+1:]
	}

	// Remove leading '-' and any non-alphanumeric characters from the binary name
	binary = strings.TrimLeft(binary, "-")
	binary = strings.TrimFunc(binary, func(c rune) bool {
		return c >= '0' && c <= '9' || c == '.'
	})

	return binary
}

// detect walks up the process tree to find out which shell is in use.
func detect() (kind.Kind, error) {
	parentPid := fmt.Sprintf("%d", os.Getppid())

	// Walk up the process tree until we find a shell.
	for parentPid != "1" {
		cmd, err := commandOf(parentPid)
		if err != nil {
			return "", err
		}

		name := parseCommand(cmd)
		if kind, ok := kind.ShellKindFromString(name); ok {
			return kind, nil
		}

		parentPid, err = parentOf(parentPid)
		if err != nil {
			return "", err
		}
	}

	return "", errors.New("could not detect shell in use")
}

func DetectWithDefault() kind.Kind {
	knd, err := detect()
	if err != nil {
		shell, ok := kind.ShellKindFromString(parseCommand(os.Getenv("SHELL")))
		if !ok {
			// exit if we can't even proceed with the default shell
			err := errors.New("could not detect your default shell")
			display.ErrorWithSupportCTA(err)
			os.Exit(1)
		}
		err = fmt.Errorf("could not detect your shell: %w. Defaulting to %s", err, shell)
		display.Error(err)
		return shell
	}
	return knd
}

// parentOf finds the parent PID of a given PID.
// This function is left as an exercise since it involves parsing /proc or similar,
// which is not directly related to the conversion task.
func parentOf(pid string) (string, error) {
	// -o: user-defined format
	// ppid=: parent process ID (no header because of the =)
	// $pid: select by PID
	out, err := runPS([]string{"-o", "ppid=", pid})
	if err != nil {
		return "", fmt.Errorf("could not get parent of pid=%s: %w", pid, err)
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) == 0 {
		return "", fmt.Errorf("parentOf: empty output for pid=%s", pid)
	}
	return strings.TrimSpace(lines[0]), nil
}
