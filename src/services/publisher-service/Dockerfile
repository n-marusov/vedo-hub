FROM rust:1.89 AS builder
WORKDIR /src

# Install protoc for tonic_build to compile gRPC protobuf stubs
RUN apt-get update && apt-get install -y protobuf-compiler && rm -rf /var/lib/apt/lists/*

# Copy workspace Cargo.toml
COPY Cargo.toml ./

# Copy ALL workspace member manifests so cargo can validate the workspace
COPY shared/Cargo.toml shared/
COPY ontology-service/Cargo.toml ontology-service/
COPY versioning-service/Cargo.toml versioning-service/
COPY publisher-service/Cargo.toml publisher-service/
COPY public-browse-api/Cargo.toml public-browse-api/

# Create dummy source files for all workspace members to prefetch dependencies
RUN mkdir -p shared/src shared/proto ontology-service/src \
             versioning-service/src publisher-service/src public-browse-api/src \
    && echo "fn main() {}" > shared/build.rs \
    && echo "// placeholder" > shared/src/lib.rs \
    && echo "fn main() {}" > ontology-service/src/main.rs \
    && echo "fn main() {}" > versioning-service/src/main.rs \
    && echo "fn main() {}" > publisher-service/src/main.rs \
    && echo "fn main() {}" > public-browse-api/src/main.rs

# Cache dependencies (expected to fail due to dummy sources, that's OK)
RUN cargo build --package publisher-service --release 2>/dev/null || true

# Copy real sources
COPY shared/ shared/
COPY publisher-service/ publisher-service/

# Build for real (clean cached build output so real build.rs generates protobuf stubs)
RUN rm -rf target/release/build/vedo-shared-* 2>/dev/null; true
RUN cargo build --package publisher-service --release

FROM debian:bookworm-slim
WORKDIR /app

# Install runtime dependencies (libssl3 for reqwest, ca-certificates for HTTPS)
RUN apt-get update && apt-get install -y --no-install-recommends \
    libssl3 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /src/target/release/publisher-service /app/service
ENV SERVICE_PORT=8086
EXPOSE 8086
ENTRYPOINT ["/app/service"]
