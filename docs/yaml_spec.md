---
title: "Yaml Spec"
path: "yaml_spec"
description: ""
created\_at: "2025-09-18T23:26:46+09:00"
updated\_at: "2025-09-18T23:26:46+09:00"
---

# Overview
This document describes the specification for the `.dodo.yaml` configuration file used by dodo. The sample below will be referenced throughout.

```yaml
version: 1
project:
  name: dodo
  description: The dodo documentation
pages:
  - markdown: docs/index.md
    path: "what_is_dodo"
    title: "What is dodo?"
  - markdown: docs/markdown_syntax.md
    path: "markdown"
    title: "Markdown Syntax"
  - directory: "Work with CI/CD"
    children:
      - markdown: docs/cicd_github.md
        path: "cicd_github"
        title: "GitHub Actions"
```

The YAML file has three top‑level sections:

* **`version`** – specifies the version of the `.dodo.yaml` spec.
* **`project`** – configures project‑level settings.
* **`pages`** – defines the structure and content of the documentation.

Let’s look at each in detail.

## Version

The `version` field indicates the `.dodo.yaml` spec version. Currently, only `1` is supported.

## Project

Use the `project` section to set the name and description shown at the top‑left of the document.

```yaml
project:
  name: dodo
  description: The dodo documentation
```

* **`name`** *(string, required)*: The document’s name. If omitted, the name set in the dodo dashboard is used.
* **`description`** *(string, required)*: The document’s description. If omitted, an empty string is used.

## Pages

The `pages` section configures the content and layout of your document. Provide an array of nodes. Each node is one of: `markdown`, `match`, or `directory`.

::: message info
The `pages` array must contain at least one element. When users visit the document root, they are redirected to the first node in `pages`.
:::

With the sample YAML, the hosted document will appear as:

```
|- "What is dodo?"       /what_is_dodo
|- "Markdown Syntax"     /markdown
|- "Work with CI/CD"
|  |- "GitHub Actions"   /cicd_github
```

::: message info
Document paths are **not** hierarchical; every path is placed directly under the root. If two documents share the same `path`, the upload will fail with an error.
:::

### Markdown node

Nodes with a `markdown` field are **markdown nodes**. Each represents a single document.

```yaml
- markdown: docs/index.md
  title: "What is dodo?"
  path: "what_is_dodo"
```

* **`markdown`** *(string, required)*: File path to the Markdown content.
* **`title`** *(string, optional)*: Document title. If omitted here, dodo uses the value from the document’s front matter.
* **`path`** *(string, optional)*: URL path for the document. Only alphanumeric characters are allowed. If omitted, dodo uses the value from front matter.
* **`description`** *(string, optional)*: Document description. If omitted, dodo uses the value from front matter. (Does not affect its appearance in the management view.)
* **`updated_at`** *(string, optional)*: Document update date. If omitted, dodo uses the value from front matter. (Does not affect its appearance in the management view.)

:::message info
### How fallback works
Values provided in .dodo.yaml take precedence.
If a field is not specified in .dodo.yaml, dodo reads it from the Markdown file’s front matter.
At minimum, title and path must be resolvable from either .dodo.yaml or front matter; otherwise the upload fails.
:::

### Match node

Nodes with a `match` field are **match nodes**. Use a match node to include all Markdown files that match a pattern.

* **`match`** *(string, required)*: Glob pattern for Markdown files to include. Pattern syntax follows this [Go library](https://pkg.go.dev/v.io/v23/glob).
* **`sort_key`** *("title", optional)*: How to sort the matched documents. Currently, only `"title"` is supported.
* **`sort_order`** *("asc" | "desc", optional)*: Sort order: ascending `"asc"` or descending `"desc"`.

When using match nodes, each matched document must declare its own title and path in front matter:

```yaml
---
title: "What is dodo"
path: "what_is_dodo"
---
```

* **`title`** *(string, required)*: The document title.
* **`path`** *(string, required)*: The URL path for the uploaded document. Only alphanumeric characters are allowed.

### Directory node

Nodes with a `directory` field are **directory nodes**. Use a directory node to group related documents in the sidebar hierarchy.

```yaml
- directory: "Work with CI/CD"
  children:
    - markdown: docs/cicd_github.md
      path: "cicd_github"
      title: "GitHub Actions"
```

* **`directory`** *(string, required)*: The directory label.
* **`children`** *(Node\[], required)*: Child nodes contained in the directory.
