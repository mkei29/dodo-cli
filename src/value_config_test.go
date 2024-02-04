package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const TestCaseValid = `
version: 1
page:
  - filepath: "README.md"
    name: "readme"
    title: "README"
  - match: "docs/*.md"
  - match: "docs/**.md"
`

func TestValidCase(t *testing.T) {
	t.Parallel()
	conf, err := ParseDocumentConfig(strings.NewReader(TestCaseValid))
	require.NoError(t, err)
	// conf := src.ParseDocumentConfig(TestCase1)
	require.Equal(t, conf.Version, "1")
	require.Equal(t, *conf.Page[0].Filepath, "README.md")
	require.Equal(t, *conf.Page[0].Name, "readme")
	require.Equal(t, *conf.Page[0].Title, "README")
}
