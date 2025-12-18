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
	Type        string                  `json:"type"`
	Filepath    string                  `json:"filepath"`
	Hash        string                  `json:"hash"`
	Path        string                  `json:"path"`
	Title       string                  `json:"title"`
	Description string                  `json:"description"`
	UpdatedAt   config.SerializableTime `json:"updated_at"`
	Children    []Page                  `json:"children"`
}

func NewLeafNodeFromConfigPge(config *config.ConfigPage) Page {
	page := Page{
		Type:        PageTypeLeafNode,
		Filepath:    config.Markdown,
		Hash:        fmt.Sprintf("%x", sha256.Sum256([]byte(config.Markdown))),
		Path:        config.Path,
		Title:       config.Title,
		Description: config.Description,
		UpdatedAt:   config.UpdatedAt,
		Children:    []Page{},
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
func (p *Page) IsValid() *appErrors.MultiError {
	errorSet := appErrors.NewMultiError()
	p.isValid(true, &errorSet)
	if errorSet.HasError() {
		return &errorSet
	}

	pathMap := make(map[string]int)
	p.duplicationCount(pathMap, "")
	for path, value := range pathMap {
		if value > 1 {
			errorSet.Add(appErrors.NewAppError(fmt.Sprintf("duplicated path was found. path: `%s`", path)))
		}
	}
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
	if p.Type == PageTypeLeafNode && p.Path == "" {
		errorSet.Add(appErrors.NewAppError("'path' is required for leaf node"))
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

func CreatePageTree(conf *config.Config, rootDir string) (*Page, *appErrors.MultiError) {
	errorSet := appErrors.NewMultiError()
	root := Page{
		Type: PageTypeRootNode,
	}

	children := make([]Page, 0, len(conf.Pages))
	for _, p := range conf.Pages {
		c, merr := buildPage(rootDir, p)
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

func buildPage(rootDir string, c config.ConfigPage) ([]Page, *appErrors.MultiError) {
	if c.MatchMarkdown() {
		return transformMarkdown(rootDir, &c)
	}

	if c.MatchDirectory() {
		return transformDirectory(rootDir, &c)
	}

	err := appErrors.NewMultiError()
	err.Add(appErrors.NewAppError("invalid configuration: doesn't match any pattern"))
	return nil, &err
}

func transformMarkdown(rootDir string, c *config.ConfigPage) ([]Page, *appErrors.MultiError) {
	merr := appErrors.NewMultiError()
	filepath := filepath.Clean(filepath.Join(rootDir, c.Markdown))

	if err := config.IsUnderRootPath(rootDir, filepath); err != nil {
		merr.Add(fmt.Errorf("path should be under the rootDir. passed: %s", filepath))
		return nil, &merr
	}

	// First, populate the fields from the markdown front matter.
	p := NewLeafNodeFromConfigPge(c)
	log.Debugf("Node Found. Type: Markdown, Filepath: '%s', Title: '%s', Path: '%s'", p.Filepath, p.Title, p.Path)
	return []Page{p}, nil
}

func transformDirectory(rootDir string, c *config.ConfigPage) ([]Page, *appErrors.MultiError) {
	merr := appErrors.NewMultiError()

	children := make([]Page, 0, len(c.Children))
	for _, child := range c.Children {
		pages, err := buildPage(rootDir, child)
		if err != nil {
			merr.Merge(*err)
			continue
		}
		children = append(children, pages...)
	}

	p := Page{
		Type:     PageTypeDirNode,
		Title:    c.Directory,
		Children: children,
	}
	if merr.HasError() {
		return nil, &merr
	}
	log.Debugf("Node Found. Type: Directory, Title: %s", p.Title)
	return []Page{p}, nil
}
