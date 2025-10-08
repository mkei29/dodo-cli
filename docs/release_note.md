---
title: "Release Note"
path: "release_note"
description: ""
created_at: "2025-09-19T00:36:29+09:00"
updated_at: "2025-09-19T00:36:29+09:00"
---

# 2025/10/08 - version 0.1.1
* Released the beta version.
* Fixed the sanitizer bug for the `touch` command.
* Enhanced the MCP server prompt.

# 2025/09/07 - version 0.0.29
* Updated documents.
* Fixed the issue where error messages were being written to STDOUT.

# 2025/09/07 - version 0.0.28
* Applied hotfix to the `preview` command.

# 2025/09/06 - version 0.0.27
* Added the `read` command.
* Added the `read_document` tool for the `mcp` command.
* Refined the template used in the `init` command.
* Improved error handling.


# 2025/09/01 - version 0.0.26
* feat(experimental): add `mcp` subcommand.
* feat: add `--format` option for the `search` subcommand.
* fix: minor bugs.

# 2025/08/26
* fix: Fixed the issue where the `preview` command didn't work as expected.
* feat: Introduced the `--format` option for the `upload` and `preview` commands.
* feat: Prepared npm packages to make installation easier.

# 2025/06/07
* feat: Added the `preview` command to deploy preview documents.

# 2025/05/22
* fix: Fixed the issue where the `touch` command generated paths including slashes.

# 2025/03/25
* feat: Introduced a `search` command to search documents from the CLI.
* feat: Added the `project_id` field to the config file schema.

# 2025/02/28
* feat: Introduced a `check` command to validate the config file before uploading.
* feat: Added the `repository` field to describe the corresponding repository.

# 2025/02/24 - v0.0.8
* feat: Introduced a `touch` command to create a new markdown file with frontmatter.
* feat: Improved error logging for the `upload` command.

# 2024/12/11
* feat: Added the `asset` field to the `.dodo.yaml` schema.

# 2024/08/10
* feat: The document URL is now logged when running `dodo-cli upload`.
* docs: Added `YAML Spec` and `Roadmap` pages.
