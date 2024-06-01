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
	Index   *ConfigIndex   `yaml:"index"`
	Pages   []*ConfigPage  `yaml:"pages"`
}

type ConfigProject struct {
	Name        *string `yaml:"name"`
	Description *string `yaml:"description"`
	Version     *string `yaml:"version"`
}

type ConfigIndex struct {
	Filepath    *string           `yaml:"filepath"`
	Title       *string           `yaml:"title"`
	Description *string           `yaml:"description"`
	CreatedAt   *SerializableTime `yaml:"created_at"`
}

type ConfigPage struct {
	Filepath    *string           `yaml:"filepath"`
	Match       *string           `yaml:"match"`
	Title       *string           `yaml:"title"`
	Path        *string           `yaml:"path"`
	Description *string           `yaml:"description"`
	SortKey     *string           `yaml:"sort_key"`
	SortOrder   *string           `yaml:"sort_order"`
	CreatedAt   *SerializableTime `yaml:"created_at"`
	Children    []*ConfigPage     `yaml:"children"`
}

// Check if the page is a valid single page.
// A valid single page should satisfy the following conditions:
// 1. The page must have a filepath field.
// 2. The page must not have a match field.
// 3. The page must have a title field.
// 4. The page must have a name field.
func (c *ConfigPage) MatchLeafNode() bool {
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

	// Cannot use sort key and sort order for leaf node.
	if c.SortKey != nil {
		return false
	}
	if c.SortOrder != nil {
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
// 1. The page must have a match field.
func (c *ConfigPage) MatchDirNode() bool {
	if c.Filepath != nil {
		return false
	}
	if c.Match == nil {
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

func ParseConfig(reader io.Reader) (*Config, error) {
	var config Config

	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read a document config: %w", err)
	}
	err = yaml.Unmarshal(buf.Bytes(), &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse a document config: %w", err)
	}
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid document config: %w", err)
	}
	return &config, nil
}

func validateConfig(config *Config) error {
	if config.Version != "1" {
		return fmt.Errorf("invalid version: %s", config.Version)
	}
	return nil
}
