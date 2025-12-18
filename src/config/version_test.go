package config

import (
	"strings"
	"testing"
)

const (
	validInput = `
version: 1
project:
  project_id: "pid"
  name: "name"
pages: []
`
	missingVersionInput = `
project:
  project_id: "pid"
  name: "name"
pages: []
`
	nonIntegerVersionInput = `
version: "1"
project:
  project_id: "pid"
  name: "name"
pages: []
`
	invalidYAMLInput = `
version: [
`
)

func TestDetectConfigVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		expected    int
		expectError bool
	}{
		{
			name:     "valid version",
			input:    validInput,
			expected: 1,
		},
		{
			name:        "missing version",
			input:       missingVersionInput,
			expectError: true,
		},
		{
			name:        "non-integer version",
			input:       nonIntegerVersionInput,
			expectError: true,
		},
		{
			name:        "invalid yaml",
			input:       invalidYAMLInput,
			expectError: true,
		},
	}

	for _, tc := range tests {
		testCase := tc
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			version, err := DetectConfigVersion(strings.NewReader(testCase.input))
			if testCase.expectError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if version != testCase.expected {
				t.Fatalf("expected version %d, got %d", testCase.expected, version)
			}
		})
	}
}
