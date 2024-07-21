
# Single Page Syntax

```yaml
path:
  # 最初のページはドキュメントのindexとなります。
  # 厳密にはindexにアクセスされた場合にはこのドキュメントにリダイレクトされます。
  - markdown: "article1.md"
    path: dir
    title: "First Article"
  # Markdown内のFront Matter内でpathやtitleを指定した場合には省略できます。
  # Markdown内にこれらの設定が記載されていない場合にはエラーになります。
  - markdown: "article2.md"
  - markdown: "article3.md"
```

```markdown
---
path: "test"
title: "Article"
---

Document Content Here
```

# Multi page syntax

```yaml
path:
  # match文に一致するmarkdownをすべて表示します。
  # どの順番で表示されるかはデフォルトではランダムです。
  - match: "*.md"
```

# Directory
ディレクトリを作成します。

```yaml
path:
  - directory: "Directory"
    children:
      - markdown: "article1.md"
      - markdown: "article2.md"
```