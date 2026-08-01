#!/bin/bash
# CT-SHELL-001: All shell scripts in the build/test automation must have valid bash syntax.
# Runs `bash -n` (syntax-only parse) on every tracked .sh file that is NOT a template
# (skill templates under .agents/ contain {{ placeholders and are excluded).
# Guards against bugs like an unterminated quote that silently kills a gate at EOF.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT_DIR"

if ! command -v git >/dev/null 2>&1; then
  echo "SKIP: git not available — cannot enumerate tracked .sh files"
  exit 0
fi

# Enumerate tracked shell scripts, excluding agent skill templates (placeholders).
scripts=$(git ls-files "*.sh" | grep -v "^\.agents/" || true)

if [ -z "$scripts" ]; then
  echo "SKIP: no tracked .sh files found"
  exit 0
fi

count=0
errors=0
for script in $scripts; do
  count=$((count + 1))
  if ! bash -n "$script" 2>&1; then
    echo "FAIL: syntax error in $script"
    errors=$((errors + 1))
  fi
done

if [ "$errors" -gt 0 ]; then
  echo "FAILED: $errors shell script(s) have syntax errors (run: bash -n <file>)"
  exit 1
fi

echo "PASS: $count shell scripts have valid bash syntax"
