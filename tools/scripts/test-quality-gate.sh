#!/usr/bin/env bash
# test-quality-gate.sh — Static test quality gate for CI
# Scans test files for blocking anti-patterns (B1-B7), inline tests, and TQS scoring.
# Exit 0 = PASS, Exit 1 = BLOCK
#
# Usage:
#   ./test-quality-gate.sh [--score] [--json] [--strict] [target-dir]
#   target-dir   Directory to scan (default: apps/services)
#   --score      Enable TQS scoring mode (outputs per-service quality score)
#   --json       Output machine-readable JSON instead of human-readable text
#   --strict     Treat warnings as failures

set -euo pipefail

# Parse arguments
TARGET_DIR=
SCORE_MODE=0
JSON_MODE=0
STRICT=

for arg in "$@"; do
  case $arg in
    --score) SCORE_MODE=1 ;;
    --json) JSON_MODE=1 ;;
    --strict) STRICT="--strict" ;;
    *)
      if [ -z "$TARGET_DIR" ] && [ -d "$arg" ]; then
        TARGET_DIR="$arg"
      elif [ -z "$TARGET_DIR" ]; then
        # Assume it's a directory that doesn't exist yet (will be checked later)
        TARGET_DIR="$arg"
      fi
      ;;
  esac
done

TARGET_DIR="${TARGET_DIR:-apps/services}"

if [ ! -d "$TARGET_DIR" ]; then
  echo "ERROR: target directory not found: $TARGET_DIR" >&2
  exit 1
fi

EXIT_CODE=0
BLOCK_COUNT=0
WARN_COUNT=0
FILES_SCANNED=0
B5_COUNT=0
TQS_SCORE=
TQS_GRADE=

# Colors (disabled in CI / JSON mode)
if [ -t 1 ] && [ "$JSON_MODE" -eq 0 ]; then
  RED='\033[0;31m'; YELLOW='\033[0;33m'; GREEN='\033[0;32m'; CYAN='\033[0;36m'; NC='\033[0m'
else
  RED=''; YELLOW=''; GREEN=''; CYAN=''; NC=''
fi

# Colors (disabled in CI)
if [ -t 1 ]; then
  RED='\033[0;31m'; YELLOW='\033[0;33m'; GREEN='\033[0;32m'; NC='\033[0m'
else
  RED=''; YELLOW=''; GREEN=''; NC=''
fi

block() {
  echo -e "${RED}⛔ BLOCK${NC} $1"
  BLOCK_COUNT=$((BLOCK_COUNT + 1))
  EXIT_CODE=1
}

warn() {
  echo -e "${YELLOW}⚠️  WARN${NC} $1"
  WARN_COUNT=$((WARN_COUNT + 1))
  if [ "$STRICT" = "--strict" ]; then
    EXIT_CODE=1
  fi
}

pass() {
  echo -e "${GREEN}✅ PASS${NC} $1"
}

info() {
  echo "   ℹ️  $1"
}

tqs_grade() {
  # TQS as integer 0-1000 (610 = 61.0 = bronze)
  local tqs_int=$1
  if [ -z "$tqs_int" ] || [ "$tqs_int" -eq 0 ]; then
    echo "na"
  elif [ "$tqs_int" -ge 950 ]; then
    echo "gold"
  elif [ "$tqs_int" -ge 800 ]; then
    echo "silver"
  elif [ "$tqs_int" -ge 600 ]; then
    echo "bronze"
  else
    echo "fail"
  fi
}

# Quick heuristic: count BDD-named test functions
# Looks for patterns like TestX_Valid_Y_Expected, test_x_valid_y, should_*, 'should X when Y'
count_bdd_tests() {
  local dir=$1; local count=0
  # Go: TestX_Y_Z pattern
  count=$((count + $(grep -rns 'func Test.*_.*_.*(' "$dir" --include='*_test.go' 2>/dev/null | wc -l || true) ))
  # Rust: should_* or *_when_* pattern
  count=$((count + $(grep -rns 'fn should_' "$dir" --include='*_test.rs' 2>/dev/null | wc -l || true) ))
  count=$((count + $(grep -rs 'fn .*_when_.*' "$dir" --include='*_test.rs' 2>/dev/null | wc -l || true) ))
  # Python: test_x_y_z pattern
  count=$((count + $(grep -rs 'def test.*_.*_.*(' "$dir" --include='test_*.py' --include='*_test.py' 2>/dev/null | wc -l || true) ))
  # TS: 'should X when Y' pattern
  count=$((count + $(grep -rs "it('\\(should'\|" "$dir" --include='*.spec.ts' --include='*.test.ts' 2>/dev/null | wc -l || true) ))
  echo $count
}

# Quick heuristic: count mock/assert patterns
count_pattern() {
  local dir=$1; local pattern=$2; shift 2
  grep -rs "$pattern" "$dir" "$@" 2>/dev/null | wc -l || true
}

compute_tqs() {
  local dir=$1
  local total_tests bdd_tests mock_count fixture_count assert_count
  local give_count sleep_count skip_count test_count file_count

  total_tests=$(count_bdd_tests "$dir")
  bdd_tests=$total_tests  # reuse
  mock_count=$(count_pattern "$dir" 'mock\|Mock\|@patch\|vi\.mock' --include='*_test.go' --include='*_test.rs' --include='test_*.py' --include='*_test.py' --include='*.spec.ts' --include='*.test.ts' --exclude-dir='.venv' 2>/dev/null)
  fixture_count=$(count_pattern "$dir" 'fixture\|Fixture\|TestFixture\|setUp\|beforeEach\|before_all' --include='*_test.go' --include='*_test.rs' --include='test_*.py' --include='*_test.py' --include='*.spec.ts' --include='*.test.ts' --exclude-dir='.venv' 2>/dev/null)
  assert_count=$(count_pattern "$dir" 'assert\|require\.\|expect(' --include='*_test.go' --include='*_test.rs' --include='test_*.py' --include='*_test.py' --include='*.spec.ts' --include='*.test.ts' --exclude-dir='.venv' 2>/dev/null)
  give_count=$(count_pattern "$dir" 'given\|// given\|# given\|// arrange' --include='*_test.go' --include='*_test.rs' --include='test_*.py' --include='*_test.py' --include='*.spec.ts' --include='*.test.ts' --exclude-dir='.venv' 2>/dev/null)
  sleep_count=$(count_pattern "$dir" 'Sleep\|sleep(' --include='*_test.go' --include='*_test.rs' --include='test_*.py' --include='*_test.py' --include='*.spec.ts' --include='*.test.ts' --exclude-dir='.venv' 2>/dev/null)
  skip_count=$(count_pattern "$dir" 'Skip\|#[ignore]\|@skip\|test\.skip' --include='*_test.go' --include='*_test.rs' --include='test_*.py' --include='*_test.py' --include='*.spec.ts' --include='*.test.ts' --exclude-dir='.venv' 2>/dev/null)

  test_count=$(find "$dir" -name '*_test.go' -o -name '*_test.rs' -o -name 'test_*.py' -o -name '*_test.py' -o -name '*.spec.ts' -o -name '*.test.ts' 2>/dev/null | wc -l)
  test_count=$((test_count + $(find "$dir" -path '*/tests/*.rs' 2>/dev/null | wc -l || echo 0) ))

  if [ "$test_count" -eq 0 ]; then
    echo "0|0|0|0|0|0|0|na"
    return
  fi

  # Structural (0.30): BDD naming, file count
  local structural
  if [ "$total_tests" -gt 0 ]; then
    local bdd_ratio=$((bdd_tests * 10 / total_tests))
    [ "$bdd_ratio" -gt 10 ] && bdd_ratio=10
    structural=$bdd_ratio
  else
    structural=5  # no tests found, neutral
  fi

  # Dependencies (0.25): mock quality, fixtures
  local deps=5
  if [ "$test_count" -gt 0 ]; then
    local avg_mocks=$((mock_count / test_count))
    if [ "$avg_mocks" -le 2 ]; then deps=8; fi
    if [ "$avg_mocks" -le 5 ]; then deps=6; else deps=4; fi
    if [ "$fixture_count" -gt 0 ]; then deps=$((deps + 1)); fi
  fi
  [ "$deps" -gt 10 ] && deps=10

  # Readability (0.20): given/when/then, test length
  local readability=5
  if [ "$test_count" -gt 0 ] && [ "$give_count" -gt 0 ]; then
    readability=7
  fi
  if [ "$give_count" -gt "$((test_count / 2))" ]; then
    readability=9
  fi

  # Safety (0.15): anti-patterns, sleep, skips
  local safety=10
  if [ "$BLOCK_COUNT" -gt 0 ]; then safety=0; fi
  if [ "$sleep_count" -gt 0 ]; then safety=$((safety - 3)); fi
  if [ "$skip_count" -gt 0 ]; then safety=$((safety - 2)); fi
  [ "$safety" -lt 0 ] && safety=0

  # Coverage (0.10): assertions per function, scenarios
  local coverage=5
  if [ "$total_tests" -gt 0 ] && [ "$assert_count" -gt 0 ]; then
    local avg_asserts=$((assert_count / total_tests))
    if [ "$avg_asserts" -ge 3 ]; then coverage=8; fi
    if [ "$avg_asserts" -ge 5 ]; then coverage=10; fi
    if [ "$avg_asserts" -ge 1 ] && [ "$avg_asserts" -lt 3 ]; then coverage=6; fi
  fi

  # Compute final TQS as integer 0-100 (75 = 7.5)
  # Formula: structural*30 + deps*25 + readability*20 + safety*15 + coverage*10 = score 0-1000, /10 = 0-100
  local tqs_int=$(( (structural*3 + deps*25/10 + readability*2 + safety*15/10 + coverage*1) * 10 ))

  local grade=$(tqs_grade "$tqs_int")
  echo "${tqs_int}|${grade}|${structural}|${deps}|${readability}|${safety}|${coverage}|${test_count}"
}

echo "═══════════════════════════════════════════════════"
echo " Test Quality Gate — Static Analysis"
echo " Target: ${TARGET_DIR}"
echo "═══════════════════════════════════════════════════"
echo ""

# ── B1: Tautology Assertions ───────────────────────────────────────────────
echo "── B1: Tautology Assertions ──"
B1_HITS=$(grep -rn \
  -e 'assert\.True(t\\?,\\s*true)' \
  -e 'require\.True(t\\?,\\s*true)' \
  -e 'assert\.False(t\\?,\\s*false)' \
  -e 'assert!(true)' \
  -e 'assert_eq!(\\(\\w+\\),\\s*\\1)' \
  -e '^\s*assert\\s\+True\s*$' \
  -e 'expect(true)\.toBe(true)' \
  "$TARGET_DIR" \
  --include='*_test.go' --include='*_test.rs' --include='test_*.py' --include='*_test.py' \
  --include='*.spec.ts' --include='*.test.ts' \
  --exclude-dir='.venv' \
  2>/dev/null || true)

if [ -n "$B1_HITS" ]; then
  while IFS= read -r line; do
    block "B1: $line"
  done <<< "$B1_HITS"
else
  pass "B1: No tautology assertions"
fi

# ── B2: Sleep in Test ──────────────────────────────────────────────────────
echo "── B2: Sleep in Test ──"
B2_HITS=$(grep -rn \
  -e 'time\.Sleep(' \
  -e 'std::thread::sleep(' \
  -e 'time\.sleep(' \
  "$TARGET_DIR" \
  --include='*_test.go' --include='*_test.rs' --include='test_*.py' --include='*_test.py' \
  --include='*.spec.ts' --include='*.test.ts' \
  --exclude-dir='.venv' \
  2>/dev/null || true)
# TypeScript: only bare setTimeout (not inside Promise)
B2_TS=$(grep -rn 'setTimeout(' "$TARGET_DIR" \
  --include='*.spec.ts' --include='*.test.ts' \
  --exclude-dir='.venv' \
  2>/dev/null | grep -v 'new Promise' || true)
if [ -n "$B2_TS" ]; then
  B2_HITS=$(printf '%s\n' "$B2_HITS" "$B2_TS" | grep -v '^$' || true)
fi

if [ -n "$B2_HITS" ]; then
  while IFS= read -r line; do
    block "B2: $line"
  done <<< "$B2_HITS"
else
  pass "B2: No sleep in tests"
fi

# ── B3: Empty Catch / Swallowed Error ──────────────────────────────────────
echo "── B3: Empty Catch / Swallowed Error ──"
B3_HITS=$(grep -rn \
  -e 'if\s*err\s*!=\s*nil\s*{\s*}' \
  -e 'except.*:\s*pass' \
  -e 'catch\s*(\w*)\s*{\s*}' \
  "$TARGET_DIR" \
  --include='*_test.go' --include='*_test.rs' --include='test_*.py' --include='*_test.py' \
  --include='*.spec.ts' --include='*.test.ts' \
  2>/dev/null || true)

if [ -n "$B3_HITS" ]; then
  while IFS= read -r line; do
    block "B3: $line"
  done <<< "$B3_HITS"
else
  pass "B3: No empty catch blocks"
fi

# ── B4: Process Exit in Test ───────────────────────────────────────────────
echo "── B4: Process Exit in Test ──"
B4_HITS=$(grep -rn \
  -e 'os\.Exit(' \
  -e 'std::process::exit(' \
  -e 'sys\.exit(' \
  -e 'process\.exit(' \
  "$TARGET_DIR" \
  --include='*_test.go' --include='*_test.rs' --include='test_*.py' --include='*_test.py' \
  --include='*.spec.ts' --include='*.test.ts' \
  --exclude-dir='.venv' \
  2>/dev/null || true)

if [ -n "$B4_HITS" ]; then
  while IFS= read -r line; do
    block "B4: $line"
  done <<< "$B4_HITS"
else
  pass "B4: No process exit in tests"
fi

# ── B6: Test Code in Production Source Files (non-idiomatic) ───────────────
echo "── B6: Inline Tests in Source Files ──"
B6_Go=$(grep -rn 'func Test' "$TARGET_DIR" \
  --include='*.go' 2>/dev/null \
  --exclude-dir=vendor \
  --exclude-dir=node_modules \
  | grep -v '_test\.go' || true)

B6_TS=$(grep -rn -e '^\s*describe(' -e '^\s*it(' -e '^\s*test(' \
  "$TARGET_DIR" \
  --include='*.ts' 2>/dev/null \
  --exclude-dir=vendor \
  --exclude-dir=node_modules \
  --exclude-dir='__tests__' \
  | grep -v -E '\.spec\.|\.test\.' || true)

B6_Python=$(grep -rn '__main__' "$TARGET_DIR" \
  --include='*.py' 2>/dev/null \
  --exclude-dir=vendor \
  --exclude-dir=node_modules \
  --exclude-dir=.venv \
  | while IFS= read -r line; do
      file=$(echo "$line" | cut -d: -f1)
      # Check if this file also has assert nearby (B7 pattern)
      if grep -q '^\s*assert ' "$file" 2>/dev/null; then
        echo "$line"
      fi
    done || true)

if [ -n "$B6_Go" ]; then
  while IFS= read -r line; do
    block "B6 (Go): $line"
  done <<< "$B6_Go"
fi

if [ -n "$B6_TS" ]; then
  while IFS= read -r line; do
    block "B6 (TS): $line"
  done <<< "$B6_TS"
fi

if [ -n "$B6_Python" ]; then
  while IFS= read -r line; do
    block "B6 (Python): $line"
  done <<< "$B6_Python"
fi

if [ -z "$B6_Go" ] && [ -z "$B6_TS" ] && [ -z "$B6_Python" ]; then
  pass "B6: No inline tests in production files"
fi

# ── B7: __main__ + Assert (Python) ─────────────────────────────────────────
echo "── B7: __main__ + Assert (Python) ──"
B7_HITS=$(find "$TARGET_DIR" -name '.venv' -prune -o -name '*.py' \
  -not -name 'test_*' -not -name '*_test.py' -print \
  2>/dev/null | while IFS= read -r file; do
    if grep -q '__main__' "$file" 2>/dev/null && grep -q '^\s*assert ' "$file" 2>/dev/null; then
      echo "$file"
    fi
  done || true)

if [ -n "$B7_HITS" ]; then
  while IFS= read -r line; do
    block "B7: $line"
  done <<< "$B7_HITS"
else
  pass "B7: No __main__+assert anti-pattern"
fi

# ── B5: Zero-Assertion Tests ───────────

echo "── B5: Zero-Assertion Tests ──"

# Heuristic: detect test functions with zero assertion calls
B5_Go=$(find "$TARGET_DIR" -name '*_test.go' 2>/dev/null \
  -exec grep -l 'func Test' {} \; 2>/dev/null | while read -r file; do
  if ! grep -qE 'assert\.|require\.' "$file" 2>/dev/null; then
    awk '/^func Test/ { name=\$0; has_assert=0; in_func=1; next } in_func && /^}/ { if (!has_assert) print FILENAME \": \" name; in_func=0 } in_func && /assert\.|require\./ { has_assert=1 }' "$file" 2>/dev/null || true
  fi
done || true)

B5_COUNT=0
if [ -n "$B5_Go" ]; then
  while IFS= read -r line; do
    block "B5 (Go): $line"
    B5_COUNT=$((B5_COUNT + 1))
  done <<< "$B5_Go"
fi

if [ "$B5_COUNT" -eq 0 ]; then
  pass "B5: No zero-assertion tests"
fi


# ── Inline Test Detection (Informational) ───
echo ""
echo "── Inline Test Detection (Informational) ──"
INLINE_RUST=$(grep -rn '#\[cfg(test)\]' "$TARGET_DIR" --include='*.rs' 2>/dev/null | wc -l) || true
if [ "$INLINE_RUST" -gt 0 ]; then
  info "Rust #[cfg(test)] modules found: ${INLINE_RUST} (idiomatic — counted as tests)"
fi

INLINE_PYTHON_DOC=$(grep -rn '>>> ' "$TARGET_DIR" --include='*.py' 2>/dev/null | wc -l) || true
if [ "$INLINE_PYTHON_DOC" -gt 0 ]; then
  info "Python doctests found: ${INLINE_PYTHON_DOC} (allowed — counted as coverage)"
fi

# ── File Inventory ──────────────────────────────────────────────────────────
echo ""
echo "── Test File Inventory ──"
GO_TESTS=$(find "$TARGET_DIR" -name '*_test.go' 2>/dev/null | wc -l)
RUST_TESTS=$(find "$TARGET_DIR" -name '*_test.rs' -o -path '*/tests/*.rs' 2>/dev/null | wc -l)
PY_TESTS=$(find "$TARGET_DIR" -name 'test_*.py' -o -name '*_test.py' 2>/dev/null | wc -l)
TS_TESTS=$(find "$TARGET_DIR" -name '*.spec.ts' -o -name '*.test.ts' 2>/dev/null | wc -l)
FILES_SCANNED=$((GO_TESTS + RUST_TESTS + PY_TESTS + TS_TESTS))

echo "   Go:     ${GO_TESTS} test files"
echo "   Rust:   ${RUST_TESTS} test files"
echo "   Python: ${PY_TESTS} test files"
echo "   TS:     ${TS_TESTS} test files"
echo "   Total:  ${FILES_SCANNED} test files scanned"

# ── TQS Calculation ──────────────────────────────────────────────────────────
TQS_INT=
TQS_GRADE_STR=
if [ "$SCORE_MODE" -eq 1 ]; then
  tqs_result=$(compute_tqs "$TARGET_DIR")
  TQS_INT=$(echo "$tqs_result" | cut -d'|' -f1)
  TQS_GRADE_STR=$(echo "$tqs_result" | cut -d'|' -f2)
  tqs_struct=$(echo "$tqs_result" | cut -d'|' -f3)
  tqs_deps=$(echo "$tqs_result" | cut -d'|' -f4)
  tqs_read=$(echo "$tqs_result" | cut -d'|' -f5)
  tqs_safe=$(echo "$tqs_result" | cut -d'|' -f6)
  tqs_cov=$(echo "$tqs_result" | cut -d'|' -f7)
  tqs_files=$(echo "$tqs_result" | cut -d'|' -f8)
fi

# Determine GATE verdict (used by both JSON and human-readable output)
if [ $EXIT_CODE -eq 0 ]; then
  GATE="PASS"
else
  GATE="BLOCK"
fi

# ── JSON Output ─────────────────────────────────────────────────────────────
JSON_OUTPUT=""
if [ "$JSON_MODE" -eq 1 ]; then
  json_tqs="null"
  json_grade="null"
  if [ -n "$TQS_INT" ]; then
    json_tqs=$(awk "BEGIN { printf \"%.1f\", $TQS_INT / 10 }" 2>/dev/null || echo "$TQS_INT")
    json_grade="\"$TQS_GRADE_STR\""
  fi
  JSON_OUTPUT=$(cat <<EOJSON
{
  "gate": "${GATE}",
  "tqs": ${json_tqs},
  "grade": ${json_grade},
  "blocking": ${BLOCK_COUNT},
  "warnings": ${WARN_COUNT},
  "b5_count": ${B5_COUNT:-0},
  "files_scanned": ${FILES_SCANNED}
}
EOJSON
)
fi

# ── Summary ─────────────────────────────────────────────────────────────────
if [ "$JSON_MODE" -eq 0 ]; then
  echo ""
  if [ -n "$TQS_INT" ] && [ "$TQS_INT" -gt 0 ]; then
    tqs_display=$(awk "BEGIN { printf \"%.1f\", $TQS_INT / 10 }" 2>/dev/null || echo "$TQS_INT")
    echo "── TQS Score ──"
    tqs_color="$GREEN"; [ "$TQS_GRADE_STR" = "fail" ] && tqs_color="$RED"
    [ "$TQS_GRADE_STR" = "bronze" ] && tqs_color="$YELLOW"
    echo -e "  ${tqs_color}TQS: ${tqs_display} (${TQS_GRADE_STR})${NC}"
    echo -e "  Groups: structural=${tqs_struct} deps=${tqs_deps} readability=${tqs_read} safety=${tqs_safe} coverage=${tqs_cov}"
    echo ""
  fi
  echo "═══════════════════════════════════════════════════"
  if [ $EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}✅ PASS${NC} — No blocking anti-patterns"
    GATE="PASS"
  else
    echo -e "${RED}⛔ BLOCK${NC} — ${BLOCK_COUNT} blocking violation(s), ${WARN_COUNT} warning(s)"
    GATE="BLOCK"
  fi
  echo "═══════════════════════════════════════════════════"
fi

# Machine-readable output
if [ "$JSON_MODE" -eq 1 ]; then
  echo "$JSON_OUTPUT"
else
  echo ""
  echo "<!-- test-quality-gate: ${GATE} -->"
  echo "<!-- blocking: ${BLOCK_COUNT} | warnings: ${WARN_COUNT} | files: ${FILES_SCANNED} | b5: ${B5_COUNT:-0} -->"
  if [ -n "$TQS_INT" ]; then
    tqs_display=$(awk "BEGIN { printf \"%.1f\", $TQS_INT / 10 }" 2>/dev/null || echo "$TQS_INT")
    echo "<!-- tqs: ${tqs_display} | grade: ${TQS_GRADE_STR} -->"
  fi
fi

exit $EXIT_CODE
