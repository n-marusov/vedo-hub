#!/bin/sh
# Pre-commit hook: gofmt + go vet for Go services
# Runs per-service — only in directories with staged Go changes.
set -e

changed=$(git diff --cached --name-only --diff-filter=ACM | grep '\.go$' | grep '^src/services/' || true)

if [ -z "$changed" ]; then
  exit 0
fi

# Derive unique Go service directories from changed files
dirs=$(echo "$changed" | sed 's|^\(src/services/[^/]*\)/.*|\1|' | sort -u)

for dir in $dirs; do
  if [ -f "$dir/go.mod" ]; then
    echo "==> go fmt  $dir"
    (cd "$dir" && go fmt ./...)
    echo "==> go vet  $dir"
    (cd "$dir" && go vet ./...)
  fi
done
