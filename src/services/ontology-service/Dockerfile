FROM rust:1.77 AS builder
WORKDIR /src
COPY Cargo.toml ./
COPY src ./src
RUN cargo build --release

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /src/target/release/ontology-service /app/service
ENV STUB_PORT=8082
EXPOSE 8082
ENTRYPOINT ["/app/service"]
