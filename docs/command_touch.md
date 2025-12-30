---
title: touch
link: command_touch
description: 
created_at: 2025-02-27T20:53:00+09:00
updated_at: 2025-02-27T20:53:00+09:00
---

# `touch` Command

The `touch` command helps manage markdown files. It creates a new markdown file with proper frontmatter if it doesn't exist, or updates the frontmatter fields if it already does. This ensures consistent metadata across your documentation.

## Usage

```bash
dodo touch [flags]
```

## Use Cases
* Quickly scaffold new documentation files with correct frontmatter structure
* Update timestamps and metadata for existing files without manual editing

## Flags
* `-t, --title string`  
  The title of the newly created file. This is used to set the `title` field in the frontmatter.

* `-p, --path string`  
  The URL path of the file. This sets the `path` field in the frontmatter, defining where the file will be accessible.

* `--debug`  
  Enable debug mode. Provides additional output for troubleshooting.

* `--no-color`  
  Disable color output. Useful for environments that do not support colored text.

* `--now string`  
  The current time in RFC3339 format. This can be used to set the `created_at` or `updated_at` fields in the frontmatter.

## Frontmatter Management

The `touch` command allows you to manage the frontmatter of markdown files, ensuring that metadata such as title, path, and timestamps are consistently applied.

## Examples

```bash
# Create a new markdown file with frontmatter
$ dodo-cli touch example.md

# Create a markdown file with custom metadata
$ dodo-cli touch example.md --title "New Title" --path new-markdown
```
