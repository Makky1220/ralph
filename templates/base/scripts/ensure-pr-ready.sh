#!/usr/bin/env sh
set -eu

usage() {
  echo "Usage: ./scripts/ensure-pr-ready.sh <pr-url-or-branch>"
  exit 1
}

if [ "${1:-}" = "" ]; then
  usage
fi

target="$1"

if command -v gh >/dev/null 2>&1; then
  state="$(gh pr view "$target" --json state -q '.state' 2>/dev/null || echo "UNKNOWN")"
  if [ "$state" = "CLOSED" ]; then
    echo "Error: PR is closed."
    exit 1
  fi
  draft="$(gh pr view "$target" --json isDraft -q '.isDraft' 2>/dev/null || echo "false")"
  if [ "$draft" = "true" ]; then
    echo "Warning: PR is still a draft."
  fi
  echo "PR state: $state"
else
  echo "gh CLI not available — skipping PR state check."
fi
