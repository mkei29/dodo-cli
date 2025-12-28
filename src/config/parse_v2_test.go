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
				page := findPageByLink(conf.Pages, "readme")
				require.NotNil(t, page)
				assert.Equal(t, ConfigPageTypeMarkdownV2, page.Type)
				require.Len(t, page.LangPage, 1)
				assert.Equal(t, "README.md", page.LangPage["en"].Filepath)
				assert.Equal(t, "README", page.LangPage["en"].Title)
				assert.Equal(t, "Readme description", page.LangPage["en"].Description)
				assert.Equal(t, "readme-path", page.LangPage["en"].Path)
				assert.Equal(t, "readme", page.LangPage["en"].Link)
			},
		},
		{
			name:     "markdown_multi_locale",
			caseName: "2_valid_multi_language",
			assert: func(t *testing.T, conf *ConfigV2) {
				page := findPageByLink(conf.Pages, "guide")
				require.NotNil(t, page)
				assert.Equal(t, ConfigPageTypeMarkdownMultiLanguageV2, page.Type)
				require.Len(t, page.LangPage, 2)
				assert.Equal(t, "Guide", page.LangPage["en"].Title)
				assert.Equal(t, "Guide description", page.LangPage["en"].Description)
				assert.Equal(t, "guide-path-en", page.LangPage["en"].Path)
				assert.Equal(t, "guide", page.LangPage["en"].Link)
				assert.Equal(t, "Guide JA", page.LangPage["ja"].Title)
				assert.Equal(t, "Guide description JA", page.LangPage["ja"].Description)
				assert.Equal(t, "guide-path-ja", page.LangPage["ja"].Path)
				assert.Equal(t, "guide", page.LangPage["ja"].Link)
			},
		},
		{
			name:     "match_multi_locale",
			caseName: "2_valid_multi_language",
			assert: func(t *testing.T, conf *ConfigV2) {
				page := findPageByLink(conf.Pages, "match-guide")
				require.NotNil(t, page)
				require.Len(t, page.LangPage, 2)
				assert.Equal(t, "Guide", page.LangPage["en"].Title)
				assert.Equal(t, "Match guide description", page.LangPage["en"].Description)
				assert.Equal(t, "match-guide-en", page.LangPage["en"].Path)
				assert.Equal(t, "Guide JA", page.LangPage["ja"].Title)
				assert.Equal(t, "Match guide description JA", page.LangPage["ja"].Description)
				assert.Equal(t, "match-guide-ja", page.LangPage["ja"].Path)
			},
		},
		{
			name:     "directory_with_lang",
			caseName: "2_valid_multi_language",
			assert: func(t *testing.T, conf *ConfigV2) {
				page := findPageByType(conf.Pages, ConfigPageTypeDirectoryV2)
				require.NotNil(t, page)
				require.Len(t, page.LangDirectory, 2)
				assert.Equal(t, "English", page.LangDirectory["en"].Title)
				assert.Equal(t, "Japanese", page.LangDirectory["ja"].Title)
				require.Len(t, page.Children, 1)
				assert.Equal(t, ConfigPageTypeMarkdownV2, page.Children[0].Type)
			},
		},
		{
			name:     "section_single_locale",
			caseName: "1_valid_single_language",
			assert: func(t *testing.T, conf *ConfigV2) {
				page := findPageByType(conf.Pages, ConfigPageTypeSectionV2)
				require.NotNil(t, page)
				assert.Equal(t, "guide", page.Path)
				require.Len(t, page.LangPage, 1)
				assert.Equal(t, "README", page.LangPage["en"].Title)
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
