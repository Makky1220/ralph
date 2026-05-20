#!/usr/bin/env sh
set -eu

branch="unknown"
if command -v git >/dev/null 2>&1; then
  branch="$(git rev-parse --abbrev-ref HEAD 2>/dev/null || printf '%s' 'unknown')"
fi

uncommitted=""
if command -v git >/dev/null 2>&1; then
  uncommitted="$(git status --porcelain 2>/dev/null | wc -l | tr -d ' ')"
fi

plan="none"
if [ -d docs/plans/active ]; then
  plan="$(find docs/plans/active -maxdepth 1 -type f -name '*.md' 2>/dev/null | sort | tail -n 1)"
  [ -n "$plan" ] || plan="none"
fi

needs_verify="false"
[ -f .harness/state/needs-verify ] && needs_verify="true"

if [ "${uncommitted:-0}" -gt 0 ] && [ "$branch" != "main" ] && [ "$branch" != "master" ] && [ "$branch" != "unknown" ]; then
  if command -v git >/dev/null 2>&1; then
    git add -A 2>/dev/null || true
    git commit -m 'wip: checkpoint before session end' 2>/dev/null || true
    msg="セッション終了: ブランチ=$branch, WIPコミット作成済み, プラン=$plan"
  else
    msg="セッション終了: ブランチ=$branch, 未コミット変更あり (git なし), プラン=$plan"
  fi
else
  msg="セッション終了: ブランチ=$branch, クリーン, プラン=$plan"
fi

if [ "$needs_verify" = "true" ]; then
  msg="${msg}。verify が未実行です — 次のセッションで run-verify.sh を実行してください。"
fi

escaped="$(printf '%s' "$msg" | sed 's/"/\\\"/g')"
printf '{"hookSpecificOutput":{"hookEventName":"SessionEnd","additionalContext":"%s"}}\n' "$escaped"
