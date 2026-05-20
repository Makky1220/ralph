# TypeScript パック

TypeScript プロジェクト向けの追加ルールと検証スクリプト。

## 含まれるもの

- `.claude/rules/typescript.md` — TypeScript 固有のコーディングルール
- `scripts/verify-typescript.sh` — 型チェック・Lint・フォーマット検証

## 要件

- Node.js 20+
- TypeScript 5+
- プロジェクトに `tsconfig.json` が存在すること

## インストール

```sh
ralph pack add typescript
```

## 検証スクリプトが使うツール

| ツール | 目的 | 設定ファイル |
|-------|------|------------|
| `tsc` | 型チェック | `tsconfig.json` |
| `eslint` | Lint | `.eslintrc.*` / `eslint.config.*` |
| `prettier` | フォーマット | `.prettierrc.*` |

ツールが存在しない場合はそのステップをスキップする（エラーにならない）。
