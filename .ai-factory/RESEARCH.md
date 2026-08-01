# Research

Updated: 2026-08-01 12:00
Status: active

## Active Summary (input for $aif-plan)
<!-- aif:active-summary:start -->
Topic: Publishing model + write-path invariant + GitLab-aligned REST API — exploration done; artifact revision mapped

Goal:
- Explore mode session 2026-08-01: ontology publishing (snapshot serving + access model), write-path invariant (no direct writes), external AI integration (capability scopes), REST API realignment to GitLab.
- Next: formalize via ADRs (write-invariant, REST refactoring, publishing extension) + $aif-plan; apply glossary/vision/roadmap revisions via owners.

Decisions:
1. Publishing: snapshot = read-only serving layer (CQRS; reads >> writes). Access binary: public (no auth) / restricted (API key). Publish gate = maintainer+ (no separate publisher role). Snapshot visibility decoupled from project visibility. Snapshots exposed as /releases (GitLab naming).
2. Write-path invariant: NO direct writes to ontology — nobody (humans, AI, admins), neither ABox nor TBox. Only branch → commit → MR → review → merge; every mutation = logged commit. REST writes to branch snapshot; ontology changes only via logged commit.
3. External AI (DMZ, not part of VEDO): integrates via platform API (F6) — ADR premise "no external integration" stays valid. Capability scopes for machines (NOT role ladder): read, propose:abox, propose:tbox, mr:create, publish:restricted; never write:direct, merge:self, publish:public without human. Autonomy×Authority: ABox auto, TBox deterministic gate, public = human. Review gate = deterministic code (reasoner/policy), never "AI reviews AI". Security containment: AI writes ONLY to own branches (ownership namespace branches/{client}/*); identity separation (proposer ≠ approver ≠ publisher) even in full automation; TBox/ABox split: different channels, frequencies, gates (ABox streamed, TBox versioned).
4. REST API: GitLab-aligned — nested under /projects/{pid}/ (verified vs official GitLab docs; flat contradicts GitLab; global endpoints = read-aggregations only). /api/v1/ontologies/* removed entirely. Content under /projects/{pid}/repository/* with ?branch= (GitLab ?ref= pattern). MR lifecycle 1:1 GitLab (iid, PUT merge, state_event, /diffs). Publishing = /projects/{pid}/releases. Branch protection = /projects/{pid}/protected_branches. Feature *_access_level on PUT /projects/{pid}.

Findings (code vs specs):
- ADR-DES.INFRA.ontology-publishing NOT implemented: publisher-service + public-browse-api = in-memory stubs; no separate public Neo4j; no rate limiting; publish-browse-ui = stub.
- Write invariant violated in code: ontology-service direct CRUD (classes/properties/individuals) → Neo4j without branch/commit; import replace/merge direct.
- Versioning and CRUD disconnected: commit writes delta to PostgreSQL, does NOT apply to Neo4j (only checkout re-materializes + internal state push).
- MR workflow not implemented (deferred to M10); webhooks (F6.3) declared but absent in code.
- Roles: no publisher (use maintainer weight 2); Keycloak editor="write" contradicts branch model; reviewer in realm, absent in collaboration spec.
- specs/requirements/branching-model.md — broken reference (file missing).
- Specs partially cover invariant: collaboration §3.3 "no direct main editing", storage-stack ADR "deltas per commit"; missing absolute "no direct writes by design" declaration.

Artifact revision (from session 2026-08-01; owners to apply):
- glossary.md: update restApi (project-scoped paths, branch-scoped writes), publisherService/publicBrowseApi (visibility public/restricted + API key), rbac (maintainer publish gate, capability scopes), serviceAccount/m2mToken (external AI machine identity); add terms: Snapshot/Release, API Key, Capability Scope, write-path invariant.
- vision.md: F6.1/F16.1 (project-scoped REST), F11 (snapshot visibility + API key + maintainer gate), F3 (checkout internal), MVP 2.5 REST section.
- ROADMAP.md: M5 Semantic Diff path change (/projects/{pid}/repository/commits/{id}/diffs), M10 MR workflow (invariant enforcement), M11 Publishing (binary visibility, API key, releases naming), M5 REST & Auth (branch-scoped writes).

Open questions:
- ADR-DES.API.organization-rest-endpoints: remove /api/v1/ontologies/{id} from public surface (ontology_id stays internal).
- GraphQL: args ontology_id → project_id (+branch), cursor pagination, DataLoader (ontology-service graphql/query.rs).
- HITL decision: full AI automation requires revising REQ-NFR.SECURITY.llm-write-human-approval (P0) — keep HITL or consciously replace with deterministic gates.
- Capability model: bot-user vs principal_type in membership (auth-service/membership).
- M10 MR workflow is the enforcement backbone — needed before write-invariant holds.
- Public browse shape: flat /api/v1/public/snapshots/{id} vs GitLab-nested /api/v1/public/projects/{pid}/releases/{id} — to decide (write API is nested; public perimeter has no exact GitLab analog).
- Storage alternatives for serving layer were analyzed (second Neo4j vs MinIO+index vs Neo4j Enterprise read replica) — decision: separate read-only serving store (CQRS); import mechanism for second Neo4j (neosemantics vs custom mapper) still open if Neo4j route is taken.

FULL BLAST RADIUS (from session 2026-08-01, verified against repo — this is a cross-cutting change, NOT a 3-file edit):
- Specs ADR (~20): organization-rest-endpoints (remove /api/v1/ontologies/{id}), graphql-sparql-split-strategy + protocol-stack-strategy + rest-graphql-mutation-boundary (endpoint tables /api/v1/ontologies/{id}/... → /projects/{pid}/repository/...), ontology-publishing (snapshot → releases, visibility, API key, maintainer gate), doc-extractor-service-strategy (extract/confirm paths), merge-request-strategy (MR = only path to main + automated review), monolith-vs-microservices (checkout internal), public-ontology-access (binary access), gitlab-like-organization-model (maintainer publish gate, capability scopes), storage-stack-strategy (deltas per commit, separate serving store), bola-bfla-negative-tests-mandate (endpoint-class table), vedo-cli-admin-boundary + cli-mfa-strategy (service account scopes), deprecation-policy + backward-compatibility (path migration), + all ADRs referencing /api/v1/ontologies (grep).
- Specs REQ (~10 direct + ~30 by path): ontology-publishing, DATA.versioning (paths /ontologies/{id}/commits|branches|merge → project-scoped), API.doc-extract-flow (paths), DATA.ontology-identifier-standard (GET /api/v1/ontologies/{id}), API.rest-versioning (examples), CROSS.sequences (publish/browse sequence diagrams, GET /public/ontologies/{slug}/metadata), INTEGRATION.saga-pattern (publish snapshot), PROCESS.e2e-testing (test dirs), SECURITY.llm-write-human-approval (HITL revision), SECURITY.llm-excessive-agency-control (Autonomy×Authority for publishing).
- Specs UC (~10): io.publish.publish-ontology-snapshot, browse.public.view-published-ontology, editor.classes/properties.* (branch+commit writes), git.branches/commits.* (checkout internal, branch=snapshot), team.reviews.review-merge-request, api.integration.integrate-through-platform-apis, api.webhooks.*, org.projects.manage-project-lifecycle, platform.landing.* (already in submodule bump).
- Specs US (~15): api.classes.create-rest, api.ontologies.read-rest, browse.public.view + view-accessible, io.publish.snapshot, editor.classes.*, abox.individuals.*, git.branches.* + commits.*, team.reviews.approve, E2E-api.integration.rest, E2E-versioning.branches.switch-rollback, + US with /ontologies/ (grep).
- openapi.json (apps/services/api-gateway/docs/openapi.json): ALL paths block — /ontologies (L23), /ontologies/{id} (L99), /ontologies/{id}/classes (L208), /properties (L443), /individuals (L544), /{propertyId} (L1756), /{individualId} (L1881), /export (L2006), /import (L2056), /versioning/commits* (L2106+), /versioning/branches* (L2373+) → /projects/{pid}/repository/* + new /merge_requests, /releases, /protected_branches. Keep ProjectDetail.ontology_id (internal); revise CreateOntologyRequest/UpdateOntologyRequest.
- Code (7 services): api-gateway routes.go + auth/auth.go (role weights, RequiredRoleLevel, publish gate maintainer=2) + main.go (exempt paths); ontology-service lib.rs (routes) + classes/properties/individuals.rs + handlers (export/import/validate) + graphql/query.rs (ontology_id → project_id); versioning-service routes.rs + handlers + sync_client.rs (project-scoped, MR endpoints); publisher-service storage.rs + handlers (releases, visibility, API key); public-browse-api snapshot_reader.rs + handlers (binary access); frontend api/ontology.ts + versioning.ts + validation.ts + extraction.ts + merge-requests.ts + fork.ts + org.ts + pages (OntologyWorkspace, VersioningPage, ValidationPage, PublishSnapshotDialog, ...); vedo-cli commands; shared/proto (publisher, public_browse stubs → real).
- Tests (84+ spec files): tests/e2e/specs/api/rest/* (api-gateway-full, api-integration, jwt-auth, org-api, query-execution), tests/e2e/specs/gui/flows/* (a11y, admin, ai-completion, browse, commenting-flow, document-*, editor, ontology-lifecycle), tests/e2e/pages/*.page.ts (mock intercepts **/api/v1/ontologies/*), tests/security/* (BOLA/BFLA /api/v1/ontologies/{ontologyId}/*), apps/services/frontend/src/__tests__/*.spec.ts (PublishSnapshotDialog, VersioningPage, OntologyWorkspaceSave, ...), tests/gates, tests/integration.
- Docs (Antora): developer-guide (architecture, development, testing, configuration, deployment), user-guide (versioning.adoc — protected branches, MR), integrator-guide (API reference, REST paths), admin-guide.
- Note: .ai-factory/plans/*, evolution/*, patches/* also reference old paths (M3 plan, AI-creation plan, patches) — historical, update only if still active.

Success signals:
- Exploration decisions captured; artifact revision mapped to owners.
- RESEARCH.md actualized.

Next step: $aif-plan full rest-api-realignment (write-invariant + GitLab-aligned REST + publishing) or first formalize ADRs; apply glossary/vision/roadmap edits via owners.
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
### 2026-07-22 18:00 — Templates-via-Forks & Project-Ontology Separation: implementation complete

What changed:
- templates-via-forks plan (17 tasks, 8 commits) fully implemented and merged to main.
  - Removed: template handler, models, 9 REQ drafts templates-*.
  - Added: fork endpoint POST /api/v1/projects/{id}/fork, Fork Saga (auth-service org.go), upstream_project_id, forks_count, 5 demo projects in VEDO Demos, UI (ForkDemoDialog, DemoProjectCard), GUI + security tests, Antora docs, Pencil designs.
  - Last commit d823d8e: refactor ForkDemoDialog (dedup card template, server error fallback, export slug).
- project-ontology-separation plan fully implemented and merged to main (164d02d).
  - PostgreSQL migrations (scopes.type='project', ontologies table), OpenAPI update, /projects/{id}/... paths, Antora docs, negative auth tests.
- Both branches (feature/templates-via-forks, feature/project-ontology-separation) still exist locally and in origin.

Key notes:
- Fork Saga orchestrator — auth-service org.go (decided in plan, implemented).
- Guest role can fork (confirmed: read access to source, fork creates new Project in user space).
- ROADMAP M5 now includes F13.1 Forks as part of MVP Gap Closure.
- Carry-over open questions from session 2026-07-22 14:00: Audit, Social Hub, MCP.
- F13.1 Forks was the Templates-via-Forks headline item; M5 still has TBox/ABox/class hierarchy/comments/SPARQL limits/Semantic Diff gaps.

Links (paths):
- .ai-factory/plans/feature-templates-via-forks.md (17 tasks, completed)
- .ai-factory/plans/feature-project-ontology-separation.md (completed 2026-07-21)
- src/services/auth-service/internal/org/org_handler.go (HandleForkProject)
- src/services/frontend/src/components/projects/ForkDemoDialog.vue
- src/services/frontend/src/api/fork.ts
- deploy/seeds/vedo-demos/bootstrap.sh (5 demo projects)
- specs/adr/ADR-DES.PROCESS.templates-via-forks.md
- ROADMAP.md (M5 updated with F13.1)

### 2026-08-01 12:00 — Ревизия: публикация снэпшотов, write-path инвариант, GitLab-выровненный REST API

What changed (exploration arc, starting from publishing model):
- Проанализирован ADR-DES.INFRA.ontology-publishing + REQ-CON.STACK.ontology-publishing: снэпшот = read-only serving-слой, CQRS (чтения ≫ записи), отдельное хранилище.
- Доступ к снэпшоту: бинарный public (без auth) / restricted (API key); видимость снэпшота не связана с видимостью Project; публикация — роль maintainer+ (отдельную publisher-роль не вводим).
- Зафиксирован write-path инвариант: никто (включая AI и админов) не меняет онтологию напрямую — ни ABox, ни TBox; только ветка → коммит → MR → review → merge; каждое изменение = логгируемый коммит. REST работает с локальным снимком ветки.
- Сценарий внешней AI-системы (DMZ, вне VEDO): интеграция через платформенный API (F6), premise ADR «интеграция не планируется» остаётся в силе; capability-скоупы (не лестница ролей); автономия × авторитет; review-gate = детерминированный код.
- REST API выровнен по GitLab (проверено по официальной документации: плоские пути противоречат GitLab — всё вложено в /projects/:id/, глобальные эндпоинты только read-агрегации): /api/v1/ontologies/* убрано; контент /projects/{pid}/repository/* с ?branch=; MR 1:1 GitLab (iid, PUT merge, state_event, /diffs); публикация = /projects/{pid}/releases; protected_branches отдельным ресурсом.
- Проведена ревизия затрагиваемых артефактов: glossary.md, vision.md, ROADMAP.md (детали в Active Summary и ниже).

Key notes (код vs спеки):
- ADR публикации не реализован: publisher-service и public-browse-api — in-memory заглушки; нет отдельного public Neo4j; нет rate limiting; publish-browse-ui — stub.
- Инвариант в коде нарушен: ontology-service пишет прямо в Neo4j (CRUD + import replace/merge) без ветки/коммита; versioning и CRUD развязаны (коммит пишет дельту в PostgreSQL, Neo4j не меняет; только checkout перематериализует через /internal/ontologies/{id}/state).
- MR-workflow не реализован (M10); webhooks (F6.3) заявлены, в коде нет.
- Роли: в Keycloak нет publisher (нужен гейт maintainer=2 в gateway RequiredRoleLevel); editor описан как «write» — противоречит веточной модели; reviewer есть в realm, нет в collaboration-спеке.
- specs/requirements/branching-model.md — битая ссылка (файла нет).
- Спеки частично покрывают инвариант (collaboration §3.3, storage-stack ADR «дельты = каждый коммит»), но нет абсолютной декларации «прямых записей не существует by design» — требуется новый ADR/требование.
- TBox/ABox split как модель эволюции знаний: ABox — высокочастотные факты, стрим/авто, TBox — редкие рискованные изменения, versioned + гейт; разные каналы публикации (live vs stable/release train).
- Анализ альтернатив serving-хранилища: (A) второй Neo4j — готовый граф/FTS, но двойная эксплуатация + сложность импорта (neosemantics/кастомный маппер); (B) MinIO снэпшот + индекс (Postgres JSONB/tantivy) — нулевой контакт с prod-БД, но строить индекс самим; (C) Neo4j Enterprise read replica — лучшее из обоих, но лицензия. Итог: отдельное read-only serving-хранилище (CQRS); конкретный механизм импорта — open question.
- Принципы безопасности автоматизации: AI пишет только в свои ветки (ownership namespace branches/{client}/*); разделение identity (proposer ≠ approver ≠ publisher) даже при полной автоматизации; review-gate = детерминированный код; человек — на границе public-публикации; rate limits/circuit breakers; провинанс+аудит (модель, версия, промпт, входы).

Links (paths):
- specs/adr/ADR-DES.INFRA.ontology-publishing.md
- specs/requirements/REQ-CON.STACK.ontology-publishing.md
- specs/requirements/REQ-FUN.INTEGRATION.collaboration.md (§1, §3.1, §3.3)
- specs/adr/ADR-DES.PROCESS.merge-request-strategy.md
- specs/adr/ADR-DES.API.organization-rest-endpoints.md
- specs/adr/ADR-DES.DATA.storage-stack-strategy.md
- specs/requirements/REQ-NFR.SECURITY.llm-write-human-approval.md, REQ-NFR.SECURITY.llm-excessive-agency-control.md
- .ai-factory/references/gitlab-projects-groups-api.md + официальные доки GitLab (merge_requests API: /projects/:id/...)
- apps/services/ontology-service/src/lib.rs, classes.rs, handlers/import_handler.rs
- apps/services/versioning-service/src/handlers/commit_handler.rs, services/sync_client.rs
- apps/services/publisher-service/src/storage.rs, public-browse-api/src/snapshot_reader.rs
- apps/services/api-gateway/auth/auth.go (role weights), routes.go, main.go
- deploy/keycloak/vedo-core-realm.json (роли), specs/glossary.md, specs/vision.md, .ai-factory/ROADMAP.md
