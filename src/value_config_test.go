package main

import (
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
`

func TestValidCase(t *testing.T) {
	t.Parallel()
	conf, err := ParseConfig(strings.NewReader(TestCaseValid))
	require.NoError(t, err)

	assert.Equal(t, "1", conf.Version)
	assert.Equal(t, "README1.md", *conf.Pages[0].Markdown)
	assert.Equal(t, "readme1", *conf.Pages[0].Path)
	assert.Equal(t, "README1", *conf.Pages[0].Title)
	assert.Equal(t, "2021-01-01T00:00:00Z", conf.Pages[0].UpdatedAt.String())

	assert.Nil(t, conf.Pages[1].UpdatedAt, "CreatedAt should be nil if there is no explicit value")
}
