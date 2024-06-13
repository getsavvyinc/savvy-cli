package internal

import (
	"reflect"
	"testing"
)

func TestExtractParams(t *testing.T) {

	testCases := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:  "no params",
			input: "No parameters here!",
		},
		{
			name:     "param with alphabets",
			input:    `jobrunner.sh --file="<file>"`,
			expected: []string{"<file>"},
		},
		{
			name:     "param with numbers",
			input:    `script --id="<id1>" --name="<name2>"`,
			expected: []string{"<id1>", "<name2>"},
		},
		{
			name:     "incomplete param",
			input:    "Edge case with incomplete <param and another<param2>",
			expected: []string{"<param2>"},
		},
		{
			name:     "param with hyphen, underscore",
			input:    `script --id="<id-1>" --name="<name_2>"`,
			expected: []string{"<id-1>", "<name_2>"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := extractParams(tc.input)
			if !reflect.DeepEqual(actual, tc.expected) {
				t.Errorf("expected %v; got %v", tc.expected, actual)
			}
		})
	}
}
