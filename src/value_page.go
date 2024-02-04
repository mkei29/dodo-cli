package main

import "fmt"

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
	Path     string
	Title    string
	Children []Page `json:"children"`
}

func NewPageFromConfig(config Config) (*Page, error) {
	children := make([]Page, 0, len(config.Page))
	for _, c := range config.Page {
		var err error
		children, err = convertToPage(children, c)
		if err != nil {
			return nil, err
		}
	}
	return &Page{
		Path:     "",
		Title:    "",
		Children: children,
	}, nil
}

func convertToPage(slice []Page, c *ConfigPage) ([]Page, error) {
	if c.IsValidSinglePage() {
		children := make([]Page, 0, len(c.Children))
		var err error
		for _, c := range c.Children {
			children, err = convertToPage(children, c)
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
