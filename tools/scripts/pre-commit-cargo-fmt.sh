#!/bin/sh
# Pre-commit hook: cargo fmt for Rust services
# Runs per-service — only in directories with staged Rust changes.
# cargo clippy is reserved for CI (too slow for pre-commit).
set -e

changed=$(git diff --cached --name-only --diff-filter=ACM | grep '\.rs$' | grep '^apps/services/' || true)

if [ -z "$changed" ]; then
  exit 0
fi

# Derive unique Rust service directories from changed files
dirs=$(echo "$changed" | sed 's|^\(apps/services/[^/]*\)/.*|\1|' | sort -u)

for dir in $dirs; do
  if [ -f "$dir/Cargo.toml" ]; then
    echo "==> cargo fmt  $dir"
    (cd "$dir" && cargo fmt)
  fi
done
