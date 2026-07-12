#!/bin/bash
# Native service layout and production-readiness checks
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
SERVICES_DIR="$ROOT_DIR/src/services"

go_services=(api-gateway auth-service commenting-service ticket-api ticket-telemetry-listener ticket-notifier)
rust_services=(ontology-service versioning-service publisher-service public-browse-api)
python_services=(metrics-service ticket-classifier)
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
  for svc in "${go_services[@]}"; do
    grep -q '^FROM golang:1.22-alpine AS builder$' "$SERVICES_DIR/$svc/Dockerfile"
    grep -q '^FROM alpine:3.19$' "$SERVICES_DIR/$svc/Dockerfile"
  done
}

check_rust_base_images() {
  for svc in "${rust_services[@]}"; do
    grep -q '^FROM rust:1.77 AS builder$' "$SERVICES_DIR/$svc/Dockerfile"
    grep -q '^FROM debian:bookworm-slim$' "$SERVICES_DIR/$svc/Dockerfile"
  done
}

check_python_base_images() {
  for svc in "${python_services[@]}"; do
    grep -q '^FROM python:3.12-slim AS builder$' "$SERVICES_DIR/$svc/Dockerfile"
    grep -q '^FROM python:3.12-slim$' "$SERVICES_DIR/$svc/Dockerfile"
  done
}

check_typescript_base_images() {
  for svc in "${ts_services[@]}"; do
    grep -q '^FROM node:20 AS builder$' "$SERVICES_DIR/$svc/Dockerfile"
    grep -q '^FROM nginx:1.25-alpine$' "$SERVICES_DIR/$svc/Dockerfile"
  done
}

check_language_manifests() {
  for svc in "${go_services[@]}"; do
    test -f "$SERVICES_DIR/$svc/go.mod"
  done
  for svc in "${rust_services[@]}"; do
    test -f "$SERVICES_DIR/$svc/Cargo.toml"
  done
  for svc in "${python_services[@]}"; do
    test -f "$SERVICES_DIR/$svc/requirements.txt"
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
  for svc in "${built_services[@]}"; do
    grep -q "^  ${svc}:" "$compose"
    grep -q "context: ../src/services/${svc}" "$compose"
  done
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
