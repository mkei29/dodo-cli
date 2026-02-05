---
title: "CI/CD"
link: "cicd_github_ja"
description: ""
created_at: "2025-09-16T00:51:21+09:00"
updated_at: "2025-09-16T00:51:21+09:00"
---

このページではGitHub Actionsのワークフローテンプレートを紹介します。
`.github/workflows`にコピーしてそのまま使えます。

## 前提条件
APIキーをGitHub Actions Secretとして事前に登録してください。
以下を参考に`DODO_API_KEY`という名前のSecretを作成します。

https://docs.github.com/ja/actions/security-guides/using-secrets-in-github-actions

## ドキュメントのアップロード
mainブランチへのマージ時にドキュメントを自動アップロードするワークフローです。

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
        uses: actions/checkout@08c6903cd8c0fde910a37f88322edcfb5dd907a8  # v5.0.0
        with:
          lfs: true
          fetch-depth : 0
      - name: Install dodo CLI
        run: |
          npm install -g @dodo-doc/cli
      - name: Publish docs
        run: |
          dodo upload
        env:
          DODO_API_KEY: ${{ secrets.DODO_API_KEY }}
```

## PRでドキュメントをプレビュー
PRの作成・更新時にドキュメントのプレビューを自動生成するワークフローです。
レビュー中にドキュメントの表示を確認できます。

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
        with:
          lfs: true
          fetch-depth : 0
      - name: Install dodo CLI
        run: |
          npm install -g @dodo-doc/cli
      - name: Preview docs
        id: preview
        continue-on-error: true
        run: |
          dodo preview --format json 1> stdout.txt 2> stderr.txt
        env:
          DODO_API_KEY: ${{ secrets.DODO_API_KEY }}
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
            echo "<pre><code>" >> body.txt
            cat stderr.txt >> body.txt
            echo "</code></pre>" >> body.txt
            echo "</details>" >> body.txt

          else
            echo "success"
            echo "status=success" >> $GITHUB_OUTPUT

            # Build comment body
            echo "✅ Document preview is available [here]($URL)" >> body.txt
          fi
      - name: Comment on PR
        uses: peter-evans/create-or-update-comment@71345be0265236311c031f5c7866368bd1eff043 # v4.0.0
        with:
          issue-number: ${{ github.event.pull_request.number }}
          body-path: body.txt
      - name: Fail if preview failed
        if: steps.prepare-comment.outputs.status == 'failed'
        run: exit 1
```
