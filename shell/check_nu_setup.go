package shell

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
)

type nuSetupChecker struct{}

var _ SetupChecker = (*nuSetupChecker)(nil)

func (n *nuSetupChecker) CheckSetup() error {
	executablePath, _ := os.Executable()

	if _, err := exec.LookPath(filepath.Base(executablePath)); err != nil {
		return errors.New(n.pathInstruction())
	}
	return nil
}

func (n *nuSetupChecker) pathInstruction() string {
	return `
  Please add savvy to your $PATH by running the following commands:

  $env.PATH = ($env.PATH | prepend $"($env.HOME)/bin")
  `
}
