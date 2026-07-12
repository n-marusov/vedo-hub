FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/service .

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /out/service /app/service
ENV STUB_PORT=8091
EXPOSE 8091
ENTRYPOINT ["/app/service"]
