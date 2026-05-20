#!/usr/bin/env bash
set -euo pipefail

PASS=0
FAIL=0
SKIP=0

_pass() { echo "  [PASS] $*"; ((PASS++)); }
_fail() { echo "  [FAIL] $*"; ((FAIL++)); }
_skip() { echo "  [SKIP] $*"; ((SKIP++)); }

echo "=== Python 検証 ==="

# Python バージョン確認
PYTHON_BIN=""
for bin in python3 python; do
  if command -v "$bin" &>/dev/null; then
    PYTHON_BIN="$bin"
    break
  fi
done

if [ -z "$PYTHON_BIN" ]; then
  echo "[ERROR] python3/python が見つからない" >&2
  exit 1
fi

echo "Python: $($PYTHON_BIN --version)"

# --- mypy ---
echo ""
echo "--- 型チェック (mypy) ---"
if ! command -v mypy &>/dev/null; then
  _skip "mypy が見つからない"
else
  if mypy . 2>&1; then
    _pass "mypy"
  else
    _fail "型エラーあり"
  fi
fi

# --- ruff check ---
echo ""
echo "--- Lint (ruff check) ---"
if ! command -v ruff &>/dev/null; then
  _skip "ruff が見つからない"
else
  if ruff check . 2>&1; then
    _pass "ruff check"
  else
    _fail "Lint エラーあり"
  fi
fi

# --- ruff format ---
echo ""
echo "--- フォーマット (ruff format) ---"
if ! command -v ruff &>/dev/null; then
  _skip "ruff が見つからない"
else
  if ruff format --check . 2>&1; then
    _pass "ruff format"
  else
    _fail "フォーマット違反あり (ruff format で修正)"
  fi
fi

# --- pytest ---
echo ""
echo "--- テスト (pytest) ---"
if ! command -v pytest &>/dev/null; then
  _skip "pytest が見つからない"
else
  if pytest --tb=short -q 2>&1; then
    _pass "pytest"
  else
    _fail "テスト失敗あり"
  fi
fi

# --- 結果 ---
echo ""
echo "=== 結果: PASS=$PASS  FAIL=$FAIL  SKIP=$SKIP ==="

if [ "$FAIL" -gt 0 ]; then
  exit 1
fi
exit 0
