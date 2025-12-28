package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/toritoritori29/dodo-cli/src/config"
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

	// Verify content by parsing frontmatter
	matter, err := config.NewFrontMatterFromMarkdown(path)
	require.NoError(t, err)
	assert.Equal(t, "Test Title", matter.Title)
	assert.Equal(t, "test-path", matter.Path)
	assert.Equal(t, "", matter.Description)
	assert.Equal(t, "2025-01-01T00:00:00+09:00", matter.CreatedAt.String())
	assert.Equal(t, "2025-01-01T00:00:00+09:00", matter.UpdatedAt.String())
	assert.NotEmpty(t, matter.LanguageGroupID, "language_group_id should be auto-generated")

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

	// Verify updated content
	updatedMatter, err := config.NewFrontMatterFromMarkdown(path)
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", updatedMatter.Title)
	assert.Equal(t, "updated-path", updatedMatter.Path)
	assert.Equal(t, "", updatedMatter.Description)
	assert.Equal(t, "2025-01-01T00:00:00+09:00", updatedMatter.CreatedAt.String())
	assert.Equal(t, "2025-01-02T00:00:00+09:00", updatedMatter.UpdatedAt.String())
	assert.Equal(t, matter.LanguageGroupID, updatedMatter.LanguageGroupID, "language_group_id should be preserved")
}

func TestSanitizePath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"./path/to/file.md", "path_to_file"},
		{"/path/to/file.md", "path_to_file"},
		{"path/to/file.md", "path_to_file"},
		{"path/to/1234.md", "path_to_1234"},
		{"/path/to/file", "path_to_file"},
		{"path/to/file", "path_to_file"},
		{"/file.md", "file"},
		{"file.md", "file"},
		{"/file", "file"},
		{"file", "file"},
		{"(file)", "file"},
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
