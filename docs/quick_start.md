---
title: "Quick Start"
link: "quick_start"
description: ""
created_at: "2025-09-15T23:48:10+09:00"
updated_at: "2025-09-15T23:48:10+09:00"
---
## Steps to Upload a Document

Here’s the usual flow for publishing a document:

1. Sign up for dodo-doc and create a project.
2. Generate an API key on the project page and export it as an environment variable.
3. Run `dodo init` to create a configuration file.
4. Run `dodo upload` to publish your documents.

## Sign Up and Create a Project

You’ll need an account and a project before uploading. If you haven’t done this yet, visit the sign-up page and complete the steps:

https://www.dodo-doc.com/signup

From the dashboard, click **New Project** (top right) and fill in the dialog:

* **Visibility**: Who can view the document.
  * `public`: visible to anyone
  * `private`: visible only to members of your organization
* **Project ID**: A unique identifier for the project. You’ll use this in later steps.
* **Project Name**: A human-friendly name for the project.

## Issue a New API Key

The API key proves that the user running the CLI has permission to access the project.
Open the **API Key** page and click **New API Key**. By default, the key includes both **Read** and **Upload** permissions.

* **Read**: Required for `docs` and `search`.
* **Upload**: Required for `upload` and `preview`.

:::message warning
Your API key is shown only once. You won’t be able to view it again after closing the screen.
:::

:::message alert
Never share your API key publicly. If it is leaked, your documents could be tampered with.
:::

Then, export the key as an environment variable:

```bash
export DODO_API_KEY="<YOUR_API_KEY>"
```

:::message info
If you upload frequently from a local environment, tools like [direnv](https://direnv.net/) can help manage environment variables.
:::

## Create a Configuration File

dodo-doc uses a `.dodo.yaml` file for configuration.
Run the interactive helper to generate it:

```bash
dodo init
```

You’ll be prompted for:

* **Project ID**: The ID you set when creating the project.
* **Project Name**: The name shown on the document page.
* **Description** (optional): A short description shown in the document sidebar.

After running `dodo init`, confirm that `.dodo.yaml` was created:

```yaml
version: 1
project:
  project_id: <Your Project ID>
  name: <Your Project Name>
  version: 1
  description: <Your project description>
pages:
  - markdown: README.md
    path: "README"
    title: "README"
  ## Create the directory and place all markdown files in the docs
  #- directory: "Directory"
  #  children:
  #    - match: "docs/*.md"
```

By default, `README.md` is the top page.
If needed, see the [configuration spec](/yaml_specification) to adjust the `pages` section. When your config looks good, move on to uploading.

## Upload Documents

You’re ready to publish. Run:

```bash
dodo upload
```

On success, you’ll see `successfully uploaded` and a URL to the document.
Open the link in your browser to check the result.

To upload again, simply run `dodo upload` once more.

## Next Steps

That’s the basics of uploading. For more advanced usage, see:

https://document.do.dodo-doc.com/yaml_specification

https://document.do.dodo-doc.com/cicd_github
