---
title: preview
path: command_preview
description: 
created_at: 2025-05-25T16:35:00+09:00
updated_at: 2025-05-25T16:35:00+09:00
---

# `preview` Command

The `preview` command is used to upload the project to dodo-doc's preview environment. It is an alias of the `upload` command with a different default endpoint. This command facilitates the transfer of your project's documentation to the dodo-doc preview platform for testing purposes before deploying to production.

## Flags
* `-c, --config string`  
  Path to the configuration file (default is ".dodo.yaml"). Use this flag to specify a different configuration file if needed.

* `-w, --workingDir string`  
  Defines the root path of the project for the command's execution context (default is "."). This is useful for uploading projects located in different directories.

* `--debug`  
  Enable debug mode. Provides additional output for troubleshooting.

* `-o, --output string`  
  Archive file path (Deprecated). 

* `--endpoint string`  
  Endpoint to upload (default is "https://api-demo.dodo-doc.com/project/upload"). Use this flag to specify a custom upload endpoint if needed.

* `--no-color`  
  Disable color output. Useful for environments that do not support colored text.

## Examples

```bash
# Upload the document to dodo preview environment.
$ dodo-cli preview
  • successfully uploaded
  • please open this link to view the document: https://xxx-preview.do.dodo-doc.com
