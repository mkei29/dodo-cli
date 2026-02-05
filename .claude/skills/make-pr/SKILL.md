---
name: make-pr
description: "現在のブランチの変更内容からPRを作成します。コミット履歴と差分を分析し、このリポジトリの規約に沿ったタイトルとサマリーを生成します。ドキュメント変更がある場合は dodo check を事前に実行します。"
argument-hint: "[base-branch]"
allowed-tools:
  - Bash(git:*)
  - Bash(gh pr:*)
  - Bash(dodo check)
---

現在のブランチからPRを作成する。引数でベースブランチを指定できる（デフォルト: `main`）。

ベースブランチ: $ARGUMENTS（未指定なら `main`）

## 手順

### 1. 事前チェック

- `git status` で未コミットの変更がないか確認する。未コミットの変更がある場合はユーザーに報告し、コミットするか確認する。
- `git log <base>..HEAD --oneline` でこのブランチのコミット一覧を取得する。
- `git diff <base>...HEAD --stat` で変更ファイルの統計を取得する。
- `git diff <base>...HEAD` で差分の詳細を取得する。
- リモートブランチが最新か確認する。push が必要なら push する。

### 2. ドキュメントの検証

`docs/` 配下または `.dodo.yaml` に変更がある場合：
1. `dodo check` を実行して設定の整合性を確認する
2. エラーがあればユーザーに報告し、修正を促す

### 3. PRタイトルとサマリーの生成

コミット履歴と差分を分析して、以下のルールでPRを作成する。

**タイトル:**
- このリポジトリのコミットメッセージ規約に従う（`feat:`, `fix:`, `docs:` 等のプレフィックス）
- 70文字以内
- 変更の本質を簡潔に表現する

**ボディ:**
```
## Summary
<変更内容を1〜3行の箇条書きで>

## Test plan
<テスト方法を箇条書きで>
```

### 4. PR作成

```bash
gh pr create --title "<title>" --body "$(cat <<'EOF'
## Summary
<bullets>

## Test plan
<bullets>
EOF
)"
```

- PR作成後、URLをユーザーに表示する。

## 注意事項

- すべてのコミット（最新だけでなく）を分析してタイトルとサマリーを作成する
- `--force` や `--no-verify` は使わない
- main/masterへの force push は絶対にしない
- PRのタイトルは英語で書く
