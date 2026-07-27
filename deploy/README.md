# Vedo Core Docker Compose

This directory contains Docker Compose setup split into multiple layers:

- `docker-compose.yml` — core application services
- `docker-compose.observability.yml` — monitoring stack (`obs` profile)
- `docker-compose.llm.yaml` — local LLM (Ollama, `llm` profile)
- `docker-compose.docs.yaml` — documentation servers (`documentation` profile)
- `docker-compose.test.yml` — E2E test overrides (dev JWT public key)

> **Vault (HashiCorp Vault) НЕ включён** в compose-файлы milestone 001. Для локальной разработки токены и секреты хранятся в `~/.vedo/config.yaml` или передаются через переменные окружения. Vault используется только в CI/CD и production.

## Prerequisites

- Docker Engine with Compose v2 (`docker compose`)
- Enough free ports (see port maps in compose files)

## Multi-Environment Setup

Проект поддерживает три окружения с непересекающимися портами, что позволяет
запускать их параллельно:

| Окружение | Env-файл | Смещение портов | COMPOSE_PROJECT_NAME |
|-----------|----------|----------------|---------------------|
| **dev** | `config/.env.dev` | без смещения (по умолчанию) | `vedo-core-dev` |
| **test** | `config/.env.test` | +10000 (RabbitMQ AMQP: +10001) | `vedo-core-test` |
| **staging** | `config/.env.staging` | +20000 (RabbitMQ AMQP: +20002) | `vedo-core-staging` |

### Быстрый старт

# Dev
make docker-up ENV=dev

# Test (параллельно с dev, порты не конфликтуют)
make docker-up ENV=test

# Staging (параллельно с dev и test)
make docker-up ENV=staging
```

> **Важно:** Внутренние сервисы (gRPC/REST) НЕ публикуются на host. Их порты доступны только внутри Docker bridge-сети (`vedo-network`).
> Для debug-доступа к внутреннему порту используйте: `docker compose exec <service> wget -q -O- http://127.0.0.1:<port>/health`
> или временно переопределите порт в локальном `.env.override`.

### Соглашение об именах переменных

- `XXX_PORT` — хост-порт (меняется между окружениями)
- `XXX_CONTAINER_PORT` — контейнерный порт (одинаков во всех окружениях)
- `SERVICE_PORT` / `GRPC_PORT` внутри контейнера всегда используют контейнерные порты

### Проверка конфигурации

```bash
docker compose --env-file config/.env.dev  -f deploy/docker-compose.yml config > /dev/null
docker compose --env-file config/.env.test -f deploy/docker-compose.yml config > /dev/null
docker compose --env-file config/.env.staging -f deploy/docker-compose.yml config > /dev/null
```

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

## Port Reference (dev — без смещения)

| Service | Host Port | Container Port | Protocol |
|---------|----------|---------------|----------|
| Frontend (SPA) | 3000 | 3000 | HTTP |
| Publish Browse UI | 3002 | 3002 | HTTP |
| API Gateway | 8080 | 8080 | REST/GraphQL |
| Auth Service | — (internal) | 8081 / 9003 | REST / gRPC |
| Ontology Service | — (internal) | 8082 / 9001 | REST / gRPC |
| Versioning Service | — (internal) | 8083 / 9002 | REST / gRPC |
| Metrics Service | — (internal) | 8084 | HTTP |
| Commenting Service | — (internal) | 8085 / 9004 | REST / gRPC |
| Publisher Service | — (internal) | 8086 / 9005 | REST / gRPC |
| Public Browse API | — (internal) | 8087 / 9011 | REST / gRPC |
| Ticket API | — (internal) | 8088 / 9010 | HTTP / gRPC |
| Ticket Classifier | — (internal) | 8089 | HTTP |
| Ticket Telemetry Listener | — (internal) | 8090 | HTTP |
| Ticket Notifier | — (internal) | 8091 | HTTP |
| Document Extractor | — (internal) | 8092 / 9013 | HTTP / gRPC |
| AI Orchestration | — (internal) | 8093 / 9014 | HTTP / gRPC |

| Infrastructure | Host Port | Container Port |
|---------------|----------|---------------|
| Neo4j | 7474 / 7687 | 7474 / 7687 |
| PostgreSQL | 5432 | 5432 |
| Redis | 6379 | 6379 |
| RabbitMQ | 5672 / 15672 | 5672 / 15672 |
| MinIO | 9000 / 9001 | 9000 / 9001 |
| Keycloak | 8180 | 8080 |

| Observability | Host Port | Container Port |
|--------------|----------|---------------|
| Prometheus | 9090 | 9090 |
| Grafana | 3001 | 3000 |
| Loki | 3100 | 3100 |
| Tempo | 3200 / 4317 / 4318 | 3200 / 4317 / 4318 |
| OTEL Collector | 8888 / 8889 | 8888 / 8889 |

| Extras | Host Port | Container Port |
|--------|----------|---------------|
| Ollama (LLM) | 11434 | 11434 |
| Docs (user/dev/admin/integrator) | 5000–5003 | 5000–5003 |

### Порты test и staging окружений

Для получения порта test-окружения добавьте **+10000** (кроме RabbitMQ AMQP — **+10001**).
Для staging: **+20000** (RabbitMQ AMQP — **+20002**).

Пример: API Gateway в dev = 8080, в test = 18080, в staging = 28080.

> **Применимо только к публичным сервисам** (frontend, API Gateway, publish-browse-ui,
> базы данных, Keycloak, observability, extras).
> Внутренние сервисы (gRPC/REST) не имеют host-портов и доступны только
> внутри bridge-сети Docker по container-портам.
> Их порты одинаковы во всех окружениях.

## Useful Endpoints (dev)

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
- Для смены окружения используйте соответствующий `--env-file` (`config/.env.dev`, `config/.env.test`, `config/.env.staging`).
- Контейнерные порты одинаковы во всех окружениях — межсервисное взаимодействие всегда работает через container network.
