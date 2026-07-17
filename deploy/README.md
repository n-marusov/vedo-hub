# Vedo Core Docker Compose

This directory contains Docker Compose setup split into two layers:

- `docker-compose.yml` for core application services
- `docker-compose.observability.yml` for monitoring stack under `obs` profile

> **Vault (HashiCorp Vault) НЕ включён** в compose-файлы milestone 001. Для локальной разработки токены и секреты хранятся в `~/.vedo/config.yaml` или передаются через переменные окружения. Vault используется только в CI/CD и production. Подробнее: `human/artifacts/requirements/REQ-FUN.PROCESS.ci-cd-requirements.md`.

## Prerequisites

- Docker Engine with Compose v2 (`docker compose`)
- Enough free ports (see port maps in compose files)

## Run Modes

### 1) Core services only

```bash
docker compose up -d
```

### 2) Core + observability

```bash
docker compose \
  -f docker-compose.yml \
  -f docker-compose.observability.yml \
  --profile obs up -d
```

### 3) Core + LLM (local Ollama for AI-assisted ontology features)

```bash
docker compose \
  -f docker-compose.yml \
  -f docker-compose.llm.yaml \
  --profile llm up -d
```

After startup, configure the document-extractor to use the local LLM:

```bash
export LLM_PROVIDER=ollama
export LLM_PROVIDERS_OLLAMA_BASE_URL=http://ollama:11434/v1
export LLM_PROVIDERS_OLLAMA_MODEL=llama3
```

### 4) Stop and remove containers

```bash
docker compose \
  -f docker-compose.yml \
  -f docker-compose.observability.yml \
  --profile obs down
```

## Validation and Status

Validate core config:

```bash
docker compose config
```

Validate merged config with observability:

```bash
docker compose \
  -f docker-compose.yml \
  -f docker-compose.observability.yml \
  --profile obs config
```

Check container status:

```bash
docker compose ps --all
```

Tail observability logs:

```bash
docker compose \
  -f docker-compose.yml \
  -f docker-compose.observability.yml \
  --profile obs logs --tail=200 grafana otel-collector prometheus loki tempo
```

## Port Reference

| Service | HTTP (health) | gRPC (internal) | Protocol |
|---------|---------------|-----------------|----------|
| API Gateway | 8080 | — | REST/GraphQL → gRPC |
| Auth Service | 8081 | 9003 | gRPC |
| Ontology Service | 8082 | 9001 | gRPC |
| Versioning Service | 8083 | 9002 | gRPC |
| Metrics Service | 8084 | — | HTTP (Prometheus) |
| Commenting Service | 8085 | 9004 | gRPC |
| Publisher Service | 8086 | 9005 | gRPC |
| Public Browse API | 8087 | 9011 | gRPC |
| Ticket API | 8088 (CLI) | 9010 | HTTP + gRPC |
| Document Extractor | 8092 | 9013 | gRPC |
| AI Orchestration | 8093 | 9014 | gRPC |

Ports 9006–9009 are reserved for future services.
Port 9012 is reserved for Metrics Service gRPC (if needed).
Port 9014 is the AI Orchestration Service gRPC endpoint.

> **Internal communication** uses gRPC on ports 9001–9014.
> **Health checks** remain on HTTP ports (8081–8092) for Docker health probes.
> **External clients** connect to the API Gateway on port 8080 (REST/GraphQL).

### Legacy Ports (deprecated)

Pre-M2 services exposed their functional HTTP ports directly to the host.
These ports are now internal-only (`expose` in Docker Compose) or removed:

- `8082` → internal (ontology-service HTTP health)
- `8083` → internal (versioning-service HTTP health)
- `8081` → internal (auth-service HTTP health)
- `8084` → internal (metrics-service HTTP health)
- `8085` → internal (commenting-service HTTP health)
- `8086` → internal (publisher-service HTTP health)
- `8087` → internal (public-browse-api HTTP health)

## Useful Endpoints

- Frontend: `http://localhost:3000`
- API gateway health: `http://localhost:8080/health`
- Grafana: `http://localhost:3001`
- Prometheus: `http://localhost:9090`
- Loki: `http://localhost:3100/ready`
- Tempo: `http://localhost:3200/ready`

## Document Extractor & LLM Configuration

The **document-extractor** service provides AI-assisted ontology extraction from documents.
It uses an LLM provider (OpenAI-compatible or Anthropic) for natural language processing.

### Configuration

Set the following environment variables before starting the stack:

```bash
# LLM Provider (required for document-extractor functionality)
export LLM_PROVIDER=openai          # openai, anthropic, or ollama
export LLM_API_KEY=sk-...          # API key for the provider
export LLM_MODEL=gpt-4o-mini       # Model identifier

# Optional: Custom base URL for OpenAI-compatible APIs
export LLM_BASE_URL=http://localhost:11434/v1

# Provider-specific overrides
export LLM_PROVIDERS_OLLAMA_BASE_URL=http://localhost:11434/v1
export LLM_PROVIDERS_OLLAMA_MODEL=llama3
```

### Local LLM (Ollama)

For development without external API calls, use the included Ollama profile:

```bash
docker compose -f docker-compose.yml -f docker-compose.llm.yaml --profile llm up -d
```

After pulling a model (e.g., `ollama pull llama3`), configure:

```bash
export LLM_PROVIDER=ollama
export LLM_PROVIDERS_OLLAMA_BASE_URL=http://ollama:11434/v1
export LLM_PROVIDERS_OLLAMA_MODEL=llama3
docker compose up -d  # restart document-extractor with LLM config
```

### API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/upload` | POST | Upload a document for ontology extraction |
| `/api/v1/convert` | POST | NL→OWL conversion from natural language |
| `/health` | GET | Health check |
| `/ready` | GET | Readiness check |

## Notes

- Observability services are inactive by default and start only with `--profile obs`.
- Grafana dashboards are provisioned from `deploy/observability/grafana/dashboards`.
