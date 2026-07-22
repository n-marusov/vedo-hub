# VEDO Hub Roadmap

> GitHub + Hugging Face for ontologies: a team platform for creating, governing, versioning, and AI-assisted ontology projects through Web UI, REST API, and MCP.

## Milestones

### MVP — minimal working product demonstration

- [x] **M0: Foundation & Infrastructure** — Microservices monorepo, Docker Compose orchestration, CI/CD baseline, OpenTelemetry/Grafana/Prometheus/Loki/Tempo observability, Keycloak authentication, health/readiness endpoints, and the initial `vedo-cli` operator boundary.
- [x] **M1: Ontology Core Engine Baseline** — Core ontology storage and editing backend for TBox/ABox entities, graph browsing primitives, Git-like commits/history/branches/rollback, REST/GraphQL access paths, and SPARQL execution baseline.
- [x] **M2: Document-to-Ontology & Demos Baseline** — Backend/API baseline for AI-assisted ontology creation from text/documents and predefined demo projects (templates replaced by Demos+Forks per ADR-DES.PROCESS.templates-via-forks): parsing, preview sequence generation, user confirmation/apply flow, ontology update, commit creation, and original document artifact attachment. Baseline GUI coverage is limited to document upload/preview/apply wiring; NL→OWL prompt UI, iterative refinement controls, AI class/property suggestion panels, and advanced batch deduplication UI are carried into M4/M5/M9 as noted below.
- [x] **M3: Multi-Team Organization Model** — GitLab-like groups, subgroups, ontology projects, project catalog, visibility levels, Owner/Developer/Reporter/Guest-style memberships, role inheritance, last-owner protection, and group/project/member APIs.
- [x] **M4: MVP GUI Wiring & Frontend Integration** — Replace hardcoded frontend placeholders with real API/GraphQL calls where backend capabilities exist; keep future-only functions visible as disabled/read-only GUI stubs. This milestone provides the first coherent Web UI path through groups, projects, ontology editing, versioning, SPARQL, dashboard widgets, and planned-feature placeholders.

    **Блок A — Быстрые победы (API + GraphQL-запросы готовы, только подключить):**
    - [x] **A1. VersioningPage** — подключить `GET_COMMIT_HISTORY_QUERY` и `GET_BRANCHES_QUERY` (запросы уже в `queries.ts`, бэкенд versioning-service работает). Заменить хардкодные `commits[]`, `branches[]`, `tags[]`, `graphNodes[]` на реальные Apollo-запросы. Добавить состояния загрузки/ошибки.
    - [x] **A2. OntologyWorkspace (Save)** — подключить кнопку «Save» к `UPDATE_DRAFT_MUTATION` (мутация уже в `queries.ts`). Добавить optimistic UI и dirty-state management через `useDraftState`.
    - [x] **A3. SPARQLPage** — дописать `onRunQuery()`: Apollo-вызов к SPARQL GraphQL-резолверу (endpoint уже работает). Подключить результат в `SPARQLQueryEditor`. Добавить `onFormatQuery()` — клиентское форматирование или серверный format-эндпоинт.

    **Блок Б — Новые GraphQL-запросы (бэкенд готов, запросов на фронтенде нет):**
    - [x] **Б1. ProjectsPage** — добавить `LIST_PROJECTS_QUERY` в `queries.ts` → подключить к странице. Заменить хардкодный `projects[]` на данные из ontology-service (список онтологий с метаданными). Подключить поиск и сортировку.
    - [x] **Б2. GroupsPage** — добавить `LIST_GROUPS_QUERY` → подключить к странице. Заменить `groupRows[]` на данные из auth-service. Реализовать раскрытие/сворачивание групп с ленивой загрузкой подгрупп.
    - [x] **Б3. MembersPage** — добавить `LIST_MEMBERS_QUERY` и мутации `UPDATE_MEMBER_ROLE`, `REMOVE_MEMBER` → подключить к странице. Заменить хардкодный `members[]`. Подключить обработчики кнопок Edit/Delete.
    - [x] **Б4. VersioningPage (Tags, Graph, Compare)** — добавить `GET_TAGS_QUERY` для списка тегов. Подключить `GRAPH_NEIGHBORHOOD_QUERY` (уже в `queries.ts`) для Repository Graph. Добавить `COMPARE_REVISIONS_QUERY` для DiffView.

    **Блок В — Требуют мок-заглушек на бэкенде (сервисы есть, но frontend-facing API отсутствует или требует эмуляции):**
    - [x] **В1. DashboardPage** — создать aggregate-запрос `DASHBOARD_QUERY` (суммирует MR, комментарии, активность из comment/versioning сервисов). Подключить виджеты, attention items, activity feed, recent ontologies к реальным данным.
    - [x] **В2. MetricsPage** — добавить `ONTOLOGY_METRICS_QUERY` → подключить к metrics-service (Python). Заменить хардкодные KPI и chart placeholders на реальные данные. Добавить реальные графики (chart.js или аналог).
    - [x] **В3. ValidationPage** — SHACL-валидатор на бэкенде — **заглушка**, которая возвращает мок-данные с фиктивным статусом `OK` и пустым списком нарушений. Фронтенд: добавить `RUN_VALIDATION_MUTATION` и `VALIDATION_REPORT_QUERY` → подключить к кнопке «Run validation». `ValidationReport` organism уже готов принимать props.
    - [x] **В4. DeploymentsPage + MergeRequestsPage** — добавить мок-резолверы на бэкенде для `/api/v1/deployments` и `/api/v1/merge-requests`, возвращающие правдоподобные фиктивные данные. На фронтенде: добавить `LIST_DEPLOYMENTS_QUERY` и `LIST_MERGE_REQUESTS_QUERY` → подключить к organism-компонентам.

    **Блок Г — Навигация и общие элементы:**
    - [x] **Г1. Header/Sidebar** — подключить навигационные элементы к реальным данным пользователя (аватар, имя из Keycloak-сессии). Активная подсветка текущего маршрута.
    - [x] **Г2. DashboardPage (онтологии)** — строки «Recent project» сделать кликабельными, ведущими на `OntologyWorkspace`.

    **Блок Д — Carry-over wiring from M2/M3 backend capabilities (GUI entry points missing or partial):**
    - [x] **Д1. OntologyWorkspace (NL→OWL)** — wire a visible NL prompt input and Generate action to the existing AI generation backend/API; render the generated sequence in the existing preview/apply workflow.
    - [x] **Д2. OntologyWorkspace (Iterative refinement)** — wire feedback/refinement controls to the existing refinement endpoint; preserve previous sequence state and show refined steps before apply.
    - [x] **Д3. AI completion panels** — wire class/property suggestion entry points, ranked suggestions, accept/reject actions, and confidence/rationale display to ai-orchestration completion APIs.
    - [x] **Д4. Ontology template GUI flow** — verify and finish the user-facing template selection/apply workflow against real backend/API data, not only test fixtures.
    - [x] **Д5. Ontology entity create workflows** — wire class/property/individual create buttons and dialogs in `OntologyWorkspace` to existing ontology-service/api-gateway CRUD; backend CRUD is part of M1 and must not be treated as missing.
    - [x] **Д6. Organization GUI read-after-write** — ensure groups/projects/members created through REST APIs appear in `GroupsPage`, `ProjectsPage`, and `MembersPage` via the same persisted backend data source.

    **Технические требования ко всем страницам:**
    - Каждая страница должна проходить три состояния: **loading** (скелетон/спиннер), **error** (понятное сообщение + кнопка Retry), **data** (рендер данных).
    - Все запросы — через Apollo Client с кешированием и optimistic UI где применимо.
    - Обработка пустых состояний (empty state) — вместо «No commits yet» показывать контекстный CTA (например, «Create your first class»).
    - Единый подход к обработке ошибок через `useErrorPresentation` composable.

- [ ] **M5: MVP Scope Gap Closure** — Close the remaining gaps between implemented baseline capabilities and the final MVP scope from `specs/vision.md` section 2.5: strict group/project/member CRUD semantics, TBox editor polish, minimal ABox CRUD, separated Class Hierarchy / TBox Graph / ABox Graph views, basic Semantic Diff, basic merge blocking on conflicts, entity/project comments with frontend feed/reply wiring to commenting-service, read-only SPARQL limits (`LIMIT`, timeout, audit),
    **F13.1 Forks** — `POST /api/v1/projects/{id}/fork`, `upstream_project_id` on Project, 5 demo projects in `VEDO Demos` group as seed data (`deploy/seeds/vedo-demos/bootstrap.sh`).
- [ ] **M6: MCP Server & API Contract for MVP** — Deliver the MVP MCP server for ontology work: group/project discovery, class/property/individual read tools, ontology search, commit history, basic Semantic Diff, read-only SPARQL, safe TBox/ABox write tools, `create_commit`, shared RBAC/audit with REST, OpenAPI documentation, and explicit `501`/`x-vedo-status: planned` stubs for post-MVP APIs.
- [ ] **M7: MVP Acceptance, Security & Demo Readiness** — Stabilize the minimal end-to-end demo chain: group → project → members → demo project (fork)/document → ontology generation → TBox/ABox edit → class/TBox/ABox navigation → commit/history/diff → REST/MCP. Add authorization regression checks for group/project roles, destructive-action safety, empty/error/loading states, smoke tests, demo seed data, and concise operator runbook.

### Version 1.0 — productization after MVP

- [ ] **M8: Query Experience 1.0** — Upgrade query capabilities beyond MVP: visual SPARQL builder, ontology introspection for query construction, saved query library, unified table/graph results, result export to CSV/JSON/Turtle, explicit opt-in NL-query mode, and Visual CYPHER Builder as a planned or experimental capability.
- [ ] **M9: Document AI 1.0** — Extend document-to-ontology beyond MVP: OCR for scanned PDFs, password/encrypted-document UX, expanded batch upload semantics, advanced merged-source deduplication, XML/CSV/XLSX polish, custom prompts, relationship hints beyond baseline suggestions, and quality review of generated ontology fragments. Baseline NL→OWL/refinement GUI wiring is not deferred here when backend/API already exists; it belongs to M4/M5.
- [ ] **M10: Collaboration & Review 1.0** — Add full collaboration workflows: threaded discussions, resolve/unresolve, Merge/Pull Request workflow, review comments, approval workflow, branch protection, manual conflict resolution, advanced Semantic Diff, and visual graph diff.
- [ ] **M11: Publishing, Import/Export & Public Browse 1.0** — Implement deferred exchange and publication capabilities: OWL/RDF import/export, XLSX/DOCX exchange where needed, snapshot publishing pipeline, public read-only browse API/UI, public ontology visibility model, public search, publication metadata, and DOI-ready metadata without making DOI mandatory for MVP.
- [ ] **M12: Ontology Quality & Reasoning 1.0** — Add advanced ontology quality capabilities: SHACL/OWL validation API, reasoner integration, semantic linter, complex OWL constructs, property characteristics, entity resolution/link discovery, quality reports, and validation gates for review workflows.
- [ ] **M13: Operations, Support & Analytics 1.0** — Productize operations and support: metrics/analytics dashboards, telemetry with explicit opt-in, support portal, feedback/NPS flows, status page, scheduled backup/restore verification, migration rollback checks, incident diagnostics by `trace_id`, and support SLA/escalation artifacts.

### Post-1.0 growth

- [ ] **M14: Enterprise, Resilience & Compliance** — SaaS/on-premise/air-gapped deployment tiers, data residency, control-plane isolation, HA/failover, RTO/RPO evidence, PITR, restore/decommission drills, secure erase, privileged access with JIT/PAM/MFA, staged rollout with auto-rollback, and release evidence gates.
- [ ] **M15: Scale, Ecosystem & Marketplace** — Large-graph visualization with WebGL/virtualization, advanced ontology analytics, dedicated enterprise APIs/SLA, GitHub/GitLab sync plugins, custom template marketplace, public profiles, stars/issues, sponsoring, and partner commissions.

## Completed

| Milestone | Date |
|-----------|------|
| M0: Foundation & Infrastructure | 2026-07-12 |
| M1: Ontology Core Engine Baseline | 2026-07-14 |
| M2: Document-to-Ontology & Template Baseline | 2026-07-18 |
| M3: Multi-Team Organization Model | 2026-07-19 |
| M4: MVP GUI Wiring & Frontend Integration | 2026-07-20 |

## Roadmap Notes

- **MVP source of truth:** MVP scope is governed by `specs/vision.md` section 2.5. The MVP is intentionally limited to the demonstrable chain: group → project → members → demo project (fork)/document → ontology generation → TBox/ABox editing → class/TBox/ABox graph navigation → commit/history/basic Semantic Diff → REST API/MCP.
- **Continuous numbering:** Milestones use a single continuous `M*` sequence. Former `M2.1` is now `M3`; former `M2.5` is now `M4`. Completed statuses and completed dates were preserved.
- **Organization model:** The GitLab-like group/project/member model is part of MVP, not a later enterprise feature. It defines ownership, visibility, RBAC inheritance, and audit boundaries for all ontology work.
- **Document AI priority:** Document-to-ontology generation is a high-priority MVP capability because the target user often starts from real `DOCX`, text `PDF`, `XLSX`, `JSON`, `MD`, or `TXT` artifacts rather than from hand-written OWL/RDF. M2 completed the backend/API and baseline upload→preview→apply path; missing GUI entry points for NL→OWL, refinement, and AI suggestions are tracked as M4/M5 wiring gaps. MVP excludes OCR, encrypted-document password handling, custom prompts, advanced merged-source deduplication, and template-as-document-context.
- **Demos in MVP:** 5 demo projects in `VEDO Demos` group serve as onboarding accelerators for the editor. Users fork demo projects via `POST /api/v1/projects/{id}/fork` — fork replaces the former `Ontology Template` concept (ADR-DES.PROCESS.templates-via-forks). Fork is also the base mechanism for Social Hub (F13.1). Demo project marketplace, custom public projects as demos, and merge-upstream-changes are post-MVP.
- **Navigation split:** MVP must not collapse class hierarchy and graph visualization into one view. It needs separate Class Hierarchy View, TBox Graph View, and ABox Graph View.
- **Semantic Diff boundary:** Basic Semantic Diff is required for MVP. Visual graph diff, reasoning-aware diff, manual conflict resolution, and Merge/Pull Request review workflows are post-MVP.
- **MCP boundary:** A minimal MCP server is part of MVP and must share RBAC, audit logging, and safety limits with REST/UI operations. Advanced agentic workflows, auto-merge, document extraction through MCP, and admin operations through MCP are post-MVP.
- **Import/export boundary:** OWL/RDF import/export is deliberately deferred after MVP. The MVP focuses on demo projects, document generation, and fork mechanics as the primary onboarding paths.
- **Stub policy:** Future functions should appear as disabled/read-only GUI stubs and documented API stubs returning `501 Not Implemented` or marked with `x-vedo-status: planned`. Stubs stabilize navigation and API contracts without expanding MVP implementation scope.
- **1.0 direction:** Version 1.0 turns the MVP into a product by expanding query UX, document AI, collaboration/review workflows, import/export/publishing, validation/reasoning, operations, support, and analytics.
- **Scope constraint:** Real-time co-editing with pessimistic node locks is not a near-term roadmap driver; collaboration is primarily modeled through projects, roles, comments, branches, Semantic Diff, and later Merge/Pull Request workflows.
- **Continuous NFRs:** Performance, security, observability, accessibility, API documentation, localization, and release evidence apply continuously across milestones rather than as one-off deliverables.
