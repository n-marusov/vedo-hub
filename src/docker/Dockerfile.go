# @ctx: Centralized multi-stage Go builder
#
# Build (from project root):
#   docker build -f src/docker/Dockerfile.go \
#     --build-arg BINARY_NAME=<name> --build-arg SERVICE_DIR=src/services \
#     -t vedo-core/<name>:latest .
#
# Args:
#   GO_BASE_IMAGE    — builder image (default: golang:1.22-alpine)
#   GO_RUNTIME_IMAGE — runtime image (default: alpine:3.19)
#   BINARY_NAME      — service directory name under SERVICE_DIR, e.g. "api-gateway"
#   SERVICE_DIR      — workspace root containing service dirs + shared/ (default: src/services)

ARG GO_BASE_IMAGE=golang:1.24-alpine
ARG GO_RUNTIME_IMAGE=alpine:3.19

# @ctx: Shared multi-stage Go builder for all Go services
# Go 1.24 is required by api-gateway; backward-compatible with 1.22 modules.
# Services with SQL migrations: copy them via builder RUN + runtime COPY.

# ── Builder Stage ──────────────────────────────────────────────────────────────
FROM ${GO_BASE_IMAGE} AS builder

# Re-declare ARGs — needed inside each build stage (Docker scoping rule)
ARG BINARY_NAME
ARG SERVICE_DIR

# WORKDIR uses the Go module path so that replace directives (../shared/* etc.)
# resolve to the correct absolute paths.
WORKDIR /vedo-core/src/services/${BINARY_NAME}

# Copy dependency files first for layer caching
COPY ${SERVICE_DIR}/${BINARY_NAME}/go.mod ${SERVICE_DIR}/${BINARY_NAME}/go.sum ./

# Copy shared/ modules required by replace ../shared/* directives
# (absolute destination matches module path resolution)
COPY ${SERVICE_DIR}/shared /vedo-core/src/services/shared/

# Copy ticket-api (needed by ticket-telemetry-listener's replace ../ticket-api)
COPY ${SERVICE_DIR}/ticket-api /vedo-core/src/services/ticket-api/

# Copy vendor directory if it exists (for offline builds — checked before go mod download)
# The --link flag is omitted because COPY --link with optional sources fails silently
COPY ${SERVICE_DIR}/${BINARY_NAME}/vendor/ ./vendor/

# Download dependencies when vendor is not available (network required)
RUN if [ ! -d vendor ] || [ ! -f vendor/modules.txt ]; then go mod download; fi

# Copy all source code for the specific service
COPY ${SERVICE_DIR}/${BINARY_NAME}/ ./

# Build: use vendored deps when vendor is populated, otherwise use module cache
# go mod tidy is only needed for non-vendor builds (vendor builds are pre-synced)
RUN if [ -f vendor/modules.txt ]; then \
      CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -o /out/service .; \
    else \
      go mod tidy && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/service .; \
    fi

# Copy migrations directory if it exists (for services with SQL migrations like auth-service)
RUN if [ -d migrations ]; then cp -r migrations /out/migrations; else mkdir -p /out/migrations; fi


# ── Runtime Stage ──────────────────────────────────────────────────────────────
FROM ${GO_RUNTIME_IMAGE} AS runtime

ARG BINARY_NAME

WORKDIR /app

# ca-certificates required by any HTTPS/gRPC-TLS calls
RUN apk add --no-cache ca-certificates

COPY --from=builder /out/service /app/service

# Copy migrations directory if present (optional — some services have SQL migrations)
COPY --from=builder /out/migrations /app/migrations/

EXPOSE 8080

CMD ["/app/service"]
