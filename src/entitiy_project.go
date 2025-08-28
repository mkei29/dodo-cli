package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/toritoritori29/dodo-cli/src/openapi"
)

type Project struct {
	BaseURL        string
	IsPublic       bool
	OrganizationID string
	ProjectID      string
	ProjectName    string
	Slug           string
}

func NewProjectFromAPI(env EnvArgs, uri string) ([]Project, error) {
	req, err := http.NewRequest(http.MethodGet, uri, strings.NewReader(""))
	if err != nil {
		return nil, fmt.Errorf("failed to create a new request from the body: %w", err)
	}
	bearer := "Bearer " + env.APIKey
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", bearer)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send a request to the server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("the server returned non-200 status code: %d", resp.StatusCode)
	}

	data := openapi.ProjectsGetResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to parse the response: %w", err)
	}

	result := make([]Project, 0, len(data.Projects))
	for _, p := range data.Projects {
		result = append(result, Project{
			BaseURL:        p.BaseUrl,
			IsPublic:       p.IsPublic,
			OrganizationID: p.OrganizationId,
			ProjectID:      p.ProjectId,
			ProjectName:    p.ProjectName,
			Slug:           p.Slug,
		})
	}
	return result, nil
}
