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

const TestCasePage1 = `
version: 1
pages:
  - filepath: "README1.md"
    path: "readme1"
    title: "README2"
  - filepath: "README2.md"
    path: "readme1"
    title: "README2"
`

func TestCreatePageTreeOnlySinglePage(t *testing.T) {
	t.Parallel()
	conf, err := ParseDocumentConfig(strings.NewReader(TestCasePage1))
	require.NoError(t, err)

	page, es := CreatePageTree(*conf, "./")
	require.False(t, es.HasError())
	assert.Equal(t, page.Path, "")
	assert.Equal(t, page.Title, "")
	assert.Len(t, page.Children, 2)
}

const TestCasePage2 = `
version: 1
pages:
  - match: "README*.md"
    title: "section1"
  - match: "docs/**/*.md"
    title: "section1"
`

func TestCreatePageTreeWithPattern(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "test_dir")
	require.NoError(t, err)

	// Create files
	os.Create(filepath.Join(dir, "README1.md"))
	os.Create(filepath.Join(dir, "README2.md"))
	os.Create(filepath.Join(dir, "README3.md"))

	os.Mkdir(filepath.Join(dir, "docs"), 0755)
	os.Create(filepath.Join(dir, "docs", "README1.md"))
	os.Create(filepath.Join(dir, "docs", "README2.md"))

	conf, err := ParseDocumentConfig(strings.NewReader(TestCasePage2))
	require.NoError(t, err, "should not return error")

	page, es := CreatePageTree(*conf, dir)
	require.False(t, es.HasError())
	assert.Equal(t, page.Path, "")
	assert.Equal(t, page.Title, "")
	assert.Len(t, page.Children, 5)
}

const TestCasePage3 = `
version: 1
pages:
  - filepath: "README.md"
    path: "readme1"
    title: "README2"
  - filepath: "README2.md"
    path: "readme1"
    title: "README2"
  - match: "README*.md"
    title: "section"
  - match: "docs/**/*.md"
    title: "section"
`

func TestCreatePageTreeWithHybridCase(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "test_dir")
	require.NoError(t, err)

	// Create files
	os.Create(filepath.Join(dir, "README1.md"))
	os.Create(filepath.Join(dir, "README2.md"))
	os.Create(filepath.Join(dir, "README3.md"))

	os.Mkdir(filepath.Join(dir, "docs"), 0755)
	os.Create(filepath.Join(dir, "docs", "README1.md"))
	os.Create(filepath.Join(dir, "docs", "README2.md"))

	conf, err := ParseDocumentConfig(strings.NewReader(TestCasePage3))
	require.NoError(t, err, "should not return error")

	page, es := CreatePageTree(*conf, dir)
	require.False(t, es.HasError())
	assert.Equal(t, page.Path, "")
	assert.Equal(t, page.Title, "")
	assert.Len(t, page.Children, 7)
}

const TestCasePage4 = `
version: 1
pages:
  - filepath: "dir1.md"
    path: "dir1"
    title: "dir1"
    children:
      - filepath: "README1.md"
        path: "readme1"
        title: "README1"
`

func TestCreatePageTreeLayeredCase(t *testing.T) {
	t.Parallel()
	conf, err := ParseDocumentConfig(strings.NewReader(TestCasePage4))
	require.NoError(t, err)

	page, es := CreatePageTree(*conf, "./")
	require.False(t, es.HasError())
	assert.Equal(t, page.Path, "")
	assert.Equal(t, page.Title, "")
	assert.Len(t, page.Children, 1)

	assert.Equal(t, page.Children[0].Path, "dir1")
	assert.Equal(t, page.Children[0].Children[0].Path, "dir1/readme1")
}

const TestCasePageMalicious1 = `
version: 1
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
pages:
  - filepath: "README2.md"
    path: "readme1"
    title: "README2"
  - match: "../**/*.md"
    title: "section"
`

const TestCasePageMalicious4 = `
version: 1
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
	}

	for i, c := range cases {
		testID := i + 1
		testCase := c
		t.Run(fmt.Sprintf("pass when malicious filepath was given. ID: %d", testID), func(t *testing.T) {
			conf, err := ParseDocumentConfig(strings.NewReader(testCase))
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
			defer file.Close()

			require.NoError(t, err)
			_, err = file.WriteString(c.content)
			require.NoError(t, err)

			page, err := NewPageFromFrontMatter(path, "")
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

const TestCasePageValid1 = `
version: 1
pages:
  - filepath: "README1.md"
    path: "readme1"
    title: "README1"
  - filepath: "README2.md"
    path: "readme2"
    title: "README2"
`

const TestCasePageInvalid1 = `
version: 1
pages:
  - filepath: "README1.md"
    path: "readme1"
    title: "README1"
  - filepath: "README1.md"
    path: "readme1"
    title: "README1"
`

const TestCasePageInvalid2 = `
version: 1
pages:
  - filepath: "README1.md"
    title: "README1"
  - filepath: "README1.md"
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
			"pass when valid content was given",
			TestCasePageValid1,
			true,
		},
		{
			"pass when page has duplicated path",
			TestCasePageInvalid1,
			false,
		},
		{
			"pass when some page doesn't have necessary fields",
			TestCasePageInvalid1,
			false,
		},
	}
	for _, tt := range testCases {
		c := tt
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			conf, err := ParseDocumentConfig(strings.NewReader(c.content))
			require.NoError(t, err, "should not return error")
			page, es := CreatePageTree(*conf, "")
			require.False(t, es.HasError(), "should not return error if valid config is given")

			es = page.IsValid()
			assert.Equal(t, c.isValid, !es.HasError())
		})
	}
}
