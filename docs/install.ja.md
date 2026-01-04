---
title: "Install"
link: "install_ja"
description: "CLIツールのセットアップ"
created_at: "2025-09-15T22:01:55+09:00"
updated_at: "2025-09-15T22:01:55+09:00"
---

## dodo CLIのインストール

**dodo-doc**を使うには、まずdodo CLIをインストールします。セットアップ方法は2つあります：npmでインストールするか、ビルド済みバイナリをダウンロードするかです。

:::message warning
Windowsは現在サポートされていません。
:::

### npmでインストール

ターミナルで以下のコマンドを実行してください：

```bash
npm install -g @dodo-doc/cli
```

### バイナリをダウンロード

または、リリースページからビルド済みバイナリをダウンロードして、`PATH`が通っている場所（例：`/usr/local/bin`）に配置します：

https://github.com/mkei29/dodo-cli/releases

### 動作確認
`dodo version`を実行して、インストールが成功したか確認してください。
他のコマンドにはAPIキーが必要です。詳細はクイックスタートを参照してください。

```bash
dodo version
```

## 次のステップ

インストールは以上です。次に、dodo-docで最初のドキュメントをアップロードしましょう。[クイックスタート](https://document.do.dodo-doc.com/quick_start)に従って始めてください。

https://document.do.dodo-doc.com/quick_start
