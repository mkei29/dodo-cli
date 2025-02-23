package main

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/mattn/go-zglob"
)

const (
	ConfigPageTypeUnknown = iota
	ConfigPageTypeMarkdown
	ConfigPageTypeMatch
	ConfigPageTypeDirectory
)

type Config struct {
	Version string
	Project ConfigProject
	Pages   []ConfigPage
	Assets  []ConfigAsset
}

type ConfigProject struct {
	Name        string
	Description string
	Version     string
}

type ConfigPage struct {
	// markdown syntax
	Markdown    string
	Title       string
	Path        string
	Description string
	UpdatedAt   SerializableTime

	// match syntax
	Match     string
	SortKey   string
	SortOrder string

	// directory syntax
	Directory string
	Children  []ConfigPage
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
		return nil, fmt.Errorf("invalid configuration: path should be under the root directory: path: %s", globPath)
	}

	matches, err := zglob.Glob(globPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list files matching '%s': %w", globPath, err)
	}
	return matches, nil
}

// A struct to keep the state of the parsing process.
type ParseState struct {
	filepath               string // The name of the file being parsed.
	config                 Config // The config object being generated.
	contents               []byte // The contents of the file being parsed.
	isVersionAlreadyParsed bool
	isProjectAlreadyParsed bool
	isPagesAlreadyParsed   bool
	isAssetsAlreadyParsed  bool
	errorSet               MultiError
}

func NewParseState(filepath string, contents []byte) *ParseState {
	return &ParseState{
		filepath: filepath,
		contents: contents,
	}
}

func (s *ParseState) buildParseError(message string, node ast.Node) error {
	line := s.getLineFromNode(node)
	return &ParseError{
		filepath: s.filepath,
		message:  message,
		line:     line,
		node:     node,
	}
}

func (s *ParseState) getLineFromNode(node ast.Node) string {
	lines := bytes.Split(s.contents, []byte("\n"))
	lineNumber := node.GetToken().Position.Line - 1
	if lineNumber < 0 || lineNumber >= len(lines) {
		return "(unknown line)"
	}
	return string(lines[lineNumber])
}

// ParseConfig takes a reader and parses it into a Config struct.
// While parsing, it validates the config at the same time.
// This is because we want to prvide a user-friendly error message.
//
// This function respects the following implementation:
// https://github.com/goccy/go-yaml/blob/abc70836f5a5623a92cf51d4bf40cbaf8fed2faa/decode.go
func ParseConfig(filepath string, reader io.Reader) (*Config, error) {
	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read a document config: %w", err)
	}

	contents := buf.Bytes()
	state := NewParseState(filepath, contents)

	// NOTE: The role of parser.Mode(0) is little bit unclear.
	// I couldn't find any documentation about it.
	root, err := parser.ParseBytes(contents, parser.Mode(0))
	if err != nil {
		// TODO: return more detailed error including the line number.
		return nil, fmt.Errorf("failed to parse a document config: %w", err)
	}
	parseRoot(state, root)

	if state.errorSet.HasError() {
		return nil, &state.errorSet
	}
	return &state.config, nil
}

func parseRoot(state *ParseState, root *ast.File) {
	if len(root.Docs) != 1 {
		state.errorSet.Add(fmt.Errorf("there should be only one document. Got %d", len(root.Docs)))
		return
	}

	body, ok := root.Docs[0].Body.(*ast.MappingNode)
	if !ok {
		state.errorSet.Add(state.buildParseError("the root node should be of mapping type", root.Docs[0].Body))
		return
	}

	// Apply parseRootItem to each node at the root level.
	for _, node := range body.Values {
		parseRootItem(state, node)
	}

	// Check if all required fields are parsed.
	// NOTE: The `assets` field is optional.
	if !state.isVersionAlreadyParsed {
		state.errorSet.Add(fmt.Errorf("the `version` field is required"))
	}
	if !state.isProjectAlreadyParsed {
		state.errorSet.Add(fmt.Errorf("the `project` field is required"))
	}
	if !state.isPagesAlreadyParsed {
		state.errorSet.Add(fmt.Errorf("the `pages` field is required"))
	}
}

func parseRootItem(state *ParseState, node ast.Node) {
	mapping, ok := node.(*ast.MappingValueNode)
	if !ok {
		state.errorSet.Add(state.buildParseError("a key-value pair is expected at the top level", node))
		return
	}
	if mapping.Key.String() == "version" {
		parseVersion(state, mapping)
		return
	}
	if mapping.Key.String() == "project" {
		parseConfigProject(state, mapping)
		return
	}
	if mapping.Key.String() == "pages" {
		parseConfigPage(state, mapping)
		return
	}
	if mapping.Key.String() == "assets" {
		parseConfigAssets(state, mapping)
		return
	}
	state.errorSet.Add(state.buildParseError("unexpected key at the top level", mapping.Key))
}

func parseVersion(state *ParseState, node *ast.MappingValueNode) {
	// This function should be called only once.
	// Receive an object like the following:
	//
	// version: "1"
	//

	if state.isVersionAlreadyParsed {
		state.errorSet.Add(state.buildParseError("there should be exactly one `version` section at the top level", node))
		return
	}
	state.isVersionAlreadyParsed = true

	intNode, ok := node.Value.(*ast.IntegerNode)
	if !ok {
		state.errorSet.Add(state.buildParseError("`version` should have an integer value", node.Value))
		return
	}

	var versionNum int
	switch v := intNode.Value.(type) {
	case int:
		versionNum = v
	case uint64:
		versionNum = int(v)
	default:
		state.errorSet.Add(state.buildParseError("internal error: `version` has an unexpected type", node.Value))
		return
	}

	if versionNum != 1 {
		state.errorSet.Add(state.buildParseError("unsupported version: only '1' is supported now", intNode))
		return
	}
	state.config.Version = strconv.Itoa(versionNum)
}

func parseConfigProject(state *ParseState, node *ast.MappingValueNode) { //nolint: cyclop
	// This function should be called only once.
	// Receive an object like the following:
	//
	// project:
	//   name: "My Project"
	//   description: "This is my project"
	//   version: "1.0.0"

	if state.isProjectAlreadyParsed {
		state.errorSet.Add(state.buildParseError("there should be exactly one `project` section at the top level", node))
		return
	}
	state.isProjectAlreadyParsed = true

	// node.value should be a mapping type.
	// ConfigProject needs name, description, and version.
	// And all of them are string type.
	children, ok := node.Value.(*ast.MappingNode)
	if !ok {
		state.errorSet.Add(state.buildParseError("the `project` should have a mapping value", node.Value))
		return
	}

	for _, item := range children.Values {
		key := item.Key.String()
		switch key {
		case "name":
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`name` field should be a string", item.Value))
				continue
			}
			state.config.Project.Name = v.Value
		case "description":
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`description` field should be a string", item.Value))
				continue
			}
			state.config.Project.Description = v.Value
		case "version":
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`version` field should be a string", item.Value))
				continue
			}
			state.config.Project.Version = v.Value
		default:
			state.errorSet.Add(state.buildParseError(fmt.Sprintf("the `project` does not accept the key: %s", key), item))
		}
	}

	if state.config.Project.Name == "" {
		state.errorSet.Add(state.buildParseError("the `project` should have a `name` field longer than 1 character", node))
	}
}

func parseConfigPage(state *ParseState, node *ast.MappingValueNode) {
	// This function should be called only once.
	// Receive an object like the following:
	//
	// pages:
	//   - markdown: "README1.md"
	//     ...
	//   - markdown: "README2.md"
	//     ..

	if state.isPagesAlreadyParsed {
		state.errorSet.Add(state.buildParseError("there should be exactly one `pages` section at the top level", node))
		return
	}
	state.isPagesAlreadyParsed = true

	sequence, ok := node.Value.(*ast.SequenceNode)
	if !ok {
		state.errorSet.Add(state.buildParseError("the `pages` field should be a sequence type", node.Value))
		return
	}
	state.config.Pages = parseConfigPageSequence(state, sequence)
}

func parseConfigPageSequence(state *ParseState, sequence *ast.SequenceNode) []ConfigPage {
	// Receive an object like the following:
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
			state.errorSet.Add(state.buildParseError("each item in the `pages` sequence should be of mapping type", item))
			continue
		}

		t := estimateConfigPageType(pageNode)
		if t == ConfigPageTypeUnknown {
			state.errorSet.Add(state.buildParseError("this mapping does not match any page type", pageNode))
			continue
		}

		if t == ConfigPageTypeMarkdown {
			p := parseConfigPageMarkdown(state, pageNode)
			configPages = append(configPages, p)
			continue
		}

		if t == ConfigPageTypeMatch {
			p := parseConfigPageMatch(state, pageNode)
			configPages = append(configPages, p)
			continue
		}

		if t == ConfigPageTypeDirectory {
			p := parseConfigPageDirectory(state, pageNode)
			configPages = append(configPages, p)
			continue
		}
		state.errorSet.Add(state.buildParseError("unreachable code", pageNode))
	}
	return configPages
}

func estimateConfigPageType(mapping *ast.MappingNode) int {
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
			return ConfigPageTypeMarkdown
		case "match":
			return ConfigPageTypeMatch
		case "directory":
			return ConfigPageTypeDirectory
		}
	}
	return ConfigPageTypeUnknown
}

func parseConfigPageMarkdown(state *ParseState, mapping *ast.MappingNode) ConfigPage { //nolint: cyclop, funlen
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
				state.errorSet.Add(state.buildParseError("`markdown` field should be a string", item.Value))
				continue
			}
			configPage.Markdown = v.Value
		case "title": //nolint: goconst
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`title` field should be a string", item.Value))
				continue
			}
			configPage.Title = v.Value
		case "path":
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`path` field should be a string", item.Value))
				continue
			}
			configPage.Path = v.Value
		case "description":
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`description` field should be a string", item.Value))
				continue
			}
			configPage.Description = v.Value
		case "updated_at":
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`updated_at` field should be a string", item.Value))
				continue
			}
			st, err := NewSerializableTime(v.Value)
			if err != nil {
				state.errorSet.Add(state.buildParseError("`updated_at` field should follow RFC3339", v))
				continue
			}
			configPage.UpdatedAt = st
		default:
			state.errorSet.Add(state.buildParseError(fmt.Sprintf("a markdown style page cannot accept the key: %s", key), item))
		}
	}
	// Ideally, we should return an error if any of the required fields are missing.
	// But in the current implementation, some fields will be populated at a later stage.
	// So, we can't validate them here.
	return configPage
}

func parseConfigPageMatch(state *ParseState, mapping *ast.MappingNode) ConfigPage { //nolint: cyclop
	// a match object has the following fields:
	// {
	//   "match": "docs/*.md",
	//   "sort_key": "title",
	//   "sort_order": "asc" // `asc` or `desc`
	// }

	configPage := ConfigPage{}
	for _, item := range mapping.Values {
		key := item.Key.String()
		switch key {
		case "match":
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`match` field should be a string", item.Value))
				continue
			}
			configPage.Match = v.Value
		case "sort_key":
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`sort_key` field should be a string", item.Value))
				continue
			}
			text := strings.ToLower(v.Value)
			if text != "title" && text != "updated_at" {
				state.errorSet.Add(state.buildParseError("`sort_key` should be either `title` or `updated_at`", item.Value))
				continue
			}
			configPage.SortKey = text
		case "sort_order":
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`sort_order` should be either `asc` or `desc`", item.Value))
				continue
			}
			text := strings.ToLower(v.Value)
			if text != "asc" && text != "desc" {
				state.errorSet.Add(state.buildParseError("`sort_order` should be either `asc` or `desc`", item.Value))
				continue
			}
			configPage.SortOrder = text
		default:
			state.errorSet.Add(state.buildParseError("a match style page cannot accept the key", item))
		}
	}

	// Validate the fields.
	if configPage.SortKey == "" && configPage.SortOrder != "" {
		state.errorSet.Add(state.buildParseError("`sort_key` should not be empty if you specify `sort_order`", mapping))
	}
	return configPage
}

func parseConfigPageDirectory(state *ParseState, mapping *ast.MappingNode) ConfigPage {
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
				state.errorSet.Add(state.buildParseError("`directory` field should be a string", item.Value))
				continue
			}
			configPage.Directory = v.Value
		case "children":
			v, ok := item.Value.(*ast.SequenceNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`children` field should be a sequence", item.Value))
				continue
			}
			configPage.Children = parseConfigPageSequence(state, v)
		default:
			state.errorSet.Add(state.buildParseError("a directory style page cannot accept the key", item))
		}
	}

	if configPage.Directory == "" {
		state.errorSet.Add(state.buildParseError("the `directory` field is required", mapping))
	}
	return configPage
}

func parseConfigAssets(state *ParseState, node *ast.MappingValueNode) {
	// This function should be called only once.
	// Receive an object like the following:
	//
	// assets:
	//   - "assets/**"
	//   - "images/**"

	if state.isAssetsAlreadyParsed {
		state.errorSet.Add(state.buildParseError("there should be exactly one `assets` section at the top level", node))
		return
	}
	state.isAssetsAlreadyParsed = true

	sequence, ok := node.Value.(*ast.SequenceNode)
	if !ok {
		state.errorSet.Add(state.buildParseError("the `assets` field should be a sequence type", node.Value))
		return
	}

	assets := make([]ConfigAsset, 0, len(sequence.Values))
	for _, item := range sequence.Values {
		v, ok := item.(*ast.StringNode)
		if !ok {
			state.errorSet.Add(state.buildParseError("an item in the `sequence` field should have a string type", item))
			continue
		}
		assets = append(assets, ConfigAsset(v.Value))
	}
	state.config.Assets = assets
}
