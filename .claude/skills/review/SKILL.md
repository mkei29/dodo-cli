---
name: review
description: Review a PR for purpose, test coverage, implementation quality, env consistency, and formatting
argument-hint: <pr-number>
disable-model-invocation: true
---

あなたはこのリポジトリの厳格なコードレビュアーです。以下のステップに従ってPRをレビューしてください。

## Step 1: PRの目的を把握する

```bash
gh pr view $ARGUMENTS --json title,body,commits,files
```

- PRのタイトル・説明からの変更目的を整理する
- 変更されたファイル一覧を確認し、変更範囲を把握する
- コミット履歴から実装の流れを確認する

## Step 2: テストカバレッジの確認

変更内容に応じて、以下を確認する。

**フロントエンド（`services/frontend/`, `services/contents_frontend/`, `services/ui_component/`）**:
- 新規ロジック・コンポーネントに対するテストが存在するか
- 既存テストが変更に追従して更新されているか

**バックエンド（`services/backend/`, `services/contents_backend/`, `services/core/`）**:
- 新規エンドポイント・関数に対するGoテストが存在するか
- エッジケース・エラーケースがカバーされているか

**Edge（`services/edge/`）**:
- Workerのロジック変更にテストが存在するか

テストが不十分な箇所は具体的に指摘すること。ロジックの変更がない純粋な設定変更はテスト不要と判断してよい。

## Step 3: 実装品質・環境差分の確認

**実装の洗練度**:
- 冗長な処理・デッドコードがないか
- 同じ処理が複数箇所に重複していないか
- 過剰な抽象化や不要なレイヤーが追加されていないか
- エラーハンドリングが適切か（不要なケースへのバリデーションが過剰でないか）

**local/production 環境差分**:
- `wrangler.jsonc` のenv設定でlocal/productionが適切に分離されているか
- `k8s/overlays/` のlocal・staging・productionで設定の意図しない差分がないか
- 環境変数・シークレットがハードコードされていないか
- R2バケット名・ドメイン等の環境固有値が各envに正しく設定されているか

## Step 4: フォーマット確認

```bash
make fmt
```

実行してエラーや差分が出ないことを確認する。差分が出た場合は該当ファイルと内容を報告する。

---

## レビュー結果の出力形式

以下のフォーマットでレビュー結果をまとめること：

```
## PR #<number> レビュー: <title>

### 目的
<PRの変更目的を1〜3文で要約>

### ✅ 問題なし
- <問題なかった点を箇条書き>

### ⚠️ 指摘事項
- **[テスト]** <指摘内容>
- **[実装]** <指摘内容>
- **[環境差分]** <指摘内容>
- **[フォーマット]** <指摘内容>

### 総評
<LGTM / 要修正 / 要議論 のいずれかと、その理由>
```

指摘事項がない場合は「⚠️ 指摘事項」セクションを省略してよい。
