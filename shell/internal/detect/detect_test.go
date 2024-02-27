package detect

import (
	"testing"
)

func TestParseCommand(t *testing.T) {
	testCases := []struct {
		rawCommandAndArgs string
		expected          string
	}{
		{
			rawCommandAndArgs: "-zsh",
			expected:          "zsh",
		},
		{rawCommandAndArgs: "tmux attach -t savvy",
			expected: "tmux",
		},
		{
			rawCommandAndArgs: "/bin/zsh -il",
			expected:          "zsh",
		},
		{rawCommandAndArgs: "/nix/store/2gs1bzkbap7r4nr7vvyi575mdrj26nby-nodejs-18.18.2/bin/node /nix/store/9kl1b483qxmq4if81i6affypgcrxpmsh-yarn-1.22.19/bin/yarn dev",
			expected: "node",
		},
		{rawCommandAndArgs: "/Users/shantanu/src/github.com/savvy-prototype/www/node_modules/@esbuild/darwin-arm64/bin/esbuild --service=0.19.9 --ping",
			expected: "esbuild",
		},
		{
			rawCommandAndArgs: "nvim main.go",
			expected:          "nvim",
		},
		{rawCommandAndArgs: "sh -c 'fzf'  --border '--color=bg+:#3b4252,bg:#2e3440,spinner:#81a1c1,hl:#616e88,fg:#d8dee9,header:#616e88,info:#81a1c1,pointer:#81a1c1,marker:#81a1c1,fg+:#d8dee9,prompt:#81a1c1,hl+:#81a1c1' +m --ansi --tiebreak=begin --header-lines=1 -d\240 '--preview' '",
			expected: "sh",
		},
	}

	for _, tc := range testCases {
		got := parseCommand(tc.rawCommandAndArgs)
		if got != tc.expected {
			t.Errorf("wrong output for commandOf(%s). expected %s, got %s", tc.rawCommandAndArgs, tc.expected, got)
		}
	}
}
