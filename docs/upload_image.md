---
title: "Upload Image"
link: "upload_image"
description: ""
created_at: "2025-09-20T00:18:32+09:00"
updated_at: "2025-09-20T00:18:32+09:00"
---

This page explains how to upload images for use in your documentation.

## Write the configuration

To upload images, add an `assets` section to your `.dodo.yaml`.
In this section, list the files you want to upload as an array.
You can also use glob patterns to include multiple files that match a pattern.

```yaml
version: 1
project:
  project_id: "document"
  name: dodo
  version: 0.0.1
pages:
  - markdown: docs/index.md
assets:
  - "assets/**"
```

The following file formats are currently supported by dodo-doc:

* image/jpeg
* image/png
* image/gif
* image/webp
* image/tiff
* image/bmp

## Write the Markdown

Once uploaded, your images can be referenced directly from Markdown.
Use the same paths you specified under `assets`.
When entering paths, always prefix them with / and use an absolute path.
In the example below, we display `/assets/preview.png`.

```markdown
![preview](/assets/preview.png)
```

After uploading, the image will render like this:

![preview](/assets/preview.png)
