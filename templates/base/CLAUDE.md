# CLAUDE.md

Claude Code 専用の常時参照ガイド。プロジェクト固有のルールはここに書かず、`.claude/rules/` に移動する。

## デフォルト動作

- `/spec` のみ手動トリガー。他のスキル（`/plan`、`/work`、`/self-review`、`/verify`、`/test`、`/sync-docs`、`/pr`）はパイプラインが自動的に呼び出す。
- リクエストが `/plan` には曖昧すぎる場合は `/spec` を使う。
- リスクがある、曖昧な、または複数ファイルにまたがる作業の前に `/plan` を使う。
- `/work` 完了後、ポストパイプラインがサブエージェント経由で自動実行される：`/self-review` → `/verify` → `/test` → `/sync-docs` → `/pr`
- `/self-review` は diff の品質のみ。`/verify` はスペック適合性 + 静的解析。`/test` は振る舞いテスト。各フェーズが個別のレポートを生成する。
- 完了を宣言する前に `./scripts/run-verify.sh` を実行する。

## Claude Code サーフェス

- `.claude/rules/` — 条件付きルール（パス・トピック別）
- `.claude/skills/` — オンデマンドワークフロー
- `.claude/agents/` — サブエージェント定義
- `.claude/hooks/` — 決定論的ランタイムガード

## ハードルール

- このファイルを短く保つ
- 詳細なトピックガイダンスは `.claude/rules/` へ
- ワークフローは `.claude/skills/` へ
- 繰り返す間違いはフック・テスト・CI・スクリプトに昇格させる
