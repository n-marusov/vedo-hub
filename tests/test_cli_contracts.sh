#!/usr/bin/env bash
# @ctx: CLI command framework contract tests — Go unit tests for vedo-cli
set -euo pipefail

cd "$(dirname "$0")/.."
echo "=== CLI Contract Tests ==="
go test ./src/cli/... -v
echo "=== CLI Contract Tests: ALL PASSED ==="
