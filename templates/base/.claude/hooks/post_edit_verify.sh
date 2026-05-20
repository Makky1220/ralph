#!/usr/bin/env sh
set -eu

mkdir -p .harness/state
printf '%s\n' "1" > .harness/state/needs-verify

printf '{"hookSpecificOutput":{"hookEventName":"PostToolUse","additionalContext":"ファイルを編集しました。スライスが完了したら ./scripts/run-verify.sh を実行してください。"}}\n'
