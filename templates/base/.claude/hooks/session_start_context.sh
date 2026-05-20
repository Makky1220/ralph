#!/usr/bin/env sh
set -eu

mkdir -p .harness/state .harness/logs docs/plans/active docs/plans/archive docs/reports
printf '%s\n' "0" > .harness/state/tool_failures.count

branch="unknown"
if command -v git >/dev/null 2>&1; then
  branch="$(git rev-parse --abbrev-ref HEAD 2>/dev/null || printf '%s' 'unknown')"
fi

plan="none"
if [ -d docs/plans/active ]; then
  plan="$(find docs/plans/active -maxdepth 1 -type f -name '*.md' 2>/dev/null | sort | tail -n 1)"
  if [ -z "$plan" ]; then
    plan="none"
  fi
fi

msg="ハーネスリマインダー: リスクのある複数ファイル作業には docs/plans/active を使用し、完了宣言の前に ./scripts/run-verify.sh を実行すること。ブランチ: $branch。アクティブプラン: $plan。"
escaped="$(printf '%s' "$msg" | sed 's/"/\\\"/g')"
printf '{"hookSpecificOutput":{"hookEventName":"SessionStart","additionalContext":"%s"}}\n' "$escaped"
