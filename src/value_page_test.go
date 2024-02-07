package main

import (
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

func TestNewPageFromConfigOnlySinglePage(t *testing.T) {
	t.Parallel()
	conf, err := ParseDocumentConfig(strings.NewReader(TestCasePage1))
	require.NoError(t, err)

	page, err := NewPageFromConfig(*conf, "./")
	require.NoError(t, err)
	assert.Equal(t, page.Path, "")
	assert.Equal(t, page.Title, "")
	assert.Len(t, page.Children, 2)
}

const TestCasePage2 = `
version: 1
pages:
  - match: "README*.md"
  - match: "docs/**/*.md"
`

func TestNewPageFromConfigWithPattern(t *testing.T) {
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

	page, err := NewPageFromConfig(*conf, dir)
	require.NoError(t, err)
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
  - match: "docs/**/*.md"
`

func TestNewPageFromConfigWithHybridCase(t *testing.T) {
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

	page, err := NewPageFromConfig(*conf, dir)
	require.NoError(t, err)
	assert.Equal(t, page.Path, "")
	assert.Equal(t, page.Title, "")
	assert.Len(t, page.Children, 7)
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
`

const TestCasePageMalicious4 = `
version: 1
pages:
  - filepath: "README2.md"
    path: "readme1"
    title: "README2"
  - match: "./dir1/../../**/*.md"
`

func TestNewPageFromConfigWithMaliciousFilepath(t *testing.T) {
	dir, err := os.MkdirTemp("", "test_dir")
	require.NoError(t, err)

	cases := []string{
		TestCasePageMalicious1,
		TestCasePageMalicious2,
		TestCasePageMalicious3,
		TestCasePageMalicious4,
	}

	for _, c := range cases {
		conf, err := ParseDocumentConfig(strings.NewReader(c))
		require.NoError(t, err, "should not return error")

		_, err = NewPageFromConfig(*conf, dir)
		require.Error(t, err, "should return error if malicious config is given")
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

			page, err := NewPageFromFrontMatter(path)
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
			page, err := NewPageFromConfig(*conf, "")
			require.NoError(t, err, "should return error if malicious config is given")
			assert.Equal(t, c.isValid, page.IsValid())
		})
	}
}
