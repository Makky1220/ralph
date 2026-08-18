<div align="center">

# ralph

**Claude Code プロジェクト向けハーネスエンジニアリング CLI**

`ralph init` 一発で、検証済みのフック・スキル・エージェント・パイプライン設定をプロジェクトに展開する。

[![Release](https://img.shields.io/github/v/release/thomas0124/ralph?style=flat-square&color=blue)](https://github.com/thomas0124/ralph/releases)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?style=flat-square&logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow?style=flat-square)](LICENSE)
[![CI](https://img.shields.io/github/actions/workflow/status/thomas0124/ralph/release.yml?style=flat-square&label=CI)](https://github.com/thomas0124/ralph/actions)
[![Homebrew](https://img.shields.io/badge/homebrew-thomas0124%2Ftap-orange?style=flat-square&logo=homebrew)](https://github.com/thomas0124/homebrew-tap)

</div>

---

## クイックスタート

```sh
brew install thomas0124/tap/ralph
```

```sh
cd your-project
ralph init
```

それだけ。あとは Claude Code を開いて `/spec` から始める。

---

## インストール

<details>
<summary><strong>Homebrew（推奨）</strong></summary>

```sh
brew install thomas0124/tap/ralph
```

</details>

<details>
<summary><strong>Go install</strong></summary>

```sh
go install github.com/thomas0124/ralph/cmd/ralph@latest
```

</details>

<details>
<summary><strong>バイナリ直接ダウンロード</strong></summary>

[Releases](https://github.com/thomas0124/ralph/releases) から OS/アーキテクチャに合ったバイナリをダウンロードして PATH に置く。

</details>

---

## コマンド一覧

| コマンド | 説明 |
|---------|------|
| `ralph init` | ハーネスをプロジェクトにセットアップ |
| `ralph upgrade` | 最新テンプレートに更新（3-way マージ） |
| `ralph doctor` | ハーネスの整合性チェック |
| `ralph pack add <lang>` | 言語パックを追加 |
| `ralph insights` | パイプライン実行データを集計・表示 |
| `ralph version` | バージョン確認 |

---

## `ralph init` — 何が展開されるか

```
your-project/
├── CLAUDE.md                          ← 常時参照ガイド
├── AGENTS.md                          ← エージェント向けプロジェクト定義
├── ralph.toml                         ← プロジェクト設定
│
├── .claude/
│   ├── settings.json                  ← 権限 + フック設定（dispatch 方式）
│   ├── hooks/
│   │   ├── ralph-dispatch.sh          ← 全フックの単一エントリポイント
│   │   ├── pre_bash_guard.sh          ← 危険な Bash コマンドをブロック
│   │   ├── post_edit_verify.sh        ← 編集後の自動検証
│   │   ├── check_mojibake.sh          ← U+FFFD 文字化け検出
│   │   ├── prompt_gate.sh             ← プロンプトごとの計画リフレッシュ促進
│   │   ├── session_start_context.sh   ← セッション開始時コンテキスト読み込み
│   │   ├── session_end_summary.sh     ← セッション終了時サマリ
│   │   ├── precompact_checkpoint.sh   ← コンパクト前チェックポイント
│   │   ├── PostToolUse.d/             ← イベント別 dispatch ディレクトリ
│   │   ├── PreToolUse.d/
│   │   └── ...
│   ├── skills/                        ← 11 スキル（後述）
│   ├── agents/                        ← 5 エージェント（後述）
│   └── rules/ralph/                   ← 10 ルール（後述）
│
├── .ralph/
│   ├── core/settings.ralph.json       ← ralph 管理の settings スナップショット
│   └── local/                         ← ユーザー拡張ポイント
│
├── scripts/                           ← 27 本のユーティリティスクリプト
│   ├── run-verify.sh / run-test.sh
│   ├── secret-scan.sh
│   ├── insights-append.sh
│   └── ...
│
├── docs/
│   ├── plans/active|archive/          ← 実装計画
│   ├── specs/  reports/               ← 仕様書・レポート
│   ├── insights/events/               ← パイプライン実行イベント（JSONL）
│   ├── recipes/  quality/
│   └── architecture/  research/  roadmap/
│
└── .github/workflows/verify.yml       ← CI ワークフロー
```

> **Git フック自動インストール**（git リポジトリの場合）
>
> | フック | 機能 |
> |--------|------|
> | `pre-commit` | ステージされたシークレットをブロック（AWS・GitHub・OpenAI 等） |
> | `commit-msg` | シークレットスキャン + Conventional Commits 検証 |
> | `pre-merge-commit` | マージ時のシークレットスキャン |
> | `prepare-commit-msg` | コミットメッセージ準備時のシークレットスキャン |

---

## オペレーティングループ

```
┌─────────────────────────────────────────────────────────────────┐
│                                                                   │
│   /spec  ──▶  /plan  ──▶  /work                                  │
│                               │                                   │
│                               ▼                                   │
│              /self-review ──▶ /verify ──▶ /test                  │
│                                               │                   │
│                                               ▼                   │
│                          /sync-docs ──▶ /cross-review ──▶ /pr    │
│                                         （任意）                  │
└─────────────────────────────────────────────────────────────────┘
```

ポストパイプラインはサブエージェントが自動実行する。`/work` を呼べば後は流れる。

| スキル | 役割 |
|--------|------|
| `/spec` | 要件の明確化（インタラクティブ、決定木形式） |
| `/plan` | クリーンベース worktree 作成 + 実装計画策定 |
| `/work` | スライス実装 → コミット → ポストパイプライン起動 |
| `/self-review` | diff 品質チェック（実装直後） |
| `/verify` | 仕様適合性 + 静的解析 |
| `/test` | 振る舞いテスト |
| `/sync-docs` | ドキュメント同期 |
| `/cross-review` | クロスモデルのセカンドオピニオン（任意） |
| `/pr` | PR 作成 + worktree クリーンアップ |

---

## 言語パック

```sh
ralph pack add typescript   # tsc + eslint
ralph pack add python       # mypy + ruff + pytest
ralph pack add golang       # go vet + staticcheck + golangci-lint
ralph pack add rust         # cargo check + clippy + fmt
ralph pack add dart         # dart analyze + flutter/pure-dart test
ralph pack add terraform    # terraform validate + tflint (tofu 優先)
```

パックが展開するファイル：

```
packs/languages/<lang>/
├── verify.sh          ← POSIX sh 検証スクリプト
├── rule.md            ← 言語ルール定義
└── README.md
.claude/rules/ralph/<lang>.md   ← Claude Code に読み込まれるルール
```

`verify.sh` は 2 つの環境変数で制御できる：

| 変数 | 説明 |
|------|------|
| `HARNESS_VERIFY_MODE` | `static` / `test` / `all`（デフォルト: `all`） |
| `RALPH_VERIFY_PROJECT_ROOTS` | モノレポで検証対象ルートを限定（スペース区切り） |

---

## `ralph upgrade` — スマートアップグレード

```sh
ralph upgrade
```

| 機能 | 説明 |
|------|------|
| **3-way マージ** | `settings.json` の ralph 管理領域を安全に更新しながらユーザー設定を保持 |
| **managed block** | `AGENTS.md`・`.gitignore` の `BEGIN/END RALPH MANAGED` ブロックのみ更新 |
| **drift 検出** | ローカル変更を上書きせず記録 |
| **色付き diff** | 変更内容をターミナルに表示 |
| **アップグレードレポート** | `docs/reports/` に自動生成 |

---

## `ralph doctor` — ヘルスチェック

```sh
ralph doctor
```

| チェック項目 | 説明 |
|-------------|------|
| Claude CLI | `require_claude_cli` で必須／警告を切り替え |
| Go | `require_go` で必須チェックを有効化 |
| settings.json | フック設定の整合性 |
| フックスクリプト | 実行可能ビットの確認 |
| マニフェスト | バージョン・ハッシュの整合性 |
| 言語パック | verify.sh の存在と実行可能ビット |

`ralph.toml` で動作を調整できる：

```toml
[doctor]
require_claude_cli = true   # false にすると警告のみ
require_go         = false  # true にすると go 必須チェックを有効化
```

---

## `ralph insights` — パイプライン分析

```sh
ralph insights              # イベント集計を表示
ralph insights --json       # JSON 出力
ralph insights backfill     # docs/reports/ から過去データを生成
```

`docs/insights/events/` の JSONL ファイルを集計して表示する：

- フェーズ別テーブル（phase / events / verdicts / findings / triage）
- エスカレーション（cycle >= 2 に達したスラッグの比較）
- ルーティング統計（honored rate per phase）

---

## モデルルーティング

```
┌─────────────────────────────────────────────────────┐
│  orchestrator  │  ユーザー選択モデル（セッション）    │
│  reviewer      │  opus   — 判断・評価                │
│  verifier      │  sonnet — 検証・テスト・実装         │
│  implementer   │  sonnet                             │
│  bulk lookups  │  haiku  — grep / ファイル調査        │
└─────────────────────────────────────────────────────┘
```

> エージェント frontmatter に `model:` を必ず明示する。省略すると継承になり、高価なモデルが想定外に多重起動される。

---

## 開発

```sh
git clone https://github.com/thomas0124/ralph
cd ralph
go build ./cmd/ralph
./ralph version
```

```sh
# Go テスト
go test -race ./...
go vet ./...

# シェルスクリプトテスト（25 本）
./tests/test-secret-scan.sh
./tests/test-ralph-dispatch.sh
./tests/test-check-mojibake.sh
# ...
```

---

## ライセンス

[MIT](LICENSE) © thomas0124
