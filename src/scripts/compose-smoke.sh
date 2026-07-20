#!/bin/bash
# @ctx: Compose smoke check for native stub domain services
set -eu

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"

# Change to the project root so Docker Compose picks up .env
cd "$ROOT_DIR/.."

services=(
  "frontend:3000"
  "publish-browse-ui:3002"
  "api-gateway:8080"
  "auth-service:8081"
  "ontology-service:8082"
  "versioning-service:8083"
  "metrics-service:8084"
  "commenting-service:8085"
  "publisher-service:8086"
  "public-browse-api:8087"
  "ticket-api:8088"
  "ticket-classifier:8089"
  "ticket-telemetry-listener:8090"
  "ticket-notifier:8091"
)

docker compose -f deploy/docker-compose.yml up -d --build

deadline=$((SECONDS + 30))
for target in "${services[@]}"; do
  service="${target%%:*}"
  port="${target##*:}"
  echo "[SMOKE] waiting for $service on $port"
  ok=0
  while [ "$SECONDS" -lt "$deadline" ]; do
    if curl -fsS "http://127.0.0.1:${port}/health" >/dev/null 2>&1; then
      ok=1
      break
    fi
    sleep 1
  done
  if [ "$ok" -ne 1 ]; then
    echo "[SMOKE] FAIL: ${service} did not become healthy within 30s"
    exit 1
  fi
done

for target in "${services[@]}"; do
  service="${target%%:*}"
  port="${target##*:}"
  if curl -fsS "http://127.0.0.1:${port}/" >/dev/null 2>&1; then
    echo "[SMOKE] PASS: ${service} root endpoint available"
  else
    echo "[SMOKE] FAIL: ${service} root endpoint unavailable"
    exit 1
  fi
done

echo "[SMOKE] PASS: all domain services healthy"
