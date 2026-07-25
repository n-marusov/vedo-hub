#!/usr/bin/env bash
# test-quality-gate.sh — Static test quality gate for CI
# Scans test files for blocking anti-patterns (B1–B7) and inline tests.
# Exit 0 = PASS, Exit 1 = BLOCK
#
# Usage: ./test-quality-gate.sh [target-dir] [--strict]
#   target-dir   Directory to scan (default: src/services)
#   --strict     Treat warnings as failures

set -euo pipefail

TARGET_DIR="${1:-src/services}"
STRICT="${2:-}"
EXIT_CODE=0
BLOCK_COUNT=0
WARN_COUNT=0
FILES_SCANNED=0

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

echo "═══════════════════════════════════════════════════"
echo " Test Quality Gate — Static Analysis"
echo " Target: ${TARGET_DIR}"
echo "═══════════════════════════════════════════════════"
echo ""

# ── B1: Tautology Assertions ───────────────────────────────────────────────
echo "── B1: Tautology Assertions ──"
B1_HITS=$(grep -rn \
  -e 'assert\.True(t\?,\s*true)' \
  -e 'require\.True(t\?,\s*true)' \
  -e 'assert\.False(t\?,\s*false)' \
  -e 'assert!(true)' \
  -e 'assert_eq!(\(\w\+\),\s*\1)' \
  -e '^\s*assert\s\+True\s*$' \
  -e 'expect(true)\.toBe(true)' \
  "$TARGET_DIR" \
  --include='*_test.go' --include='*_test.rs' --include='test_*.py' --include='*_test.py' \
  --include='*.spec.ts' --include='*.test.ts' \
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
  2>/dev/null || true)
# TypeScript: only bare setTimeout (not inside Promise)
B2_TS=$(grep -rn 'setTimeout(' "$TARGET_DIR" \
  --include='*.spec.ts' --include='*.test.ts' \
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
  | grep -v '_test\.go' || true)

B6_TS=$(grep -rn -e 'describe(' -e 'it(' -e 'test(' \
  "$TARGET_DIR" \
  --include='*.ts' 2>/dev/null \
  | grep -v -E '\.spec\.|\.test\.' || true)

B6_Python=$(grep -rn '__main__' "$TARGET_DIR" \
  --include='*.py' 2>/dev/null \
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
B7_HITS=$(find "$TARGET_DIR" -name '*.py' \
  -not -name 'test_*' -not -name '*_test.py' \
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

# ── Inline Test Detection (Informational) ──────────────────────────────────
echo ""
echo "── Inline Test Detection (Informational) ──"
INLINE_RUST=$(grep -rn '#\[cfg(test)\]' "$TARGET_DIR" --include='*.rs' 2>/dev/null | wc -l)
if [ "$INLINE_RUST" -gt 0 ]; then
  info "Rust #[cfg(test)] modules found: ${INLINE_RUST} (idiomatic — counted as tests)"
fi

INLINE_PYTHON_DOC=$(grep -rn '>>> ' "$TARGET_DIR" --include='*.py' 2>/dev/null | wc -l)
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

# ── Summary ─────────────────────────────────────────────────────────────────
echo ""
echo "═══════════════════════════════════════════════════"
if [ $EXIT_CODE -eq 0 ]; then
  echo -e "${GREEN}✅ PASS${NC} — No blocking anti-patterns"
  GATE="PASS"
else
  echo -e "${RED}⛔ BLOCK${NC} — ${BLOCK_COUNT} blocking violation(s), ${WARN_COUNT} warning(s)"
  GATE="BLOCK"
fi
echo "═══════════════════════════════════════════════════"

# Machine-readable output for CI
echo ""
echo "<!-- test-quality-gate: ${GATE} -->"
echo "<!-- blocking: ${BLOCK_COUNT} | warnings: ${WARN_COUNT} | files: ${FILES_SCANNED} -->"

exit $EXIT_CODE
