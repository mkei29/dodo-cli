---
name: go-development
description: Go言語のコードを書く・変更する・レビューするときに必ず適用するルール。.goファイルの編集、新規作成、テスト追加など、Go関連の作業すべてで自動適用する。
user-invocable: false
---

Go言語のコードを扱う際は、以下のルールを必ず守ること。

## 1. BUILD.bazel の更新

Goファイルを追加・削除・パッケージ構成を変更したあとは、必ずプロジェクトルートから以下を実行してBUILD.bazelを最新状態に保つ。

```bash
make gazelle
```

## 2. こまめなフォーマット

コードの変更を加えるたびに、以下を実行してフォーマットを維持する。コミット前に必ず実行すること。

```bash
make fmt
```

## 3. アーリーリターン

ガード節を使いネストを浅く保つ。条件を満たさない場合は早期に `return`（または `return err`）し、メインロジックのインデントを最小限に抑える。

```go
// Bad
func process(input string) (string, error) {
    if input != "" {
        result, err := doSomething(input)
        if err == nil {
            return result, nil
        } else {
            return "", err
        }
    } else {
        return "", errors.New("input is empty")
    }
}

// Good
func process(input string) (string, error) {
    if input == "" {
        return "", errors.New("input is empty")
    }
    result, err := doSomething(input)
    if err != nil {
        return "", err
    }
    return result, nil
}
```

## 4. テストはテーブルテスト形式で記述

Goのテストを書く際は、必ずテーブルテスト（Table-Driven Tests）形式で記述する。

```go
func TestFoo(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:  "正常系: ...",
            input: "...",
            want:  "...",
        },
        {
            name:    "異常系: ...",
            input:   "...",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Foo(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("Foo() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("Foo() = %v, want %v", got, tt.want)
            }
        })
    }
}
```
