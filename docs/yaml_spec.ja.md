---
title: "YAML仕様"
link: "yaml_spec_ja"
description: ""
created_at: "2025-09-18T23:26:46+09:00"
updated_at: "2025-09-18T23:26:46+09:00"
---

# 概要
このページでは`.dodo.yaml`設定ファイルの仕様を説明します。以下のサンプルを例に解説します。

```yaml
version: 1
project:
  name: dodo-doc
  description: The dodo-doc documentation
pages:
  - markdown: docs/index.md
    path: "what_is_dodo_doc"
    title: "What is dodo-doc?"
  - markdown: docs/markdown_syntax.md
    path: "markdown"
    title: "Markdown Syntax"
  - directory: "Work with CI/CD"
    children:
      - markdown: docs/cicd_github.md
        path: "cicd_github"
        title: "GitHub Actions"
annotation:
  owner: docs-team
  tags:
    - internal
    - handbook
```

YAMLファイルには4つのトップレベルセクションがあります：

* **`version`** – `.dodo.yaml`の仕様バージョンを指定します
* **`project`** – プロジェクトレベルの設定を構成します
* **`pages`** – ドキュメントの構造とコンテンツを定義します
* **`annotation`** – 任意のメタデータを付与するためのオプションセクション

以下、各セクションについて説明します。

## Version

`version`フィールドは`.dodo.yaml`の仕様バージョンを示します。現在、`1`のみがサポートされています。

## Project

`project`セクションで、ドキュメントの左上に表示される名前と説明を設定します。

```yaml
project:
  project_id: dododoc
  name: dodo-doc
  description: The dodo-doc documentation
  version: "v128"
  logo: "" # (オプション) プロジェクトロゴのパス。例：「assets/logo.png」
```

* **`project_id`** *(string, 必須)*: アップロード先のプロジェクトID。
* **`name`** *(string, 必須)*: ドキュメントの名前。サイドバーに表示されます。
* **`description`** *(string, オプション)*: ドキュメントの説明。サイドバーに表示されます。
* **`version`** *(string, オプション)*: ドキュメントバージョンの文字列。サイドバーに表示されます。省略すると内部の連番が使われます。
* **`logo`** *(string, オプション)*: ドキュメントロゴへのパス。画像は28 x 28 pxを想定しています。


## Pages

`pages`セクションでドキュメントの構成とレイアウトを定義します。各要素（ノード）は`markdown`、`match`、`directory`のいずれかです。

::: message info
`pages`配列には少なくとも1つの要素が必要です。ユーザーがドキュメントルートにアクセスすると、`pages`の最初のノードにリダイレクトされます。
:::

サンプルYAMLの場合、ホストされたドキュメントは次のように表示されます：

```
|- "What is dodo?"       /what_is_dodo
|- "Markdown Syntax"     /markdown
|- "Work with CI/CD"
|  |- "GitHub Actions"   /cicd_github
```

::: message info
ドキュメントのパスは**階層的ではありません**。すべてのパスはルート直下に配置されます。2つのドキュメントが同じ`path`を共有している場合、アップロードはエラーで失敗します。
:::

### Markdownノード

`markdown`フィールドを持つノードは**markdownノード**です。それぞれが単一のドキュメントを表します。

```yaml
- markdown: docs/index.md
  title: "What is dodo doc?"
  path: "what_is_dodo_doc"
```

* **`markdown`** *(string, 必須)* : Markdownコンテンツへのファイルパス。
* **`title`** *(string, オプション)* : ドキュメントのタイトル。省略時はフロントマターの値が使われます。
* **`path`** *(string, オプション)* : ドキュメントのURLパス。英数字のみ。省略時はフロントマターの値が使われます。
* **`description`** *(string, オプション)* : ドキュメントの説明。省略時はフロントマターの値が使われます（管理ビューには影響しません）。
* **`updated_at`** *(string, オプション)* : ドキュメントの更新日。省略時はフロントマターの値が使われます（管理ビューには影響しません）。

:::message info
### フォールバックの動作
.dodo.yamlで指定された値が優先されます。
フィールドが.dodo.yamlで指定されていない場合、dodo-docはMarkdownファイルのフロントマターから読み取ります。
最低限、titleとpathは.dodo.yamlまたはフロントマターのいずれかから解決可能でなければなりません。そうでない場合、アップロードは失敗します。
:::

### Matchノード

`match`フィールドを持つノードは**matchノード**です。パターンに一致するすべてのMarkdownファイルをまとめて含められます。

* **`match`** *(string, 必須)*: 含めるMarkdownファイルのglobパターン。パターン構文はこの[Goライブラリ](https://pkg.go.dev/v.io/v23/glob)に従います。
* **`sort_key`** *("title", オプション)*: 一致したドキュメントをどのようにソートするか。現在、`"title"`のみがサポートされています。
* **`sort_order`** *("asc" | "desc", オプション)*: ソート順：昇順`"asc"`または降順`"desc"`。

matchノードでは、一致した各ドキュメントがフロントマターでtitleとpathを宣言する必要があります：

```yaml
---
title: "What is dodo"
link: "what_is_dodo"
---
```

* **`title`** *(string, 必須)*: ドキュメントのタイトル。
* **`path`** *(string, 必須)*: アップロードされたドキュメントのURLパス。英数字のみが許可されます。

### Directoryノード

`directory`フィールドを持つノードは**directoryノード**です。サイドバー上で関連ドキュメントをグループ化できます。

```yaml
- directory: "Work with CI/CD"
  children:
    - markdown: docs/cicd_github.md
      path: "cicd_github"
      title: "GitHub Actions"
```

* **`directory`** *(string, 必須)*: ディレクトリラベル。
* **`children`** *(Node\[], 必須)*: ディレクトリに含まれる子ノード。

## Annotation

`annotation`には任意のメタデータを記述できます。dodo-docはこのセクションの内容を検証・使用しないため、パイプラインのフラグや所有者情報など、自由な用途に使えます。

```yaml
annotation:
  owner: docs-team
  tags:
    - internal
    - handbook
  feature_flags:
    search: enabled
    ai_summary: disabled
```

dodo-docは`annotation`の内容を処理時に無視します。メタデータの保存用途としてのみ存在します。
