package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/frontmatter"
	"github.com/caarlos0/log"
	"github.com/mattn/go-zglob"
	"gopkg.in/yaml.v3"
)

type PageSummary struct {
	Filepath    string `json:"filepath"`
	Path        string `json:"path"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func NewPageHeader(filepath, path, title string) PageSummary {
	return PageSummary{
		Filepath: filepath,
		Path:     path,
		Title:    title,
	}
}

func NewPageHeaderFromPage(p *Page) PageSummary {
	return NewPageHeader(p.Filepath, p.Path, p.Title)
}

type Page struct {
	IsRoot      bool   `json:"is_root"`
	Filepath    string `json:"filepath"`
	Path        string `json:"path"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Children    []Page `json:"children"`
}

func NewPageFromFrontMatter(filePath, parentPath string) (*Page, error) {
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
	page.Path = filepath.Join(parentPath, page.Path)
	page.Filepath = filePath
	page.IsRoot = false
	page.Children = []Page{}
	return &page, nil
}

func NewPageFromConfig(config Config, rootDir string) (*Page, error) {
	children := make([]Page, 0, len(config.Pages))
	for _, c := range config.Pages {
		var err error
		children, err = convertToPage(children, c, rootDir, "")
		if err != nil {
			return nil, err
		}
	}
	return &Page{
		IsRoot:      true,
		Filepath:    "",
		Title:       "",
		Path:        "",
		Description: "",
		Children:    children,
	}, nil
}

func convertToPage(slice []Page, c *ConfigPage, rootDir string, parentPath string) ([]Page, error) {
	if c.IsValidSinglePage() {
		joinedPath := filepath.Join(parentPath, *c.Path)
		path := filepath.Clean(filepath.Join(rootDir, *c.Filepath))
		if err := IsUnderRootPath(rootDir, path); err != nil {
			return nil, fmt.Errorf("path should be under the rootDir: %w", err)
		}
		children := make([]Page, 0, len(c.Children))
		var err error
		for _, c := range c.Children {
			children, err = convertToPage(children, c, rootDir, joinedPath)
			if err != nil {
				return nil, err
			}
		}
		description := ""
		if c.Description != nil {
			description = *c.Description
		}

		slice = append(slice, Page{
			IsRoot:      false,
			Filepath:    path,
			Path:        joinedPath,
			Title:       *c.Title,
			Description: description,
			Children:    children,
		})
		return slice, nil
	}

	if c.IsValidPatternPage() {
		path := filepath.Clean(filepath.Join(rootDir, *c.Match))
		if err := IsUnderRootPath(rootDir, path); err != nil {
			return nil, fmt.Errorf("path should be under the rootDir: %w", err)
		}
		matches, err := zglob.Glob(path)
		if err != nil {
			log.Info(fmt.Sprintf("error %s not found", path))
			return nil, err
		}
		for _, m := range matches {
			page, err := NewPageFromFrontMatter(m, parentPath)
			if err != nil {
				return nil, err
			}
			slice = append(slice, *page)
		}
		return slice, nil
	}

	return nil, fmt.Errorf("passed ConfigPage doesn't match any pattern")
}

// List PageSummary of the page includes.
func (p *Page) ListPageHeader() []PageSummary {
	list := make([]PageSummary, 0, len(p.Children))
	list = listPageHeader(list, p)
	return list
}

func listPageHeader(list []PageSummary, p *Page) []PageSummary {
	if !p.IsRoot {
		list = append(list, NewPageHeaderFromPage(p))
	}
	for _, c := range p.Children {
		list = listPageHeader(list, &c)
	}
	return list
}

// Check if the page is valid.
// This function checks the following conditions:
// 1. all page have necessary fields.
// 2. There are no duplicated paths.
func (p *Page) IsValid() bool {
	if !p.isValid(true) {
		return false
	}

	pathMap := make(map[string]int)
	p.duplicationCount(pathMap)
	for _, value := range pathMap {
		if value > 1 {
			return false
		}
	}
	return true
}

func (p *Page) isValid(isRoot bool) bool {
	if isRoot && !p.IsRoot {
		log.Debug("IsRoot of root page should be true")
		return false
	}
	if !isRoot && p.IsRoot {
		log.Debug("IsRoot field of child page should be false")
		return false
	}
	if !isRoot && p.Path == "" {
		log.Debugf("Path field of child page should not be empty: %s", p.Filepath)
		return false
	}

	for _, c := range p.Children {
		ok := c.isValid(false)
		if !ok {
			return false
		}
	}
	return true
}

func (p *Page) duplicationCount(pathMap map[string]int) {
	pathMap[p.Path]++
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
