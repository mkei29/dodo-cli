package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/adrg/frontmatter"
	"github.com/caarlos0/log"
	"github.com/mattn/go-zglob"
	"gopkg.in/yaml.v3"
)

const (
	PageTypeRootNode        = "RootNode"
	PageTypeLeafNode        = "LeafNode"
	PageTypeDirNode         = "DirNodeWithoutPage"
	PageTypeDirNodeWithPage = "DirNodeWithPage"
)

type PageSummary struct {
	Type        string `json:"type"`
	Filepath    string `json:"filepath"`
	Hash        string `json:"hash"`
	Path        string `json:"path"`
	Title       string `json:"title"`
	UpdatedAt   string `json:"updated_at"`
	Description string `json:"description"`
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
	}
}

type Page struct {
	Type        string           `json:"type"`
	Filepath    string           `json:"filepath"`
	Hash        string           `json:"hash"`
	Path        string           `json:"path"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	UpdatedAt   SerializableTime `json:"updated_at"`
	Children    []Page           `json:"children"`
}

func NewLeafNodeFromFrontMatter(filePath string) (*Page, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	formats := []*frontmatter.Format{
		frontmatter.NewFormat("---", "---", yaml.Unmarshal),
	}
	page := Page{}
	_, err = frontmatter.Parse(file, &page, formats...)
	if err != nil {
		return nil, fmt.Errorf("failed to parse front matter: %w", err)
	}

	page.Type = PageTypeLeafNode
	page.Filepath = filePath
	page.Hash = fmt.Sprintf("%x", sha256.Sum256([]byte(filePath)))
	page.Children = []Page{}
	return &page, nil
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
// 1. all page have necessary fields.
// 2. There are no duplicated paths.
func (p *Page) IsValid() ErrorSet {
	errorSet := NewErrorSet()
	p.isValid(true, &errorSet)
	if errorSet.HasError() {
		return errorSet
	}

	pathMap := make(map[string]int)
	p.duplicationCount(pathMap, "")
	for path, value := range pathMap {
		if value > 1 {
			errorSet.Add(NewAppError(fmt.Sprintf("duplicated path was found. path: `%s`", path)))
		}
	}
	return errorSet
}

func (p *Page) isValid(isRoot bool, errorSet *ErrorSet) {
	if isRoot && p.Type != PageTypeRootNode {
		errorSet.Add(NewAppError("Type for root node should be Root"))
		return
	}
	if !isRoot && p.Type == PageTypeRootNode {
		errorSet.Add(NewAppError("Type for non-root node should not be Root"))
		return
	}
	if p.Type == PageTypeLeafNode && p.Path == "" {
		errorSet.Add(NewAppError("'path' is required for leaf node"))
	}

	matched, err := regexp.MatchString("^[a-zA-Z-0-9._-]*$", p.Path)
	if err != nil || !matched {
		errorSet.Add(NewAppError(fmt.Sprintf("The path `%s` contains invalid characters. File paths can only contain alphanumeric characters, periods (.), underscores (_), and hyphens (-)", p.Filepath)))
	}
	for _, c := range p.Children {
		c.isValid(false, errorSet)
	}
}

func (p *Page) duplicationCount(pathMap map[string]int, parentPath string) {
	path := filepath.Join(parentPath, p.Path)
	if p.Path != "" {
		pathMap[path]++
	}
	for _, c := range p.Children {
		c.duplicationCount(pathMap, path)
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

// Count number of pages.
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

func SortPageSlice(sortKey, sortOrder *string, pages []Page) error {
	if sortKey == nil && sortOrder == nil {
		return nil
	}
	if sortKey == nil {
		return fmt.Errorf("sort key is not provided")
	}
	// Check sortOrder
	isASC := true
	if sortOrder != nil {
		switch strings.ToLower(*sortOrder) {
		case "asc":
			break
		case "desc":
			isASC = false
		default:
			return fmt.Errorf("invalid sort order: %s", *sortOrder)
		}
	}
	if *sortKey == "title" {
		sort.Slice(pages, func(i, j int) bool {
			return (pages[i].Title < pages[j].Title) == isASC
		})
		return nil
	}
	return fmt.Errorf("invalid sort key: %s", *sortKey)
}

func CreatePageTree(config Config, rootDir string) (*Page, ErrorSet) {
	errorSet := NewErrorSet()
	root := Page{
		Type: PageTypeRootNode,
	}

	children := make([]Page, 0, len(config.Pages))
	for _, p := range config.Pages {
		c, es := buildPage(rootDir, p)
		children = append(children, c...)
		errorSet.Merge(es)
	}

	root.Children = children
	return &root, errorSet
}

func buildPage(rootDir string, c ConfigPage) ([]Page, ErrorSet) {
	if c.MatchMarkdown() {
		return transformMarkdown(rootDir, &c)
	}
	if c.MatchMatch() {
		return transformMatch(rootDir, &c)
	}

	if c.MatchDirectory() {
		return transformDirectory(rootDir, &c)
	}

	es := NewErrorSet()
	es.Add(NewAppError("invalid configuration: doesn't match any pattern"))
	return nil, es
}

func transformMarkdown(rootDir string, c *ConfigPage) ([]Page, ErrorSet) {
	es := NewErrorSet()
	filepath := filepath.Clean(filepath.Join(rootDir, c.Markdown))

	if err := IsUnderRootPath(rootDir, filepath); err != nil {
		es.Add(NewAppError(fmt.Sprintf("path should be under the rootDir. passed: %s", filepath)))
		return nil, es
	}

	// First, populate the fields from the markdown front matter.
	p, err := NewLeafNodeFromFrontMatter(filepath)
	if err != nil {
		es.Add(err)
		return nil, es
	}

	// Second, populate the fields from the configuration.
	if c.Title != "" {
		p.Title = c.Title
	}
	if c.Path != "" {
		p.Path = c.Path
	}
	if c.Description != "" {
		p.Description = c.Description
	}
	if c.UpdatedAt != "" {
		p.UpdatedAt = c.UpdatedAt
	}

	log.Debugf("Node Found. Type: Markdown, Filepath: '%s', Title: '%s', Path: '%s'", p.Filepath, p.Title, p.Path)
	return []Page{*p}, es
}

func transformMatch(rootDir string, c *ConfigPage) ([]Page, ErrorSet) {
	pages := make([]Page, 0)
	es := NewErrorSet()
	dirPath := filepath.Clean(filepath.Join(rootDir, c.Match))
	if err := IsUnderRootPath(rootDir, dirPath); err != nil {
		es.Add(NewAppError(fmt.Sprintf("invalid configuration: path should be under the rootDir: path: %s", dirPath)))
		return nil, es
	}

	matches, err := zglob.Glob(dirPath)
	if err != nil {
		es.Add(NewAppError(fmt.Sprintf("internal error:  error raised during globbing %s", dirPath)))
		return nil, es
	}

	for _, m := range matches {
		p, err := NewLeafNodeFromFrontMatter(m)
		if err != nil {
			es.Add(err)
			continue
		}
		log.Debugf("Node Found. Type: Markdown, Filepath: %s, Title: %s, Path: %s", p.Filepath, p.Title, p.Path)
		pages = append(pages, *p)
	}

	if err := SortPageSlice(&c.SortKey, &c.SortOrder, pages); err != nil {
		es.Add(NewAppError(fmt.Sprintf("failed to sort pages: %v", err)))
		return nil, es
	}

	return pages, es
}

func transformDirectory(rootDir string, c *ConfigPage) ([]Page, ErrorSet) {
	es := NewErrorSet()

	children := make([]Page, 0, len(c.Children))
	for _, child := range c.Children {
		pages, err := buildPage(rootDir, child)

		if err.HasError() {
			es.Merge(err)
			continue
		}

		children = append(children, pages...)
	}

	p := Page{
		Type:     PageTypeDirNode,
		Title:    c.Directory,
		Children: children,
	}
	log.Debugf("Node Found. Type: Document, Title: %s", p.Title)
	return []Page{p}, es
}
