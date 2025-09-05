package main

import (
	"context"
	"fmt"

	"github.com/caarlos0/log"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
	"github.com/toritoritori29/dodo-cli/src/openapi"
)

const SearchDescription = `
Search documents from the dodo-doc.
---
dodo is a service that hosts users' documents.
This tool searches for documents across the entire dodo platform based on a given query.
You can use this tool to search for user-specific knowledge, design documents, or any specialized expertise.


Search results are returned in json format with the following structure:
{
"items": [
{
"title": "document title",
"contents": "document contents",
"id": "document id",
"project_id": "project id",
"project_slug": "project slug",
"url": "document url"
}
]
}
`

const DocumentDescription = `
Read the content of a document from dodo-doc.
---
dodo is a service that hosts users' documents.
This tool retrieves the content of a specific document identified by its project slug and path.
You can get the native markdown content of the document.

The slug and path can be found in the document URL.
For example, if the document URL is https://docs.do.dodo-doc.com/hello_world/getting_started,
the slug is "docs" and the path is "hello_world/getting_started".

To find the document URL, you can use the "search" tool to search for documents and obtain their URLs.
`

type MCPArgs struct {
	configPath string // config file path
	debug      bool   // server endpoint to upload
}

func CreateMCPCmd() *cobra.Command {
	opts := MCPArgs{}
	initCmd := &cobra.Command{
		Use:   "mcp",
		Short: "Start a MCP server",
		RunE: func(_ *cobra.Command, _ []string) error {
			printer := NewPrinter(ErrorLevel)
			if err := mcpCmdEntrypoint(&opts); err != nil {
				printer.PrintError(err)
				return err
			}
			return nil
		},
	}
	initCmd.Flags().StringVarP(&opts.configPath, "config", "c", ".dodo.yaml", "Path to the configuration file")
	initCmd.Flags().BoolVar(&opts.debug, "debug", false, "Enable debug mode")
	return initCmd
}

func mcpCmdEntrypoint(_ *MCPArgs) error {
	s := server.NewMCPServer(
		"dodo-doc-mcp",
		"0.0.1",
		server.WithToolCapabilities(false),
	)
	h := &Handler{
		Env:         NewEnvArgs(),
		SearchURI:   "http://contents.test-doc.com/search/v1",
		DocumentURI: "http://contents.test-doc.com/document/v1",
	}
	s.AddTool(addSearchTool(), h.addSearchHandler)
	s.AddTool(addReadDocumentTool(), h.addReadDocumentHandler)

	log.Infof("Starting MCP server...")
	if err := server.ServeStdio(s); err != nil {
		return fmt.Errorf("failed to start MCP server: %w", err)
	}
	return nil
}

func addSearchTool() mcp.Tool {
	return mcp.NewTool("search", mcp.WithDescription(SearchDescription), mcp.WithString("query", mcp.Required(), mcp.Description("The search query")))
}

func addReadDocumentTool() mcp.Tool {
	tool := mcp.NewTool(
		"read_document",
		mcp.WithDescription("Read the content of a document from dodo-doc"),
		mcp.WithString("slug", mcp.Required(), mcp.Description("The project slug where the document belongs")),
		mcp.WithString("path", mcp.Required(), mcp.Description("The path of the document to read")),
	)
	return tool
}

type Handler struct {
	Env         EnvArgs
	SearchURI   string
	DocumentURI string
}

type addSearchHandlerResult struct {
	Items []openapi.SearchRecord `json:"items"`
}

func (h *Handler) addSearchHandler(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query := request.GetString("query", "")
	items, err := sendSearchRequest(&h.Env, h.SearchURI, query)
	if err != nil {
		return nil, err
	}

	result := addSearchHandlerResult{
		Items: items,
	}
	return mcp.NewToolResultStructured(result, ""), nil
}

func (h *Handler) addReadDocumentHandler(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	slug := request.GetString("slug", "")
	path := request.GetString("path", "")
	content, err := sendReadDocumentRequest(&h.Env, h.DocumentURI, slug, path)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(content), nil
}
