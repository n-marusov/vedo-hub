FROM python:3.12-slim AS builder

WORKDIR /build

# Install build system dependencies for native extensions
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    libc6-dev \
    && rm -rf /var/lib/apt/lists/*

COPY pyproject.toml ./
RUN pip install --no-cache-dir --upgrade pip \
    && pip install --no-cache-dir \
        fastapi==0.111.0 \
        uvicorn[standard]==0.30.1 \
        grpcio==1.64.1 \
        grpcio-tools==1.64.1 \
        pydantic==2.7.4 \
        pydantic-settings==2.3.4 \
        python-docx==1.1.2 \
        PyPDF2==3.0.1 \
        pdfplumber==0.11.0 \
        openpyxl==3.1.5 \
        aiofiles==24.1.0 \
        httpx==0.27.0 \
        prometheus-client==0.20.0

# Copy all source code
COPY . .


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

COPY --from=builder /build /app/

ENV SERVICE_PORT=8092
EXPOSE 8092
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD python -c "import http.client; conn=http.client.HTTPConnection('localhost', 8092); conn.request('GET', '/health'); resp=conn.getresponse(); exit(0 if resp.status==200 else 1)"

ENTRYPOINT ["python", "-u", "main.py"]
