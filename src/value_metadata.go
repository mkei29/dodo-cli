package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

type Metadata struct {
	Version string          `json:"version"`
	Project MetadataProject `json:"project"`
	Page    Page            `json:"page"`
	Asset   []MetadataAsset `json:"asset"`
}

func (m *Metadata) Serialize() ([]byte, error) {
	s, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize metadata: %w", err)
	}
	return s, nil
}

type MetadataProject struct {
	ProjectID   string `json:"project_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Repository  string `json:"repository"`
}

func NewMetadataProjectFromConfig(c *Config) MetadataProject {
	return MetadataProject{
		ProjectID:   c.Project.ProjectID,
		Name:        c.Project.Name,
		Description: c.Project.Description,
		Version:     c.Project.Version,
		Repository:  c.Project.Repository,
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
