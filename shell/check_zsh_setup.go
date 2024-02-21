package shell

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
)

type zshSetupChecker struct{}

var _ SetupChecker = (*zshSetupChecker)(nil)

func (z *zshSetupChecker) CheckSetup() error {
	executablePath, _ := os.Executable()

	if _, err := exec.LookPath(filepath.Base(executablePath)); err != nil {
		return errors.New(z.pathInstruction())
	}
	return nil
}

func (z *zshSetupChecker) pathInstruction() string {
	return `
  Please add savvy to your $PATH by running the following commands:

  echo 'export PATH="$HOME/bin:$PATH"' >> ~/.zshrc
  source ~/.zshrc
  `
}
