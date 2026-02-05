---
title: touch
link: command_touch_ja
description:
created_at: 2025-02-27T20:53:00+09:00
updated_at: 2025-02-27T20:53:00+09:00
---

# `touch`コマンド

`touch`コマンドはMarkdownファイルの作成・更新を行います。ファイルが存在しない場合はフロントマター付きで新規作成し、既に存在する場合はフロントマターを更新します。

## 使い方

```bash
dodo touch [flags]
```

## ユースケース
* フロントマター付きのドキュメントファイルを素早く作成
* 既存ファイルのタイムスタンプやメタデータを更新

## フラグ
* `-t, --title string`
  フロントマターの`title`フィールドに設定する値

* `-p, --path string`
  フロントマターの`path`フィールドに設定するURLパス

* `--debug`
  デバッグモードを有効にします。

* `--no-color`
  カラー出力を無効にします。

* `--now string`
  RFC3339形式の現在時刻。フロントマターの`created_at`や`updated_at`に設定されます。

## 例

```bash
# フロントマター付きの新しいMarkdownファイルを作成
$ dodo-cli touch example.md

# カスタムメタデータ付きのMarkdownファイルを作成
$ dodo-cli touch example.md --title "New Title" --path new-markdown
```
