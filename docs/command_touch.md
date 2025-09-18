---
title: touch
path: command_touch
description: 
created_at: 2025-02-27T20:53:00+09:00
updated_at: 2025-02-27T20:53:00+09:00
---

# `touch` Command

The `touch` command helps manage markdown files. It creates a new markdown file with the given title if it does not exist, or updates the frontmatter fields if the file already exists. This command is useful for maintaining consistent metadata across your documentation.

## Usage

```bash
dodo touch [flags]
```

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
# Create a new markdown file including the front mater
$ dodo-cli touch example.md

# Create a new markdown file with metadata for the front matter
$ dodo-cli touch example.md --title "New Title" --path new-markdown
```
