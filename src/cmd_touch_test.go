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

	err = executeTouchWrapper(args)
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
	err = executeTouchWrapper(updateArgs)
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
