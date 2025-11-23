---
title: read
path: command_read
description: "A document about read command"
created_at: "2025-09-06T17:04:46+09:00"
updated_at: "2025-09-06T17:04:46+09:00"
---

# `read` Command
This command retrieves and displays the raw markdown content of a doc from dodo-doc. It fetches the document from the server and outputs it to stdout, making it easy to pipe to other tools.

## Use Cases
* Fetch raw markdown to pass to AI agents or other processing tools
* Read documentation content from the command line without opening a browser

## Usage

```bash
dodo read [flags]
```

## Flags
* `-u, --url string`  
  The full URL of the document to read (overrides project-id and path if set)
* `-s, --project-id string`  
  The project ID (slug) to read the document from
* `-p, --path string`  
  The path of the document to read
* `--endpoint string`  
  Server endpoint for reading documents (default: "https://contents.dodo-doc.com/")
* `--debug`  
  Enable debug mode for detailed logging
* `--no-color`  
  Disable color output

## Examples

```bash
# Read a document using project ID and path
$ dodo-cli read --project-id my-project --path /docs/introduction.md
# Introduction

Welcome to my project...

# Read a document using full URL
$ dodo-cli read --url https://my-project.dodo-doc.com/docs/introduction.md
# Introduction

Welcome to my project...

# Read with debug mode enabled
$ dodo-cli read --project-id my-project --path /docs/guide.md --debug
```

## Requirements
* `DODO_API_KEY` environment variable must be set with a valid API key
* Either provide both `--project-id` and `--path`, or use the `--url` flag
* The document must exist and be accessible with the provided API key