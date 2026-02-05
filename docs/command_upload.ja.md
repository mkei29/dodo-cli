---
title: upload
link: command_upload_ja
description:
created_at: 2025-02-27T20:51:20+09:00
updated_at: 2025-02-27T20:51:20+09:00
---

# `upload`コマンド

`upload`コマンドはドキュメントをdodo-docに公開します。
`.dodo.yaml`を読み取り、定義されたMarkdownファイルをまとめてデプロイします。

## 使い方

```bash
dodo-cli upload [flags]
```

## ユースケース
* ドキュメントをdodo-docにデプロイ
* CIと連携して、mainへのマージ時にドキュメントを自動デプロイ

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
  アップロード先のエンドポイント（デフォルト: `http://api.dodo-doc.com/project/upload`）

* `--no-color`
  カラー出力を無効にします。


## 例

```bash
# ドキュメントをdodoにアップロード
$ dodo-cli upload
  • successfully uploaded
  • please open this link to view the document: https://xxx.do.dodo-doc.com
```
