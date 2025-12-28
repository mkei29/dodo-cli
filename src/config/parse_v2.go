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
	"github.com/toritoritori29/dodo-cli/src/utils"
)

const (
	ConfigPageTypeMarkdownV2               = "markdown"
	ConfigPageTypeMarkdownMultiLanguageV2  = "markdown_multilanguage"
	ConfigPageTypeMatchV2                  = "match"
	ConfigPageTypeDirectoryV2              = "directory"
	ConfigPageTypeDirectoryMultiLanguageV2 = "directory_multilanguage"
	ConfigPageTypeSectionV2                = "section"
	ConfigPageTypeSectionV2MultiLanguage   = "section_multilanguage"
)

const (
	ConfigPageV2KeyType     = "type"
	ConfigPageV2KeyLink     = "link"
	ConfigPageV2KeyTitle    = "title"
	ConfigPageV2KeyDesc     = "description"
	ConfigPageV2KeyFilepath = "filepath"
	ConfigPageV2KeyLang     = "lang"
	ConfigPageV2KeyPattern  = "pattern"
	ConfigPageV2KeySortKey  = "sort_key"
	ConfigPageV2KeySortOrd  = "sort_order"
	ConfigPageV2KeyChildren = "children"
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

	// LangPage holds per-locale page variants (keyed by ISO 639-1 code).
	LangPage map[string]ConfigPageLangPage
	// LangDirectory holds per-locale directory titles (keyed by ISO 639-1 code).
	LangDirectory map[string]ConfigPageLangDirectory
	// LangSection holds per-locale section titles (keyed by ISO 639-1 code).
	LangSection map[string]ConfigPageLangSection

	// Children is used by directory/section entries.
	Children []ConfigPageV2
}

func (c *ConfigPageV2) SortKeyTitle(defaultLang string) string {
	if len(c.LangPage) == 0 {
		return "<invalid page>" // invalid state
	}
	if entry, ok := c.LangPage[defaultLang]; ok {
		return entry.Title
	}
	return "<invalid page>" // invalid state
}

type ConfigPageLangPage struct {
	Link        string
	Title       string
	Description string
	Filepath    string
}

type ConfigPageLangDirectory struct {
	Title       string
	Description string
}

type ConfigPageLangSection struct {
	Title       string
	Description string
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
		case ConfigPageTypeDirectoryMultiLanguageV2:
			p := parseConfigPageDirectoryMultiLanguageV2(state, pageNode)
			configPages = append(configPages, p)
		case ConfigPageTypeSectionV2:
			p := parseConfigPageSectionV2(state, pageNode)
			configPages = append(configPages, p)
		case ConfigPageTypeSectionV2MultiLanguage:
			p := parseConfigPageSectionMultiLanguageV2(state, pageNode)
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
		case ConfigPageTypeDirectoryV2:
			if hasLang {
				return ConfigPageTypeDirectoryMultiLanguageV2
			}
			return ConfigPageTypeDirectoryV2
		case ConfigPageTypeSectionV2:
			if hasLang {
				return ConfigPageTypeSectionV2MultiLanguage
			}
			return ConfigPageTypeSectionV2
		case ConfigPageTypeMatchV2:
			return strings.ToLower(v.Value)
		default:
			state.errorSet.Add(state.buildParseError("unknown page type: "+v.Value, item.Value))
			return ""
		}
	}

	state.errorSet.Add(state.buildParseError("this mapping does not match any page type", mapping))
	return ""
}

// Parse Markdown Page --------------------------------------------------------
func parseConfigPageMarkdownV2(state *ParseStateV2, mapping *ast.MappingNode) ConfigPageV2 {
	configPage := ConfigPageV2{
		Type: ConfigPageTypeMarkdownV2,
	}
	langItem := ConfigPageLangPage{}

	for _, item := range mapping.Values {
		key := item.Key.String()
		switch key {
		case ConfigPageV2KeyType:
			continue
		case ConfigPageV2KeyLink:
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`link` field must be a string", item.Value))
				continue
			}
			langItem.Link = v.Value
		case ConfigPageV2KeyTitle:
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`title` field must be a string", item.Value))
				continue
			}
			langItem.Title = v.Value
		case ConfigPageV2KeyDesc:
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`description` field must be a string", item.Value))
				continue
			}
			langItem.Description = v.Value
		case ConfigPageV2KeyFilepath:
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`filepath` field must be a string", item.Value))
				continue
			}
			langItem.Filepath = v.Value
		default:
			state.errorSet.Add(state.buildParseError("a markdown style page cannot accept the key: "+key, item))
		}
	}

	fillSingleLangFromMarkdownV2(state, &langItem, mapping)
	defaultLang := state.config.Project.DefaultLanguage
	if defaultLang == "" {
		defaultLang = "en"
	}
	configPage.LangPage = map[string]ConfigPageLangPage{
		defaultLang: langItem,
	}
	validateConfigPageMarkdown(state, &configPage, mapping)
	return configPage
}

func parseConfigPageMarkdownMultiLanguageV2(state *ParseStateV2, mapping *ast.MappingNode) ConfigPageV2 {
	configPage := ConfigPageV2{
		Type: ConfigPageTypeMarkdownMultiLanguageV2,
	}

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
			configPage.LangPage = parseMarkdownLangEntriesV2(state, v)
		case ConfigPageV2KeyLink, ConfigPageV2KeyTitle, ConfigPageV2KeyDesc, ConfigPageV2KeyFilepath:
			state.errorSet.Add(state.buildParseError("single-locale fields cannot be used with `lang`", item.Value))
		default:
			state.errorSet.Add(state.buildParseError("a markdown style page cannot accept the key: "+key, item))
		}
	}
	validateConfigPageMarkdown(state, &configPage, mapping)
	return configPage
}

func parseMarkdownLangEntriesV2(state *ParseStateV2, langNode *ast.MappingNode) map[string]ConfigPageLangPage {
	langMap := make(map[string]ConfigPageLangPage)
	for _, item := range langNode.Values {
		key := strings.ToLower(item.Key.String())
		valueNode, ok := item.Value.(*ast.MappingNode)
		if !ok {
			state.errorSet.Add(state.buildParseError("each language entry must be a mapping", item.Value))
			continue
		}
		langConfig := ConfigPageLangPage{}
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
			case ConfigPageV2KeyDesc:
				v, ok := child.Value.(*ast.StringNode)
				if !ok {
					state.errorSet.Add(state.buildParseError("`description` field must be a string", child.Value))
					continue
				}
				langConfig.Description = v.Value
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
		fillSingleLangFromMarkdownV2(state, &langConfig, valueNode)
		langMap[key] = langConfig
	}
	return langMap
}

func fillSingleLangFromMarkdownV2(state *ParseStateV2, langPage *ConfigPageLangPage, mapping *ast.MappingNode) {
	if langPage.Filepath == "" {
		state.errorSet.Add(state.buildParseError("the `filepath` field is required", mapping))
		return
	}

	clean, err := state.getSecurePath(langPage.Filepath)
	if err != nil {
		state.errorSet.Add(state.buildParseError(err.Error(), mapping))
		return
	}

	p, err := NewFrontMatterFromMarkdown(clean)
	if err != nil {
		message := fmt.Sprintf("cannot read the markdown file: %s, %v", langPage.Filepath, err.Error())
		state.errorSet.Add(state.buildParseError(message, mapping))
		return
	}

	// Fill missing fields from the markdown front matter.
	if langPage.Title == "" && p.Title != "" {
		langPage.Title = p.Title
	}
	if langPage.Link == "" && p.Link != "" {
		langPage.Link = p.Link
	}
	if langPage.Description == "" && p.Description != "" {
		langPage.Description = p.Description
	}
}

func validateConfigPageMarkdown(state *ParseStateV2, langConfig *ConfigPageV2, mapping *ast.MappingNode) {

	for _, langItem := range langConfig.LangPage {
		if langItem.Title == "" {
			state.errorSet.Add(state.buildParseError("the `title` field is required", mapping))
		}
		if langItem.Link == "" {
			state.errorSet.Add(state.buildParseError("the `link` field is required", mapping))
		}
		if langItem.Filepath == "" {
			state.errorSet.Add(state.buildParseError("the `filepath` field is required", mapping))
		}
	}
}

func validateLangKeySetV2(state *ParseStateV2, mapping *ast.MappingNode, keySet map[string]struct{}) {
	for code := range keySet {
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
	if _, ok := keySet[defaultLang]; !ok {
		state.errorSet.Add(state.buildParseError("`lang` must include the default language: "+defaultLang, mapping))
	}
}

// Parse Match Page -----------------------------------------------------------
func parseConfigPageMatchV2(state *ParseStateV2, mapping *ast.MappingNode) []ConfigPageV2 {
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
		default:
			state.errorSet.Add(state.buildParseError("a match style page cannot accept the key", item))
		}
	}

	if sortKey == "" && sortOrder != "" {
		state.errorSet.Add(state.buildParseError("`sort_key` must not be empty if you specify `sort_order`", mapping))
		return nil
	}
	pages := buildConfigPageFromMatchStatementV2(state, mapping, pattern, sortKey, sortOrder)

	// Run validation
	for _, page := range pages {
		validateConfigPageMarkdown(state, &page, mapping)
	}
	return pages
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

	pagesByGroupID := make(map[string]ConfigPageV2)
	for _, m := range matches {
		matter, err := NewFrontMatterFromMarkdown(m)
		if err != nil {
			message := fmt.Sprintf("%s: %s", err.Error(), m)
			state.errorSet.Add(state.buildParseError(message, mapping))
			continue
		}
		langItem := ConfigPageLangPage{}
		fillSingleLangFromMarkdownV2(state, &langItem, mapping)

		lang := getLanguageFromFrontmatterV2(matter, state.config.Project.DefaultLanguage)
		if !isValidISOLanguageCode(lang) {
			message := fmt.Sprintf("`lang` must be a valid ISO 639-1 language code. given: %s", lang)
			state.errorSet.Add(state.buildParseError(message, mapping))
			continue
		}

		page, ok := pagesByGroupID[matter.LanguageGroupID]
		// If not exists, create a new page entry.
		if !ok {
			pagesByGroupID[matter.LanguageGroupID] = ConfigPageV2{
				Type: ConfigPageTypeMarkdownMultiLanguageV2,
				LangPage: map[string]ConfigPageLangPage{
					lang: langItem,
				},
			}
		}
		// If exists, just add the lang entry.
		if _, exists := page.LangPage[lang]; exists {
			message := fmt.Sprintf("duplicate `lang` detected for link in match pattern: %s", lang)
			state.errorSet.Add(state.buildParseError(message, mapping))
			continue
		}
		page.LangPage[lang] = langItem
	}

	pages := utils.Values(pagesByGroupID)
	if err := sortPageSliceV2(sortKey, sortOrder, pages, state.config.Project.DefaultLanguage); err != nil {
		state.errorSet.Add(state.buildParseError(err.Error(), mapping))
		return nil
	}
	return pages
}

func getLanguageFromFrontmatterV2(matter *FrontMatter, defaultLang string) string {
	lang := strings.ToLower(matter.Lang())
	if lang == "" {
		lang = defaultLang
	}
	return lang
}

func sortPageSliceV2(sortKey, sortOrder string, pages []ConfigPageV2, defaultLang string) error {
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
			left := pages[i].SortKeyTitle(defaultLang)
			right := pages[j].SortKeyTitle(defaultLang)
			return (left < right) == isASC
		})
		return nil
	}
	return fmt.Errorf("invalid sort key: %s", sortKey)
}

// Parse Directory Page -------------------------------------------------------
func parseConfigPageDirectoryV2(state *ParseStateV2, mapping *ast.MappingNode) ConfigPageV2 {
	configPage := ConfigPageV2{
		Type: ConfigPageTypeDirectoryV2,
	}
	langItem := &ConfigPageLangDirectory{}

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
			langItem.Title = v.Value
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

	defaultLang := state.config.Project.DefaultLanguage
	if defaultLang == "" {
		defaultLang = "en"
	}
	configPage.LangDirectory = map[string]ConfigPageLangDirectory{
		defaultLang: *langItem,
	}
	validateConfigPageDirectory(state, configPage, mapping)
	return configPage
}

func parseConfigPageDirectoryMultiLanguageV2(state *ParseStateV2, mapping *ast.MappingNode) ConfigPageV2 {
	configPage := ConfigPageV2{
		Type: ConfigPageTypeDirectoryMultiLanguageV2,
	}
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
			configPage.LangDirectory = parseDirectoryLangEntriesV2(state, v)
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
	validateConfigPageDirectory(state, configPage, mapping)
	return configPage
}

func parseDirectoryLangEntriesV2(state *ParseStateV2, mapping *ast.MappingNode) map[string]ConfigPageLangDirectory {
	langMap := make(map[string]ConfigPageLangDirectory)

	for _, item := range mapping.Values {
		lang := strings.ToLower(item.Key.String())
		entry, ok := item.Value.(*ast.MappingNode)
		if !ok {
			state.errorSet.Add(state.buildParseError("each language entry must be a mapping", item.Value))
			continue
		}

		item := &ConfigPageLangDirectory{}
		for _, field := range entry.Values {
			key := field.Key.String()
			switch key {
			case ConfigPageV2KeyTitle:
				v, ok := field.Value.(*ast.StringNode)
				if !ok {
					state.errorSet.Add(state.buildParseError("`title` field must be a string", field.Value))
					continue
				}
				item.Title = v.Value
			case ConfigPageV2KeyDesc:
				v, ok := field.Value.(*ast.StringNode)
				if !ok {
					state.errorSet.Add(state.buildParseError("`description` field must be a string", field.Value))
					continue
				}
				item.Description = v.Value
			default:
				state.errorSet.Add(state.buildParseError("a directory language entry cannot accept the key: "+key, field))
			}
		}
		langMap[lang] = *item
	}
	return langMap
}

func validateConfigPageDirectory(state *ParseStateV2, page ConfigPageV2, mapping *ast.MappingNode) {
	if len(page.LangSection) == 0 {
		state.errorSet.Add(state.buildParseError("`lang` must not be empty", mapping))
		return
	}

	// Check language keys follow ISO 639-1 codes
	keySet := make(map[string]struct{}, len(page.LangSection))
	for key := range page.LangSection {
		keySet[key] = struct{}{}
	}
	validateLangKeySetV2(state, mapping, keySet)

	// Check each language entry
	for lang, entry := range page.LangSection {
		if entry.Title == "" {
			message := fmt.Sprintf("the `title` field is required for language: %s", lang)
			state.errorSet.Add(state.buildParseError(message, mapping))
		}
		if len(page.Children) == 0 {
			message := fmt.Sprintf("`children` is required for directory for language: %s", lang)
			state.errorSet.Add(state.buildParseError(message, mapping))
		}
	}
}

// Parse Section Page -------------------------------------------------------
func parseConfigPageSectionV2(state *ParseStateV2, mapping *ast.MappingNode) ConfigPageV2 {
	configPage := ConfigPageV2{
		Type: ConfigPageTypeSectionV2,
	}
	langItem := &ConfigPageLangSection{}

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
			langItem.Title = v.Value
		case ConfigPageV2KeyDesc:
			v, ok := item.Value.(*ast.StringNode)
			if !ok {
				state.errorSet.Add(state.buildParseError("`description` field must be a string", item.Value))
				continue
			}
			langItem.Description = v.Value
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

	defaultLang := state.config.Project.DefaultLanguage
	if defaultLang == "" {
		defaultLang = "en"
	}
	configPage.LangSection = map[string]ConfigPageLangSection{
		defaultLang: *langItem,
	}
	validateConfigPageSection(state, configPage, mapping)
	return configPage
}

func parseConfigPageSectionMultiLanguageV2(state *ParseStateV2, mapping *ast.MappingNode) ConfigPageV2 {
	configPage := ConfigPageV2{
		Type: ConfigPageTypeSectionV2,
	}

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
			configPage.LangSection = parseSectionLangEntriesV2(state, v)
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
	validateConfigPageSection(state, configPage, mapping)
	return configPage
}

func parseSectionLangEntriesV2(state *ParseStateV2, mapping *ast.MappingNode) map[string]ConfigPageLangSection {
	langMap := make(map[string]ConfigPageLangSection)

	for _, item := range mapping.Values {
		lang := strings.ToLower(item.Key.String())
		entry, ok := item.Value.(*ast.MappingNode)
		if !ok {
			state.errorSet.Add(state.buildParseError("each language entry must be a mapping", item.Value))
			continue
		}

		item := &ConfigPageLangSection{}
		for _, field := range entry.Values {
			key := field.Key.String()
			switch key {
			case ConfigPageV2KeyTitle:
				v, ok := field.Value.(*ast.StringNode)
				if !ok {
					state.errorSet.Add(state.buildParseError("`title` field must be a string", field.Value))
					continue
				}
				item.Title = v.Value
			case ConfigPageV2KeyDesc:
				v, ok := field.Value.(*ast.StringNode)
				if !ok {
					state.errorSet.Add(state.buildParseError("`description` field must be a string", field.Value))
					continue
				}
				item.Description = v.Value
			default:
				state.errorSet.Add(state.buildParseError("a section language entry cannot accept the key: "+key, field))
			}
		}
		langMap[lang] = *item
	}
	return langMap
}

func validateConfigPageSection(state *ParseStateV2, page ConfigPageV2, mapping *ast.MappingNode) {
	if len(page.LangSection) == 0 {
		state.errorSet.Add(state.buildParseError("`lang` must not be empty", mapping))
		return
	}

	// Check language keys follow ISO 639-1 codes
	keySet := make(map[string]struct{}, len(page.LangSection))
	for key := range page.LangSection {
		keySet[key] = struct{}{}
	}
	validateLangKeySetV2(state, mapping, keySet)

	// Check each language entry
	for lang, entry := range page.LangSection {
		if entry.Title == "" {
			message := fmt.Sprintf("the `title` field is required for language: %s", lang)
			state.errorSet.Add(state.buildParseError(message, mapping))
		}
	}
}

// Other Sections --------------------------------------------------------------
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
