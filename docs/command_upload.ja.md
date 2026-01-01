---
title: upload
link: command_upload_ja
description:
created_at: 2025-02-27T20:51:20+09:00
updated_at: 2025-02-27T20:51:20+09:00
---

# `upload`コマンド

`upload`コマンドは、ドキュメントをdodo-docに公開します。
`.dodo.yaml`を読み取り、参照されているすべてのMarkdownファイルをバンドルして、ドキュメントを読者がアクセスできるようにデプロイします。

## 使い方

```bash
dodo-cli upload [flags]
```

## ユースケース
* ドキュメントをdodo-docにデプロイ
* CI統合により、変更がmainにマージされたときにドキュメントを自動デプロイ

## フラグ

* `-c, --config string`
  設定ファイルへのパス（デフォルトは「.dodo.yaml」）。必要に応じて別の設定ファイルを指定するには、このフラグを使用します。

* `-w, --workingDir string`
  コマンドの実行コンテキストのプロジェクトルートパスを定義します（デフォルトは「.」）。異なるディレクトリにあるプロジェクトをアップロードする場合に便利です。

* `-f, --format string`
  出力フォーマット。「text」または「json」を指定できます。

* `--debug`
  デバッグモードを有効にします。トラブルシューティングのための追加出力を提供します。

* `-o, --output string`
  アーカイブファイルパス（非推奨）。

* `--endpoint string`
  アップロード先のエンドポイント（デフォルトは「http://api.dodo-doc.com/project/upload」）。必要に応じてカスタムアップロードエンドポイントを指定するには、このフラグを使用します。

* `--no-color`
  カラー出力を無効にします。カラーテキストをサポートしない環境で便利です。


## 例

```bash
# ドキュメントをdodoにアップロード
$ dodo-cli upload
  • successfully uploaded
  • please open this link to view the document: https://xxx.do.dodo-doc.com
```
