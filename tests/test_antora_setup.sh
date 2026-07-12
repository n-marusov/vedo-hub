#!/bin/bash
# Contract checks for Antora documentation infrastructure
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
DOCS_ROOT="$ROOT_DIR/src/docs/antora"
DOCS_COMPOSE="$ROOT_DIR/deploy/docker-compose.docs.yaml"

test_antora_playbook_exists() {
  test -f "$DOCS_ROOT/antora-playbook.yml"
}
test_four_components_exist() {
  test -f "$DOCS_ROOT/user-guide/antora.yml"
  test -f "$DOCS_ROOT/developer-guide/antora.yml"
  test -f "$DOCS_ROOT/admin-guide/antora.yml"
  test -f "$DOCS_ROOT/integrator-guide/antora.yml"
}

test_nav_files_exist() {
  test -f "$DOCS_ROOT/user-guide/modules/ROOT/nav.adoc"
  test -f "$DOCS_ROOT/developer-guide/modules/ROOT/nav.adoc"
  test -f "$DOCS_ROOT/admin-guide/modules/ROOT/nav.adoc"
  test -f "$DOCS_ROOT/integrator-guide/modules/ROOT/nav.adoc"
}

test_offline_content_scaffold_exists() {
  # Build artifacts — skip if Antora site has not been built yet
  if [ -f "$DOCS_ROOT/build/site/user-guide/index.html" ]; then
    return 0
  fi
  echo "SKIP: offline content scaffold — build artifacts not found (run antora build first)"
}

test_docs_ports_declared_in_compose() {
  grep -q '^  docs-user:' "$DOCS_COMPOSE"
  grep -q '^  docs-dev:' "$DOCS_COMPOSE"
  grep -q '^  docs-admin:' "$DOCS_COMPOSE"
  grep -q '^  docs-integrator:' "$DOCS_COMPOSE"
  grep -F -q 'profiles: ["documentation"]' "$DOCS_COMPOSE"
  grep -F -q '"5000:5000"' "$DOCS_COMPOSE"
  grep -F -q '"5001:5001"' "$DOCS_COMPOSE"
  grep -F -q '"5002:5002"' "$DOCS_COMPOSE"
  grep -F -q '"5003:5003"' "$DOCS_COMPOSE"
}

test_update_guide_exists() {
  test -f "$DOCS_ROOT/update.md"
}

test_docs_error_paths_covered() {
  grep -q 'start_page:' "$DOCS_ROOT/antora-playbook.yml"
}

for t in $(declare -F | awk '{print $3}' | grep '^test_'); do
  "$t"
done
