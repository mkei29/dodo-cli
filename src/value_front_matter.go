package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/adrg/frontmatter"
	"gopkg.in/yaml.v3"
)

const (
	FrontMatterStart = "---"
	FrontMatterEnd   = "---"

	FrontMatterKeyTitle     = "title"
	FrontMatterKeyPath      = "path"
	FrontMatterDescription  = "description"
	FrontMatterKeyCreatedAt = "created_at"
	FrontMatterKeyUpdatedAt = "updated_at"
)

// A struct that describes the markdown header.
type FrontMatter struct {
	Title       string
	Path        string
	Description string
	CreatedAt   SerializableTime
	UpdatedAt   SerializableTime
	UnknownTags map[string]interface{}
}

func NewFrontMatter(title string, path string, now ...time.Time) *FrontMatter {
	var currentTime time.Time
	if len(now) > 0 {
		currentTime = now[0]
	} else {
		currentTime = time.Now()
	}
	return &FrontMatter{
		Title:       title,
		Path:        path,
		Description: "",
		CreatedAt:   NewSerializableTimeFromTime(currentTime),
		UpdatedAt:   NewSerializableTimeFromTime(currentTime),
		UnknownTags: make(map[string]interface{}),
	}
}

func NewFrontMatterFromMarkdown(filepath string) (*FrontMatter, error) { //nolint: cyclop
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", filepath)
	}
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	formats := []*frontmatter.Format{
		frontmatter.NewFormat(FrontMatterStart, FrontMatterEnd, yaml.Unmarshal),
	}
	var kv map[string]string
	_, err = frontmatter.Parse(file, &kv, formats...)
	if err != nil {
		return nil, fmt.Errorf("failed to parse front matter: %w", err)
	}

	matter := FrontMatter{
		UnknownTags: make(map[string]interface{}),
	}
	for k, v := range kv {
		switch strings.ToLower(k) {
		case FrontMatterKeyTitle:
			matter.Title = v
		case FrontMatterKeyPath:
			matter.Path = v
		case FrontMatterDescription:
			matter.Description = v
		case FrontMatterKeyCreatedAt:
			st, err := NewSerializableTime(v)
			if err != nil {
				return nil, fmt.Errorf("`created_at` must follow the RFC3339 format. Got: %s", v)
			}
			matter.CreatedAt = st
		case FrontMatterKeyUpdatedAt:
			st, err := NewSerializableTime(v)
			if err != nil {
				return nil, fmt.Errorf("`updated_at` must follow the RFC3339 format. Got: %s", v)
			}
			matter.UpdatedAt = st
		default:
			matter.UnknownTags[k] = v
		}
	}
	return &matter, nil
}

// UpdateMarkdown updates the front matter of the specified markdown file.
// It keeps the remaining contents of the file intact.
func (f *FrontMatter) UpdateMarkdown(filepath string) error {
	file, err := os.OpenFile(filepath, os.O_RDWR, 0o644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	formats := []*frontmatter.Format{
		frontmatter.NewFormat(FrontMatterStart, FrontMatterEnd, yaml.Unmarshal),
	}
	var v map[string]interface{}
	remaining, err := frontmatter.Parse(file, &v, formats...)
	if err != nil {
		return fmt.Errorf("failed to parse front matter: %w", err)
	}

	contents := []byte(f.String())
	contents = append(contents, remaining...)

	// Truncate file and rewrite the contents
	if err := file.Truncate(0); err != nil {
		return fmt.Errorf("failed to truncate file: %w", err)
	}
	_, err = file.WriteAt(contents, 0)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}
	return nil
}

func (f *FrontMatter) String() string {
	// Prepare sorted unknown tag keys
	sortedKeys := make([]string, 0, len(f.UnknownTags))
	for k := range f.UnknownTags {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	var text string
	text += FrontMatterStart + "\n"
	text += fmt.Sprintf("title: \"%s\"\n", f.Title)
	text += fmt.Sprintf("path: \"%s\"\n", f.Path)
	text += fmt.Sprintf("description: \"%s\"\n", f.Description)
	text += fmt.Sprintf("created_at: \"%s\"\n", f.CreatedAt)
	text += fmt.Sprintf("updated_at: \"%s\"\n", f.UpdatedAt)
	for _, k := range sortedKeys {
		text += fmt.Sprintf("%s: %s\n", k, f.UnknownTags[k])
	}
	text += FrontMatterEnd + "\n"
	return text
}
