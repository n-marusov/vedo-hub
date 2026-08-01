#!/bin/sh
# Pre-push JSON validation for committed files.
# Uses `uv run python` — the project's Python toolchain (uv), which resolves
# to a real interpreter on any machine (bare python3 may hit MS Store stubs
# on Windows).
set -e

files=$(git diff --cached --name-only --diff-filter=ACM | grep '\.json$' | grep -v 'pnpm-lock\.yaml' || true)
[ -z "$files" ] && exit 0

failed=0
for f in $files; do
  uv run python -m json.tool "$f" > /dev/null 2>&1 || { echo "✖ Invalid JSON: $f"; failed=1; }
done
[ "$failed" = "1" ] && exit 1
exit 0
