package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/mattn/go-zglob"
	appErrors "github.com/toritoritori29/dodo-cli/src/errors"
)

const (
	ConfigPageTypeMarkdownV2              = "markdown"
	ConfigPageTypeMarkdownMultiLanguageV2 = "markdown_multilanguage"
	ConfigPageTypeMatchV2                 = "match"
	ConfigPageTypeDirectoryV2             = "directory"
	ConfigPageTypeSectionV2               = "section"
)

const (
	ConfigPageV2KeyType     = "type"
	ConfigPageV2KeyLink     = "link"
	ConfigPageV2KeyTitle    = "title"
	ConfigPageV2KeyFilepath = "filepath"
	ConfigPageV2KeyLang     = "lang"
	ConfigPageV2KeyPattern  = "pattern"
	ConfigPageV2KeySortKey  = "sort_key"
	ConfigPageV2KeySortOrd  = "sort_order"
	ConfigPageV2KeyChildren = "children"
)

const (
	ConfigPageV2LegacyKeyMatch     = "match"
	ConfigPageV2LegacyKeyDirectory = "directory"
)

type ConfigV2 struct {
	Version string
	Project ConfigProjectV2
	Pages   []ConfigPageV2
	Assets  []ConfigAssetV2
}

type ConfigProjectV2 struct {
	ProjectID       string
	Name            string
	Description     string
	Version         string
	Logo            string
	Repository      string
	DefaultLanguage string
}

type ConfigPageV2 struct {
	// Type defines how this page entry should be parsed (markdown/markdown_multilanguage/match/directory/section).
	Type string

	// Lang holds per-locale variants (keyed by ISO 639-1 code).
	Lang map[string]ConfigPageLangV2

	// Path is used by section pages to define URL path segments.
	Path string

	// Match pattern fields for auto-discovery.
	Pattern   string
	SortKey   string
	SortOrder string

	// Children is used by directory/section entries.
	Children []ConfigPageV2
}

type ConfigPageLangV2 struct {
	// Link/Title/Filepath are the localized markdown fields.
	Link     string
	Title    string
	Filepath string
}

type ConfigAssetV2 string

func (m ConfigAssetV2) List(rootDir string) ([]string, error) {
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

type ParseStateV2 struct {
	filepath                  string
	config                    ConfigV2
	contents                  []byte
	rootPath                  string
	isVersionAlreadyParsed    bool
	isProjectAlreadyParsed    bool
	isPagesAlreadyParsed      bool
	isAssetsAlreadyParsed     bool
	isAnnotationAlreadyParsed bool
	errorSet                  appErrors.MultiError
}

func NewParseStateV2(filepath, workingDir string) *ParseStateV2 {
	return &ParseStateV2{
		filepath: filepath,
		rootPath: workingDir,
	}
}

func (s *ParseStateV2) buildParseError(message string, node ast.Node) error {
	line := s.getLineFromNode(node)
	return &appErrors.ParseError{
		Filepath: s.filepath,
		Message:  message,
		Line:     line,
		Node:     node,
	}
}

func (s *ParseStateV2) getLineFromNode(node ast.Node) string {
	lines := bytes.Split(s.contents, []byte("\n"))
	lineNumber := node.GetToken().Position.Line - 1
	if lineNumber < 0 || lineNumber >= len(lines) {
		return "(unknown line)"
	}
	return string(lines[lineNumber])
}

func (s *ParseStateV2) getSecurePath(path string) (string, error) {
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

func ParseConfigV2(state *ParseStateV2, reader io.Reader) (*ConfigV2, error) {
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, reader); err != nil {
		return nil, fmt.Errorf("failed to read a document config: %w", err)
	}
	contents := buf.Bytes()
	state.contents = contents

	root, err := parser.ParseBytes(contents, parser.Mode(0))
	if err != nil {
		return nil, fmt.Errorf("failed to parse a document config: %w", err)
	}
	parseRootV2(state, root)

	if state.errorSet.HasError() {
		return nil, &state.errorSet
	}
	return &state.config, nil
}

func parseRootV2(state *ParseStateV2, root *ast.File) {
	if len(root.Docs) != 1 {
		state.errorSet.Add(fmt.Errorf("there should be only one document. Got %d", len(root.Docs)))
		return
	}

	body, ok := root.Docs[0].Body.(*ast.MappingNode)
	if !ok {
		state.errorSet.Add(state.buildParseError("the root node must be of mapping type", root.Docs[0].Body))
		return
	}

	// Pages parsing depends on project.default_language, so defer until after project is parsed.
	var pagesNode *ast.MappingValueNode
	for _, mapping := range body.Values {
		key := mapping.Key.String()
		switch key {
		case "version":
			parseVersionV2(state, mapping)
		case "project":
			parseConfigProjectV2(state, mapping)
		case "pages":
			if pagesNode != nil {
				state.errorSet.Add(state.buildParseError("there should be exactly one `pages` section at the top level", mapping))
				continue
			}
			pagesNode = mapping
		case "assets":
			parseConfigAssetsV2(state, mapping)
		case "annotation":
			parseConfigAnnotationV2(state, mapping)
		default:
			state.errorSet.Add(state.buildParseError("unexpected key at the top level", mapping.Key))
		}
	}

	if pagesNode != nil {
		parseConfigPageV2(state, pagesNode)
	}

	// Validation
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

func parseVersionV2(state *ParseStateV2, node *ast.MappingValueNode) {
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

	if versionNum != 2 {
		state.errorSet.Add(state.buildParseError("unsupported version: only '2' is supported now", intNode))
		return
	}
	state.config.Version = strconv.Itoa(versionNum)
}

func parseConfigProjectV2(state *ParseStateV2, node *ast.MappingValueNode) { //nolint: cyclop, funlen
	if state.isProjectAlreadyParsed {
		state.errorSet.Add(state.buildParseError("there should be exactly one `project` section at the top level", node))
		return
	}
	state.isProjectAlreadyParsed = true

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
			state.config.Project.DefaultLanguage = strings.ToLower(v.Value)
		default:
			state.errorSet.Add(state.buildParseError("the `project` does not accept the key: "+key, item))
		}
	}

	if state.config.Project.ProjectID == "" {
		state.errorSet.Add(state.buildParseError("the `project` must have a `project_id` field longer than 1 character", node))
	}
	if state.config.Project.Name == "" {
		state.errorSet.Add(state.buildParseError("the `project` must have a `name` field longer than 1 character", node))
	}
	if state.config.Project.DefaultLanguage == "" {
		state.config.Project.DefaultLanguage = "en"
	}
	if !isValidISOLanguageCode(state.config.Project.DefaultLanguage) {
		message := fmt.Sprintf("`default_language` field must be a valid ISO 639-1 language code (e.g., 'ja'). given: %s", state.config.Project.DefaultLanguage)
		state.errorSet.Add(state.buildParseError(message, node))
	}

	repoURL := state.config.Project.Repository
	if repoURL != "" {
		if _, err := url.ParseRequestURI(repoURL); err != nil {
			state.errorSet.Add(state.buildParseError("the `repository` field must be a valid URL", node))
		}
	}
}

func parseConfigPageV2(state *ParseStateV2, node *ast.MappingValueNode) {
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
	state.config.Pages = parseConfigPageSequenceV2(state, sequence)
}

func parseConfigPageSequenceV2(state *ParseStateV2, sequence *ast.SequenceNode) []ConfigPageV2 {
	configPages := make([]ConfigPageV2, 0, len(sequence.Values))
	for _, item := range sequence.Values {
		pageNode, ok := item.(*ast.MappingNode)
		if !ok {
			state.errorSet.Add(state.buildParseError("each item in the `pages` sequence must be of mapping type", item))
			continue
		}

		pageType := estimateConfigPageTypeV2(state, pageNode)
		if pageType == "" {
			continue
		}

		switch pageType {
		case ConfigPageTypeMarkdownV2:
			p := parseConfigPageMarkdownV2(state, pageNode)
			configPages = append(configPages, p)
		case ConfigPageTypeMarkdownMultiLanguageV2:
			p := parseConfigPageMarkdownMultiLanguageV2(state, pageNode)
			configPages = append(configPages, p)
		case ConfigPageTypeMatchV2:
			pages := parseConfigPageMatchV2(state, pageNode)
			configPages = append(configPages, pages...)
		case ConfigPageTypeDirectoryV2:
			p := parseConfigPageDirectoryV2(state, pageNode)
			configPages = append(configPages, p)
		case ConfigPageTypeSectionV2:
			p := parseConfigPageSectionV2(state, pageNode)
			configPages = append(configPages, p)
		default:
			state.errorSet.Add(state.buildParseError("unknown page type", pageNode))
		}
	}
	return configPages
}

func estimateConfigPageTypeV2(state *ParseStateV2, mapping *ast.MappingNode) string {
	hasLang := false
	for _, item := range mapping.Values {
		if item.Key.String() == ConfigPageV2KeyLang {
			hasLang = true
			break
		}
	}
	for _, item := range mapping.Values {
		if item.Key.String() != ConfigPageV2KeyType {
			continue
		}
		v, ok := item.Value.(*ast.StringNode)
		if !ok {
			state.errorSet.Add(state.buildParseError("`type` field must be a string", item.Value))
			return ""
		}
		switch strings.ToLower(v.Value) {
		case ConfigPageTypeMarkdownV2:
			if hasLang {
				return ConfigPageTypeMarkdownMultiLanguageV2
			}
			return ConfigPageTypeMarkdownV2
		case ConfigPageTypeMatchV2, ConfigPageTypeDirectoryV2, ConfigPageTypeSectionV2:
			return strings.ToLower(v.Value)
		default:
			state.errorSet.Add(state.buildParseError("unknown page type: "+v.Value, item.Value))
			return ""
		}
	}

	for _, item := range mapping.Values {
		switch item.Key.String() {
		case ConfigPageV2LegacyKeyMatch:
			return ConfigPageTypeMatchV2
		case ConfigPageV2LegacyKeyDirectory:
			return ConfigPageTypeDirectoryV2
		}
	}

	state.errorSet.Add(state.buildParseError("this mapping does not match any page type", mapping))
	return ""
}

func parseConfigPageMarkdownV2(state *ParseStateV2, mapping *ast.MappingNode) ConfigPageV2 { //nolint: cyclop, funlen
	configPage := ConfigPageV2{
		Type: ConfigPageTypeMarkdownV2,
	}
	langConfig := ConfigPageLangV2{}

	for _, item := range mapping.Values {
		key := item.Key.String()
		switch key {
		case ConfigPageV2KeyType:
			continue
		case ConfigPageV2KeyLang:
			state.errorSet.Add(state.buildParseError("`lang` cannot be used in single-locale markdown", item.Value))
		case ConfigPageV2KeyLink:
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`link` field must be a string", item.Value))
				continue
			}
			langConfig.Link = v.Value
		case ConfigPageV2KeyTitle:
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`title` field must be a string", item.Value))
				continue
			}
			langConfig.Title = v.Value
		case ConfigPageV2KeyFilepath:
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`filepath` field must be a string", item.Value))
				continue
			}
			langConfig.Filepath = v.Value
		default:
			state.errorSet.Add(state.buildParseError("a markdown style page cannot accept the key: "+key, item))
		}
	}

	fillSingleLangFromMarkdownV2(state, &langConfig, mapping)
	validateMarkdownLangEntryV2(state, &langConfig, mapping)
	defaultLang := state.config.Project.DefaultLanguage
	if defaultLang == "" {
		defaultLang = "en"
	}
	configPage.Lang = map[string]ConfigPageLangV2{
		defaultLang: langConfig,
	}
	return configPage
}

func parseConfigPageMarkdownMultiLanguageV2(state *ParseStateV2, mapping *ast.MappingNode) ConfigPageV2 { //nolint: cyclop, funlen
	configPage := ConfigPageV2{
		Type: ConfigPageTypeMarkdownMultiLanguageV2,
	}
	var langNode *ast.MappingNode

	for _, item := range mapping.Values {
		key := item.Key.String()
		switch key {
		case ConfigPageV2KeyType:
			continue
		case ConfigPageV2KeyLang:
			v, ok := item.Value.(*ast.MappingNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`lang` field must be a mapping", item.Value))
				continue
			}
			langNode = v
		case ConfigPageV2KeyLink, ConfigPageV2KeyTitle, ConfigPageV2KeyFilepath:
			state.errorSet.Add(state.buildParseError("single-locale fields cannot be used with `lang`", item.Value))
		default:
			state.errorSet.Add(state.buildParseError("a markdown style page cannot accept the key: "+key, item))
		}
	}

	if langNode == nil {
		state.errorSet.Add(state.buildParseError("`lang` field is required for multi-locale markdown", mapping))
		return configPage
	}
	configPage.Lang = parseMarkdownLangMapV2(state, langNode)
	validateLangMapV2(state, configPage.Lang, mapping)
	return configPage
}

func parseMarkdownLangMapV2(state *ParseStateV2, langNode *ast.MappingNode) map[string]ConfigPageLangV2 { //nolint: cyclop, funlen
	langMap := make(map[string]ConfigPageLangV2)
	for _, item := range langNode.Values {
		key := strings.ToLower(item.Key.String())
		valueNode, ok := item.Value.(*ast.MappingNode)
		if !ok {
			state.errorSet.Add(state.buildParseError("each language entry must be a mapping", item.Value))
			continue
		}
		langConfig := ConfigPageLangV2{}
		for _, child := range valueNode.Values {
			childKey := child.Key.String()
			switch childKey {
			case ConfigPageV2KeyLink:
				v, ok := child.Value.(*ast.StringNode)
				if !ok {
					state.errorSet.Add(state.buildParseError("`link` field must be a string", child.Value))
					continue
				}
				langConfig.Link = v.Value
			case ConfigPageV2KeyTitle:
				v, ok := child.Value.(*ast.StringNode)
				if !ok {
					state.errorSet.Add(state.buildParseError("`title` field must be a string", child.Value))
					continue
				}
				langConfig.Title = v.Value
			case ConfigPageV2KeyFilepath:
				v, ok := child.Value.(*ast.StringNode)
				if !ok {
					state.errorSet.Add(state.buildParseError("`filepath` field must be a string", child.Value))
					continue
				}
				langConfig.Filepath = v.Value
			default:
				state.errorSet.Add(state.buildParseError("a markdown language entry cannot accept the key: "+childKey, child))
			}
		}

		fillLangFieldsFromMarkdownV2(state, &langConfig, valueNode)
		validateMarkdownLangEntryV2(state, &langConfig, valueNode)
		langMap[key] = langConfig
	}
	return langMap
}

func fillSingleLangFromMarkdownV2(state *ParseStateV2, langConfig *ConfigPageLangV2, mapping *ast.MappingNode) {
	if langConfig.Filepath == "" {
		state.errorSet.Add(state.buildParseError("the `filepath` field is required", mapping))
		return
	}

	clean, err := state.getSecurePath(langConfig.Filepath)
	if err != nil {
		state.errorSet.Add(state.buildParseError(err.Error(), mapping))
		return
	}

	p, err := NewFrontMatterFromMarkdown(clean)
	if err != nil {
		message := fmt.Sprintf("cannot read the markdown file: %s, %v", langConfig.Filepath, err.Error())
		state.errorSet.Add(state.buildParseError(message, mapping))
		return
	}

	if langConfig.Title == "" && p.Title != "" {
		langConfig.Title = p.Title
	}
	if langConfig.Link == "" && p.Link != "" {
		langConfig.Link = p.Link
	}
}

func fillLangFieldsFromMarkdownV2(state *ParseStateV2, langConfig *ConfigPageLangV2, mapping *ast.MappingNode) {
	if langConfig.Filepath == "" {
		state.errorSet.Add(state.buildParseError("the `filepath` field is required", mapping))
		return
	}

	clean, err := state.getSecurePath(langConfig.Filepath)
	if err != nil {
		state.errorSet.Add(state.buildParseError(err.Error(), mapping))
		return
	}

	p, err := NewFrontMatterFromMarkdown(clean)
	if err != nil {
		message := fmt.Sprintf("cannot read the markdown file: %s, %v", langConfig.Filepath, err.Error())
		state.errorSet.Add(state.buildParseError(message, mapping))
		return
	}

	if langConfig.Title == "" && p.Title != "" {
		langConfig.Title = p.Title
	}
}

func validateMarkdownLangEntryV2(state *ParseStateV2, langConfig *ConfigPageLangV2, mapping *ast.MappingNode) {
	if langConfig.Title == "" {
		state.errorSet.Add(state.buildParseError("the `title` field is required", mapping))
	}
	if langConfig.Link == "" {
		state.errorSet.Add(state.buildParseError("the `link` field is required", mapping))
	}
	if langConfig.Filepath == "" {
		state.errorSet.Add(state.buildParseError("the `filepath` field is required", mapping))
	}
}

func validateLangMapV2(state *ParseStateV2, langMap map[string]ConfigPageLangV2, mapping *ast.MappingNode) {
	if len(langMap) == 0 {
		state.errorSet.Add(state.buildParseError("`lang` must not be empty", mapping))
		return
	}
	for code := range langMap {
		if !isValidISOLanguageCode(code) {
			message := fmt.Sprintf("`lang` key must be a valid ISO 639-1 language code (e.g., 'ja'). given: %s", code)
			state.errorSet.Add(state.buildParseError(message, mapping))
		}
	}
	defaultLang := state.config.Project.DefaultLanguage
	if defaultLang == "" {
		state.errorSet.Add(state.buildParseError("internal error: default_language is empty", mapping))
		return
	}
	if _, ok := langMap[defaultLang]; !ok {
		state.errorSet.Add(state.buildParseError("`lang` must include the default language: "+defaultLang, mapping))
	}
}

func parseConfigPageMatchV2(state *ParseStateV2, mapping *ast.MappingNode) []ConfigPageV2 { //nolint: cyclop
	var pattern string
	var sortKey string
	var sortOrder string

	for _, item := range mapping.Values {
		key := item.Key.String()
		switch key {
		case ConfigPageV2KeyType:
			continue
		case ConfigPageV2KeyPattern:
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`pattern` field must be a string", item.Value))
				continue
			}
			pattern = v.Value
		case ConfigPageV2KeySortKey:
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`sort_key` field must be a string", item.Value))
				continue
			}
			text := strings.ToLower(v.Value)
			if text != "title" {
				state.errorSet.Add(state.buildParseError("`sort_key` must be `title`", item.Value))
				continue
			}
			sortKey = text
		case ConfigPageV2KeySortOrd:
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
		case ConfigPageV2LegacyKeyMatch:
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`match` field must be a string", item.Value))
				continue
			}
			pattern = v.Value
		default:
			state.errorSet.Add(state.buildParseError("a match style page cannot accept the key", item))
		}
	}

	if sortKey == "" && sortOrder != "" {
		state.errorSet.Add(state.buildParseError("`sort_key` must not be empty if you specify `sort_order`", mapping))
		return nil
	}
	return buildConfigPageFromMatchStatementV2(state, mapping, pattern, sortKey, sortOrder)
}

func buildConfigPageFromMatchStatementV2(state *ParseStateV2, mapping *ast.MappingNode, pattern, sortKey, sortOrder string) []ConfigPageV2 {
	clean, err := state.getSecurePath(pattern)
	if err != nil {
		state.errorSet.Add(state.buildParseError(err.Error(), mapping))
		return nil
	}

	matches, err := zglob.Glob(clean)
	if err != nil {
		message := fmt.Sprintf("failed to list files matching '%s': %v", pattern, err)
		state.errorSet.Add(state.buildParseError(message, mapping))
		return nil
	}

	entriesByGroup := make(map[string][]string)
	matterByPath := make(map[string]*FrontMatter)
	for _, m := range matches {
		matter, err := NewFrontMatterFromMarkdown(m)
		if err != nil {
			message := fmt.Sprintf("%s: %s", err.Error(), m)
			state.errorSet.Add(state.buildParseError(message, mapping))
			continue
		}

		groupID := strings.TrimSpace(matter.LanguageGroupID)
		if groupID == "" {
			state.errorSet.Add(state.buildParseError("`language_group_id` is required in markdown front matter for match pattern", mapping))
			continue
		}
		if matter.Link == "" {
			state.errorSet.Add(state.buildParseError("`link` is required in markdown front matter for match pattern", mapping))
			continue
		}
		if matter.Title == "" {
			state.errorSet.Add(state.buildParseError("`title` is required in markdown front matter for match pattern", mapping))
			continue
		}

		matterByPath[m] = matter
		entriesByGroup[groupID] = append(entriesByGroup[groupID], m)
	}

	pages := make([]ConfigPageV2, 0, len(entriesByGroup))
	for _, entries := range entriesByGroup {
		page := buildMatchGroupV2(state, mapping, entries, matterByPath)
		if page.Type == "" {
			continue
		}
		pages = append(pages, page)
	}

	if err := sortPageSliceV2(sortKey, sortOrder, pages, state.config.Project.DefaultLanguage); err != nil {
		state.errorSet.Add(state.buildParseError(err.Error(), mapping))
		return nil
	}
	return pages
}

func buildMatchGroupV2(state *ParseStateV2, mapping *ast.MappingNode, entries []string, matterByPath map[string]*FrontMatter) ConfigPageV2 {
	if len(entries) == 0 {
		state.errorSet.Add(state.buildParseError("internal error: match group is empty", mapping))
		return ConfigPageV2{}
	}
	langMap := make(map[string]ConfigPageLangV2)
	for _, path := range entries {
		matter, ok := matterByPath[path]
		if !ok || matter == nil {
			state.errorSet.Add(state.buildParseError("internal error: missing front matter for matched file", mapping))
			return ConfigPageV2{}
		}
		lang := strings.ToLower(matter.Lang())
		if lang == "" {
			lang = state.config.Project.DefaultLanguage
		}
		if !isValidISOLanguageCode(lang) {
			message := fmt.Sprintf("`lang` must be a valid ISO 639-1 language code. given: %s", lang)
			state.errorSet.Add(state.buildParseError(message, mapping))
			return ConfigPageV2{}
		}
		if _, ok := langMap[lang]; ok {
			state.errorSet.Add(state.buildParseError("duplicate `lang` detected for link", mapping))
			return ConfigPageV2{}
		}
		link := matter.Link
		if link == "" {
			state.errorSet.Add(state.buildParseError("`link` is required in markdown front matter for match pattern", mapping))
			return ConfigPageV2{}
		}
		if matter.Title == "" {
			state.errorSet.Add(state.buildParseError("`title` is required in markdown front matter for match pattern", mapping))
			return ConfigPageV2{}
		}
		langMap[lang] = ConfigPageLangV2{
			Link:     link,
			Title:    matter.Title,
			Filepath: path,
		}
	}

	return ConfigPageV2{
		Type: ConfigPageTypeMarkdownMultiLanguageV2,
		Lang: langMap,
	}
}

func sortPageSliceV2(sortKey, sortOrder string, pages []ConfigPageV2, defaultLang string) error { //nolint: cyclop
	if sortKey == "" && sortOrder == "" {
		return nil
	}
	if sortKey == "" {
		return errors.New("sort key is not provided")
	}

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
			left := pageTitleForSortV2(&pages[i], defaultLang)
			right := pageTitleForSortV2(&pages[j], defaultLang)
			return (left < right) == isASC
		})
		return nil
	}
	return fmt.Errorf("invalid sort key: %s", sortKey)
}

func pageTitleForSortV2(page *ConfigPageV2, defaultLang string) string {
	if len(page.Lang) == 0 {
		return ""
	}
	if defaultLang != "" {
		if entry, ok := page.Lang[defaultLang]; ok {
			return entry.Title
		}
	}

	keys := make([]string, 0, len(page.Lang))
	for key := range page.Lang {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return page.Lang[keys[0]].Title
}

func parseConfigPageDirectoryV2(state *ParseStateV2, mapping *ast.MappingNode) ConfigPageV2 { //nolint: cyclop
	configPage := ConfigPageV2{
		Type: ConfigPageTypeDirectoryV2,
	}
	var langNode *ast.MappingNode
	var title string

	for _, item := range mapping.Values {
		key := item.Key.String()
		switch key {
		case ConfigPageV2KeyType:
			continue
		case ConfigPageV2KeyTitle:
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`title` field must be a string", item.Value))
				continue
			}
			title = v.Value
		case ConfigPageV2KeyLang:
			v, ok := item.Value.(*ast.MappingNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`lang` field must be a mapping", item.Value))
				continue
			}
			langNode = v
		case ConfigPageV2KeyChildren:
			v, ok := item.Value.(*ast.SequenceNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`children` field must be a sequence", item.Value))
				continue
			}
			configPage.Children = parseConfigPageSequenceV2(state, v)
		default:
			state.errorSet.Add(state.buildParseError("a directory style page cannot accept the key", item))
		}
	}

	if langNode != nil {
		if title != "" {
			state.errorSet.Add(state.buildParseError("`lang` cannot be used with `title`", mapping))
		}
		configPage.Lang = parseDirectoryLangMapV2(state, langNode)
		validateLangMapV2(state, configPage.Lang, mapping)
	}

	if title != "" && len(configPage.Lang) == 0 {
		defaultLang := state.config.Project.DefaultLanguage
		if defaultLang == "" {
			defaultLang = "en"
		}
		configPage.Lang = map[string]ConfigPageLangV2{
			defaultLang: {
				Title: title,
			},
		}
	}

	if len(configPage.Lang) == 0 {
		state.errorSet.Add(state.buildParseError("`title` or `lang` is required for directory", mapping))
	}
	if len(configPage.Children) == 0 {
		state.errorSet.Add(state.buildParseError("`children` is required for directory", mapping))
	}
	return configPage
}

func parseDirectoryLangMapV2(state *ParseStateV2, langNode *ast.MappingNode) map[string]ConfigPageLangV2 { //nolint: cyclop
	langMap := make(map[string]ConfigPageLangV2)
	for _, item := range langNode.Values {
		key := strings.ToLower(item.Key.String())
		valueNode, ok := item.Value.(*ast.MappingNode)
		if !ok {
			state.errorSet.Add(state.buildParseError("each language entry must be a mapping", item.Value))
			continue
		}
		langConfig := ConfigPageLangV2{}
		for _, child := range valueNode.Values {
			childKey := child.Key.String()
			switch childKey {
			case ConfigPageV2KeyTitle:
				v, ok := child.Value.(*ast.StringNode)
				if !ok {
					state.errorSet.Add(state.buildParseError("`title` field must be a string", child.Value))
					continue
				}
				langConfig.Title = v.Value
			default:
				state.errorSet.Add(state.buildParseError("a directory language entry cannot accept the key: "+childKey, child))
			}
		}
		if langConfig.Title == "" {
			state.errorSet.Add(state.buildParseError("the `title` field is required", valueNode))
		}
		langMap[key] = langConfig
	}
	return langMap
}

func parseConfigPageSectionV2(state *ParseStateV2, mapping *ast.MappingNode) ConfigPageV2 { //nolint: cyclop, funlen
	configPage := ConfigPageV2{
		Type: ConfigPageTypeSectionV2,
	}
	var langNode *ast.MappingNode
	var title string
	var filepath string

	for _, item := range mapping.Values {
		key := item.Key.String()
		switch key {
		case ConfigPageV2KeyType:
			continue
		case "path":
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`path` field must be a string", item.Value))
				continue
			}
			configPage.Path = v.Value
		case ConfigPageV2KeyTitle:
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`title` field must be a string", item.Value))
				continue
			}
			title = v.Value
		case ConfigPageV2KeyFilepath:
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`filepath` field must be a string", item.Value))
				continue
			}
			filepath = v.Value
		case ConfigPageV2KeyLang:
			v, ok := item.Value.(*ast.MappingNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`lang` field must be a mapping", item.Value))
				continue
			}
			langNode = v
		case ConfigPageV2KeyChildren:
			v, ok := item.Value.(*ast.SequenceNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`children` field must be a sequence", item.Value))
				continue
			}
			configPage.Children = parseConfigPageSequenceV2(state, v)
		default:
			state.errorSet.Add(state.buildParseError("a section style page cannot accept the key: "+key, item))
		}
	}

	if configPage.Path == "" {
		state.errorSet.Add(state.buildParseError("the `path` field is required for section", mapping))
	}

	if langNode != nil {
		if title != "" || filepath != "" {
			state.errorSet.Add(state.buildParseError("`lang` cannot be used with `title` or `filepath`", mapping))
		}
		configPage.Lang = parseSectionLangMapV2(state, langNode)
		validateLangMapV2(state, configPage.Lang, mapping)
		return configPage
	}

	defaultLang := state.config.Project.DefaultLanguage
	if defaultLang == "" {
		defaultLang = "en"
	}
	singleLang := ConfigPageLangV2{
		Title:    title,
		Filepath: filepath,
	}
	fillSectionLangFromMarkdownV2(state, &singleLang, mapping)
	if singleLang.Title == "" {
		state.errorSet.Add(state.buildParseError("the `title` field is required", mapping))
	}
	if singleLang.Filepath == "" {
		state.errorSet.Add(state.buildParseError("the `filepath` field is required", mapping))
	}
	configPage.Lang = map[string]ConfigPageLangV2{
		defaultLang: singleLang,
	}
	return configPage
}

func parseSectionLangMapV2(state *ParseStateV2, langNode *ast.MappingNode) map[string]ConfigPageLangV2 { //nolint: cyclop
	langMap := make(map[string]ConfigPageLangV2)
	for _, item := range langNode.Values {
		key := strings.ToLower(item.Key.String())
		valueNode, ok := item.Value.(*ast.MappingNode)
		if !ok {
			state.errorSet.Add(state.buildParseError("each language entry must be a mapping", item.Value))
			continue
		}
		langConfig := ConfigPageLangV2{}
		for _, child := range valueNode.Values {
			childKey := child.Key.String()
			switch childKey {
			case ConfigPageV2KeyTitle:
				v, ok := child.Value.(*ast.StringNode)
				if !ok {
					state.errorSet.Add(state.buildParseError("`title` field must be a string", child.Value))
					continue
				}
				langConfig.Title = v.Value
			case ConfigPageV2KeyFilepath, "path":
				v, ok := child.Value.(*ast.StringNode)
				if !ok {
					state.errorSet.Add(state.buildParseError("`filepath` field must be a string", child.Value))
					continue
				}
				langConfig.Filepath = v.Value
			default:
				state.errorSet.Add(state.buildParseError("a section language entry cannot accept the key: "+childKey, child))
			}
		}

		fillSectionLangFromMarkdownV2(state, &langConfig, valueNode)
		if langConfig.Title == "" {
			state.errorSet.Add(state.buildParseError("the `title` field is required", valueNode))
		}
		if langConfig.Filepath == "" {
			state.errorSet.Add(state.buildParseError("the `filepath` field is required", valueNode))
		}
		langMap[key] = langConfig
	}
	return langMap
}

func fillSectionLangFromMarkdownV2(state *ParseStateV2, langConfig *ConfigPageLangV2, mapping *ast.MappingNode) {
	if langConfig.Filepath == "" {
		state.errorSet.Add(state.buildParseError("the `filepath` field is required", mapping))
		return
	}

	clean, err := state.getSecurePath(langConfig.Filepath)
	if err != nil {
		state.errorSet.Add(state.buildParseError(err.Error(), mapping))
		return
	}

	p, err := NewFrontMatterFromMarkdown(clean)
	if err != nil {
		message := fmt.Sprintf("cannot read the markdown file: %s, %v", langConfig.Filepath, err.Error())
		state.errorSet.Add(state.buildParseError(message, mapping))
		return
	}

	if langConfig.Title == "" && p.Title != "" {
		langConfig.Title = p.Title
	}
}

func parseConfigAssetsV2(state *ParseStateV2, node *ast.MappingValueNode) {
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

	assets := make([]ConfigAssetV2, 0, len(sequence.Values))
	for _, item := range sequence.Values {
		v, ok := item.(*ast.StringNode)
		if !ok {
			state.errorSet.Add(state.buildParseError("an item in the `sequence` field must have a string type", item))
			continue
		}
		assets = append(assets, ConfigAssetV2(v.Value))
	}
	state.config.Assets = assets
}

func parseConfigAnnotationV2(state *ParseStateV2, node *ast.MappingValueNode) {
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
