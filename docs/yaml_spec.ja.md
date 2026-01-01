---
title: "Yaml仕様"
link: "yaml_spec_ja"
description: ""
created_at: "2025-09-18T23:26:46+09:00"
updated_at: "2025-09-18T23:26:46+09:00"
---

# 概要
このドキュメントでは、dodoが使用する`.dodo.yaml`設定ファイルの仕様を説明します。以下のサンプルを参照しながら説明します。

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
* **`annotation`** – 独自のワークフロー用のオプションの自由形式メタデータ

それぞれを詳しく見ていきましょう。

## Version

`version`フィールドは`.dodo.yaml`の仕様バージョンを示します。現在、`1`のみがサポートされています。

## Project

`project`セクションを使用して、ドキュメントの左上に表示される名前と説明を設定します。

```yaml
project:
  project_id: dododoc
  name: dodo-doc
  description: The dodo-doc documentation
  version: "v128"
  logo: "" # (オプション) プロジェクトロゴのパス。例：「assets/logo.png」
```

* **`project_id`** *(string, 必須)*: このドキュメントに関連付けられたプロジェクトID。ドキュメントはここで定義されたプロジェクトにアップロードされます。
* **`name`** *(string, 必須)*: ドキュメントの名前。この値はドキュメントのサイドバーに表示されます。
* **`description`** *(string, オプション)*: ドキュメントの説明。この値はドキュメントのサイドバーに表示されます。
* **`version`** *(string, オプション)*: ドキュメントバージョンを説明する文字列。この値はドキュメントのサイドバーに表示されます。省略すると、内部の連番が使用されます。
* **`logo`** *(string, オプション)*: ドキュメントロゴへのパス。この値がドキュメントのロゴとして使用されます。画像は28 x 28 pxを想定しています。


## Pages

`pages`セクションはドキュメントのコンテンツとレイアウトを構成します。ノードの配列を指定します。各ノードは`markdown`、`match`、または`directory`のいずれかです。

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
* **`title`** *(string, オプション)* : ドキュメントのタイトル。ここで省略された場合、dodo-docはドキュメントのフロントマターから値を使用します。
* **`path`** *(string, オプション)* : ドキュメントのURLパス。英数字のみが許可されます。省略された場合、dodo-docはフロントマターから値を使用します。
* **`description`** *(string, オプション)* : ドキュメントの説明。省略された場合、dodo-docはフロントマターから値を使用します。（管理ビューでの表示には影響しません。）
* **`updated_at`** *(string, オプション)* : ドキュメントの更新日。省略された場合、dodo-docはフロントマターから値を使用します。（管理ビューでの表示には影響しません。）

:::message info
### フォールバックの動作
.dodo.yamlで指定された値が優先されます。
フィールドが.dodo.yamlで指定されていない場合、dodo-docはMarkdownファイルのフロントマターから読み取ります。
最低限、titleとpathは.dodo.yamlまたはフロントマターのいずれかから解決可能でなければなりません。そうでない場合、アップロードは失敗します。
:::

### Matchノード

`match`フィールドを持つノードは**matchノード**です。matchノードを使用して、パターンに一致するすべてのMarkdownファイルを含めます。

* **`match`** *(string, 必須)*: 含めるMarkdownファイルのglobパターン。パターン構文はこの[Goライブラリ](https://pkg.go.dev/v.io/v23/glob)に従います。
* **`sort_key`** *("title", オプション)*: 一致したドキュメントをどのようにソートするか。現在、`"title"`のみがサポートされています。
* **`sort_order`** *("asc" | "desc", オプション)*: ソート順：昇順`"asc"`または降順`"desc"`。

matchノードを使用する場合、一致した各ドキュメントはフロントマターで独自のtitleとpathを宣言する必要があります：

```yaml
---
title: "What is dodo"
link: "what_is_dodo"
---
```

* **`title`** *(string, 必須)*: ドキュメントのタイトル。
* **`path`** *(string, 必須)*: アップロードされたドキュメントのURLパス。英数字のみが許可されます。

### Directoryノード

`directory`フィールドを持つノードは**directoryノード**です。directoryノードを使用して、サイドバー階層内の関連ドキュメントをグループ化します。

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

`annotation`を使用して、`project`や`pages`と並んで任意のメタデータを保存します。dodo-docはこのセクションをそのまま保持し、その内容を検証または使用しないため、ワークフロー（パイプラインフラグ、所有権、機能トグルなど）に合わせて形成できます。

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

dodo-docは処理中に`annotation`の内容を無視します。これは、任意のメタデータを添付するためだけに存在します。
