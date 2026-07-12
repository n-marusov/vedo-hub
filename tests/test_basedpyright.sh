#!/usr/bin/env bash
# Python typecheck gate — runs basedpyright on all Python services.
set -euo pipefail

ROOT="$(realpath "$(dirname "$0")/..")"

# Find Python project directories
PYTHON_DIRS="$(find "$ROOT/src" -maxdepth 4 -name pyproject.toml -not -path "*/node_modules/*" -not -path "*\.venv/*" -not -path "*/__pycache__/*" -exec dirname {} \; 2>/dev/null | sort -u)"

if [ -z "$PYTHON_DIRS" ]; then
  echo "[basedpyright] No Python services found — skipping"
  exit 0
fi

FAILED=0
for dir in $PYTHON_DIRS; do
  svc="$(basename "$dir")"
  echo "[basedpyright] $svc — running basedpyright"
  (cd "$dir" && uvx basedpyright . 2>&1) || { echo "BASEDPYRIGHT_FAILED: $svc"; FAILED=1; }
done

if [ "$FAILED" -eq 1 ]; then
  echo "[basedpyright] One or more typecheck failures"
  exit 1
fi

echo "[basedpyright] All Python services passed"
