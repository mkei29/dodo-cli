package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const TestCasePage1 = `
version: 1
page:
  - filepath: "README1.md"
    name: "readme1"
    title: "README2"
  - filepath: "README2.md"
    name: "readme1"
    title: "README2"
`

func TestNewPageFromConfig(t *testing.T) {
	t.Parallel()
	conf, err := ParseDocumentConfig(strings.NewReader(TestCasePage1))
	require.NoError(t, err)

	page, err := NewPageFromConfig(*conf)
	require.NoError(t, err)
	assert.Equal(t, page.Path, "")
	assert.Equal(t, page.Title, "")
	assert.Len(t, page.Children, 2)
}
