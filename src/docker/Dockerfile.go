# @ctx: Multi-stage Go builder — PLAT-LOCAL-002 / DEPLOY-OPS-001
# Build: docker build -f llm/src/docker/Dockerfile.go \
#   --build-arg BINARY_NAME=<name> --build-arg SERVICE_DIR=llm/src/<name> \
#   -t vedo-core/<name>:latest .

ARG GO_BASE_IMAGE
ARG GO_RUNTIME_IMAGE
ARG BINARY_NAME
ARG SERVICE_DIR

FROM ${GO_BASE_IMAGE} AS builder
WORKDIR /app
COPY ${SERVICE_DIR}/ .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/${BINARY_NAME} .

FROM ${GO_RUNTIME_IMAGE} AS runtime
WORKDIR /app
COPY --from=builder /app/bin/${BINARY_NAME} /app/service
EXPOSE 8080
CMD ["/app/service"]
