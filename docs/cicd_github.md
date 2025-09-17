---
title: "CI/CD"
path: "cicd_github"
description: ""
created_at: "2025-09-16T00:51:21+09:00"
updated_at: "2025-09-16T00:51:21+09:00"
---

This page introduces GitHub Action workflow file templates that are ready to use.
Copy them to .github/workflows as needed.

## Prerequisite
You need to register the API Key as a GitHub Actions Secret in advance.
Please create a Secret named DODO_API_KEY by referring to the following article.

https://docs.github.com/en/actions/security-guides/using-secrets-in-github-actions

## Upload New Document
This workflow is a CI that automatically uploads documents when merged to the main branch.
Please use this when you want to automate document updates.

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
This workflow is a CI that automatically prepares a preview version of documents when a PR is created or updated.
Please use this when you want to check the document output during the review process.

```yaml
name: preview-document
on:
  pull_request:
    types: [opened, synchronize, reopened]
permissions:
  contents: write
  pull-requests: write
jobs: 
  publish-docs:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
      - name: Install dodo CLI
        run: |
          npm install -g @dodo-doc/cli
      - name: Preview docs
        id: preview
        continue-on-error: true
        run: |
          dodo preview --format json 1> stdout.txt 2> stderr.txt
      - name: Prepare comment body
        id: prepare-comment
        run: |
          STATUS=$(jq -r '.status' stdout.txt 2>/dev/null || echo "false")
          URL=$(jq -r '.document_url' stdout.txt 2>/dev/null || echo "")
      
          if [ "$STATUS" = "false" ] || [ -z "$URL" ]; then
            echo "failed"
            echo "status=failed" >> $GITHUB_OUTPUT

            # Build comment body
            echo "❌ Document preview failed" >> body.txt
            echo "<details><summary>Error Message</summary>" >> body.txt
            cat stderr.txt >> body.txt
            echo "</details>" >> body.txt

          else
            echo "success"
            echo "status=success" >> $GITHUB_OUTPUT

            # Build comment body
            echo "✅ Document preview is available [here]($URL)" >> body.txt
          fi
        env:
          DODO_API_KEY: ${{ secrets.DODO_ACCESS_KEY }}
      - name: Comment on PR
        uses: peter-evans/create-or-update-comment@71345be0265236311c031f5c7866368bd1eff043  # v4.0.0
        with:
          issue-number: ${{ github.event.pull_request.number }}
          body-path: body.txt
      - name: Fail if preview failed
        if: steps.prepare-comment.outputs.status == 'failed'
        run: exit 1
```
