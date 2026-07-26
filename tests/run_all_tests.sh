#!/bin/bash
# Combined test runner
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"

echo "=== Running build command tests ==="
bash "$ROOT/tests/gates/test_build_commands.sh"

echo ""
echo "=== Running Docker build tests ==="
bash "$ROOT/tests/gates/test_docker_build.sh"

echo ""
echo "=== Running CI + Compose tests ==="
bash "$ROOT/tests/gates/test_ci_and_compose.sh"

echo ""
echo "=== Running health + metadata tests ==="
bash "$ROOT/tests/gates/test_health_metadata.sh"

echo ""
echo "=== Running observability stack tests ==="
bash "$ROOT/tests/gates/test_observability_stack.sh"

echo ""
echo "=== Running Antora setup tests ==="
bash "$ROOT/tests/gates/test_antora_setup.sh"

echo ""
echo "=== Running Playwright setup tests ==="
bash "$ROOT/tests/gates/test_playwright_setup.sh"

echo ""
echo "=== Running service template observability tests ==="
bash "$ROOT/tests/gates/test_service_observability_templates.sh"

echo ""
echo "=== Running stub services and layout tests ==="
bash "$ROOT/tests/gates/test_stub_services_and_layout.sh"

echo ""
echo "=== Running native stub tests ==="
bash "$ROOT/tests/gates/test_native_stubs.sh"

echo ""
echo "=== Running milestone Docker gate (health-only) ==="
bash "$ROOT/tests/gates/test_milestone_docker_gate.sh" --health-only

echo ""
echo "=== All tests passed ==="
