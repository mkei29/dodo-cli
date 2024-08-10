
このページでは`.dodo.yaml`の仕様について説明します。
次に示すyamlファイルは、このページの説明で使用する`.dodo.yaml`のサンプルです。

```yaml
version: 1
project:
  name: dodo
  description: The dodo documentation
pages:
  - markdown: docs/index.md
    path: "what_is_dodo"
    title: "What is dodo?"
  - markdown: docs/markdown_syntax.md
    path: "markdown"
    title: "Markdown Syntax"
  - directory: "Work with CI/CD"
    children:
      - markdown: docs/cicd_github.md
        path: "cicd_github"
        title: "GitHub Actions"
```

YAMLファイルは大きく分けて3つのセクションで構成されています。

* `version`: `.dodo.yaml`の仕様のバージョンです。
* `project`: Projectの設定について記載するセクションです。
* `pages`: ドキュメントの構成を指定するセクションです。

それでは各セクションの詳細を見ていきましょう。

# version
この項目は`.dodo.yaml`の仕様のバージョンを表します。
現在`1`のみが指定できます。

dodoは可能な限り`.dodo.yaml`の後方互換性を維持しますが、将来よりよい体験を提供する為に新しい`.dodo.yaml`の仕様を提供する可能性があります。
将来、複数の仕様が提供されるようになった場合には、この項目でどのバージョンを利用するか指定することができます。

# project
このセクションの項目を書き換えることによって、ドキュメントの左上に表示されるドキュメントの名前や説明を設定することができます。

```yaml
project:
  name: dodo
  description: The dodo documentation
```

* `name` (string, Required): ドキュメントの名称です。ドキュメント上に表示されるドキュメント名はこの値になります。値が指定されなかった場合には、dodoダッシュボードで指定されたドキュメント名が使用されます。
* `description` (string, Required): ドキュメントの説明です。ドキュメント上に表示されるドキュメント名はこの値になります。値が指定されなかった場合には空白文字列が使用されます。

# pages
このセクションの項目を書き換えることによって、ドキュメントのコンテンツやレイアウトを設定することができます。
`pages`セクションには配列で`markdown`、`match`、`directory`形式のノードを指定することができます。

`pages`セクションは最低でも1つ以上の要素を持つ必要があります。
ユーザーがドキュメントのルートにアクセスした場合には、`pages`で指定された最初のノードにリダイレクトされます。

```yaml
pages:
  - markdown: docs/index.md
    path: "what_is_dodo"
    title: "What is dodo?"
  - markdown: docs/markdown_syntax.md
    path: "markdown"
    title: "Markdown Syntax"
  - directory: "Work with CI/CD"
    children:
      - markdown: docs/cicd_github.md
        path: "cicd_github"
        title: "GitHub Actions"
```

このyamlファイルを元にdodoへドキュメントをアップロードすると、以下のレイアウトを持つドキュメントがホストされます。

```
|- "What is dodo?" /what_is_dodo
|- "Markdown Syntax" /markdown
|- "Work with CI/CD"
|  |- "GitHub Actions" "/cicd_github"
```

重要な点として各ドキュメントのパスはディレクトリなどを利用しても階層構造を形成せず、ルート直下に配置されます。
そのため同じ`path`を持つドキュメントが存在した場合にはアップロード時にエラーがでます。

### `markdown` Node
`markdown`エントリを含むノードは`markdown`ノードと見做されます。
`markdown`ノードは単一のドキュメントを表します。

```yaml
- markdown: docs/index.md
  title: "What is dodo?"
  path: "what_is_dodo"
```

* `markdown` (string, Required): ドキュメントのコンテンツとなるMarkdownへのファイルパス。
* `title` (string, Required): ドキュメントのタイトル。
* `path` (string, Required): アップロードしたドキュメントのURLのパス。英数字のみ指定することができます。
* `description` (string, Optional): ドキュメントの説明文です。管理用のエントリでドキュメントの見た目は変化しません。
* `updated_at` (string, Optional): ドキュメントが更新された日付です。管理用のエントリでドキュメントの見た目は変化しません。

### `match` Node
`match`エントリを含むノードは`match`ノードと見做されます。
`match`ノードを使うことで、パターンに一致するmarkdownをまとめてレイアウトに追加することができます。


* `match`: 追加したいマークダウンのパターン。パターンの仕様はgolangの[このライブラリ](https://pkg.go.dev/v.io/v23/glob)に基づいています。
* `sort_key` "title: どのような順番でドキュメントを並べるか指定できます。現時点ではタイトルでソートする"title"のみが利用可能です。
* `sort_order` "desc" | "asc":  降順でソートするか昇順でソートするかを指定できます　

このノードを利用する際には、ドキュメント冒頭に明示的にタイトルとパスの情報を記載する必要があります。

```yaml
---
title: "What is dodo"
path: "what_is_dodo"
---
```

* `title` string: ドキュメントのタイトル。
* `path` string: アップロードしたドキュメントのURLのパス。英数字のみ指定することができます。

### `directory` Node
`directory`エントリを含むノードは`directory`ノードと見做されます。
`directory`ノードを設定することでドキュメントの改装構造を表現することができます。

```yaml
- directory: "Work with CI/CD"
  children:
    - markdown: docs/cicd_github.md
      path: "cicd_github"
      title: "GitHub Actions"
```

* `directory` string: ディレクトリの名前
* `children` Node[]:　ディレクトリの子ノード
