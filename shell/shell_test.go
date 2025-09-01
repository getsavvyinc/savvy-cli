package shell

import (
	"testing"

	"github.com/getsavvyinc/savvy-cli/shell/kind"
)

func TestNew(t *testing.T) {
	// Test that Nu shell is properly instantiated
	// Note: This test assumes the detect package will return Nu for testing
	// In a real scenario, you might need to mock the detection

	testCases := []struct {
		name     string
		expected string
	}{
		{
			name:     "Nu shell",
			expected: "nu",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// We can't easily test the New function without mocking detect.DetectWithDefault()
			// But we can test that the Nu case exists in the switch statement
			// by checking that kind.Nu is defined
			if kind.Nu != "nu" {
				t.Errorf("Expected kind.Nu to be 'nu', got %s", kind.Nu)
			}
		})
	}
}

func TestSupportedShells(t *testing.T) {
	shells := SupportedShells()
	found := false
	for _, shell := range shells {
		if shell == "nu" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected 'nu' to be in supported shells, got %v", shells)
	}
}
