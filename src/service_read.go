package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/toritoritori29/dodo-cli/src/openapi"
)

func sendReadDocumentRequest(env *EnvArgs, endpoint Endpoint, slug, path string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, endpoint.DocumentURL(slug, path), strings.NewReader(""))
	if err != nil {
		return "", fmt.Errorf("failed to create a new request from the body: %w", err)
	}
	bearer := "Bearer " + env.APIKey
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", bearer)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send a request to the server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("the server returned non-200 status code: %d", resp.StatusCode)
	}

	data := openapi.DocumentGetResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", fmt.Errorf("failed to parse the response: %w", err)
	}
	if data.Markdown == nil {
		return "", errors.New("the response does not contain markdown data")
	}
	return *data.Markdown, nil
}

func parseURL(documentURL string) (string, string, error) {
	u, err := url.Parse(documentURL)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse the document URL: %w", err)
	}
	hostname := u.Hostname()
	domains := strings.Split(hostname, ".")
	if len(domains) != 4 {
		return "", "", errors.New("invalid document URL format")
	}
	slug := domains[0]
	p := u.Path
	return slug, p, nil
}
