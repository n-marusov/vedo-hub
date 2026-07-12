FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/service ./cmd/ticket-api/

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /out/service /app/service
ENV STUB_PORT=8088
EXPOSE 8088
ENTRYPOINT ["/app/service"]
