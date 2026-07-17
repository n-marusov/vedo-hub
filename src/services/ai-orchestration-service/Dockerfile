FROM golang:1.22-alpine AS builder

# Match directory structure to go.mod module paths so replace directives resolve
WORKDIR /vedo-core/src/services/ai-orchestration-service
COPY ai-orchestration-service/go.mod ai-orchestration-service/go.sum ./
COPY shared/llm /vedo-core/src/services/shared/llm/
COPY shared/proto /vedo-core/src/services/shared/proto/
RUN go mod download

COPY ai-orchestration-service/ ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/service .

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /out/service /app/service
EXPOSE 8093 9014
HEALTHCHECK --interval=15s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -q -O /dev/null http://127.0.0.1:8093/health
ENTRYPOINT ["/app/service"]
