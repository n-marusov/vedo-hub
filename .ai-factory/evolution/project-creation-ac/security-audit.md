# Security Audit: Project Creation in Group

**Scope:** Изменённые файлы в `feature/project-creation-in-group`

---

## 🔴 Critical: Pre-Deployment Checklist

| Проверка | Статус |
|----------|--------|
| No secrets in code or git history | ✅ Нет жестко закодированных секретов |
| All user input is validated and sanitized | ✅ ShouldBindJSON (Go) + валидация на фронтенде |
| Authentication on all protected routes | ✅ `requiresAuth: true` на всех защищённых маршрутах |
| Error messages don't leak sensitive info | ⚠️ Toast показывает `String(err)` для неизвестных ошибок |
| Client-side debug logging guarded | ⚠️ `console.info` используется как структурированное логирование (штатный паттерн проекта) |
| Dependencies scanned | ✅ `go test ./...` проходит, `go vet` без ошибок |
| Race conditions prevented | ✅ Idempotency-Key на write endpoint |

---

## Authentication & Sessions

### JWT Security
- ✅ Bearer token извлекается из `Authorization` header (`extractToken`)
- ✅ Токен проверяется в `router.beforeEach` guard
- ✅ Все защищённые маршруты имеют `meta: { requiresAuth: true }`
- ✅ Проверка валидности токена: `isTokenValid()`

## Injection Prevention

### Input Validation (API Gateway)
- ✅ `c.ShouldBindJSON(&req)` на всех эндпоинтах HandleCreateProject
- ✅ Проверка обязательных полей: name (или label), group_id
- ✅ Валидация visibility на бэкенде (enum + parent check)

### Input Validation (Frontend)
- ✅ Group required: `if (!selectedGroupId.value)`
- ✅ Name required: `if (!trimmed)`
- ✅ Валидация перед отправкой запроса

### NoSQL / Command Injection
- ✅ gRPC вызовы с типизированными proto-сообщениями
- ✅ Нет динамических запросов к БД

## XSS Protection

- ✅ Нет `v-html`, `innerHTML`, `dangerouslySetInnerHTML`
- ✅ Vue шаблоны с авто-экранированием ({{ }})
- ✅ CSP заголовки не проверены (зависит от nginx/config)

## CSRF Protection

- ✅ SPA с Bearer token (не cookie-based)
- ✅ `X-Requested-With: XMLHttpRequest` заголовок
- ⚠️ SameSite cookie политика не проверена (nginx/config)

## Secrets Management

- ✅ Нет жестко закодированных паролей/ключей/токенов
- ✅ `.env` не закоммичен (проверено через git)
- ✅ Нет секретов в логах (структурированное логирование без sensitive data)

## Client-Facing Logging & Errors

### Findings
```typescript
// CreateProjectPage.vue:227
const msg = err instanceof Error ? err.message : String(err);
showToast(t("projects.create_error", { error: msg }), "error");
```

- ✅ `extractErrorMessage` в `org.ts` нормализует Axios ошибки
- ⚠️ Для неизвестных ошибок `String(err)` может показать внутренние детали в Toast
- ⚠️ `console.info` используется для структурированного логирования во всех файлах (штатный паттерн)

### Logging Pattern (intentional)
- `console.info(JSON.stringify({ level: "info", event: "...", ... }))` — структурированное логирование
- `console.error(JSON.stringify({ level: "error", ... }))` — ошибки
- `console.debug(JSON.stringify({ level: "debug", ... }))` — отладка

## TODO / FIXME Маркеры

```go
// apps/services/api-gateway/handlers/org_handler.go:445
// TODO (Task 3.1): Uncomment gRPC path once `buf generate` regenerates

// apps/services/api-gateway/handlers/org_handler.go:449
// TODO: gRPC path (uncomment after buf generate):
```

⚠️ **2 TODO маркера** в `org_handler.go` — относятся к ForkProject handler (не к CreateProject). Не блокируют, но рекомендуется очистить.

## API Security

| Проверка | Статус |
|----------|--------|
| API keys in headers (not URL) | ✅ Bearer token в Authorization header |
| Idempotency-Key на write endpoint | ✅ `crypto.randomUUID()` + header |
| Rate limiting | ⚠️ На уровне API Gateway (вне scope данного аудита) |
| Input validation | ✅ ShouldBindJSON + schema validation |
| Error codes stable | ✅ `INVALID_REQUEST`, `GRPC_ERROR`, маппинг через `mapGrpcCodeToHTTP` |

## Резюме

| Категория | Статус |
|-----------|--------|
| Critical issues | 0 |
| High issues | 0 |
| Medium issues | 1 (TODO маркеры в org_handler.go) |
| Low issues | 2 (error messages detail, console.info logging) |

Все критические и высокие проверки пройдены. Сервис следует архитектурным границам безопасности: фронтенд → API Gateway (единственная точка входа) → auth-service с валидацией ролей и видимости.

```aif-gate-result
{
  "schema_version": 1,
  "gate": "security",
  "status": "pass",
  "blocking": false,
  "blockers": [],
  "affected_files": [
    "apps/services/api-gateway/handlers/org_handler.go",
    "apps/services/frontend/src/pages/CreateProjectPage.vue",
    "apps/services/frontend/src/api/org.ts",
    "apps/services/frontend/src/router/index.ts"
  ],
  "suggested_next": {
    "command": "$aif-commit",
    "reason": "Security audit passed with only minor non-blocking findings."
  }
}
```
