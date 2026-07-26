#!/bin/sh
# Pre-commit hook: cargo clippy for Rust services
# Runs per-service — only in directories with staged Rust changes.
# Uses per-service [lints.clippy] in Cargo.toml.
set -e

changed=$(git diff --cached --name-only --diff-filter=ACM | grep '\.rs$' | grep '^src/services/' || true)

if [ -z "$changed" ]; then
  exit 0
fi

# Derive unique Rust service directories from changed files
dirs=$(echo "$changed" | sed 's|^\(src/services/[^/]*\)/.*|\1|' | sort -u)

for dir in $dirs; do
  if [ -f "$dir/Cargo.toml" ]; then
    echo "==> cargo clippy  $dir"
    (cd "$dir" && cargo clippy -- -D warnings)
  fi
done
