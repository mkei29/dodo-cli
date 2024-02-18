---
title: How to use dodo
path: /index
description: This page describe how to setup dodo.
---

# What is dodo

`dodo` is a documentation web service for modern developers.
dodoは現代の開発者のためのドキュメントツールです。

これまで優秀なドキュメントツールは多々ありましたが、多くのツールではホストするたの環境を自前で整える必要があります。
dodoはpublicに公開したい場合にもプライベートで使いたい場合にも数分でセットアップできます。

* シンプルなyamlファイルで設定を記述可能
* publicに公開することも、privateで一部のユーザーだけで利用することもできる
* デプロイ操作がCLIで完結して開発者フレンドリー
* 各種CIへの組み込みも簡単


## Install

First of all, let's install dodo to your machine. 

```bash
curl url
```

### For mac user

### For linux user

### For windows user

## Create First Project
次にdodoダッシュボードで新しくプロジェクトを作成しましょう。
Then, go to the dodo dashboard.


## How to write `.dodo.yaml`
dodoは`.dodo.yaml`ファイルが配置されているディレクトリをドキュメント用のディレクトリとして認識します。


```yaml
pages:
  - path: "./index.md"
  - path: "./index.md"
```

pagesを指定した場合には, pagesに列挙した順番に従ってレイアウトが構成されます。
pagesを指定した上で列挙されなかったファイルはレイアウトに追加されません。
同じファイルを複数指定した場合にはエラーが発生します。


## Client Logics
Call /upload_archive
This endpoint returns the id for uploaded archive

