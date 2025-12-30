---
title: mcp
link: command_mcp
description: "A document about mcp command"
created_at: "2025-09-06T17:04:46+09:00"
updated_at: "2025-09-06T17:04:46+09:00"
---

# `mcp` Command
This command starts a Model Context Protocol (MCP) server that provides tools for interacting with dodo-doc documents.
The server runs over stdio and exposes two tools for searching and reading documents.

## Available Tools
* **search**: Search documents across the entire dodo platform based on a query. Returns structured results with document titles, contents, IDs, project information, and URLs.
* **read_document**: Read the full markdown content of a specific document using its URL.

## Usage

```bash
dodo mcp
```

## Flags
* `--endpoint string`  
  Server endpoint for document operations (default: "https://contents.dodo-doc.com/")

## Examples

```bash
# Start the MCP server with default settings
$ dodo mcp
```

## Integration with Claude Code
To install the MCP server in Claude Code, use the following command:

```bash
$ claude mcp add dodo --env DODO_API_KEY=<YOUR_API_KEY> -- dodo mcp
```

## Tool Details

### Search Tool
- **Input**: Query string
- **Output**: JSON array with document search results containing:
  - `title`: Document title
  - `contents`: Document contents preview
  - `id`: Document ID
  - `project_id`: Project ID
  - `project_slug`: Project slug
  - `url`: Document URL

### Read Document Tool
- **Input**: Document URL (obtained from search results)
- **Output**: Full markdown content of the document

## Requirements
* `DODO_API_KEY` environment variable must be set with a valid API key
* MCP-compatible client (such as Claude Code) to interact with the server
