---
title: check
path: command_check
description: 
created_at: 2025-02-26T00:44:06+09:00
updated_at: 2025-02-26T00:44:06+09:00
---

# `check` Command

The `check` command is used to validate the configuration file for dodo-doc.

## Flags
* `-c, --config string`
  Path to the configuration file (default is ".dodo.yaml").

* `--debug`
  Enable debug mode.

* `--no-color`
  Disable color output.

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

