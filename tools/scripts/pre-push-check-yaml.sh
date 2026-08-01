#!/bin/sh
# Pre-push YAML validation for committed files.
# Uses `uv run python` — the project's Python toolchain (uv), which resolves
# to a real interpreter on any machine (bare python3 may hit MS Store stubs
# on Windows).
set -e

files=$(git diff --cached --name-only --diff-filter=ACM | grep -E '\.(yml|yaml)$' | grep -v 'pnpm-lock\.yaml' || true)
[ -z "$files" ] && exit 0

failed=0
for f in $files; do
  uv run python -c "import yaml,sys; yaml.safe_load(open(sys.argv[1]))" "$f" 2>/dev/null || { echo "✖ Invalid YAML: $f"; failed=1; }
done
[ "$failed" = "1" ] && exit 1
exit 0
