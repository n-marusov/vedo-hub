# Code Review Summary

**Files Reviewed:** 16 ключевых файлов (из 60+ изменённых в ветке)
**Risk Level:** 🟢 Low

---

## Контекстные гейты

- **Архитектура:** ✅ PASS — Изменения следуют архитектурным границам: фронтенд → API Gateway → auth-service. Нет прямых обращений к Neo4j/PostgreSQL. Нет циклических зависимостей.
- **Правила:** ✅ PASS — Именование соответствует глоссарию (Project, Group, Ontology, Owner, Maintainer). Traceability annotations присутствуют. Структурированное логирование через slog.
- **Дорожная карта:** ✅ PASS — Изменения соответствуют M5 (MVP Scope Gap Closure).

---

## Critical Issues

Нет.

---

## Suggestions

### 1. `visibility` enum: frontend использует lowercase, бэкенд — PascalCase

**Проблема:** На фронтенде `CreateProjectPage.vue` значение visibility по умолчанию — `"private"` (lowercase), а gRPC proto и OpenAPI spec ожидают `"Private"` (PascalCase). Серверная сторона (`server.go`) по умолчанию устанавливает `"Private"`, но если фронтенд пришлёт `"private"`, это будет обработано как `""` (пустое) и заменено на `"Private"` благодаря `normalizeVisibility()`. Однако при явной передаче `"private"` (lowercase) может возникнуть несоответствие.

**Файл:** `apps/services/frontend/src/pages/CreateProjectPage.vue:158`
**Рекомендация:** Унифицировать регистр — либо использовать `"Private"` на фронтенде, либо добавить case-insensitive нормализацию на бэкенде (в `normalizeVisibility`).

```typescript
// Сейчас:
const selectedVisibility = ref("private");

// Рекомендация:
const selectedVisibility = ref("Private");
```

**Severity:** 🟡 Medium

### 2. `showToast` успеха — тип не указан

**Проблема:** Вызов `showToast(t("projects.create_success"))` не указывает тип тоста (по умолчанию `"success"`). В `useToast()` сигнатура `showToast(message: string, type?: ToastType)`. Явное указание типа улучшит читаемость.

**Файл:** `apps/services/frontend/src/pages/CreateProjectPage.vue:287`
**Рекомендация:** Добавить явный тип: `showToast(t("projects.create_success"), "success")`

### 3. `nextTick` после `setTimeout` в тестах

**Проблема:** Тесты `CreateProjectPage.spec.ts` используют `await new Promise((resolve) => setTimeout(resolve, N))` с последующим `await nextTick()` для ожидания асинхронных рендеров. Это B1 анти-паттерн. Предпочтительнее использовать `flushPromises()` или `waitFor()` из `@vue/test-utils`.

**Файл:** `apps/services/frontend/src/__tests__/CreateProjectPage.spec.ts` (строки 130, 206, 223, 253, 267, 285, 299)
**Рекомендация:** Заменить на `await flushPromises()` из `@vue/test-utils` или использовать `waitFor(() => { expect(...) })`.

### 4. TODO маркеры в `org_handler.go`

**Проблема:** В хендлере `HandleForkProject` остались закомментированные TODO с grpc путями.

**Файл:** `apps/services/api-gateway/handlers/org_handler.go:445,449,469`
**Рекомендация:** Удалить или раскомментировать после `buf generate`.

### 5. Сырые ошибки в Toast

**Проблема:** `CreateProjectPage.vue` использует `String(err)` для неизвестных ошибок в Toast. Это может показать пользователю внутренние детали реализации.

**Файл:** `apps/services/frontend/src/pages/CreateProjectPage.vue:290`
**Рекомендация:** Использовать `extractErrorMessage(err, "Unknown error")` для единообразной обработки ошибок.

---

## Questions

Нет.

---

## Positive Notes

1. **Чистая архитектура:** Отличное разделение ответственности — CreateProjectPage.vue не содержит бизнес-логики, только UI и вызов API.
2. **Idempotency-Key:** Корректно реализован на фронтенде через `crypto.randomUUID()` для безопасного retry запросов.
3. **Role validation:** Проверка роли Maintainer+ на уровне `CreateScope` — защита от эскалации привилегий.
4. **Visibility inheritance:** Наследование visibility от родительской группы при пустом значении — правильное поведение.
5. **Compensating cleanup:** При ошибке создания Ontology выполняется cleanup созданного Scope — правильный saga-паттерн.
6. **Traceability:** Все новые артефакты добавлены в `traceability.ttl`.
7. **i18n:** Полный набор `projects.*` ключей в обоих языках (en + ru).

---

```aif-gate-result
{
  "schema_version": 1,
  "gate": "review",
  "status": "pass",
  "blocking": false,
  "blockers": [],
  "affected_files": [
    "apps/services/frontend/src/pages/CreateProjectPage.vue",
    "apps/services/frontend/src/__tests__/CreateProjectPage.spec.ts",
    "apps/services/api-gateway/handlers/org_handler.go",
    "apps/services/auth-service/internal/grpc/server.go",
    "apps/services/auth-service/org/org.go",
    "apps/services/frontend/src/api/org.ts",
    "apps/services/api-gateway/docs/openapi.json",
    "apps/shared/proto/auth/v1/org.proto"
  ],
  "suggested_next": {
    "command": "$aif-commit",
    "reason": "Review passed. 5 minor suggestions, no critical issues."
  }
}
```
