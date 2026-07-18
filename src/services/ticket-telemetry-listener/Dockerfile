FROM golang:1.22-alpine AS builder
WORKDIR /src
RUN mkdir -p /ticket-api
COPY ticket-api /ticket-api/
COPY ticket-telemetry-listener/go.mod ticket-telemetry-listener/go.sum ./
RUN go mod download
COPY ticket-telemetry-listener ./
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/service .

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /out/service /app/service
ENV STUB_PORT=8090
EXPOSE 8090
ENTRYPOINT ["/app/service"]
