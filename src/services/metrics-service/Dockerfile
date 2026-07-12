FROM python:3.12-slim AS builder
WORKDIR /build
COPY main.py ./main.py

FROM python:3.12-slim
WORKDIR /app
COPY --from=builder /build/main.py ./main.py
ENV STUB_PORT=8084
EXPOSE 8084
ENTRYPOINT ["python", "-u", "main.py"]
