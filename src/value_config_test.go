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

func TestValidCase(t *testing.T) {
	t.Parallel()
	conf, err := ParseConfig(strings.NewReader(TestCaseValid))
	require.NoError(t, err)

	// Check metadata
	assert.Equal(t, "1", conf.Version)

	// Check pages
	assert.Equal(t, "README1.md", *conf.Pages[0].Markdown)
	assert.Equal(t, "readme1", *conf.Pages[0].Path)
	assert.Equal(t, "README1", *conf.Pages[0].Title)
	assert.Equal(t, "2021-01-01T00:00:00Z", conf.Pages[0].UpdatedAt.String())
	assert.Nil(t, conf.Pages[1].UpdatedAt, "CreatedAt should be nil if there is no explicit value")

	// Check assets
	assert.Equal(t, "assets/**", string(conf.Assets[0]))
	assert.Equal(t, "images/**", string(conf.Assets[1]))
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
