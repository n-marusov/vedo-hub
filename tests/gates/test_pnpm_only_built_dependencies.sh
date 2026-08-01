#!/bin/bash
# Gate: pnpm onlyBuiltDependencies — verifies that .npmrc files declare
# allowed build-script packages for all TypeScript frontend services.
# Without these entries, pnpm blocks native binary installs and build fails
# with ERR_PNPM_IGNORED_BUILDS.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"

# Per-service list of packages that require build-script approval.
# These packages ship native binaries or need postinstall hooks.
declare -A REQUIRED_BUILT_DEPS
REQUIRED_BUILT_DEPS["frontend"]="@biomejs/biome esbuild lefthook vue-demi"
REQUIRED_BUILT_DEPS["publish-browse-ui"]="@biomejs/biome esbuild"

TS_SERVICES=(
  "frontend"
  "publish-browse-ui"
)

test_frontend_npmrc_has_only_built_dependencies() {
  echo "TEST: frontend/.npmrc declares all required onlyBuiltDependencies"

  local file="$ROOT_DIR/apps/services/frontend/.npmrc"

  if [ ! -f "$file" ]; then
    echo "FAIL: $file not found"
    return 1
  fi

  local failed=0
  for dep in ${REQUIRED_BUILT_DEPS["frontend"]}; do
    if ! grep -qF "onlyBuiltDependencies[]=${dep}" "$file"; then
      echo "  FAIL: missing onlyBuiltDependencies[]=${dep} in $file"
      failed=$((failed + 1))
    fi
  done

  if [ "$failed" -gt 0 ]; then
    echo "  Run: pnpm approve-builds   or   add missing entries to .npmrc"
    return 1
  fi

  echo "PASS: all required deps present — ${REQUIRED_BUILT_DEPS["frontend"]}"
}

test_publish_browse_ui_npmrc_has_only_built_dependencies() {
  echo "TEST: publish-browse-ui/.npmrc declares all required onlyBuiltDependencies"

  local file="$ROOT_DIR/apps/services/publish-browse-ui/.npmrc"

  if [ ! -f "$file" ]; then
    echo "FAIL: $file not found"
    return 1
  fi

  local failed=0
  for dep in ${REQUIRED_BUILT_DEPS["publish-browse-ui"]}; do
    if ! grep -qF "onlyBuiltDependencies[]=${dep}" "$file"; then
      echo "  FAIL: missing onlyBuiltDependencies[]=${dep} in $file"
      failed=$((failed + 1))
    fi
  done

  if [ "$failed" -gt 0 ]; then
    return 1
  fi

  echo "PASS: all required deps present — ${REQUIRED_BUILT_DEPS["publish-browse-ui"]}"
}

test_npmrc_exists_for_all_ts_services() {
  echo "TEST: .npmrc exists for every TypeScript service"

  local failed=0
  for svc in "${TS_SERVICES[@]}"; do
    local file="$ROOT_DIR/apps/services/$svc/.npmrc"
    if [ ! -f "$file" ]; then
      echo "  FAIL: missing $file"
      failed=$((failed + 1))
    fi
  done

  if [ "$failed" -gt 0 ]; then
    return 1
  fi

  echo "PASS: all ${#TS_SERVICES[@]} services have .npmrc"
}

test_only_built_dependencies_have_no_duplicates() {
  echo "TEST: onlyBuiltDependencies entries have no duplicates"

  local failed=0
  for svc in "${TS_SERVICES[@]}"; do
    local file="$ROOT_DIR/apps/services/$svc/.npmrc"
    if [ ! -f "$file" ]; then
      continue
    fi

    local duplicates
    duplicates=$(grep -oE '^onlyBuiltDependencies\[\]=.+$' "$file" | sort | uniq -d)
    if [ -n "$duplicates" ]; then
      echo "  FAIL: duplicate onlyBuiltDependencies entries in $file:"
      echo "$duplicates"
      failed=$((failed + 1))
    fi
  done

  if [ "$failed" -gt 0 ]; then
    return 1
  fi

  echo "PASS: no duplicate entries"
}

test_pnpm_workspace_allow_builds_are_valid_booleans() {
  echo "TEST: pnpm-workspace.yaml allowBuilds has valid boolean values"

  local failed=0
  for svc in "${TS_SERVICES[@]}"; do
    local file="$ROOT_DIR/apps/services/$svc/pnpm-workspace.yaml"
    if [ ! -f "$file" ]; then
      continue
    fi

    # Check for non-boolean placeholder values in allowBuilds section.
    # Valid values are 'true' or 'false'. Invalid values include
    # placeholder strings like "set this to true or false".
    local invalid
    invalid=$(grep -E '^  ["'\'']?[a-zA-Z@].*: (.+)$' "$file" | \
      grep -vE ': (true|false)$' | \
      grep -vE '^allowBuilds:' || true)

    # More precise: extract allowBuilds section and check each value
    if grep -q '^allowBuilds:' "$file"; then
      local in_allow_builds=0
      while IFS= read -r line; do
        if [ "$in_allow_builds" -eq 1 ]; then
          # Stop at next top-level key
          if echo "$line" | grep -qE '^[a-zA-Z]'; then
            break
          fi
          # Check indented key: value
          if echo "$line" | grep -qE '^  .*: (.+)$'; then
            local val
            val=$(echo "$line" | sed -E 's/^  .*: //')
            if [ "$val" != "true" ] && [ "$val" != "false" ]; then
              echo "  FAIL: $svc/pnpm-workspace.yaml — invalid allowBuilds value: $line"
              failed=$((failed + 1))
            fi
          fi
        fi
        if echo "$line" | grep -q '^allowBuilds:'; then
          in_allow_builds=1
        fi
      done < "$file"
    fi
  done

  if [ "$failed" -gt 0 ]; then
    echo "  Hint: Change placeholder values to 'true' or 'false'"
    return 1
  fi

  echo "PASS: all allowBuilds values are valid booleans"
}

test_dockerfile_copies_pnpm_workspace_yaml() {
  echo "TEST: Dockerfile.typescript copies pnpm-workspace.yaml into image"

  local file="$ROOT_DIR/tools/dockerfiles/Dockerfile.typescript"

  if [ ! -f "$file" ]; then
    echo "FAIL: $file not found"
    return 1
  fi

  # Without pnpm-workspace.yaml in the image, --frozen-lockfile fails with
  # ERR_PNPM_LOCKFILE_CONFIG_MISMATCH because the lockfile records overrides
  # from the workspace config.
  if ! grep -q 'pnpm-workspace.yaml' "$file"; then
    echo "  FAIL: $file does not COPY pnpm-workspace.yaml"
    echo "  Hint: The lockfile records overrides from pnpm-workspace.yaml;"
    echo "        the image must contain it for --frozen-lockfile to match."
    return 1
  fi

  echo "PASS: Dockerfile.typescript copies pnpm-workspace.yaml"
}

run_all() {
  local failed=0

  for test_fn in $(declare -F | awk '{print $3}' | grep '^test_'); do
    local rc=0
    "$test_fn" || rc=$?
    if [ "$rc" -ne 0 ]; then
      echo "FAIL: $test_fn (exit $rc)"
      failed=$((failed + 1))
    else
      echo ""
    fi
  done

  echo "=== Results: $failed failures ==="
  return $failed
}

run_all
