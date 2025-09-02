package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecuteTouchNew(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	path := filepath.Join(tempDir, "test.md")
	args := TouchArgs{
		filepath: path,
		title:    "Test Title",
		path:     "test-path",
		debug:    true,
		noColor:  true,
		now:      "2025-01-01T00:00:00+09:00",
	}

	err = touchCmdEntrypoint(args)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(path)
	require.NoError(t, err)

	// Verify content
	content, err := os.ReadFile(path)
	require.NoError(t, err)

	expectedContent := `---
title: "Test Title"
path: "test-path"
description: ""
created_at: "2025-01-01T00:00:00+09:00"
updated_at: "2025-01-01T00:00:00+09:00"
---
`
	assert.Equal(t, expectedContent, string(content))

	// Then update the markdown.
	updateArgs := TouchArgs{
		filepath: path,
		title:    "Updated Title",
		path:     "updated-path",
		debug:    true,
		noColor:  true,
		now:      "2025-01-02T00:00:00+09:00",
	}
	err = touchCmdEntrypoint(updateArgs)
	require.NoError(t, err)

	// Verify content
	content, err = os.ReadFile(path)
	require.NoError(t, err)

	expectedContent = `---
title: "Updated Title"
path: "updated-path"
description: ""
created_at: "2025-01-01T00:00:00+09:00"
updated_at: "2025-01-02T00:00:00+09:00"
---
`
	assert.Equal(t, expectedContent, string(content))
}

func TestSanitizePath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"/path/to/file.md", "path_to_file"},
		{"path/to/file.md", "path_to_file"},
		{"/path/to/file", "path_to_file"},
		{"path/to/file", "path_to_file"},
		{"/file.md", "file"},
		{"file.md", "file"},
		{"/file", "file"},
		{"file", "file"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := sanitizePath(test.input)
			if result != test.expected {
				t.Errorf("sanitizePath(%q) = %q; want %q", test.input, result, test.expected)
			}
		})
	}
}
