# Implementation Plan: Migration Integration Tests — Graceful Guard Pattern

Branch: feature/test-improvements
Created: 2026-07-23

## Settings
- Testing: yes (the plan IS about test infrastructure changes)
- Logging: verbose
- Docs: no (internal test infrastructure, no user docs needed)

## Roadmap Linkage
Milestone: M7 — MVP Acceptance, Security & Demo Readiness
Rationale: Стабильный `make test` — базовое требование acceptance. Падение интеграционных тестов без БД блокирует разработчикам быстрый фидбэк.

## Research Context
Source: Исследование от 2026-07-23 (см. историю — explore mode по `make test`)

Goal: Мигрировать интеграционные тесты с panic на graceful guard, разделить unit/integration в Makefile
Constraints:
- TQS B4: нельзя молча завершать тест без проверки
- Минимальные изменения — не переписывать существующие тесты
- Сохранить возможность запуска интеграционных тестов с БД отдельно
Decisions:
- `eprintln + return` вместо `panic!` — TQS B4 не нарушен (stderr ≠ silent)
- `make test` = только unit (T0), `make test-integration` = с БД
- Docker Compose profile "test" для инфраструктуры

## Commit Plan
- **Commit 1** (tasks 1-2): "fix(versioning-service): replace panic with graceful skip in integration tests"
- **Commit 2** (tasks 3-4): "fix(ontology-service): replace panic with graceful skip in integration tests"
- **Commit 3** (task 5): "build: add test-unit and test-integration targets to Makefile"
- **Commit 4** (task 6): "chore: add test DB services to Docker Compose"
- **Commit 5** (task 7): "docs: update run_integration_rust.sh and README"

## Acceptance Criteria
- [ ] All tests pass: `make test` (unit only, без инфраструктуры)
- [ ] `make test-integration` запускает тесты с БД
- [ ] TQS B4 не нарушен — все guard выводят сообщение в stderr
- [ ] ontology-service: `skip_if_no_neo4j()` — `eprintln + return`, не `panic!`
- [ ] versioning-service: `skip_if_no_pg()` — `eprintln + return`, не `panic!`

## Tasks

### Phase 1: versioning-service — graceful skip вместо panic
- [x] **Task 1: Обновить `skip_if_no_pg()` в versioning-service** (depends on: none)
  - Файл: `src/services/versioning-service/tests/common/mod.rs`
  - Заменить `panic!(...)` на `eprintln!(...)` + `return`
  - Сигнатура функции остаётся `pub fn skip_if_no_pg()`
  - Сообщение: `eprintln!("⚠️  Skipping PostgreSQL integration test. Set PG_TEST_DATABASE_URL to run.")`
  - Убедиться, что тест не делает `panic` — просто возвращается рано
  - LOGGING: INFO-level message to stderr, not stdout

- [x] **Task 2: Обновить integration-тесты versioning-service** (depends on: 1)
  - Файлы: `src/services/versioning-service/tests/branch_integration_test.rs` и остальные 6 test файлов
  - Для асинхронных тестов (`#[tokio::test]`): проверить, что `skip_if_no_pg()` вызывается первой строкой и тест корректно завершается после return
  - Все 4 теста в branch_integration_test.rs, плюс остальные 6 test файлов
  - BDD naming: `skip_if_no_pg` → test returns early without panic
  - **Валидация:** `cargo test -p versioning-service --test '*'` — exit 0 без PG

### Phase 2: ontology-service — graceful skip вместо panic
- [x] **Task 3: Обновить `skip_if_no_neo4j()` в ontology-service** (depends on: none)
  - Файл: `src/services/ontology-service/tests/common/mod.rs`
  - Заменить `panic!(...)` на `eprintln!(...)` + ранний return
  - Сигнатура: `pub fn skip_if_no_neo4j()`
  - Сообщение: `eprintln!("⚠️  Skipping Neo4j integration test. Set NEO4J_TEST_URI to run.")`
  - LOGGING: INFO-level to stderr

- [x] **Task 4: Обновить integration-тесты ontology-service** (depends on: 3)
  - Файлы: все файлы в `src/services/ontology-service/tests/*.rs` (9 файлов)
  - Для тестов, которые используют `skip_if_no_neo4j()` → проверить ранний return без panic
  - Для тестов без внешних зависимостей (org_resolvers_test, route_registration_test) — оставить как есть
  - **Валидация:** `cargo test -p ontology-service --test '*'` — exit 0 без Neo4j

### Phase 3: Makefile — разделение unit/integration
- [x] **Task 5: Добавить `test-rust-unit` и `test-integration` в Makefile** (depends on: 2, 4)
  - Файл: `src/build/rust.mk` и `src/Makefile`
  - Добавить `test-rust-unit`:
    ```makefile
    .PHONY: test-rust-unit
    test-rust-unit:
        @if [ -z "$(RUST_DIRS)" ]; then echo "No Rust services found"; exit 0; fi
        @for dir in $(RUST_DIRS); do \
            echo "[Rust] unit testing $$(basename $$dir)"; \
            cd $$dir && cargo test --lib 2>&1 || ...; \
        done
    ```
  - Изменить `test` в Makefile:
    ```makefile
    test: test-rust-unit test-go test-python test-typescript
    ```
  - Добавить `test-all` (старое поведение: запускает всё)
    ```makefile
    test-all: test-rust test-go test-python test-typescript
    ```
  - **Валидация:** `make test` — exit 0 без БД

### Phase 4: Docker Compose — test profile (опционально, если нужна БД для локального запуска)
- [ ] **Task 6: Добавить test-зависимости в Docker Compose** (depends on: none, optional)
  - Файл: `deploy/docker-compose.yml`
  - Добавить сервисы под `profiles: ["test"]`:
    - `postgres-test` (postgres:16-alpine, порт 5433, БД vedo_test)
    - `neo4j-test` (neo4j:5, порт 7688)
  - Убедиться, что сервисы не запускаются в `docker compose up` без `--profile test`
  - **Валидация:** `docker compose --profile test up -d postgres-test neo4j-test` — контейнеры стартуют

### Phase 5: Документация
- [ ] **Task 7: Обновить документацию тестов** (depends on: 5)
  - Файл: `tests/run_integration_rust.sh` — исправить комментарий «normally SKIPPED» → «normally gracefully skipped with eprintln»
  - Обновить `deploy/README.md` или `src/docs/antora/developer-guide/` — секция про запуск тестов:
    - `make test` → unit-тесты (без инфры)
    - `make test-all` → все тесты (требует инфру)
    - `PG_TEST_DATABASE_URL=... NEO4J_TEST_URI=... make test-integration`
  - **Валидация:** `grep -r "normally SKIPPED" tests/run_integration_rust.sh` — больше не показывает устаревший текст

### Phase 6: Traceability
- [ ] **Task 8: Обновить traceability.ttl** (depends on: 5)
  - Файл: `.ai-factory/traceability/traceability.ttl`
  - Проверить, нужно ли обновлять записи для изменённых `tests/common/mod.rs` файлов
  - Если изменённые файлы уже покрыты директорией service → skip
  - Если нет → добавить `vdo:TestSuite` / `vdo:TestContract`
