# Python パック

Python プロジェクト向けの追加ルールと検証スクリプト。

## 含まれるもの

- `.claude/rules/python.md` — Python 固有のコーディングルール
- `scripts/verify-python.sh` — 型チェック・Lint・フォーマット検証

## 要件

- Python 3.11+
- プロジェクトに `pyproject.toml` または `setup.py` が存在すること

## インストール

```sh
ralph pack add python
```

## 検証スクリプトが使うツール

| ツール | 目的 | 設定ファイル |
|-------|------|------------|
| `mypy` | 型チェック | `pyproject.toml` / `mypy.ini` |
| `ruff` | Lint + フォーマット | `pyproject.toml` / `ruff.toml` |
| `pytest` | テスト | `pyproject.toml` / `pytest.ini` |

ツールが存在しない場合はそのステップをスキップする（エラーにならない）。
