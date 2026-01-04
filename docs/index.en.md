---
title: "What is dodo-doc"
description: "Docs in sync with your code"
link: "what_is_dodo_doc"
created_at: "2025-09-15T19:49:41+09:00"
updated_at: "2025-09-15T19:49:41+09:00"
---

Keeping documentation accurate is hard — teams move fast, pages go stale, and readers tune out. 
dodo-doc fixes that by keeping docs in step with your code and easy to read.

[dodo-doc](https://www.dodo-doc.com) is a lightweight, powerful documentation hosting service built for developers. 
Create an account, install a single CLI, and publish Markdown in minutes. 

Browse live examples—including this page—to see what’s possible.
Ready to try it? Follow the link below to deploy your first docs.

https://document.do.dodo-doc.com/install

:::message info
dodo-doc is currently in beta and offers only a Free plan.
A paid Team Plan for teams is planned for a future release.
:::

# What dodo-doc can do

dodo-doc is a documentation hosting service for developers.
It's designed to fit naturally into your development workflow, making it easy to keep documentation up to date.

## Advantages for document readers
* **Beautiful, modern interface**: Clean design that makes documentation easy to navigate and pleasant to read
* **Terminal-first access**: Open and search docs with simple commands—no need to leave your terminal
* **Privacy controls**: Keep docs private with flexible visibility settings—great for personal notes or sensitive content

## Advantages for document writers
* **Developer-friendly**: Drop Markdown files into your workflow and upload—no special formats required
* **Easy setup**: Run `dodo init` to scaffold a .dodo.yaml config in seconds
* **CI/CD integration**: Ship docs from your pipeline with a single-binary CLI for automated, reliable deployments

dodo-doc is built for people who live in editors and terminals. Here are a few CLI commands you’ll use often:

```bash
# Initialize config: interactively create .dodo.yaml
dodo init

# Upload documents defined in .dodo.yaml
dodo upload

# Publish a time-limited preview for quick reviews
dodo preview

# List your documents and open one in the browser
dodo docs

# Search your documents and open a result
dodo search
```

We’re actively shipping new features to make dodo-doc even better. Check the [roadmap](/roadmap) to see what’s coming.


## Other Options

### Other Documentation Hosting Services
Other documentation hosting services such as [GitBook](https://www.gitbook.com/) and [Mintlify](https://www.mintlify.com/) also exist.
If you want enterprise-grade reliability or advanced AI features, these may be a better fit.

### Static File Hosting Services
You can generate HTML files for documents using tools like [mdBook](https://rust-lang.github.io/mdBook/) or [Fumadocs](https://fumadocs.dev/) and upload them to static hosting services.
There are various services such as Cloudflare Pages and GitHub Pages.
With GitHub Pages, you can also host private docs visible only to users within your organization.

## Next Steps
Ready to dive in? Create your dodo-doc account and host your first document.
Use the guides below to set up the client and get started:

- [Installation Guide](https://document.do.dodo-doc.com/install)
- [Quick Start Guide](https://document.do.dodo-doc.com/quick_start)
