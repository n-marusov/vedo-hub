#!/usr/bin/env bash
# traceability-validator.sh — TTL ↔ Filesystem cross-validation
# Parses traceability.ttl, inventories filesystem artifacts, cross-validates,
# computes RCS (Requirements Coverage Score), and detects P0 orphans.
#
# Usage: ./traceability-validator.sh [--json] [--strict] [project-root]
#   project-root  Project root directory (default: auto-detect)
#   --json        Output JSON report
#   --strict      Treat WARN as BLOCK (exit 1 for all issues)

set -euo pipefail

# ── Auto-detect project root ────────────────────────────────────────────────
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Parse arguments
JSON_MODE=0
STRICT=0
CUSTOM_ROOT=""

for arg in "$@"; do
  case $arg in
    --json) JSON_MODE=1 ;;
    --strict) STRICT=1 ;;
    *)
      if [ -z "$CUSTOM_ROOT" ] && [ -d "$arg" ]; then
        CUSTOM_ROOT="$arg"
      fi
      ;;
  esac
done

if [ -n "$CUSTOM_ROOT" ]; then
  PROJECT_ROOT="$CUSTOM_ROOT"
fi

TTL_FILE="$PROJECT_ROOT/.ai-factory/traceability/traceability.ttl"
REQ_DIR="$PROJECT_ROOT/specs/requirements"
US_DIR="$PROJECT_ROOT/specs/user-stories"
UC_DIR="$PROJECT_ROOT/specs/use-cases"

EXIT_CODE=0
RCS_SCORE=""
RCS_GRADE=""
P0_ORPHAN_COUNT=0
STALE_COUNT=0
UNTRACKED_COUNT=0
VALIDATION_ISSUES=""

# Colors
if [ -t 1 ] && [ "$JSON_MODE" -eq 0 ]; then
  RED='\033[0;31m'; YELLOW='\033[0;33m'; GREEN='\033[0;32m'; CYAN='\033[0;36m'; NC='\033[0m'
else
  RED=''; YELLOW=''; GREEN=''; CYAN=''; NC=''
fi

block() { echo -e "${RED}⛔ BLOCK${NC} $1"; EXIT_CODE=1; }
warn() { echo -e "${YELLOW}⚠️  WARN${NC} $1"; [ "$STRICT" -eq 1 ] && EXIT_CODE=1; }
pass() { echo -e "${GREEN}✅ PASS${NC} $1"; }
info() { echo -e "   ${CYAN}ℹ️${NC} $1"; }

echo "═══════════════════════════════════════════════════"
echo " Traceability Validator — TTL ↔ Filesystem"
echo "═══════════════════════════════════════════════════"
echo ""

# ── Phase 1: Parse TTL ──────────────────────────────────────────────────────
echo "── Phase 1: Parse traceability.ttl ──"

if [ ! -f "$TTL_FILE" ]; then
  warn "traceability.ttl not found at: $TTL_FILE"
  echo ""
  echo "── RCS Score ──"
  echo -e "  ${YELLOW}RCS: N/A — no TTL file${NC}"
  echo ""
  echo "═══════════════════════════════════════════════════"
  [ "$EXIT_CODE" -eq 0 ] && echo -e "${GREEN}✅ PASS${NC} — No traceability issues (no TTL file)" || echo -e "${RED}⛔ BLOCK${NC} — Issues found"
  echo "═══════════════════════════════════════════════════"
  exit $EXIT_CODE
fi

info "TTL file found: $TTL_FILE"

# Extract entities
REQ_TTL=$(grep 'a vdo:Requirement' "$TTL_FILE" 2>/dev/null | sed 's/^[[:space:]]*//;s/[[:space:]]*a.*//' | sort -u || true)
US_TTL=$(grep 'a vdo:UserStory' "$TTL_FILE" 2>/dev/null | sed 's/^[[:space:]]*//;s/[[:space:]]*a.*//' | sort -u || true)
TEST_TTL=$(grep 'a vdo:TestSuite' "$TTL_FILE" 2>/dev/null | sed 's/^[[:space:]]*//;s/[[:space:]]*a.*//' | sort -u || true)
VALIDATES=$(grep 'vdo:validates' "$TTL_FILE" 2>/dev/null || true)
VERIFIED_BY=$(grep 'vdo:verifiedBy' "$TTL_FILE" 2>/dev/null || true)

REQ_TTL_COUNT=0
US_TTL_COUNT=0
TEST_TTL_COUNT=0
VALIDATES_COUNT=0

[ -n "$REQ_TTL" ] && REQ_TTL_COUNT=$(echo "$REQ_TTL" | grep -c . 2>/dev/null || true)
[ -n "$US_TTL" ] && US_TTL_COUNT=$(echo "$US_TTL" | grep -c . 2>/dev/null || true)
[ -n "$TEST_TTL" ] && TEST_TTL_COUNT=$(echo "$TEST_TTL" | grep -c . 2>/dev/null || true)
[ -n "$VALIDATES" ] && VALIDATES_COUNT=$(echo "$VALIDATES" | grep -c 'vdo:validates' 2>/dev/null || true)

info "$REQ_TTL_COUNT requirements in TTL"
info "$TEST_TTL_COUNT test suites in TTL"
info "$VALIDATES_COUNT validation links in TTL"

# ── Phase 2: Filesystem inventory ─────────────────────────────────────────────
echo ""
echo "── Phase 2: Filesystem Inventory ──"

REQ_FS_COUNT=$(find "$REQ_DIR" -name 'REQ-*.md' 2>/dev/null | wc -l) || true
US_FS_COUNT=$(find "$US_DIR" -name 'US-*.md' 2>/dev/null | wc -l) || true
UC_FS_COUNT=$(find "$UC_DIR" -name 'UC-*.md' 2>/dev/null | wc -l) || true
TEST_FS_COUNT=$(find "$PROJECT_ROOT/apps/services" \( -name '*_test.go' -o -name '*_test.rs' -o -name 'test_*.py' -o -name '*_test.py' -o -name '*.spec.ts' -o -name '*.test.ts' \) 2>/dev/null | wc -l) || true

info "$REQ_FS_COUNT REQ-*.md files on disk"
info "$US_FS_COUNT US-*.md files on disk"
info "$UC_FS_COUNT UC-*.md files on disk"
info "$TEST_FS_COUNT test files on disk"

# ── Phase 3: Cross-validation ────────────────────────────────────────────────
echo ""
echo "── Phase 3: Cross-validation ──"

# TTL → FS: check that TTL-referenced requirements exist on filesystem
if [ -n "$REQ_TTL" ]; then
  while IFS= read -r req; do
    req_id=$(echo "$req" | sed 's/^vdo://')
    req_file=$(find "$REQ_DIR" -name "${req_id}.md" 2>/dev/null | head -1) || true
    if [ -z "$req_file" ]; then
      warn "TTL→FS STALE: $req referenced in TTL but no ${req_id}.md on disk"
      STALE_COUNT=$((STALE_COUNT + 1))
    fi
  done <<< "$REQ_TTL"
fi

# FS → TTL: count untracked REQ files (exist on disk but not in TTL)
if [ "$REQ_FS_COUNT" -gt 0 ] && [ "$REQ_TTL_COUNT" -gt 0 ]; then
  untracked_reqs=$((REQ_FS_COUNT - REQ_TTL_COUNT))
  if [ "$untracked_reqs" -gt 0 ]; then
    warn "$untracked_reqs REQ-*.md files not tracked in traceability.ttl"
    UNTRACKED_COUNT=$((UNTRACKED_COUNT + untracked_reqs))
  fi
fi

# Check test files exist for TTL test suite entries
if [ -n "$TEST_TTL" ]; then
  while IFS= read -r ts; do
    ts_id=$(echo "$ts" | sed 's/^vdo://')
    # Try to find corresponding test file
    ts_found=$(find "$PROJECT_ROOT/apps/services" -path "*${ts_id}*" -name '*_test*' 2>/dev/null | head -1) || true
    if [ -z "$ts_found" ]; then
      # Check if the test suite references a source file
      source_file=$(grep -A2 "^${ts}[[:space:]]" "$TTL_FILE" 2>/dev/null | grep 'vdo:sourceFile' | sed 's/.*"\(.*\)".*/\1/' || true)
      if [ -n "$source_file" ] && [ ! -f "$PROJECT_ROOT/$source_file" ]; then
        warn "TTL→FS STALE: $ts references source file not found: $source_file"
        STALE_COUNT=$((STALE_COUNT + 1))
      fi
    fi
  done <<< "$TEST_TTL"
fi

# Count P0 requirements without test coverage
echo ""
echo "── Phase 4: P0 Orphan Detection & RCS ──"

P0_WEIGHT=3
P1_WEIGHT=2
P2_WEIGHT=1
P3_WEIGHT=0

total_weight=0
covered_weight=0
P0_ORPHAN_COUNT=0

if [ -n "$REQ_TTL" ]; then
  while IFS= read -r req; do
    req_id=$(echo "$req" | sed 's/^vdo://')
    # Get priority from TTL
    priority=$(grep -A5 "^${req}[[:space:]]" "$TTL_FILE" 2>/dev/null | grep 'vdo:priority' | sed 's/.*"\(P[0-9]\)".*/\1/' || echo "P3")

    # Determine weight
    weight=$P3_WEIGHT
    [ "$priority" = "P0" ] && weight=$P0_WEIGHT
    [ "$priority" = "P1" ] && weight=$P1_WEIGHT
    [ "$priority" = "P2" ] && weight=$P2_WEIGHT

    # Check if this requirement has any vdo:validates links from tests
    is_covered=$(echo "$VALIDATES" | grep -c "vdo:validates.*${req}\b" 2>/dev/null || echo 0)

    total_weight=$((total_weight + weight))
    if [ "$is_covered" -gt 0 ]; then
      covered_weight=$((covered_weight + weight))
    elif [ "$priority" = "P0" ]; then
      warn "P0 ORPHAN: $req_id has zero test coverage"
      P0_ORPHAN_COUNT=$((P0_ORPHAN_COUNT + 1))
    fi
  done <<< "$REQ_TTL"
fi

# Compute RCS
if [ "$total_weight" -gt 0 ]; then
  RCS_INT=$((covered_weight * 1000 / total_weight))
  RCS_DISPLAY=$(awk "BEGIN { printf \"%.1f\", $RCS_INT / 100 }" 2>/dev/null || echo "$RCS_INT")
  if [ "$RCS_INT" -ge 900 ]; then RCS_GRADE="excellent"
  elif [ "$RCS_INT" -ge 700 ]; then RCS_GRADE="good"
  elif [ "$RCS_INT" -ge 500 ]; then RCS_GRADE="partial"
  else RCS_GRADE="poor"
  fi
else
  RCS_INT=0
  RCS_DISPLAY="0.0"
  RCS_GRADE="na"
fi

# ── Output ───────────────────────────────────────────────────────────────────
if [ "$JSON_MODE" -eq 1 ]; then
  cat <<EOJSON
{
  "traceability": "$([ "$EXIT_CODE" -eq 0 ] && echo "VALID" || echo "ISSUES")",
  "rcs": $RCS_DISPLAY,
  "grade": "$RCS_GRADE",
  "orphan_p0": $P0_ORPHAN_COUNT,
  "stale_refs": $STALE_COUNT,
  "untracked": $UNTRACKED_COUNT,
  "ttl_requirements": $REQ_TTL_COUNT,
  "fs_requirements": $REQ_FS_COUNT,
  "ttl_test_suites": $TEST_TTL_COUNT,
  "fs_test_files": $TEST_FS_COUNT
}
EOJSON
else
  echo ""
  rcs_color="$GREEN"
  [ "$RCS_GRADE" = "partial" ] && rcs_color="$YELLOW"
  [ "$RCS_GRADE" = "poor" ] && rcs_color="$RED"
  echo -e "  ${rcs_color}RCS: ${RCS_DISPLAY} (${RCS_GRADE})${NC}"
  if [ "$P0_ORPHAN_COUNT" -gt 0 ]; then
    echo -e "  ${RED}P0 orphans: ${P0_ORPHAN_COUNT}${NC}"
  fi
  echo -e "  Stale refs: ${STALE_COUNT} | Untracked: ${UNTRACKED_COUNT}"
  echo ""
  echo "═══════════════════════════════════════════════════"
  if [ "$EXIT_CODE" -eq 0 ]; then
    echo -e "${GREEN}✅ PASS${NC} — Traceability OK"
  else
    echo -e "${RED}⛔ BLOCK${NC} — Traceability issues found"
  fi
  echo "═══════════════════════════════════════════════════"

  # Machine-readable output
  echo ""
  echo "<!-- traceability: $([ "$EXIT_CODE" -eq 0 ] && echo "VALID" || echo "ISSUES") | rcs: ${RCS_DISPLAY} | grade: ${RCS_GRADE} | orphan-p0: ${P0_ORPHAN_COUNT} -->"
fi

exit $EXIT_CODE
