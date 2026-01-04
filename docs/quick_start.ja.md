---
title: "クイックスタート"
link: "quick_start_ja"
description: ""
created_at: "2025-09-15T23:48:10+09:00"
updated_at: "2025-09-15T23:48:10+09:00"
---
## ドキュメントをアップロードする手順

ドキュメントを公開するための通常の流れは以下の通りです：

1. dodo-docにサインアップしてプロジェクトを作成する
2. プロジェクトページでAPIキーを生成して環境変数にエクスポートする
3. `dodo init`を実行して設定ファイルを作成する
4. `dodo upload`を実行してドキュメントを公開する

## サインアップとプロジェクト作成

アップロードする前にアカウントとプロジェクトが必要です。まだ作成していない場合は、サインアップページにアクセスして手順を完了してください：

https://www.dodo-doc.com/signup

ダッシュボードから**New Project**（右上）をクリックして、ダイアログに必要事項を入力してください：

* **Visibility**: ドキュメントを誰が閲覧できるか
  * `public`: 誰でも閲覧可能
  * `private`: 組織のメンバーのみ閲覧可能
* **Project ID**: プロジェクトの一意な識別子。後の手順で使用します
* **Project Name**: プロジェクトのわかりやすい表示名

## 新しいAPIキーを発行

APIキーは、CLIを実行するユーザーがプロジェクトにアクセスする権限を持っていることを証明します。
**API Key**ページを開いて**New API Key**をクリックします。デフォルトでは、キーには**Read**と**Upload**の両方の権限が含まれます。

* **Read**: `docs`と`search`に必要
* **Upload**: `upload`と`preview`に必要

:::message warning
APIキーは一度だけ表示されます。画面を閉じた後は再度確認できません。
:::

:::message alert
APIキーを公開しないでください。漏洩すると、ドキュメントが改ざんされる可能性があります。
:::

次に、キーを環境変数としてエクスポートします：

```bash
export DODO_API_KEY="<YOUR_API_KEY>"
```

:::message info
ローカル環境から頻繁にアップロードする場合は、[direnv](https://direnv.net/)のようなツールが環境変数の管理に役立ちます。
:::

## 設定ファイルを作成

dodo-docは設定に`.dodo.yaml`ファイルを使用します。
対話的なヘルパーを実行して生成します：

```bash
dodo init
```

以下の項目の入力を求められます：

* **Project ID**: プロジェクト作成時に設定したID
* **Project Name**: ドキュメントページに表示される名前
* **Description**（オプション）: ドキュメントのサイドバーに表示される簡単な説明

`dodo init`を実行した後、`.dodo.yaml`が作成されたことを確認してください：

```yaml
version: 1
project:
  project_id: <Your Project ID>
  name: <Your Project Name>
  version: 1
  description: <Your project description>
pages:
  - markdown: README.md
    path: "README"
    title: "README"
  ## Create the directory and place all markdown files in the docs
  #- directory: "Directory"
  #  children:
  #    - match: "docs/*.md"
```

デフォルトでは、`README.md`がトップページになります。
必要に応じて、[設定仕様](/yaml_specification)を参照して`pages`セクションを調整してください。設定が適切になったら、アップロードに進みましょう。

## ドキュメントをアップロード

公開する準備ができました。以下を実行してください：

```bash
dodo upload
```

成功すると、`successfully uploaded`とドキュメントへのURLが表示されます。
ブラウザでリンクを開いて結果を確認してください。

再度アップロードするには、もう一度`dodo upload`を実行するだけです。

## 次のステップ

以上がアップロードの基本です。より高度な使い方については、以下を参照してください：

https://document.do.dodo-doc.com/yaml_specification

https://document.do.dodo-doc.com/cicd_github
