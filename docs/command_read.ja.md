---
title: read
link: command_read_ja
description: "readコマンドに関するドキュメント"
created_at: "2025-09-06T17:04:46+09:00"
updated_at: "2025-09-06T17:04:46+09:00"
---

# `read`コマンド
このコマンドはdodo-docからドキュメントのMarkdownコンテンツを取得し、標準出力に出力します。パイプで他のツールに渡すこともできます。

## ユースケース
* AIエージェントや他のツールにMarkdownを渡す
* ブラウザを開かずにコマンドラインでドキュメントを確認

## 使い方

```bash
dodo read [flags]
```

## フラグ
* `-u, --url string`
  ドキュメントのURL。指定すると`--project-id`と`--path`より優先されます。
* `-s, --project-id string`
  プロジェクトID（スラッグ）
* `-p, --path string`
  ドキュメントのパス
* `--endpoint string`
  サーバーエンドポイント（デフォルト: `https://contents.dodo-doc.com/`）
* `--debug`
  デバッグモードを有効にします。
* `--no-color`
  カラー出力を無効にします。

## 例

```bash
# プロジェクトIDとパスを指定してドキュメントを読み取る
$ dodo-cli read --project-id my-project --path /docs/introduction.md
# Introduction

Welcome to my project...

# URLを指定してドキュメントを読み取る
$ dodo-cli read --url https://my-project.dodo-doc.com/docs/introduction.md
# Introduction

Welcome to my project...

# デバッグモードを有効にして読み取る
$ dodo-cli read --project-id my-project --path /docs/guide.md --debug
```

## 要件
* 環境変数`DODO_API_KEY`に有効なAPIキーを設定してください
* `--project-id`と`--path`の両方を指定するか、`--url`を指定してください
* 指定したドキュメントが存在し、APIキーでアクセス可能である必要があります
