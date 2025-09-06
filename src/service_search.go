package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/toritoritori29/dodo-cli/src/openapi"
)

func sendSearchRequest(env *EnvArgs, endpoint Endpoint, query string) ([]openapi.SearchRecord, error) {
	body := openapi.SearchPostRequest{
		Query: query,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal the search request body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, endpoint.SearchURL(), strings.NewReader(string(bodyBytes)))
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

	data := openapi.SearchPostResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to parse the response: %w", err)
	}
	return data.Records, nil
}
