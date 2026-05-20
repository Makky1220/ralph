# TypeScript ルール

このルールは TypeScript プロジェクトでのみアクティブになる。

## 型安全性

- `any` を使わない — 代わりに `unknown` または具体的な型を使う
- `as` キャストは型ガードが型推論より正確な場合のみ使う
- すべての公開 API に明示的な戻り値型を付ける
- `null` と `undefined` を区別する — `strictNullChecks: true` が必須

## 非同期

- コールバックより `async/await` を優先する
- Promise のエラーを必ずハンドルする（`catch` または `try/catch`）
- `Promise<void>` を返す関数に `await` を付け忘れない

## モジュール

- デフォルトエクスポートより名前付きエクスポートを優先する（ツールとの親和性が高い）
- 循環依存を避ける — `madge` または `eslint-plugin-import` で検出する
- インポートパスにバレルファイル（`index.ts`）を使いすぎない

## コード品質

- 関数は 40 行以内に収める
- ユニオン型のすべてのケースを網羅した `switch` には `default: throw new Error(...)` を使う
- クラスより関数コンポーネント・純粋関数を優先する

## テスト

- Vitest または Jest でユニットテストを書く
- モックは使わず統合テストを優先できる場合はそうする
- `expect.assertions(n)` で非同期テストの assertion 漏れを防ぐ

## 検証

```sh
./scripts/verify-typescript.sh
```
