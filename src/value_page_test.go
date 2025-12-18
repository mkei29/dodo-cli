package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/toritoritori29/dodo-cli/src/config"
)

func prepareTempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	return dir
}

func prepareSubDir(t *testing.T, rootDir, subDir string) string {
	t.Helper()
	dir := filepath.Join(rootDir, subDir)
	err := os.Mkdir(dir, 0o755)
	require.NoError(t, err)
	return dir
}

func prepareFile(t *testing.T, rootDir, filename, content string) {
	t.Helper()
	filepath := filepath.Join(rootDir, filename)
	file, err := os.Create(filepath)
	require.NoError(t, err)
	defer file.Close()
	file.WriteString(content)
}

const TestCaseCreatePageWithMarkdown = `
version: 1
project:
  project_id: "project_id"
  name: "Test Project"
  default_language: "en"
pages:
  - markdown: "README1.md"
    path: "readme1"
    title: "README2"
    updated_at: "2021-01-01T00:00:00Z"
  - markdown: "README2.md"
    path: "readme2"
    title: "README2"
`

func TestCreatePageTreeWithMarkdown(t *testing.T) {
	t.Parallel()
	dir := prepareTempDir(t)
	prepareFile(t, dir, "README1.md", "content")
	prepareFile(t, dir, "README2.md", "content")

	state := config.NewParseStateV1("config.yaml", dir)
	conf, err := config.ParseConfigV1(state, strings.NewReader(TestCaseCreatePageWithMarkdown))
	require.NoError(t, err)

	page, merr := CreatePageTree(conf, dir)
	require.Nil(t, merr, "CreatePageTree should not return error if the valid case is specified")

	assert.Equal(t, "RootNode", page.Type)
	assert.Len(t, page.Children, 2)

	page1 := page.Children[0]
	assert.Equal(t, "LeafNode", page1.Type)
	assert.Equal(t, "readme1", page1.Path)
	assert.Equal(t, "README2", page1.Title)
	assert.Equal(t, "", page1.Description)
	assert.Equal(t, "en", page1.Language[0].Language)
	assert.Equal(t, "README2", page1.Language[0].Title)
	assert.Equal(t, "", page1.Language[0].Description)
	assert.Equal(t, "2021-01-01T00:00:00Z", page1.UpdatedAt.String())

	page2 := page.Children[1]
	assert.Equal(t, "LeafNode", page2.Type)
	assert.Equal(t, "readme2", page2.Path)
	assert.Equal(t, "README2", page2.Title)
	assert.Equal(t, "", page2.Description)
	assert.Equal(t, "en", page2.Language[0].Language)
	assert.Equal(t, "README2", page2.Language[0].Title)
	assert.Equal(t, "", page2.Language[0].Description)
	assert.Equal(t, "", page2.UpdatedAt.String())
}

const TestCaseCreatePageTreeMatch = `
version: 1
project:
  project_id: "project_id"
  name: "project"
pages:
  - match: "./**/README*.md"
    sort_key: "title"
    sort_order: "asc"
`

func TestCreatePageTreeWithMatch(t *testing.T) {
	t.Parallel()

	// Create files
	dir := prepareTempDir(t)
	prepareFile(t, dir, "README1.md", `
	---
  title: README1
  path: readme1
	---
	`)
	prepareFile(t, dir, "README2.md", `
	---
  title: README2
  path: readme2
	---
	`)

	sub := prepareSubDir(t, dir, "docs")
	prepareFile(t, dir, "README3.md", `
	---
  title: README3
  path: readme3
	---
	`)
	prepareFile(t, sub, "README4.md", `
	---
  title: README4
  path: readme4
	---
	`)
	prepareFile(t, sub, "README5.md", `
	---
  title: README5
  path: readme5
	---
	`)

	state := config.NewParseStateV1("config.yaml", dir)
	conf, err := config.ParseConfigV1(state, strings.NewReader(TestCaseCreatePageTreeMatch))
	require.NoError(t, err, "should not return error")

	page, merr := CreatePageTree(conf, dir)
	require.Nil(t, merr, "CreatePageTree should not return error")

	// Root node should have 2 children
	require.Len(t, page.Children, 5, "root node should have 5 children")

	// Check the first child
	page1 := page.Children[0]
	assert.Equal(t, "LeafNode", page1.Type)
	assert.Equal(t, "readme1", page1.Path)
	assert.Equal(t, "README1", page1.Title)
	assert.Equal(t, "en", page1.Language[0].Language)
	assert.Equal(t, "README1", page1.Language[0].Title)
	assert.Equal(t, "", page1.Language[0].Description)
	assert.Equal(t, "", page1.Description)
	assert.Equal(t, "", page1.UpdatedAt.String())

	// Check the remaining children
	assert.Equal(t, "README2", page.Children[1].Title)
	assert.Equal(t, "README3", page.Children[2].Title)
	assert.Equal(t, "README4", page.Children[3].Title)
	assert.Equal(t, "README5", page.Children[4].Title)
}

const TestCaseCreatePageHybridCase = `
version: 1
project:
  project_id: "project_id"
  name: "Test Project"
pages:
  - markdown: "README.md"
  - match: "*.md"
`

func TestCreatePageTreeWithHybridCase(t *testing.T) {
	t.Parallel()

	dir := prepareTempDir(t)
	prepareFile(t, dir, "README.md", `
	---
  title: README
  path: readme
	---
	`)

	state := config.NewParseStateV1("config.yaml", dir)
	conf, err := config.ParseConfigV1(state, strings.NewReader(TestCaseCreatePageHybridCase))
	require.NoError(t, err, "should not return error")

	page, merr := CreatePageTree(conf, dir)
	require.Nil(t, merr, "CreatePageTree should not return error if the valid case is specified: %w", err)
	require.Len(t, page.Children, 2, "root node should have 4 children")

	page1 := page.Children[0]
	assert.Equal(t, "LeafNode", page1.Type)
	assert.Equal(t, "README", page1.Title)
	assert.Equal(t, "", page1.Description)
	assert.Equal(t, "en", page1.Language[0].Language)
	assert.Equal(t, "README", page1.Language[0].Title)
	assert.Equal(t, "", page1.Language[0].Description)
	assert.Equal(t, "readme", page1.Path)

	page2 := page.Children[1]
	assert.Equal(t, "LeafNode", page2.Type)
	assert.Equal(t, "README", page2.Title)
	assert.Equal(t, "", page2.Description)
	assert.Equal(t, "en", page2.Language[0].Language)
	assert.Equal(t, "README", page2.Language[0].Title)
	assert.Equal(t, "", page2.Language[0].Description)
	assert.Equal(t, "readme", page2.Path)
}

const TestCaseCreatePageWithDirectory = `
version: 1
project:
  project_id: "project_id"
  name: "project"
pages:
  - directory: "directory"
    children:
      - markdown: "README1.md"
        path: "readme1"
        title: "README1"
`

func TestCreatePageTreeWithDirectory(t *testing.T) {
	t.Parallel()

	dir := prepareTempDir(t)
	prepareFile(t, dir, "README1.md", `
	---
  title: README1
  path: readme1
	---
	`)

	state := config.NewParseStateV1("config.yaml", dir)
	conf, err := config.ParseConfigV1(state, strings.NewReader(TestCaseCreatePageWithDirectory))
	require.NoError(t, err)

	page, merr := CreatePageTree(conf, dir)
	require.Nil(t, merr, "CreatePageTree should not return error if the valid case is specified: %v", err)

	require.Len(t, page.Children, 1)

	dir1 := page.Children[0]
	assert.Equal(t, "DirNodeWithoutPage", dir1.Type)
	assert.Equal(t, "directory", dir1.Title)
	assert.Equal(t, "en", dir1.Language[0].Language)
	assert.Equal(t, "directory", dir1.Language[0].Title)
	assert.Equal(t, "", dir1.Language[0].Description)

	page1 := dir1.Children[0]
	assert.Equal(t, "README1", page1.Title)
	assert.Equal(t, "readme1", page1.Path)
	assert.Equal(t, "en", page1.Language[0].Language)
	assert.Equal(t, "README1", page1.Language[0].Title)
	assert.Equal(t, "", page1.Language[0].Description)
}

// Valid case.
const TestCasePageValid1 = `
version: 1
project:
  project_id: "project_id"
  name: "project"
  default_language: "ja"
pages:
  - markdown: "README1.md"
    path: "readme1"
    title: "README1"
  - markdown: "README2.md"
    path: "readme2"
    title: "README2"
`

const TestCasePageValid2 = `
version: 1
project:
  project_id: "project_id"
  name: "project"
pages:
  - directory: "DIR1"
    children:
      - markdown: "README1.md"
        path: "readme1"
        title: "README1"
  - directory: "DIR2"
    children:
      - markdown: "README1.md"
        path: "readme1"
        title: "README1"
`

// Invalid Case: Paths are duplicated.
const TestCasePageInvalid1 = `
version: 1
project:
  project_id: "project_id"
  name: "project"
pages:
  - markdown: "README1.md"
    path: "readme1"
    title: "README1"
  - markdown: "README1.md"
    path: "readme1"
    title: "README1"
`

// Invalid Case: Duplicated path under the same parent.
const TestCasePageInvalid2 = `
version: 1
project:
  project_id: "project_id"
  name: "project"
pages:
  - directory: "DIR1"
    children:
      - markdown: "README1.md"
        path: "readme1"
        title: "README1"
      - markdown: "README1.md"
        path: "readme1"
        title: "README1"
`

func TestIsValid(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name    string
		content string
		isValid bool
	}{
		{
			"valid content",
			TestCasePageValid1,
			true,
		},
		{
			"invalid: same path but different parent",
			TestCasePageValid2,
			false,
		},
		{
			"invalid: page has duplicated paths in the different parent",
			TestCasePageInvalid1,
			false,
		},
		{
			"invalid: page has duplicated paths under the same parent",
			TestCasePageInvalid2,
			false,
		},
	}

	for _, tt := range testCases {
		c := tt
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			dir := prepareTempDir(t)
			prepareFile(t, dir, "README1.md", "")
			prepareFile(t, dir, "README2.md", "")

			state := config.NewParseStateV1("config.yaml", dir)
			conf, err := config.ParseConfigV1(state, strings.NewReader(c.content))
			require.NoError(t, err, "should not return error: %v", err)
			page, merr := CreatePageTree(conf, dir)
			require.Nil(t, merr, "CreatePageTree should not failed if the valid case is specified: %v", err)

			merr = page.IsValid()
			if c.isValid {
				require.Nil(t, merr, "should not return error: %v", merr)
			} else {
				require.NotNil(t, merr, "should return error")
			}
		})
	}
}
