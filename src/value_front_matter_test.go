package main

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewFrontMatter(t *testing.T) {
	title := "Test Title"
	path := "/test/path"
	fm := NewFrontMatter(title, path)

	assert.Equal(t, title, fm.Title, "expected title to be %s, got %s", title, fm.Title)
	assert.Equal(t, path, fm.Path, "expected path to be %s, got %s", path, fm.Path)
	assert.False(t, fm.CreatedAt.IsNull(), "expected CreatedAt to be set, got null")
	assert.False(t, fm.UpdatedAt.IsNull(), "expected UpdatedAt to be set, got null")
}

func TestNewFrontMatterFromMarkdown(t *testing.T) {
	content := `---
title: Test Title
path: /test/path
created_at: 2023-10-01T00:00:00Z
updated_at: 2023-10-01T00:00:00Z
---`

	tmpfile, err := os.CreateTemp("", "test*.md")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("failed to close temp file: %v", err)
	}

	fm, err := NewFrontMatterFromMarkdown(tmpfile.Name())
	if err != nil {
		t.Fatalf("failed to create FrontMatter from markdown: %v", err)
	}

	assert.Equal(t, "Test Title", fm.Title, "expected title 'Test Title', got %s", fm.Title)
	assert.Equal(t, "/test/path", fm.Path, "expected path '/test/path', got %s", fm.Path)

	expectedTime := SerializableTime("2023-10-01T00:00:00Z")
	assert.Equal(t, expectedTime, fm.CreatedAt, "expected CreatedAt %v, got %v", expectedTime, fm.CreatedAt)
	assert.Equal(t, expectedTime, fm.UpdatedAt, "expected UpdatedAt %v, got %v", expectedTime, fm.UpdatedAt)
}

func TestFrontMatterString(t *testing.T) {
	fm := NewFrontMatter("Test Title", "/test/path")
	fm.CreatedAt = NewSerializableTimeFromTime(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	fm.UpdatedAt = NewSerializableTimeFromTime(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))

	expected := `---
title: Test Title
path: /test/path
created_at: 2025-01-01T00:00:00Z
updated_at: 2025-01-01T00:00:00Z
---
`

	if fm.String() != expected {
		t.Errorf("expected %s, got %s", expected, fm.String())
	}
}
