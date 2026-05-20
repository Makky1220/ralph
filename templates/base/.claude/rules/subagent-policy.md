# サブエージェント委任ポリシー

作業をサブエージェントに委任するタイミングと方法。パイプライン順序は `post-implementation-pipeline.md` で定義されている。

## /work のポストパイプライン — フェーズロール

`/work` が完了したら、以下のフェーズ固有のサブエージェントでポストパイプラインを実行する：

| ステップ | サブエージェント | スキル | 目的 |
|---------|----------------|-------|------|
| 1 | `reviewer` | `/self-review` | diff 品質 |
| 2 | `verifier` | `/verify` | スペック適合性 + 静的解析 |
| 3 | `tester` | `/test` | 振る舞いテスト |
| 4 | `doc-maintainer` | `/sync-docs` | ドキュメント同期 |

ステップ 1-3 は順次実行する（ある出力が次の入力に影響するため）。ステップ 4 はテストがパスした後に実行する。Claude Code は `Task(subagent_type=<エージェント名>)` ツールを使用する。

### 実行

```
reviewer: run /self-review for the current diff against plan <slug>
  → reviewer produces docs/reports/self-review-*.md
  → CRITICAL 所見がある場合：継続前に修正して停止

verifier: run /verify against plan <slug>
  → verifier produces docs/reports/verify-*.md
  → fail 判定の場合：継続前に修正して停止

tester: run /test against plan <slug>
  → tester produces docs/reports/test-*.md
  → fail 判定の場合：/pr に進まない
```

### フォールバック

サブエージェントの実行が失敗した場合（ツールエラー、レビュー所見ではない）、対応するスキルをインラインで実行してレポートにフォールバックを記録する。フェーズをサイレントにスキップしない。

## /spec — 常にインライン

`/spec` はユーザーとのインタラクティブな対話（`AskUserQuestion` による要件の明確化）に大きく依存するためメインコンテキストで実行する。

## /plan — 常にインライン

`/plan` はユーザー対話（フロー選択・目的確認・重要分岐の解決）に大きく依存するためメインコンテキストで実行する。

## /pr — 常にインライン

`/pr` は実装コンテキスト（なぜそのコードが書かれたか、設計決定が何か）に依存するためメインコンテキストで実行する。
