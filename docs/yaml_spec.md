# Overview
This document outlines the specifications of the `.dodo.yaml` configuration file used in dodo. Below is a sample `.dodo.yaml` file that will be referenced throughout this guide.

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

The YAML file is divided into three main sections:

* **`version`**: Specifies the version of the `.dodo.yaml` specification.
* **`project`**: Configures project-specific settings.
* **`pages`**: Defines the structure and content of the documentation.

Let's explore each section in detail.

## Version
The `version` field indicates the version of the `.dodo.yaml` specification. Currently, only `1` is supported.

> **Note**: While dodo aims to maintain backward compatibility, future updates may introduce new specifications. You can specify the desired version in this field when multiple versions are available.

## Project
The `project` section allows you to set the name and description displayed at the top left of the document.

```yaml
project:
  name: dodo
  description: The dodo documentation
```

* **`name`** (string, Required): The document's name. If not specified, the name set on the dodo dashboard will be used.
* **`description`** (string, Required): The document's description. If not specified, an empty string will be used.

## Pages
The `pages` section configures the content and layout of the document. You can specify nodes in the format of `markdown`, `match`, or `directory` within an array.

::: message info
> The `pages` section must contain at least one element. When users access the root of the document, they are redirected to the first node specified in `pages`.
:::

With this YAML configuration, the hosted document will have the following layout:

```
|- "What is dodo?" /what_is_dodo
|- "Markdown Syntax" /markdown
|- "Work with CI/CD"
|  |- "GitHub Actions" "/cicd_github"
```

::: message info

Document paths do not form a hierarchical structure; they are placed directly under the root. If documents share the same `path`, an error will occur during upload.
:::

### Markdown Node
Nodes containing `markdown` entries are considered as `markdown` nodes.
A `markdown` node represents a single document.

```yaml
- markdown: docs/index.md
  title: "What is dodo?"
  path: "what_is_dodo"
```

* **`markdown`** (string, Required): File path to the Markdown content.
* **`title`** (string, Required): Document title.
* **`path`** (string, Required): URL path for the document. Only alphanumeric characters are allowed.
* **`description`** (string, Optional): Document description. Does not affect appearance in the management entry.
* **`updated_at`** (string, Optional): Document update date. Does not affect appearance in the management entry.

### Match Node
Nodes containing a `match` entry are considered as `match` nodes.
By using a `match` node, you can add markdown that matches a pattern to the layout.

* `match` (string, Required): Pattern of the markdown to be added. The pattern specification is based on this [library](https://pkg.go.dev/v.io/v23/glob) in Go.
* `sort_key` ("title", Optional): You can specify how to sort the documents. Currently, only "title" for sorting by title is available.
* `sort_order` ("asc" | "desc", Optional): You can specify whether to sort in descending "desc" or ascending "asc" order.

When using this node, it is necessary to explicitly include the title and path information at the beginning of the document.

```yaml
---
title: "What is dodo"
path: "what_is_dodo"
---
```

* `title` (string, Required): Title of the document.
* `path` (string, Required): Path of the uploaded document's URL. Only alphanumeric characters can be specified.

### Directory Node
Nodes containing a `directory` entry are considered as `directory` nodes.
By setting a `directory` node, you can represent the structural layout of the document.

```yaml
- directory: "Work with CI/CD"
  children:
    - markdown: docs/cicd_github.md
      path: "cicd_github"
      title: "GitHub Actions"
```

* `directory` (string, Required): Name of the directory.
* `children` (Node[], Required): Children nodes of the directory.
