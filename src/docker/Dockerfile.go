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

ARG GO_BASE_IMAGE=golang:1.22-alpine
ARG GO_RUNTIME_IMAGE=alpine:3.19

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

# Download and cache dependencies
RUN go mod download

# Copy all source code for the specific service
COPY ${SERVICE_DIR}/${BINARY_NAME}/ ./

# Sync go.mod/go.sum for the build environment, then compile a static binary
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/service .


# ── Runtime Stage ──────────────────────────────────────────────────────────────
FROM ${GO_RUNTIME_IMAGE} AS runtime

ARG BINARY_NAME

WORKDIR /app

# ca-certificates required by any HTTPS/gRPC-TLS calls
RUN apk add --no-cache ca-certificates

COPY --from=builder /out/service /app/service

EXPOSE 8080

CMD ["/app/service"]
