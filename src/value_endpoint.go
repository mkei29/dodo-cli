package main

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"
)

var ErrNoHost = errors.New("the endpoint URL must include a scheme and a host")

type Endpoint string

func NewEndpoint(base string) (Endpoint, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("failed to parse the endpoint URL: %w", err)
	}
	if u.Scheme == "" || u.Host == "" {
		return "", ErrNoHost
	}
	return Endpoint(base), nil
}

func (endpoint Endpoint) String() string {
	return string(endpoint)
}

func (endpoint Endpoint) SearchURL() string {
	baseURL := strings.TrimSuffix(string(endpoint), "/")
	return baseURL + "/search/v1"
}

func (endpoint Endpoint) DocumentURL(slug, pathStr string) string {
	baseURL := strings.TrimSuffix(string(endpoint), "/")
	url := path.Join(slug, pathStr)
	return fmt.Sprintf("%s/document/v1/%s?format=markdown", baseURL, url)
}
