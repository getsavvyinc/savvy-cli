package shell

import (
	"github.com/getsavvyinc/savvy-cli/shell/kind"
)

type SetupChecker interface {
	// CheckSetup returns an non nil error if the shell setup is not correct
	// The error message should contain instructions on how to fix the setup and is safe to display to the user
	CheckSetup() error
}

func NewSetupChecker() SetupChecker {
	shell := detectWithDefault()
	switch shell {
	case kind.Zsh:
		return &zshSetupChecker{}
	case kind.Bash:
		return &bashSetupChecker{}
	default:
		return &nopSetupChecker{}
	}
}

type nopSetupChecker struct{}

var _ SetupChecker = (*nopSetupChecker)(nil)

func (n *nopSetupChecker) CheckSetup() error {
	return nil
}
