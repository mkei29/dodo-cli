package main

import (
	"encoding/json"
	"fmt"
	"io"
)

type UploadResponse struct {
	Status      string `json:"status"`
	Message     string `json:"message"`
	DocumentURL string `json:"documentURL"`
}

func ParseUploadResponse(body io.Reader) (*UploadResponse, error) {
	var response UploadResponse
	data, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	if err = json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response body: %w", err)
	}
	return &response, nil
}
