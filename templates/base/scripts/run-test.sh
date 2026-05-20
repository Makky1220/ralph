#!/usr/bin/env sh
# テストのみ実行（静的解析は含まない）。
# プロジェクト固有のテストコマンドで上書きするには ./scripts/test.local.sh を作成する。
set -eu

mkdir -p .harness/state docs/evidence

ts="$(date -u '+%Y-%m-%dT%H:%M:%SZ')"
evidence_file="docs/evidence/test-$(date -u '+%Y-%m-%d-%H%M%S').log"
status_file=".harness/state/test-exit-code"

{
  ran_any=0
  status=0

  echo "# Test run"
  echo "- Timestamp: $ts"
  echo ""

  if [ -x ./scripts/test.local.sh ]; then
    echo "==> Running local test runner"
    ran_any=1
    if ! ./scripts/test.local.sh; then
      status=1
    fi
  fi

  languages=""
  if [ -x ./scripts/detect-languages.sh ]; then
    languages="$(./scripts/detect-languages.sh 2>/dev/null || true)"
  fi

  for lang in $languages; do
    runner="packs/languages/$lang/test.sh"
    if [ -x "$runner" ]; then
      echo "==> Running $lang test runner"
      ran_any=1
      if ! "$runner"; then
        status=1
      fi
    fi
  done

  if [ "$ran_any" -eq 0 ]; then
    echo "テストランナーが見つかりませんでした。"
    echo "./scripts/test.local.sh を作成するか、packs/languages/<name>/test.sh を追加してください。"
    status=2
  else
    echo ""
    if [ "$status" -eq 0 ]; then
      echo "==> すべてのテストがパスしました。"
    else
      echo "==> 一部のテストが失敗しました。"
    fi
  fi

  printf '%s' "$status" > "$status_file"
} 2>&1 | tee "$evidence_file"

echo ""
echo "Evidence saved to: $evidence_file"

test_status=0
if [ -f "$status_file" ]; then
  test_status="$(cat "$status_file")"
  rm -f "$status_file"
fi

exit "$test_status"
