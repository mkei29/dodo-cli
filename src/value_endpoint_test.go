package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEndpoint(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		hasError bool
	}{
		{"https://example.com", "https://example.com", false},
		{"http://localhost:8080", "http://localhost:8080", false},
		{"https://api.dodo.dev", "https://api.dodo.dev", false},
		{"https://example.com/path", "https://example.com/path", false},
		{"", "", true},
		{"example.com", "", true},
		{"invalid-url", "", true},
		{"/path/only", "", true},
	}

	for _, test := range tests {
		result, err := NewEndpoint(test.input)
		if test.hasError {
			require.Error(t, err, "Expected error for input: %s", test.input)
		} else {
			require.NoError(t, err, "Unexpected error for input: %s", test.input)
			assert.Equal(t, Endpoint(test.expected), result, "Endpoint(%s) = %v, expected %v", test.input, result, test.expected)
		}
	}
}

func TestEndpoint_SearchURL(t *testing.T) {
	tests := []struct {
		endpoint Endpoint
		expected string
	}{
		{Endpoint("https://example.com"), "https://example.com/search/v1"},
		{Endpoint("https://example.com/"), "https://example.com/search/v1"},
		{Endpoint("http://localhost:8080"), "http://localhost:8080/search/v1"},
		{Endpoint("https://api.dodo.dev/base"), "https://api.dodo.dev/base/search/v1"},
	}

	for _, test := range tests {
		result := test.endpoint.SearchURL()
		assert.Equal(t, test.expected, result, "Endpoint(%s).SearchURL() = %v, expected %v", test.endpoint, result, test.expected)
	}
}

func TestEndpoint_DocumentURL(t *testing.T) {
	tests := []struct {
		endpoint Endpoint
		slug     string
		path     string
		expected string
	}{
		{Endpoint("https://example.com"), "myproject", "doc.md", "https://example.com/document/v1/myproject/doc.md?format=markdown"},
		{Endpoint("https://example.com/"), "myproject", "folder/doc.md", "https://example.com/document/v1/myproject/folder/doc.md?format=markdown"},
		{Endpoint("http://localhost:8080"), "test", "index.md", "http://localhost:8080/document/v1/test/index.md?format=markdown"},
		{Endpoint("https://api.dodo.dev/base"), "proj", "docs/readme.md", "https://api.dodo.dev/base/document/v1/proj/docs/readme.md?format=markdown"},
	}

	for _, test := range tests {
		result := test.endpoint.DocumentURL(test.slug, test.path)
		assert.Equal(t, test.expected, result, "Endpoint(%s).DocumentURL(%s, %s) = %v, expected %v", test.endpoint, test.slug, test.path, result, test.expected)
	}
}
