FROM python:3.12-slim AS builder

WORKDIR /build

# Install uv
COPY --from=ghcr.io/astral-sh/uv:latest /uv /uvx /bin/

# Copy dependency files first for layer caching
COPY pyproject.toml uv.lock ./

# Install production dependencies with uv
RUN uv sync --frozen --no-dev --no-editable

# Copy source code
COPY main.py ./main.py
COPY analytics/ ./analytics/
COPY collectors/ ./collectors/
COPY prometheus_exporter.py ./prometheus_exporter.py

# Re-run uv sync to install the package itself
RUN uv sync --frozen --no-dev --no-editable


# ─── Runtime Stage ─────────────────────────────────────────────────────────────

FROM python:3.12-slim

WORKDIR /app

# Copy installed dependencies and source from builder
COPY --from=builder /build/.venv /app/.venv
COPY --from=builder /build/main.py /app/main.py
COPY --from=builder /build/analytics /app/analytics
COPY --from=builder /build/collectors /app/collectors
COPY --from=builder /build/prometheus_exporter.py /app/prometheus_exporter.py

ENV PATH="/app/.venv/bin:$PATH"
ENV SERVICE_PORT=8084
EXPOSE 8084
ENTRYPOINT ["python", "-u", "main.py"]
