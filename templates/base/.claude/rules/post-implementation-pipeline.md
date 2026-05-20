# ポストパイプライン順序

ポストパイプラインの単一の正規情報源。すべてのフローはこの順序に従わなければならない。

## 正規順序

```
/self-review → /verify → /test → /sync-docs → /pr
```

どのステップもスキップしてはならない。

## ステップの責任

| ステップ | エージェント | 目的 | 停止条件 |
|---------|-------------|------|---------|
| `/self-review` | `reviewer` | diff 品質のみ | CRITICAL 所見 |
| `/verify` | `verifier` | スペック適合性 + 静的解析 | fail 判定 |
| `/test` | `tester` | 振る舞いテスト | fail 判定 |
| `/sync-docs` | `doc-maintainer` | ドキュメント同期 | — |
| `/pr` | インライン | PR 作成 + プランアーカイブ | — |

## fix 後の再実行

問題を修正した後の再実行は **すべての** ステップを含む：

```
fix → /self-review → /verify → /test → /sync-docs → /pr
```

## この順序が参照されている場所

この順序を更新する場合、以下のすべての場所も更新する：
- `.claude/skills/work/SKILL.md`（ステップ 9）
- `.claude/rules/subagent-policy.md`（ポストパイプラインテーブル）
- `CLAUDE.md`（デフォルト動作）
- `README.md`（オペレーティングループ）
