package config

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const v2ReadmeContents = `
---
title: "README"
link: "readme"
---
`

func TestParseConfigV2MarkdownSingleLocale(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	readme := createTempFile(t, dir, "README.md")
	require.NoError(t, os.WriteFile(readme, []byte(v2ReadmeContents), 0o600))

	input := `
version: 2
project:
  project_id: "project_id"
  name: "Test Project"
pages:
  - type: markdown
    filepath: "README.md"
`

	state := NewParseStateV2("config.yaml", dir)
	conf, err := ParseConfigV2(state, strings.NewReader(input))
	require.NoError(t, err)

	require.Len(t, conf.Pages, 1)
	assert.Equal(t, ConfigPageTypeMarkdownV2, conf.Pages[0].Type)
	require.Len(t, conf.Pages[0].Lang, 1)
	assert.Equal(t, "README.md", conf.Pages[0].Lang["en"].Filepath)
	assert.Equal(t, "README", conf.Pages[0].Lang["en"].Title)
	assert.Equal(t, "readme", conf.Pages[0].Lang["en"].Link)
}

func TestParseConfigV2MarkdownMultiLocale(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	readmeEN := createTempFile(t, dir, "README.en.md")
	require.NoError(t, os.WriteFile(readmeEN, []byte(v2ReadmeContents), 0o600))
	readmeJA := createTempFile(t, dir, "README.ja.md")
	require.NoError(t, os.WriteFile(readmeJA, []byte(v2ReadmeContents), 0o600))

	input := `
version: 2
project:
  project_id: "project_id"
  name: "Test Project"
  default_language: "en"
pages:
  - type: markdown
    lang:
      en:
        filepath: "README.en.md"
        link: "guide"
        title: "Guide"
      ja:
        filepath: "README.ja.md"
        link: "guide"
        title: "Guide JA"
`

	state := NewParseStateV2("config.yaml", dir)
	conf, err := ParseConfigV2(state, strings.NewReader(input))
	require.NoError(t, err)

	require.Len(t, conf.Pages, 1)
	assert.Equal(t, ConfigPageTypeMarkdownV2, conf.Pages[0].Type)
	require.Len(t, conf.Pages[0].Lang, 2)
	assert.Equal(t, "Guide", conf.Pages[0].Lang["en"].Title)
	assert.Equal(t, "guide", conf.Pages[0].Lang["en"].Link)
	assert.Equal(t, "Guide JA", conf.Pages[0].Lang["ja"].Title)
	assert.Equal(t, "guide", conf.Pages[0].Lang["ja"].Link)
}

func TestParseConfigV2MatchMultiLocale(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	docEN := createTempFile(t, dir, "doc.en.md")
	require.NoError(t, os.WriteFile(docEN, []byte(`
---
title: "Guide"
link: "guide"
lang: "en"
language_group_id: "guide"
---
`), 0o600))
	docJA := createTempFile(t, dir, "doc.ja.md")
	require.NoError(t, os.WriteFile(docJA, []byte(`
---
title: "Guide JA"
link: "guide"
lang: "ja"
language_group_id: "guide"
---
`), 0o600))

	input := `
version: 2
project:
  project_id: "project_id"
  name: "Test Project"
  default_language: "en"
pages:
  - type: match
    pattern: "./*.md"
    sort_key: "title"
    sort_order: "asc"
`

	state := NewParseStateV2("config.yaml", dir)
	conf, err := ParseConfigV2(state, strings.NewReader(input))
	require.NoError(t, err)

	require.Len(t, conf.Pages, 1)
	assert.Equal(t, ConfigPageTypeMarkdownV2, conf.Pages[0].Type)
	require.Len(t, conf.Pages[0].Lang, 2)
	assert.Equal(t, "Guide", conf.Pages[0].Lang["en"].Title)
	assert.Equal(t, "Guide JA", conf.Pages[0].Lang["ja"].Title)
}

func TestParseConfigV2DirectoryWithLang(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	child := createTempFile(t, dir, "child.md")
	require.NoError(t, os.WriteFile(child, []byte(v2ReadmeContents), 0o600))

	input := `
version: 2
project:
  project_id: "project_id"
  name: "Test Project"
  default_language: "en"
pages:
  - type: directory
    lang:
      en:
        title: "English"
      ja:
        title: "Japanese"
    children:
      - type: markdown
        filepath: "child.md"
`

	state := NewParseStateV2("config.yaml", dir)
	conf, err := ParseConfigV2(state, strings.NewReader(input))
	require.NoError(t, err)

	require.Len(t, conf.Pages, 1)
	assert.Equal(t, ConfigPageTypeDirectoryV2, conf.Pages[0].Type)
	require.Len(t, conf.Pages[0].Lang, 2)
	assert.Equal(t, "English", conf.Pages[0].Lang["en"].Title)
	assert.Equal(t, "Japanese", conf.Pages[0].Lang["ja"].Title)
	require.Len(t, conf.Pages[0].Children, 1)
	assert.Equal(t, ConfigPageTypeMarkdownV2, conf.Pages[0].Children[0].Type)
}

func TestParseConfigV2SectionSingleLocale(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	section := createTempFile(t, dir, "section.md")
	require.NoError(t, os.WriteFile(section, []byte(v2ReadmeContents), 0o600))

	input := `
version: 2
project:
  project_id: "project_id"
  name: "Test Project"
pages:
  - type: section
    path: "guide"
    filepath: "section.md"
`

	state := NewParseStateV2("config.yaml", dir)
	conf, err := ParseConfigV2(state, strings.NewReader(input))
	require.NoError(t, err)

	require.Len(t, conf.Pages, 1)
	assert.Equal(t, ConfigPageTypeSectionV2, conf.Pages[0].Type)
	assert.Equal(t, "guide", conf.Pages[0].Path)
	require.Len(t, conf.Pages[0].Lang, 1)
	assert.Equal(t, "README", conf.Pages[0].Lang["en"].Title)
}
