# Research

Updated: 2026-07-21 18:30
Status: active

## Active Summary (input for $aif-plan)
<!-- aif:active-summary:start -->
Topic: Концептуальное противоречие Project vs Ontology в модели организации VEDO Core

Goal: Разрешить тождество Project = Ontology, закреплённое в ADR/REQ/context.md, в пользу модели «Project — контейнер, Ontology — содержимое (граф понятий)». Синхронизировать glossary, ADR, REQ, Antora doc, OpenAPI, routes.go, org_handler.go.

Constraints:
- Не ломать миграционный путь ADR-DES.API.rest-graphql-mutation-boundary (2026-07-21) — он ссылается на `/ontologies/{id}/members/{userId}` как уже существующий
- Сохранить GitLab-like идеологию (Group → Project → members/visibility/policies с наследованием)
- Согласовать с PostgreSQL-схемой `scopes` (type=group | type=ontology) в auth-service — возможно потребуется ввести `type=project` или переосмыслить

Decisions (по результатам explore):
1. **Project ≠ Ontology.** Project — платформенная сущность для совместной работы: среда, в которой живёт онтология, с собственным Git-like репозиторием, участниками, ролями, visibility, policies. Ontology — содержимое Project: формальная спецификация концептуализации (TBox/ABox, классы, свойства, индивиды, аксиомы).
2. Соответствие GitLab остаётся: **VEDO Project = GitLab Project** (единица доступа и версионирования). Ontology — дополнительный слой внутри Project.
3. **1:1 (Project ↔ Ontology) — обязательно.** Один Project содержит ровно одну Ontology. Группировка нескольких онтологий — через Group (иерархия групп), не через упаковку в один Project. Project без Ontology не существует; Ontology вне Project не существует.
4. Members/visibility/policies живут **на Project** (как в GitLab `/projects/:id/members|visibility`), не на Ontology. Текущие `/ontologies/{id}/members|visibility|policies` — последствие неверной идентификации Ontology с Project и подлежат переименованию в `/projects/{id}/...` (без aliases — при 1:1 алиасы только маскируют модель).
5. `ProjectSummary.ontology_count` в OpenAPI — удалить (всегда 1 при 1:1).

Open questions:
- ~~1:1 (один Project = одна Ontology) или 1:N (Project может содержать несколько онтологий)?~~ **РЕШЕНО 2026-07-21: 1:1.** Один Project содержит ровно одну Ontology. Группировка нескольких онтологий — через Group (иерархия групп), не через упаковку в один Project. Project без Ontology не существует; Ontology вне Project не существует.
- Нужно ли переименовать `/ontologies/{id}/members` → `/projects/{id}/members`, или оставить как alias? → **РЕШЕНО: переименовать.** `ontology_count` в ProjectSummary — удалить (всегда 1). При 1:1 алиасы не нужны — они только маскируют модель.
- Нужна ли visibility/policies на отдельных онтологиях внутри Project, или только на Project? → **РЕШЕНО: только на Project.** При 1:1 это эквивалентно, но канонический путь — `/projects/{id}/visibility|policies`, чтобы соответствовать GitLab-эталону `/projects/:id/visibility`.
- Что делать с PostgreSQL `scopes.type`? Сейчас `group` | `ontology`. → **Предложение:** переименовать `type='ontology'` → `type='project'`, добавить отдельную таблицу `ontologies` с FK `project_id` (1:1). Либо — оставить `scopes` как есть и ввести `projects` как views/alias. Решение — в плане.
- Текущая реализация `org_handler.go` уже передаёт scope как `"ontology/" + c.Param("id")` для members/visibility/policies — это «онтология как scope». При переходе к Project-контейнеру scope должен быть `"project/" + id`. → Будет покрыто планом.

Success signals:
- glossary.md: чёткое разделение Project (платформенный контейнер) vs Ontology (graph content); устранены формулировки «онтологий (проектов)» как синонимы; 1:1 закреплено как обязательное — ✅ ВЫПОЛНЕНО 2026-07-21 в surgical-правке
- ADR `gitlab-like-organization-model` обновлён: GitLab Project = VEDO Project, а не VEDO Ontology; 1:1 закреплено
- REQ-NFR.SECURITY.organization-access-model: убрано «Ontology — аналог GitLab Project»; убрано «Идея "Project содержит несколько онтологий" не используется» (это решение было следствием неверной идентификации — теперь рассуждение меняется на противоположное при том же выводе 1:1)
- context.md (L32): обновлено с «Ontology — аналог GitLab Project» на «Project — аналог GitLab Project, содержит одну Ontology»
- Antora organization-model.adoc: обновлено (Project ≠ Ontology, 1:1)
- OpenAPI: `ProjectSummary.ontology_count` удалён; `/ontologies/{id}/members|visibility|policies` → `/projects/{id}/members|visibility|policies`
- routes.go + org_handler.go: эндпоинты и scope-строки консистентны с новой моделью (`scope = "project/" + id`)
- PostgreSQL `scopes`: миграция `type='ontology'` → `type='project'` + новая таблица `ontologies` с FK `project_id` (или эквивалентный рефакторинг)

Next step: Выйти из explore → $aif-plan full project-ontology-separation для согласованного изменения specs/glossary.md, specs/context.md, ADR, REQ, Antora doc, openapi.json, routes.go, org_handler.go, auth-service migrations.
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
<!-- aif:sessions:end -->
