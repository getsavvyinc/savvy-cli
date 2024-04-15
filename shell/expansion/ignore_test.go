package expansion_test

import (
	"testing"

	"github.com/getsavvyinc/savvy-cli/shell/expansion"
)

func TestIgnoreGrep(t *testing.T) {
	testCases := []struct {
		name     string
		message  string
		expected string
	}{
		{
			name:     "no grep",
			message:  "echo hello world",
			expected: "echo hello world",
		},
		{
			name:     "grep",
			message:  `grep --color=auto --exclude-dir={.bzr,CVS,.git,.hg,.svn} "pattern"`,
			expected: `grep "pattern"`,
		},
		{
			name:     "grep with flags",
			message:  `grep --color=auto --exclude-dir={.bzr,CVS,.git,.hg,.svn} -i "pattern"`,
			expected: `grep -i "pattern"`,
		},
		{
			name:     "grep with multiple flags",
			message:  `grep --color=auto --exclude-dir={.bzr,CVS,.git,.hg,.svn} -i -n "pattern"`,
			expected: `grep -i -n "pattern"`,
		},
		{
			name:     "multiple grep instances",
			message:  `grep --color=auto --exclude-dir={.bzr,CVS,.git,.hg,.svn} -i -n "pattern1" | grep --color=auto --exclude-dir={.bzr,CVS,.git,.hg,.svn} "pattern2"`,
			expected: `grep -i -n "pattern1" | grep "pattern2"`,
		},
		{
			name:     "grep with other commands",
			message:  `echo "hello world" | grep --color=auto --exclude-dir={.bzr,CVS,.git,.hg,.svn} "pattern"`,
			expected: `echo "hello world" | grep "pattern"`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := expansion.IgnoreGrep(tc.message)
			if actual != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, actual)
			}
		})
	}
}
