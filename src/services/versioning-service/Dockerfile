FROM rust:1.77 AS builder
WORKDIR /src
COPY Cargo.toml ./
COPY src ./src
RUN cargo build --release

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /src/target/release/versioning-service /app/service
ENV STUB_PORT=8083
EXPOSE 8083
ENTRYPOINT ["/app/service"]
