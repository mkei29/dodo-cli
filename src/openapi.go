// Package main provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.4.1 DO NOT EDIT.
package main

// SearchPostRequest defines model for search_post_request.
type SearchPostRequest struct {
	Projects []string `json:"projects"`
	Query    string   `json:"query"`
}

// SearchPostResponse defines model for search_post_response.
type SearchPostResponse struct {
	Message string         `json:"message"`
	Records []SearchRecord `json:"records"`
	Status  string         `json:"status"`
}

// SearchPostResponseError defines model for search_post_response_error.
type SearchPostResponseError struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

// SearchRecord defines model for search_record.
type SearchRecord struct {
	Contents    string `json:"contents"`
	Id          string `json:"id"`
	ProjectId   string `json:"project_id"`
	ProjectSlug string `json:"project_slug"`
	Title       string `json:"title"`
	Url         string `json:"url"`
}

// SearchContentsJSONRequestBody defines body for SearchContents for application/json ContentType.
type SearchContentsJSONRequestBody = SearchPostRequest
