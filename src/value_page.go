package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/frontmatter"
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
	Type        string
	Filepath    string `json:"filepath"`
	Path        string `json:"path"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func NewPageSummary(filepath, path, title string) PageSummary {
	return PageSummary{
		Filepath: filepath,
		Path:     path,
		Title:    title,
	}
}

func NewPageHeaderFromPage(p *Page) PageSummary {
	return PageSummary{
		Type:        p.Type,
		Filepath:    p.Filepath,
		Path:        p.Path,
		Title:       p.Title,
		Description: p.Description,
	}
}

type Page struct {
	Type        string `json:"type"`
	Filepath    string `json:"filepath"`
	Path        string `json:"path"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Children    []Page `json:"children"`
}

func NewPageFromFrontMatter(filePath string) (*Page, error) {
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
	page.Children = []Page{}
	return &page, nil
}

func CreatePageTree(config Config, rootDir string) (*Page, *ErrorSet) {
	children := make([]Page, 0, len(config.Pages))
	errorSet := NewErrorSet()
	for _, c := range config.Pages {
		children = convertToPage(errorSet, children, c, rootDir)
	}
	return &Page{
		Type:        PageTypeRootNode,
		Filepath:    "",
		Title:       "",
		Path:        "",
		Description: "",
		Children:    children,
	}, errorSet
}

func convertToPage(errorSet *ErrorSet, slice []Page, c *ConfigPage, rootDir string) []Page { //nolint: funlen, cyclop
	if c.IsValidSinglePage() {
		filepath := filepath.Clean(filepath.Join(rootDir, *c.Filepath))
		if err := IsUnderRootPath(rootDir, filepath); err != nil {
			errorSet.Add(NewAppError(fmt.Sprintf("path should be under the rootDir. passed: %s", filepath)))
			return nil
		}

		children := make([]Page, 0, len(c.Children))
		for _, c := range c.Children {
			child := convertToPage(errorSet, children, c, rootDir)
			if child != nil {
				children = append(children, child...)
			}
		}

		description := ""
		if c.Description != nil {
			description = *c.Description
		}

		slice = append(slice, Page{
			Type:        PageTypeLeafNode,
			Filepath:    filepath,
			Path:        *c.Path,
			Title:       *c.Title,
			Description: description,
			Children:    children,
		})
		return slice
	}

	if c.IsValidMatchDirectory() {
		dirPath := filepath.Clean(filepath.Join(rootDir, *c.Match))
		if err := IsUnderRootPath(rootDir, dirPath); err != nil {
			errorSet.Add(NewAppError(fmt.Sprintf("invalid configuration: path should be under the rootDir: path: %s", dirPath)))
			return nil
		}
		matches, err := zglob.Glob(dirPath)
		if err != nil {
			errorSet.Add(NewAppError(fmt.Sprintf("internal error:  error raised during globbing %s", dirPath)))
			return nil
		}

		children := make([]Page, 0, len(matches))
		for _, m := range matches {
			page, err := NewPageFromFrontMatter(m)
			if err != nil {
				errorSet.Add(NewAppError(fmt.Sprintf("invalid configuration: cannot read a file: path %s", dirPath)))
			}
			children = append(children, *page)
		}

		slice = append(slice, Page{
			Type:     PageTypeDirNodeWithPage,
			Title:    *c.Title,
			Children: children,
		})
		return slice
	}

	errorSet.Add(NewAppError("invalid configuration: doesn't match any pattern"))
	return nil
}

// List PageSummary of the page includes.
func (p *Page) ListPageHeader() []PageSummary {
	list := make([]PageSummary, 0, len(p.Children))
	list = listPageHeader(list, p)
	return list
}

func listPageHeader(list []PageSummary, p *Page) []PageSummary {
	if p.Type != PageTypeRootNode {
		list = append(list, NewPageHeaderFromPage(p))
	}
	for idx := range p.Children {
		list = listPageHeader(list, &p.Children[idx])
	}
	return list
}

// Check if the page is valid.
// This function checks the following conditions:
// 1. all page have necessary fields.
// 2. There are no duplicated paths.
func (p *Page) IsValid() *ErrorSet {
	errorSet := NewErrorSet()
	p.isValid(true, errorSet)
	if errorSet.HasError() {
		return errorSet
	}

	pathMap := make(map[string]int)
	p.duplicationCount(pathMap)
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
		errorSet.Add(NewAppError(fmt.Sprintf("%s Path field of child page should not be empty", p.Filepath)))
	}
	for _, c := range p.Children {
		c.isValid(false, errorSet)
	}
}

func (p *Page) duplicationCount(pathMap map[string]int) {
	if p.Path != "" {
		pathMap[p.Path]++
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
