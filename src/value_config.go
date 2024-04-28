package main

import (
	"bytes"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Version string         `yaml:"version"`
	Project *ConfigProject `yaml:"project"`
	Pages   []*ConfigPage  `yaml:"pages"`
}

type ConfigProject struct {
	Name        *string `yaml:"name"`
	Description *string `yaml:"description"`
	Version     *string `yaml:"version"`
}

type ConfigPage struct {
	Filepath    *string       `yaml:"filepath"`
	Match       *string       `yaml:"match"`
	Title       *string       `yaml:"title"`
	Path        *string       `yaml:"path"`
	Description *string       `yaml:"description"`
	Children    []*ConfigPage `yaml:"children"`
}

// Check if the page is a valid single page.
// A valid single page should satisfy the following conditions:
// 1. The page must have a filepath field.
// 2. The page must not have a match field.
// 3. The page must have a title field.
// 4. The page must have a name field.
func (c *ConfigPage) IsValidSinglePage() bool {
	if c.Filepath == nil {
		return false
	}
	if c.Match != nil {
		return false
	}
	if c.Title == nil {
		return false
	}
	if c.Path == nil {
		return false
	}
	return true
}

func (c *ConfigPage) IsValidMatchPage() bool {
	if c.Filepath != nil {
		return false
	}
	if c.Match == nil {
		return false
	}
	if c.Title == nil {
		return false
	}
	if c.Path != nil {
		return false
	}
	return true
}

// Check if the page consist of multiple pages.
// A valid single page should satisfy the following conditions:
// 1. The page must not have a filepath field.
// 2. The page must have a match field.
// 4. The page must not have a name field.
func (c *ConfigPage) IsValidMatchDirectory() bool {
	if c.Filepath != nil {
		return false
	}
	if c.Match == nil {
		return false
	}
	if c.Title == nil {
		return false
	}
	if c.Path != nil {
		return false
	}
	return true
}

func ParseDocumentConfig(reader io.Reader) (*Config, error) {
	var definition Config

	buf := new(bytes.Buffer)
	io.Copy(buf, reader)
	err := yaml.Unmarshal(buf.Bytes(), &definition)
	if err != nil {
		return nil, fmt.Errorf("failed to parse document definition: %w", err)
	}
	if err := validateDocumentDefinition(&definition); err != nil {
		return nil, fmt.Errorf("invalid document definition: %w", err)
	}
	return &definition, nil
}

func validateDocumentDefinition(definition *Config) error {
	if definition.Version != "1" {
		return fmt.Errorf("invalid version: %s", definition.Version)
	}
	return nil
}
