package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// Valid Case.
const TestCaseParseConfig1 = `
version: 1
pages:
  - markdown: "README2.md"
    path: "readme1"
    title: "README2"
`

// Invalid Case with unknown date format in the pages field.
const TestCaseParseConfig2 = `
version: 1
pages:
  - markdown: "README2.md"
    path: "readme1"
    title: "README2"
		updated_at: "23/1/2024
`

func TestParseConfig(t *testing.T) {
	t.Parallel()

	t.Run("should not return error when valid config was given", func(t *testing.T) {
		t.Parallel()
		dir := prepareTempDir(t)
		prepareFile(t, dir, "README1.md", "content")

		_, err := ParseConfig(strings.NewReader(TestCaseParseConfig1))
		require.NoError(t, err)
	})
	t.Run("should return error when invalid config was given", func(t *testing.T) {
		t.Parallel()
		_, err := ParseConfig(strings.NewReader(TestCaseParseConfig2))
		require.Error(t, err)
	})
}

const TestCaseCreatePageWithMarkdown = `
version: 1
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

	conf, err := ParseConfig(strings.NewReader(TestCaseCreatePageWithMarkdown))
	require.NoError(t, err)

	page, es := CreatePageTree(*conf, dir)
	require.False(t, es.HasError())
	assert.Equal(t, "RootNode", page.Type)
	assert.Len(t, page.Children, 2)

	page1 := page.Children[0]
	assert.Equal(t, "readme1", page1.Path)
	assert.Equal(t, "README2", page1.Title)
	assert.Equal(t, "", page1.Description)
	assert.Equal(t, "2021-01-01T00:00:00Z", page1.UpdatedAt.String())

	page2 := page.Children[1]
	assert.Equal(t, "readme2", page2.Path)
	assert.Equal(t, "README2", page2.Title)
	assert.Equal(t, "", page2.Description)
	assert.Equal(t, "", page2.UpdatedAt.String())
}

const TestCaseCreatePageTreeMatch = `
version: 1
pages:
  - match: "./**/README*.md"
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

	conf, err := ParseConfig(strings.NewReader(TestCaseCreatePageTreeMatch))
	require.NoError(t, err, "should not return error")

	page, es := CreatePageTree(*conf, dir)
	require.False(t, es.HasError(), "should not return error when valid config is given")

	// Root node should have 2 children
	assert.Len(t, page.Children, 5, "root node should have 5 children")
}

const TestCaseCreatePageHybridCase = `
version: 1
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

	conf, err := ParseConfig(strings.NewReader(TestCaseCreatePageHybridCase))
	require.NoError(t, err, "should not return error")

	page, es := CreatePageTree(*conf, dir)
	require.False(t, es.HasError())
	assert.Len(t, page.Children, 2, "root node should have 4 children")

	page1 := page.Children[0]
	assert.Equal(t, "LeafNode", page1.Type)
	assert.Equal(t, "readme", page1.Path)
	assert.Equal(t, "", page1.Description)

	page2 := page.Children[1]
	assert.Equal(t, "LeafNode", page2.Type)
	assert.Equal(t, "readme", page2.Path)
	assert.Equal(t, "", page2.Description)
}

const TestCaseCreatePageWithDirectory = `
version: 1
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

	conf, err := ParseConfig(strings.NewReader(TestCaseCreatePageWithDirectory))
	require.NoError(t, err)

	page, es := CreatePageTree(*conf, dir)
	require.False(t, es.HasError())
	assert.Len(t, page.Children, 1)

	dir1 := page.Children[0]
	assert.Equal(t, "DirNodeWithoutPage", dir1.Type)
	assert.Equal(t, "directory", dir1.Title)

	page1 := dir1.Children[0]
	assert.Equal(t, "README1", page1.Title)
	assert.Equal(t, "readme1", page1.Path)
}

const TestCasePageMalicious1 = `
version: 1
index:
  title: "root"
  filepath: "./README.md"
pages:
  - filepath: "../README.md"
    path: "readme1"
    title: "README2"
  - filepath: "README2.md"
    path: "readme1"
    title: "README2"
`

const TestCasePageMalicious2 = `
version: 1
index:
  title: "root"
  filepath: "./README.md"
pages:
  - filepath: "README2.md"
    path: "readme1"
    title: "README2"
  - filepath: "./dir1/.././../confidential"
    path: "readme1"
    title: "README2"
`

const TestCasePageMalicious3 = `
version: 1
index:
  title: "root"
  filepath: "./README.md"
pages:
  - filepath: "README2.md"
    path: "readme1"
    title: "README2"
  - match: "../**/*.md"
    title: "section"
`

const TestCasePageMalicious4 = `
version: 1
index:
  title: "root"
  filepath: "./README.md"
pages:
  - filepath: "README2.md"
    path: "readme1"
    title: "README2"
  - match: "./dir1/../../**/*.md"
    title: "section"
`

// No index field.
const TestCasePageMalicious5 = `
version: 1
index:
  title: "root"
  filepath: "./README.md"
pages:
  - filepath: "README2.md"
    path: "readme1"
    title: "README2"
  - match: "./dir1/../../**/*.md"
    title: "section"
`

// No title in index field.
const TestCasePageMalicious6 = `
version: 1
index:
  filepath: "./README.md"
pages:
  - filepath: "README2.md"
    path: "readme1"
    title: "README2"
  - match: "./dir1/../../**/*.md"
    title: "section"
`

// No filepath in index field.
const TestCasePageMalicious7 = `
version: 1
index:
  title: "root"
  filepath: "./README.md"
pages:
  - filepath: "README2.md"
    path: "readme1"
    title: "README2"
  - match: "./dir1/../../**/*.md"
    title: "section"
`

func TestCreatePageTreeWithMaliciousFilepath(t *testing.T) {
	dir, err := os.MkdirTemp("", "test_dir")
	require.NoError(t, err)

	cases := []string{
		TestCasePageMalicious1,
		TestCasePageMalicious2,
		TestCasePageMalicious3,
		TestCasePageMalicious4,
		TestCasePageMalicious5,
		TestCasePageMalicious6,
		TestCasePageMalicious7,
	}

	for i, c := range cases {
		testID := i + 1
		testCase := c
		t.Run(fmt.Sprintf("pass when malicious filepath was given. ID: %d", testID), func(t *testing.T) {
			conf, err := ParseConfig(strings.NewReader(testCase))
			require.NoError(t, err, "should not return error")

			_, es := CreatePageTree(*conf, dir)
			es.Summary()
			assert.True(t, es.HasError(), "should fail when malicious filepath was given")
		})
	}
}

const TestCaseReadPageFromFile1 = `
---
title: "title"
path: "path"
---
`

const TestCaseReadPageFromFile2 = `
---
title: "title"
---
`

const TestCaseReadPageFromFile3 = ""

func TestReadPageFromFile(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		content     string
		expectError bool
		expectPath  string
		expectTitle string
	}{
		{
			"pass when valid content with path and title was given",
			TestCaseReadPageFromFile1,
			false,
			"path",
			"title",
		},
		{
			"pass when valid content with only title was given",
			TestCaseReadPageFromFile2,
			false,
			"",
			"title",
		},
		{
			"pass when empty string was given",
			TestCaseReadPageFromFile3,
			false,
			"",
			"",
		},
	}

	for _, tt := range testCases {
		c := tt
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			dir, err := os.MkdirTemp("", "test_dir")
			require.NoError(t, err)

			path := filepath.Join(dir, "README1.md")
			file, err := os.Create(path)
			if err != nil {
				t.Fatalf("failed to create file: %v", err)
			}
			defer file.Close()

			require.NoError(t, err)
			_, err = file.WriteString(c.content)
			require.NoError(t, err)

			page, err := NewLeafNodeFromFrontMatter(path)
			if c.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, c.expectPath, page.Path)
			assert.Equal(t, c.expectTitle, page.Title)
		})
	}
}

// Valid case.
const TestCasePageValid1 = `
version: 1
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
pages:
  - directory: "DIR1"
    children:
      - markdown: "README1.md"
        path: "readme1"
        title: "README1"
  - directory: "DIR2"
    children:
      - markdown: "README2.md"
        path: "readme2"
        title: "README2"
`

// Invalid Case: Paths are duplicated.
const TestCasePageInvalid1 = `
version: 1
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

// Path field is invalid.
const TestCasePageInvalid3 = `
version: 1
pages:
  - markdown: "README1.md"
    path: "test/readme1"
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
		{
			"invalid: path field is invalid",
			TestCasePageInvalid3,
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

			conf, err := ParseConfig(strings.NewReader(c.content))
			require.NoError(t, err, "should not return error")
			page, es := CreatePageTree(*conf, dir)
			require.False(t, es.HasError(), "should not return error if valid config is given")

			es = page.IsValid()
			assert.Equal(t, c.isValid, !es.HasError())
		})
	}
}
