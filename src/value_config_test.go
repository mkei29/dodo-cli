package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const TestCaseValid = `
version: 1
project:
  name: "Test Project"
pages:
  - markdown: "README1.md"
    path: "readme1"
    title: "README1"
    updated_at: "2021-01-01T00:00:00Z"
  - markdown: "README2.md"
    path: "readme1"
    title: "README1"
  - match: "docs/*.md"
  - match: "docs/**.md"
assets:
  - "assets/**"
  - "images/**"
`

func createTempFile(t *testing.T, dir, path string) {
	t.Helper()
	f, err := os.Create(filepath.Join(dir, path))
	require.NoError(t, err)
	defer f.Close()
}

func TestParseConfigDetails(t *testing.T) {
	t.Parallel()
	conf, err := ParseConfig("config.yaml", strings.NewReader(TestCaseValid))
	require.NoError(t, err)

	// Check metadata
	assert.Equal(t, "1", conf.Version)

	// Check pages
	assert.Equal(t, "README1.md", conf.Pages[0].Markdown)
	assert.Equal(t, "readme1", conf.Pages[0].Path)
	assert.Equal(t, "README1", conf.Pages[0].Title)
	assert.Equal(t, "2021-01-01T00:00:00Z", conf.Pages[0].UpdatedAt.String())
	assert.True(t, conf.Pages[1].UpdatedAt.IsNull(), "UpdatedAt should be nil if there is no explicit value")

	// Check assets
	assert.Equal(t, "assets/**", string(conf.Assets[0]))
	assert.Equal(t, "images/**", string(conf.Assets[1]))
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
	- sort_key: "title"
	- sort_order: "asc"
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
	- title: "README2"
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
	- path: "readme2"
	- title: "README2"
	- updated_at: "2021-01-01T00:00:00Z"
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
	- sort_order: "asc"
assets:
	- "assets/**"
`

func TestParseConfig(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid config with markdown",
			input:    TestCaseCreatePageWithMarkdown,
			expected: true,
		},
		{
			name:     "valid config with match",
			input:    TestCaseCreatePageTreeMatch,
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
	}

	for _, tc := range testCases {
		testCase := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := ParseConfig("config.yaml", strings.NewReader(testCase.input))
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
