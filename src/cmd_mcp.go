package main

import (
	"context"
	"errors"
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
After you receive the search results, you can use the "read_document" tool to read the content of a specific document.


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
Read the content of a document from dodo-doc
---
This tool allows you to retrieve the full content of a document hosted on dodo, a document hosting service.
You can fetch the document’s native Markdown content.
The document URL can be obtained from the results of the search tool.
Use this tool when you need to read or process the raw content of a document directly.
`

type MCPArgs struct {
	configPath string // config file path
	debug      bool   // server endpoint to upload
	endpoint   string
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
	initCmd.Flags().StringVar(&opts.endpoint, "endpoint", "https://contents.dodo-doc.com/", "server endpoint to upload")
	return initCmd
}

func mcpCmdEntrypoint(args *MCPArgs) error {
	s := server.NewMCPServer(
		"dodo-doc-mcp",
		Version,
		server.WithToolCapabilities(false),
	)
	h, err := NewHandler(NewEnvArgs(), args.endpoint)
	if err != nil {
		return fmt.Errorf("failed to create handler: %w", err)
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
		mcp.WithString("url", mcp.Required(), mcp.Description("The document URL")),
	)
	return tool
}

type Handler struct {
	Env      EnvArgs
	Endpoint Endpoint
}

func NewHandler(env EnvArgs, endpointStr string) (*Handler, error) {
	endpoint, err := NewEndpoint(endpointStr)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint: %w", err)
	}
	return &Handler{
		Env:      env,
		Endpoint: endpoint,
	}, nil
}

type addSearchHandlerResult struct {
	Items []openapi.SearchRecord `json:"items" jsonschema_description:"Search results"`
}

func (h *Handler) addSearchHandler(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query := request.GetString("query", "")
	items, err := sendSearchRequest(&h.Env, h.Endpoint, query)
	if err != nil {
		return nil, err
	}

	result := addSearchHandlerResult{
		Items: items,
	}
	return mcp.NewToolResultStructuredOnly(result), nil
}

func (h *Handler) addReadDocumentHandler(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	u := request.GetString("url", "")
	if u == "" {
		return nil, errors.New("url is required")
	}
	slug, path, err := parseURL(u)
	if err != nil {
		return nil, err
	}
	content, err := sendReadDocumentRequest(&h.Env, h.Endpoint, slug, path)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(content), nil
}
