---
title: "init"
path: "command_init"
description: 
created_at: "2025-02-27T21:09:00+09:00"
updated_at: "2025-02-27T21:09:00+09:00"
---

# `init` Command

The `init` command creates a new configuration file for your project.
If key details aren’t provided via flags, it will prompt you interactively, making it easy to bootstrap a project with the required settings.

## Usage

```bash
dodo init [flags]
```

## Flags
* `-c, --config string`  
  Path to the configuration file. Use this flag to specify a custom configuration file path.

* `-w, --working-dir string`  
  Defines the root path of the project for the command's execution context. This is useful for initializing projects in different directories.

* `-f, --force`  
  Overwrite the configuration file if it already exists. Use with caution to avoid losing existing configurations.

* `--debug`  
  Enable debug mode. Provides additional output for troubleshooting.

* `--project-name string`  
  Project Name. Specify the name of the project to be initialized.

* `--description string`  
  Project Description. Provide a brief description of the project.

## Interactive Mode

When run without options, the `init` command will prompt for project details interactively. This is useful for users who prefer a guided setup process.

## Examples

```bash
# Create a project interactively
$ dodo-cli init
Project Name: My Project
Description: A sample project


# Create a project with options
$ dodo-cli init --project-name "My Project" --description "A sample project"
```
