package main

import (
	"bytes"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Version string `yaml:"version"`
}

func parseConfig(reader io.Reader) (*Config, error) {
	var config *Config

	buf := new(bytes.Buffer)
	io.Copy(buf, reader)
	err := yaml.Unmarshal(buf.Bytes(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return config, nil
}
