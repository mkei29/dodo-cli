---
title: read
link: command_read_ja
description: "readコマンドに関するドキュメント"
created_at: "2025-09-06T17:04:46+09:00"
updated_at: "2025-09-06T17:04:46+09:00"
---

# `read`コマンド
このコマンドは、dodo-docからドキュメントの生のMarkdownコンテンツを取得して表示します。サーバーからドキュメントを取得し、標準出力に出力するため、他のツールにパイプすることが簡単にできます。

## ユースケース
* AIエージェントや他の処理ツールに渡すために生のMarkdownを取得
* ブラウザを開かずにコマンドラインからドキュメントコンテンツを読む

## 使い方

```bash
dodo read [flags]
```

## フラグ
* `-u, --url string`
  読み取るドキュメントの完全なURL（設定されている場合、project-idとpathを上書きします）
* `-s, --project-id string`
  ドキュメントを読み取るプロジェクトID（スラッグ）
* `-p, --path string`
  読み取るドキュメントのパス
* `--endpoint string`
  ドキュメント読み取り用のサーバーエンドポイント（デフォルト：「https://contents.dodo-doc.com/」）
* `--debug`
  詳細なロギングのためのデバッグモードを有効にします
* `--no-color`
  カラー出力を無効にします

## 例

```bash
# プロジェクトIDとパスを使用してドキュメントを読み取る
$ dodo-cli read --project-id my-project --path /docs/introduction.md
# Introduction

Welcome to my project...

# 完全なURLを使用してドキュメントを読み取る
$ dodo-cli read --url https://my-project.dodo-doc.com/docs/introduction.md
# Introduction

Welcome to my project...

# デバッグモードを有効にして読み取る
$ dodo-cli read --project-id my-project --path /docs/guide.md --debug
```

## 要件
* 有効なAPIキーを含む`DODO_API_KEY`環境変数を設定する必要があります
* `--project-id`と`--path`の両方を指定するか、`--url`フラグを使用してください
* ドキュメントが存在し、提供されたAPIキーでアクセス可能である必要があります
