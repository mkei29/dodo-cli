---
name: make-branch
description: Pull latest main and create a new branch with conventional commit-style name (feat/xxx, fix/xxx, etc.)
argument-hint: <branch-name>
allowed-tools: Bash(git fetch *), Bash(git switch *), Bash(git status)
---

以下の手順でブランチを作成する。

## Step 1: 作業中の変更を確認

```bash
git status
```

コミットされていない変更（modified・untracked）がある場合は、**作業を続けるかユーザーに確認してから先に進む**。

## Step 2: リモートを最新化

```bash
git fetch -a
```

## Step 3: ブランチを作成

`origin/main` を起点にブランチを作成する。

```bash
git switch -c $ARGUMENTS origin/main
```

## ブランチ命名規則

引数はconventional commitスタイルのプレフィックスを使うこと：

| プレフィックス | 用途 |
|:---|:---|
| `feat/xxx` | 新機能 |
| `fix/xxx` | バグ修正 |
| `chore/xxx` | ビルド・設定・依存関係など |
| `docs/xxx` | ドキュメント |
| `refactor/xxx` | リファクタリング |
| `test/xxx` | テスト追加・修正 |

`xxx` は短く英語で内容を表すケバブケース（例: `feat/add-user-auth`）。

## 引数が指定されていない場合

作業内容をユーザーに確認し、適切なブランチ名を提案してから作成する。
