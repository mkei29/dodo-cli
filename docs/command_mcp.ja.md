---
title: mcp
link: command_mcp_ja
description: "mcpコマンドに関するドキュメント"
created_at: "2025-09-06T17:04:46+09:00"
updated_at: "2025-09-06T17:04:46+09:00"
---

# `mcp`コマンド
このコマンドはModel Context Protocol（MCP）サーバーを起動します。
stdio上で動作し、ドキュメントの検索・読み取り用の2つのツールを提供します。

## 利用可能なツール
* **search**: dodoプラットフォーム全体からドキュメントを検索します。タイトル、コンテンツ、ID、プロジェクト情報、URLを含む結果を返します。
* **read_document**: URLを指定してドキュメントのMarkdownコンテンツを取得します。

## 使い方

```bash
dodo mcp
```

## フラグ
* `--endpoint string`
  サーバーエンドポイント（デフォルト: `https://contents.dodo-doc.com/`）

## 例

```bash
# デフォルト設定でMCPサーバーを起動
$ dodo mcp
```

## Claude Codeとの統合
Claude CodeにMCPサーバーを追加するには、以下のコマンドを実行します：

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
* 環境変数`DODO_API_KEY`に有効なAPIキーを設定してください
* MCP対応クライアント（Claude Codeなど）が必要です
