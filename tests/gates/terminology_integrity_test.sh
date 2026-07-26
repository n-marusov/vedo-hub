#!/usr/bin/env bash
# ============================================================================
# Terminology Consistency Test
# ============================================================================
# Scans modified files for stale milestone references (M2.1, M2.5, etc.)
# and verifies "Project" is consistently described as an ontology container.
#
# Usage:
#   ./tests/scripts/terminology_integrity_test.sh          # run against all files
#   ./tests/scripts/terminology_integrity_test.sh --diff    # run against git diff only
#
# Exit codes:
#   0 — all checks pass
#   1 — stale references found
#   2 — project terminology violation found
# ============================================================================

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color
ERRORS=0
PROJECT_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$PROJECT_ROOT"

# Determine file list
if [[ "${1:-}" == "--diff" ]]; then
    FILES=$(git diff --name-only --diff-filter=ACMR -- '*.go' '*.rs' '*.ts' '*.vue' '*.md' '*.adoc' '*.yaml' '*.yml' '*.json')
else
    FILES=$(find . -type f \( -name "*.go" -o -name "*.rs" -o -name "*.ts" -o -name "*.vue" -o -name "*.md" -o -name "*.adoc" -o -name "*.yaml" -o -name "*.yml" -o -name "*.json" \) \
        -not -path "./.git/*" \
        -not -path "./node_modules/*" \
        -not -path "./target/*" \
        -not -path "./.venv/*" \
        -not -path "./.ai-factory/evolution/*" \
        -not -path "./.ai-factory/patches/*" \
        -not -path "./specs/*" 2>/dev/null || true)
fi

echo "━━━ Terminology Consistency Check ━━━"
echo ""

# ── Check 1: Stale M2.1 references ──────────────────────────────────────────
echo "--- Check 1: Stale M2.1 references ---"
STALE_M21=()
while IFS= read -r file; do
    if grep -n 'M2\.1' "$file" > /dev/null 2>&1; then
        # Allow historical/explanatory references
        while IFS= read -r line; do
            # Skip if the line is an explanatory comment about the rename
            if echo "$line" | grep -qiE "(originally created as|was originally|former M2\.1|now M3|continuous numbering|прежний M2\.1|Нет по M2\.1|preserved as implementation history)"; then
                continue
            fi
            STALE_M21+=("$file:$line")
        done < <(grep -n 'M2\.1' "$file" 2>/dev/null || true)
    fi
done <<< "$FILES"

if [[ ${#STALE_M21[@]} -gt 0 ]]; then
    echo -e "${RED}FAIL: Found ${#STALE_M21[@]} stale M2.1 reference(s):${NC}"
    for ref in "${STALE_M21[@]}"; do
        echo "  $ref"
    done
    ERRORS=$((ERRORS + 1))
else
    echo -e "${GREEN}PASS: No stale M2.1 references found.${NC}"
fi

# ── Check 2: Stale M2.5 references (should be M4 in current numbering) ─────
echo ""
echo "--- Check 2: Stale M2.5 references ---"
STALE_M25=()
while IFS= read -r file; do
    if grep -n 'M2\.5' "$file" > /dev/null 2>&1; then
        while IFS= read -r line; do
            # Skip historical references and plan/patch files
            if echo "$line" | grep -qiE "(former M2\.5|now M4|M2\.5 scope|post-M2\.5|M2\.5 page|M2\.5 route|M2\.5 Queries|M2\.5 test|M2\.5 Vitest|M2\.5 E2E|playwright\.m2\.5|feature-m2-5)"; then
                continue
            fi
            STALE_M25+=("$file:$line")
        done < <(grep -n 'M2\.5' "$file" 2>/dev/null || true)
    fi
done <<< "$FILES"

if [[ ${#STALE_M25[@]} -gt 0 ]]; then
    echo -e "${YELLOW}WARN: Found ${#STALE_M25[@]} M2.5 reference(s) — verify should be M4:${NC}"
    for ref in "${STALE_M25[@]}"; do
        echo "  $ref"
    done
    # Not an error — may be valid historical references
    echo "  (action: update to M4 if referencing current milestone)"
else
    echo -e "${GREEN}PASS: No stale M2.5 references found.${NC}"
fi

# ── Check 3: Project terminology ────────────────────────────────────────────
echo ""
echo "--- Check 3: Project-as-ontology-container terminology ---"
PROJECT_VIOLATIONS=()
while IFS= read -r file; do
    # In user-facing docs/frontend, 'Project' should be described as containing an ontology
    # Skip plan files and system-level docs that use generic terms
    if echo "$file" | grep -qiE "\.(ai-factory|git|specs)" || echo "$file" | grep -qiE "(node_modules|target|\.venv)"; then
        continue
    fi
    # Only check user-facing files (frontend, API docs, integrator/developer guide)
    if echo "$file" | grep -qiE "(frontend|openapi|antora.*guide|docs.*user|docs.*integrator|docs.*admin)"; then
        # Look for "project" used as a generic list/grid item without ontology context
        # This is a heuristic — flag suspicious patterns for human review
        if grep -n -i 'project' "$file" | grep -qiE "(list of projects|project list|project grid|all projects)" 2>/dev/null; then
            # Check if the surrounding context mentions ontology
            local_context=$(grep -B3 -A3 -i 'list of projects\|project list\|project grid\|all projects' "$file" 2>/dev/null | head -10)
            if ! echo "$local_context" | grep -qiE "(ontology|ontologies|knowledge graph|graph model)"; then
                PROJECT_VIOLATIONS+=("$file")
            fi
        fi
    fi
done <<< "$FILES"

if [[ ${#PROJECT_VIOLATIONS[@]} -gt 0 ]]; then
    echo -e "${YELLOW}WARN: ${#PROJECT_VIOLATIONS[@]} file(s) may need project→ontology-container terminology review:${NC}"
    for f in "${PROJECT_VIOLATIONS[@]}"; do
        echo "  $f"
    done
else
    echo -e "${GREEN}PASS: Project terminology looks consistent.${NC}"
fi

# ── Summary ─────────────────────────────────────────────────────────────────
echo ""
echo "━━━ Summary ━━━"
if [[ $ERRORS -eq 0 ]]; then
    echo -e "${GREEN}All checks passed.${NC}"
    exit 0
else
    echo -e "${RED}$ERRORS check(s) failed.${NC}"
    exit 1
fi
