---
title: preview
link: command_preview_ja
description:
created_at: 2025-05-25T16:35:00+09:00
updated_at: 2025-05-25T16:35:00+09:00
---

# `preview`コマンド

`preview`コマンドはドキュメントをdodo-docのプレビュー環境にアップロードします。本番デプロイ前の確認用に、期間限定のプレビューURLを生成します。

## 使い方

```bash
dodo preview [flags]
```

## ユースケース
* ローカルのMarkdownが正しく表示されるか確認
* CIでmainにマージする前にドキュメントの変更をチェック

## フラグ
* `-c, --config string`
  設定ファイルのパス（デフォルト: `.dodo.yaml`）

* `-w, --workingDir string`
  プロジェクトのルートパス（デフォルト: `.`）

* `-f, --format string`
  出力フォーマット（`text` または `json`）

* `--debug`
  デバッグモードを有効にします。

* `-o, --output string`
  アーカイブファイルパス（非推奨）

* `--endpoint string`
  アップロード先のエンドポイント（デフォルト: `https://api-demo.dodo-doc.com/project/upload`）

* `--no-color`
  カラー出力を無効にします。

## 例

```bash
# ドキュメントをdodoプレビュー環境にアップロード
$ dodo-cli preview
  • successfully uploaded
  • please open this link to view the document: https://xxx-preview.do.dodo-doc.com
```
