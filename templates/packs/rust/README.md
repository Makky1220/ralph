# Rust パック

Rust プロジェクト向けの追加ルールと検証スクリプト。

## 含まれるもの

- `.claude/rules/rust.md` — Rust 固有のコーディングルール
- `scripts/verify-rust.sh` — チェック・Clippy・フォーマット検証

## 要件

- Rust 1.75+ (edition 2021)
- プロジェクトに `Cargo.toml` が存在すること

## インストール

```sh
ralph pack add rust
```

## 検証スクリプトが使うツール

| ツール | 目的 | 設定ファイル |
|-------|------|------------|
| `cargo check` | コンパイル検査 | `Cargo.toml` |
| `cargo clippy` | Lint | `.cargo/config.toml` |
| `cargo fmt` | フォーマット | `rustfmt.toml` |
| `cargo test` | テスト | — |
