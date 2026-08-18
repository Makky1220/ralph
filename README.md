# ralph

Claude Code プロジェクト向けハーネスエンジニアリング CLI。\
`ralph init` 一発で、検証済みのフック・スキル・エージェント・パイプライン設定をプロジェクトに展開する。

```
ralph init                  # ハーネスをセットアップ
ralph upgrade               # 最新テンプレートに更新（3-way マージ）
ralph doctor                # ハーネスの整合性チェック
ralph pack add typescript   # 言語パックを追加
ralph insights              # パイプライン実行データを集計・表示
ralph version               # バージョン確認
```

## インストール

```sh
brew install thomas0124/tap/ralph
```

```sh
go install github.com/thomas0124/ralph/cmd/ralph@latest
```

[Releases](https://github.com/thomas0124/ralph/releases) からバイナリを直接ダウンロードすることもできる。

---

## `ralph init` — 何が展開されるか

```
CLAUDE.md                             ← 常時参照ガイド
AGENTS.md                             ← エージェント向けプロジェクト定義
ralph.toml                            ← プロジェクト設定

.claude/
  settings.json                       ← 権限 + フック設定（dispatch 方式）
  hooks/
    ralph-dispatch.sh                 ← 全フックの単一エントリポイント
    pre_bash_guard.sh                 ← 危険な Bash コマンドをブロック
    post_edit_verify.sh               ← 編集後の自動検証
    check_mojibake.sh                 ← U+FFFD 文字化け検出
    prompt_gate.sh                    ← プロンプトごとの計画リフレッシュ促進
    session_start_context.sh          ← セッション開始時コンテキスト読み込み
    session_end_summary.sh            ← セッション終了時サマリ
    precompact_checkpoint.sh          ← コンパクト前チェックポイント
    PostToolUse.d/                    ← イベント別 dispatch ディレクトリ
    PreToolUse.d/  PreCompact.d/  …
  skills/
    spec/   plan/   work/             ← 開発フロー（仕様・計画・実装）
    self-review/  verify/  test/      ← 品質チェック
    sync-docs/  pr/                   ← ドキュメント同期・PR 作成
    anti-bottleneck/                  ← 不要な確認を避けるポリシー
    audit-harness/                    ← ハーネス自体の監査
    cross-review/                     ← クロスモデルレビュー
  agents/
    implementer.md                    ← sonnet 固定のスコープ限定実装
    reviewer.md   verifier.md
    tester.md     doc-maintainer.md
  rules/ralph/
    ralph-workflow.md                 ← ワークフロー全体ガイド
    model-routing.md                  ← opus/sonnet/haiku モデル選択ルール
    agent-messaging.md                ← エージェント間メッセージングプロトコル
    architecture.md  planning.md  testing.md  …（計 10 ルール）

.ralph/
  core/settings.ralph.json            ← ralph 管理の settings スナップショット
  local/hooks/  local/verify.d/       ← ユーザー拡張ポイント

scripts/
  run-verify.sh                       ← 全検証を実行
  run-static-verify.sh                ← 静的解析のみ
  run-test.sh                         ← テストのみ
  secret-scan.sh                      ← シークレットスキャン（git hook から呼ばれる）
  ralph-config.sh  ralph-common.sh    ← 共通設定・ユーティリティ
  detect-languages.sh                 ← 使用言語の自動検出
  insights-append.sh                  ← インサイトイベント記録
  gc-artifacts.sh  ralph-worktree.sh  …（計 27 本）

docs/
  plans/active/   plans/archive/      ← 実装計画
  specs/          reports/            ← 仕様書・レポート
  insights/events/                    ← パイプライン実行イベント（JSONL）
  recipes/        quality/            ← ハウツー・品質定義
  architecture/   research/  roadmap/
```

`ralph init` は git リポジトリなら `.git/hooks/` にも自動インストールする：

- `pre-commit` — ステージされたシークレットをブロック（AWS・GitHub・OpenAI 等）
- `commit-msg` — コミットメッセージのシークレットスキャン + Conventional Commits 検証
- `pre-merge-commit` / `prepare-commit-msg` — マージ時の秘密情報スキャン

---

## オペレーティングループ

```
/spec  →  /plan  →  /work
                       ↓
         /self-review → /verify → /test → /sync-docs → /pr
```

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

ポストパイプラインはサブエージェントが自動実行する。`/work` を呼べば後は流れる。

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
  verify.sh        ← POSIX sh 検証スクリプト
  rule.md          ← 言語ルール定義
  README.md
.claude/rules/ralph/<lang>.md  ← Claude Code に読み込まれるルール
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

- **3-way マージ** で `settings.json` の ralph 管理領域（env / permissions / hooks）を安全に更新しながらユーザー設定を保持
- **managed block** で `AGENTS.md`・`.gitignore` の `BEGIN/END RALPH MANAGED` ブロックのみ更新
- **drift 検出** でローカル変更を上書きせず記録
- **色付き diff** で変更内容をターミナルに表示
- **アップグレードレポート** を `docs/reports/` に自動生成

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

| 役割 | モデル |
|------|--------|
| オーケストレーター（セッション） | ユーザー選択モデル |
| 判断系（reviewer） | `opus` |
| 手続き系（verifier / tester / implementer） | `sonnet` |
| 一括機械的作業（grep / ファイル調査） | `haiku` |

エージェント frontmatter に `model:` を必ず明示する。省略すると継承になり、高価なモデルが想定外に多重起動される。

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

シェルスクリプトのテスト：

```sh
./tests/test-secret-scan.sh
./tests/test-ralph-dispatch.sh
./tests/test-check-mojibake.sh
# ... 25 本のテストスクリプト
```

---

## ライセンス

MIT
