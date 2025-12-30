---
title: check
link: command_check
description: 
created_at: 2025-02-26T00:44:06+09:00
updated_at: 2025-02-26T00:44:06+09:00
---

# `check` Command

The `check` command validates your dodo-doc configuration file. It ensures all required fields are present and correctly formatted, catching errors before you deploy your documentation.

## Use Cases
* Validate your .dodo.yaml configuration after editing
* Run in CI to validate .dodo.yaml changes in pull requests before merging to main

## Usage

```bash
dodo check [flags]
```

## Flags
* `-c, --config string`  
  Path to the configuration file (default is ".dodo.yaml"). Use this flag to specify a different configuration file if needed.

* `--debug`  
  Enable debug mode. This provides additional output useful for troubleshooting.

* `--no-color`  
  Disable color output. Useful for environments that do not support colored text.

## Error Handling

The `check` command will output errors if the configuration file does not meet the required standards. Common errors include missing fields and incorrect date formats. Ensure that all fields follow the expected format to avoid these errors.

## Examples

```bash
$ dodo-cli check
  ⨯ .dodo.yaml:10:12 the `title` field should exist in the markdown file when you use `match`: /xxx/usage1.md
    >     - match: "/xxx/*"
  ⨯ .dodo.yaml:10:12 the `path` field should exist in the markdown file when you use `match`: /xxx/usage2.md
    >     - match: "/xxx/*"
  ⨯ .dodo.yaml:10:12 the `title` field should exist in the markdown file when you use `match`: /xxx/usage3.md
    >     - match: "/xxx/*"

...

  ⨯ .dodo.yaml:20:12 `created_at` should follow the RFC3339 format. Got: 20241113: /yyy/20241113.md
    >     - match: "/yyy/*"
  ⨯ .dodo.yaml:20:12 `created_at` should follow the RFC3339 format. Got: 20240818: /yyy/20240818.md
    >     - match: "/yyy/*"
  ⨯ .dodo.yaml:20:12 `created_at` should follow the RFC3339 format. Got: 20240617: /yyy/20240617.md
    >     - match: "/yyy/*"
Error: 79 errors: 
```
