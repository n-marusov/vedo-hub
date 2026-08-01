#!/bin/sh
# Pre-push Go module hygiene: verifies go.mod/go.sum are tidy in every
# Go module of the monorepo.
set -e

failed=0
for dir in $(find apps/services apps/shared apps/vedo-cli -name go.mod -exec dirname {} \;); do
  if ! (cd "$dir" && go mod tidy && git diff --exit-code go.mod go.sum > /dev/null 2>&1); then
    echo "✖ go.mod/go.sum not tidy in $dir. Run 'go mod tidy' and commit the changes."
    failed=1
  fi
done
[ "$failed" = "1" ] && exit 1
exit 0
