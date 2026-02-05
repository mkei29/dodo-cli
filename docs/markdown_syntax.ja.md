---
title: "Markdown構文"
link: "markdown_syntax_ja"
description: ""
created_at: "2025-09-18T23:25:21+09:00"
updated_at: "2025-09-18T23:25:21+09:00"
---

このページでは、dodo-docで使用できるMarkdown構文を紹介します。
dodo-docは[CommonMark](https://commonmark.org/)で定義されているほとんどの構文をサポートしています。

## 見出し

```markdown
# 見出し 1
## 見出し 2
### 見出し 3
#### 見出し 4
##### 見出し 5
###### 見出し 6
```

## イタリック

```markdown
これは*イタリック*のテキストです。

これは_イタリック_のテキストです。
```

これは*イタリック*のテキストです。

これは*イタリック*のテキストです。

## 太字

```markdown
これは**太字**のテキストです。

これは__太字__のテキストです。
```

これは**太字**のテキストです。

これは**太字**のテキストです。

## インラインコード

```markdown
`インラインコード`の例
```

`インラインコード`の例

## インライン画像

```markdown
![preview](assets/preview.png)
```

![preview](assets/preview.png)

## コードブロック

3つのバッククォート（\`\`\`）と、オプションで言語名を指定します。

````markdown
```bash
echo "Hello from bash"
```
````

```bash
echo "Hello from bash"
```

## 引用

```markdown
> 引用テキスト
```

> 引用テキスト

## 順序付きリスト

```markdown
1. アイテム 1
2. アイテム 2
3. アイテム 3
4. アイテム 4
```

1. アイテム 1
2. アイテム 2
3. アイテム 3
4. アイテム 4

## 順序なしリスト

```markdown
* アイテム 1
* アイテム 2
- アイテム 3
- アイテム 4
```

* アイテム 1
* アイテム 2

- アイテム 3
- アイテム 4

## リンク

```markdown
[dodo top](https://www.dodo-doc.com)
```

[dodo top](https://www.dodo-doc.com)

## 水平線

```markdown
---
***
___
```

---

# dodo-doc固有の構文

以下の機能は、標準Markdownに対するdodo-doc拡張です。

## リンクカード

URLが前後を空行で区切られた独立した行にある場合、dodo-docはリンクカードとして表示します。

```markdown

https://www.dodo-doc.com

```

[https://www.dodo-doc.com](https://www.dodo-doc.com)

## メッセージ

メッセージブロックで重要な情報を目立たせることができます。`info`、`warning`、`alert`の3種類があります。

```markdown
:::message info
情報メッセージ
:::

:::message warning
警告メッセージ
:::

:::message alert
アラートメッセージ
:::
```
