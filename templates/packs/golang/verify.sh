#!/usr/bin/env bash
set -euo pipefail

PASS=0
FAIL=0
SKIP=0

_pass() { echo "  [PASS] $*"; ((PASS++)); }
_fail() { echo "  [FAIL] $*"; ((FAIL++)); }
_skip() { echo "  [SKIP] $*"; ((SKIP++)); }

echo "=== Go 検証 ==="

if ! command -v go &>/dev/null; then
  echo "[ERROR] go が見つからない" >&2
  exit 1
fi

echo "Go: $(go version)"

# --- go vet ---
echo ""
echo "--- 静的解析 (go vet) ---"
if go vet ./... 2>&1; then
  _pass "go vet"
else
  _fail "go vet エラーあり"
fi

# --- gofmt ---
echo ""
echo "--- フォーマット (gofmt) ---"
UNFORMATTED=$(gofmt -l . 2>&1)
if [ -z "$UNFORMATTED" ]; then
  _pass "gofmt"
else
  echo "  フォーマット違反ファイル:"
  echo "$UNFORMATTED" | sed 's/^/    /'
  _fail "フォーマット違反あり (gofmt -w で修正)"
fi

# --- staticcheck ---
echo ""
echo "--- 拡張解析 (staticcheck) ---"
if ! command -v staticcheck &>/dev/null; then
  _skip "staticcheck が見つからない (go install honnef.co/go/tools/cmd/staticcheck@latest)"
else
  if staticcheck ./... 2>&1; then
    _pass "staticcheck"
  else
    _fail "staticcheck エラーあり"
  fi
fi

# --- golangci-lint ---
echo ""
echo "--- 統合 Lint (golangci-lint) ---"
if ! command -v golangci-lint &>/dev/null; then
  _skip "golangci-lint が見つからない"
else
  if golangci-lint run ./... 2>&1; then
    _pass "golangci-lint"
  else
    _fail "golangci-lint エラーあり"
  fi
fi

# --- go test ---
echo ""
echo "--- テスト (go test) ---"
if go test -race ./... 2>&1; then
  _pass "go test -race"
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
