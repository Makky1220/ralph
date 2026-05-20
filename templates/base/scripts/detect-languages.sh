#!/usr/bin/env sh
# プロジェクトで使われている言語を検出して、対応する言語パックが存在するものを出力する。
set -eu

detected=""

# TypeScript / JavaScript
if find . \( -name "*.ts" -o -name "*.tsx" -o -name "package.json" \) \
     -not -path "./.git/*" -not -path "./node_modules/*" \
     2>/dev/null | head -1 | grep -q .; then
  if [ -d "packs/languages/typescript" ]; then
    detected="$detected typescript"
  fi
fi

# Python
if find . -name "*.py" -not -path "./.git/*" \
     2>/dev/null | head -1 | grep -q .; then
  if [ -d "packs/languages/python" ]; then
    detected="$detected python"
  fi
fi

# Go
if find . -name "*.go" -not -path "./.git/*" \
     2>/dev/null | head -1 | grep -q .; then
  if [ -d "packs/languages/golang" ]; then
    detected="$detected golang"
  fi
fi

# Rust
if find . -name "*.rs" -not -path "./.git/*" \
     2>/dev/null | head -1 | grep -q .; then
  if [ -d "packs/languages/rust" ]; then
    detected="$detected rust"
  fi
fi

printf '%s\n' "$detected" | tr -s ' ' '\n' | grep -v '^$' || true
