#!/bin/bash
# =============================================================================
# Gate: Runtime config secrets check
# =============================================================================
# Validates that:
#   1. docker-compose config resolves without errors
#   2. VEDO_* env vars (used for frontend config.js) contain NO secrets
#   3. Entrypoint scripts only expose whitelisted non-sensitive vars
#   4. No hardcoded default passwords appear in compose config output
#
# Security boundary: /config.js is served publicly by nginx.
# It MUST only contain non-sensitive configuration.
# =============================================================================
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"
PASSED=0
FAILED=0

# ── Sensitive patterns that MUST NOT appear in VEDO_* env vars ──────────
# Patterns are case-insensitive. Add new patterns as new attack vectors emerge.
SENSITIVE_PATTERNS=(
  "password"
  "secret"
  "api_key"
  "apikey"
  "token"
  "private_key"
  "privatekey"
  "credential"
  "jwt_secret"
  "signing_key"
)

# ── Whitelist: allowed VEDO_* variable names ────────────────────────────
# Only these VEDO_* vars may appear in docker-compose environment blocks.
# Adding a new var requires a security review (update this whitelist).
ALLOWED_VEDO_VARS=(
  "VEDO_KEYCLOAK_URL"
  "VEDO_KEYCLOAK_REALM"
  "VEDO_KEYCLOAK_CLIENT_ID"
  "VEDO_SKIP_AUTH"
  "VEDO_GRAPHQL_ENDPOINT"
  "VEDO_USE_MOCK_API"
  "VEDO_APP_VERSION"
  "VEDO_APP_ENV"
)

# ── Whitelist: values exposed in entrypoint heredocs ────────────────────
# These are the keys that appear in window.__VEDO_CONFIG__ inside
# docker-entrypoint.sh. Must match ALLOWED_VEDO_VARS (without VEDO_ prefix).
EXPOSED_CONFIG_KEYS=(
  "KEYCLOAK_URL"
  "KEYCLOAK_REALM"
  "KEYCLOAK_CLIENT_ID"
  "SKIP_AUTH"
  "GRAPHQL_ENDPOINT"
  "USE_MOCK_API"
  "APP_VERSION"
  "APP_ENV"
)

log_pass() { printf "  [PASS] %s\n" "$1"; PASSED=$((PASSED + 1)); }
log_fail() { printf "  [FAIL] %s\n" "$1"; FAILED=$((FAILED + 1)); }
log_info() { printf "  [INFO] %s\n" "$1"; }

# =========================================================================
# Test 1: docker compose config resolves for all environments
# =========================================================================
test_compose_config_resolves() {
  log_info "Checking compose config resolution for all environments..."

  for env in dev test staging; do
    local env_file="$ROOT_DIR/config/.env.$env"
    if [ ! -f "$env_file" ]; then
      log_info "  env file $env_file not found — skipping $env check"
      continue
    fi
    if docker compose --env-file "$env_file" -f "$ROOT_DIR/deploy/docker-compose.yml" config > /dev/null 2>&1; then
      log_pass "compose config $env resolves"
    else
      log_fail "compose config $env fails to resolve"
    fi
  done
}

# =========================================================================
# Test 2: VEDO_* env vars contain no secrets
# =========================================================================
test_vedo_vars_no_secrets() {
  log_info "Checking VEDO_* vars for sensitive patterns..."

  for env in dev test staging; do
    local env_file="$ROOT_DIR/config/.env.$env"
    if [ ! -f "$env_file" ]; then continue; fi

    local config_output
    config_output=$(docker compose --env-file "$env_file" -f "$ROOT_DIR/deploy/docker-compose.yml" config 2>/dev/null || true)

    # Extract VEDO_* values from compose config
    local vedo_values
    vedo_values=$(echo "$config_output" | grep -oP 'VEDO_[A-Z_]+:\s*\K.*' || true)

    for pattern in "${SENSITIVE_PATTERNS[@]}"; do
      if echo "$vedo_values" | grep -qi "$pattern"; then
        local matched
        matched=$(echo "$vedo_values" | grep -i "$pattern" | head -3)
        log_fail "VEDO_* vars in $env contain sensitive pattern '$pattern': $matched"
        return
      fi
    done

    log_pass "VEDO_* vars in $env: no secrets detected"
  done
}

# =========================================================================
# Test 3: VEDO_* variable names are whitelisted
# =========================================================================
test_vedo_var_names_whitelisted() {
  log_info "Checking VEDO_* variable names against whitelist..."

  for compose_file in "$ROOT_DIR/deploy/docker-compose.yml" "$ROOT_DIR/deploy/docker-compose.test.yml"; do
    [ ! -f "$compose_file" ] && continue
    local filename
    filename=$(basename "$compose_file")

    # Extract VEDO_* variable names from compose files
    local vedo_names
    vedo_names=$(grep -oP 'VEDO_[A-Z_]+' "$compose_file" | sort -u || true)

    for vname in $vedo_names; do
      local allowed=false
      for allowed_name in "${ALLOWED_VEDO_VARS[@]}"; do
        if [ "$vname" = "$allowed_name" ]; then
          allowed=true
          break
        fi
      done
      if [ "$allowed" = false ]; then
        log_fail "$filename: VEDO_* var '$vname' is NOT in the security whitelist"
      fi
    done
  done

  log_pass "VEDO_* var names in compose files: all whitelisted"
}

# =========================================================================
# Test 4: Entrypoint heredocs only expose whitelisted keys
# =========================================================================
test_entrypoint_keys_whitelisted() {
  log_info "Checking entrypoint scripts for exposed config keys..."

  for entrypoint in \
    "$ROOT_DIR/apps/services/frontend/docker-entrypoint.sh" \
    "$ROOT_DIR/apps/services/publish-browse-ui/docker-entrypoint.sh"; do

    [ ! -f "$entrypoint" ] && continue
    local filename
    filename=$(basename "$(dirname "$entrypoint")")/$(basename "$entrypoint")

    # Extract keys from window.__VEDO_CONFIG__ block in the heredoc
    local exposed_keys
    exposed_keys=$(sed -n '/window\.__VEDO_CONFIG__/,/};/p' "$entrypoint" | grep -oP '^\s+\w+' | tr -d ' ' | sort -u || true)

    for key in $exposed_keys; do
      # Skip JS keywords and structural tokens
      case "$key" in
        window|__VEDO_CONFIG__|APP_VERSION) continue ;;
      esac

      local key_allowed=false
      for allowed_key in "${EXPOSED_CONFIG_KEYS[@]}"; do
        if [ "$key" = "$allowed_key" ]; then
          key_allowed=true
          break
        fi
      done
      if [ "$key_allowed" = false ]; then
        log_fail "$filename: exposes key '$key' not in security whitelist"
      fi
    done
  done

  log_pass "entrypoint config keys: all whitelisted"
}

# =========================================================================
# Test 5: No hardcoded default passwords in compose environment blocks
# =========================================================================
test_no_default_passwords_in_compose() {
  log_info "Checking compose for hardcoded default credentials..."

  local compose_file="$ROOT_DIR/deploy/docker-compose.yml"

  # Check for common default password patterns in environment blocks
  # (password, admin, root without variable substitution)
  local defaults_found
  defaults_found=$(grep -nE ':\s*(password|admin|root|guest)\s*$' "$compose_file" | grep -v '#' | grep -v 'KEYCLOAK_ADMIN\b' || true)

  if [ -n "$defaults_found" ]; then
    log_fail "docker-compose.yml contains hardcoded default credentials:"
    echo "$defaults_found" | while read -r line; do
      printf "    %s\n" "$line"
    done
  else
    log_pass "no hardcoded default credentials in compose environment"
  fi
}

# =========================================================================
# Run all tests
# =========================================================================
echo "=== Runtime Config Secrets Gate ==="
echo ""

test_compose_config_resolves
test_vedo_vars_no_secrets
test_vedo_var_names_whitelisted
test_entrypoint_keys_whitelisted
test_no_default_passwords_in_compose

echo ""
echo "=== Results: $PASSED passed, $FAILED failed ==="

if [ "$FAILED" -gt 0 ]; then
  echo "GATE: FAIL"
  exit 1
else
  echo "GATE: PASS"
  exit 0
fi
