#!/bin/sh
# Pre-commit hook: golangci-lint for Go services
# Runs per-service — only in directories with staged Go changes.
# Uses per-service .golangci.yml configuration.
set -e

changed=$(git diff --cached --name-only --diff-filter=ACM | grep '\.go$' | grep '^src/services/' || true)

if [ -z "$changed" ]; then
  exit 0
fi

# Derive unique Go service directories from changed files
dirs=$(echo "$changed" | sed 's|^\(src/services/[^/]*\)/.*|\1|' | sort -u)

for dir in $dirs; do
  if [ -f "$dir/go.mod" ] && [ -f "$dir/.golangci.yml" ]; then
    echo "==> golangci-lint  $dir"
    (cd "$dir" && golangci-lint run --timeout 5m)
  fi
done
