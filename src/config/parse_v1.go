package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/caarlos0/log"
	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/mattn/go-zglob"
	appErrors "github.com/toritoritori29/dodo-cli/src/errors"
	"golang.org/x/text/language"
)

const (
	ConfigPageTypeUnknown = iota
	ConfigPageTypeMarkdown
	ConfigPageTypeMatch
	ConfigPageTypeDirectory
)

const (
	ConfigPageKeyMarkdown    = "markdown"
	ConfigPageKeyTitle       = "title"
	ConfigPageKeyPath        = "path"
	ConfigPageKeyDescription = "description"
	ConfigPageKeyUpdatedAt   = "updated_at"
	ConfigPageKeyCreatedAt   = "created_at"
)

const (
	ConfigPageMatchKeyMatch     = "match"
	ConfigPageMatchKeySortKey   = "sort_key"
	ConfigPageMatchKeySortOrder = "sort_order"
)

const (
	ConfigPageDirectoryKeyDirectory = "directory"
	ConfigPageDirectoryKeyChildren  = "children"
)

type ConfigV1 struct {
	Version string
	Project ConfigProjectV1
	Pages   []ConfigPageV1
	Assets  []ConfigAssetV1
}

type ConfigProjectV1 struct {
	ProjectID       string
	Name            string
	Description     string
	Version         string
	Logo            string
	Repository      string
	DefaultLanguage string
}

type ConfigPageV1 struct {
	// markdown syntax
	Markdown    string
	Title       string
	Path        string
	Description string
	UpdatedAt   SerializableTime
	CreatedAt   SerializableTime

	// directory syntax
	Directory string
	Children  []ConfigPageV1

	// NOTE: match syntax is translated to a markdown statement.
}

// Check if the page is a valid single page.
func (c *ConfigPageV1) MatchMarkdown() bool {
	return c.Markdown != ""
}

func (c *ConfigPageV1) MatchDirectory() bool {
	return c.Directory != ""
}

func (c *ConfigPageV1) isValidTitle() error {
	if c.Title == "" {
		return errors.New("the `title` field is required")
	}
	return nil
}

func (c *ConfigPageV1) isValidPath() error {
	// The path must contain only alphanumeric characters, periods (.), underscores (_), and hyphens (-).
	if c.Path == "" {
		return errors.New("the `path` field is required")
	}
	matched, err := regexp.MatchString("^[a-zA-Z-0-9_-]*$", c.Path)
	if err != nil || !matched {
		return fmt.Errorf("the path `%s` contains invalid characters. Paths can only contain alphanumeric characters, underscores (_), and hyphens (-)", c.Path)
	}
	return nil
}

type ConfigAssetV1 string

func (m ConfigAssetV1) List(rootDir string) ([]string, error) {
	globPath := filepath.Clean(filepath.Join(rootDir, string(m)))
	if err := IsUnderRootPath(rootDir, globPath); err != nil {
		return nil, fmt.Errorf("invalid configuration: path must be under the root directory: path: %s", globPath)
	}

	matches, err := zglob.Glob(globPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list files matching '%s': %w", globPath, err)
	}
	return matches, nil
}

// A struct to keep the state of the parsing process.
type ParseStateV1 struct {
	filepath                  string   // The name of the file being parsed.
	config                    ConfigV1 // The config object being generated.
	contents                  []byte   // The contents of the file being parsed.
	rootPath                  string
	isVersionAlreadyParsed    bool
	isProjectAlreadyParsed    bool
	isPagesAlreadyParsed      bool
	isAssetsAlreadyParsed     bool
	isAnnotationAlreadyParsed bool
	errorSet                  appErrors.MultiError
}

func NewParseStateV1(filepath, workingDir string) *ParseStateV1 {
	return &ParseStateV1{
		filepath: filepath,
		rootPath: workingDir,
	}
}

func (s *ParseStateV1) buildParseError(message string, node ast.Node) error {
	line := s.getLineFromNode(node)
	return &appErrors.ParseError{
		Filepath: s.filepath,
		Message:  message,
		Line:     line,
		Node:     node,
	}
}

func (s *ParseStateV1) getLineFromNode(node ast.Node) string {
	lines := bytes.Split(s.contents, []byte("\n"))
	lineNumber := node.GetToken().Position.Line - 1
	if lineNumber < 0 || lineNumber >= len(lines) {
		return "(unknown line)"
	}
	return string(lines[lineNumber])
}

// getAbsolutePath converts a relative path to an absolute path and validates it's under the root directory.
func (s *ParseStateV1) getAbsolutePath(path string) (string, error) {
	absRootPath, err := filepath.Abs(s.rootPath)
	if err != nil {
		return "", fmt.Errorf("failed to get the absolute path of the working directory: %w", err)
	}

	absPath, err := filepath.Abs(filepath.Join(s.rootPath, path))
	if err != nil {
		return "", fmt.Errorf("failed to get the absolute path of the specified file: %w", err)
	}

	absRootPath = filepath.Clean(absRootPath)
	absPath = filepath.Clean(absPath)

	if !strings.HasPrefix(absPath, absRootPath) {
		return "", fmt.Errorf("the file being parsed is not under the root directory: %s", absPath)
	}
	return absPath, nil
}

// ParseConfigV1 takes a reader and parses it into a ConfigV1 struct.
// While parsing, it validates the config at the same time.
// This is to provide a user-friendly error message.
//
// This function follows the implementation at:
// https://github.com/goccy/go-yaml/blob/abc70836f5a5623a92cf51d4bf40cbaf8fed2faa/decode.go
func ParseConfigV1(state *ParseStateV1, reader io.Reader) (*ConfigV1, error) {
	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read a document config: %w", err)
	}
	contents := buf.Bytes()
	state.contents = contents

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

func parseRoot(state *ParseStateV1, root *ast.File) {
	if len(root.Docs) != 1 {
		state.errorSet.Add(fmt.Errorf("there should be only one document. Got %d", len(root.Docs)))
		return
	}

	body, ok := root.Docs[0].Body.(*ast.MappingNode)
	if !ok {
		state.errorSet.Add(state.buildParseError("the root node must be of mapping type", root.Docs[0].Body))
		return
	}

	// Apply parseRootItem to each node at the root level.
	for _, node := range body.Values {
		parseRootItem(state, node)
	}

	// Check if all required fields are parsed.
	// NOTE: The `assets` field is optional.
	if !state.isVersionAlreadyParsed {
		state.errorSet.Add(errors.New("the `version` field is required"))
	}
	if !state.isProjectAlreadyParsed {
		state.errorSet.Add(errors.New("the `project` field is required"))
	}
	if !state.isPagesAlreadyParsed {
		state.errorSet.Add(errors.New("the `pages` field is required"))
	}
}

func parseRootItem(state *ParseStateV1, node ast.Node) {
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
	if mapping.Key.String() == "annotation" {
		parseConfigAnnotation(state, mapping)
		return
	}
	state.errorSet.Add(state.buildParseError("unexpected key at the top level", mapping.Key))
}

func parseVersion(state *ParseStateV1, node *ast.MappingValueNode) {
	// This function should be called only once.
	// Receives an object like the following:
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
		state.errorSet.Add(state.buildParseError("`version` must have an integer value", node.Value))
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

func parseConfigProject(state *ParseStateV1, node *ast.MappingValueNode) { //nolint: cyclop, funlen
	// This function should be called only once.
	// Receives an object like the following:
	//
	// project:
	//   project_id: "project_123"
	//   name: "My Project"
	//   description: "This is my project"
	//   version: "1.0.0"

	if state.isProjectAlreadyParsed {
		state.errorSet.Add(state.buildParseError("there should be exactly one `project` section at the top level", node))
		return
	}
	state.isProjectAlreadyParsed = true

	// node.value should be a mapping type.
	// ConfigProjectV1 needs name, description, and version.
	// And all of them are string type.
	children, ok := node.Value.(*ast.MappingNode)
	if !ok {
		state.errorSet.Add(state.buildParseError("the `project` must have a mapping value", node.Value))
		return
	}

	for _, item := range children.Values {
		key := item.Key.String()
		switch key {
		case "project_id":
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`project_id` field must be a string", item.Value))
				continue
			}
			state.config.Project.ProjectID = v.Value
		case "name":
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`name` field must be a string", item.Value))
				continue
			}
			state.config.Project.Name = v.Value
		case "description":
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`description` field must be a string", item.Value))
				continue
			}
			state.config.Project.Description = v.Value
		case "version":
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`version` field must be a string", item.Value))
				continue
			}
			state.config.Project.Version = v.Value
		case "logo":
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`logo` field must be a string", item.Value))
				continue
			}
			state.config.Project.Logo = v.Value
		case "repository":
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`repository` field must be a string", item.Value))
				continue
			}
			state.config.Project.Repository = v.Value
		case "default_language":
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`default_language` field must be a string", item.Value))
				continue
			}
			defaultLanguage := strings.ToLower(v.Value)

			state.config.Project.DefaultLanguage = defaultLanguage
		default:
			state.errorSet.Add(state.buildParseError("the `project` does not accept the key: "+key, item))
		}
	}

	// Post processing and validation.
	if state.config.Project.DefaultLanguage == "" {
		state.config.Project.DefaultLanguage = "en"
	}

	// Validate the required fields.
	if state.config.Project.ProjectID == "" {
		state.errorSet.Add(state.buildParseError("the `project` must have a `project_id` field longer than 1 character", node))
	}
	if state.config.Project.Name == "" {
		state.errorSet.Add(state.buildParseError("the `project` must have a `name` field longer than 1 character", node))
	}
	if !isValidISOLanguageCode(state.config.Project.DefaultLanguage) {
		state.errorSet.Add(state.buildParseError(fmt.Sprintf("`default_language` field must be a valid ISO 639-1 language code (e.g., 'ja'). given: %s", state.config.Project.DefaultLanguage), node))
	}

	// Validate the repository field.
	repoURL := state.config.Project.Repository
	if repoURL != "" {
		_, err := url.ParseRequestURI(repoURL)
		if err != nil {
			state.errorSet.Add(state.buildParseError("the `repository` field must be a valid URL", node))
		}
	}
}

func isValidISOLanguageCode(code string) bool {
	if len(code) != 2 {
		return false
	}
	lower := strings.ToLower(code)
	if _, err := language.ParseBase(lower); err == nil {
		return true
	}
	return false
}

func parseConfigPage(state *ParseStateV1, node *ast.MappingValueNode) {
	// This function should be called only once.
	// Receives an object like the following:
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
		state.errorSet.Add(state.buildParseError("the `pages` field must be a sequence type", node.Value))
		return
	}
	state.config.Pages = parseConfigPageSequence(state, sequence)
}

func parseConfigPageSequence(state *ParseStateV1, sequence *ast.SequenceNode) []ConfigPageV1 {
	// Receives an object like the following:
	//
	// xxx:
	//   - markdown: "README1.md"
	//     ...
	//   - markdown: "README2.md"
	//     ...

	configPages := make([]ConfigPageV1, 0, len(sequence.Values))
	for _, item := range sequence.Values {
		pageNode, ok := item.(*ast.MappingNode)
		if !ok {
			state.errorSet.Add(state.buildParseError("each item in the `pages` sequence must be of mapping type", item))
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
			configPages = append(configPages, p...)
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
		case ConfigPageKeyMarkdown:
			return ConfigPageTypeMarkdown
		case ConfigPageMatchKeyMatch:
			return ConfigPageTypeMatch
		case ConfigPageDirectoryKeyDirectory:
			return ConfigPageTypeDirectory
		}
	}
	return ConfigPageTypeUnknown
}

func parseConfigPageMarkdown(state *ParseStateV1, mapping *ast.MappingNode) ConfigPageV1 { //nolint: cyclop, funlen
	// A markdown object has the following fields:
	// {
	//   "markdown": "README1.md",
	//   "title": "README1",
	//   "path": "readme1",
	//   "description": "This is README1",
	//   "updated_at": "2021-01-01T00:00:00Z"
	// }

	configPage := ConfigPageV1{}

	for _, item := range mapping.Values {
		key := item.Key.String()

		switch key {
		case ConfigPageKeyMarkdown:
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`markdown` field must be a string", item.Value))
				continue
			}
			configPage.Markdown = v.Value
		case ConfigPageKeyTitle:
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`title` field must be a string", item.Value))
				continue
			}
			configPage.Title = v.Value
		case ConfigPageKeyPath:
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`path` field must be a string", item.Value))
				continue
			}
			configPage.Path = v.Value
		case ConfigPageKeyDescription:
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`description` field must be a string", item.Value))
				continue
			}
			configPage.Description = v.Value
		case ConfigPageKeyUpdatedAt:
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`updated_at` field must be a string", item.Value))
				continue
			}
			st, err := NewSerializableTime(v.Value)
			if err != nil {
				state.errorSet.Add(state.buildParseError("`updated_at` field must follow RFC3339", v))
				continue
			}
			configPage.UpdatedAt = st
		case ConfigPageKeyCreatedAt:
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`created_at` field must be a string", item.Value))
				continue
			}
			st, err := NewSerializableTime(v.Value)
			if err != nil {
				state.errorSet.Add(state.buildParseError("`created_at` field must follow RFC3339", v))
				continue
			}
			configPage.CreatedAt = st
		default:
			state.errorSet.Add(state.buildParseError("a markdown style page cannot accept the key: "+key, item))
		}
	}

	// If the markdown has a frontmatter, populate the empty fields with it.
	// Then validate the fields.
	fillFieldsFromMarkdown(state, &configPage, mapping)
	validateMarkdownPage(state, &configPage, mapping)
	return configPage
}

func fillFieldsFromMarkdown(state *ParseStateV1, configPage *ConfigPageV1, mapping *ast.MappingNode) { //nolint: cyclop
	if configPage.Markdown == "" {
		state.errorSet.Add(state.buildParseError("the `markdown` field is required", mapping))
		return
	}
	clean, err := state.getAbsolutePath(configPage.Markdown)
	if err != nil {
		state.errorSet.Add(state.buildParseError(err.Error(), mapping))
		return
	}

	// First, populate the fields from the markdown front matter.
	p, err := NewFrontMatterFromMarkdown(clean)
	if err != nil {
		message := fmt.Sprintf("cannot read the markdown file: %s, %v", configPage.Markdown, err.Error())
		state.errorSet.Add(state.buildParseError(message, mapping))
		return
	}

	if configPage.Title == "" && p.Title != "" {
		configPage.Title = p.Title
	}
	if configPage.Path == "" && p.Link != "" {
		configPage.Path = p.Link
	}
	if configPage.Description == "" && p.Description != "" {
		configPage.Description = p.Description
	}
	if !configPage.UpdatedAt.HasValue() && !p.UpdatedAt.HasValue() {
		configPage.UpdatedAt = p.UpdatedAt
	}
	if !configPage.CreatedAt.HasValue() && !p.CreatedAt.HasValue() {
		configPage.CreatedAt = p.CreatedAt
	}
}

func validateMarkdownPage(state *ParseStateV1, configPage *ConfigPageV1, mapping *ast.MappingNode) bool {
	ok := true
	if err := configPage.isValidTitle(); err != nil {
		state.errorSet.Add(state.buildParseError(err.Error(), mapping))
		ok = false
	}
	if err := configPage.isValidPath(); err != nil {
		state.errorSet.Add(state.buildParseError(err.Error(), mapping))
		ok = false
	}
	return ok
}

func parseConfigPageMatch(state *ParseStateV1, mapping *ast.MappingNode) []ConfigPageV1 { //nolint: cyclop
	// A match object has the following fields:
	// {
	//   "match": "docs/*.md",
	//   "sort_key": "title",
	//   "sort_order": "asc" // `asc` or `desc`
	// }

	var match string
	var sortKey string
	var sortOrder string

	for _, item := range mapping.Values {
		key := item.Key.String()
		switch key {
		case ConfigPageMatchKeyMatch:
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`match` field must be a string", item.Value))
				continue
			}
			match = v.Value
		case ConfigPageMatchKeySortKey:
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`sort_key` field must be a string", item.Value))
				continue
			}
			text := strings.ToLower(v.Value)
			if text != "title" && text != "updated_at" && text != "created_at" {
				state.errorSet.Add(state.buildParseError("`sort_key` must be either `title` or `updated_at`", item.Value))
				continue
			}
			sortKey = text
		case ConfigPageMatchKeySortOrder:
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`sort_order` must be either `asc` or `desc`", item.Value))
				continue
			}
			text := strings.ToLower(v.Value)
			if text != "asc" && text != "desc" {
				state.errorSet.Add(state.buildParseError("`sort_order` must be either `asc` or `desc`", item.Value))
				continue
			}
			sortOrder = text
		default:
			state.errorSet.Add(state.buildParseError("a match style page cannot accept the key", item))
		}
	}

	// Validate the fields.
	if sortKey == "" && sortOrder != "" {
		state.errorSet.Add(state.buildParseError("`sort_key` must not be empty if you specify `sort_order`", mapping))
		return nil
	}
	return buildConfigPageFromMatchStatement(state, mapping, match, sortKey, sortOrder)
}

func buildConfigPageFromMatchStatement(state *ParseStateV1, mapping *ast.MappingNode, match, sortKey, sortOrder string) []ConfigPageV1 {
	clean, err := state.getAbsolutePath(match)
	if err != nil {
		state.errorSet.Add(state.buildParseError(err.Error(), mapping))
		return nil
	}

	matches, err := zglob.Glob(clean)
	if err != nil {
		state.errorSet.Add(state.buildParseError(fmt.Sprintf("failed to list files matching '%s': %v", match, err), mapping))
		return nil
	}

	pages := make([]ConfigPageV1, 0, len(matches))
	for _, m := range matches {
		matter, err := NewFrontMatterFromMarkdown(m)
		if err != nil {
			message := fmt.Sprintf("%s: %s", err.Error(), m)
			state.errorSet.Add(state.buildParseError(message, mapping))
			continue
		}

		p := ConfigPageV1{
			Markdown:    m,
			Title:       matter.Title,
			Path:        matter.Link,
			Description: matter.Description,
			UpdatedAt:   matter.UpdatedAt,
			CreatedAt:   matter.CreatedAt,
		}

		if ok := validateMatchPage(state, &p, mapping); !ok {
			continue
		}
		log.Debugf("Node Found. Type: Markdown, Filepath: %s, Title: %s, Path: %s", p.Markdown, p.Title, p.Path)
		pages = append(pages, p)
	}
	if err := sortPageSlice(sortKey, sortOrder, pages); err != nil {
		state.errorSet.Add(state.buildParseError(err.Error(), mapping))
		return nil
	}
	return pages
}

func validateMatchPage(state *ParseStateV1, configPage *ConfigPageV1, mapping *ast.MappingNode) bool {
	// Almost the same as validateMarkdownPage.
	// But the error message is different.
	ok := true
	if err := configPage.isValidTitle(); err != nil {
		message := fmt.Sprintf("%v: %s", err, configPage.Markdown)
		state.errorSet.Add(state.buildParseError(message, mapping))
		ok = false
	}
	if err := configPage.isValidPath(); err != nil {
		message := fmt.Sprintf("%v: %s", err, configPage.Markdown)
		state.errorSet.Add(state.buildParseError(message, mapping))
		ok = false
	}
	return ok
}

func sortPageSlice(sortKey, sortOrder string, pages []ConfigPageV1) error { //nolint: cyclop
	if sortKey == "" && sortOrder == "" {
		return nil
	}
	if sortKey == "" {
		return errors.New("sort key is not provided")
	}
	// Check sortOrder
	isASC := true
	if sortOrder != "" {
		switch strings.ToLower(sortOrder) {
		case "asc":
			break
		case "desc":
			isASC = false
		default:
			return fmt.Errorf("invalid sort order: `%s`", sortOrder)
		}
	}

	if sortKey == "title" {
		sort.Slice(pages, func(i, j int) bool {
			return (pages[i].Title < pages[j].Title) == isASC
		})
		return nil
	}
	if sortKey == "updated_at" {
		sort.Slice(pages, func(i, j int) bool {
			return (pages[i].UpdatedAt.Before(pages[j].UpdatedAt.Time)) == isASC
		})
		return nil
	}
	if sortKey == "created_at" {
		sort.Slice(pages, func(i, j int) bool {
			return (pages[i].CreatedAt.Before(pages[j].CreatedAt.Time)) == isASC
		})
		return nil
	}
	return fmt.Errorf("invalid sort key: %s", sortKey)
}

func parseConfigPageDirectory(state *ParseStateV1, mapping *ast.MappingNode) ConfigPageV1 {
	// A directory object has the following fields:
	//
	// {
	//   "directory": "path/to/directory",
	//   "children": [
	//     {
	//       "markdown": "README1.md",
	//       ...
	//     },
	// }

	configPage := ConfigPageV1{}
	for _, item := range mapping.Values {
		key := item.Key.String()
		switch key {
		case "directory":
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`directory` field must be a string", item.Value))
				continue
			}
			configPage.Directory = v.Value
		case "children":
			v, ok := item.Value.(*ast.SequenceNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`children` field must be a sequence", item.Value))
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

func parseConfigAssets(state *ParseStateV1, node *ast.MappingValueNode) {
	// This function should be called only once.
	// Receives an object like the following:
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
		state.errorSet.Add(state.buildParseError("the `assets` field must be a sequence type", node.Value))
		return
	}

	assets := make([]ConfigAssetV1, 0, len(sequence.Values))
	for _, item := range sequence.Values {
		v, ok := item.(*ast.StringNode)
		if !ok {
			state.errorSet.Add(state.buildParseError("an item in the `sequence` field must have a string type", item))
			continue
		}
		assets = append(assets, ConfigAssetV1(v.Value))
	}
	state.config.Assets = assets
}

func parseConfigAnnotation(state *ParseStateV1, node *ast.MappingValueNode) {
	// This function should be called only once.
	// annotation:
	//   foo: bar
	if state.isAnnotationAlreadyParsed {
		state.errorSet.Add(state.buildParseError("there should be exactly one `annotation` section at the top level", node))
		return
	}
	state.isAnnotationAlreadyParsed = true

	var annotation any
	if err := yaml.NodeToValue(node.Value, &annotation); err != nil {
		state.errorSet.Add(state.buildParseError("failed to parse `annotation` field", node.Value))
		return
	}
}
