package main

import (
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/caarlos0/log"
	"github.com/toritoritori29/dodo-cli/src/config"
	appErrors "github.com/toritoritori29/dodo-cli/src/errors"
)

const (
	PageTypeRootNode        = "RootNode"
	PageTypeLeafNode        = "LeafNode"
	PageTypeDirNode         = "DirNodeWithoutPage"
	PageTypeDirNodeWithPage = "DirNodeWithPage"
	PageTypeSectionNode     = "SectionNode"
)

type PageLanguageWiseInfo struct {
	Language    string `json:"language"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Path        string `json:"path"`
	Hash        string `json:"hash"`
	Filepath    string `json:"filepath"`
}

type PageSummary struct {
	Type        string                  `json:"type"`
	Filepath    string                  `json:"filepath"`
	Hash        string                  `json:"hash"`
	Path        string                  `json:"path"`
	Title       string                  `json:"title"`
	Description string                  `json:"description"`
	Language    []PageLanguageWiseInfo  `json:"language"`
	UpdatedAt   config.SerializableTime `json:"updated_at"`
}

func NewPageSummary(filepath, path, title string) PageSummary {
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(filepath)))
	return PageSummary{
		Filepath: filepath,
		Path:     path,
		Hash:     hash,
		Title:    title,
	}
}

func NewPageHeaderFromPage(p *Page) PageSummary {
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(p.Filepath)))
	return PageSummary{
		Type:        p.Type,
		Filepath:    p.Filepath,
		Hash:        hash,
		Path:        p.Path,
		Title:       p.Title,
		Description: p.Description,
		Language:    p.Language,
	}
}

type Page struct {
	Type        string                  `json:"type"`
	Filepath    string                  `json:"filepath"`
	Hash        string                  `json:"hash"`
	Path        string                  `json:"path"`
	Title       string                  `json:"title"`
	Description string                  `json:"description"`
	Language    []PageLanguageWiseInfo  `json:"language"`
	UpdatedAt   config.SerializableTime `json:"updated_at"`
	Children    []Page                  `json:"children"`
}

func NewLeafNodeFromConfigPage(configProject *config.ConfigProjectV1, configPage *config.ConfigPageV1) Page {
	page := Page{
		Type: PageTypeLeafNode,
		Language: []PageLanguageWiseInfo{
			{
				Language:    configProject.DefaultLanguage,
				Title:       configPage.Title,
				Description: configPage.Description,
				Path:        configPage.Path,
				Filepath:    configPage.Markdown,
				Hash:        fmt.Sprintf("%x", sha256.Sum256([]byte(configPage.Markdown))),
			},
		},
		UpdatedAt: configPage.UpdatedAt,
		Children:  []Page{},
	}
	return page
}

// List PageSummary of the page includes.
func (p *Page) ListPageHeader() []PageSummary {
	list := make([]PageSummary, 0, len(p.Children))
	list = listPageHeader(list, p)
	return list
}

func listPageHeader(list []PageSummary, p *Page) []PageSummary {
	list = append(list, NewPageHeaderFromPage(p))
	for idx := range p.Children {
		list = listPageHeader(list, &p.Children[idx])
	}
	return list
}

// Check if the page is valid.
// This function checks the following conditions:
// 1. All pages have necessary fields.
// 2. There are no duplicated paths.
func (p *Page) IsValid(defaultLang string) *appErrors.MultiError {
	errorSet := appErrors.NewMultiError()
	p.isValid(true, &errorSet)
	if errorSet.HasError() {
		return &errorSet
	}

	// Check if there are duplicated paths.
	pathMap := make(map[string]int)
	p.duplicationCount(pathMap)
	for path, value := range pathMap {
		if value > 1 {
			errorSet.Add(appErrors.NewAppError(fmt.Sprintf("duplicated path was found. path: `%s`", path)))
		}
	}

	// Check if all pages implement the default language.
	p.isImplementDefaultLanguage(defaultLang, &errorSet)

	if errorSet.HasError() {
		return &errorSet
	}
	return nil
}

func (p *Page) isValid(isRoot bool, errorSet *appErrors.MultiError) {
	if isRoot && p.Type != PageTypeRootNode {
		errorSet.Add(appErrors.NewAppError("Type for root node should be RootNode"))
		return
	}
	if !isRoot && p.Type == PageTypeRootNode {
		errorSet.Add(appErrors.NewAppError("Type for non-root node should not be RootNode"))
		return
	}
	if p.Type == PageTypeLeafNode {
		for _, langInfo := range p.Language {
			if langInfo.Path == "" {
				errorSet.Add(appErrors.NewAppError("'path' is required for leaf node"))
			}
		}
	}
	for _, c := range p.Children {
		c.isValid(false, errorSet)
	}
}

func (p *Page) isImplementDefaultLanguage(lang string, errorSet *appErrors.MultiError) {
	for i := range p.Children {
		p.Children[i].isImplementDefaultLanguage(lang, errorSet)
	}
	if p.Type == PageTypeRootNode {
		return
	}

	var otherLanguage PageSummary
	for _, l := range p.Language {
		if l.Language == lang {
			return
		}
		otherLanguage = NewPageHeaderFromPage(p)
	}
	errorSet.Add(appErrors.NewAppError(fmt.Sprintf("there is no default language page corresponding to: %+v", otherLanguage.Title)))
}

func (p *Page) duplicationCount(pathMap map[string]int) {
	if p.Type == PageTypeLeafNode {
		for _, langInfo := range p.Language {
			if langInfo.Path != "" {
				pathMap[langInfo.Path]++
			}
		}
	}
	for _, c := range p.Children {
		c.duplicationCount(pathMap)
	}
}

// Generate a string representation of the page.
func (p *Page) String() string {
	return p.buildString(0)
}

func (p *Page) buildString(depth int) string {
	offset := strings.Repeat("-", depth*2)
	lines := make([]string, 0, len(p.Children)+1)
	lines = append(lines, fmt.Sprintf("%sTitle: %s, Path: %s", offset, p.Title, p.Path))
	for _, c := range p.Children {
		lines = append(lines, c.buildString(depth+1))
	}
	return strings.Join(lines, "\n")
}

// Count the number of pages.
func (p *Page) Count() int {
	return p.buildCount() - 1
}

func (p *Page) buildCount() int {
	c := 1
	for _, child := range p.Children {
		c += child.buildCount()
	}
	return c
}

// Add a child page.
func (p *Page) Add(page Page) {
	p.Children = append(p.Children, page)
}

func CreatePageTree(conf *config.ConfigV1, rootDir string) (*Page, *appErrors.MultiError) {
	errorSet := appErrors.NewMultiError()
	root := Page{
		Type: PageTypeRootNode,
	}

	children := make([]Page, 0, len(conf.Pages))
	for _, p := range conf.Pages {
		c, merr := buildPage(rootDir, &conf.Project, &p)
		if merr != nil {
			errorSet.Merge(*merr)
			continue
		}
		children = append(children, c...)
	}
	if errorSet.HasError() {
		return nil, &errorSet
	}

	root.Children = children
	return &root, nil
}

func buildPage(rootDir string, configProject *config.ConfigProjectV1, configPage *config.ConfigPageV1) ([]Page, *appErrors.MultiError) {
	if configPage.MatchMarkdown() {
		return transformMarkdown(rootDir, configProject, configPage)
	}

	if configPage.MatchDirectory() {
		return transformDirectory(rootDir, configProject, configPage)
	}

	err := appErrors.NewMultiError()
	err.Add(appErrors.NewAppError("invalid configuration: doesn't match any pattern"))
	return nil, &err
}

func transformMarkdown(rootDir string, configProject *config.ConfigProjectV1, configPage *config.ConfigPageV1) ([]Page, *appErrors.MultiError) {
	merr := appErrors.NewMultiError()
	filepath := filepath.Clean(filepath.Join(rootDir, configPage.Markdown))

	if err := config.IsUnderRootPath(rootDir, filepath); err != nil {
		merr.Add(fmt.Errorf("path should be under the rootDir. passed: %s", filepath))
		return nil, &merr
	}

	// First, populate the fields from the markdown front matter.
	p := NewLeafNodeFromConfigPage(configProject, configPage)
	log.Debugf("Node Found. Type: Markdown, Filepath: '%s', Title: '%s', Path: '%s'", p.Filepath, p.Title, p.Path)
	return []Page{p}, nil
}

func transformDirectory(rootDir string, configProject *config.ConfigProjectV1, configPage *config.ConfigPageV1) ([]Page, *appErrors.MultiError) {
	merr := appErrors.NewMultiError()

	children := make([]Page, 0, len(configPage.Children))
	for _, child := range configPage.Children {
		pages, err := buildPage(rootDir, configProject, &child)
		if err != nil {
			merr.Merge(*err)
			continue
		}
		children = append(children, pages...)
	}

	p := Page{
		Type: PageTypeDirNode,
		Language: []PageLanguageWiseInfo{
			{
				Language:    configProject.DefaultLanguage,
				Title:       configPage.Directory,
				Description: "",
			},
		},
		Children: children,
	}
	if merr.HasError() {
		return nil, &merr
	}
	log.Debugf("Node Found. Type: Directory, Title: %s", p.Title)
	return []Page{p}, nil
}

// CreatePageTreeV2 creates a page tree from ConfigV2.
func CreatePageTreeV2(conf *config.ConfigV2, rootDir string) (*Page, *appErrors.MultiError) {
	errorSet := appErrors.NewMultiError()
	root := Page{
		Type: PageTypeRootNode,
	}

	children := make([]Page, 0, len(conf.Pages))
	for _, p := range conf.Pages {
		c, merr := buildPageV2(rootDir, &conf.Project, &p)
		if merr != nil {
			errorSet.Merge(*merr)
			continue
		}
		children = append(children, c...)
	}
	if errorSet.HasError() {
		return nil, &errorSet
	}

	root.Children = children
	return &root, nil
}

func buildPageV2(rootDir string, configProject *config.ConfigProjectV2, configPage *config.ConfigPageV2) ([]Page, *appErrors.MultiError) {
	defaultLang := configProject.GetDefaultLanguageOrFallback()

	switch configPage.Type {
	case config.ConfigPageTypeMarkdownV2, config.ConfigPageTypeMarkdownMultiLanguageV2:
		return transformMarkdownV2(rootDir, defaultLang, configPage)
	case config.ConfigPageTypeSectionV2, config.ConfigPageTypeSectionV2MultiLanguage:
		return transformSectionV2(rootDir, configProject, configPage)
	case config.ConfigPageTypeDirectoryV2, config.ConfigPageTypeDirectoryMultiLanguageV2:
		return transformDirectoryV2(rootDir, configProject, configPage)
	default:
		err := appErrors.NewMultiError()
		err.Add(appErrors.NewAppError("unknown page type: " + configPage.Type))
		return nil, &err
	}
}

func transformMarkdownV2(rootDir, defaultLang string, configPage *config.ConfigPageV2) ([]Page, *appErrors.MultiError) {
	merr := appErrors.NewMultiError()

	// Get default language page info
	langPage, ok := configPage.LangPage[defaultLang]
	if !ok {
		merr.Add(fmt.Errorf("default language (%s) not found in page", defaultLang))
		return nil, &merr
	}

	filepath := filepath.Clean(filepath.Join(rootDir, langPage.Filepath))
	if err := config.IsUnderRootPath(rootDir, filepath); err != nil {
		merr.Add(fmt.Errorf("path should be under the rootDir. passed: %s", filepath))
		return nil, &merr
	}

	// Build language info for all languages
	languageInfo := make([]PageLanguageWiseInfo, 0, len(configPage.LangPage))
	for lang, lp := range configPage.LangPage {
		languageInfo = append(languageInfo, PageLanguageWiseInfo{
			Language:    lang,
			Title:       lp.Title,
			Description: lp.Description,
			Path:        lp.Link,
			Filepath:    lp.Filepath,
			Hash:        fmt.Sprintf("%x", sha256.Sum256([]byte(lp.Filepath))),
		})
	}

	p := Page{
		Type:     PageTypeLeafNode,
		Language: languageInfo,
		Children: []Page{},
	}

	log.Debugf("Node Found. Type: Markdown, Filepath: '%s', Title: '%s', Path: '%s'", p.Filepath, p.Title, p.Path)
	return []Page{p}, nil
}

func transformSectionV2(rootDir string, configProject *config.ConfigProjectV2, configPage *config.ConfigPageV2) ([]Page, *appErrors.MultiError) {
	merr := appErrors.NewMultiError()

	children := make([]Page, 0, len(configPage.Children))
	for _, child := range configPage.Children {
		pages, err := buildPageV2(rootDir, configProject, &child)
		if err != nil {
			merr.Merge(*err)
			continue
		}
		children = append(children, pages...)
	}

	// Build language info for all languages
	languageInfo := make([]PageLanguageWiseInfo, 0, len(configPage.LangSection))
	for lang, ls := range configPage.LangSection {
		languageInfo = append(languageInfo, PageLanguageWiseInfo{
			Language:    lang,
			Title:       ls.Title,
			Description: ls.Description,
			Path:        "",
		})
	}
	p := Page{
		Type:     PageTypeSectionNode,
		Language: languageInfo,
		Children: children,
	}

	if merr.HasError() {
		return nil, &merr
	}
	log.Debugf("Node Found. Type: Section, Title: %s", p.Title)
	return []Page{p}, nil
}

func transformDirectoryV2(rootDir string, configProject *config.ConfigProjectV2, configPage *config.ConfigPageV2) ([]Page, *appErrors.MultiError) {
	merr := appErrors.NewMultiError()

	children := make([]Page, 0, len(configPage.Children))
	for _, child := range configPage.Children {
		pages, err := buildPageV2(rootDir, configProject, &child)
		if err != nil {
			merr.Merge(*err)
			continue
		}
		children = append(children, pages...)
	}

	// Build language info for all languages
	languageInfo := make([]PageLanguageWiseInfo, 0, len(configPage.LangDirectory))
	for lang, ld := range configPage.LangDirectory {
		languageInfo = append(languageInfo, PageLanguageWiseInfo{
			Language:    lang,
			Title:       ld.Title,
			Description: ld.Description,
			Path:        "",
		})
	}
	p := Page{
		Type:     PageTypeDirNode,
		Language: languageInfo,
		Children: children,
	}

	if merr.HasError() {
		return nil, &merr
	}
	log.Debugf("Node Found. Type: Directory, Title: %s", p.Title)
	return []Page{p}, nil
}
