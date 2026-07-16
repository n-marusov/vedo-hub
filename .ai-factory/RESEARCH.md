# Research

Updated: 2026-07-16 16:00
Status: active

## Active Summary (input for $aif-plan)
<!-- aif:active-summary:start -->
Topic: Результаты ревью $aif-review ветки feature/ai-ontology-creation (17 коммитов, M2)

Goal: Зафиксировать проблемы, требующие исправления плана и кода:
- Нарушение ADR-DES.INFRA.ai-orchestration-service-strategy (AI-хендлеры в Gateway вместо отдельного сервиса)
- 15 блокирующих критических проблем (безопасность, атомарность, баги)
- План не ссылается на C4-диаграммы и актуальные ADR
- `grpcPool.Close()` в routes.go — gRPC-пул закрывается до старта сервера

Constraints:
- Менять код и план в рамках отдельных сессий ($aif-improve, $aif-fix)
- Обновить ADR и добавить C4-ссылки в план
- Рефакторинг AI в ai-orchestration-service — отдельная задача (не в M2 critical path)

Decisions:
1. $aif-improve: уточнить план M2 — добавить ссылки на C4-диаграммы, синхронизировать с ADR
2. $aif-fix: план исправлений для 15 блокирующих проблем

Open questions:
- Когда именно мигрировать AI в ai-orchestration-service? → **M2, Phase 7** (6 новых задач: 7.1–7.6)
- Нужен ли новый ADR для gRPC-аутентификации (токен-проброс)? → Добавлена задача 0.7 в Phase 0, ссылка на ADR-DES.SECURITY.authorization-policy-gates-strategy

Success signals:
- План синхронизирован с ADR и C4-диаграммами
- Fix-план покрывает все 15 Critical Issues
- RESEARCH.md сохранён до завершения всех проблем

Next step: Вызвать $aif-improve, затем $aif-fix
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
<!-- aif:sessions:end -->
