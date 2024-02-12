package main

import (
	"encoding/json"
)

type Metadata struct {
	Version string          `json:"version"`
	Project MetadataProject `json:"project"`
	Page    Page            `json:"page"`
}

func (m *Metadata) Serialize() ([]byte, error) {
	s, err := json.Marshal(m)
	return s, err
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
