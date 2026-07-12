FROM rust:1.77 AS builder
WORKDIR /src
COPY Cargo.toml ./
COPY src ./src
RUN cargo build --release

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /src/target/release/public-browse-api /app/service
ENV STUB_PORT=8087
EXPOSE 8087
ENTRYPOINT ["/app/service"]
