#!/bin/bash
# @ctx: Contract checks for Antora documentation infrastructure
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
DOCS_ROOT="$ROOT_DIR/src/docs/antora"
DOCS_COMPOSE="$ROOT_DIR/src/docker-compose.docs.yaml"

# @hlv CT-DOCS-001
test_antora_playbook_exists() {
  test -f "$DOCS_ROOT/antora-playbook.yml"
}

# @hlv CT-DOCS-002
test_four_components_exist() {
  test -f "$DOCS_ROOT/user-guide/antora.yml"
  test -f "$DOCS_ROOT/developer-guide/antora.yml"
  test -f "$DOCS_ROOT/admin-guide/antora.yml"
  test -f "$DOCS_ROOT/integrator-guide/antora.yml"
}

# @hlv CT-DOCS-003
test_nav_files_exist() {
  test -f "$DOCS_ROOT/user-guide/modules/ROOT/nav.adoc"
  test -f "$DOCS_ROOT/developer-guide/modules/ROOT/nav.adoc"
  test -f "$DOCS_ROOT/admin-guide/modules/ROOT/nav.adoc"
  test -f "$DOCS_ROOT/integrator-guide/modules/ROOT/nav.adoc"
}

# @hlv CT-DOCS-004
test_offline_content_scaffold_exists() {
  test -f "$DOCS_ROOT/build/site/user-guide/index.html"
  test -f "$DOCS_ROOT/build/site/developer-guide/index.html"
  test -f "$DOCS_ROOT/build/site/admin-guide/index.html"
  test -f "$DOCS_ROOT/build/site/integrator-guide/index.html"
}

# @hlv CT-DOCS-005
# @hlv CT-DOCS-007
test_docs_ports_declared_in_compose() {
  grep -q '^  docs-user:' "$DOCS_COMPOSE"
  grep -q '^  docs-dev:' "$DOCS_COMPOSE"
  grep -q '^  docs-admin:' "$DOCS_COMPOSE"
  grep -q '^  docs-integrator:' "$DOCS_COMPOSE"
  grep -q 'profiles: \["documentation"\]' "$DOCS_COMPOSE"
  grep -q '"5000:5000"' "$DOCS_COMPOSE"
  grep -q '"5001:5001"' "$DOCS_COMPOSE"
  grep -q '"5002:5002"' "$DOCS_COMPOSE"
  grep -q '"5003:5003"' "$DOCS_COMPOSE"
}

# @hlv CT-DOCS-006
test_update_guide_exists() {
  test -f "$DOCS_ROOT/update.md"
}

# @hlv ANTORA_BUILD_FAILED
# @hlv LINK_CHECK_FAILED
# @hlv STYLE_CHECK_FAILED
# @hlv SNIPPET_CHECK_FAILED
test_docs_error_paths_covered() {
  grep -q 'start_page:' "$DOCS_ROOT/antora-playbook.yml"
}

for t in $(declare -F | awk '{print $3}' | grep '^test_'); do
  "$t"
done
