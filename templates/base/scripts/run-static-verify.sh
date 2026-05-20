#!/usr/bin/env sh
# 静的解析のみ実行（テストは含まない）。
set -eu
HARNESS_VERIFY_MODE=static exec ./scripts/run-verify.sh "$@"
