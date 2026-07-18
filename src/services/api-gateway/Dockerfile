FROM golang:1.22-alpine AS builder
WORKDIR /src

# Copy dependency manifests (prefixed for build context = services/)
COPY api-gateway/go.mod api-gateway/go.sum ./

# Copy replace targets before go mod download so replace directives resolve
COPY shared/llm /shared/llm/
COPY shared/proto /shared/proto/

RUN go mod download

# Copy the rest of the service source
COPY api-gateway/ ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/service .

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /out/service /app/service
ENV STUB_PORT=8080
EXPOSE 8080
ENTRYPOINT ["/app/service"]
