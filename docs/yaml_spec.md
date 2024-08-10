This page explains the specifications of `.dodo.yaml`.
Below is a sample `.dodo.yaml` file that will be used in the explanations on this page.

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

The YAML file consists of three main sections.

* `version`: Represents the version of the `.dodo.yaml` specification.
* `project`: Section for configuring the project.
* `pages`: Section specifying the structure of the documentation.

Let's now delve into the details of each section.

# Version
This item represents the version of the `.dodo.yaml` specification.
Currently, only `1` can be specified.

While dodo strives to maintain backward compatibility with `.dodo.yaml`, there is a possibility of providing a new specification for a better experience in the future.
In case multiple specifications are provided in the future, you can specify which version to use in this item.

# Project
By modifying the items in this section, you can set the name and description that will be displayed at the top left of the document.

```yaml
project:
  name: dodo
  description: The dodo documentation
```

* `name` (string, Required): Name of the document. The name displayed on the document will be this value. If not specified, the document name set on the dodo dashboard will be used.
* `description` (string, Required): Description of the document. The description displayed on the document will be this value. If not specified, an empty string will be used.

# Pages
By modifying the items in this section, you can configure the content and layout of the document.
In the `pages` section, you can specify nodes in the format of `markdown`, `match`, or `directory` in an array.

The `pages` section must have at least one element.
When a user accesses the root of the document, they will be redirected to the first node specified in `pages`.

With this YAML file, when uploading a document to dodo, the hosted document will have the following layout:

```
|- "What is dodo?" /what_is_dodo
|- "Markdown Syntax" /markdown
|- "Work with CI/CD"
|  |- "GitHub Actions" "/cicd_github"
```

An important point is that the path of each document will not form a hierarchical structure but will be placed directly under the root, even if directories are used.
Therefore, if there are documents with the same `path`, an error will occur during upload.


### Markdown Node
Nodes containing `markdown` entries are considered as `markdown` nodes.
A `markdown` node represents a single document.

```yaml
- markdown: docs/index.md
  title: "What is dodo?"
  path: "what_is_dodo"
```

* `markdown` (string, Required): File path to the Markdown content of the document.
* `title` (string, Required): Title of the document.
* `path` (string, Required): Path of the uploaded document's URL. Only alphanumeric characters can be specified.
* `description` (string, Optional): Description of the document. The appearance of the document does not change in the management entry.
* `updated_at` (string, Optional): Date when the document was updated. The appearance of the document does not change in the management entry.

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
