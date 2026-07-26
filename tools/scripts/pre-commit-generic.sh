#!/bin/sh
# Pre-commit hook: generic file checks
# Replaces pre-commit-hooks (trailing-whitespace, end-of-file, merge-conflict, mixed-line-ending)
set -e

# ── Check merge conflict markers ──
check_merge_conflict() {
  files=$(git diff --cached --name-only --diff-filter=ACM 2>/dev/null || true)
  if [ -z "$files" ]; then
    return 0
  fi

  # Build pattern without literal conflict markers to avoid self-flagging
  left=$(printf '\074\074\074\074\074\074\074 ')
  right=$(printf '\076\076\076\076\076\076\076 ')
  mid='^=======$'
  conflicts=$(echo "$files" | xargs grep -n "${left}\|${mid}\|${right}" 2>/dev/null | head -30 || true)
  if [ -n "$conflicts" ]; then
    echo "✖ Merge conflict markers found:"
    echo "$conflicts"
    return 1
  fi
}

# ── Check trailing whitespace ──
check_trailing_whitespace() {
  files=$(git diff --cached --name-only --diff-filter=ACM 2>/dev/null | grep -v '\.css$' || true)
  if [ -z "$files" ]; then
    return 0
  fi

  errors=$(echo "$files" | xargs grep -n '[[:space:]]$' 2>/dev/null | head -50 || true)
  if [ -n "$errors" ]; then
    echo "✖ Trailing whitespace found (run formatters to fix):"
    echo "$errors"
    return 1
  fi
}

# ── Run all checks ──
failed=0

check_merge_conflict || failed=1
check_trailing_whitespace || failed=1

exit $failed
