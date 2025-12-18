package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"mime"
	"path/filepath"
	"slices"
	"strings"

	"github.com/caarlos0/log"
	"github.com/toritoritori29/dodo-cli/src/config"
	appErrors "github.com/toritoritori29/dodo-cli/src/errors"
)

var AvailableMimeTypes = []string{ //nolint: gochecknoglobals
	"image/jpeg",
	"image/png",
	"image/gif",
	"image/webp",
	"image/bmp",
}

type Metadata struct {
	Version string          `json:"version"`
	Project MetadataProject `json:"project"`
	Page    Page            `json:"page"`
	Asset   []MetadataAsset `json:"asset"`
}

func NewMetadataFromConfig(conf *config.ConfigV1) (*Metadata, error) {
	project := NewMetadataProjectFromConfig(conf)
	merr := appErrors.NewMultiError()

	// Validate Page structs from config.
	page, err := CreatePageTree(conf, ".")
	if err != nil {
		merr.Merge(*err)
	}
	if err = page.IsValid(); err != nil {
		merr.Merge(*err)
	}

	// Validate Assets struct from config.
	assets, err := NewMetadataAssetFromConfig(conf, ".")
	if err != nil {
		merr.Merge(*err)
	}

	if merr.HasError() {
		return nil, &merr
	}
	log.Debugf("successfully created a project from the config. project ID: %s", project.ProjectID)
	log.Debugf("successfully created pages from the config. found %d pages", page.Count())
	log.Debugf("successfully created assets from the config. found %d assets", len(assets))

	metadata := Metadata{
		Version: "1",
		Project: project,
		Page:    *page,
		Asset:   assets,
	}
	return &metadata, nil
}

func (m *Metadata) Serialize() ([]byte, error) {
	s, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize metadata: %w", err)
	}
	return s, nil
}

type MetadataProject struct {
	ProjectID       string `json:"project_id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	Version         string `json:"version"`
	Logo            string `json:"logo"`
	Repository      string `json:"repository"`
	DefaultLanguage string `json:"default_language"`
}

func NewMetadataProjectFromConfig(c *config.ConfigV1) MetadataProject {
	return MetadataProject{
		ProjectID:       c.Project.ProjectID,
		Name:            c.Project.Name,
		Description:     c.Project.Description,
		Version:         c.Project.Version,
		Logo:            c.Project.Logo,
		Repository:      c.Project.Repository,
		DefaultLanguage: c.Project.DefaultLanguage,
	}
}

type MetadataAsset struct {
	Path string `json:"path"`
	Hash string `json:"hash"`
}

func NewMetadataAsset(path string) MetadataAsset {
	sum := sha256.Sum256([]byte(path))
	hash := hex.EncodeToString(sum[:])
	return MetadataAsset{
		path,
		hash,
	}
}

func NewMetadataAssetFromConfig(c *config.ConfigV1, rootDir string) ([]MetadataAsset, *appErrors.MultiError) {
	// Create Assets struct from config.
	merr := appErrors.NewMultiError()
	metadataAssets := make([]MetadataAsset, 0, len(c.Assets)*10)
	for _, a := range c.Assets {
		files, err := a.List(rootDir)
		if err != nil {
			merr.Add(err)
		}

		for _, f := range files {
			ma := NewMetadataAsset(f)
			if err = ma.IsValidDataType(); err != nil {
				merr.Add(fmt.Errorf("asset file is invalid: %s: %w", f, err))
				continue
			}
			metadataAssets = append(metadataAssets, ma)
		}
	}

	// Add logo as an asset if exists.
	if c.Project.Logo != "" {
		logoAsset := NewMetadataAsset(c.Project.Logo)
		if err := logoAsset.IsValidDataType(); err != nil {
			merr.Add(err)
		} else {
			metadataAssets = append(metadataAssets, logoAsset)
		}
	}
	if merr.HasError() {
		return nil, &merr
	}
	return metadataAssets, nil
}

func (a *MetadataAsset) IsValidDataType() error {
	if !slices.Contains(AvailableMimeTypes, a.EstimateMimeType()) {
		return fmt.Errorf("the file `%s` has invalid mime type: %s", a.Path, a.EstimateMimeType())
	}
	return nil
}

func (a *MetadataAsset) EstimateMimeType() string {
	ext := strings.ToLower(filepath.Ext(a.Path))
	return mime.TypeByExtension(ext)
}
