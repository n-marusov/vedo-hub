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

### 3) Stop and remove containers

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

## Useful Endpoints

- Frontend: `http://localhost:3000`
- API gateway health: `http://localhost:8080/health`
- Grafana: `http://localhost:3001`
- Prometheus: `http://localhost:9090`
- Loki: `http://localhost:3100/ready`
- Tempo: `http://localhost:3200/ready`

## Notes

- Observability services are inactive by default and start only with `--profile obs`.
- Grafana dashboards are provisioned from `deploy/observability/grafana/dashboards`.
