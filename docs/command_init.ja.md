---
title: "init"
link: "command_init_ja"
description:
created_at: "2025-02-27T21:09:00+09:00"
updated_at: "2025-02-27T21:09:00+09:00"
---

# `init`コマンド

`init`コマンドはプロジェクト用の設定ファイルを新規作成します。
フラグを省略すると対話形式で入力を求められるため、手軽にプロジェクトを初期化できます。

## 使い方

```bash
dodo init [flags]
```

## フラグ
* `-c, --config string`
  設定ファイルのパス

* `-w, --working-dir string`
  プロジェクトのルートパス。別のディレクトリで初期化したい場合に指定します。

* `-f, --force`
  既存の設定ファイルを上書きします。

* `--debug`
  デバッグモードを有効にします。

* `--project-name string`
  プロジェクト名を指定します。

* `--description string`
  プロジェクトの説明を指定します。

## 対話モード

オプションなしで実行すると、対話形式でプロジェクト情報の入力を求められます。

## 例

```bash
# 対話的にプロジェクトを作成
$ dodo-cli init
Project Name: My Project
Description: A sample project


# オプション付きでプロジェクトを作成
$ dodo-cli init --project-name "My Project" --description "A sample project"
```
