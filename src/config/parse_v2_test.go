package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func loadTestCaseV2(t *testing.T, name string) (string, string) {
	t.Helper()
	root := filepath.Join("test_cases", "v2", name)
	data, err := os.ReadFile(filepath.Join(root, ".dodo.yaml"))
	require.NoError(t, err)
	return root, string(data)
}

func findPageByType(pages []ConfigPageV2, pageType string) *ConfigPageV2 {
	for i := range pages {
		if pages[i].Type == pageType {
			return &pages[i]
		}
	}
	return nil
}

func findPageByLink(pages []ConfigPageV2, link string) *ConfigPageV2 {
	for i := range pages {
		for _, lang := range pages[i].LangPage {
			if lang.Link == link {
				return &pages[i]
			}
		}
	}
	return nil
}

func TestParseConfigV2(t *testing.T) {
	tests := []struct {
		name     string
		caseName string
		assert   func(t *testing.T, conf *ConfigV2)
	}{
		{
			name:     "markdown_single_locale",
			caseName: "1_valid_single_language",
			assert: func(t *testing.T, conf *ConfigV2) {
				// Readme
				require.Len(t, conf.Pages, 3, "there should be 3 pages at the top level")
				page := conf.Pages[0]
				assert.Equal(t, ConfigPageTypeMarkdownV2, page.Type)
				require.NotNil(t, page.LangPage["en"])
				assert.Equal(t, "README.md", page.LangPage["en"].Filepath)
				assert.Equal(t, "README", page.LangPage["en"].Title)
				assert.Equal(t, "readme", page.LangPage["en"].Link)

				// Section
				page = conf.Pages[1]
				assert.Equal(t, ConfigPageTypeSectionV2, page.Type)
				require.Len(t, page.LangSection, 1)
				assert.Equal(t, "Section", page.LangSection["en"].Title)

				// Section Children
				require.Len(t, page.Children, 2)
				child1 := page.Children[0]
				assert.Equal(t, ConfigPageTypeMarkdownV2, child1.Type)
				require.NotNil(t, child1.LangPage["en"])
				assert.Equal(t, "file1.md", child1.LangPage["en"].Filepath)
				assert.Equal(t, "File1", child1.LangPage["en"].Title)
				assert.Equal(t, "file1", child1.LangPage["en"].Link)

				child2 := page.Children[1]
				assert.Equal(t, ConfigPageTypeMarkdownV2, child2.Type)
				require.NotNil(t, child2.LangPage["en"])
				assert.Equal(t, "file2.md", child2.LangPage["en"].Filepath)
				assert.Equal(t, "File2", child2.LangPage["en"].Title)
				assert.Equal(t, "file2", child2.LangPage["en"].Link)
			},
		},
		{
			name:     "markdown_multi_locale",
			caseName: "2_valid_multi_language",
			assert: func(t *testing.T, conf *ConfigV2) {
				require.Len(t, conf.Pages, 2, "there should be 2 pages at the top level")

				// Section (single locale with multi-locale children)
				page := conf.Pages[0]
				assert.Equal(t, ConfigPageTypeSectionV2, page.Type)
				require.Len(t, page.LangSection, 1)
				assert.Equal(t, "Section", page.LangSection["en"].Title)

				// Section Children
				require.Len(t, page.Children, 2, "section should have 2 children")

				// First child: explicit title and link
				child1 := page.Children[0]
				assert.Equal(t, ConfigPageTypeMarkdownMultiLanguageV2, child1.Type)
				require.Len(t, child1.LangPage, 2, "child1 should have 2 languages")
				require.NotNil(t, child1.LangPage["en"])
				assert.Equal(t, "file1.en.md", child1.LangPage["en"].Filepath)
				assert.Equal(t, "File1 EN", child1.LangPage["en"].Title)
				assert.Equal(t, "file1_en", child1.LangPage["en"].Link)
				require.NotNil(t, child1.LangPage["ja"])
				assert.Equal(t, "file1.ja.md", child1.LangPage["ja"].Filepath)
				assert.Equal(t, "File1 JA", child1.LangPage["ja"].Title)
				assert.Equal(t, "file1_ja", child1.LangPage["ja"].Link)

				// Second child: implicit title and link from frontmatter
				child2 := page.Children[1]
				assert.Equal(t, ConfigPageTypeMarkdownMultiLanguageV2, child2.Type)
				require.Len(t, child2.LangPage, 2, "child2 should have 2 languages")
				require.NotNil(t, child2.LangPage["en"])
				assert.Equal(t, "file2.en.md", child2.LangPage["en"].Filepath)
				assert.Equal(t, "File2", child2.LangPage["en"].Title)
				assert.Equal(t, "file2", child2.LangPage["en"].Link)
				require.NotNil(t, child2.LangPage["ja"])
				assert.Equal(t, "file2.ja.md", child2.LangPage["ja"].Filepath)
				assert.Equal(t, "File2", child2.LangPage["ja"].Title)
				assert.Equal(t, "file2", child2.LangPage["ja"].Link)

				// Directory (single locale)
				page = conf.Pages[1]
				assert.Equal(t, ConfigPageTypeDirectoryV2, page.Type)
				require.Len(t, page.LangDirectory, 1)
				assert.Equal(t, "Docs", page.LangDirectory["en"].Title)

				// Directory Children (from match pattern)
				require.Len(t, page.Children, 2, "directory should have 2 children from match")

				// Children should be sorted by title (asc)
				// Child1 comes before Child2
				matchChild1 := page.Children[0]
				assert.Equal(t, ConfigPageTypeMarkdownMultiLanguageV2, matchChild1.Type)
				require.Len(t, matchChild1.LangPage, 2, "match child1 should have 2 languages")
				require.NotNil(t, matchChild1.LangPage["en"])
				assert.Equal(t, "directory/child1.en.md", matchChild1.LangPage["en"].Filepath)
				assert.Equal(t, "Child1 EN", matchChild1.LangPage["en"].Title)
				assert.Equal(t, "child1_en", matchChild1.LangPage["en"].Link)
				require.NotNil(t, matchChild1.LangPage["ja"])
				assert.Equal(t, "directory/child1.ja.md", matchChild1.LangPage["ja"].Filepath)
				assert.Equal(t, "Child1_JA", matchChild1.LangPage["ja"].Title)
				assert.Equal(t, "child1_ja", matchChild1.LangPage["ja"].Link)

				matchChild2 := page.Children[1]
				assert.Equal(t, ConfigPageTypeMarkdownMultiLanguageV2, matchChild2.Type)
				require.Len(t, matchChild2.LangPage, 2, "match child2 should have 2 languages")
				require.NotNil(t, matchChild2.LangPage["en"])
				assert.Equal(t, "directory/child2.en.md", matchChild2.LangPage["en"].Filepath)
				assert.Equal(t, "Child2 EN", matchChild2.LangPage["en"].Title)
				assert.Equal(t, "child2_en", matchChild2.LangPage["en"].Link)
				require.NotNil(t, matchChild2.LangPage["ja"])
				assert.Equal(t, "directory/child2.ja.md", matchChild2.LangPage["ja"].Filepath)
				assert.Equal(t, "Child2 JA", matchChild2.LangPage["ja"].Title)
				assert.Equal(t, "child2_ja", matchChild2.LangPage["ja"].Link)
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir, input := loadTestCaseV2(t, tc.caseName)
			state := NewParseStateV2("config.yaml", dir)
			conf, err := ParseConfigV2(state, strings.NewReader(input))
			require.NoError(t, err)
			tc.assert(t, conf)
		})
	}
}
