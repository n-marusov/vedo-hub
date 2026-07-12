# VEDO CLI — Административная утилита командной строки

## Назначение

`vedo-cli` — единая административная утилита командной строки для экосистемы VEDO Core. Консолидирует операции, которые иначе распадаются на `kubectl`, `helm`, `neo4j-admin`, `psql`, S3-клиенты и ручные Grafana/Loki/Tempo/Prometheus запросы.

Не заменяет web UI и public API. Зона ответственности ограничена операциями, требующими автоматизации, повышенных привилегий или работы в air-gapped окружениях.

## Пользователи

- **DevOps-инженер** — backup/restore, миграции, air-gap подготовка, развёртывание, SCA/SBOM
- **Инженер поддержки (SRE)** — диагностика инцидентов по trace_id, региональная диагностика, поддержка tenant через support metadata
- **Security Lead** — управление emergency access, compliance evidence, escalation matrix, SCA-анализ
- **Администратор онтологий** — экспорт/импорт/сравнение онтологий между окружениями
- **Owner группы/онтологии** — управление членством, передача владения

## Команды

### Backup и Restore

| Команда | Описание | Guardrail |
|---------|----------|-----------|
| `vedo-cli backup create --full` | Полный backup Neo4j + PostgreSQL, загрузка в S3/MinIO, верификация | — |
| `vedo-cli backup verify` | Проверка целостности backup-пакета | — |
| `vedo-cli backup delete` | Удаление backup | G3 |
| `vedo-cli backup schedule` | Настройка расписания backup | — |
| `vedo-cli restore --id <backup_id>` | Восстановление из backup | G3 |
| `vedo-cli backup purge-request --tenant <id> --jurisdiction <jurisdiction>` | Запрос на удаление backup тенанта по юрисдикции | — |

### Миграции

| Команда | Описание | Guardrail |
|---------|----------|-----------|
| `vedo-cli migrate plan` | Чтение состояния миграций, формирование плана | — |
| `vedo-cli migrate apply` | Применение миграций PostgreSQL + Neo4j с проверкой pre-migration backup | G3 |
| `vedo-cli migrate rollback` | Откат миграции из pre-migration backup | G3 |

### Диагностика, Emergency и Кэш

| Команда | Описание |
|---------|----------|
| `vedo-cli diagnose trace --id <trace_id>` | Поиск trace в Tempo, логов в Loki, метрик в Prometheus |
| `vedo-cli diagnose` | Общая диагностика состояния системы |
| `vedo-cli diagnose region --region <id>` | Диагностика региона (health, replication lag, ошибки) для edge-region failover |
| `vedo-cli cache warmup` | Разогрев кэша Redis в регионе восстановления после failover |
| `vedo-cli emergency readonly` | Kill Switch — перевод системы в режим «только чтение» |
| `vedo-cli emergency clear` | Снятие emergency-режима |

LLM-режим в `diagnose` опционален (отключается флагом), в air-gapped окружениях отключён по умолчанию.

### Decommission и Data Lifecycle

| Команда | Описание | Guardrail |
|---------|----------|-----------|
| `vedo-cli decommission plan --tenant <id>` | Dry-run план вывода тенанта (без мутации состояния) | — |
| `vedo-cli decommission export` | Экспорт данных тенанта (Turtle, JSON Lines, Git, LFS) | — |
| `vedo-cli decommission verify --package <dir>` | Верификация экспортированного пакета | — |
| `vedo-cli decommission --purge-data` | Безвозвратное удаление данных тенанта после подтверждения экспорта | G4 |
| `vedo-cli decommission --verify-purge` | Проверка отсутствия остаточных данных | — |
| `vedo-cli account close --account-id <id> --jurisdiction <gdpr\|ru_152fz>` | Закрытие аккаунта по юрисдикции | G3 |

### Онтологии

| Команда | Описание | Guardrail |
|---------|----------|-----------|
| `vedo-cli ontology export --env <env> --canonical` | Экспорт онтологии в canonical Turtle | — |
| `vedo-cli ontology import --env <env> <file>` | Импорт онтологии из Turtle/RDF/XML | — |
| `vedo-cli ontology diff <env1> <env2>` | Semantic diff между окружениями | — |
| `vedo-cli ontology delete` | Удаление онтологии | G3 |
| `vedo-cli ontology restore --ontology-id <id>` | Восстановление онтологии из backup | — |
| `vedo-cli ontology transfer-owner --ontology-id <id> --new-owner <user>` | Передача владения онтологией | — |

### Tenant Management

| Команда | Описание | Guardrail |
|---------|----------|-----------|
| `vedo-cli tenant seed --profile canonical` | Развёртывание эталонного профиля нагрузки (1M аксиом) | — |
| `vedo-cli tenant seed --force` | Принудительный seed | G2 |
| `vedo-cli tenant delete` | Удаление тенанта | G4 |
| `vedo-cli tenant restore --tenant-id <id>` | Восстановление тенанта | — |
| `vedo-cli tenant rename --tenant-id <id> --name <name>` | Переименование тенанта | — |

### Air-Gap

| Команда | Описание |
|---------|----------|
| `vedo-cli airgap prepare` | Сбор offline-пакета (образы, Helm-чарты, docs, checksums) |
| `vedo-cli airgap prepare --advisory-bundle <file>` | Подготовка SCA advisory bundle (RustSec, Go VulnDB, PyPA, npm, NVD, Trivy) для offline-сканирования |
| `vedo-cli airgap verify package` | Проверка целостности offline-пакета |
| `vedo-cli airgap import-advisories --bundle <file>` | Импорт advisory баз в изолированном контуре |
| `vedo-cli airgap check-advisories --max-age-hours 48` | Проверка свежести advisory баз перед SCA-сканированием |

### SCA / SBOM (Supply Chain Security)

| Команда | Описание | Guardrail |
|---------|----------|-----------|
| `vedo-cli sca scan <target>` | SCA-сканирование зависимостей (RustSec, Trivy, Safety, Govulncheck) | — |
| `vedo-cli sca sbom generate <target>` | Генерация SBOM в SPDX 2.3 / CycloneDX 1.5 | — |

SCA-сканирование в air-gapped среде использует флаг `VEDO_OFFLINE_MODE=true`, импорт advisory баз выполняется через `vedo-cli airgap import-advisories`.

### Version Control

| Команда | Описание | Guardrail |
|---------|----------|-----------|
| `vedo-cli branch delete` | Удаление ветки | G1 |
| `vedo-cli commit revert` | Откат коммита | G1 |
| `vedo-cli mr create --source <branch> --target main` | Создание Merge Request | — |
| `vedo-cli mr list --status open` | Список Merge Request'ов по статусу | — |
| `vedo-cli mr review --id <id> --approve` | Ревью и утверждение Merge Request | — |
| `vedo-cli mr merge --id <id>` | Слияние одобренного Merge Request | G1 |

### Ticket Management

| Команда | Описание | Guardrail |
|---------|----------|-----------|
| `vedo-cli ticket create --title "<title>" --description "<text>" --category <category> --severity <level>` | Создание тикета через CLI | — |
| `vedo-cli ticket list --status <status> --category <category> --source <manual\|telemetry>` | Список тикетов с фильтрацией | — |
| `vedo-cli ticket get --id <ticket_id>` | Просмотр карточки тикета и истории | — |
| `vedo-cli ticket update --id <ticket_id> --priority <p0\|p1\|p2\|p3> --assignee <user>` | Обновление полей тикета | — |
| `vedo-cli ticket comment --id <ticket_id> --text "<comment>"` | Добавление комментария в тикет | — |
| `vedo-cli ticket close --id <ticket_id> --resolution "<text>"` | Закрытие тикета | — |
| `vedo-cli ticket reopen --id <ticket_id> --reason "<text>"` | Переоткрытие тикета | — |
| `vedo-cli ticket delete --id <ticket_id>` | Удаление тикета | G3 |

Все операции `vedo-cli ticket` работают с единым бэкендом подсистемы тикетов VEDO Core; события фиксируются в аудите с каналом `channel=cli`.

### Support Metadata

| Команда | Описание | Guardrail |
|---------|----------|-----------|
| `vedo-cli support tenant-info <id>` | Полная информация о tenant (контакты, SLA, deployment config) | — |
| `vedo-cli support tenant-info --search <name/email>` | Поиск tenant по названию или email администратора | — |
| `vedo-cli support audit-trail <id>` | Audit trail tenant (все события) | — |
| `vedo-cli support audit-trail <id> --since <date>` | Фильтр audit trail по дате | — |
| `vedo-cli support audit-trail <id> --type <event_type>` | Фильтр audit trail по типу события (tenant_purged и т.д.) | — |
| `vedo-cli support list-backups <id>` | Список backup tenant | — |
| `vedo-cli support list-backups <id> --type worm` | Только WORM archive | — |
| `vedo-cli support list-backups <id> --status verified` | Только верифицированные backup | — |
| `vedo-cli support emergency-access request <id> --reason "<reason>"` | Запрос break-glass ключа | G3 (Security Lead approval) |
| `vedo-cli support emergency-access list <id>` | История выданных emergency-ключей | — |
| `vedo-cli support emergency-access revoke <key_id>` | Отзыв ключа emergency-доступа | G3 |
| `vedo-cli support compliance-evidence list <id>` | Список compliance evidence tenant | — |
| `vedo-cli support compliance-evidence upload <id> --type dpa --file <file>` | Загрузка compliance evidence (DPA, SOC2, ISO27001) | G3 |
| `vedo-cli support escalation-matrix` | Текущая матрица эскалации (SLA tiers, response times) | — |
| `vedo-cli support escalation-matrix --tenant <id>` | Матрица эскалации с учётом SLA tenant | — |
| `vedo-cli support health` | Статус всех компонентов support storage (Support DB, WORM, Vault) | — |

Команды `vedo-cli support` работают полностью offline в air-gapped окружениях, обращаясь к локальной Support DB, MinIO и Vault.

## Архитектура

```
vedo-cli (binary)
  │
  ├──→ API Gateway (REST admin endpoints)
  │     ├──→ Ontology Service (Rust)
  │     ├──→ Versioning Service (Rust)
  │     └──→ Auth Service (Go / Keycloak)
  │     └──→ Ticket API (Go / Ticket Management)
  │
  ├──→ Neo4j (прямые операции: backup, migrate)
  ├──→ PostgreSQL (прямые операции: backup, migrate)
  ├──→ S3/MinIO (backup storage)
  ├──→ Tempo (query traces)
  ├──→ Loki (query logs)
  ├──→ Prometheus (query metrics)
  │
  ├──→ Support DB (отдельный PostgreSQL: tenant-info, audit-trail, escalation-matrix, tickets)
  ├──→ WORM S3 (Object Lock: compliance-evidence, audit archive)
  └──→ Vault / KMS (emergency-access keys)
```

`vedo-cli` — бинарный CLI-инструмент, не Docker-сервис. Предназначен для запуска на рабочей станции оператора, в CI/CD пайплайнах и в air-gapped окружениях без доступа к registry.

Ключевые административные операции (создание access point, привязка storage, установка квот, настройка backup policy) выполняются через REST API Gateway (`POST /api/v1/admin/*`). Прямые операции с БД (backup, migrate) выполняются напрямую.

## Аутентификация и авторизация

Решение зафиксировано в `ADR-DES.SECURITY.cli-mfa-strategy`.

### Режимы аутентификации

**1. Интерактивный (администратор в терминале):**

Выполняется OIDC Resource Owner Password Grant (ROPG) с TOTP claim в Keycloak:

```bash
$ vedo-cli backup create --full
Enter username: admin@vedo.local
Enter password: 
Enter TOTP code: 123456
✅ Authenticated (token valid for 1h)
```

- `access_token` (TTL: 1 час) — JWT с roles и tenant scope, используется для вызова API Gateway и прямых операций с БД.
- `refresh_token` кэшируется в `~/.vedo/config.yaml` (зашифрован мастер-ключом из credential chain). Повторный ввод пароля/TOTP не требуется до истечения refresh_token.
- Для команд категорий **A/B** (`backup delete`, `tenant purge`, `production restore`, `migration rollback`) MFA не кэшируется — каждый вызов требует новый TOTP, даже если `access_token` действителен.
- Для команд категории **C** (`ontology delete`, `account close`, `branch delete`) TOTP запрашивается 1 раз за lifetime refresh_token.
- Пароль не выводится на экран, не логируется, не сохраняется в shell history.

**2. Неинтерактивный (CI/CD, automation):**

```bash
$ vedo-cli backup create --token $SERVICE_ACCOUNT_TOKEN
```

- Service account в Keycloak с `client_credentials` grant.
- MFA не требуется — scope ограничен предопределённым набором операций (категория D, read-only, backup create).
- Операции категорий A/B через service account запрещены на уровне Keycloak client scopes.
- Audit логирует `actor=service-account:<client_id>`.

### Кэширование сессии

`refresh_token` хранится в `~/.vedo/config.yaml`:

```yaml
auth:
  realm: vedo-core
  client_id: vedo-cli
  refresh_token: <encrypted>
  token_endpoint: https://keycloak.vedo.local/realms/vedo-core/protocol/openid-connect/token
```

### Break-glass (Keycloak недоступен)

Используется L1 Emergency Admin из `ADR-DES.SECURITY.break-glass-access-strategy` — отдельная учётная запись вне Keycloak с паролем по схеме Шамира (M из N), доступ через `/emergency/login`. MFA не требуется — компенсируется Shamir splitting, immutable audit и немедленными уведомлениями security-команды.

### Разграничение прав

- `access_token` содержит claims: `roles` (admin, support, ontology_admin), `tenant_id`, `email`.
- Права проверяются через AuthDecision на стороне API Gateway — только роль `admin` допускает admin-операции.
- Операции скоупированы по тенанту.

### Destructive Command Guardrails (G1-G4)

| Уровень | Требования | Команды |
|---------|------------|---------|
| G1 | typed confirmation | branch delete, commit revert |
| G2 | env guard + typed confirmation | tenant seed --force |
| G3 | MFA + env guard + typed confirmation | ontology delete, account close, backup delete, migrate rollback, migrate apply, decommission purge, ticket delete, support emergency-access request, support emergency-access revoke, support compliance-evidence upload |
| G4 | cool-down delay + full backup verification + всё из G3 | tenant delete, decommission --purge-data |

Peer approval для разрушительных операций: `vedo-cli --require-peer-approval`, подтверждение через `vedo-cli approval approve <request-id>`.

## Выходные форматы

- `--format json` — machine-readable вывод для automation и CI/CD
- `--format human` (по умолчанию) — человекочитаемый вывод для интерактивной работы

## ADR-ссылки

- **ADR-DES.INFRA.vedo-cli-admin-boundary** — принято 2026-05-12. `vedo-cli` как единая административная утилита, ответственность, ограничения, альтернативы. Включает управление Merge Request'ами.
- **ADR-DES.INFRA.vedo-cli-diagnostics-entrypoint** — принято 2026-05-12. `vedo-cli diagnose trace --id <trace_id>` как точка входа в диагностику.
- **ADR-DES.INFRA.support-metadata-isolation-strategy** — принято 2026-05-16. `vedo-cli support` команды для работы с support metadata (tenant-info, audit-trail, emergency-access, compliance-evidence, escalation-matrix).
- **ADR-DES.INFRA.edge-region-failover-strategy** — принято 2026-05-16. `vedo-cli diagnose region` и `vedo-cli cache warmup` для регионального failover.
- **ADR-DES.SECURITY.supply-chain-vulnerability-policy** — принято 2026-05-16. `vedo-cli sca` и `vedo-cli airgap --advisory-bundle` для supply chain security.
- **ADR-DES.SECURITY.cli-mfa-strategy** — принято 2026-05-19. MFA для vedo-cli через Keycloak: интерактивный ROPG+TOTP и неинтерактивный service account.
- **ADR-DES.PROCESS.merge-request-strategy** — принято 2026-05-16. Процесс Merge Request как механизм внесения изменений в защищённые ветки.
- **ADR-IMPL.OPS.ticket-management-system-architecture** — принято 2026-05-18. Архитектура единой подсистемы тикетов.

## Связанные контракты

- **administration.md** (ADMIN-OPS-001) — административные операции как API и CLI-действия
- **observability-operations.md** (OBSERV-001) — `vedo-cli diagnose` как точка входа диагностики
- **deployment-operations.md** (DEPLOY-OPS-001) — P0 runbook начинается с `vedo-cli diagnose trace --id <trace_id>`
- **data-lifecycle.md** (DATA-LIFECYCLE-001) — decommission plan, canonical profile seed
- **supply-chain-security.md** — SCA-сканирование, `vedo-cli sca`, `vedo-cli airgap --advisory-bundle`
- **support-metadata-isolation.md** — Support DB, `vedo-cli support` команды

## Версионирование

`vedo-cli` версионируется вместе с VEDO Core release. Совместимость команд с версиями серверных служб поддерживается на уровне мажорной версии.

## Открытые вопросы

- Интеграция с внешними Secret Manager (HashiCorp Vault, AWS Secrets Manager) для хранения credentials backup storage?

## Принятые решения

- **Язык реализации:** Go (решение зафиксировано в `ADR-IMPL.STACK.vedo-cli-language-strategy`).
- **Модель распространения:** Single binary под каждую платформу, бинарные артефакты в релизе.
