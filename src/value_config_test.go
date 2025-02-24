package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Readme1Contents and Readme2Contents are used in multiple tests in this file.
const Readme1Contents = `
---
title: "README1"
path: "readme1"
---
`

const Readme2Contents = `
---
title: "README2"
path: "readme2"
---
`

const TestCaseForDetailCheckMarkdown = `
version: 1
project:
  name: "Test Project"
pages:
  - markdown: "README1.md"
    path: "readme1"
    title: "README1"
    updated_at: "2021-01-01T00:00:00Z"
  - markdown: "README2.md"
assets:
  - "assets/**"
  - "images/**"
`

func TestParseConfigDetailsMarkdown(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	createTempFile(t, dir, "README1.md")
	readme2 := createTempFile(t, dir, "README2.md")
	os.WriteFile(readme2, []byte(Readme2Contents), 0o600)

	state := NewParseState("config.yaml", dir)
	conf, err := ParseConfig(state, strings.NewReader(TestCaseForDetailCheckMarkdown))
	require.NoError(t, err)

	// Check metadata
	assert.Equal(t, "1", conf.Version)

	// Check README1
	assert.Equal(t, "README1.md", conf.Pages[0].Markdown)
	assert.Equal(t, "readme1", conf.Pages[0].Path)
	assert.Equal(t, "README1", conf.Pages[0].Title)
	assert.Equal(t, "2021-01-01T00:00:00Z", conf.Pages[0].UpdatedAt.String())
	assert.True(t, conf.Pages[1].UpdatedAt.IsZero(), "UpdatedAt should be nil if there is no explicit value")

	// Check README2
	// There are no fields in the YAML file, but we can read the fields from the markdown file.
	assert.Equal(t, "README2.md", conf.Pages[1].Markdown)
	assert.Equal(t, "readme2", conf.Pages[1].Path)
	assert.Equal(t, "README2", conf.Pages[1].Title)

	// Check assets
	assert.Equal(t, "assets/**", string(conf.Assets[0]))
	assert.Equal(t, "images/**", string(conf.Assets[1]))
}

const TestCaseForDetailCheckMatch = `
version: 1
project:
  name: "Test Project"
pages:
  - match: "./*.md"
`

func TestParseConfigDetailsMatch(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	createTempFile(t, dir, "README1.md")
	readme2 := createTempFile(t, dir, "README2.md")
	os.WriteFile(readme2, []byte(Readme2Contents), 0o600)

	state := NewParseState("config.yaml", dir)
	conf, err := ParseConfig(state, strings.NewReader(TestCaseForDetailCheckMarkdown))
	require.NoError(t, err)

	// Check metadata
	assert.Equal(t, "1", conf.Version)

	// Check README1
	assert.Equal(t, "README1.md", conf.Pages[0].Markdown)
	assert.Equal(t, "readme1", conf.Pages[0].Path)
	assert.Equal(t, "README1", conf.Pages[0].Title)

	// Check README2
	assert.Equal(t, "README2.md", conf.Pages[1].Markdown)
	assert.Equal(t, "readme2", conf.Pages[1].Path)
	assert.Equal(t, "README2", conf.Pages[1].Title)
}

// Valid Case.
const TestCaseParseMarkdown = `
version: 1
project:
  name: "Test Project"
pages:
  - markdown: "README2.md"
    path: "readme2"
    title: "README2"
`

// Invalid Case with a sort order without a sort_key.
const TestCaseWithMatchPage = `
version: 1
project:
  name: "Test Project"
pages:
	- match: "./docs/*.md"
	  sort_key: "title"
	  sort_order: "asc"
assets:
	- "assets/**"
`

// Invalid Case with an unknown version.
const TestCaseWithUnknownVersion = `
version: 2
project:
	name: "Test Project"
pages:
	- markdown: "README2.md"
`

// Invalid Case with an empty project name.
const TestCaseWithEmptyProjectName = `
version: 1
project:
	name: ""
	description: "This is a test project."
	version: "1.0.0"
pages:
	- markdown: "README2.md"
`

const TestCaseWithUnknownField = `
version: 1
project:
	name: "Test Project"
pages:
	- markdown: "README2.md"
unknown: "unknown"
`

// Invalid Case with unknown page type in the pages field.
const TestCaseWithUnknownPageType = `
version: 1
project:
  name: "Test Project"
pages:
	- path: "readme2"
	  title: "README2"
`

// Invalid Case with children in the markdown item.
// Can't use children in the markdown item.
const TestCaseWithBrokenMarkdownPage = `
version: 1
project:
  name: "Test Project"
pages:
  - markdown: "README1.md"
    path: "readme1"
    title: "README1"
    children:
      - markdown: "./README1.md"
        path: "./another"
        title: "./ANOTHER"
`

// Invalid Case with unknown date format in the pages field.
const TestCaseWithUnknownUpdatedAtFormat = `
version: 1
project:
  name: "Test Project"
pages:
  - markdown: "README2.md"
    path: "readme2"
    title: "README2"
		updated_at: "23/1/2024
`

// Invalid case with multiple assets fields.
const TestCaseWithMultipleAssets = `
version: 1
project:
  name: "Test Project"
pages:
	- markdown: "README2.md"
	  path: "readme2"
	  title: "README2"
	  updated_at: "2021-01-01T00:00:00Z"
assets:
	- "assets/**"
assets:
	- "assets/**"
`

// Invalid Case with a sort order without a sort_key.
const TestCaseWithSortOrderWithoutSortKey = `
version: 1
project:
  name: "Test Project"
pages:
	- match: "./docs/*.md"
	  sort_order: "asc"
assets:
	- "assets/**"
`

// Directory Traversal Attack.
const TestCaseWithDirectoryTraversal1 = `
version: 1
project:
  name: "Test Project"
pages:
  - markdown: "../forbid/README1.md"
    path: "readme1"
    title: "README1"
`

// Directory Traversal Attack.
const TestCaseWithDirectoryTraversal2 = `
version: 1
project:
  name: "Test Project"
pages:
  - markdown: "./.././forbid/README1.md"
    path: "readme2"
    title: "README2"
`

// Directory Traversal Attack.
const TestCasePageDirectoryTraversal3 = `
version: 1
project:
  name: "Test Project"
pages:
  - match: "../forbid/*.md"
`

// Directory Traversal Attack.
const TestCasePageDirectoryTraversal4 = `
version: 1
project:
  name: "Test Project"
pages:
  - markdown: "README1.md"
    path: "readme1"
    title: "README1"
  - match: "./.././forbid/*.md"
`

func TestParseConfig(t *testing.T) {
	// Create a temporary directory.
	dir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	// Create a working directory.
	workingDir := filepath.Join(dir, "working")
	require.NoError(t, os.Mkdir(workingDir, 0o755))
	f := createTempFile(t, workingDir, "README1.md")
	require.NoError(t, os.WriteFile(f, []byte(Readme1Contents), 0o600))
	f = createTempFile(t, workingDir, "README2.md")
	require.NoError(t, os.WriteFile(f, []byte(Readme2Contents), 0o600))

	// Create a temporary directory to test the directory traversal attack.
	forbidDir := filepath.Join(dir, "forbid")
	require.NoError(t, os.Mkdir(forbidDir, 0o755))
	f = createTempFile(t, forbidDir, "README1.md")
	require.NoError(t, os.WriteFile(f, []byte(Readme1Contents), 0o600))

	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid config with markdown",
			input:    TestCaseParseMarkdown,
			expected: true,
		},
		{
			name:     "valid config with match",
			input:    TestCaseForDetailCheckMatch,
			expected: true,
		},
		{
			name:     "invalid config: unknown version",
			input:    TestCaseWithUnknownVersion,
			expected: false,
		},
		{
			name:     "invalid config: empty project name",
			input:    TestCaseWithEmptyProjectName,
			expected: false,
		},
		{
			name:     "invalid config: unknown field",
			input:    TestCaseWithUnknownField,
			expected: false,
		},
		{
			name:     "invalid config: unknown page type in the `pages` field",
			input:    TestCaseWithUnknownPageType,
			expected: false,
		},
		{
			name:     "invalid config: children in the markdown item",
			input:    TestCaseWithBrokenMarkdownPage,
			expected: false,
		},
		{
			name:     "invalid config: unknown date format in the `updated_at` field",
			input:    TestCaseWithUnknownUpdatedAtFormat,
			expected: false,
		},
		{
			name:     "invalid config: multiple assets fields",
			input:    TestCaseWithMultipleAssets,
			expected: false,
		},
		{
			name:     "invalid config: sort order without a sort key",
			input:    TestCaseWithSortOrderWithoutSortKey,
			expected: false,
		},
		{
			name:     "invalid config: including directory traversal1",
			input:    TestCaseWithDirectoryTraversal1,
			expected: false,
		},
		{
			name:     "invalid config: including directory traversal2",
			input:    TestCaseWithDirectoryTraversal2,
			expected: false,
		},
		{
			name:     "invalid config: including directory traversal3",
			input:    TestCasePageDirectoryTraversal3,
			expected: false,
		},
		{
			name:     "invalid config: including directory traversal4",
			input:    TestCasePageDirectoryTraversal4,
			expected: false,
		},
	}

	for _, tc := range testCases {
		testCase := tc
		t.Run(tc.name, func(t *testing.T) {
			// DO NOT parallelize this test.
			// If you parallelize this test, the temp file cleanup will occur before testing.
			state := NewParseState("config.yaml", workingDir)
			_, err := ParseConfig(state, strings.NewReader(testCase.input))
			if testCase.expected {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestConfigAsset(t *testing.T) {
	dir, err := os.MkdirTemp("", "")
	require.NoError(t, err)

	createTempFile(t, dir, "image1.png")
	createTempFile(t, dir, "image2.png")
	createTempFile(t, dir, "image3.jpg")

	c := ConfigAsset("image*.png")
	ls, err := c.List(dir)
	require.NoError(t, err)

	require.Len(t, ls, 2)
	assert.Contains(t, ls, filepath.Join(dir, "image1.png"))
	assert.Contains(t, ls, filepath.Join(dir, "image2.png"))
	assert.NotContains(t, ls, filepath.Join(dir, "image3.jpg"))
}

func TestDirectoryTraversal(t *testing.T) {
	dir, err := os.MkdirTemp("", "")
	require.NoError(t, err)

	subdir := filepath.Join(dir, "subdir")
	err = os.Mkdir(subdir, 0o700)
	require.NoError(t, err)

	createTempFile(t, dir, "image1.png")
	createTempFile(t, subdir, "image2.png")

	c := ConfigAsset("../../image*.png")
	_, err = c.List(subdir)
	require.Error(t, err)
}

func createTempFile(t *testing.T, dir, path string) string {
	t.Helper()
	f, err := os.Create(filepath.Join(dir, path))
	require.NoError(t, err)
	defer f.Close()
	return f.Name()
}
