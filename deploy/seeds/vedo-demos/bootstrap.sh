#!/bin/bash
# Bootstrap script: creates VEDO Demos group and 5 demo projects.
# Idempotent — safe to run multiple times.
#
# Usage: ./bootstrap.sh [--base-url <url>] [--token <jwt>]
#   --base-url  API Gateway base URL (default: http://localhost:8080)
#   --token     JWT token with admin privileges (default: uses VEDO_ADMIN_TOKEN env var)

set -euo pipefail

BASE_URL="${1:-http://localhost:8080}"
TOKEN="${VEDO_ADMIN_TOKEN}"

if [ -z "$TOKEN" ]; then
  echo "ERROR: VEDO_ADMIN_TOKEN environment variable is required" >&2
  exit 1
fi

AUTH_HEADER="Authorization: Bearer $TOKEN"

echo "=== VEDO Demos Bootstrap ==="

# Step 1: Create VEDO Demos group (idempotent)
echo "Creating VEDO Demos group..."
GROUP_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/groups" \
  -H "$AUTH_HEADER" \
  -H "Content-Type: application/json" \
  -d '{"name": "VEDO Demos", "description": "Demo projects for getting started with VEDO Core"}')

GROUP_ID=$(echo "$GROUP_RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

if [ -z "$GROUP_ID" ]; then
  # Group may already exist — try to find it
  GROUP_ID=$(curl -s "$BASE_URL/api/v1/groups?search=VEDO+Demos" \
    -H "$AUTH_HEADER" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
  if [ -z "$GROUP_ID" ]; then
    echo "ERROR: Failed to create or find VEDO Demos group" >&2
    exit 1
  fi
  echo "VEDO Demos group found (id: $GROUP_ID)"
else
  echo "VEDO Demos group created (id: $GROUP_ID)"
fi

# Step 2: Create demo projects from sequence files
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEMOS=("organization" "product" "process" "glossary" "event")

for demo in "${DEMOS[@]}"; do
  echo "Creating demo project: $demo..."

  # Create project in VEDO Demos group
  PROJECT_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/projects" \
    -H "$AUTH_HEADER" \
    -H "Content-Type: application/json" \
    -d "{\"name\": \"$demo\", \"description\": \"$demo demo project\", \"group_id\": \"$GROUP_ID\"}")

  PROJECT_ID=$(echo "$PROJECT_RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
  ONTOLOGY_ID=$(echo "$PROJECT_RESPONSE" | grep -o '"ontology_id":"[^"]*"' | cut -d'"' -f4)

  if [ -z "$PROJECT_ID" ]; then
    echo "  SKIP (may already exist)"
    continue
  fi

  echo "  Project created: $PROJECT_ID (ontology: $ONTOLOGY_ID)"

  # Apply sequence from JSON file
  SEQUENCE=$(cat "$SCRIPT_DIR/$demo.json")
  curl -s -X POST "$BASE_URL/api/v1/ontologies/$ONTOLOGY_ID/apply-template" \
    -H "$AUTH_HEADER" \
    -H "Content-Type: application/json" \
    -d "{\"template_id\": \"$demo\"}" \
    > /dev/null

  # Set visibility to public
  curl -s -X PUT "$BASE_URL/api/v1/projects/$PROJECT_ID/visibility" \
    -H "$AUTH_HEADER" \
    -H "Content-Type: application/json" \
    -d '{"visibility": "public"}' \
    > /dev/null

  echo "  Done: $demo (public)"
done

echo "=== Bootstrap complete ==="
echo "5 demo projects created in VEDO Demos group"
