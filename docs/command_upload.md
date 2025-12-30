---
title: upload
link: command_upload
description: 
created_at: 2025-02-27T20:51:20+09:00
updated_at: 2025-02-27T20:51:20+09:00
---

# `upload` Command

The `upload` command publishes your documentation to dodo-doc.
It reads your `.dodo.yaml`, bundles all referenced Markdown files, and deploys them to make your docs accessible to readers.

## Usage

```bash
dodo-cli upload [flags]
```

## Use Cases
* Deploy your documentation to dodo-doc
* Automatically deploy documentation when changes are merged to main via CI integration

## Flags

* `-c, --config string`  
  Path to the configuration file (default is ".dodo.yaml"). Use this flag to specify a different configuration file if needed.

* `-w, --workingDir string`  
  Defines the root path of the project for the command's execution context (default is "."). This is useful for uploading projects located in different directories.

* `-f, --format string`
  Output format. Accepts either "text" or "json".

* `--debug`  
  Enable debug mode. Provides additional output for troubleshooting.

* `-o, --output string`  
  Archive file path (Deprecated). 

* `--endpoint string`  
  Endpoint to upload (default is "http://api.dodo-doc.com/project/upload"). Use this flag to specify a custom upload endpoint if needed.

* `--no-color`  
  Disable color output. Useful for environments that do not support colored text.


## Examples

```bash
# Upload the document to dodo.
$ dodo-cli upload
  • successfully uploaded
  • please open this link to view the document: https://xxx.do.dodo-doc.com
```
