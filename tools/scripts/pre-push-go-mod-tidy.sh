#!/bin/sh
# Pre-push Go module hygiene: verifies go.mod/go.sum are tidy in every
# Go module of the monorepo.
#
# A module is "tidy" when `go mod tidy` produces no changes to tracked
# files. Only tracked files are diffed: a module with no external
# dependencies has no go.sum at all, and `git diff` on an absent path
# exits 128, which would otherwise create a false "not tidy" failure.
set -e

failed=0
for dir in $(find apps/services apps/shared apps/vedo-cli -name go.mod -exec dirname {} \;); do
  ok=1
  if ! (cd "$dir" && go mod tidy) > /dev/null 2>&1; then
    echo "✖ go mod tidy failed in $dir"
    ok=0
  fi
  if ! (cd "$dir" && git diff --exit-code -- go.mod) > /dev/null 2>&1; then
    echo "✖ go.mod not tidy in $dir. Run 'go mod tidy' and commit the changes."
    ok=0
  fi
  if git ls-files --error-unmatch -- "$dir/go.sum" > /dev/null 2>&1; then
    if ! (cd "$dir" && git diff --exit-code -- go.sum) > /dev/null 2>&1; then
      echo "✖ go.sum not tidy in $dir. Run 'go mod tidy' and commit the changes."
      ok=0
    fi
  fi
  if [ "$ok" = "0" ]; then
    failed=1
  fi
done
[ "$failed" = "1" ] && exit 1
exit 0
