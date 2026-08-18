# ralph

Claude Code 専用のハーネスエンジニアリング CLI。プロジェクトに Claude Code の設定・フック・スキル・エージェント・パイプラインを一括でセットアップする。

```
ralph init       # 新規プロジェクトにハーネスをセットアップ
ralph upgrade    # ハーネスを最新テンプレートに更新（コンフリクト検出付き）
ralph doctor     # ハーネスの整合性を検査
ralph pack add typescript  # TypeScript 言語パックを追加
ralph version    # バージョン情報を表示
```

## インストール

### Homebrew

```sh
brew install thomas0124/tap/ralph
```

### go install

```sh
go install github.com/thomas0124/ralph/cmd/ralph@latest
```

### バイナリダウンロード

[Releases](https://github.com/thomas0124/ralph/releases) から OS に合ったバイナリをダウンロードして `PATH` に置く。

---

## 何をするのか

`ralph init` はプロジェクトルートに以下を展開する：

```
CLAUDE.md                        ← 常時参照ガイド
ralph.toml                       ← プロジェクト設定
.claude/
  settings.json                  ← フック + パーミッション設定
  hooks/                         ← 7 本のランタイムガード
  skills/                        ← 8 スキル（/spec /plan /work … /pr）
  agents/                        ← 4 サブエージェント定義
  rules/                         ← 7 条件付きルール
scripts/
  run-verify.sh                  ← 全検証をまとめて実行
  run-static-verify.sh           ← 静的検証のみ
  run-test.sh                    ← テストのみ
  …（他 5 本）
docs/
  plans/active/   plans/archive/
  specs/          reports/
  tech-debt/      evidence/
  quality/definition-of-done.md
```

---

## オペレーティングループ

```
/spec   →   /plan   →   /work
                            ↓
              /self-review → /verify → /test → /sync-docs → /pr
```

- `/spec` — 要件の明確化（インタラクティブ）
- `/plan` — 実装計画の策定（スライス分割）
- `/work` — 実装（スライスごとにコミット → 検証）
- ポストパイプライン — サブエージェントが自動実行

---

## 言語パック

```sh
ralph pack add typescript   # tsc + eslint
ralph pack add python       # mypy + ruff + pytest
ralph pack add golang       # go vet + staticcheck + golangci-lint
ralph pack add rust         # cargo check + clippy + fmt
ralph pack add dart         # dart analyze + flutter test
ralph pack add terraform    # terraform validate + tflint
```

パックは以下のファイルをプロジェクトに追加する：

```
packs/languages/<lang>/
  verify.sh              ← POSIX sh 製の検証スクリプト
  rule.md                ← 言語固有のルール定義
  README.md              ← パックのドキュメント
.claude/rules/ralph/
  <lang>.md              ← Claude Code に読み込まれるルール
```

`verify.sh` は `HARNESS_VERIFY_MODE`（`static` / `test` / `all`）と `RALPH_VERIFY_PROJECT_ROOTS`（モノレポ向けルート指定）に対応している。

---

## アップグレード

```sh
ralph upgrade
```

テンプレートの変更とローカル変更を 3-way マージで統合する。コンフリクトは対話的に解決できる。

---

## doctor

```sh
ralph doctor
```

チェック内容：

- Claude CLI (`claude`) の存在（`ralph.toml` の `require_claude_cli` で制御）
- Go のインストール状況（`require_go` で制御）
- `.claude/settings.json` の整合性
- フックスクリプトの実行可能ビット
- マニフェストバージョン
- 有効な言語パックの検証スクリプト

### ralph.toml

`ralph.toml` でプロジェクトレベルの挙動を設定できる：

```toml
[doctor]
require_claude_cli = true   # false にすると警告のみ
require_go         = false  # true にすると go 必須チェックを有効化
```

---

## 開発

```sh
git clone https://github.com/thomas0124/ralph
cd ralph
go build ./cmd/ralph
./ralph version
```

```sh
go test -race ./...
go vet ./...
```

---

## ライセンス

MIT
