---
title: "Markdown Syntax"
link: "markdown_syntax"
description: ""
created_at: "2025-09-18T23:25:21+09:00"
updated_at: "2025-09-18T23:25:21+09:00"
---

This page introduces the Markdown syntax you can use in dodo-doc.
dodo-doc supports most of the syntax defined by [CommonMark](https://commonmark.org/).

## Headings

```markdown
# Heading 1
## Heading 2
### Heading 3
#### Heading 4
##### Heading 5
###### Heading 6
```

## Italic

```markdown
This is *italic* text.

This is _italic_ text.
```

This is *italic* text.

This is *italic* text.

## Bold

```markdown
This is **bold** text.

This is __bold__ text.
```

This is **bold** text.

This is **bold** text.

## Inline code

```markdown
`Inline code` example
```

`Inline code` example

## Inline image

```markdown
![preview](assets/preview.png)
```

![preview](assets/preview.png)

## Fenced code blocks

Use triple backticks (\`\`\`) and an optional language hint.

````markdown
```bash
echo "Hello from bash"
```
````

```bash
echo "Hello from bash"
```

## Blockquotes

```markdown
> Blockquote text
```

> Blockquote text

## Ordered lists

```markdown
1. Item 1
2. Item 2
3. Item 3
4. Item 4
```

1. Item 1
2. Item 2
3. Item 3
4. Item 4

## Unordered lists

```markdown
* Item 1
* Item 2
- Item 3
- Item 4
```

* Item 1
* Item 2

- Item 3
- Item 4

## Links

```markdown
[dodo top](https://www.dodo-doc.com)
```

[dodo top](https://www.dodo-doc.com)

## Thematic breaks (horizontal rules)

```markdown
---
***
___
```

---

# dodo-doc–specific syntax

The following features are dodo-doc extensions to standard Markdown.

## Link cards

If a URL appears on its own line, separated by blank lines, dodo-doc renders it as a link card.

```markdown

https://www.dodo-doc.com

```

[https://www.dodo-doc.com](https://www.dodo-doc.com)

## Messages

Use message blocks to call attention to information. Supported types are `info`, `warning`, and `alert`.

```markdown
:::message info
Info message
:::

:::message warning
Warning message
:::

:::message alert
Alert message
:::
```
