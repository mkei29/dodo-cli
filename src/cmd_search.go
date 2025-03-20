package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/caarlos0/log"
	"github.com/spf13/cobra"
)

type SearchArgs struct {
	query    []string // search query
	endpoint string   // server endpoint to search
	debug    bool     // enable debug mode
}

func CreateSearchCmd() *cobra.Command {
	opts := SearchArgs{}
	searchCmd := &cobra.Command{
		Use:   "search",
		Short: "Search for a string in the project files.",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.query = args
			return executeSearch(cmd, opts)
		},
	}
	searchCmd.Flags().BoolVar(&opts.debug, "debug", false, "Enable debug mode")
	searchCmd.Flags().StringVar(&opts.endpoint, "endpoint", "https://contents.dodo-doc.com/search/v1", "Server endpoint to search")
	return searchCmd
}

func executeSearch(cmd *cobra.Command, args SearchArgs) error {
	// Initialize logger and so on, then execute the main function.

	if args.debug {
		log.SetLevel(log.DebugLevel)
		log.Debug("running in debug mode")
	}

	env := NewEnvArgs()
	if args.debug {
		log.SetLevel(log.DebugLevel)
		log.Debug("running in debug mode")
	}

	query := strings.Join(args.query, " ")
	query = strings.TrimSpace(query)
	if query == "" {
		return errors.New("search query cannot be empty")
	}

	log.Infof("Searching for '%s'", query)
	if args.endpoint == "" {
		return errors.New("endpoint cannot be empty")
	}

	// Call the function to send the search request
	log.Infof("Using endpoint: %s", args.endpoint)
	records, err := sendSearchRequest(env, args.endpoint, query)
	if err != nil {
		return fmt.Errorf("failed to execute search: %w", err)
	}

	for _, record := range records {
		log.Infof("Found: %s", record.Title)
	}

	return nil
}

func sendSearchRequest(env EnvArgs, uri, query string) ([]SearchRecord, error) {
	body := SearchPostRequest{
		Query: query,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search request body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, uri, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, fmt.Errorf("failed to create a new upload request from body: %w", err)
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
	log.Infof("Response status: %s", resp.Status)

	data := SearchPostResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to parse the response: %w", err)
	}
	return data.Records, nil
}
