#!/bin/bash
# Generate vdo:verifiedBy reverse links from existing vdo:validates links.
# Usage: bash .ai-factory/scripts/generate_verifiedby.sh

TTL=".ai-factory/traceability/traceability.ttl"
TMPFILE=$(mktemp)
trap 'rm -f "$TMPFILE"' EXIT

grep 'vdo:validates' "$TTL" | grep -E 'base:ts/' > "$TMPFILE"

echo "# ============================================================================="
echo "# REVERSE LINKS: vdo:verifiedBy (auto-generated from vdo:validates)"
echo "# ============================================================================="
echo ""

awk '
function trim(s) {
  gsub(/^[ \t]+|[ \t]+$/, "", s)
  return s
}
{
  line = $0
  ts = $1

  # Extract all base:req/ and base:nfr/ URIs including dots
  # Match everything from "base:req/" or "base:nfr/" up to whitespace, comma, or period-at-end
  while (match(line, /base:(req|nfr)\/[A-Za-z0-9_.-]+/)) {
    req = substr(line, RSTART, RLENGTH)
    req = trim(req)

    # Add to map, deduplicating test suites per requirement
    if (map[req] == "") map[req] = ts
    else if (map[req] !~ "(^|,)" ts "(,|$)") map[req] = map[req] "," ts

    line = substr(line, RSTART + RLENGTH)
  }
}
END {
  for (req in map) {
    n = split(map[req], tests, ",")
    if (n == 1) {
      print req " vdo:verifiedBy " tests[1] " ."
    } else if (n == 2) {
      print req " vdo:verifiedBy " tests[1] ", " tests[2] " ."
    } else {
      printf "%s vdo:verifiedBy %s,\n", req, tests[1]
      for (i=2; i<n; i++) {
        printf "    %s,\n", tests[i]
      }
      printf "    %s .\n", tests[n]
    }
  }
}' "$TMPFILE"

echo ""
echo "# End of auto-generated verifiedBy links"
