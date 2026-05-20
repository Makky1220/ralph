# 完了の定義

タスクは以下がすべて真のときのみ「完了」とみなす：

## 実装チェックリスト

- [ ] プランの受け入れ基準がすべて満たされている
- [ ] 各スライスが個別にコミット済み（conventional commit 形式）
- [ ] `./scripts/run-verify.sh` が 0 で終了
- [ ] 発見された技術的負債が `docs/tech-debt/` に記録済み

## ポストパイプラインチェックリスト

- [ ] `/self-review` 完了（`docs/reports/self-review-*.md` 存在）— CRITICAL 所見なし
- [ ] `/verify` 完了（`docs/reports/verify-*.md` 存在）— pass または partial-pass
- [ ] `/test` 完了（`docs/reports/test-*.md` 存在）— **pass 必須**
- [ ] `/sync-docs` 完了
- [ ] `/pr` 作成済み（URL 表示済み）
- [ ] プランがアーカイブ済み（`docs/plans/archive/`）

## 重要

「完了」と言う前に、何が検証済みで何が未検証かを明示すること。テストなしに完了と宣言しない。
