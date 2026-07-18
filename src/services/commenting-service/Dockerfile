FROM golang:1.22-alpine AS builder
WORKDIR /src

# Copy dependency manifest (prefixed for build context = services/)
COPY commenting-service/go.mod ./

# Copy replace target before go mod download so replace directive resolves
COPY shared/proto /shared/proto/

RUN go mod download

# Copy the rest of the service source
COPY commenting-service/ ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/service .

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /out/service /app/service
ENV STUB_PORT=8085
EXPOSE 8085
ENTRYPOINT ["/app/service"]
