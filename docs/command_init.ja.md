---
title: "init"
link: "command_init_ja"
description:
created_at: "2025-02-27T21:09:00+09:00"
updated_at: "2025-02-27T21:09:00+09:00"
---

# `init`コマンド

`init`コマンドは、プロジェクト用の新しい設定ファイルを作成します。
フラグで主要な詳細が提供されていない場合、対話的にプロンプトが表示され、必要な設定でプロジェクトを簡単に初期化できます。

## 使い方

```bash
dodo init [flags]
```

## フラグ
* `-c, --config string`
  設定ファイルへのパス。このフラグを使用して、カスタム設定ファイルパスを指定します。

* `-w, --working-dir string`
  コマンドの実行コンテキストのプロジェクトルートパスを定義します。異なるディレクトリでプロジェクトを初期化する場合に便利です。

* `-f, --force`
  設定ファイルが既に存在する場合は上書きします。既存の設定を失わないように注意して使用してください。

* `--debug`
  デバッグモードを有効にします。トラブルシューティングのための追加出力を提供します。

* `--project-name string`
  プロジェクト名。初期化するプロジェクトの名前を指定します。

* `--description string`
  プロジェクトの説明。プロジェクトの簡単な説明を提供します。

## 対話モード

オプションなしで実行すると、`init`コマンドは対話的にプロジェクトの詳細を要求します。ガイド付きのセットアッププロセスを好むユーザーに便利です。

## 例

```bash
# 対話的にプロジェクトを作成
$ dodo-cli init
Project Name: My Project
Description: A sample project


# オプション付きでプロジェクトを作成
$ dodo-cli init --project-name "My Project" --description "A sample project"
```
