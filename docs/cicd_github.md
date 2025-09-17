---
title: "CI/CD"
path: ""
description: ""
created_at: ""
updated_at: "2025-09-16T00:51:21+09:00"
---

このページにはGithub Actionで使えるWorkflowファイルのテンプレートを紹介しています。
必要に応じて.github/workflowsなどにコピーして利用してください。

## Prerequisite

## Upload New Document
This workflow

```yaml
name: upload-document
on:
  push:
    branches:
      - main
  workflow_dispatch:
permissions:
  contents: read
jobs: 
  publish-docs:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
      - name: Install dodo CLI
        run: |
          npm install -g @dodo-cli/cli
      - name: Publish docs
        run: |
          dodo upload
        env:
          DODO_API_KEY: ${{ secrets.DODO_API_KEY }}
```

## Preview Document on PR 

## Check .dodo.yaml Syntax

```
```

