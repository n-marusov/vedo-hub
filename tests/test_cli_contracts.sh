#!/usr/bin/env bash
# CLI command framework contract tests — Go unit tests for vedo-cli
set -euo pipefail

cd "$(dirname "$0")/.."
echo "=== CLI Contract Tests ==="
go -C src/cli test ./... -v
echo "=== CLI Contract Tests: ALL PASSED ==="
