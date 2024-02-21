package shell

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
)

type bashSetupChecker struct{}

var _ SetupChecker = (*bashSetupChecker)(nil)

func (b *bashSetupChecker) CheckSetup() error {
	executablePath, _ := os.Executable()

	if _, err := exec.LookPath(filepath.Base(executablePath)); err != nil {
		return errors.New(b.pathInstruction())
	}
	return nil
}

func (b *bashSetupChecker) pathInstruction() string {
	return `
  Please add savvy to your $PATH by running the following command:

  echo 'export PATH="$HOME/bin:$PATH"' >> ~/.bashrc
  source ~/.bashrc
  `
}
