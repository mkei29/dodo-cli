package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const TestCaseValid = `
version: 1
pages:
  - filepath: "README.md"
    path: "readme"
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
	require.Equal(t, *conf.Pages[0].Filepath, "README.md")
	require.Equal(t, *conf.Pages[0].Path, "readme")
	require.Equal(t, *conf.Pages[0].Title, "README")
}
