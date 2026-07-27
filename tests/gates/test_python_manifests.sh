#!/bin/bash
# CT-PYTHON-001: All Python services defined in docker-compose must have uv manifests
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/../.." && pwd)"
COMPOSE="$ROOT_DIR/deploy/docker-compose.yml"

if [ ! -f "$COMPOSE" ]; then
  echo "FAIL: compose file not found"
  exit 1
fi

errors=0

# Extract service name + SERVICE_DIR pairs from docker-compose for all Dockerfile.python services
while IFS='|' read -r svc_name svc_dir; do
  abs_dir="$ROOT_DIR/$svc_dir"
  if [ ! -f "$abs_dir/pyproject.toml" ]; then
    echo "FAIL: $svc_name missing pyproject.toml (expected at $svc_dir/pyproject.toml)"
    errors=$((errors + 1))
  fi
  if [ ! -f "$abs_dir/uv.lock" ]; then
    echo "FAIL: $svc_name missing uv.lock (expected at $svc_dir/uv.lock)"
    errors=$((errors + 1))
  fi
done < <(
  grep -B5 -A2 'dockerfile: tools/dockerfiles/Dockerfile.python' "$COMPOSE" | \
    awk -F': ' '
      /^  [a-z]/   { svc=$1; gsub(/[ :]/, "", svc) }
      /SERVICE_DIR/ { dir=$2; print svc "|" dir }
    '
)

if [ "$errors" -gt 0 ]; then
  echo "FAILED: $errors Python service(s) missing uv manifests"
  exit 1
fi

echo "PASS: All Python services have pyproject.toml and uv.lock"
