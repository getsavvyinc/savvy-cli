package parser_test

import (
	"testing"

	"github.com/getsavvyinc/savvy-cli/ingest/cheatsheet"
	"github.com/getsavvyinc/savvy-cli/ingest/cheatsheet/parser"
	"github.com/stretchr/testify/assert"
)

func TestTLDR(t *testing.T) {
	tests := []struct {
		name       string
		filepath   string
		err        error
		cheatSheet *cheatsheet.CheatSheet
	}{
		{
			name:     "Wrong File",
			filepath: "testdata/wrong-extension.txt",
			err:      parser.ErrRequiredMdFile,
		},
		{
			name:     "sh.md",
			filepath: "testdata/sh.md",
			cheatSheet: &cheatsheet.CheatSheet{
				Title:       "sh",
				Description: "Bourne shell, the standard command language interpreter. See also `histexpand` for history expansion. More information: <https://manned.org/sh>.",
				Examples: []*cheatsheet.Example{
					{
						Command:     "sh",
						Explanation: "Start an interactive shell session",
					},
					{
						Command:     `sh -c "{{command}}"`,
						Explanation: "Execute a command and then exit",
					},
					{
						Explanation: "Execute a script",
						Command:     `sh {{path/to/script.sh}}`,
					},
					{
						Command:     `sh -s`,
						Explanation: "Read and execute commands from `stdin`",
					},
				},
			},
		},
	}

	tldr := parser.New(cheatsheet.TLDR)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cs, err := tldr.Parse(tc.filepath)
			if tc.err != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tc.err)
				assert.Nil(t, cs)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cs)
				assert.Equal(t, tc.cheatSheet, cs)
			}
		})
	}
}
