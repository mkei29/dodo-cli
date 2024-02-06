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

type PageSummary struct {
	Path  string `json:"path"`
	Title string `json:"title"`
}

func NewPageHeader(path string, title string) PageSummary {
	return PageSummary{
		Path:  path,
		Title: title,
	}
}

type Page struct {
	IsRoot   bool   `json:"is_root"`
	Title    string `json:"title"`
	Path     string `json:"path"`
	Children []Page `json:"children"`
}

func NewPageFromFrontMatter(path string) (*Page, error) {
	file, err := os.Open(path)
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
	return &page, nil
}

func NewPageFromConfig(config Config, rootDir string) (*Page, error) {
	children := make([]Page, 0, len(config.Pages))
	for _, c := range config.Pages {
		var err error
		children, err = convertToPage(children, c, rootDir)
		if err != nil {
			return nil, err
		}
	}
	return &Page{
		Title:    "",
		Path:     "",
		Children: children,
	}, nil
}

func convertToPage(slice []Page, c *ConfigPage, rootDir string) ([]Page, error) {
	if c.IsValidSinglePage() {
		path := filepath.Clean(filepath.Join(rootDir, *c.Filepath))
		if err := IsUnderRootPath(rootDir, path); err != nil {
			return nil, fmt.Errorf("path should be under the rootDir: %w", err)
		}
		children := make([]Page, 0, len(c.Children))
		var err error
		for _, c := range c.Children {
			children, err = convertToPage(children, c, rootDir)
			if err != nil {
				return nil, err
			}
		}

		slice = append(slice, Page{
			Path:     *c.Name,
			Title:    *c.Title,
			Children: children,
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
			return nil, err
		}
		for _, m := range matches {
			page, err := NewPageFromFrontMatter(m)
			if err != nil {
				return nil, err
			}
			slice = append(slice, *page)
		}
		return slice, nil
	}

	return nil, fmt.Errorf("passed ConfigPage doesn't match any pattern")
}

func (p *Page) ListPageHeader() []PageSummary {
	list := make([]PageSummary, 0, len(p.Children))
	list = listPageHeader(list, p)
	return list
}

func listPageHeader(list []PageSummary, p *Page) []PageSummary {
	list = append(list, NewPageHeader(p.Path, p.Title))
	for _, c := range p.Children {
		listPageHeader(list, &c)
	}
	return list
}

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
