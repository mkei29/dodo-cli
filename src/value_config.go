package main

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"

	"github.com/mattn/go-zglob"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Version string         `yaml:"version"`
	Project *ConfigProject `yaml:"project"`
	Pages   []*ConfigPage  `yaml:"pages"`
	Assets  []ConfigAsset  `yaml:"assets"`
}

type ConfigProject struct {
	Name        *string `yaml:"name"`
	Description *string `yaml:"description"`
	Version     *string `yaml:"version"`
}

type ConfigPage struct {
	// markdown syntax
	Markdown    *string           `yaml:"markdown"`
	Title       *string           `yaml:"title"`
	Path        *string           `yaml:"path"`
	Description *string           `yaml:"description"`
	UpdatedAt   *SerializableTime `yaml:"updated_at"`

	// match syntax
	Match     *string `yaml:"match"`
	SortKey   *string `yaml:"sort_key"`
	SortOrder *string `yaml:"sort_order"`

	// directory syntax
	Directory *string       `yaml:"directory"`
	Children  []*ConfigPage `yaml:"children"`
}

// Check if the page is a valid single page.
func (c *ConfigPage) MatchMarkdown() bool {
	if c.Markdown == nil {
		return false
	}

	// prohibit match syntax
	if c.Match != nil {
		return false
	}
	if c.SortKey != nil {
		return false
	}
	if c.SortOrder != nil {
		return false
	}

	// prohibit directory syntax
	if c.Directory != nil {
		return false
	}
	if c.Children != nil {
		return false
	}
	return true
}

func (c *ConfigPage) MatchMatch() bool {
	if c.Match == nil {
		return false
	}

	// prohibit markdown syntax
	if c.Markdown != nil {
		return false
	}
	if c.Title != nil {
		return false
	}
	if c.Path != nil {
		return false
	}
	if c.Description != nil {
		return false
	}
	if c.UpdatedAt != nil {
		return false
	}
	// prohibit directory syntax
	if c.Directory != nil {
		return false
	}
	if c.Children != nil {
		return false
	}
	return true
}

func (c *ConfigPage) MatchDirectory() bool {
	if c.Directory == nil {
		return false
	}

	// prohibit markdown syntax
	if c.Markdown != nil {
		return false
	}
	if c.Title != nil {
		return false
	}
	if c.Path != nil {
		return false
	}
	if c.Description != nil {
		return false
	}
	if c.UpdatedAt != nil {
		return false
	}

	// prohibit match syntax
	if c.Match != nil {
		return false
	}
	if c.SortKey != nil {
		return false
	}
	if c.SortOrder != nil {
		return false
	}
	return true
}

type ConfigAsset string

func (m ConfigAsset) List(rootDir string) ([]string, error) {
	globPath := filepath.Clean(filepath.Join(rootDir, string(m)))
	if err := IsUnderRootPath(rootDir, globPath); err != nil {
		return nil, fmt.Errorf("invalid configuration: path should be under the rootDir: path: %s", globPath)
	}

	matches, err := zglob.Glob(globPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list files match '%s' : %w", globPath, err)
	}
	return matches, nil
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
