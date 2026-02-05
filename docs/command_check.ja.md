---
title: check
link: command_check_ja
description:
created_at: 2025-02-26T00:44:06+09:00
updated_at: 2025-02-26T00:44:06+09:00
---

# `check`コマンド

`check`コマンドは`.dodo.yaml`設定ファイルを検証します。必須フィールドの有無やフォーマットをチェックし、デプロイ前にエラーを検出できます。

## ユースケース
* 編集後の`.dodo.yaml`の検証
* CIでmainにマージする前に`.dodo.yaml`の変更をチェック

## 使い方

```bash
dodo check [flags]
```

## フラグ
* `-c, --config string`
  設定ファイルのパス（デフォルト: `.dodo.yaml`）

* `--debug`
  デバッグモードを有効にします。トラブルシューティング用の詳細な出力が表示されます。

* `--no-color`
  カラー出力を無効にします。

## エラーハンドリング

設定ファイルに問題がある場合、`check`コマンドはエラーを出力します。よくあるエラーとしては、必須フィールドの欠落や日付フォーマットの誤りがあります。

## 例

```bash
$ dodo-cli check
  ⨯ .dodo.yaml:10:12 the `title` field should exist in the markdown file when you use `match`: /xxx/usage1.md
    >     - match: "/xxx/*"
  ⨯ .dodo.yaml:10:12 the `path` field should exist in the markdown file when you use `match`: /xxx/usage2.md
    >     - match: "/xxx/*"
  ⨯ .dodo.yaml:10:12 the `title` field should exist in the markdown file when you use `match`: /xxx/usage3.md
    >     - match: "/xxx/*"

...

  ⨯ .dodo.yaml:20:12 `created_at` should follow the RFC3339 format. Got: 20241113: /yyy/20241113.md
    >     - match: "/yyy/*"
  ⨯ .dodo.yaml:20:12 `created_at` should follow the RFC3339 format. Got: 20240818: /yyy/20240818.md
    >     - match: "/yyy/*"
  ⨯ .dodo.yaml:20:12 `created_at` should follow the RFC3339 format. Got: 20240617: /yyy/20240617.md
    >     - match: "/yyy/*"
Error: 79 errors:
```
