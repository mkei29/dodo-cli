
# dodo-cli

The CLI tool for dodo-doc.  
dodo-doc is a web service that provides document hosting.

## Document Guidelines

All documents exist under the `docs` directory.  
After you add or update a document, please also update the `.dodo.yaml` at the project root.  
Finally, you must run `dodo check` to verify the configuration is valid.

When you write or modify documents, you need to follow the rules below.

### Rules for Command Documentation

You must write a summary of the document at the top like the following:

```markdown
# `init` command
This command creates a new `.dodo.yaml` at the current directory.
...
```

If there are options or flags that are valuable for dodo users, please add a `Flags` section:

```markdown
## Flags
* `-c, --config string`  
  Path to the configuration file (default is ".dodo.yaml"). Use this flag to specify a different configuration file if needed.
```

After that, you must add an `Examples` section that describes typical use cases:

```markdown
## Examples

```bash
# Upload the document to dodo preview environment.
$ dodo-cli preview
  • successfully uploaded
  • please open this link to view the document: https://xxx-preview.do.dodo-doc.com
```
```

If there is additional information that you should notify users about, please write additional sections after the examples section.