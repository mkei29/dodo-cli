type: markdownによる単一のmarkdown指定
```yaml
  # 明示的に書くケース: locale省略. default_languageが使われる
  # (内部的には default_language の lang に格納される)
  - type: markdown
    link: "" # filepathのmarkdown内に記述があれば省略可能
    title: "" # filepathのmarkdown内に記述があれば省略可能
    filepath: ""

  # 複数ロケールに対応する場合
  - type: markdown
    # defaultLangの指定が
    lang:
      en:
        link: ""
        filepath: ""
        title: "" # filepathのmarkdown内に記述があれば省略可能
      ja:
        link: "" # localesを使う場合には省略不可. filepathにlinkが書かれていても無視してwarningを出す
        filepath: ""
        title: ""
```

matchによるパターンマッチ.
同一のlinkがあった場合にはlangフィールドを読み取りlangに重複が無いか + defaultLangに対応する言語が存在するかチェックする.
チェックが通れば複数ロケールの`type: markdown`と同様に処理する
```yaml
  - type: match
    pattern: "" # globパターン
    sort_key: "title"
    sort_order: "asc"
```

`type: directory` あまりv1と変わらないが以下の通り複数ロケールの対応ができる。
```yaml
# 単一ロケール版 (内部的には default_language の lang に格納される)
- type: directory
  title: "English"
  children:
    - markdown: docs/cicd_github.md
      path: "cicd_github"
      title: "GitHub Actions"

# マルチロケール対応版
- type: directory
  lang:
    en:
      title: "English"
    ja:
      title: "Japanese"
  children:
    - markdown: docs/cicd_github.md
      path: "cicd_github"
      title: "GitHub Actions"
```

`type: section` 殆どtype directoryと同じ
```yaml
  - type: section
      en:
        title: ""
        path
    path: "test"
      en: 
        title: ""
        path: docs/index.md
      ja: 
        title: ""
        path: docs/index.ja.md
```


* Pathの説明がおかしそう
* PageにDescription相当の機能がなさそう
* fillLangFieldsFromMarkdownV2でtitle以外の補完を指定なさそう
* matchで言語毎のパスを指定できなさそう
* language_group_idは必須じゃなくしたい。

matchEntryV2のような中途半端な構造体を作らないでください。FrontmatterはLangを持つ可能性があります。