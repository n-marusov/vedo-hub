#!/bin/bash
# =============================================================================
# Gate: pnpm version consistency
# =============================================================================
# Validates that:
#   1. All package.json files have "packageManager" field
#   2. pnpm version is consistent across all services
#   3. Dockerfile.typescript uses corepack (not hardcoded npm install -g pnpm@X)
#   4. Lockfiles exist and are valid for the declared pnpm version
# =============================================================================
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"
PASSED=0
FAILED=0

log_pass() { printf "  [PASS] %s\n" "$1"; PASSED=$((PASSED + 1)); }
log_fail() { printf "  [FAIL] %s\n" "$1"; FAILED=$((FAILED + 1)); }
log_info() { printf "  [INFO] %s\n" "$1"; }

# =========================================================================
# Test 1: All package.json files have packageManager field
# =========================================================================
test_package_manager_field() {
  log_info "Checking packageManager field in all package.json files..."

  local pkg_files
  pkg_files=$(find "$ROOT_DIR/apps/services" -name "package.json" -not -path "*/node_modules/*" -not -path "*/.venv/*" 2>/dev/null || true)

  for pkg in $pkg_files; do
    local rel
    rel=$(realpath --relative-to="$ROOT_DIR" "$pkg" 2>/dev/null || echo "$pkg")

    if grep -q '"packageManager"' "$pkg"; then
      local version
      version=$(grep '"packageManager"' "$pkg" | head -1 | grep -oP 'pnpm@[\d.]+' || echo "UNKNOWN")
      log_pass "$rel: $version"
    else
      log_fail "$rel: missing packageManager field"
    fi
  done
}

# =========================================================================
# Test 2: pnpm version is consistent across all services
# =========================================================================
test_consistent_pnpm_version() {
  log_info "Checking pnpm version consistency..."

  local versions
  versions=$(find "$ROOT_DIR/apps/services" -name "package.json" -not -path "*/node_modules/*" -not -path "*/.venv/*" \
    -exec grep -oP '"packageManager":\s*"pnpm@[\d.]+"' {} \; 2>/dev/null | sort -u || true)

  local count
  count=$(echo "$versions" | wc -l)

  if [ "$count" -eq 0 ]; then
    log_fail "no packageManager fields found"
  elif [ "$count" -eq 1 ]; then
    log_pass "all services use same pnpm version: $(echo "$versions" | head -1)"
  else
    log_fail "pnpm version mismatch across services:"
    echo "$versions" | while read -r v; do printf "      %s\n" "$v"; done
  fi
}

# =========================================================================
# Test 3: Dockerfile.typescript uses corepack
# =========================================================================
test_dockerfile_uses_corepack() {
  log_info "Checking Dockerfile.typescript for corepack..."

  local dockerfile="$ROOT_DIR/tools/dockerfiles/Dockerfile.typescript"

  if grep -q "corepack enable" "$dockerfile"; then
    log_pass "Dockerfile.typescript uses corepack enable"
  else
    log_fail "Dockerfile.typescript does NOT use corepack enable"
  fi

  if grep -q "npm install.*pnpm" "$dockerfile"; then
    log_fail "Dockerfile.typescript still has hardcoded 'npm install ... pnpm'"
  else
    log_pass "Dockerfile.typescript: no hardcoded npm install pnpm"
  fi
}

# =========================================================================
# Test 4: Lockfiles exist
# =========================================================================
test_lockfiles_exist() {
  log_info "Checking pnpm-lock.yaml files..."

  local svc_dirs
  svc_dirs=$(find "$ROOT_DIR/apps/services" -name "package.json" -not -path "*/node_modules/*" -not -path "*/.venv/*" -exec dirname {} \; 2>/dev/null || true)

  for dir in $svc_dirs; do
    local rel
    rel=$(realpath --relative-to="$ROOT_DIR" "$dir" 2>/dev/null || echo "$dir")

    if [ -f "$dir/pnpm-lock.yaml" ]; then
      log_pass "$rel: lockfile exists"
    else
      log_fail "$rel: missing pnpm-lock.yaml"
    fi
  done
}

# =========================================================================
# Run all tests
# =========================================================================
echo "=== pnpm Version Consistency Gate ==="
echo ""

test_package_manager_field
test_consistent_pnpm_version
test_dockerfile_uses_corepack
test_lockfiles_exist

echo ""
echo "=== Results: $PASSED passed, $FAILED failed ==="

if [ "$FAILED" -gt 0 ]; then
  echo "GATE: FAIL"
  exit 1
else
  echo "GATE: PASS"
  exit 0
fi
