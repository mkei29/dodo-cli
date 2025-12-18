package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/toritoritori29/dodo-cli/src/config"
)

func TestNewMetadataFromConfig(t *testing.T) {
	t.Parallel()

	// Create temporary directory for test files
	dir, err := os.MkdirTemp("", "metadata_test")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(dir) })

	// Create test markdown files
	readme1 := filepath.Join(dir, "README1.md")
	require.NoError(t, os.WriteFile(readme1, []byte(`---
title: "Test Page 1"
path: "test-page-1"
---
# Test Page 1 Content`), 0o600))

	readme2 := filepath.Join(dir, "README2.md")
	require.NoError(t, os.WriteFile(readme2, []byte(`---
title: "Test Page 2"
path: "test-page-2"
---
# Test Page 2 Content`), 0o600))

	// Create assets directory with test files
	assetsDir := filepath.Join(dir, "assets")
	require.NoError(t, os.Mkdir(assetsDir, 0o755))

	// Create a dummy PNG file with proper PNG header
	// PNG magic bytes: 89 50 4E 47 0D 0A 1A 0A
	testImage := filepath.Join(assetsDir, "test.png")
	pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	require.NoError(t, os.WriteFile(testImage, pngHeader, 0o600))

	// Create a text file with .txt extension (not in AvailableMimeTypes)
	invalidAsset := filepath.Join(dir, "invalid.txt")
	require.NoError(t, os.WriteFile(invalidAsset, []byte("not an image"), 0o600))

	// Test successful metadata creation
	t.Run("successful metadata creation", func(t *testing.T) {
		t.Parallel()
		// Create test config
		config := &config.ConfigV1{
			Version: "1",
			Project: config.ConfigProjectV1{
				ProjectID:       "test-project-id",
				Name:            "Test Project",
				Description:     "Test project description",
				Version:         "1.0.0",
				Logo:            "assets/test.png",
				Repository:      "https://github.com/test/repo",
				DefaultLanguage: "JP",
			},
			Pages: []config.ConfigPageV1{
				{
					Markdown: "README1.md",
					Title:    "Test Page 1",
					Path:     "test-page-1",
				},
				{
					Markdown: "README2.md",
					Title:    "Test Page 2",
					Path:     "test-page-2",
				},
			},
			Assets: []config.ConfigAssetV1{
				"assets/*",
			},
		}

		// Change to test directory for relative path resolution
		oldWd, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(dir))
		defer os.Chdir(oldWd)

		metadata, err := NewMetadataFromConfigV1(config)
		require.NoError(t, err)
		require.NotNil(t, metadata)

		// Verify metadata version
		assert.Equal(t, "1", metadata.Version)

		// Verify project metadata
		assert.Equal(t, "test-project-id", metadata.Project.ProjectID)
		assert.Equal(t, "Test Project", metadata.Project.Name)
		assert.Equal(t, "Test project description", metadata.Project.Description)
		assert.Equal(t, "1.0.0", metadata.Project.Version)
		assert.Equal(t, "assets/test.png", metadata.Project.Logo)
		assert.Equal(t, "https://github.com/test/repo", metadata.Project.Repository)
		assert.Equal(t, "JP", metadata.Project.DefaultLanguage)

		// Verify page structure exists (detailed validation would require understanding Page struct)
		assert.NotNil(t, metadata.Page)
		assert.Len(t, metadata.Page.Children, 2)

		// Verify assets (number of assets + logo)
		assert.Len(t, metadata.Asset, 2)
	})

	// Test with empty config
	t.Run("empty pages config", func(t *testing.T) {
		t.Parallel()
		config := &config.ConfigV1{
			Version: "1",
			Project: config.ConfigProjectV1{
				ProjectID: "test-project-id",
				Name:      "Test Project",
			},
			Pages:  []config.ConfigPageV1{},
			Assets: []config.ConfigAssetV1{},
		}

		oldWd, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(dir))
		defer os.Chdir(oldWd)

		metadata, err := NewMetadataFromConfigV1(config)
		require.NoError(t, err)
		require.NotNil(t, metadata)

		// Should have empty pages and assets
		assert.Empty(t, metadata.Asset)
		assert.Equal(t, "1", metadata.Version)
	})

	// Test with invalid asset MIME type
	t.Run("invalid asset MIME type", func(t *testing.T) {
		t.Parallel()
		invalidConfig := &config.ConfigV1{
			Version: "1",
			Project: config.ConfigProjectV1{
				ProjectID: "test-project-id",
				Name:      "Test Project",
			},
			Pages: []config.ConfigPageV1{},
			Assets: []config.ConfigAssetV1{
				"*.txt", // This will match invalid.txt which has unsupported MIME type
			},
		}

		oldWd, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(dir))
		defer os.Chdir(oldWd)

		metadata, err := NewMetadataFromConfigV1(invalidConfig)
		require.Error(t, err)
		assert.Nil(t, metadata)
	})

	// Test with invalid asset MIME type
	t.Run("invalid logo MIME type", func(t *testing.T) {
		t.Parallel()
		invalidConfig := &config.ConfigV1{
			Version: "1",
			Project: config.ConfigProjectV1{
				ProjectID: "test-project-id",
				Name:      "Test Project",
				Logo:      "invalid.txt", // invalid.txt has unsupported MIME type
			},
			Pages:  []config.ConfigPageV1{},
			Assets: []config.ConfigAssetV1{},
		}

		oldWd, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(dir))
		defer os.Chdir(oldWd)

		metadata, err := NewMetadataFromConfigV1(invalidConfig)
		require.Error(t, err)
		assert.Nil(t, metadata)
	})
}

func TestMetadataAsset(t *testing.T) {
	asset := NewMetadataAsset("test/image.png")
	assert.Equal(t, "image/png", asset.EstimateMimeType())
}
