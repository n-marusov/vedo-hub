# Сводка: Пропущенные тесты до вехи M5

> Анализ: какие P0 требования в рамках M5 реально тестируемы, но не имеют тестов.
> Дата: 2026-07-23

## Ключевой вывод

**Все функциональные требования M5 уже покрыты существующими тестами.** 
Orphan P0 — это либо нетестируемые артефакты, либо требования, требующие реализации фич.

---

## M5: MVP Scope Gap Closure — Покрытие тестами

| Функциональность M5 | P0 требования | Статус | Тесты |
|--------------------|---------------|--------|-------|
| Strict group/project/member CRUD | REQ-NFR.SECURITY.organization-access-model | ✅ **Covered** | auth-service (12 тестов) + e2e |
| TBox editor polish | REQ-USR.UI.tbox-editor | ✅ **Covered** | ontology class + frontend tests |
| Minimal ABox CRUD | REQ-USR.UI.abox-editor | ✅ **Covered** | ontology individual tests |
| Separated Class/TBox/ABox views | REQ-USR.UI.graph-navigation | ✅ **Covered** | frontend graph_search tests |
| Basic Semantic Diff | REQ-FUN.DATA.versioning | ✅ **Covered** | versioning-service (8 тестов) |
| Basic merge blocking on conflicts | REQ-FUN.DATA.versioning | ✅ **Covered** | versioning merge tests |
| Comments + frontend feed/reply | REQ-FUN.INTEGRATION.collaboration | ✅ **Covered** | commenting-service tests |
| Read-only SPARQL limits | REQ-FUN.API.graphql-sparql | ✅ **Covered** | api-gateway + e2e tests |
| Forks (POST /projects/{id}/fork) | — (ADR, не P0) | ✅ **Covered** | fork acceptance/migration tests |

**Все 9 функциональных блоков M5 покрыты существующими тестами.**

---

## Orphan P0, относящиеся к зоне M0–M5 (тестируемые)

После реализации `REQ-FUN.API.sunset-header` в этой сессии, **тестируемых orphan P0 в зоне M0–M5 не осталось.**

| P0 Requirement | Статус | Почему |
|---------------|--------|--------|
| ~~REQ-FUN.API.sunset-header~~ | ✅ **DONE** | Middleware + 11 тестов |
| ~~REQ-FUN.CROSS.sequences~~ | ✅ **DONE** | Связано с document-extractor validator |
| ~~REQ-FUN.INTEGRATION.audit-integration~~ | ✅ **DONE** | Связано с CLI audit tests |
| ~~REQ-FUN.INTEGRATION.audit-log-content~~ | ✅ **DONE** | Связано с CLI audit tests |
| ~~REQ-FUN.API.owl-no-cycles~~ | ✅ **DONE** | Связано с ontology class_integration |
| ~~REQ-FUN.API.owl-syntax-validation~~ | ✅ **DONE** | Связано с ontology validation |
| ~~REQ-FUN.API.doc-extract-flow~~ | ✅ **DONE** | Связано с document-extractor routes |
| ~~REQ-FUN.API.integration~~ | ✅ **DONE** | Связано с e2e api-gateway-full |

---

## Оставшиеся orphan P0 (39 шт.) — категоризация

### 🔴 Требуют реализации фич (10 шт.)

Эти требования описывают функциональность, которая ещё не реализована в коде. 
Тесты появятся после реализации самих фич.

| P0 | Область | Статус фичи |
|----|---------|-------------|
| REQ-FUN.INTEGRATION.cost-calculation | Биллинг | Не реализовано |
| REQ-FUN.INTEGRATION.credit-top-up | Биллинг | Не реализовано |
| REQ-FUN.INTEGRATION.full-response-storage | Интеграция | Не реализовано |
| REQ-FUN.INTEGRATION.provider-billing-integration | Биллинг | Не реализовано |
| REQ-FUN.INTEGRATION.saga-pattern | Интеграция | Не реализовано |
| REQ-NFR.API.circuit-breaker | API | Не реализовано |
| REQ-NFR.API.circuit-breaker-integration | API | Не реализовано |
| REQ-FUN.DATA.gdpr-export-mandate | Data | Не реализовано |
| REQ-FUN.DATA.major-version-migration | Data | Не реализовано |
| REQ-FUN.DATA.migration-rollback | Data | Не реализовано |

### 🟡 Требуют специализированных тестов (6 шт.)

Эти требования описывают нефункциональные характеристики, требующие
специализированных бенчмарков, а не unit/integration тестов.

| P0 | Область | Что нужно |
|----|---------|-----------|
| REQ-NFR.PERF.performance | Perf | p95 < 120ms, p99 < 250ms — бенчмарки |
| REQ-NFR.PERF.canonical-workload-profile | Perf | До 1M axioms — нагрузочное тестирование |
| REQ-NFR.API.generation-latency | Perf | Латенсия NL→OWL генерации |
| REQ-NFR.API.suggestion-latency | Perf | Латенсия AI-рекомендаций |
| REQ-NFR.API.validation-latency | Perf | Латенсия валидации |
| REQ-NFR.API.excel-max-rows / max-size | Perf | Лимиты Excel (2 reqs) |

### 🟢 Требуют небольших frontend-тестов (1 шт.)

| P0 | Область | Что нужно |
|----|---------|-----------|
| REQ-USR.UI.excel-preview | UI | Frontend-тест для предпросмотра Excel при импорте |

### ⚪ Нетестируемые по природе (22 шт.)

Эти требования описывают архитектурные решения, политики, документацию,
процессы — они не могут быть верифицированы автоматическими тестами.

| Категория | Кол-во | Примеры |
|-----------|--------|---------|
| Архитектурные | 2 | REQ-FUN.API.protocol-stack (gRPC/REST), backward-compatibility |
| Политики данных | 3 | REQ-FUN.DATA.storage-stack, major-version-policy |
| Инфраструктурные | 4 | REQ-FUN.INFRA.*, REQ-NFR.INTEGRATION.failover |
| UI/доступность | 4 | REQ-NFR.UI.accessibility, wcag, data-loss, error-feedback |
| Документация | 2 | REQ-FUN.API.migration-guide, REQ-NFR.DOC.localization |
| API deprecation | 2 | REQ-NFR.API.deprecation-periods, REQ-USR.API.deprecation-notification |
| CLI | 1 | REQ-USR.INFRA.cli-admin-tool |
| Безопасность/CI | 1 | REQ-NFR.SECURITY.sca-sbom-gating |
| UI элементы | 1 | REQ-FUN.UI.external-llm-override |
| Прочие | 2 | REQ-FUN.API.excel-max-*, deprecation-periods |

---

## Итоговая карта действий до M5

| # | Задача | Требование | Сложность | Статус |
|---|-------|-----------|-----------|--------|
| 1 | Middleware Sunset/Deprecation + тесты | REQ-FUN.API.sunset-header | 🟢 1 час | ✅ **DONE** |
| 2 | Frontend-тест Excel preview | REQ-USR.UI.excel-preview | 🟡 2-3 часа | ❌ Открыто |
| 3 | Performance-бенчмарки | REQ-NFR.PERF.* | 🔴 1-2 дня | ❌ Отложено (M9-M13) |
| 4 | Интеграционные тесты биллинга | REQ-FUN.INTEGRATION.* (5) | 🔴 Зависит от фич | ❌ Блокировано |
| 5 | Circuit breaker тесты | REQ-NFR.API.circuit-breaker* | 🔴 Зависит от фич | ❌ Блокировано |

**Единственная реалистичная задача в пределах M5 — `REQ-USR.UI.excel-preview` (frontend-тест).**
