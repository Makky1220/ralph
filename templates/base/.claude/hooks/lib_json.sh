#!/usr/bin/env sh

extract_json_field() {
  _json="$1"
  _field="$2"
  printf '%s' "$_json" | python3 -c "
import sys, json
try:
    d = json.loads(sys.stdin.read())
    parts = '$_field'.split('.')
    v = d
    for p in parts:
        v = v[p]
    print(v)
except Exception:
    pass
" 2>/dev/null || true
}
