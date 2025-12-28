type: markdownによる単一のmarkdown指定
```yaml
  # 明示的に書くケース: locale省略. default_languageが使われる
  # (内部的には default_language の lang に格納される)
  - type: markdown
    link: "" # filepathのmarkdown内に記述があれば省略可能
    title: "" # filepathのmarkdown内に記述があれば省略可能
    description: "" # filepathのmarkdown内に記述があれば省略可能
    filepath: ""

  # 複数ロケールに対応する場合
  - type: markdown
    lang:
      en:
        link: ""
        title: "" # filepathのmarkdown内に記述があれば省略可能
        description: "" # filepathのmarkdown内に記述があれば省略可能
        filepath: ""
      ja:
        link: ""
        title: ""
        description: ""
        filepath: ""
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
  description: "" # 省略可能
  children:
    - type: markdown
      link: "cicd_github"
      title: "GitHub Actions"
      filepath: "./test.md"

# マルチロケール対応版
- type: directory
  lang:
    en:
      title: "English"
      description: "" # 省略可能
    ja:
      title: "Japanese"
      description: "" # 省略可能
  children:
    - type: markdown
      link: "cicd_github"
      title: "GitHub Actions"
      filepath: "./test.md"
```

`type: section` 殆どtype directoryと同じ
```yaml
# シングルロケール対応版
- type: section
  title: "test"
  description: "" # 省略可能
  children:
    - type: markdown
      link: "cicd_github"
      title: "GitHub Actions"
      filepath: "./test.md"

# マルチロケール対応版
- type: section
  lang:
    en:
      title: "English"
      description: "" # 省略可能
    ja:
      title: "Japanese"
      description: "" # 省略可能
  children:
    - type: markdown
      link: "cicd_github"
      title: "GitHub Actions"
      filepath: "./test.md"
```


## 実装状況

### 完了した機能
- ✅ Description機能（全てのページタイプでdescriptionフィールドをサポート）
- ✅ fillSingleLangFromMarkdownV2でtitle, link, descriptionを補完
- ✅ matchで言語毎のパスを指定（buildConfigPageFromMatchStatementV2で実装）
- ✅ validateConfigPageSection/DirectoryでDefaultLanguageの存在確認（validateLangKeySetV2経由）
- ✅ 各validationでvalidateLangKeySetV2を呼び出し

### 未完了・検討中の項目
* language_group_idの必須を緩和する（frontmatter仕様の変更が必要）