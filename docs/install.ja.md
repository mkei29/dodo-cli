---
title: "Install"
link: "install_ja"
description: "CLIツールのセットアップ"
created_at: "2025-09-15T22:01:55+09:00"
updated_at: "2025-09-15T22:01:55+09:00"
---

## dodo CLIのインストール

**dodo-doc**を使うには、まずdodo CLIをインストールします。npmでインストールする方法と、ビルド済みバイナリをダウンロードする方法があります。

:::message warning
Windowsは現在サポートされていません。
:::

### npmでインストール

ターミナルで以下のコマンドを実行してください：

```bash
npm install -g @dodo-doc/cli
```

### バイナリをダウンロード

リリースページからビルド済みバイナリをダウンロードして、`PATH`の通ったディレクトリ（例: `/usr/local/bin`）に配置する方法もあります：

https://github.com/mkei29/dodo-cli/releases

### 動作確認
`dodo version`を実行して、インストールが成功したか確認してください。
他のコマンドにはAPIキーが必要です。詳細はクイックスタートを参照してください。

```bash
dodo version
```

## 次のステップ

インストールは以上です。次は[クイックスタート](https://document.do.dodo-doc.com/quick_start)に従って、最初のドキュメントをアップロードしてみましょう。

https://document.do.dodo-doc.com/quick_start
