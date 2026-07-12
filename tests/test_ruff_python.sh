#!/usr/bin/env bash
# @ctx: Python Ruff lint + format gate — runs ruff check --fix and ruff format
# on all Python services (directories with pyproject.toml).
set -euo pipefail

ROOT="$(realpath "$(dirname "$0")/..")"

# Find Python project directories (same logic as python.mk)
PYTHON_DIRS="$(find "$ROOT/src" -maxdepth 4 -name pyproject.toml -not -path "*/node_modules/*" -not -path "*\.venv/*" -not -path "*/__pycache__/*" -exec dirname {} \; 2>/dev/null | sort -u)"

if [ -z "$PYTHON_DIRS" ]; then
  echo "[ruff] No Python services found — skipping"
  exit 0
fi

FAILED=0
for dir in $PYTHON_DIRS; do
  svc="$(basename "$dir")"
  echo "[ruff] $svc — running ruff check --fix"
  (cd "$dir" && uvx ruff check --fix . 2>&1) || { echo "RUFF_CHECK_FAILED: $svc"; FAILED=1; }

  echo "[ruff] $svc — running ruff format"
  (cd "$dir" && uvx ruff format . 2>&1) || { echo "RUFF_FORMAT_FAILED: $svc"; FAILED=1; }
done

if [ "$FAILED" -eq 1 ]; then
  echo "[ruff] One or more checks failed"
  exit 1
fi

echo "[ruff] All Python services passed"
