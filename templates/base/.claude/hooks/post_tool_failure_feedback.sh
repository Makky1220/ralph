#!/usr/bin/env sh
set -eu

mkdir -p .harness/state
count_file=".harness/state/tool_failures.count"
count=0
if [ -f "$count_file" ]; then
  count="$(cat "$count_file")"
fi
count=$((count + 1))
printf '%s\n' "$count" > "$count_file"

msg="ツール失敗 #${count}。根本原因を調査してください。同じ失敗したアプローチを繰り返さないこと。"
escaped="$(printf '%s' "$msg" | sed 's/"/\\\"/g')"
printf '{"hookSpecificOutput":{"hookEventName":"PostToolUseFailure","additionalContext":"%s"}}\n' "$escaped"
