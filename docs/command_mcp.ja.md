---
title: mcp
link: command_mcp_ja
description: "mcpコマンドに関するドキュメント"
created_at: "2025-09-06T17:04:46+09:00"
updated_at: "2025-09-06T17:04:46+09:00"
---

# `mcp`コマンド
このコマンドは、dodo-docドキュメントと対話するためのツールを提供するModel Context Protocol（MCP）サーバーを起動します。
サーバーはstdio上で実行され、ドキュメントの検索と読み取りのための2つのツールを公開します。

## 利用可能なツール
* **search**: クエリに基づいてdodoプラットフォーム全体のドキュメントを検索します。ドキュメントのタイトル、コンテンツ、ID、プロジェクト情報、URLを含む構造化された結果を返します。
* **read_document**: URLを使用して特定のドキュメントの完全なMarkdownコンテンツを読み取ります。

## 使い方

```bash
dodo mcp
```

## フラグ
* `--endpoint string`
  ドキュメント操作用のサーバーエンドポイント（デフォルト：「https://contents.dodo-doc.com/」）

## 例

```bash
# デフォルト設定でMCPサーバーを起動
$ dodo mcp
```

## Claude Codeとの統合
Claude CodeにMCPサーバーをインストールするには、以下のコマンドを使用します：

```bash
$ claude mcp add dodo --env DODO_API_KEY=<YOUR_API_KEY> -- dodo mcp
```

## ツールの詳細

### Searchツール
- **入力**: クエリ文字列
- **出力**: 以下を含むドキュメント検索結果のJSON配列：
  - `title`: ドキュメントのタイトル
  - `contents`: ドキュメントコンテンツのプレビュー
  - `id`: ドキュメントID
  - `project_id`: プロジェクトID
  - `project_slug`: プロジェクトスラッグ
  - `url`: ドキュメントURL

### Read Documentツール
- **入力**: ドキュメントURL（検索結果から取得）
- **出力**: ドキュメントの完全なMarkdownコンテンツ

## 要件
* 有効なAPIキーを含む`DODO_API_KEY`環境変数を設定する必要があります
* サーバーと対話するためのMCP互換クライアント（Claude Codeなど）
