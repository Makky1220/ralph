---
name: sync-docs
description: 振る舞い・コマンド・契約・ワークフローが変化した後に、プラン・ドキュメント・指示ファイルを同期する。スキル・フック・ルール・スクリプトの変更後のハーネス内部一貫性もカバーする。/test の後、/pr の前に doc-maintainer エージェント経由で呼び出される。
allowed-tools: Read, Grep, Glob, Write, Edit, Bash
---
実装またはハーネス構造が変わって、ドキュメントがドリフトした可能性がある場合にこのスキルを使う。

## プロダクトレベルの同期

必要に応じて更新：
- アクティブプランの進捗
- `README.md`
- `CLAUDE.md`
- `.claude/rules/`
- `docs/quality/`
- `docs/reports/` のリンクや参照

`CLAUDE.md` を短く安定した状態に保つ。新しいルールがパスまたはトピック固有の場合は `.claude/rules/` に入れる。

## ハーネス内部の同期

スキル・フック・ルール・スクリプト・言語パックが変わった場合：

- **スキルの追加/削除/リネーム**：`CLAUDE.md` のリポジトリマップが現在のスキルセットを反映しているか？ `README.md` が現在のオペレーティングループを示しているか？
- **フックの追加/削除**：`.claude/settings.json` が正しいフックスクリプトを参照しているか？ 削除されたフックが消えているか？
- **ルールの追加/削除**：`.claude/rules/` がプロジェクトの言語とトピックと一致しているか？
- **言語パックの追加/削除**：`scripts/detect-languages.sh` がその言語を検出するか？ 対応する `.claude/rules/<lang>.md` があるか？ `packs/languages/<lang>/verify.sh` が実際のベリファイアを実行するか？
- **スクリプトの追加/削除**：`README.md` のクイックスタートが有効なスクリプトを参照しているか？
- **品質ゲートの変更**：`docs/quality/definition-of-done.md` が `/work` の実際の完了ワークフローと一致しているか？
