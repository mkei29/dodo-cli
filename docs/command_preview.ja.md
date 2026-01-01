---
title: preview
link: command_preview_ja
description:
created_at: 2025-05-25T16:35:00+09:00
updated_at: 2025-05-25T16:35:00+09:00
---

# `preview`コマンド

`preview`コマンドは、プロジェクトをdodo-docのプレビュー環境にアップロードします。これは、本番デプロイ前にテストするための一時的で共有可能なバージョンのドキュメントです。`upload`と同じように動作しますが、異なるエンドポイントを対象とし、期間限定のプレビューURLを生成します。

## 使い方

```bash
dodo preview [flags]
```

## ユースケース
* ローカルのMarkdownドキュメントが正しくレンダリングされることを確認
* mainにマージする前にドキュメントの変更を検証するためにCIで実行

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
  アップロード先のエンドポイント（デフォルトは「https://api-demo.dodo-doc.com/project/upload」）。必要に応じてカスタムアップロードエンドポイントを指定するには、このフラグを使用します。

* `--no-color`
  カラー出力を無効にします。カラーテキストをサポートしない環境で便利です。

## 例

```bash
# ドキュメントをdodoプレビュー環境にアップロード
$ dodo-cli preview
  • successfully uploaded
  • please open this link to view the document: https://xxx-preview.do.dodo-doc.com
```
