# Research

Updated: 2026-07-22 14:00
Status: active

## Active Summary (input for $aif-plan)
<!-- aif:active-summary:start -->
Topic: DDD Context Map для VEDO Hub + Templates-via-Forks решение

Goal: (1) Сформировать оптимальную карту Bounded Contexts на основе анализа specs/ (vision.md, ADR, C4, glossary, REQ). (2) Разрешить 5 открытых вопросов: Templates ownership, Audit BC, Search BC, MCP placement, Social Hub. (3) Решение пользователя: выпилить Templates как отдельный BC, заменить на Demos Group + Forks (F13.1 переносится в MVP).

Background — pending plan (не забыть):
- project-ontology-separation: план для Project ≠ Ontology, 1:1, REST `/projects/{id}/...`, PostgreSQL `scopes.type='project'` + таблица `ontologies`. Решение принято 2026-07-21, план ещё не запущен. См. сессию 2026-07-21 18:45.

DDD Context Map (12 BC + façade):
- Core (3): Ontology/Knowledge Graph BC (ontology-service, Neo4j), Versioning BC (versioning-service, PostgreSQL), Organization/Access BC (auth-service, PostgreSQL scopes)
- Supporting MVP (4, после выпиливания Templates): AI Orchestration BC (ai-orchestration-service), Document Extraction BC (document-extractor), Publishing BC (publisher+public-browse-api), Collaboration BC (commenting-service)
- Supporting post-MVP (3): MCP BC (thin proxy), Search BC (cross-ontology, external index), Social BC (forks/stars/DOI/profile)
- Generic (4): Identity BC (auth+Keycloak), Support/Ticketing BC (5 ticket services), Metrics & Observability BC (metrics-service+Grafana), Administration BC (vedo-cli)
- Façade: API Gateway (OHS+ACL), Frontend BC (BFF), Publish Browse UI

Ключевые отношения:
- Ontology ↔ Versioning: Partnership + Materialized Cache (active branch в Neo4j, deltas в PostgreSQL). Нужен формальный Published Language `OntologySnapshot`/`TripleDelta`.
- Organization ↔ Ontology: Shared Kernel (project_id/ontology_id, 1:1 pairing)
- AI ↔ External LLM: Anti-Corruption Layer (LLM Adapter Layer + LLM Policy Router). Router мигрирует из Gateway в ai-orchestration-service (Phase 7, в прогрессе).
- Document Extraction → Ontology: Customer-Supplier + Published Language `OntologyBuildSequence` (JSON, JSON-schema validated, ApplySequence atomic в Neo4j).
- Commenting → Ontology: Conformist (entity_type+entity_id без FK, orphaned comments expected).
- API Gateway: OHS (OpenAPI=REST writes, GraphQL=read-only navigation, SPARQL=REST only с DoS-защитой per ADR rest-graphql-mutation-boundary).

Decisions (по результатам explore 2026-07-22):
1. **Templates BC выпиливается.** Шаблоны = демо-проекты в группе `VEDO Demos` + fork mechanism. F13.1 Forks переезжает из post-MVP в MVP. Теряем: semver для шаблонов (git tags богаче), auto-deprecate по usage_count (premature automation для 5-20 проектов), personal-catalog-UX (через Group "My Templates"). Приобретаем: −1 BC, −1 cross-BC Saga (ApplyTemplate), −1 Published Language (OWL+metadata.json), F13.1 в MVP, community PR через MR flow, единая fork-mechanics для templates и social hub.
2. **Fork model:** `upstream_project_id` на Project (nullable, ВАРИАНТ 1 — Git-модель). `forks_count` = count WHERE upstream_project_id=X (query или Redis cache). Endpoint: `POST /api/v1/projects/{id}/fork`.
3. **Fork Saga:** Organization create Project + Versioning copy branch + Ontology materialize. Orchestrator — TBD (не API Gateway — anti-pattern "leaky gateway").
4. **Audit: гибрид.** Shared library (`audit-rs`/`audit-go`/`audit-py`) для write path (каждый BC эмитит локально) + Audit Query BC (thin, Generic) для read path (RBAC, export, UI, masking на чтение). Published Language `AuditEvent` schema (versioned protobuf/JSON Schema). Store: Support DB (метаданные, 365 дней) + WORM S3 (critical ops, 7 лет) + S3 (полные ответы, 90 дней TTL). mcp-audit-strategy — хороший паттерн, но покрывает только MCP; нужно обобщить на все BC.
5. **Search Level 1 (in-ontology)** — остаётся в Ontology BC (Neo4j fulltext index). **Search Level 2 (cross-ontology, F16.2)** — отложить до network effect (Social Hub + MCP, post-MVP). F16.2 не имеет ни одного REQ.
6. **MCP — thin BC, post-MVP.** Transport (JSON-RPC/SSE) не подходит для API Gateway. Attack surface изоляция от internal AI. Делегирует CheckPolicy в AI Orchestration через gRPC. mcp-audit-strategy требует обновления (middleware в MCP service, не в Gateway).
7. **Social Hub (F13):** не строить сейчас, но заложить hooks: `upstream_project_id` (MVP через Decision 1), `public_profile` опционально в User (Identity BC), `doi` опционально в Publication (Publishing BC). Issues (F13.4) ≠ Support Tickets — отдельный Ontology Issues BC (post-MVP). Social Hub priority #1 в vision.md, но 0 specs — нужен ADR "post-MVP defer".

Open questions (после deep dive):
- Audit: кто владеет shared library? Не нарушает ли "no shared mutable state" из ARCHITECTURE.md? (Нет — library immutable, store shared через Published Language.)
- Templates/Fork Saga: кто orchestrator? API Gateway (anti-pattern) или отдельный orchestration service?
- Social Hub: нужен ли ADR "post-MVP defer" для формализации отсрочки priority #1?
- MCP: M2 migration plan (Phase 7) не включает MCP. Должен ли включать хотя бы contract design?
- Audit: `audit_events` таблица из ADR-DES.API.organization-rest-endpoints — где живёт? Должна быть в Support DB (изолировано от tenant).

Success signals:
- RESEARCH.md обновлён с DDD context map + 5 deep dives + Templates-via-Forks decision
- Next: $aif-plan full templates-via-forks для реализации (12 действий: удалить 9 REQ-черновиков templates-*, обновить vision.md/glossary.md, добавить upstream_project_id + fork endpoint, 5 демо-проектов как seed data, F13.1 US/UC)

Next step: $aif-plan full templates-via-forks — выпилить Templates BC, реализовать Demos + Forks (F13.1 в MVP). Параллельно pending: $aif-plan full project-ontology-separation (решение 2026-07-21, план не запущен).
<!-- aif:active-summary:end -->

## Sessions
<!-- aif:sessions:start -->
### 2026-07-16 16:00 — Анализ 4 вопросов ревью

What changed:
1. **`ontology_context` в document-extractor:** Подтверждено — мёртвый параметр. Не упоминается в плане (Task 3.4), не используется в коде. Удалить или задокументировать с security note.

2. **`buildOntologyContext`/`buildClassContext` — стабы:** План (Task 4.2) требует реальный контекст (class tree, existing properties). Текущая реализация возвращает только ID онтологии. Это незавершённая реализация — качество LLM-подсказок критически страдает.

3. **AI-хендлеры в API Gateway — нарушение ADR:** Найден ADR-DES.INFRA.ai-orchestration-service-strategy, который ПРЕДПИСЫВАЕТ выделение AI-оркестрации в отдельный сервис. Текущая реализация — прямое нарушение. C4-диаграммы (ai-orchestration-components.md, api-gateway-components.md) описывают целевую архитектуру. ADR перечисляет конкретный план миграции и чек-лист.

4. **`grpcPool.Close()` — блокирующий баг:** Подтверждено. `defer grpcPool.Close()` в `RegisterRoutes()` срабатывает до старта сервера. HTTP-прокси маскирует проблему — старые CRUD-ручки продолжают работать через HTTP fallback.

Key notes:
- ADR утверждает: AI-хендлеры, LLM Policy Router, prompt injection, templates → всё переезжает в ai-orchestration-service
- C4 api-gateway-components.md показывает Gateway с gRPC-прокси к aiOrch
- C4 ai-orchestration-components.md показывает полную целевую архитектуру AI-сервиса
- План feature-ai-ontology-creation.md не ссылается ни на ADR, ни на C4-диаграммы
- document-extractor остаётся с собственным LLM-клиентом (Hybrid Model), но политики — через gRPC
- MCP-сервер — Thin Proxy Model, не содержит LLM-клиента

Links (paths):
- specs/adr/ADR-DES.INFRA.ai-orchestration-service-strategy.md
- specs/c4/ai-orchestration-components.md
- specs/c4/api-gateway-components.md
- specs/c4/container.md
- .ai-factory/plans/feature-ai-ontology-creation.md
- src/services/api-gateway/routes.go (grpcPool.Close bug)
- src/services/api-gateway/handlers/ai_completion_handler.go (buildOntologyContext stub)
- src/services/document-extractor/api/routes.py (ontology_context dead param)
### 2026-07-16 17:00 — Миграция AI в M2 (Phase 7)

What changed:
- **Решение:** Миграция ai-orchestration-service — в M2 (Phase 7), не в M3
- Добавлены задачи 7.1–7.6 (proto contract, scaffold, миграция хендлеров, Gateway proxy, doc-extractor CheckPolicy/LogLLMUsage, Docker/CI)
- Task 6.2–6.4 возвращены в Phase 6 (были ошибочно под заголовком Phase 7)
- План обновлён: прогресс-трек (39 задач), commit plan (18 коммитов), карта зависимостей
- ADR/C4 таблицы: M3 → M2 (Phase 7)
- RESEARCH.md: открытые вопросы закрыты

Links (paths):
- .ai-factory/plans/feature-ai-ontology-creation.md (Phase 7: Tasks 7.1–7.6)
- specs/adr/ADR-DES.INFRA.ai-orchestration-service-strategy.md (чек-лист реализации)
- specs/c4/ai-orchestration-components.md (целевая архитектура)
- specs/c4/api-gateway-components.md (Gateway → aiOrch gRPC proxy)
### 2026-07-21 18:45 — Решение 1:1 (Project ↔ Ontology)

What changed:
- Пользователь принял архитектурное решение: **один Project = одна Ontology (1:1)**. Группировка нескольких онтологий — через Group (иерархия групп), не через упаковку в один Project.
- Обновлён `specs/glossary.md` запись «Проект» (L666-671): 1:1 закреплено как обязательное, не как «простейший случай». Добавлено: «Project без Ontology не существует; Ontology вне Project не существует.»
- Закрыты все open questions в Active Summary: путь переименования, отсутствие aliases, размещение visibility/policies на Project, предложение по PostgreSQL-схеме.

Key notes:
- Решение 1:1 согласуется с `context.md` L32 («одна онтология соответствует одному Git-like репозиторию»), с ADR `gitlab-like` (отклонил 1:N), с GitLab-эталоном (один Project = один repo).
- `ProjectSummary.ontology_count` в OpenAPI — противоречит 1:1 и подлежит удалению (всегда 1).
- При 1:1 aliases `/ontologies/{id}/members` ↔ `/projects/{id}/members` не нужны — они только маскируют модель. Канонический путь — `/projects/{id}/...`.
- visibility/policies канонически на Project (соответствует GitLab `/projects/:id/visibility`), не на Ontology.
- Рассуждение в ADR `gitlab-like` отклоняло «Project содержит несколько онтологий» **потому что приравнивало Project к Ontology**. После разделения сущностей рассуждение меняется на противоположное: 1:1 Project↔Ontology — это и есть GitLab-модель (один repo = один Project), только теперь Project — контейнер, а Ontology — его содержимое.

Links (paths):
- specs/glossary.md (L666-671 запись «Проект» — обновлена)
- .ai-factory/RESEARCH.md (Active Summary — закрыты open questions, Decisions расширены)
- Остальные артефакты — в плане $aif-plan full project-ontology-separation
### 2026-07-22 14:00 — DDD Context Map + 5 Deep Dives + Templates-via-Forks

What changed:
- Проведён DDD-анализ specs/ (vision.md F1-F17, C4 context/container, 8 ключевых ADR, glossary ubiquitous language).
- Сформирована карта: 12 Bounded Contexts (3 Core, 4 Supporting MVP после выпиливания Templates, 3 Supporting post-MVP, 4 Generic) + façade (API Gateway OHS, Frontend BFF).
- Углублены 5 открытых вопросов: Templates, Audit, Search, MCP, Social Hub.
- **Решение пользователя:** Templates BC выпиливается. Шаблоны = демо-проекты в группе `VEDO Demos` + fork mechanism. F13.1 Forks переносится из post-MVP в MVP.

Key notes — DDD карта:
- Core триада: Ontology BC (Neo4j), Versioning BC (PostgreSQL), Organization BC (PostgreSQL scopes). Partnership Ontology↔Versioning (Materialized Cache), Shared Kernel Organization↔Ontology (1:1 project_id).
- AI Orchestration BC — Supporting; LLM Policy Router мигрирует из Gateway в ai-orchestration-service (Phase 7, в прогрессе per RESEARCH sess. 2026-07-16).
- Document Extraction → Ontology: Customer-Supplier, Published Language `OntologyBuildSequence`.
- Commenting → Ontology: Conformist (entity_type+entity_id без FK, orphaned comments expected).
- API Gateway: OHS (OpenAPI=REST writes, GraphQL=read-only, SPARQL=REST с DoS-защитой per ADR rest-graphql-mutation-boundary).

Key notes — Templates-via-Forks:
- Templates BC выпилен. 5 MVP-шаблонов → 5 демо-проектов в group `VEDO Demos`. "Сохранить как шаблон" → `visibility=public`. Apply template → fork. Community PR → MR flow. `usage_count` → `forks_count`.
- Теряем: semver (git tags богаче), auto-deprecate по usage (premature automation для 5-20 проектов), personal-catalog-UX (через Group "My Templates").
- Приобретаем: −1 BC, −1 Saga (ApplyTemplate), −1 Published Language, F13.1 в MVP, единая fork-mechanics для templates и social hub.
- Fork model: `upstream_project_id` на Project (nullable, ВАРИАНТ 1 Git-модель). Endpoint `POST /api/v1/projects/{id}/fork`. Forks_count = count query или Redis cache.
- Fork Saga: Organization create Project + Versioning copy branch + Ontology materialize. Orchestrator TBD (не API Gateway).
- Grep `template` в src/ = 0 совпадений — templates не реализованы в коде, выпиливание безопасно (только REQ-черновики).

Key notes — Audit:
- Гибрид: shared library (`audit-rs`/`audit-go`/`audit-py`) write path + Audit Query BC read path. Published Language `AuditEvent`.
- Store: Support DB (метаданные 365 дней) + WORM S3 (critical ops 7 лет) + S3 (полные ответы 90 дней TTL).
- mcp-audit-strategy (ПРЕДЛОЖЕНО) — паттерн для MCP, но не покрывает Ontology/Versioning/Auth/Admin gRPC-операции. Нужно обобщить.
- Tension: `audit_events` из ADR organization-rest-endpoints — где живёт? Должна быть в Support DB.

Key notes — Search:
- Level 1 (in-ontology) — в Ontology BC (Neo4j fulltext index), не выделять.
- Level 2 (cross-ontology, F16.2) — отложить до network effect (Social Hub + MCP). F16.2 не имеет ни одного REQ, US, UC, ADR. Grep semantic search|elasticsearch|meilisearch = 0 совпадений.

Key notes — MCP:
- Thin BC, post-MVP. Transport (JSON-RPC/SSE) не подходит для gin Gateway. Attack surface изоляция.
- Делегирует CheckPolicy в AI Orchestration через gRPC. introspect-ontology → Ontology BC.
- Tension: mcp-audit-strategy — middleware в Gateway, но если MCP separate service, audit должен быть в MCP service. ADR нужно обновить.

Key notes — Social Hub:
- F13 priority #1 в vision.md, но 0 specs (grep fork|star|DOI|sponsor|social = 0 совпадений). Огромный vision-specs gap.
- Не строить сейчас. Hooks: `upstream_project_id` (MVP через Templates-via-Forks), `public_profile` в User (Identity BC, опц.), `doi` в Publication (Publishing BC, опц.).
- Issues (F13.4) ≠ Support Tickets — разные BC. Issues → Ontology Issues BC (post-MVP), Tickets → Ticketing BC.
- Нужен ADR "Social Hub post-MVP defer" для формализации отсрочки priority #1.

5 главных инсайтов (severity):
1. Templates — выпиливается per Decision 1 (Templates-via-Forks) — Medium
2. Audit — гибрид library + Query BC; mcp-audit покрывает только MCP, остальные BC без formal audit — High (compliance risk)
3. Search Level 2 — отложить; F16.2 vision-only — Low
4. MCP — thin BC post-MVP; transport не подходит для Gateway — Medium
5. Social Hub — priority #1 в vision, 0 specs; нужен ADR "post-MVP defer" — High (vision-specs alignment gap)

Links (paths):
- specs/vision.md (F1-F17, L150-619 функция decomposition, domain events L585-618)
- specs/c4/context.md, specs/c4/container.md (14 контейнеров)
- specs/adr/ADR-DES.INFRA.monolith-vs-microservices.md
- specs/adr/ADR-IMPL.PROCESS.repository-layout-strategy.md
- specs/adr/ADR-DES.SECURITY.gitlab-like-organization-model.md
- specs/adr/ADR-DES.INFRA.ai-orchestration-service-strategy.md
- specs/adr/ADR-DES.API.protocol-stack-strategy.md
- specs/adr/ADR-DES.API.organization-rest-endpoints.md
- specs/adr/ADR-DES.PROCESS.merge-request-strategy.md
- specs/adr/ADR-IMPL.INTEGRATION.commenting-service-architecture.md
- specs/adr/ADR-IMPL.OPS.ticket-management-system-architecture.md
- specs/adr/ADR-DES.API.llm-policy-router-strategy.md
- specs/adr/ADR-DES.INFRA.control-plane-isolation-strategy.md
- specs/adr/ADR-DES.API.rest-graphql-mutation-boundary.md
- specs/adr/ADR-DES.DATA.mcp-audit-strategy.md (ПРЕДЛОЖЕНО)
- specs/adr/ADR-DES.INFRA.support-metadata-isolation-strategy.md
- specs/adr/ADR-DES.INTEGRATION.mcp-server-query-adoption.md
- specs/glossary.md (L631-1133 VEDO platform terms)
- specs/requirements/REQ-FUN.OPS.templates-*.md (3 файла — под удаление)
- specs/requirements/REQ-FUN.API.templates-*.md (4 файла — под удаление)
- specs/requirements/REQ-FUN.DATA.templates-metadata.md (под удаление)
- specs/requirements/REQ-USR.UI.templates-*.md (4 файла — под удаление)
- specs/requirements/REQ-NFR.OPS.templates-analytics.md (под удаление)
- specs/requirements/REQ-FUN.INTEGRATION.audit-*.md (3 файла — audit pattern)
- specs/requirements/REQ-NFR.SECURITY.audit-*.md (3 файла — audit protection)
- specs/requirements/REQ-NFR.DATA.audit-retention.md
<!-- aif:sessions:end -->
