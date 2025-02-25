package main

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFrontMatter(t *testing.T) {
	title := "Test Title"
	path := "/test/path"
	fm := NewFrontMatter(title, path)

	assert.Equal(t, title, fm.Title, "expected title to be %s, got %s", title, fm.Title)
	assert.Equal(t, path, fm.Path, "expected path to be %s, got %s", path, fm.Path)
	assert.False(t, fm.CreatedAt.IsZero(), "expected CreatedAt to be set, got zero value")
	assert.False(t, fm.UpdatedAt.IsZero(), "expected UpdatedAt to be set, got zero value")
	assert.Equal(t, "", fm.Description, "expected description to be empty, got %s", fm.Description)
}

func TestNewFrontMatterFromMarkdown(t *testing.T) {
	content := `---
title: Test Title
path: /test/path
description: Test Description
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
	assert.Equal(t, "Test Description", fm.Description, "expected description 'Test Description', got %s", fm.Description)

	expectedTime, err := NewSerializableTime("2023-10-01T00:00:00Z")
	require.NoError(t, err)
	assert.Equal(t, expectedTime, fm.CreatedAt, "expected CreatedAt %v, got %v", expectedTime, fm.CreatedAt)
	assert.Equal(t, expectedTime, fm.UpdatedAt, "expected UpdatedAt %v, got %v", expectedTime, fm.UpdatedAt)
}

func TestNewFrontMatterFromMarkdownWithoutFrontMater(t *testing.T) {
	content := "Test Content"

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

	_, err = NewFrontMatterFromMarkdown(tmpfile.Name())
	require.NoError(t, err, "expected no error when creating FrontMatter from empty file")
}

func TestNewFrontMatterFromMarkdownWithUnknownTags(t *testing.T) {
	content := `---
title: Test Title
path: /test/path
description: Test Description
created_at: 2023-10-01T00:00:00Z
updated_at: 2023-10-01T00:00:00Z
UnknownTag1: "unknown1"
UnknownTag2: "unknown2"
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

	assert.Contains(t, fm.UnknownTags, "UnknownTag1", "expected UnknownTag1 to be included in UnknownTags")
	assert.Contains(t, fm.UnknownTags, "UnknownTag2", "expected UnknownTag2 to be included in UnknownTags")

	assert.Equal(t, "unknown1", fm.UnknownTags["UnknownTag1"], "expected UnknownTag1 to be 'unknown1', got %s", fm.UnknownTags["UnknownTag1"])
	assert.Equal(t, "unknown2", fm.UnknownTags["UnknownTag2"], "expected UnknownTag2 to be 'unknown2', got %s", fm.UnknownTags["UnknownTag2"])
	assert.Equal(t, "Test Description", fm.Description, "expected description 'Test Description', got %s", fm.Description)
}

func TestFrontMatterString(t *testing.T) {
	fm := NewFrontMatter("Test Title", "/test/path")
	fm.Description = "Test Description"
	fm.CreatedAt = NewSerializableTimeFromTime(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	fm.UpdatedAt = NewSerializableTimeFromTime(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	fm.UnknownTags["UnknownTag1"] = "unknown1"
	fm.UnknownTags["UnknownTag2"] = "unknown2"

	expected := `---
title: Test Title
path: /test/path
description: Test Description
created_at: 2025-01-01T00:00:00Z
updated_at: 2025-01-01T00:00:00Z
UnknownTag1: unknown1
UnknownTag2: unknown2
---
`

	if fm.String() != expected {
		t.Errorf("expected %s, got %s", expected, fm.String())
	}
}
