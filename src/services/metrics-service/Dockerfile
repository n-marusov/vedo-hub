FROM python:3.12-slim AS builder
WORKDIR /build
COPY pyproject.toml uv.lock ./
COPY main.py ./main.py
COPY analytics/ ./analytics/
COPY collectors/ ./collectors/
COPY prometheus_exporter.py ./prometheus_exporter.py

FROM python:3.12-slim
WORKDIR /app
COPY --from=builder /build /app/
RUN pip install --no-cache-dir fastapi==0.111.0 uvicorn[standard]==0.30.1 aio-pika redis prometheus-client
ENV SERVICE_PORT=8084
EXPOSE 8084
ENTRYPOINT ["python", "-u", "main.py"]
