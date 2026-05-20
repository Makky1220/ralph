#!/usr/bin/env sh
set -eu

mkdir -p .harness/state .harness/logs docs/evidence

ts="$(date -u '+%Y-%m-%dT%H:%M:%SZ')"
evidence_file="docs/evidence/verify-$(date -u '+%Y-%m-%d-%H%M%S').log"
status_file=".harness/state/verify-exit-code"

{
  ran_any=0
  status=0

  echo "# Verification run"
  echo "- Timestamp: $ts"
  echo ""

  if [ -x ./scripts/verify.local.sh ]; then
    echo "==> Running local verifier"
    ran_any=1
    if ! ./scripts/verify.local.sh; then
      status=1
    fi
  fi

  languages=""
  if [ -x ./scripts/detect-languages.sh ]; then
    languages="$(./scripts/detect-languages.sh 2>/dev/null || true)"
  fi

  if [ -n "$languages" ]; then
    echo "==> Language packs detected: $languages"
  fi

  for lang in $languages; do
    verifier="packs/languages/$lang/verify.sh"
    if [ -x "$verifier" ]; then
      echo "==> Running $lang verifier"
      ran_any=1
      if ! "$verifier"; then
        status=1
      fi
    fi
  done

  changed_files=""
  if command -v git >/dev/null 2>&1; then
    changed_files="$( (git diff --name-only 2>/dev/null; git diff --name-only --cached 2>/dev/null) | sort -u )"
  fi

  docs_only=1
  if [ -n "$changed_files" ]; then
    printf '%s\n' "$changed_files" | while IFS= read -r file; do
      case "$file" in
        ""|docs/*|README.md|CLAUDE.md|.claude/*)
          ;;
        *)
          echo "$file" > .harness/state/non_docs_change
          ;;
      esac
    done
    if [ -f .harness/state/non_docs_change ]; then
      docs_only=0
      rm -f .harness/state/non_docs_change
    fi
  fi

  if [ "$ran_any" -eq 0 ]; then
    if [ "$docs_only" -eq 1 ]; then
      echo "ベリファイアが実行されませんでした。ドキュメントのみの変更のようです。"
    else
      echo "コードの変更に対するベリファイアが実行されませんでした。"
      echo "./scripts/verify.local.sh または packs/languages/<name>/verify.sh に実際のベリファイアを追加してください。"
      status=2
    fi
  else
    echo ""
    if [ "$status" -eq 0 ]; then
      echo "==> すべてのベリファイアがパスしました。"
    else
      echo "==> 一部のベリファイアが失敗しました。"
    fi
  fi

  printf '%s' "$status" > "$status_file"
} 2>&1 | tee "$evidence_file"

echo ""
echo "Evidence saved to: $evidence_file"

verify_status=0
if [ -f "$status_file" ]; then
  verify_status="$(cat "$status_file")"
  rm -f "$status_file"
fi

if [ "$verify_status" = "0" ] && [ -f .harness/state/needs-verify ]; then
  rm -f .harness/state/needs-verify
fi

exit "$verify_status"
