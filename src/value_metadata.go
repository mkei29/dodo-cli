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
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

func NewMetadataProjectFromConfig(c *Config) MetadataProject {
	if c.Project == nil {
		return MetadataProject{}
	}

	name := ""
	if c.Project.Name != nil {
		name = *c.Project.Name
	}
	description := ""
	if c.Project.Description != nil {
		description = *c.Project.Description
	}
	version := ""
	if c.Project.Version != nil {
		version = *c.Project.Version
	}

	return MetadataProject{
		Name:        name,
		Description: description,
		Version:     version,
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
