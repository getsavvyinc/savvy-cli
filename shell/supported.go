package shell

import "github.com/getsavvyinc/savvy-cli/shell/kind"

func SupportedShells() []string {
	return []string{string(kind.Bash), string(kind.Zsh), string(kind.Dash)}
}
