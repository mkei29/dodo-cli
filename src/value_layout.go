package main

type Layout struct {
	IsDir    bool     `yaml:"is_dir"`
	Title    string   `yaml:"title"`
	Path     string   `yaml:"path"`
	Children []Layout `yaml:"children"`
}

func NewLayout() Layout {
	return Layout{IsDir: true, Title: "", Path: "", Children: []Layout{}}
}

type PageHeader struct {
	Path  string `json:"path"`
	Title string `json:"title"`
}

func NewPageHeader(path string, title string) PageHeader {
	return PageHeader{
		Path:  path,
		Title: title,
	}
}
