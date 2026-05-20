# Go パック

Go プロジェクト向けの追加ルールと検証スクリプト。

## 含まれるもの

- `.claude/rules/golang.md` — Go 固有のコーディングルール
- `scripts/verify-golang.sh` — vet・Lint・テスト検証

## 要件

- Go 1.21+
- プロジェクトに `go.mod` が存在すること

## インストール

```sh
ralph pack add golang
```

## 検証スクリプトが使うツール

| ツール | 目的 | 設定ファイル |
|-------|------|------------|
| `go vet` | 静的解析 | — |
| `staticcheck` | 拡張解析 | `staticcheck.conf` |
| `golangci-lint` | 統合 Lint | `.golangci.yml` |
| `go test` | テスト | — |

`staticcheck` と `golangci-lint` が存在しない場合はそのステップをスキップする。
