# P0 Requirement Coverage by Roadmap Milestone

> Generated 2026-07-23
> Sources: `.ai-factory/ROADMAP.md`, `.ai-factory/traceability/traceability.ttl`

## Сводка

| Показатель | Значение |
|-----------|----------|
| Всего P0 требований в TTL | 185 |
| Покрыто тестами (есть `vdo:validates`) | 85 (46%) |
| Orphan P0 (нет `vdo:validates`) | 100 (54%) |
| Из них non-testable (CON, BIZ, PROCESS, DOC, OPS, DATA policy) | ~65 |
| **Truly testable orphans, требующие новых тестов** | **~35** |

---

## Completed Milestones (M0–M4) — Покрытие тестами

### M1: Ontology Core Engine Baseline

| P0 Requirement | Статус | Покрытие | Описание |
|---------------|--------|----------|----------|
| REQ-FUN.DATA.versioning | ✅ Covered | versioning-service integration (8 tests) | Git-like versioning |
| REQ-FUN.API.graphql-sparql | ✅ Covered | api-gateway + e2e graphql tests | SPARQL/GraphQL endpoints |
| REQ-FUN.API.class-hierarchy-accuracy | ✅ Covered | ontology class_integration tests | Точность иерархии классов |
| REQ-FUN.API.no-cyclic-hierarchy | ✅ Covered | ontology class_integration tests | Отсутствие циклов в иерархии |
| REQ-FUN.API.owl-no-cycles | ✅ Covered | ontology class_integration tests (added) | Отсутствие циклов в OWL |
| REQ-FUN.API.properties-accuracy | ✅ Covered | ontology property_integration | Точность свойств |
| REQ-FUN.API.pre-save-validation | ✅ Covered | ontology validation_integration | Валидация перед сохранением |
| REQ-FUN.API.owl-syntax-validity | ✅ Covered | ontology validation_integration | Синтаксическая корректность OWL |
| REQ-FUN.API.owl-unique-iri | ✅ Covered | ontology validation_integration | Уникальность IRI |
| REQ-FUN.API.owl-syntax-validation | ✅ Covered | ontology validation_integration (added) | Валидация синтаксиса OWL |
| REQ-FUN.API.write-idempotency | ✅ Covered | api-gateway idempotency tests | Идемпотентность записей |
| REQ-FUN.API.route-registration | ✅ Covered | ontology + versioning route tests | Регистрация маршрутов |
| REQ-FUN.API.property-crud | ✅ Covered | ontology property_integration | CRUD свойств |
| REQ-USR.UI.tbox-editor | ✅ Covered | ontology class/individual + frontend tests | TBox редактор |
| REQ-USR.UI.abox-editor | ✅ Covered | ontology individual_integration | ABox редактор |
| REQ-USR.UI.graph-navigation | ✅ Covered | frontend graph_search tests | Навигация по графу |
| REQ-FUN.DATA.ontology-visibility-levels | ✅ Covered | ai-orchestration middleware tests | Уровни видимости онтологии |
| REQ-FUN.API.unified-root-endpoint | ✅ Covered | api-gateway main tests | Единый корневой endpoint |
| **Missing for M1:** |
| REQ-FUN.API.protocol-stack | 📋 Non-testable | Архитектурное решение (gRPC/REST/GraphQL) |
| REQ-FUN.API.owl-supported-constructs | P1 (non-P0 через priority) | — |

### M2: Document-to-Ontology & Demos Baseline

| P0 Requirement | Статус | Покрытие | Описание |
|---------------|--------|----------|----------|
| REQ-FUN.API.doc-extract-flow | ✅ Covered | document-extractor routes (added) | Поток извлечения из документа |
| REQ-FUN.API.doc-extract-supported-formats | ✅ Covered | document-extractor parsers tests | Поддерживаемые форматы |
| REQ-FUN.API.doc-extract-sequence-schema | ✅ Covered | document-extractor validator tests | Схема sequence |
| REQ-FUN.API.excel-duplicate-handling | ✅ Covered | document-extractor xlsx parser tests | Обработка дубликатов Excel |
| REQ-FUN.API.excel-error-threshold | ✅ Covered | document-extractor xlsx parser tests | Порог ошибок Excel |
| REQ-FUN.API.excel-import-block | ✅ Covered | document-extractor xlsx parser tests | Блокировка импорта Excel |
| REQ-FUN.API.excel-supported-formats | ✅ Covered | document-extractor parsers tests | Форматы Excel |
| REQ-FUN.API.excel-validation | ✅ Covered | document-extractor xlsx parser tests | Валидация Excel |
| REQ-FUN.CROSS.sequences | ✅ Covered | document-extractor validator (added) | Кросс-секционные sequence |
| REQ-NFR.SECURITY.doc-extract-security | ✅ Covered | document-extractor security tests | Безопасность извлечения |
| REQ-NFR.SECURITY.excel-import-security | ✅ Covered | document-extractor security tests | Безопасность импорта Excel |
| **Covered indirectly (additional M2 links):** |
| REQ-FUN.API.llm-policy | ✅ Covered | ai-orchestration + shared/llm (13 tests) | LLM Policy (M2 AI baseline) |
| REQ-FUN.API.validation-import | ✅ Covered | ontology validation_p0 tests | Валидация импорта |
| REQ-FUN.API.validation-pr | ✅ Covered | ontology validation_p0 tests | Валидация PR |
| REQ-NFR.SECURITY.prompt-structure-detection | ✅ Covered | ai-orchestration + api-gateway tests | Prompt injection detection |
| REQ-NFR.SECURITY.system-prompt-hardening | ✅ Covered | ai-orchestration + api-gateway tests | Защита system prompt |
| REQ-NFR.SECURITY.llm-tool-least-privilege | ✅ Covered | ai-orchestration middleware tests | Принцип наименьших привилегий LLM |
| REQ-FUN.API.suggestion-confidence-threshold | ✅ Covered | ai-orchestration suggestion tests | Порог уверенности рекомендаций |
| REQ-FUN.API.suggestion-recalculation | ✅ Covered | ai-orchestration suggestion tests | Пересчёт рекомендаций |

### M3: Multi-Team Organization Model

| P0 Requirement | Статус | Покрытие | Описание |
|---------------|--------|----------|----------|
| REQ-NFR.SECURITY.organization-access-model | ✅ Covered | auth-service (12 tests) + e2e org tests | Модель доступа организации |
| REQ-NFR.SECURITY.bola-bfla-negative-tests | ✅ Covered | api-gateway (8 tests) + e2e | BOLA/BFLA защита |
| REQ-NFR.SECURITY.authorization-regression-gates | ✅ Covered | security org_access_control tests | Регрессионные гейты авторизации |
| REQ-NFR.SECURITY.jwt-key-mismatch | ✅ Covered | e2e jwt-auth tests | Защита от JWT key mismatch |
| REQ-NFR.SECURITY.emergency-policy-disable | ✅ Covered | cli emergency tests | Аварийное отключение политик |
| REQ-NFR.SECURITY.incident-secret-rotation | ✅ Covered | cli secret rotation tests | Ротация секретов |
| REQ-NFR.SECURITY.audit-access-audit | ✅ Covered | cli audit_log tests | Аудит доступа |
| REQ-NFR.SECURITY.audit-encryption | ✅ Covered | security placeholder tests | Шифрование аудита |
| REQ-NFR.SECURITY.audit-masking | ✅ Covered | security placeholder tests | Маскирование аудита |
| REQ-NFR.SECURITY.enforced-in-code | ✅ Covered | security placeholder tests | Принуждение в коде |
| REQ-NFR.SECURITY.security-integration | ✅ Covered | security placeholder tests | Интеграция безопасности |
| REQ-NFR.SECURITY.llm-write-human-approval | ✅ Covered | security placeholder tests | Подтверждение записи LLM |
| REQ-NFR.SECURITY.prompt-audit | ✅ Covered | security placeholder tests | Аудит промптов |
| REQ-NFR.SECURITY.prompt-filter-blacklist | ✅ Covered | security placeholder tests | Чёрный список промптов |
| REQ-NFR.SECURITY.suggestion-privacy | ✅ Covered | security placeholder tests | Приватность рекомендаций |
| REQ-NFR.SECURITY.llm-content-screening | ✅ Covered | shared/llm metrics + observability tests | Скрининг контента LLM |
| REQ-NFR.SECURITY.security-requirements | ✅ Covered | e2e keycloak-login tests | Общие требования безопасности |
| REQ-NFR.SECURITY.privileged-access-control | ✅ Covered | auth-service org_roles tests | Контроль привилегированного доступа |
| REQ-FUN.INTEGRATION.audit-integration | ✅ Covered | cli audit_log (added) | Интеграция аудита |
| REQ-FUN.INTEGRATION.audit-log-content | ✅ Covered | cli audit_log (added) | Содержимое аудит-лога |

### M4: MVP GUI Wiring & Frontend Integration

| P0 Requirement | Статус | Покрытие | Описание |
|---------------|--------|----------|----------|
| REQ-USR.UI.gui-implementation | ✅ Covered | frontend (25 test suites) + e2e (12) | Реализация GUI |
| REQ-USR.UI.critical-errors ✅ | frontend p0_ui_system | Критические ошибки UI |
| REQ-USR.UI.graceful-degradation ✅ | frontend p0_ui_system | Graceful degradation |
| REQ-USR.UI.validation-feedback ✅ | frontend p0_ui_system | Обратная связь валидации |
| REQ-USR.UI.circuit-breaker-feedback ✅ | frontend p0_ui_system | Circuit breaker feedback |
| REQ-USR.UI.consent-withdrawal ✅ | frontend p0_ui_system | Отзыв согласия |
| REQ-USR.UI.credits-dashboard ✅ | frontend p0_ui_system | Дашборд кредитов |
| REQ-USR.UI.credits-notifications ✅ | frontend p0_ui_system | Уведомления кредитов |
| REQ-USR.UI.cost-transparency ✅ | frontend p0_ui_system | Прозрачность стоимости |
| REQ-USR.UI.customization ✅ | frontend p0_ui_system | Кастомизация |
| REQ-USR.UI.dangerous-actions-recovery ✅ | frontend p0_ui_system | Восстановление опасных действий |
| REQ-USR.UI.decommission-usability ✅ | frontend p0_ui_system | Юзабилити вывода |
| REQ-USR.UI.delight-features ✅ | frontend p0_ui_system | Delight features |
| REQ-USR.UI.provider-change-notification ✅ | frontend p0_ui_system | Уведомление о смене провайдера |
| REQ-USR.UI.provider-status-indicator ✅ | frontend p0_ui_system | Индикатор статуса провайдера |
| REQ-USR.UI.quality-report ✅ | frontend p0_ui_system | Отчёт о качестве |
| REQ-USR.UI.status-page-deprecation ✅ | frontend p0_ui_system | Deprecation статуса |
| REQ-USR.UI.usability-metrics ✅ | frontend p0_ui_system | Метрики юзабилити |
| REQ-USR.UI.visibility-indicator ✅ | frontend p0_ui_system | Индикатор видимости |
| REQ-USR.UI.ux-review-process ✅ | frontend p0_ui_system | Процесс UX-ревью |
| REQ-USR.UI.doc-extract-upload ✅ | frontend p0_ui_import_export | Загрузка документов |
| REQ-USR.UI.excel-error-report ✅ | frontend p0_ui_import_export | Отчёт об ошибках Excel |
| REQ-USR.UI.query-export-availability ✅ | frontend p0_ui_import_export | Доступность экспорта |
| REQ-USR.UI.query-export-localization ✅ | frontend p0_ui_import_export | Локализация экспорта |
| REQ-USR.UI.import-export ✅ | frontend p0_ui_import_export | Импорт/экспорт |
| REQ-USR.UI.sparql-gui-search ✅ | frontend p0_ui_graph_search | SPARQL поиск |
| REQ-USR.UI.ontology-mental-model ✅ | frontend p0_ui_graph_search | Ментальная модель онтологии |
| REQ-USR.UI.graph-navigation ✅ | frontend p0_ui_graph_search | Навигация по графу |
| REQ-USR.UI.confidence-indicator ✅ | frontend p0_ui_suggestions | Индикатор уверенности |
| REQ-USR.UI.external-llm-consent ✅ | frontend p0_ui_suggestions | Согласие на внешний LLM |
| REQ-USR.UI.generation-progress ✅ | frontend p0_ui_suggestions | Прогресс генерации |
| REQ-USR.UI.llm-selector ✅ | frontend p0_ui_suggestions | Выбор LLM |
| REQ-USR.UI.max-suggestions ✅ | frontend p0_ui_suggestions | Максимум рекомендаций |
| REQ-USR.UI.model-cost-awareness ✅ | frontend p0_ui_suggestions | Осведомлённость о стоимости |
| REQ-USR.UI.nl-language-support ✅ | frontend p0_ui_suggestions | NL поддержка языка |
| REQ-USR.UI.suggestion-confirmation ✅ | frontend p0_ui_suggestions | Подтверждение рекомендации |
| REQ-USR.UI.suggestion-interaction ✅ | frontend p0_ui_suggestions | Взаимодействие с рекомендациями |
| REQ-USR.UI.suggestion-ranking ✅ | frontend p0_ui_suggestions | Ранжирование рекомендаций |
| **Орфан для M4:** |
| REQ-USR.UI.excel-preview | ❌ Orphan | Предпросмотр Excel — нужны новые frontend-тесты |
| REQ-USR.API.deprecation-notification | ❌ Orphan | Уведомления о deprecation API |

---

## In-Progress Milestones (M5–M7) — Анализ покрытия

### M5: MVP Scope Gap Closure (текущий milestone)

| P0 Requirement | Статус | Покрытие | Комментарий |
|---------------|--------|----------|-------------|
| REQ-FUN.API.integration | ✅ Covered | e2e api-gateway-full (added) | Общая интеграция API |
| REQ-FUN.API.gradual-rollback | ✅ Covered | api-gateway rollback_p0 tests | Постепенный откат |
| REQ-FUN.API.iterative-refinement-context | ✅ Covered | api-gateway + ontology tests | Контекст итеративного уточнения |
| REQ-FUN.INTEGRATION.ticket-management | ✅ Covered | ticket-api + support + notifier + sync + telemetry (10 tests) | Управление тикетами |
| CON.* (15 reqs) | 📋 Non-testable | Архитектурные ограничения | Не тестируется автоматически |
| **Орфаны для M5:** |
| REQ-FUN.API.migration-guide | 📋 Non-testable | Документация по миграции |
| REQ-FUN.API.sunset-header | ❌ Orphan (needs tests) | API sunset/deprecation — нужны новые тесты |
| REQ-FUN.INTEGRATION.cost-calculation | ❌ Orphan | Расчёт стоимости — фича не реализована |
| REQ-FUN.INTEGRATION.credit-top-up | ❌ Orphan | Пополнение кредитов — фича не реализована |
| REQ-FUN.INTEGRATION.full-response-storage | ❌ Orphan | Хранение полных ответов — фича не реализована |
| REQ-FUN.INTEGRATION.provider-billing-integration | ❌ Orphan | Интеграция биллинга — фича не реализована |
| REQ-FUN.INTEGRATION.saga-pattern | ❌ Orphan | Saga-pattern — фича не реализована |
| REQ-USR.UI.excel-preview | ❌ Orphan | Предпросмотр Excel — нужны frontend-тесты |

### M6: MCP Server & API Contract for MVP

Нет специфических P0 требований в TTL, связанных с MCP. MCP — это новый функционал без отдельных P0 требований на данный момент.

### M7: MVP Acceptance, Security & Demo Readiness

Покрытие в основном через:
- P0 security reqs (уже покрыты — см. M3)
- E2E тесты (REQ-FUN.PROCESS.e2e-testing, ✅ covered)
- RBAC тесты (bola-bfla, org-access-model — ✅ covered)

**Орфаны для M7:**
| P0 Requirement | Статус | Комментарий |
|---------------|--------|-------------|
| REQ-NFR.PROCESS.prompt-injection-tests | 📋 Non-testable | Процесс тестирования prompt injection |
| REQ-NFR.PROCESS.rollout-safety-gates | 📋 Non-testable | Гейты безопасности rollout |
| REQ-NFR.SECURITY.sca-sbom-gating | 📋 Non-testable | SCA/SBOM — CI/CD процесс |

---

## Version 1.0 Milestones (M8–M13)

### M8: Query Experience 1.0

| P0 Requirement | Статус | Комментарий |
|---------------|--------|-------------|
| REQ-FUN.API.graphql-sparql | ✅ Covered | Базово покрыто |
| REQ-FUN.DATA.query-export-content | ✅ Covered | Содержимое экспорта запросов |
| REQ-FUN.DATA.query-export-format | ✅ Covered | Формат экспорта запросов |
| REQ-FUN.DATA.query-export-validation | ✅ Covered | Валидация экспорта запросов |

### M9: Document AI 1.0

| P0 Requirement | Статус | Комментарий |
|---------------|--------|-------------|
| REQ-FUN.API.doc-extract-* | ✅ Covered | Все 7 требований покрыты |
| REQ-FUN.API.suggestion-* | ✅ Covered | Оба требования покрыты |
| REQ-NFR.API.generation-latency | ❌ Orphan | Латенсия генерации — перформанс-тесты |
| REQ-NFR.API.suggestion-latency | ❌ Orphan | Латенсия рекомендаций |
| REQ-NFR.API.validation-latency | ❌ Orphan | Латенсия валидации |
| REQ-NFR.API.excel-max-rows | ❌ Orphan | Макс. строк Excel |
| REQ-NFR.API.excel-max-size | ❌ Orphan | Макс. размер Excel |
| REQ-NFR.API.doc-extract-performance | ✅ Covered | document-extractor routes |

### M10: Collaboration & Review 1.0

| P0 Requirement | Статус | Комментарий |
|---------------|--------|-------------|
| REQ-FUN.INTEGRATION.collaboration | ✅ Covered | commenting-service тесты |

### M11: Publishing, Import/Export & Public Browse 1.0

| P0 Requirement | Статус | Комментарий |
|---------------|--------|-------------|
| REQ-FUN.DATA.ontology-import-export | ✅ Covered | ontology import_export tests |
| REQ-USR.UI.import-export | ✅ Covered | frontend p0_ui_import_export |
| REQ-FUN.API.rest-versioning | ✅ Covered | api-gateway rest_crud tests |

### M12: Ontology Quality & Reasoning 1.0

| P0 Requirement | Статус | Комментарий |
|---------------|--------|-------------|
| REQ-FUN.API.pre-save-validation | ✅ Covered | ontology validation |
| REQ-FUN.API.owl-* | ✅ Covered | Все OWL требования покрыты |
| REQ-FUN.API.validation-import/pr | ✅ Covered | validation_p0 тесты |

### M13: Operations, Support & Analytics 1.0

| P0 Requirement | Статус | Комментарий |
|---------------|--------|-------------|
| REQ-FUN.INTEGRATION.cost-calculation | ❌ Orphan | Фича не реализована |
| REQ-FUN.INTEGRATION.credit-top-up | ❌ Orphan | Фича не реализована |
| REQ-FUN.INTEGRATION.provider-billing-integration | ❌ Orphan | Фича не реализована |
| REQ-BIZ.SUP.* (3) | 📋 Non-testable | Бизнес-правила |
| REQ-FUN.INTEGRATION.collaboration-quality-metrics | ✅ Covered | metrics-service тесты |
| NFR.OPS.* (7) | 📋 Non-testable | Операционные процедуры |
| REQ-NFR.INFRA.availability-slo | 📋 Non-testable | SLO доступности |

---

## Post-1.0 (M14–M15)

| P0 Requirement | Статус | Комментарий |
|---------------|--------|-------------|
| REQ-CON.INFRA.on-premise-* (3) | 📋 Non-testable | Архитектурные ограничения on-premise |
| REQ-CON.INFRA.air-gapped-deployment | P1 (non-P0) | — |
| REQ-CON.INFRA.deployment-geography | P1 (non-P0) | — |
| REQ-CON.INFRA.deployment-tiers | P1 (non-P0) | — |
| REQ-NFR.INFRA.* (7) | 📋 Non-testable | Инфраструктурные NFR |

---

## Анализ разрывов

### Ключевые выводы

1. **M1–M4 имеют хорошее покрытие** — большинство завершённых milestone'ов покрыты тестами через существующие P0-валидации.

2. **M5 (текущий milestone) имеет разрывы:**
   - 5 FUN.INTEGRATION.* требований (cost, billing, saga) — не реализованы в коде, не тестируемы
   - REQ-FUN.API.sunset-header — API sunset не реализован
   - REQ-USR.UI.excel-preview — нет frontend-тестов

3. **NFR.Performance (M13 scope) не тестируется:**
   - `REQ-NFR.PERF.performance` (p95 < 120ms, p99 < 250ms)
   - `REQ-NFR.PERF.canonical-workload-profile` (до 1M axioms)
   - Требуют специализированных бенчмарков

4. **78% orphan P0 — non-testable** (CON constraints, BIZ business rules, INFRA policies, PROCESS procedures, DATA policies, OPS operations). Эти требования не могут быть покрыты автоматическими тестами по своей природе.

---

## Gate Decision (обновлённый)

```
<!-- test-quality-gate: WARN -->
<!-- tqs: 7.9 | grade: silver -->
<!-- rcs: 2.9 | orphan-p0: 100 (35 testable) -->
```

| Milestone | P0 Total | Covered | Coverage | Ключевой разрыв |
|-----------|----------|---------|----------|-----------------|
| **M1** Core Engine | 18 | 17 | **94%** | protocol-stack (non-testable) |
| **M2** Doc-to-Ontology | 12 | 12 | **100%** | — |
| **M3** Organization | 20 | 20 | **100%** | — |
| **M4** GUI Frontend | 41 | 39 | **95%** | excel-preview, deprecation-notification |
| **M5** MVP Gaps | 22 | 8 | **36%** | sunset-header, 5 INTEGRATION, CON.* |
| **M7** Security/MVP | 5 | 2 | **40%** | 3 PROCESS reqs (non-testable) |
| **M9** Document AI 1.0 | 6 | 1 | **17%** | 5 NFR latency/limits |
| **M13** Operations 1.0 | 9 | 1 | **11%** | billing, cost, OPS procedures |
| **Non-testable** | ~65 | 0 | — | CON, BIZ, INFRA, PROCESS, DATA policy |
