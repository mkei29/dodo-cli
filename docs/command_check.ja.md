---
title: check
link: command_check_ja
description:
created_at: 2025-02-26T00:44:06+09:00
updated_at: 2025-02-26T00:44:06+09:00
---

# `check`コマンド

`check`コマンドは、dodo-docの設定ファイルを検証します。必要なフィールドがすべて存在し、正しくフォーマットされていることを確認し、ドキュメントをデプロイする前にエラーをキャッチします。

## ユースケース
* 編集後に.dodo.yaml設定を検証
* mainにマージする前にプルリクエストで.dodo.yamlの変更を検証するためにCIで実行

## 使い方

```bash
dodo check [flags]
```

## フラグ
* `-c, --config string`
  設定ファイルへのパス（デフォルトは「.dodo.yaml」）。必要に応じて別の設定ファイルを指定するには、このフラグを使用します。

* `--debug`
  デバッグモードを有効にします。トラブルシューティングに役立つ追加出力を提供します。

* `--no-color`
  カラー出力を無効にします。カラーテキストをサポートしない環境で便利です。

## エラーハンドリング

`check`コマンドは、設定ファイルが必要な基準を満たしていない場合、エラーを出力します。一般的なエラーには、フィールドの欠落や不正な日付フォーマットが含まれます。これらのエラーを避けるために、すべてのフィールドが期待されるフォーマットに従っていることを確認してください。

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
