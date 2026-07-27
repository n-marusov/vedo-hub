#!/bin/bash
# Native service layout and production-readiness checks
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"
SERVICES_DIR="$ROOT_DIR/apps/services"
DOCKERFILES_DIR="$ROOT_DIR/tools/dockerfiles"

go_services=(api-gateway auth-service commenting-service ticket-api ticket-telemetry-listener ticket-notifier)
rust_services=(ontology-service versioning-service publisher-service public-browse-api)
python_services=(metrics-service ticket-classifier document-extractor)
ts_services=(frontend publish-browse-ui)

failures=0

check() {
  local label="$1"
  shift
  if "$@"; then
    echo "PASS: $label"
  else
    echo "FAIL: $label"
    failures=$((failures + 1))
  fi
}

check_go_base_images() {
  test -f "$DOCKERFILES_DIR/Dockerfile.go" && \
    grep -Eq 'FROM \$\{GO_BASE_IMAGE\}' "$DOCKERFILES_DIR/Dockerfile.go" && \
    grep -Eq 'FROM \$\{GO_RUNTIME_IMAGE\}' "$DOCKERFILES_DIR/Dockerfile.go"
}

check_rust_base_images() {
  test -f "$DOCKERFILES_DIR/Dockerfile.rust" && \
    grep -Eq 'FROM \$\{RUST_BASE_IMAGE\}' "$DOCKERFILES_DIR/Dockerfile.rust" && \
    grep -Eq 'FROM \$\{RUST_RUNTIME_IMAGE\}' "$DOCKERFILES_DIR/Dockerfile.rust"
}

check_python_base_images() {
  test -f "$DOCKERFILES_DIR/Dockerfile.python" && \
    grep -Eq 'FROM \$\{PYTHON_IMAGE\}' "$DOCKERFILES_DIR/Dockerfile.python" && \
    grep -Eq 'FROM \$\{PYTHON_RUNTIME_IMAGE\}' "$DOCKERFILES_DIR/Dockerfile.python"
}

check_typescript_base_images() {
  test -f "$DOCKERFILES_DIR/Dockerfile.typescript" && \
    grep -Eq 'FROM \$\{NODE_IMAGE\}' "$DOCKERFILES_DIR/Dockerfile.typescript" && \
    grep -Eq 'FROM \$\{NGINX_IMAGE\}' "$DOCKERFILES_DIR/Dockerfile.typescript"
}

check_language_manifests() {
  for svc in "${go_services[@]}"; do
    test -f "$SERVICES_DIR/$svc/go.mod"
  done
  for svc in "${rust_services[@]}"; do
    test -f "$SERVICES_DIR/$svc/Cargo.toml"
  done
  for svc in "${python_services[@]}"; do
    test -f "$SERVICES_DIR/$svc/pyproject.toml" && \
      test -f "$SERVICES_DIR/$svc/uv.lock"
  done
  for svc in "${ts_services[@]}"; do
    test -f "$SERVICES_DIR/$svc/package.json"
  done
}

check_entrypoints_exist() {
  for svc in "${go_services[@]}"; do
    test -f "$SERVICES_DIR/$svc/main.go"
  done
  for svc in "${rust_services[@]}"; do
    test -f "$SERVICES_DIR/$svc/src/main.rs"
  done
  for svc in "${python_services[@]}"; do
    test -f "$SERVICES_DIR/$svc/main.py"
  done
  for svc in "${ts_services[@]}"; do
    test -f "$SERVICES_DIR/$svc/src/main.ts"
    test -f "$SERVICES_DIR/$svc/nginx.conf"
  done
}

check_compose_builds_native_services() {
  local compose="$ROOT_DIR/deploy/docker-compose.yml"
  local built_services=(api-gateway auth-service ontology-service versioning-service metrics-service frontend)
  test -f "$compose" || return 1
  grep -q 'dockerfile: tools/dockerfiles/Dockerfile\.' "$compose" || return 1
  for svc in "${built_services[@]}"; do
    grep -q "^  ${svc}:" "$compose" || return 1
  done
}

# CT-PYTHON-001: All Python services defined in docker-compose must have uv manifests
check_compose_python_services_have_manifests() {
  local compose="$ROOT_DIR/deploy/docker-compose.yml"
  test -f "$compose" || return 1

  local current_svc=""
  local errors=0

  # Use fd 3 to avoid nested read conflicts
  while IFS= read -r line <&3; do
    # Capture service name line (e.g. "  ticket-classifier:")
    if echo "$line" | grep -qP '^  [a-zA-Z0-9_-]+:$'; then
      current_svc=$(echo "$line" | sed 's/^  \(.*\):$/\1/')
    fi

    # When a Python dockerfile is found, peek ahead for SERVICE_DIR
    if echo "$line" | grep -q 'dockerfile: tools/dockerfiles/Dockerfile\.python'; then
      local svc_dir=""
      # Peek next up to 3 lines for SERVICE_DIR (structure: dockerfile, args:, SERVICE_DIR)
      for _ in 1 2 3; do
        IFS= read -r peek <&3 || break
        if echo "$peek" | grep -qP '^\s+SERVICE_DIR:\s+'; then
          svc_dir=$(echo "$peek" | sed 's/.*SERVICE_DIR:[[:space:]]*//')
          break
        fi
      done

      if [ -n "$svc_dir" ]; then
        local abs_dir="$ROOT_DIR/$svc_dir"
        if [ ! -f "$abs_dir/pyproject.toml" ]; then
          echo "  FAIL: $current_svc (SERVICE_DIR=$svc_dir): missing pyproject.toml"
          errors=$((errors + 1))
        fi
        if [ ! -f "$abs_dir/uv.lock" ]; then
          echo "  FAIL: $current_svc (SERVICE_DIR=$svc_dir): missing uv.lock"
          errors=$((errors + 1))
        fi
      fi
    fi
  done 3< "$compose"

  [ "$errors" -eq 0 ]
}

check_ts_nginx_method_guard_syntax() {
  for svc in "${ts_services[@]}"; do
    local cfg="$SERVICES_DIR/$svc/nginx.conf"
    ! grep -q 'limit_except GET { return 405; }' "$cfg"
    grep -q 'if ($request_method != GET) { return 405; }' "$cfg"
  done
}

check_no_stub_version_in_sources() {
  ! grep -rq '0\.1\.0-stub' "$SERVICES_DIR" --include='*.go' --include='*.rs' --include='*.py' --include='*.json'
}

check_no_stub_log_prefix() {
  # Exclude test files and assertion code that checks for absence of [STUB]
  local found
  found=$(grep -rl '"\[STUB\]"' "$SERVICES_DIR" --include='*.go' --include='*.rs' --include='*.py' --include='*.conf' | grep -v '_test\.' | grep -v 'test_' | grep -v 'stub_server.py' || true)
  [ -z "$found" ]
}

check_no_stub_flag_true() {
  ! grep -rq '"stub": *true' "$SERVICES_DIR" --include='*.go' --include='*.rs' --include='*.py' --include='*.json'
}

check_compose_image_tags_no_stub() {
  local compose="$ROOT_DIR/deploy/docker-compose.yml"
  ! grep -q '\-stub:' "$compose"
}

check "go Dockerfiles use golang->alpine" check_go_base_images
check "rust Dockerfiles use rust->debian" check_rust_base_images
check "python Dockerfiles use python slim multistage" check_python_base_images
check "ts Dockerfiles use node->nginx" check_typescript_base_images
check "language manifests exist" check_language_manifests
check "language entrypoints exist" check_entrypoints_exist
check "compose builds native service contexts" check_compose_builds_native_services
check "compose Python services have pyproject.toml + uv.lock" check_compose_python_services_have_manifests
check "ts nginx configs avoid invalid limit_except return" check_ts_nginx_method_guard_syntax
check "no -stub version in source files" check_no_stub_version_in_sources
check "no [STUB] log prefix in source files" check_no_stub_log_prefix
check "no stub:true flag in source files" check_no_stub_flag_true
check "compose image tags have no -stub suffix" check_compose_image_tags_no_stub

if [ "$failures" -gt 0 ]; then
  echo "FAILED: $failures checks failed"
  exit 1
fi

echo "ALL PASSED"
