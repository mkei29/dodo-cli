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
