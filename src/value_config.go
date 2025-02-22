package main

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"

	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/mattn/go-zglob"
)

const (
	ConfigPageTypeMarkdown = iota
	ConfigPageTypeMatch
	ConfigPageTypeDirectory
)

type Config struct {
	Version string        `yaml:"version"`
	Project ConfigProject `yaml:"project"`
	Pages   []ConfigPage  `yaml:"pages"`
	Assets  []ConfigAsset `yaml:"assets"`
}

type ConfigProject struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Version     string `yaml:"version"`
}

type ConfigPage struct {
	// markdown syntax
	Markdown    string           `yaml:"markdown"`
	Title       string           `yaml:"title"`
	Path        string           `yaml:"path"`
	Description string           `yaml:"description"`
	UpdatedAt   SerializableTime `yaml:"updated_at"`

	// match syntax
	Match     string `yaml:"match"`
	SortKey   string `yaml:"sort_key"`
	SortOrder string `yaml:"sort_order"`

	// directory syntax
	Directory string       `yaml:"directory"`
	Children  []ConfigPage `yaml:"children"`
}

// Check if the page is a valid single page.
func (c *ConfigPage) MatchMarkdown() bool {
	return c.Markdown != ""
}

func (c *ConfigPage) MatchMatch() bool {
	return c.Match != ""
}

func (c *ConfigPage) MatchDirectory() bool {
	return c.Directory != ""
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

type ParseState struct {
	config                 Config
	isVersionAlreadyParsed bool
	isProjectAlreadyParsed bool
	isPagesAlreadyParsed   bool
	isAssetsAlreadyParsed  bool
}

func NewParseState() *ParseState {
	return &ParseState{}
}

// ParseConfig takes a reader and parses it into a Config struct.
// While parsing, it validates the config at the same time.
// This is because we want to prvide a user-friendly error message.
//
// This function respects the following implementation:
// https://github.com/goccy/go-yaml/blob/abc70836f5a5623a92cf51d4bf40cbaf8fed2faa/decode.go
func ParseConfig(reader io.Reader) (*Config, error) {
	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read a document config: %w", err)
	}

	state := NewParseState()

	// NOTE: The role of parser.Mode(0) is little bit unclear.
	// I couldn't find any documentation about it.
	root, err := parser.ParseBytes(buf.Bytes(), parser.Mode(0))
	if err != nil {
		return nil, fmt.Errorf("failed to parse a document config: %w", err)
	}
	if len(root.Docs) != 1 {
		return nil, fmt.Errorf("invalid document config: there should be only one document. got %d", len(root.Docs))
	}

	body, ok := root.Docs[0].Body.(*ast.MappingNode)
	if !ok {
		return nil, fmt.Errorf("invalid document config: the document should be a mapping type")
	}

	// Apply parseRootItem to each node at the root level.
	for _, node := range body.Values {
		if err := parseRootItem(state, node); err != nil {
			return nil, err
		}
	}
	return &state.config, nil
}

func parseRootItem(state *ParseState, node ast.Node) error {
	mapping, ok := node.(*ast.MappingValueNode)
	if !ok {
		return fmt.Errorf("invalid document config: the root item should be a mapping type")
	}
	if mapping.Key.String() == "version" {
		return parseVersion(state, mapping)
	}
	if mapping.Key.String() == "project" {
		return parseConfigProject(state, mapping)
	}
	if mapping.Key.String() == "pages" {
		return parseConfigPage(state, mapping)
	}
	if mapping.Key.String() == "assets" {
		return parseConfigAssets(state, mapping)
	}

	return nil
}

func parseVersion(state *ParseState, node *ast.MappingValueNode) error {
	// This function should be called only once.
	// Receive a object like following:
	//
	// version: "1"
	//

	if state.isVersionAlreadyParsed {
		return fmt.Errorf("duplicate version")
	}
	value := node.Value.String()
	if value != "1" {
		return fmt.Errorf("invalid version. version should be '1' : %s", value)
	}

	state.isVersionAlreadyParsed = true
	state.config.Version = value
	return nil
}

func parseConfigProject(state *ParseState, node *ast.MappingValueNode) error {
	// This function should be called only once.
	// Receive a object like following:
	//
	// project:
	//   name: "My Project"
	//   description: "This is my project"
	//   version: "1.0.0"

	if state.isProjectAlreadyParsed {
		return fmt.Errorf("duplicate project")
	}

	// node.value should be a mapping type.
	// ConfigProject needs name, description, and version.
	// And all of them are string type.
	children, ok := node.Value.(*ast.MappingNode)
	if !ok {
		return fmt.Errorf("invalid project: the project should be a mapping type")
	}

	for _, item := range children.Values {
		key := item.Key.String()
		switch key {
		case "name":
			state.config.Project.Name = item.Value.String()
		case "description":
			state.config.Project.Description = item.Value.String()
		case "version":
			state.config.Project.Version = item.Value.String()
		default:
			return fmt.Errorf("invalid project key: %s", key)
		}
	}
	state.isProjectAlreadyParsed = true
	return nil
}

func parseConfigPage(state *ParseState, node *ast.MappingValueNode) error {
	// This function should be called only once.
	// Receive a object like following:
	//
	// pages:
	//   - markdown: "README1.md"
	//     ...
	//   - markdown: "README2.md"
	//     ..

	if state.isPagesAlreadyParsed {
		return fmt.Errorf("duplicate pages")
	}

	sequence, ok := node.Value.(*ast.SequenceNode)
	if !ok {
		return fmt.Errorf("invalid pages: the pages should be a sequence type")
	}

	pages, err := parseConfigPageSequence(sequence)
	if err != nil {
		return err
	}
	state.isPagesAlreadyParsed = true
	state.config.Pages = pages
	return nil
}

func parseConfigPageSequence(sequence *ast.SequenceNode) ([]ConfigPage, error) {
	// Receive a object like following:
	//
	// xxx:
	//   - markdown: "README1.md"
	//     ...
	//   - markdown: "README2.md"
	//     ...

	configPages := make([]ConfigPage, 0, len(sequence.Values))
	for _, item := range sequence.Values {
		pageNode, ok := item.(*ast.MappingNode)
		if !ok {
			// TODO: エラーメッセージを改善
			return nil, ErrUnexpectedNode("`page` should have a mapping", item)
		}

		t, err := estimateConfigPageType(pageNode)
		if err != nil {
			return nil, ErrUnexpectedNode("this mapping does not match any page type", item)
		}

		if t == ConfigPageTypeMarkdown {
			p, err := parseConfigPageMarkdown(pageNode)
			if err != nil {
				return nil, fmt.Errorf("invalid page: %w", err)
			}
			configPages = append(configPages, p)
		}

		if t == ConfigPageTypeMatch {
			p, err := parseConfigPageMatch(pageNode)
			if err != nil {
				return nil, fmt.Errorf("invalid page: %w", err)
			}
			configPages = append(configPages, p)
		}

		if t == ConfigPageTypeDirectory {
			p, err := parseConfigPageDirectory(pageNode)
			if err != nil {
				return nil, fmt.Errorf("invalid page: %w", err)
			}
			configPages = append(configPages, p)
		}
	}
	return configPages, nil
}

func estimateConfigPageType(mapping *ast.MappingNode) (int, error) {
	// a page object must match one of the following patterns:
	// {
	//   "markdown": "README1.md",
	//   ...
	// }
	// {
	//   "match": "docs/*.md",
	//   ...
	// }
	// {
	//   "directory": "path/to/directory",
	//   ...
	// }

	for _, item := range mapping.Values {
		key := item.Key.String()
		switch key {
		case "markdown":
			return ConfigPageTypeMarkdown, nil
		case "match":
			return ConfigPageTypeMatch, nil
		case "directory":
			return ConfigPageTypeDirectory, nil
		}
	}
	return -1, ErrUnexpectedNode("this mapping does not match any page type", mapping)
}

func parseConfigPageMarkdown(mapping *ast.MappingNode) (ConfigPage, error) { //nolint: cyclop
	// a markdown object has the following fields:
	// {
	//   "markdown": "README1.md",
	//   "title": "README1",
	//   "path": "readme1",
	//   "description": "This is README1",
	//   "updated_at": "2021-01-01T00:00:00Z"
	// }
	configPage := ConfigPage{}

	for _, item := range mapping.Values {
		key := item.Key.String()

		switch key {
		case "markdown":
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				return ConfigPage{}, ErrUnexpectedNode("`markdown` field should be a string", item.Value)
			}
			configPage.Markdown = v.Value
		case "title":
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				return ConfigPage{}, ErrUnexpectedNode("`title` field should be a string", item.Value)
			}
			configPage.Title = v.Value
		case "path":
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				return ConfigPage{}, ErrUnexpectedNode("`path` field should be a string", item.Value)
			}
			configPage.Path = v.Value
		case "description":
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				return ConfigPage{}, ErrUnexpectedNode("`description` field should be a string", item.Value)
			}
			configPage.Description = v.Value
		case "updated_at":
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				return ConfigPage{}, ErrUnexpectedNode("`updated_at` field should be a string", item.Value)
			}
			// TODO: updated_atのフォーマットが正しいことを検証する
			configPage.UpdatedAt = SerializableTime(v.Value)
		default:
			return ConfigPage{}, ErrUnexpectedNode(fmt.Sprintf("a markdown style page cannot accept the key: %s", key), item)
		}
	}
	// TODO: ここで必要なフィールドがすべてあるか検証する
	return configPage, nil
}

func parseConfigPageMatch(mapping *ast.MappingNode) (ConfigPage, error) {
	// a match object has the following fields:
	// {
	//   "match": "docs/*.md",
	//   "sort_key": "title",
	//   "sort_order": "asc"
	// }

	configPage := ConfigPage{}
	for _, item := range mapping.Values {
		key := item.Key.String()
		switch key {
		case "match":
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				return ConfigPage{}, ErrUnexpectedNode("`match` field should be a string", item.Value)
			}
			configPage.Match = v.Value
		case "sort_key":
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				return ConfigPage{}, ErrUnexpectedNode("`sort_key` field should be a string", item.Value)
			}
			// TODO: sort_keyで受け入れ可能な値であることを検証する
			configPage.SortKey = v.Value
		case "sort_order":
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				return ConfigPage{}, ErrUnexpectedNode("`sort_order` field should be a string", item.Value)
			}
			// TODO: sort_orderはasc, descのいずれかであることを検証する
			configPage.SortOrder = v.Value
		default:
			return ConfigPage{}, ErrUnexpectedNode("a match style page cannot accept the key: %s", item.Value)
		}
	}
	// TODO: ここで必要なフィールドがすべてあるか検証する
	return configPage, nil
}

func parseConfigPageDirectory(mapping *ast.MappingNode) (ConfigPage, error) {
	// a directory object has the following fields:
	//
	// {
	//   "directory": "path/to/directory",
	//   "children": [
	//     {
	//       "markdown": "README1.md",
	//       ...
	//     },
	// }

	configPage := ConfigPage{}
	for _, item := range mapping.Values {
		key := item.Key.String()
		switch key {
		case "directory":
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				return ConfigPage{}, ErrUnexpectedNode("`directory` field should be a string", item.Value)
			}
			configPage.Directory = v.Value
		case "children":
			v, ok := item.Value.(*ast.SequenceNode)
			if !ok {
				return ConfigPage{}, ErrUnexpectedNode("`children` field should be a sequence", item.Value)
			}
			children, err := parseConfigPageSequence(v)
			if err != nil {
				return ConfigPage{}, err
			}
			configPage.Children = children
		default:
			return ConfigPage{}, ErrUnexpectedNode("a directory style page cannot accept a key", item)
		}
	}
	// TODO: ここで必要なフィールドがすべてあるか検証する
	return configPage, nil
}

func parseConfigAssets(state *ParseState, node *ast.MappingValueNode) error {
	// This function should be called only once.
	// Receive a object like following:
	//
	// assets:
	//   - "assets/**"
	//   - "images/**"

	if state.isAssetsAlreadyParsed {
		return ErrUnexpectedNode("there should be a exact one `assets` field in the config file", node)
	}

	sequence, ok := node.Value.(*ast.SequenceNode)
	if !ok {
		return ErrUnexpectedNode("the `assets` field should be a sequence of string", node.Value)
	}

	assets := make([]ConfigAsset, 0, len(sequence.Values))
	for _, item := range sequence.Values {
		v, ok := item.(*ast.StringNode)
		if !ok {
			return ErrUnexpectedNode("a item in the `sequence` field should have a string type", item)
		}
		assets = append(assets, ConfigAsset(v.Value))
	}
	state.config.Assets = assets
	return nil
}
