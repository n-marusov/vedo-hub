FROM golang:1.22-alpine AS builder
WORKDIR /src

# Copy dependency manifests (prefixed for build context = services/)
COPY ticket-api/go.mod ticket-api/go.sum ./

# Copy replace target before go mod download so replace directive resolves
COPY shared/proto /shared/proto/

RUN go mod download

# Copy the rest of the service source
COPY ticket-api/ ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/service ./cmd/ticket-api/

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /out/service /app/service
ENV STUB_PORT=8088
EXPOSE 8088
ENTRYPOINT ["/app/service"]
