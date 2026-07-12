#!/usr/bin/env bash
# @ctx: SCA gate contract tests — Go unit tests for security supply-chain gate
set -euo pipefail

cd "$(dirname "$0")/.."
echo "=== SCA Contract Tests ==="
go test ./src/security/sca/... -v
echo "=== SCA Contract Tests: ALL PASSED ==="
