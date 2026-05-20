---
name: pr
description: 自己レビュー・検証・テストがパスした後にプルリクエストを作成する。ブランチプッシュ・PR 作成・プランアーカイブ・ハンドオフを処理する。/sync-docs 完了後に自動呼び出し。
---
完了した作業を人間のレビューとマージのためにハンドオフする PR を作成する。

## 事前チェック

PR を作成する前に、以下をすべて確認する：

1. `docs/reports/` に CRITICAL 所見がない自己レビューレポートが存在する。
2. `docs/reports/` に pass または partial-pass 判定の検証レポートが存在する。
3. `docs/reports/` に **pass 判定**のテストレポートが存在する。**テストが失敗した場合は PR を作成しない。**
4. 生のエビデンスが `docs/evidence/` に保存されている。
5. ブランチ名が `./scripts/branch-name.sh validate "$(git branch --show-current)"` をパスする。
6. main または master にいない。
7. `gh` CLI が利用可能（ない場合はマニュアルコマンドを提供する）。

事前チェックが失敗した場合は、何が足りないかを説明して停止する。

## ステップ

1. 未コミットの変更を `git status --porcelain` で確認する。
   - **未コミットの変更がある場合**：`git add`（特定のファイルを優先、`-A` は避ける）でステージし conventional commit を作成する。
   - **作業ツリーがクリーン**（中間コミット済み）：ステージとコミットをスキップしてプッシュへ。
2. ブランチをプッシュする：`git push -u origin HEAD`
3. `gh pr create` で PR を作成する。**PR タイトルと本文は日本語で書く。**
   - PR タイトルは `./scripts/branch-name.sh title-prefix "$(git branch --show-current)"` で取得したプレフィックスで始める（例：`feat/...` → `feat: ...`）。
   - `--draft` はオペレーターが明示的に要求した場合のみ使う。
   - ツール（Claude、AI など）を PR 提供者として識別する記述を避ける。
   - ブランチが `<type>/<issue>/<slug>` の場合、PR 本文に `Closes #<number>` を含める。部分的な作業は `Refs #<number>` を使う。
4. `./scripts/ensure-pr-ready.sh <pr-url>` を実行する。失敗した場合は停止して PR が完了したと報告しない。
5. 大きな diff（500行以上）の場合、`docs/reports/walkthrough-<date>-<slug>.md` を作成する。
6. プランをアーカイブする：`./scripts/archive-plan.sh <plan-path>`

## 完了ゲート

以下がすべて真でなければ PR が完了したと表示しない：

- [ ] PR が作成され、URL がユーザーに表示された
- [ ] `./scripts/ensure-pr-ready.sh <pr-url>` がパスした
- [ ] プランが `docs/plans/active/` から `docs/plans/archive/` にアーカイブされた
- [ ] 大きな diff の場合：ウォークスルーレポートが存在する
- [ ] コミットが conventional commit 形式に従っている
