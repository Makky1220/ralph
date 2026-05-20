#!/usr/bin/env bash
set -euo pipefail

PASS=0
FAIL=0
SKIP=0

_pass() { echo "  [PASS] $*"; ((PASS++)); }
_fail() { echo "  [FAIL] $*"; ((FAIL++)); }
_skip() { echo "  [SKIP] $*"; ((SKIP++)); }

echo "=== Rust 検証 ==="

if ! command -v cargo &>/dev/null; then
  echo "[ERROR] cargo が見つからない" >&2
  exit 1
fi

echo "Rust: $(rustc --version)"
echo "Cargo: $(cargo --version)"

# --- cargo check ---
echo ""
echo "--- コンパイル検査 (cargo check) ---"
if cargo check --all-targets 2>&1; then
  _pass "cargo check"
else
  _fail "コンパイルエラーあり"
fi

# --- cargo clippy ---
echo ""
echo "--- Lint (cargo clippy) ---"
if cargo clippy --all-targets -- -D warnings 2>&1; then
  _pass "cargo clippy"
else
  _fail "Clippy 警告あり"
fi

# --- cargo fmt ---
echo ""
echo "--- フォーマット (cargo fmt) ---"
if cargo fmt --check 2>&1; then
  _pass "cargo fmt"
else
  _fail "フォーマット違反あり (cargo fmt で修正)"
fi

# --- cargo test ---
echo ""
echo "--- テスト (cargo test) ---"
if cargo test 2>&1; then
  _pass "cargo test"
else
  _fail "テスト失敗あり"
fi

# --- 結果 ---
echo ""
echo "=== 結果: PASS=$PASS  FAIL=$FAIL  SKIP=$SKIP ==="

if [ "$FAIL" -gt 0 ]; then
  exit 1
fi
exit 0
