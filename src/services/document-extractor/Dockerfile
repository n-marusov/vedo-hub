FROM python:3.12-slim AS builder

WORKDIR /build

# Install build system dependencies for native extensions
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    libc6-dev \
    && rm -rf /var/lib/apt/lists/*

# Install uv
COPY --from=ghcr.io/astral-sh/uv:latest /uv /uvx /bin/

# Copy dependency files first for layer caching
COPY pyproject.toml uv.lock ./

# Install production dependencies with uv
RUN uv sync --frozen --no-dev --no-editable

# Copy all source code
COPY . .

# Re-run uv sync to install the package itself
RUN uv sync --frozen --no-dev --no-editable


# ─── Runtime Stage ─────────────────────────────────────────────────────────────

FROM python:3.12-slim

WORKDIR /app

# Install runtime system dependencies for PDF processing
RUN apt-get update && apt-get install -y --no-install-recommends \
    libglib2.0-0 \
    libsm6 \
    libxext6 \
    libxrender-dev \
    libgomp1 \
    && rm -rf /var/lib/apt/lists/*

# Copy installed dependencies from builder
COPY --from=builder /build/.venv /app/.venv
# Copy all source code
COPY --from=builder /build /app/

ENV PATH="/app/.venv/bin:$PATH"
ENV SERVICE_PORT=8092
EXPOSE 8092
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD python -c "import http.client; conn=http.client.HTTPConnection('localhost', 8092); conn.request('GET', '/health'); resp=conn.getresponse(); exit(0 if resp.status==200 else 1)"

ENTRYPOINT ["python", "-u", "main.py"]
