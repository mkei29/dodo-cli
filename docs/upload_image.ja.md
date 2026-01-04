---
title: "画像のアップロード"
link: "upload_image_ja"
description: ""
created_at: "2025-09-20T00:18:32+09:00"
updated_at: "2025-09-20T00:18:32+09:00"
---

このページでは、ドキュメントで使用する画像をアップロードする方法を説明します。

## 設定を書く

画像をアップロードするには、`.dodo.yaml`に`assets`セクションを追加します。
このセクションには、アップロードしたいファイルを配列として列挙します。
globパターンを使用して、パターンに一致する複数のファイルを含めることもできます。

```yaml
version: 1
project:
  project_id: "document"
  name: dodo
  version: 0.0.1
pages:
  - markdown: docs/index.md
assets:
  - "assets/**"
```

dodo-docは現在、以下のファイル形式をサポートしています：

* image/jpeg
* image/png
* image/gif
* image/webp
* image/tiff
* image/bmp

## Markdownを書く

アップロードすると、画像をMarkdownから直接参照できます。
`assets`で指定したのと同じパスを使用してください。
パスを入力する際は、先頭に / を付けた絶対パスを使用します。
以下の例では、`/assets/preview.png`を表示しています。

```markdown
![preview](/assets/preview.png)
```

アップロード後、画像は次のようにレンダリングされます：

![preview](/assets/preview.png)
