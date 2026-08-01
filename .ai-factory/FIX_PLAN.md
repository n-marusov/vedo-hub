# Fix Plan: Auth regression — dynamic runtime config + SERVICE_VERSION

**Problem:** Authentication fails in dev environment. VITE_* env vars were baked into the frontend Docker image at build time, requiring a rebuild when switching between environments (dev/test/staging).
**Created:** 2026-08-01 22:00
**Status:** Implemented

## Analysis

### Root cause

`VITE_KEYCLOAK_URL` and other auth config were passed as Docker build ARGs and baked into the JS bundle by Vite at build time. When switching between dev (port 8180) and test (port 18180), the frontend Docker image retained the old URL, causing redirects to the wrong Keycloak instance.

### Impact scope

- Frontend OIDC flow redirected to wrong Keycloak port
- Any VITE_* env var change required a full frontend rebuild
- No way to use the same Docker image across environments

## Changes Implemented

### 1. Dynamic Runtime Config (VEDO_* vars)

- **`tools/dockerfiles/Dockerfile.typescript`** — removed all VITE_* build args, added `docker-entrypoint.sh` as ENTRYPOINT
- **`apps/services/frontend/docker-entrypoint.sh`** (new) — generates `/usr/share/nginx/html/config.js` at container startup from `VEDO_*` env vars
- **`apps/services/publish-browse-ui/docker-entrypoint.sh`** (new) — minimal entrypoint for public viewer
- **`apps/services/frontend/index.html`** — added `<script src="/config.js"></script>` before app bundle
- **`apps/services/publish-browse-ui/index.html`** — same
- **`apps/services/frontend/src/auth/keycloak.ts`** — `import.meta.env.VITE_*` → `window.__VEDO_CONFIG__.*`
- **`apps/services/frontend/src/auth/session.ts`** — `import.meta.env.VITE_SKIP_AUTH` → `window.__VEDO_CONFIG__.SKIP_AUTH`

### 2. Docker Compose Env Updates

- **`deploy/docker-compose.yml`** — frontend: removed `VITE_*` build args, added `VEDO_KEYCLOAK_URL`, `VEDO_KEYCLOAK_REALM`, `VEDO_KEYCLOAK_CLIENT_ID`, `VEDO_SKIP_AUTH`, `VEDO_APP_VERSION` as runtime env vars
- **`deploy/docker-compose.test.yml`** — changed `VITE_SKIP_AUTH` build arg to `VEDO_SKIP_AUTH` runtime env var

### 3. SERVICE_VERSION Added to All Services

- **`deploy/docker-compose.yml`** — `SERVICE_VERSION=${SERVICE_VERSION:-dev}` added to all 16 application services: api-gateway, ontology-service, versioning-service, auth-service, metrics-service, publish-browse-ui, publisher-service, public-browse-api, commenting-service, ticket-api, ticket-classifier, ticket-telemetry-listener, document-extractor, ai-orchestration-service, ticket-notifier
- **`deploy/docker-compose.yml`** — frontend gets `VEDO_APP_VERSION=${SERVICE_VERSION:-dev}`
- `ticket-api` retains its existing `APP_VERSION` alongside new `SERVICE_VERSION`

### 4. Makefile Updates

- **`Makefile`** — `docker-up` target now passes `SERVICE_VERSION=$(VERSION)` (version from `git describe --tags`)
- **`Makefile`** — new `docker-env-check` target shows resolved KC_HOSTNAME_PORT, VEDO_KEYCLOAK_URL, SERVICE_VERSION, FRONTEND_PORT, API_GATEWAY_PORT

## Files Modified

| File | Change |
|------|--------|
| `tools/dockerfiles/Dockerfile.typescript` | Removed VITE_* ARG/ENV, added entrypoint COPY + ENTRYPOINT |
| `apps/services/frontend/docker-entrypoint.sh` | **NEW** — generates config.js from VEDO_* env vars |
| `apps/services/publish-browse-ui/docker-entrypoint.sh` | **NEW** — minimal entrypoint |
| `apps/services/frontend/index.html` | Added config.js script tag |
| `apps/services/publish-browse-ui/index.html` | Added config.js script tag |
| `apps/services/frontend/src/auth/keycloak.ts` | window.__VEDO_CONFIG__ instead of import.meta.env |
| `apps/services/frontend/src/auth/session.ts` | window.__VEDO_CONFIG__.SKIP_AUTH instead of import.meta.env |
| `deploy/docker-compose.yml` | VEDO_* runtime env vars + SERVICE_VERSION for all services |
| `deploy/docker-compose.test.yml` | VITE_SKIP_AUTH → VEDO_SKIP_AUTH runtime |
| `Makefile` | SERVICE_VERSION in docker-up, new docker-env-check target |

## Verification

```bash
# Dev env
docker compose --env-file config/.env.dev -f deploy/docker-compose.yml config | grep -E "VEDO_|SERVICE_VERSION"
# → VEDO_KEYCLOAK_URL: http://localhost:8180, SERVICE_VERSION: dev ✓

# Test env
docker compose --env-file config/.env.test -f deploy/docker-compose.test.yml config | grep -E "VEDO_|SERVICE_VERSION"
# → VEDO_KEYCLOAK_URL: http://localhost:18180, VEDO_SKIP_AUTH: "true" ✓
```

## Risks & Considerations

- Container images must be rebuilt: `make docker-up --no-cache` or `docker compose build --no-cache frontend`
- `publish-browse-ui` also uses Dockerfile.typescript — its entrypoint is minimal (no Keycloak vars)
- `window.__VEDO_CONFIG__` is set before the app bundle loads — no race condition
- TypeScript types: `(window as any).__VEDO_CONFIG__` is untyped; consider adding a `global.d.ts` declaration later
