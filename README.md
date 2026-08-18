<div align="center">

# ralph

**Claude Code / Codex プロジェクト向けハーネス & マルチエージェント CLI**

`ralph init` 一発でハーネスを展開。`ralph org` で Claude Code と Codex を並列シートとして動かす。

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
| `ralph org <verb>` | 自律マルチシート org ランタイム（後述） |
| `ralph status` | アクティブシートの状態一覧 |
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
ralph doctor --probe-models   # org ランタイムのモデル疎通確認まで実施
```

| チェック項目 | 説明 |
|-------------|------|
| Claude CLI | `require_claude_cli` で必須／警告を切り替え |
| Codex CLI | `require_codex_cli` で必須チェックを有効化 |
| Go | `require_go` で必須チェックを有効化 |
| settings.json | フック設定の整合性 |
| フックスクリプト | 実行可能ビットの確認 |
| マニフェスト | バージョン・ハッシュの整合性 |
| 言語パック | verify.sh の存在と実行可能ビット |
| herdr | ターミナルペイン管理ツール（org ランタイム用） |
| agmsg | シート間メッセージングスクリプト群 |
| org envelope | herdr + agmsg の疎通確認 |
| model probes | `--probe-models` 時：各ドライバーのモデル疎通確認 |

`ralph.toml` で動作を調整できる：

```toml
[doctor]
require_claude_cli = true   # false にすると警告のみ
require_codex_cli  = false  # true にすると codex 必須チェックを有効化
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

## org ランタイム — 自律マルチシート実行

**herdr**（ターミナルペイン管理）と **agmsg**（型付きメッセージング）を組み合わせて、複数の AI シートを並列に動かす仕組みです。Claude Code と Codex を混在させることができます。

### 前提ツール

| ツール | インストール |
|--------|-------------|
| [herdr](https://herdr.dev) | `brew install herdr` |
| [agmsg](https://github.com/fujibee/agmsg) | `git clone https://github.com/fujibee/agmsg ~/.agents/skills/agmsg` |

```sh
ralph doctor   # herdr / agmsg の疎通確認
```

### トポロジー

```
lead (Claude Code)
  ├── seat: implementer  (driver: claude / codex)
  ├── seat: reviewer     (driver: claude)
  └── seat: tester       (driver: codex)
```

lead がタスクを各シートに送り、RESULT を受け取って調整します。シートは herdr がターミナルペインとして起動し、シート間のメッセージは agmsg の TASK / RESULT / BLOCKED / REVIEW / QUESTION 形式でやり取りされます。

### org コマンド

```sh
ralph org spawn --driver claude --seat reviewer   # Claude Code シートを起動
ralph org spawn --driver codex  --seat verifier   # Codex シートを起動
ralph org send  --seat reviewer "TYPE: TASK\n..."  # タスク送信
ralph org wait  --seat reviewer                    # 完了待機
ralph org read  --seat reviewer                    # 最新メッセージ取得
ralph org stop  --seat reviewer                    # シート停止
ralph org disband                                  # 全シート停止
ralph status                                       # アクティブシート一覧
```

### `ralph.toml` org 設定

```toml
[org]
driver_pool      = ["claude", "codex"]
model_pool       = [
  { driver = "claude", model = "opus" },    # 判断・レビュー系
  { driver = "claude", model = "sonnet" },  # 実装系
  { driver = "codex",  model = "gpt-5.5" }, # 検証系
]
max_seats        = 5
deadman_minutes  = 10
agmsg_home       = "~/.agents/skills/agmsg"

[org.permissions]
default          = "default"
codex_verified   = false   # true: codex に広い権限を付与

[org.budget]
seat_wall_clock_minutes  = 30
total_wall_clock_minutes = 120
max_fix_rounds           = 3
```

### Claude Code + Codex 橋渡し

`/cross-review` スキルで、一方が実装してもう一方がレビューする非同期ワークフローが使えます：

```
Claude が実装 → Codex がレビュー
Codex が実装  → Claude がレビュー
```

org ランタイムなしでも `/cross-review` 単体で動作します。

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
