package main

import (
	"bytes"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

type DocumentDefinition struct {
	Version string   `yaml:"version"`
	Layout  []Layout `yaml:"layout"`
}

func (d *DocumentDefinition) ListPageHeader() []PageHeader {
	list := make([]PageHeader, 0, len(d.Layout))
	listPageHeader(d.Layout, &list)
	return list
}

func listPageHeader(layoutList []Layout, list *[]PageHeader) {
	for _, l := range layoutList {
		*list = append(*list, NewPageHeader(l.Path, l.Title))
		listPageHeader(l.Children, list)
	}
}

func ParseDocumentDefinition(reader io.Reader) (*DocumentDefinition, error) {
	var definition DocumentDefinition

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

func validateDocumentDefinition(definition *DocumentDefinition) error {
	if definition.Version != "1" {
		return fmt.Errorf("invalid version: %s", definition.Version)
	}
	return nil
}
