#!/usr/bin/env sh
set -eu

branch="unknown"
if command -v git >/dev/null 2>&1; then
  branch="$(git rev-parse --abbrev-ref HEAD 2>/dev/null || printf '%s' 'unknown')"
fi

plan="none"
if [ -d docs/plans/active ]; then
  plan="$(find docs/plans/active -maxdepth 1 -type f -name '*.md' 2>/dev/null | sort | tail -n 1)"
  [ -n "$plan" ] || plan="none"
fi

needs_verify="false"
[ -f .harness/state/needs-verify ] && needs_verify="true"

if [ "$needs_verify" = "true" ]; then
  msg="コンパクション前チェックポイント: ブランチ=$branch, プラン=$plan。未コミットの変更がある場合は WIP コミットを作成してください（feature ブランチのみ）。verify が未実行です。"
else
  msg="コンパクション前チェックポイント: ブランチ=$branch, プラン=$plan。"
fi

escaped="$(printf '%s' "$msg" | sed 's/"/\\\"/g')"
printf '{"hookSpecificOutput":{"hookEventName":"PreCompact","additionalContext":"%s"}}\n' "$escaped"
