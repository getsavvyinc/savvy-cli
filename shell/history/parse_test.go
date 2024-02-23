package history_test

import (
	"testing"

	"github.com/getsavvyinc/savvy-cli/shell/history"
	"github.com/stretchr/testify/assert"
)

func TestHistoryParser(t *testing.T) {
	testCases := []struct {
		name           string
		input          []string
		expectedOutput []*history.Command
	}{
		{
			name:           "empty history",
			input:          []string{},
			expectedOutput: nil,
		},
		{
			name: "history with multiple commands and spaces",

			input: []string{
				"  1  echo hello ",
				"  2  echo world ",
				"  3  echo hello world ",
			},
			expectedOutput: []*history.Command{
				history.NewCommand("echo hello"),
				history.NewCommand("echo world"),
				history.NewCommand("echo hello world"),
			},
		},
		{
			name: "history with malformed lines",
			input: []string{
				"  1  echo hello 44.234ms ",
				"  2  echo world ",
			},
			expectedOutput: []*history.Command{
				history.NewCommand("echo hello 44.234ms"),
				history.NewCommand("echo world"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parser := history.NewParser()
			parsedOutput := parser.ParseHistory(tc.input)
			assert.Equalf(t, len(tc.expectedOutput), len(parsedOutput), "expected %d commands, got %d", len(tc.expectedOutput), len(parsedOutput))
			for i, expected := range tc.expectedOutput {
				assert.Equalf(t, expected.GetCommand(), parsedOutput[i].GetCommand(), "index: %d", i)
			}
		})
	}
}
