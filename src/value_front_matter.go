package main

import (
	"fmt"
	"os"
	"time"

	"github.com/adrg/frontmatter"
	"gopkg.in/yaml.v3"
)

const (
	FrontMatterStart = "---"
	FrontMatterEnd   = "---"
)

// A struct that describe the markdown header.
type FrontMatter struct {
	Title     string
	Path      string
	CreatedAt SerializableTime
	UpdatedAt SerializableTime
}

func NewFrontMatter(title string, path string) *FrontMatter {
	now := time.Now()
	return &FrontMatter{
		Title:     title,
		Path:      path,
		CreatedAt: NewSerializableTimeFromTime(now),
		UpdatedAt: NewSerializableTimeFromTime(now),
	}
}

func NewFrontMatterFromMarkdown(filepath string) (*FrontMatter, error) {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", filepath)
	}
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	defer file.Close()

	formats := []*frontmatter.Format{
		frontmatter.NewFormat(FrontMatterStart, FrontMatterEnd, yaml.Unmarshal),
	}

	var v map[string]interface{}
	_, err = frontmatter.Parse(file, &v, formats...)
	if err != nil {
		return nil, fmt.Errorf("failed to parse front matter: %w", err)
	}

	matter := FrontMatter{}
	if title, ok := asString(v, "title"); ok {
		matter.Title = title
	}
	if path, ok := asString(v, "path"); ok {
		matter.Path = path
	}
	if createdAt, ok := asSerializedTime(v, "created_at"); ok {
		matter.CreatedAt = *createdAt
	}
	if updatedAt, ok := asSerializedTime(v, "updated_at"); ok {
		matter.UpdatedAt = *updatedAt
	}
	return &matter, nil
}

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

	_, err = file.WriteAt(contents, 0)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}

func (f *FrontMatter) String() string {
	var text string
	text += fmt.Sprintf("%s\n", FrontMatterStart)
	text += fmt.Sprintf("title: %s\n", f.Title)
	text += fmt.Sprintf("path: %s\n", f.Path)
	text += fmt.Sprintf("created_at: %s\n", f.CreatedAt)
	text += fmt.Sprintf("updated_at: %s\n", f.UpdatedAt)
	text += fmt.Sprintf("%s\n", FrontMatterEnd)
	return text
}

func asString(m map[string]interface{}, key string) (string, bool) {
	v, ok := m[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

func asSerializedTime(m map[string]interface{}, key string) (*SerializableTime, bool) {
	v, ok := m[key]
	if !ok {
		return nil, false
	}

	switch v := v.(type) {
	case string:
		st, err := NewSerializableTime(v)
		return &st, err == nil
	case time.Time:
		st := NewSerializableTimeFromTime(v)
		return &st, true
	}
	return nil, false
}
