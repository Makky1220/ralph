# Python ルール

このルールは Python プロジェクトでのみアクティブになる。

## 型アノテーション

- すべての公開関数・メソッドに型アノテーションを付ける
- `Any` は避ける — 代わりに `object` または具体的な型を使う
- `Optional[T]` より `T | None` (Python 3.10+) を優先する
- `TypedDict` を dict のスキーマ定義に使う

## コード品質

- 関数は 30 行以内に収める
- クラスより関数・モジュールを優先する
- グローバル状態を避ける
- 副作用のある関数と純粋関数を分離する

## エラーハンドリング

- 例外を黙って飲み込まない (`except: pass`)
- 具体的な例外クラスをキャッチする（`except Exception` は最終手段）
- カスタム例外クラスには `Error` サフィックスを付ける

## 非同期

- asyncio を使う場合は `async/await` を一貫して使う
- `asyncio.get_event_loop()` より `asyncio.run()` を優先する

## テスト

- pytest でテストを書く
- `conftest.py` でフィクスチャを共有する
- 外部サービスには `pytest-mock` を使う

## 検証

```sh
./scripts/verify-python.sh
```
