FROM python:3.12-slim AS builder

WORKDIR /build

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

# Copy installed dependencies from builder
COPY --from=builder /build/.venv /app/.venv
COPY --from=builder /build /app/

ENV PATH="/app/.venv/bin:$PATH"
ENV SERVICE_PORT=8089
EXPOSE 8089

ENTRYPOINT ["python", "-u", "main.py"]
