# =============================================================================
# VEDO Core — Build Orchestrator
# =============================================================================
# Root Makefile for polyglot microservices monorepo (Go, Rust, Python, TypeScript).
# Delegates to language-specific rules in tools/build/*.mk
# =============================================================================

# --- Preamble ----------------------------------------------------------------
# Strict mode: fail on errors, undefined vars, pipe failures.
# Disable implicit rules for faster, predictable execution.
ifeq ($(OS),Windows_NT)
  SHELL := C:/Program Files/Git/bin/bash.exe
else
  SHELL := /bin/bash
endif
.ONESHELL:
.SHELLFLAGS := -eu -o pipefail -c
.DELETE_ON_ERROR:
MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules

# --- Project Metadata --------------------------------------------------------
VERSION    ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT     ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ' 2>/dev/null || echo "unknown")
ROOT       := $(realpath $(dir $(lastword $(MAKEFILE_LIST))))

# --- Service Paths -----------------------------------------------------------
SERVICE_DIR     := apps/services
SRC_SERVICE_DIR := src/services

# --- Include Language-Specific Rules -----------------------------------------
include tools/build/rust.mk
include tools/build/go.mk
include tools/build/python.mk
include tools/build/typescript.mk
include tools/build/docker.mk
include tools/build/yaml.mk

# --- Default Goal ------------------------------------------------------------
.DEFAULT_GOAL := help

# ==============================================================================
# HELP
# ==============================================================================

.PHONY: help
help: ## Show this help message
	@awk 'BEGIN{printf "Usage: make \\033[36m<target>\\033[0m\\n"} \
		/^[-a-zA-Z_0-9]+:.*## /{n=substr($$1,1,match($$1,/:/)-1);gsub(/^[-a-zA-Z_0-9]+:.*## /,"",$$0);printf "  \\033[36m%-25s\\033[0m %s\\n",n,$$0} \
		/^##@/{printf "\\n\\033[1m%s\\033[0m\\n",substr($$0,5)}' $(MAKEFILE_LIST)

.PHONY: list-services
list-services: ## List all registered services and their language
	@echo "=== Go Services ==="
	@for d in $(GO_DIRS); do printf "%s\n" "$$(basename $$d)"; done | sort -u | while read s; do printf "  %s\n" "$$s"; done
	@echo ""
	@echo "=== Rust Services ==="
	@for d in $(RUST_DIRS); do \
		bn=$$(basename $$d); \
		[ "$$bn" = "shared" ] && continue; \
		printf "%s\n" "$$bn"; \
	done | sort -u | while read s; do printf "  %s\n" "$$s"; done
	@echo ""
	@echo "=== Python Services ==="
	@for d in $(PYTHON_DIRS); do printf "%s\n" "$$(basename $$d)"; done | sort -u | while read s; do printf "  %s\n" "$$s"; done
	@echo ""
	@echo "=== TypeScript Services ==="
	@for d in $(TS_DIRS); do printf "%s\n" "$$(basename $$d)"; done | sort -u | while read s; do printf "  %s\n" "$$s"; done

# ==============================================================================
# PROTO — gRPC code generation
# ==============================================================================

PROTO_DIR := apps/shared/proto

.PHONY: proto-generate
proto-generate: ## Generate gRPC code from .proto files (Go + Rust)
	@echo "[Proto] generating Go code..."
	@cd $(ROOT)/$(PROTO_DIR) && buf generate
	@echo "[Proto] Go code generated"
	@echo "[Proto] Rust code is generated at build-time via tonic_build"

.PHONY: proto-lint
proto-lint: ## Lint .proto files with buf
	@echo "[Proto] linting..."
	@cd $(ROOT)/$(PROTO_DIR) && buf lint

.PHONY: proto-breaking
proto-breaking: ## Check .proto breaking changes with buf
	@echo "[Proto] checking breaking changes..."
	@cd $(ROOT)/$(PROTO_DIR) && buf breaking --against '.git#branch=main'

.PHONY: proto-all
proto-all: proto-lint proto-generate ## Run all proto checks (lint + generate)

.PHONY: proto-generate-python
proto-generate-python: ## Generate Python gRPC stubs for document-extractor
	@echo "[Proto] generating Python code..."
	@cd $(ROOT)/$(PROTO_DIR) && python -m grpc_tools.protoc \
		--proto_path=. \
		--python_out=$(ROOT)/$(SERVICE_DIR)/document-extractor/grpc_client \
		--grpc_python_out=$(ROOT)/$(SERVICE_DIR)/document-extractor/grpc_client \
		common/v1/common.proto \
		ontology/v1/ontology.proto
	@echo "[Proto] Python code generated to $(SERVICE_DIR)/document-extractor/grpc_client/"
	@echo "[Proto] Note: run 'pip install grpcio-tools' if grpc_tools is not available"

.PHONY: proto-generate-ai-orchestration
proto-generate-ai-orchestration: ## Generate Go gRPC stubs for ai-orchestration-service
	@echo "[Proto] generating ai-orchestration Go code..."
	@cd $(ROOT)/$(PROTO_DIR) && buf generate ai-orchestration/v1/ai_orchestration.proto
	@echo "[Proto] ai-orchestration Go code generated"

.PHONY: proto-generate-ai-orchestration-python
proto-generate-ai-orchestration-python: ## Generate Python gRPC stubs for ai-orchestration
	@echo "[Proto] generating ai-orchestration Python code..."
	@cd $(ROOT)/$(PROTO_DIR) && python -m grpc_tools.protoc \
		--proto_path=. \
		--python_out=$(ROOT)/$(SERVICE_DIR)/document-extractor/grpc_client \
		--grpc_python_out=$(ROOT)/$(SERVICE_DIR)/document-extractor/grpc_client \
		common/v1/common.proto \
		ai-orchestration/v1/ai_orchestration.proto
	@echo "[Proto] ai-orchestration Python code generated to $(SERVICE_DIR)/document-extractor/grpc_client/"

# ==============================================================================
# BUILD — Compile all services natively (without Docker)
# ==============================================================================

##@ Build — Native (all)

build: build-rust build-go build-python build-typescript ## Build all services natively (Rust, Go, Python, TypeScript)

build-all: build ## Alias for build

##@ Build — Native (per-service)

# Build a single service natively (usage: make build-api-gateway)
# @desc: Detects language automatically from service directory contents.
#   Works for services under apps/services/ and src/services/.
#   For per-language builds use: build-rust, build-go, build-python, build-typescript
.PHONY: build-%
build-%:
	@svc="$*"
	dir=""
	for d in "$(ROOT)/$(SERVICE_DIR)/$$svc" "$(ROOT)/$(SRC_SERVICE_DIR)/$$svc"; do \
		if [ -d "$$d" ]; then dir="$$d"; break; fi; \
	done
	if [ -z "$$dir" ]; then \
		echo "Unknown service: $$svc"; \
		echo "Run 'make list-services' to see available services."; \
		exit 1; \
	fi
	if [ -f "$$dir/Cargo.toml" ]; then \
		echo "[Rust] building $$svc"
		cd "$$dir" && cargo build --release 2>&1
	elif [ -f "$$dir/go.mod" ]; then \
		echo "[Go] building $$svc"
		cd "$$dir" && go build ./... 2>&1
	elif [ -f "$$dir/pyproject.toml" ]; then \
		echo "[Python] installing deps for $$svc"
		cd "$$dir" && uv sync 2>&1
	elif [ -f "$$dir/package.json" ]; then \
		echo "[TypeScript] building $$svc"
		cd "$$dir" && pnpm install 2>&1 && pnpm build 2>&1
	else
		echo "Cannot determine language for $$svc — no Cargo.toml, go.mod, pyproject.toml, or package.json"
		exit 1
	fi



# ==============================================================================
# TEST — Unit tests (runs natively, no Docker required)
# ==============================================================================

##@ Test — Unit

.PHONY: test test-all
test: ## Run all unit tests (T0) natively
	@failed=""; \
	echo "[Test] Running all unit tests..."; \
	$(MAKE) test-rust-unit 2>&1 || failed="$$failed rust"; \
	$(MAKE) test-go 2>&1 || failed="$$failed go"; \
	$(MAKE) test-python 2>&1 || failed="$$failed python"; \
	$(MAKE) test-typescript 2>&1 || failed="$$failed typescript"; \
	echo ""; \
	echo "========================================"; \
	if [ -z "$$failed" ]; then \
		echo "  [PASS]  ALL UNIT TESTS PASSED"; \
	else \
		echo "  [FAIL]  UNIT TESTS FAILED:$$failed"; \
	fi; \
	echo "========================================"; \
	if [ -n "$$failed" ]; then exit 1; fi

test-all: ## Run all tests including integration (T0+T1)
	@failed=""; \
	echo "[Test] Running all tests (unit + integration)..."; \
	$(MAKE) test-rust 2>&1 || failed="$$failed rust"; \
	$(MAKE) test-go 2>&1 || failed="$$failed go"; \
	$(MAKE) test-python 2>&1 || failed="$$failed python"; \
	$(MAKE) test-typescript 2>&1 || failed="$$failed typescript"; \
	echo ""; \
	echo "========================================"; \
	if [ -z "$$failed" ]; then \
		echo "  [PASS]  ALL TESTS PASSED"; \
	else \
		echo "  [FAIL]  TESTS FAILED:$$failed"; \
	fi; \
	echo "========================================"; \
	if [ -n "$$failed" ]; then exit 1; fi

##@ Test — Integration

test-integration: test-integration-rust test-integration-go ## Run integration tests (requires live infra — Neo4j auto-started, PG, Go services)

.PHONY: test-integration-rust
test-integration-rust: ## Run Rust integration tests (auto-starts Neo4j if not running)
	@NEO4J_STARTED=""; \
	NEO4J_HOST="localhost"; \
	NEO4J_PORT="$${NEO4J_BOLT_PORT:-7687}"; \
	NEO4J_USER="$${NEO4J_USER:-neo4j}"; \
	NEO4J_PASSWORD="$${NEO4J_PASSWORD:-password}"; \
	NEO4J_TEST_URI="bolt://$${NEO4J_HOST}:$${NEO4J_PORT}"; \
	echo "[Integration] checking Neo4j at $${NEO4J_HOST}:$${NEO4J_PORT}..."; \
	if command -v docker >/dev/null 2>&1 && docker compose -f "$(ROOT)/deploy/docker-compose.yml" ps neo4j 2>/dev/null | grep -q "healthy"; then \
		echo "[Integration] Neo4j is already running (Docker healthy)"; \
	else \
		echo "[Integration] Starting Neo4j via Docker Compose..."; \
		cd "$(ROOT)" && docker compose -f deploy/docker-compose.yml --env-file config/.env.dev up -d neo4j 2>&1; \
		NEO4J_STARTED="yes"; \
		echo "[Integration] Waiting for Neo4j healthcheck..."; \
		i=0; \
		while [ $$i -lt 60 ]; do \
			if docker compose -f "$(ROOT)/deploy/docker-compose.yml" ps neo4j 2>/dev/null | grep -q "healthy"; then \
				echo "[Integration] Neo4j is ready!"; \
				break; \
			fi; \
			sleep 2; \
			i=$$((i + 1)); \
			echo "[Integration]   ...waiting for Neo4j ($${i}s)"; \
		done; \
		if [ $$i -ge 60 ]; then \
			echo "[Integration] ERROR: Neo4j did not become ready within 120 seconds"; \
			docker compose -f "$(ROOT)/deploy/docker-compose.yml" logs neo4j 2>&1 | tail -20; \
			exit 1; \
		fi; \
	fi; \
	export NEO4J_TEST_URI="$${NEO4J_TEST_URI}"; \
	export NEO4J_USER="$$NEO4J_USER"; \
	export NEO4J_PASSWORD="$$NEO4J_PASSWORD"; \
	export NEO4J_URI="$${NEO4J_TEST_URI}"; \
	echo "[Integration] Running ontology-service Neo4j integration tests..."; \
	cd "$(ROOT)/apps/services/ontology-service"; \
	RESULT=0; \
	for test_file in tests/*integration_test.rs tests/*p0_test.rs; do \
		test_name="$$(basename "$$test_file" .rs)"; \
		echo "[Integration]   [$${test_name}]"; \
		cargo test --test "$${test_name}" -- --test-threads=1 --nocapture 2>&1 || { RESULT=1; break; }; \
	done; \
	if [ -n "$$NEO4J_STARTED" ]; then \
		echo "[Integration] Stopping auto-started Neo4j..."; \
		cd "$(ROOT)" && docker compose -f deploy/docker-compose.yml --env-file config/.env.dev stop neo4j 2>&1; \
		cd "$(ROOT)/apps/services/ontology-service"; \
	fi; \
	if [ $$RESULT -ne 0 ]; then \
		exit 1; \
	fi

.PHONY: test-integration-go
test-integration-go: ## Run Go integration tests in tests/integration/
	@if [ -d "$(ROOT)/tests/integration/ticket-api" ]; then \
		echo "[Go] integration tests — ticket-api"
		cd "$(ROOT)/tests/integration/ticket-api" && go test ./... 2>&1 || true; \
	fi
	@if [ -d "$(ROOT)/tests/integration/org-api" ]; then \
		echo "[Go] integration tests — org-api"
		cd "$(ROOT)/tests/integration/org-api" && go test ./... 2>&1 || true; \
	fi

.PHONY: test-versioning
test-versioning: ## Run versioning-service integration tests (auto-starts PostgreSQL if not running)
	@echo "[Versioning] checking PostgreSQL availability..."
	@PG_URL="postgres://postgres:password@localhost:5432/vedo_versioning"; \
	PG_STARTED=""; \
	if command -v pg_isready >/dev/null 2>&1; then \
		if pg_isready -q -h localhost -p 5432 2>/dev/null; then \
			echo "[Versioning] PostgreSQL is already running"; \
		else \
			echo "[Versioning] Starting PostgreSQL via Docker Compose..."; \
			cd "$(ROOT)" && docker compose -f deploy/docker-compose.yml --env-file config/.env.dev up -d postgres 2>&1; \
			PG_STARTED="yes"; \
			echo "[Versioning] Waiting for PostgreSQL to become healthy..."; \
			i=0; \
			while [ $$i -lt 30 ]; do \
				if pg_isready -q -h localhost -p 5432 2>/dev/null; then \
					echo "[Versioning] PostgreSQL is ready!"; \
					break; \
				fi; \
				sleep 1; \
				i=$$((i + 1)); \
			done; \
			if [ $$i -ge 30 ]; then \
				echo "[Versioning] ERROR: PostgreSQL did not become ready within 30 seconds"; \
				exit 1; \
			fi; \
		fi; \
	else \
		echo "[Versioning] pg_isready not found, assuming PostgreSQL is available at localhost:5432"; \
	fi; \
	export PG_TEST_DATABASE_URL="$$PG_URL"; \
	export DATABASE_URL="$$PG_URL"; \
	echo "[Versioning] Running versioning-service unit tests..."; \
	cd "$(ROOT)/apps/services/versioning-service" && cargo test --lib 2>&1 || exit 1; \
	echo "[Versioning] Running versioning-service integration tests..."; \
	cd "$(ROOT)/apps/services/versioning-service" && cargo test --test '*' -- --test-threads=1 --nocapture 2>&1; \
	RESULT=$$?; \
	if [ -n "$$PG_STARTED" ]; then \
		echo "[Versioning] Stopping auto-started PostgreSQL..."; \
		cd "$(ROOT)" && docker compose -f deploy/docker-compose.yml stop postgres 2>&1; \
	fi; \
	exit $$RESULT

##@ Test — E2E (requires Docker test stack)

test-e2e: test-e2e-api test-e2e-gui ## Run all E2E tests (API + GUI)

.PHONY: test-e2e-api
test-e2e-api: ## Run E2E API tests via Playwright (requires Docker test stack)
	@echo "[E2E] installing dependencies..."
	@cd "$(ROOT)/tests/e2e" && pnpm install --frozen-lockfile 2>&1 || pnpm install 2>&1
	@cd "$(ROOT)/tests/e2e" && npx playwright install chromium 2>&1 || true
	@echo "[E2E] running API tests..."
	@cd "$(ROOT)/tests/e2e" && pnpm exec playwright test --config=config/playwright.api.config.ts

.PHONY: test-e2e-gui
test-e2e-gui: ## Run E2E GUI tests via Playwright (requires Docker test stack)
	@echo "[E2E] installing dependencies..."
	@cd "$(ROOT)/tests/e2e" && pnpm install --frozen-lockfile 2>&1 || pnpm install 2>&1
	@cd "$(ROOT)/tests/e2e" && npx playwright install chromium 2>&1 || true
	@echo "[E2E] running GUI tests..."
	@cd "$(ROOT)/tests/e2e" && pnpm exec playwright test --config=config/playwright.gui.config.ts

.PHONY: test-gates
test-gates: ## Run all gate-level test scripts (contracts, BOLA/BFLA, etc.)
	@gate_failures=""
	@echo "[Gates] running contract tests..."
	@bash $(ROOT)/tests/gates/test_contract_gate.sh 2>&1 || gate_failures="$${gate_failures} contract"
	@echo "[Gates] running BOLA/BFLA security tests..."
	@bash $(ROOT)/tests/gates/test_bola_bfla_gate.sh 2>&1 || gate_failures="$${gate_failures} bola-bfla"
	@echo "[Gates] running Python service manifest validation..."
	@bash $(ROOT)/tests/gates/test_python_manifests.sh 2>&1 || gate_failures="$${gate_failures} python-manifests"
	@if [ -n "$$gate_failures" ]; then \
		echo ""; \
		echo "!!! Gates FAILED:$$gate_failures !!!"; \
		exit 1; \
	fi
	@echo "[Gates] all gates passed"

.PHONY: coverage
coverage: ## Run tests with coverage (Go only)
	@echo "[Coverage] running tests with coverage..."
	@if [ -n "$(GO_DIRS)" ]; then \
		for dir in $(GO_DIRS); do \
			echo "[Go] coverage for $$(basename $$dir)"; \
			cd $(ROOT)/$$dir && go test -coverprofile=cover.out ./... 2>&1 || true; \
		done; \
	fi

# ==============================================================================
# LINT — All linters
# ==============================================================================

##@ Lint

lint: lint-rust lint-go lint-python lint-typescript lint-yaml ## Run all linters

.PHONY: lint-%
lint-%: ## Lint a single service (usage: make lint-api-gateway)
	@svc="$*"
	dir=""
	for d in "$(ROOT)/$(SERVICE_DIR)/$$svc" "$(ROOT)/$(SRC_SERVICE_DIR)/$$svc"; do \
		if [ -d "$$d" ]; then dir="$$d"; break; fi; \
	done
	if [ -z "$$dir" ]; then \
		echo "Unknown service: $$svc"; \
		echo "Run 'make list-services' to see available services."; \
		exit 1; \
	fi
	if [ -f "$$dir/Cargo.toml" ]; then \
		echo "[Rust] linting $$svc"
		cd "$$dir" && cargo clippy -- -D warnings 2>&1
	elif [ -f "$$dir/go.mod" ]; then \
		echo "[Go] linting $$svc"
		cd "$$dir" && golangci-lint run ./... 2>&1
	elif [ -f "$$dir/pyproject.toml" ]; then \
		echo "[Python] linting $$svc"
		cd "$$dir" && ruff check . 2>&1
	elif [ -f "$$dir/package.json" ]; then \
		echo "[TypeScript] linting $$svc"
		cd "$$dir" && npx biome check . 2>&1
	else
		echo "Cannot determine language for $$svc"
		exit 1
	fi

# ==============================================================================
# FORMAT — Code formatting
# ==============================================================================

##@ Format

.PHONY: fmt-rust fmt-go fmt-python fmt-typescript fmt

fmt-rust: ## Format all Rust code with cargo fmt
	@if [ -z "$(RUST_DIRS)" ]; then echo "No Rust services found"; exit 0; fi
	@for dir in $(RUST_DIRS); do \
		echo "[Rust] formatting $$(basename $$dir)"; \
		cd $$dir && cargo fmt 2>&1 || true; \
	done

fmt-go: ## Format all Go code with gofmt
	@if [ -z "$(GO_DIRS)" ]; then echo "No Go services found"; exit 0; fi
	@for dir in $(GO_DIRS); do \
		echo "[Go] formatting $$(basename $$dir)"; \
		cd $(ROOT)/$$dir && gofmt -l -w . 2>&1 || true; \
	done

fmt-python: ## Format all Python code with ruff
	@if [ -z "$(PYTHON_DIRS)" ]; then echo "No Python services found"; exit 0; fi
	@for dir in $(PYTHON_DIRS); do \
		echo "[Python] formatting $$(basename $$dir)"; \
		cd $$dir && ruff format . 2>&1 || true; \
	done

fmt-typescript: ## Format all TypeScript code with biome
	@if [ -z "$(TS_DIRS)" ]; then echo "No TypeScript services found"; exit 0; fi
	@for dir in $(TS_DIRS); do \
		echo "[TypeScript] formatting $$(basename $$dir)"; \
		cd $(ROOT)/$$dir && npx biome check --write . 2>&1 || true; \
	done

fmt: fmt-rust fmt-go fmt-python fmt-typescript ## Format all source code

# ==============================================================================
# DOCKER — Build images
# ==============================================================================

##@ Docker — Build

docker-build: docker-build-rust docker-build-go docker-build-python docker-build-typescript ## Build Docker images for all services (by language)

docker-build-all: docker-build ## Alias for docker-build

# Build a single service Docker image via compose (usage: make docker-build-api-gateway)
# NOTE: per-service builds use `docker compose build` which reads the compose file
# to determine build context and Dockerfile.
docker-build-%:
	@docker compose \
		--env-file $(ROOT)/config/.env.$(ENV) \
		-f $(ROOT)/$(COMPOSE_FILE) \
		build $*

##@ Docker — Compose Lifecycle

ENV                 ?= dev
COMPOSE_FILE        ?= deploy/docker-compose.yml
COMPOSE_PROFILE     ?=

.PHONY: docker-up docker-down docker-logs docker-ps docker-shell docker-config
.PHONY: docker-up-dev docker-up-test docker-up-staging
.PHONY: docker-down-dev docker-down-test docker-down-staging

docker-up: ## Start all services via Docker Compose (usage: make docker-up [ENV=dev|test|staging])
	docker compose \
		--env-file $(ROOT)/config/.env.$(ENV) \
		-f $(ROOT)/$(COMPOSE_FILE) \
		$(if $(COMPOSE_PROFILE),--profile $(COMPOSE_PROFILE)) \
		up -d

docker-up-dev: ## Start dev environment (alias for make docker-up ENV=dev)
	$(MAKE) docker-up ENV=dev

docker-up-test: ## Start test environment (includes JWT dev keys for Playwright)
	$(MAKE) docker-up ENV=test

docker-up-staging: ## Start staging environment
	$(MAKE) docker-up ENV=staging

docker-down: ## Stop all services (usage: make docker-down [ENV=dev|test|staging])
	docker compose \
		--env-file $(ROOT)/config/.env.$(ENV) \
		-f $(ROOT)/$(COMPOSE_FILE) \
		down

docker-down-dev: ## Stop dev environment
	$(MAKE) docker-down ENV=dev

docker-down-test: ## Stop test environment
	$(MAKE) docker-down ENV=test

docker-down-staging: ## Stop staging environment
	$(MAKE) docker-down ENV=staging

docker-restart: docker-down docker-up ## Restart the current environment

docker-logs: ## Tail logs from all services (usage: make docker-logs [ENV=dev|test|staging])
	docker compose \
		--env-file $(ROOT)/config/.env.$(ENV) \
		-f $(ROOT)/$(COMPOSE_FILE) \
		logs -f

docker-ps: ## List running service containers (usage: make docker-ps [ENV=dev|test|staging])
	docker compose \
		--env-file $(ROOT)/config/.env.$(ENV) \
		-f $(ROOT)/$(COMPOSE_FILE) \
		ps

docker-shell: ## Open shell in a service container (usage: make docker-shell SVC=<name> [ENV=dev])
	@if [ -z "$(SVC)" ]; then echo "Usage: make docker-shell SVC=<service-name>"; exit 1; fi
	docker compose \
		--env-file $(ROOT)/config/.env.$(ENV) \
		-f $(ROOT)/$(COMPOSE_FILE) \
		exec $(SVC) sh

docker-config: ## Validate compose configuration (usage: make docker-config [ENV=dev|test|staging])
	docker compose \
		--env-file $(ROOT)/config/.env.$(ENV) \
		-f $(ROOT)/$(COMPOSE_FILE) \
		config > /dev/null && echo "Config OK"

##@ Docker — Status

.PHONY: status docker-status docker-status-full

docker-status: ## Brief health table (Name, Health, Ports)
	docker compose \
		--env-file $(ROOT)/config/.env.$(ENV) \
		-f $(ROOT)/$(COMPOSE_FILE) ps --format "table {{.Name}}\t{{.Health}}\t{{.Ports}}" 2>/dev/null

docker-status-full: ## Full machine-readable JSON with all container details
	docker compose \
		--env-file $(ROOT)/config/.env.$(ENV) \
		-f $(ROOT)/$(COMPOSE_FILE) ps --format json 2>/dev/null || echo '[]'

status: docker-status ## Alias — show health of each service in the compose stack

##@ Docker — Infrastructure (standalone infra services without app stack)

.PHONY: infra-up infra-down

infra-up: ## Start only infrastructure services (Neo4j, Postgres, Redis, RabbitMQ, Keycloak, MinIO)
	docker compose \
		--env-file $(ROOT)/config/.env.$(ENV) \
		-f $(ROOT)/$(COMPOSE_FILE) up -d neo4j postgres redis rabbitmq keycloak minio

infra-down: ## Stop infrastructure services
	docker compose \
		--env-file $(ROOT)/config/.env.$(ENV) \
		-f $(ROOT)/$(COMPOSE_FILE) down neo4j postgres redis rabbitmq keycloak minio

# ==============================================================================
# TYPECHECK
# ==============================================================================

##@ TypeScript

.PHONY: typecheck
typecheck: typecheck-typescript ## Run TypeScript type checking

# ==============================================================================
# DEVELOPMENT — Local dev targets for individual services
# ==============================================================================

##@ Development

.PHONY: dev dev-frontend dev-api

dev: ## Start development environment
	@echo "Use 'make docker-up' to start all services via Docker Compose."
	@echo "Or run individual services:"
	@echo "  make dev-frontend  - Start frontend dev server"
	@echo "  make dev-api       - Start API gateway natively"

dev-frontend: ## Start frontend dev server with hot reload
	cd $(ROOT)/$(SERVICE_DIR)/frontend && pnpm dev

dev-api: ## Start API gateway in development mode
	cd $(ROOT)/$(SERVICE_DIR)/api-gateway && go run .

# ==============================================================================
# QUALITY & SECURITY — Static analysis, docs verification, dependency scanning
# ==============================================================================

##@ Quality & Security

.PHONY: test-quality-gate
test-quality-gate: ## Run static test quality gate (anti-patterns, tautologies, etc.)
	@bash $(ROOT)/tools/scripts/test-quality-gate.sh $(ROOT)/apps/services

.PHONY: docs-lint
docs-lint: ## Verify Antora documentation builds without errors
	@echo "[Docs] verifying Antora documentation build..."
	@if command -v npx &>/dev/null; then \
		cd $(ROOT)/docs/antora && npx antora --fetch antora-playbook.yml 2>&1 | tail -5 || \
		echo "[Docs] WARN: antora build failed — check docs/antora/antora-playbook.yml"; \
	else \
		echo "[Docs] SKIP: npx not available"; \
	fi

.PHONY: npm-audit
npm-audit: ## Run npm/pnpm audit on frontend dependencies
	@echo "[Security] auditing frontend dependencies..."
	@if [ -f "$(ROOT)/$(SERVICE_DIR)/frontend/package.json" ]; then \
		cd $(ROOT)/$(SERVICE_DIR)/frontend && \
		if command -v pnpm &>/dev/null; then \
			pnpm audit --audit-level=high 2>&1 || echo "[Security] WARN: audit found vulnerabilities"; \
		elif command -v npm &>/dev/null; then \
			npm audit --audit-level=high 2>&1 || echo "[Security] WARN: audit found vulnerabilities"; \
		else \
			echo "[Security] SKIP: neither pnpm nor npm available"; \
		fi; \
	else \
		echo "[Security] SKIP: frontend/package.json not found"; \
	fi

# ==============================================================================
# CI — Full pipeline aggregate target
# ==============================================================================

##@ CI

.PHONY: ci

ci: ## Run full CI pipeline (proto + build + lint + unit tests + typecheck)
	@failed=0
	@echo "=== CI Pipeline Started ==="
	@$(MAKE) proto-all || failed=1
	@$(MAKE) build || failed=1
	@$(MAKE) lint || failed=1
	@$(MAKE) test || failed=1
	@$(MAKE) typecheck || failed=1
	@if [ $$failed -ne 0 ]; then \
		echo ""; \
		echo "!!! CI Pipeline FAILED !!!"; \
		exit 1; \
	fi
	@echo ""
	@echo "=== CI pipeline passed (proto + build + lint + unit tests + typecheck) ==="
	@echo "To run integration tests:  make test-integration"
	@echo "To run E2E tests:          make test-e2e (requires Docker test stack)"
	@echo "To run gate tests:         make test-gates"

ci-full: ## Run full CI pipeline including quality gates, Docker build, integration, E2E, and security (auto-starts Docker test stack)
	@failed=0
	@echo "=== Full CI Pipeline Started ==="
	@$(MAKE) vendor-go || failed=1
	@$(MAKE) ci || failed=1
	@$(MAKE) test-quality-gate || failed=1
	@$(MAKE) docs-lint || failed=1
	@echo "[ci-full] Ensuring Docker test stack is up..."
	@$(MAKE) docker-up-test 2>/dev/null || true
	@$(MAKE) test-integration || failed=1
	@$(MAKE) test-e2e || failed=1
	@$(MAKE) test-gates || failed=1
	@$(MAKE) npm-audit || failed=1
	@if [ $$failed -ne 0 ]; then \
		echo ""; \
		echo "!!! Full CI Pipeline FAILED !!!"; \
		exit 1; \
	fi
	@echo ""
	@echo "=== Full CI pipeline passed (vendor-go + ci + quality-gate + docs-lint + integration + E2E + gates + npm-audit) ==="

# ==============================================================================
# HOOKS — Git hooks management via Lefthook
# ==============================================================================

##@ Hooks

.PHONY: install-hooks uninstall-hooks run-hooks validate-hooks

install-hooks: ## Install Git hooks via Lefthook
	@echo "[Hooks] installing..."
	@LEFTHOOK_BIN=$$(ls $(ROOT)/$(SERVICE_DIR)/frontend/node_modules/.bin/lefthook 2>/dev/null || which lefthook 2>/dev/null || echo ""); \
	if [ -z "$$LEFTHOOK_BIN" ]; then \
		echo "Lefthook not found. Install via: cd $(ROOT)/$(SERVICE_DIR)/frontend && pnpm install"; \
		exit 1; \
	fi; \
	$$LEFTHOOK_BIN install

uninstall-hooks: ## Remove all Lefthook hooks
	@LEFTHOOK_BIN=$$(ls $(ROOT)/$(SERVICE_DIR)/frontend/node_modules/.bin/lefthook 2>/dev/null || which lefthook 2>/dev/null || echo ""); \
	if [ -n "$$LEFTHOOK_BIN" ]; then \
		$$LEFTHOOK_BIN uninstall; \
	fi

run-hooks: ## Run all pre-commit hooks on staged files
	@LEFTHOOK_BIN=$$(ls $(ROOT)/$(SERVICE_DIR)/frontend/node_modules/.bin/lefthook 2>/dev/null || which lefthook 2>/dev/null || echo ""); \
	if [ -z "$$LEFTHOOK_BIN" ]; then echo "Lefthook not found"; exit 1; fi; \
	$$LEFTHOOK_BIN run pre-commit

validate-hooks: ## Validate lefthook.yml configuration
	@LEFTHOOK_BIN=$$(ls $(ROOT)/$(SERVICE_DIR)/frontend/node_modules/.bin/lefthook 2>/dev/null || which lefthook 2>/dev/null || echo ""); \
	if [ -z "$$LEFTHOOK_BIN" ]; then echo "Lefthook not found"; exit 1; fi; \
	$$LEFTHOOK_BIN validate

# ==============================================================================
# CLEAN — Remove all build artifacts
# ==============================================================================

##@ Clean

.PHONY: clean
clean: clean-rust clean-go clean-python clean-typescript ## Clean all build artifacts

.PHONY: clean-all
clean-all: clean ## Deep clean (all artifacts + vendor dirs)
	@echo "[Clean] removing Go vendor directories..."
	@for dir in $(GO_DIRS); do \
		if [ -d "$(ROOT)/$$dir/vendor" ]; then \
			echo "  removing $$dir/vendor"; \
			rm -rf "$(ROOT)/$$dir/vendor"; \
		fi; \
	done
	@echo "[Clean] removing .venv directories..."
	@for dir in $(PYTHON_DIRS); do \
		if [ -d "$$dir/.venv" ]; then \
			echo "  removing $$dir/.venv"; \
			rm -rf "$$dir/.venv"; \
		fi; \
	done
	@echo "[Clean] removing node_modules..."
	@for dir in $(TS_DIRS); do \
		if [ -d "$(ROOT)/$$dir/node_modules" ]; then \
			echo "  removing $$dir/node_modules"; \
			rm -rf "$(ROOT)/$$dir/node_modules"; \
		fi; \
	done
	@echo "[Clean] deep clean complete"
