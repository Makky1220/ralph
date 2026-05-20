#!/usr/bin/env bash
set -euo pipefail

PASS=0
FAIL=0
SKIP=0

_pass() { echo "  [PASS] $*"; ((PASS++)); }
_fail() { echo "  [FAIL] $*"; ((FAIL++)); }
_skip() { echo "  [SKIP] $*"; ((SKIP++)); }

echo "=== TypeScript 検証 ==="

# --- 型チェック ---
echo ""
echo "--- 型チェック (tsc) ---"
if ! command -v tsc &>/dev/null; then
  _skip "tsc が見つからない (node_modules/.bin/tsc も確認)"
  if [ -f "node_modules/.bin/tsc" ]; then
    if node_modules/.bin/tsc --noEmit 2>&1; then
      _pass "型チェック"
    else
      _fail "型エラーあり"
    fi
  fi
elif tsc --noEmit 2>&1; then
  _pass "型チェック"
else
  _fail "型エラーあり"
fi

# --- ESLint ---
echo ""
echo "--- Lint (eslint) ---"
ESLINT_BIN=""
if command -v eslint &>/dev/null; then
  ESLINT_BIN="eslint"
elif [ -f "node_modules/.bin/eslint" ]; then
  ESLINT_BIN="node_modules/.bin/eslint"
fi

if [ -z "$ESLINT_BIN" ]; then
  _skip "eslint が見つからない"
else
  if $ESLINT_BIN . --ext .ts,.tsx 2>&1; then
    _pass "ESLint"
  else
    _fail "ESLint エラーあり"
  fi
fi

# --- Prettier ---
echo ""
echo "--- フォーマット (prettier) ---"
PRETTIER_BIN=""
if command -v prettier &>/dev/null; then
  PRETTIER_BIN="prettier"
elif [ -f "node_modules/.bin/prettier" ]; then
  PRETTIER_BIN="node_modules/.bin/prettier"
fi

if [ -z "$PRETTIER_BIN" ]; then
  _skip "prettier が見つからない"
else
  if $PRETTIER_BIN --check "**/*.{ts,tsx}" 2>&1; then
    _pass "フォーマット"
  else
    _fail "フォーマット違反あり (prettier --write で修正)"
  fi
fi

# --- 結果 ---
echo ""
echo "=== 結果: PASS=$PASS  FAIL=$FAIL  SKIP=$SKIP ==="

if [ "$FAIL" -gt 0 ]; then
  exit 1
fi
exit 0
