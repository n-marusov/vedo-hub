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

# --- ANSI Color Codes (CI-safe) ----------------------------------------------
# Disabled in dumb terminals (CI pipelines, non-TTY).
# Usage: printf "$(C_GREEN)[PASS]$(C_RESET) message\n"
ifeq ($(TERM),dumb)
  C_RED   :=
  C_GREEN :=
  C_YELLOW:=
  C_CYAN  :=
  C_BOLD  :=
  C_RESET :=
else
  C_RED   := \033[31m
  C_GREEN := \033[32m
  C_YELLOW:= \033[33m
  C_CYAN  := \033[36m
  C_BOLD  := \033[1m
  C_RESET := \033[0m
endif

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

##@ Test — Unit

.PHONY: test-unit-fast test-unit-full test-fast test-full

test-unit-fast: ## Unit tests — fail-fast (unit only, stops at first failure)
	@printf "$(C_CYAN)[Unit]$(C_RESET) fail-fast mode\n"
	@$(MAKE) test-rust-fast-unit
	@$(MAKE) test-go-fast
	@$(MAKE) test-python-fast
	@$(MAKE) test-typescript-fast
	@printf "$(C_GREEN)[PASS]$(C_RESET) ALL UNIT TESTS PASSED\n"

test-unit-full: ## Unit tests — full statistics (collect all failures)
	@failed=0; \
	printf "$(C_CYAN)[Unit]$(C_RESET) full statistics mode\n"; \
	$(MAKE) test-rust-full-unit 2>&1 || failed=1; \
	$(MAKE) test-go-full 2>&1 || failed=1; \
	$(MAKE) test-python-full 2>&1 || failed=1; \
	$(MAKE) test-typescript-full 2>&1 || failed=1; \
	if [ $$failed -ne 0 ]; then \
		printf "$(C_RED)[FAIL]$(C_RESET) SOME UNIT TESTS FAILED\n"; \
		exit 1; \
	fi; \
	printf "$(C_GREEN)[PASS]$(C_RESET) ALL UNIT TESTS PASSED\n"

test-fast: ## Run ALL tests — fail-fast (unit + integration + e2e + gates)
	@printf "$(C_CYAN)=== test-fast ===$(C_RESET)\n"
	@$(MAKE) test-unit-fast
	@$(MAKE) test-integration-fast
	@$(MAKE) test-e2e-fast
	@$(MAKE) test-gates-fast
	@printf "$(C_GREEN)[PASS]$(C_RESET) test-fast complete\n"

test-full: ## Run ALL tests — full statistics (collect all failures)
	@failed=0; \
	printf "$(C_CYAN)=== test-full ===$(C_RESET)\n"; \
	$(MAKE) test-unit-full 2>&1 || failed=1; \
	$(MAKE) test-integration-full 2>&1 || failed=1; \
	$(MAKE) test-e2e-full 2>&1 || failed=1; \
	$(MAKE) test-gates-full 2>&1 || failed=1; \
	if [ $$failed -ne 0 ]; then \
		printf "$(C_RED)[FAIL]$(C_RESET) SOME TESTS FAILED\n"; \
		exit 1; \
	fi; \
	printf "$(C_GREEN)[PASS]$(C_RESET) ALL TESTS PASSED\n"

##@ Test — Integration

##@ Test — Integration

.PHONY: test-integration-fast test-integration-full
test-integration-fast: ## Integration tests — fail-fast (auto-starts Neo4j + PostgreSQL via docker-compose.test.yml)
	@printf "$(C_CYAN)[Integration]$(C_RESET) fail-fast mode\n"
	@printf "$(C_CYAN)[Integration]$(C_RESET) checking Neo4j...\n"
	@rm -f /tmp/vedo-test-neo4j /tmp/vedo-test-pg
	@if command -v docker >/dev/null 2>&1 && docker compose -f "$(ROOT)/deploy/docker-compose.test.yml" --env-file config/.env.test ps neo4j 2>/dev/null | grep -q "healthy"; then \
		printf "$(C_GREEN)[Integration]$(C_RESET) Neo4j is already running\n"; \
	else \
		printf "$(C_YELLOW)[Integration]$(C_RESET) Starting Neo4j...\n"; \
		cd "$(ROOT)" && docker compose -f deploy/docker-compose.test.yml --env-file config/.env.test up -d neo4j 2>&1; \
		printf "$(C_YELLOW)[Integration]$(C_RESET) Waiting for Neo4j healthcheck...\n"; \
		i=0; \
		while [ $$i -lt 60 ]; do \
			if docker compose -f "$(ROOT)/deploy/docker-compose.test.yml" --env-file config/.env.test ps neo4j 2>/dev/null | grep -q "healthy"; then \
				printf "$(C_GREEN)[Integration]$(C_RESET) Neo4j is ready!\n"; \
				break; \
			fi; \
			sleep 2; \
			i=$$((i + 1)); \
		done; \
		if [ $$i -ge 60 ]; then \
			printf "$(C_RED)[Integration]$(C_RESET) ERROR: Neo4j did not become ready within 120 seconds\n"; \
			docker compose -f "$(ROOT)/deploy/docker-compose.test.yml" --env-file config/.env.test logs neo4j 2>&1 | tail -20; \
			exit 1; \
		fi; \
		touch /tmp/vedo-test-neo4j; \
	fi
	@printf "$(C_CYAN)[Integration]$(C_RESET) checking PostgreSQL...\n"
	@if command -v docker >/dev/null 2>&1 && docker compose -f "$(ROOT)/deploy/docker-compose.test.yml" --env-file config/.env.test ps postgres 2>/dev/null | grep -q "healthy"; then \
		printf "$(C_GREEN)[Integration]$(C_RESET) PostgreSQL is already running (Docker healthy)\n"; \
	else \
		printf "$(C_YELLOW)[Integration]$(C_RESET) Starting PostgreSQL via Docker Compose...\n"; \
		cd "$(ROOT)" && docker compose -f deploy/docker-compose.test.yml --env-file config/.env.test up -d postgres 2>&1; \
		printf "$(C_YELLOW)[Integration]$(C_RESET) Waiting for PostgreSQL healthcheck...\n"; \
		i=0; \
		while [ $$i -lt 60 ]; do \
			if docker compose -f "$(ROOT)/deploy/docker-compose.test.yml" --env-file config/.env.test ps postgres 2>/dev/null | grep -q "healthy"; then \
				printf "$(C_GREEN)[Integration]$(C_RESET) PostgreSQL is ready!\n"; \
				break; \
			fi; \
			sleep 2; \
			i=$$((i + 1)); \
		done; \
		if [ $$i -ge 60 ]; then \
			printf "$(C_RED)[Integration]$(C_RESET) ERROR: PostgreSQL did not become ready within 120 seconds\n"; \
			docker compose -f "$(ROOT)/deploy/docker-compose.test.yml" --env-file config/.env.test logs postgres 2>&1 | tail -20; \
			exit 1; \
		fi; \
		touch /tmp/vedo-test-pg; \
	fi
	@printf "$(C_CYAN)[Integration]$(C_RESET) ensuring vedo_org_test database exists...\n"
	@if command -v docker >/dev/null 2>&1; then \
		docker compose -f "$(ROOT)/deploy/docker-compose.test.yml" --env-file config/.env.test exec -T postgres \
			psql -U vedo -d postgres -tc "SELECT 1 FROM pg_database WHERE datname='vedo_org_test'" 2>/dev/null | grep -q 1 || \
		docker compose -f "$(ROOT)/deploy/docker-compose.test.yml" --env-file config/.env.test exec -T postgres \
			psql -U vedo -d postgres -c "CREATE DATABASE vedo_org_test" 2>&1; \
	fi
	@printf "$(C_CYAN)[Integration]$(C_RESET) running tests...\n"
	@$(MAKE) test-integration-rust-fast
	@$(MAKE) test-integration-go-fast
	@$(MAKE) test-versioning-fast
	@printf "$(C_CYAN)[Integration]$(C_RESET) cleaning up infrastructure...\n"
	@if [ -f /tmp/vedo-test-neo4j ]; then \
		printf "$(C_YELLOW)[Integration]$(C_RESET) Stopping auto-started Neo4j...\n"; \
		cd "$(ROOT)" && docker compose -f deploy/docker-compose.test.yml --env-file config/.env.test stop neo4j 2>&1; \
		rm -f /tmp/vedo-test-neo4j; \
	fi
	@if [ -f /tmp/vedo-test-pg ]; then \
		printf "$(C_YELLOW)[Integration]$(C_RESET) Stopping auto-started PostgreSQL...\n"; \
		cd "$(ROOT)" && docker compose -f deploy/docker-compose.test.yml --env-file config/.env.test stop postgres 2>&1; \
		rm -f /tmp/vedo-test-pg; \
	fi
	@printf "$(C_GREEN)[PASS]$(C_RESET) All integration tests passed\n"

test-integration-full: ## Integration tests — full statistics (auto-starts Neo4j + PostgreSQL via docker-compose.test.yml)
		@printf "$(C_CYAN)[Integration]$(C_RESET) full statistics mode\n"
		@printf "$(C_CYAN)[Integration]$(C_RESET) checking Neo4j...\n"
		@rm -f /tmp/vedo-test-neo4j /tmp/vedo-test-pg
		@if command -v docker >/dev/null 2>&1 && docker compose -f "$(ROOT)/deploy/docker-compose.test.yml" --env-file config/.env.test ps neo4j 2>/dev/null | grep -q "healthy"; then \
			printf "$(C_GREEN)[Integration]$(C_RESET) Neo4j is already running\n"; \
		else \
			printf "$(C_YELLOW)[Integration]$(C_RESET) Starting Neo4j...\n"; \
			cd "$(ROOT)" && docker compose -f deploy/docker-compose.test.yml --env-file config/.env.test up -d neo4j 2>&1; \
			printf "$(C_YELLOW)[Integration]$(C_RESET) Waiting for Neo4j healthcheck...\n"; \
			i=0; \
			while [ $$i -lt 60 ]; do \
				if docker compose -f "$(ROOT)/deploy/docker-compose.test.yml" --env-file config/.env.test ps neo4j 2>/dev/null | grep -q "healthy"; then \
					printf "$(C_GREEN)[Integration]$(C_RESET) Neo4j is ready!\n"; \
					break; \
				fi; \
				sleep 2; \
				i=$$((i + 1)); \
			done; \
			if [ $$i -ge 60 ]; then \
				printf "$(C_RED)[Integration]$(C_RESET) ERROR: Neo4j did not become ready within 120 seconds\n"; \
				docker compose -f "$(ROOT)/deploy/docker-compose.test.yml" --env-file config/.env.test logs neo4j 2>&1 | tail -20; \
				exit 1; \
			fi; \
			touch /tmp/vedo-test-neo4j; \
		fi
		@printf "$(C_CYAN)[Integration]$(C_RESET) checking PostgreSQL...\n"
		@if command -v docker >/dev/null 2>&1 && docker compose -f "$(ROOT)/deploy/docker-compose.test.yml" --env-file config/.env.test ps postgres 2>/dev/null | grep -q "healthy"; then \
			printf "$(C_GREEN)[Integration]$(C_RESET) PostgreSQL is already running (Docker healthy)\n"; \
		else \
			printf "$(C_YELLOW)[Integration]$(C_RESET) Starting PostgreSQL via Docker Compose...\n"; \
			cd "$(ROOT)" && docker compose -f deploy/docker-compose.test.yml --env-file config/.env.test up -d postgres 2>&1; \
			printf "$(C_YELLOW)[Integration]$(C_RESET) Waiting for PostgreSQL healthcheck...\n"; \
			i=0; \
			while [ $$i -lt 60 ]; do \
				if docker compose -f "$(ROOT)/deploy/docker-compose.test.yml" --env-file config/.env.test ps postgres 2>/dev/null | grep -q "healthy"; then \
					printf "$(C_GREEN)[Integration]$(C_RESET) PostgreSQL is ready!\n"; \
					break; \
				fi; \
				sleep 2; \
				i=$$((i + 1)); \
			done; \
			if [ $$i -ge 60 ]; then \
				printf "$(C_RED)[Integration]$(C_RESET) ERROR: PostgreSQL did not become ready within 120 seconds\n"; \
				docker compose -f "$(ROOT)/deploy/docker-compose.test.yml" --env-file config/.env.test logs postgres 2>&1 | tail -20; \
				exit 1; \
			fi; \
			touch /tmp/vedo-test-pg; \
		fi
		@printf "$(C_CYAN)[Integration]$(C_RESET) ensuring vedo_org_test database exists...\n"
		@if command -v docker >/dev/null 2>&1; then \
			docker compose -f "$(ROOT)/deploy/docker-compose.test.yml" --env-file config/.env.test exec -T postgres \
				psql -U vedo -d postgres -tc "SELECT 1 FROM pg_database WHERE datname='vedo_org_test'" 2>/dev/null | grep -q 1 || \
			docker compose -f "$(ROOT)/deploy/docker-compose.test.yml" --env-file config/.env.test exec -T postgres \
				psql -U vedo -d postgres -c "CREATE DATABASE vedo_org_test" 2>&1; \
		fi
		@printf "$(C_CYAN)[Integration]$(C_RESET) running tests...\n"
		@failed=0; \
		$(MAKE) test-integration-rust-full 2>&1 || failed=1; \
		$(MAKE) test-integration-go-full 2>&1 || failed=1; \
		$(MAKE) test-versioning-full 2>&1 || failed=1; \
		printf "$(C_CYAN)[Integration]$(C_RESET) cleaning up infrastructure...\n"; \
		if [ -f /tmp/vedo-test-neo4j ]; then \
			printf "$(C_YELLOW)[Integration]$(C_RESET) Stopping auto-started Neo4j...\n"; \
			cd "$(ROOT)" && docker compose -f deploy/docker-compose.test.yml --env-file config/.env.test stop neo4j 2>&1; \
			rm -f /tmp/vedo-test-neo4j; \
		fi; \
		if [ -f /tmp/vedo-test-pg ]; then \
			printf "$(C_YELLOW)[Integration]$(C_RESET) Stopping auto-started PostgreSQL...\n"; \
			cd "$(ROOT)" && docker compose -f deploy/docker-compose.test.yml --env-file config/.env.test stop postgres 2>&1; \
			rm -f /tmp/vedo-test-pg; \
		fi; \
		if [ $$failed -ne 0 ]; then \
			printf "$(C_RED)[FAIL]$(C_RESET) Some integration tests failed\n"; \
			exit 1; \
		fi; \
		printf "$(C_GREEN)[PASS]$(C_RESET) All integration tests passed\n"

.PHONY: test-integration-rust-fast test-integration-rust-full
test-integration-rust-fast: ## Rust integration tests — fail-fast (auto-starts Neo4j via docker-compose.test.yml)
	@NEO4J_STARTED=""; \
	NEO4J_HOST="localhost"; \
	NEO4J_PORT="$${NEO4J_BOLT_PORT:-$$(grep -E '^NEO4J_BOLT_PORT=' "$(ROOT)/config/.env.test" 2>/dev/null | tail -1 | cut -d= -f2 | tr -d '\r[:space:]')}"; \
	NEO4J_PORT="$${NEO4J_PORT:-7687}"; \
	NEO4J_USER="$${NEO4J_USER:-neo4j}"; \
	NEO4J_PASSWORD="$${NEO4J_PASSWORD:-password}"; \
	NEO4J_TEST_URI="bolt://$${NEO4J_HOST}:$${NEO4J_PORT}"; \
	NEO4J_COMPOSE="$(ROOT)/deploy/docker-compose.test.yml"; \
	NEO4J_ENV_FILE="$(ROOT)/config/.env.test"; \
	printf "$(C_CYAN)[Integration]$(C_RESET) checking Neo4j at $${NEO4J_HOST}:$${NEO4J_PORT}...\n"; \
	if command -v docker >/dev/null 2>&1 && docker compose -f "$$NEO4J_COMPOSE" --env-file "$$NEO4J_ENV_FILE" ps neo4j 2>/dev/null | grep -q "healthy"; then \
		printf "$(C_GREEN)[Integration]$(C_RESET) Neo4j is already running (Docker healthy)\n"; \
	else \
		printf "$(C_YELLOW)[Integration]$(C_RESET) Starting Neo4j via Docker Compose...\n"; \
		cd "$(ROOT)" && docker compose -f "$$NEO4J_COMPOSE" --env-file "$$NEO4J_ENV_FILE" up -d --force-recreate neo4j 2>&1; \
		NEO4J_STARTED="yes"; \
		printf "$(C_YELLOW)[Integration]$(C_RESET) Waiting for Neo4j healthcheck...\n"; \
		i=0; \
		while [ $$i -lt 60 ]; do \
			if docker compose -f "$$NEO4J_COMPOSE" --env-file "$$NEO4J_ENV_FILE" ps neo4j 2>/dev/null | grep -q "healthy"; then \
				printf "$(C_GREEN)[Integration]$(C_RESET) Neo4j is ready!\n"; \
				break; \
			fi; \
			sleep 2; \
			i=$$((i + 1)); \
		done; \
		if [ $$i -ge 60 ]; then \
			printf "$(C_RED)[Integration]$(C_RESET) ERROR: Neo4j did not become ready within 120 seconds\n"; \
			printf "$(C_RED)[Integration]$(C_RESET) Stale Neo4j locks (\"Neo4j is already running\") are cleared with:\n"; \
			printf "  docker compose -f deploy/docker-compose.test.yml --env-file config/.env.test down -v\n"; \
			docker compose -f "$$NEO4J_COMPOSE" --env-file "$$NEO4J_ENV_FILE" logs neo4j 2>&1 | tail -20; \
			exit 1; \
		fi; \
	fi; \
	export NEO4J_TEST_URI="$${NEO4J_TEST_URI}"; \
	export NEO4J_USER="$$NEO4J_USER"; \
	export NEO4J_PASSWORD="$$NEO4J_PASSWORD"; \
	export NEO4J_URI="$${NEO4J_TEST_URI}"; \
	printf "$(C_CYAN)[Integration]$(C_RESET) Running ontology-service Neo4j integration tests (fail-fast)...\n"; \
	cd "$(ROOT)/apps/services/ontology-service"; \
	for test_file in tests/*integration_test.rs tests/*p0_test.rs; do \
		test_name="$$(basename "$$test_file" .rs)"; \
		printf "  [$${test_name}]\n"; \
		cargo test --test "$${test_name}" -- --test-threads=1 --nocapture 2>&1 || { printf "$(C_RED)[FAIL]$(C_RESET) $${test_name}\n"; exit 1; }; \
	done; \
	if [ -n "$$NEO4J_STARTED" ]; then \
		printf "$(C_YELLOW)[Integration]$(C_RESET) Stopping auto-started Neo4j...\n"; \
		cd "$(ROOT)" && docker compose -f "$$NEO4J_COMPOSE" --env-file "$$NEO4J_ENV_FILE" stop neo4j 2>&1; \
	fi; \
	printf "$(C_GREEN)[PASS]$(C_RESET) All Rust integration tests passed\n"

test-integration-rust-full: ## Rust integration tests — full statistics (collect all failures)
	@NEO4J_STARTED=""; \
	NEO4J_HOST="localhost"; \
	NEO4J_PORT="$${NEO4J_BOLT_PORT:-$$(grep -E '^NEO4J_BOLT_PORT=' "$(ROOT)/config/.env.test" 2>/dev/null | tail -1 | cut -d= -f2 | tr -d '\r[:space:]')}"; \
	NEO4J_PORT="$${NEO4J_PORT:-7687}"; \
	NEO4J_USER="$${NEO4J_USER:-neo4j}"; \
	NEO4J_PASSWORD="$${NEO4J_PASSWORD:-password}"; \
	NEO4J_TEST_URI="bolt://$${NEO4J_HOST}:$${NEO4J_PORT}"; \
	NEO4J_COMPOSE="$(ROOT)/deploy/docker-compose.test.yml"; \
	NEO4J_ENV_FILE="$(ROOT)/config/.env.test"; \
	printf "$(C_CYAN)[Integration]$(C_RESET) checking Neo4j at $${NEO4J_HOST}:$${NEO4J_PORT}...\n"; \
	if command -v docker >/dev/null 2>&1 && docker compose -f "$$NEO4J_COMPOSE" --env-file "$$NEO4J_ENV_FILE" ps neo4j 2>/dev/null | grep -q "healthy"; then \
		printf "$(C_GREEN)[Integration]$(C_RESET) Neo4j is already running (Docker healthy)\n"; \
	else \
		printf "$(C_YELLOW)[Integration]$(C_RESET) Starting Neo4j via Docker Compose...\n"; \
		cd "$(ROOT)" && docker compose -f "$$NEO4J_COMPOSE" --env-file "$$NEO4J_ENV_FILE" up -d --force-recreate neo4j 2>&1; \
		NEO4J_STARTED="yes"; \
		printf "$(C_YELLOW)[Integration]$(C_RESET) Waiting for Neo4j healthcheck...\n"; \
		i=0; \
		while [ $$i -lt 60 ]; do \
			if docker compose -f "$$NEO4J_COMPOSE" --env-file "$$NEO4J_ENV_FILE" ps neo4j 2>/dev/null | grep -q "healthy"; then \
				printf "$(C_GREEN)[Integration]$(C_RESET) Neo4j is ready!\n"; \
				break; \
			fi; \
			sleep 2; \
			i=$$((i + 1)); \
		done; \
		if [ $$i -ge 60 ]; then \
			printf "$(C_RED)[Integration]$(C_RESET) ERROR: Neo4j did not become ready within 120 seconds\n"; \
			printf "$(C_RED)[Integration]$(C_RESET) Stale Neo4j locks (\"Neo4j is already running\") are cleared with:\n"; \
			printf "  docker compose -f deploy/docker-compose.test.yml --env-file config/.env.test down -v\n"; \
			docker compose -f "$$NEO4J_COMPOSE" --env-file "$$NEO4J_ENV_FILE" logs neo4j 2>&1 | tail -20; \
			exit 1; \
		fi; \
	fi; \
	export NEO4J_TEST_URI="$${NEO4J_TEST_URI}"; \
	export NEO4J_USER="$$NEO4J_USER"; \
	export NEO4J_PASSWORD="$$NEO4J_PASSWORD"; \
	export NEO4J_URI="$${NEO4J_TEST_URI}"; \
	printf "$(C_CYAN)[Integration]$(C_RESET) Running ontology-service Neo4j integration tests (full)...\n"; \
	cd "$(ROOT)/apps/services/ontology-service"; \
	RESULT=0; \
	for test_file in tests/*integration_test.rs tests/*p0_test.rs; do \
		test_name="$$(basename "$$test_file" .rs)"; \
		printf "  [$${test_name}]\n"; \
		cargo test --test "$${test_name}" -- --test-threads=1 --nocapture 2>&1 || { RESULT=1; printf "$(C_RED)[FAIL]$(C_RESET) $${test_name}\n"; }; \
	done; \
	if [ -n "$$NEO4J_STARTED" ]; then \
		printf "$(C_YELLOW)[Integration]$(C_RESET) Stopping auto-started Neo4j...\n"; \
		cd "$(ROOT)" && docker compose -f "$$NEO4J_COMPOSE" --env-file "$$NEO4J_ENV_FILE" stop neo4j 2>&1; \
		cd "$(ROOT)/apps/services/ontology-service"; \
	fi; \
	if [ $$RESULT -ne 0 ]; then \
		printf "$(C_RED)[FAIL]$(C_RESET) Some Rust integration tests failed\n"; \
		exit 1; \
	fi; \
	printf "$(C_GREEN)[PASS]$(C_RESET) All Rust integration tests passed\n"

.PHONY: test-integration-go-fast test-integration-go-full
test-integration-go-fast: ## Go integration tests — fail-fast (auto-starts PostgreSQL via docker-compose.test.yml)
	@PG_STARTED=""; \
	PG_HOST="localhost"; \
	PG_PORT="$${PG_TEST_PORT:-$$(grep -E '^POSTGRES_PORT=' "$(ROOT)/config/.env.test" 2>/dev/null | tail -1 | cut -d= -f2 | tr -d '\r[:space:]')}"; \
	PG_PORT="$${PG_PORT:-15432}"; \
	PG_USER="$${PG_USER:-vedo}"; \
	PG_PASSWORD="$${PG_PASSWORD:-vedo}"; \
	PG_COMPOSE="$(ROOT)/deploy/docker-compose.test.yml"; \
	PG_ENV_FILE="$(ROOT)/config/.env.test"; \
	PG_URL="postgres://$${PG_USER}:$${PG_PASSWORD}@$${PG_HOST}:$${PG_PORT}/vedo_org_test?sslmode=disable"; \
	printf "$(C_CYAN)[Go]$(C_RESET) checking PostgreSQL at $${PG_HOST}:$${PG_PORT}...\n"; \
	if command -v docker >/dev/null 2>&1 && docker compose -f "$$PG_COMPOSE" --env-file "$$PG_ENV_FILE" ps postgres 2>/dev/null | grep -q "healthy"; then \
		printf "$(C_GREEN)[Go]$(C_RESET) PostgreSQL is already running (Docker healthy)\n"; \
	else \
		printf "$(C_YELLOW)[Go]$(C_RESET) Starting PostgreSQL via Docker Compose...\n"; \
		cd "$(ROOT)" && docker compose -f "$$PG_COMPOSE" --env-file "$$PG_ENV_FILE" up -d postgres 2>&1; \
		PG_STARTED="yes"; \
		printf "$(C_YELLOW)[Go]$(C_RESET) Waiting for PostgreSQL healthcheck...\n"; \
		i=0; \
		while [ $$i -lt 60 ]; do \
			if docker compose -f "$$PG_COMPOSE" --env-file "$$PG_ENV_FILE" ps postgres 2>/dev/null | grep -q "healthy"; then \
				printf "$(C_GREEN)[Go]$(C_RESET) PostgreSQL is ready!\n"; \
				break; \
			fi; \
			sleep 2; \
			i=$$((i + 1)); \
		done; \
		if [ $$i -ge 60 ]; then \
			printf "$(C_RED)[Go]$(C_RESET) ERROR: PostgreSQL did not become ready within 120 seconds\n"; \
			docker compose -f "$$PG_COMPOSE" --env-file "$$PG_ENV_FILE" logs postgres 2>&1 | tail -20; \
			exit 1; \
		fi; \
	fi; \
	printf "$(C_CYAN)[Go]$(C_RESET) ensuring vedo_org_test database exists...\n"; \
	docker compose -f "$$PG_COMPOSE" --env-file "$$PG_ENV_FILE" exec -T postgres \
		psql -U "$$PG_USER" -d postgres -tc "SELECT 1 FROM pg_database WHERE datname='vedo_org_test'" 2>/dev/null | grep -q 1 || \
	docker compose -f "$$PG_COMPOSE" --env-file "$$PG_ENV_FILE" exec -T postgres \
		psql -U "$$PG_USER" -d postgres -c "CREATE DATABASE vedo_org_test" 2>&1; \
	export DATABASE_URL="$$PG_URL"; \
	for dir in ticket-api org-api; do \
		p="$(ROOT)/tests/integration/$$dir"; \
		if [ -d "$$p" ] && (cd "$$p" && go list -tags=integration ./... >/dev/null 2>&1); then \
			printf "$(C_CYAN)[Go]$(C_RESET) integration — $$dir\n"; \
			cd "$$p" && go test -tags=integration ./... 2>&1 || { printf "$(C_RED)[FAIL]$(C_RESET) $$dir\n"; exit 1; }; \
		fi; \
	done; \
	if [ -n "$$PG_STARTED" ]; then \
		printf "$(C_YELLOW)[Go]$(C_RESET) Stopping auto-started PostgreSQL...\n"; \
		cd "$(ROOT)" && docker compose -f "$$PG_COMPOSE" --env-file "$$PG_ENV_FILE" stop postgres 2>&1; \
	fi; \
	printf "$(C_GREEN)[PASS]$(C_RESET) All Go integration tests passed\n"

test-integration-go-full: ## Go integration tests — full statistics (auto-starts PostgreSQL via docker-compose.test.yml)
		@PG_STARTED=""; \
		PG_HOST="localhost"; \
		PG_PORT="$${PG_TEST_PORT:-$$(grep -E '^POSTGRES_PORT=' "$(ROOT)/config/.env.test" 2>/dev/null | tail -1 | cut -d= -f2 | tr -d '\r[:space:]')}"; \
		PG_PORT="$${PG_PORT:-15432}"; \
		PG_USER="$${PG_USER:-vedo}"; \
		PG_PASSWORD="$${PG_PASSWORD:-vedo}"; \
		PG_COMPOSE="$(ROOT)/deploy/docker-compose.test.yml"; \
		PG_ENV_FILE="$(ROOT)/config/.env.test"; \
		PG_URL="postgres://$${PG_USER}:$${PG_PASSWORD}@$${PG_HOST}:$${PG_PORT}/vedo_org_test?sslmode=disable"; \
		printf "$(C_CYAN)[Go]$(C_RESET) checking PostgreSQL at $${PG_HOST}:$${PG_PORT}...\n"; \
		if command -v docker >/dev/null 2>&1 && docker compose -f "$$PG_COMPOSE" --env-file "$$PG_ENV_FILE" ps postgres 2>/dev/null | grep -q "healthy"; then \
			printf "$(C_GREEN)[Go]$(C_RESET) PostgreSQL is already running (Docker healthy)\n"; \
		else \
			printf "$(C_YELLOW)[Go]$(C_RESET) Starting PostgreSQL via Docker Compose...\n"; \
			cd "$(ROOT)" && docker compose -f "$$PG_COMPOSE" --env-file "$$PG_ENV_FILE" up -d postgres 2>&1; \
			PG_STARTED="yes"; \
			printf "$(C_YELLOW)[Go]$(C_RESET) Waiting for PostgreSQL healthcheck...\n"; \
			i=0; \
			while [ $$i -lt 60 ]; do \
				if docker compose -f "$$PG_COMPOSE" --env-file "$$PG_ENV_FILE" ps postgres 2>/dev/null | grep -q "healthy"; then \
					printf "$(C_GREEN)[Go]$(C_RESET) PostgreSQL is ready!\n"; \
					break; \
				fi; \
				sleep 2; \
				i=$$((i + 1)); \
			done; \
			if [ $$i -ge 60 ]; then \
				printf "$(C_RED)[Go]$(C_RESET) ERROR: PostgreSQL did not become ready within 120 seconds\n"; \
				docker compose -f "$$PG_COMPOSE" --env-file "$$PG_ENV_FILE" logs postgres 2>&1 | tail -20; \
				exit 1; \
			fi; \
		fi; \
		printf "$(C_CYAN)[Go]$(C_RESET) ensuring vedo_org_test database exists...\n"; \
		docker compose -f "$$PG_COMPOSE" --env-file "$$PG_ENV_FILE" exec -T postgres \
			psql -U "$$PG_USER" -d postgres -tc "SELECT 1 FROM pg_database WHERE datname='vedo_org_test'" 2>/dev/null | grep -q 1 || \
		docker compose -f "$$PG_COMPOSE" --env-file "$$PG_ENV_FILE" exec -T postgres \
			psql -U "$$PG_USER" -d postgres -c "CREATE DATABASE vedo_org_test" 2>&1; \
		export DATABASE_URL="$$PG_URL"; \
		failed=""; \
		for dir in ticket-api org-api; do \
			p="$(ROOT)/tests/integration/$$dir"; \
			if [ -d "$$p" ] && (cd "$$p" && go list -tags=integration ./... >/dev/null 2>&1); then \
				printf "$(C_CYAN)[Go]$(C_RESET) integration — $$dir\n"; \
				cd "$$p" && go test -tags=integration ./... 2>&1 || failed="$$failed $$dir"; \
			fi; \
		done; \
		if [ -n "$$PG_STARTED" ]; then \
			printf "$(C_YELLOW)[Go]$(C_RESET) Stopping auto-started PostgreSQL...\n"; \
			cd "$(ROOT)" && docker compose -f "$$PG_COMPOSE" --env-file "$$PG_ENV_FILE" stop postgres 2>&1; \
		fi; \
		if [ -n "$$failed" ]; then \
			printf "$(C_RED)[FAIL]$(C_RESET) Go integration tests failed in:$$failed\n"; \
			exit 1; \
		fi; \
		printf "$(C_GREEN)[PASS]$(C_RESET) All Go integration tests passed\n"

.PHONY: test-versioning-fast test-versioning-full
test-versioning-fast: ## Versioning integration — fail-fast (auto-starts PostgreSQL if not running)
	@PG_STARTED=""; \
	PG_HOST="localhost"; \
	PG_PORT="$${PG_TEST_PORT:-$$(grep -E '^POSTGRES_PORT=' "$(ROOT)/config/.env.test" 2>/dev/null | tail -1 | cut -d= -f2 | tr -d '\r[:space:]')}"; \
	PG_PORT="$${PG_PORT:-15432}"; \
	PG_USER="$${PG_USER:-vedo}"; \
	PG_PASSWORD="$${PG_PASSWORD:-vedo}"; \
	PG_COMPOSE="$(ROOT)/deploy/docker-compose.test.yml"; \
	PG_ENV_FILE="$(ROOT)/config/.env.test"; \
	PG_URL="postgres://$${PG_USER}:$${PG_PASSWORD}@$${PG_HOST}:$${PG_PORT}/vedo_versioning"; \
	printf "$(C_CYAN)[Versioning]$(C_RESET) checking PostgreSQL at $${PG_HOST}:$${PG_PORT}...\n"; \
	if command -v docker >/dev/null 2>&1 && docker compose -f "$$PG_COMPOSE" --env-file "$$PG_ENV_FILE" ps postgres 2>/dev/null | grep -q "healthy"; then \
		printf "$(C_GREEN)[Versioning]$(C_RESET) PostgreSQL is already running (Docker healthy)\n"; \
	else \
		printf "$(C_YELLOW)[Versioning]$(C_RESET) Starting PostgreSQL via Docker Compose...\n"; \
		cd "$(ROOT)" && docker compose -f "$$PG_COMPOSE" --env-file "$$PG_ENV_FILE" up -d postgres 2>&1; \
		PG_STARTED="yes"; \
		printf "$(C_YELLOW)[Versioning]$(C_RESET) Waiting for PostgreSQL healthcheck...\n"; \
		i=0; \
		while [ $$i -lt 60 ]; do \
			if docker compose -f "$$PG_COMPOSE" --env-file "$$PG_ENV_FILE" ps postgres 2>/dev/null | grep -q "healthy"; then \
				printf "$(C_GREEN)[Versioning]$(C_RESET) PostgreSQL is ready!\n"; \
				break; \
			fi; \
			sleep 2; \
			i=$$((i + 1)); \
		done; \
		if [ $$i -ge 60 ]; then \
			printf "$(C_RED)[Versioning]$(C_RESET) ERROR: PostgreSQL did not become ready within 120 seconds\n"; \
			docker compose -f "$$PG_COMPOSE" --env-file "$$PG_ENV_FILE" logs postgres 2>&1 | tail -20; \
			exit 1; \
		fi; \
	fi; \
	printf "$(C_CYAN)[Versioning]$(C_RESET) ensuring vedo_versioning database exists...\n"; \
	docker compose -f "$$PG_COMPOSE" --env-file "$$PG_ENV_FILE" exec -T postgres \
		psql -U "$$PG_USER" -d postgres -tc "SELECT 1 FROM pg_database WHERE datname='vedo_versioning'" 2>/dev/null | grep -q 1 || \
	docker compose -f "$$PG_COMPOSE" --env-file "$$PG_ENV_FILE" exec -T postgres \
		psql -U "$$PG_USER" -d postgres -c "CREATE DATABASE vedo_versioning" 2>&1; \
	export PG_TEST_DATABASE_URL="$$PG_URL"; \
	export DATABASE_URL="$$PG_URL"; \
	RESULT=0; \
	printf "$(C_CYAN)[Versioning]$(C_RESET) Running unit tests (fail-fast)...\n"; \
	cd "$(ROOT)/apps/services/versioning-service" && cargo test --lib 2>&1 || RESULT=1; \
	if [ $$RESULT -eq 0 ]; then \
		printf "$(C_CYAN)[Versioning]$(C_RESET) Running integration tests (fail-fast)...\n"; \
		cd "$(ROOT)/apps/services/versioning-service" && cargo test --test '*' -- --test-threads=1 --nocapture 2>&1 || RESULT=1; \
	fi; \
	if [ -n "$$PG_STARTED" ]; then \
		printf "$(C_YELLOW)[Versioning]$(C_RESET) Stopping auto-started PostgreSQL...\n"; \
		cd "$(ROOT)" && docker compose -f "$$PG_COMPOSE" --env-file "$$PG_ENV_FILE" stop postgres 2>&1; \
	fi; \
	if [ $$RESULT -ne 0 ]; then \
		printf "$(C_RED)[FAIL]$(C_RESET) Versioning tests failed\n"; \
		exit 1; \
	fi; \
	printf "$(C_GREEN)[PASS]$(C_RESET) All versioning tests passed\n"

test-versioning-full: ## Versioning integration — full statistics (collect all failures)
	@PG_STARTED=""; \
	PG_HOST="localhost"; \
	PG_PORT="$${PG_TEST_PORT:-$$(grep -E '^POSTGRES_PORT=' "$(ROOT)/config/.env.test" 2>/dev/null | tail -1 | cut -d= -f2 | tr -d '\r[:space:]')}"; \
	PG_PORT="$${PG_PORT:-15432}"; \
	PG_USER="$${PG_USER:-vedo}"; \
	PG_PASSWORD="$${PG_PASSWORD:-vedo}"; \
	PG_COMPOSE="$(ROOT)/deploy/docker-compose.test.yml"; \
	PG_ENV_FILE="$(ROOT)/config/.env.test"; \
	PG_URL="postgres://$${PG_USER}:$${PG_PASSWORD}@$${PG_HOST}:$${PG_PORT}/vedo_versioning"; \
	printf "$(C_CYAN)[Versioning]$(C_RESET) checking PostgreSQL at $${PG_HOST}:$${PG_PORT}...\n"; \
	if command -v docker >/dev/null 2>&1 && docker compose -f "$$PG_COMPOSE" --env-file "$$PG_ENV_FILE" ps postgres 2>/dev/null | grep -q "healthy"; then \
		printf "$(C_GREEN)[Versioning]$(C_RESET) PostgreSQL is already running (Docker healthy)\n"; \
	else \
		printf "$(C_YELLOW)[Versioning]$(C_RESET) Starting PostgreSQL via Docker Compose...\n"; \
		cd "$(ROOT)" && docker compose -f "$$PG_COMPOSE" --env-file "$$PG_ENV_FILE" up -d postgres 2>&1; \
		PG_STARTED="yes"; \
		printf "$(C_YELLOW)[Versioning]$(C_RESET) Waiting for PostgreSQL healthcheck...\n"; \
		i=0; \
		while [ $$i -lt 60 ]; do \
			if docker compose -f "$$PG_COMPOSE" --env-file "$$PG_ENV_FILE" ps postgres 2>/dev/null | grep -q "healthy"; then \
				printf "$(C_GREEN)[Versioning]$(C_RESET) PostgreSQL is ready!\n"; \
				break; \
			fi; \
			sleep 2; \
			i=$$((i + 1)); \
		done; \
		if [ $$i -ge 60 ]; then \
			printf "$(C_RED)[Versioning]$(C_RESET) ERROR: PostgreSQL did not become ready within 120 seconds\n"; \
			docker compose -f "$$PG_COMPOSE" --env-file "$$PG_ENV_FILE" logs postgres 2>&1 | tail -20; \
			exit 1; \
		fi; \
	fi; \
	printf "$(C_CYAN)[Versioning]$(C_RESET) ensuring vedo_versioning database exists...\n"; \
	docker compose -f "$$PG_COMPOSE" --env-file "$$PG_ENV_FILE" exec -T postgres \
		psql -U "$$PG_USER" -d postgres -tc "SELECT 1 FROM pg_database WHERE datname='vedo_versioning'" 2>/dev/null | grep -q 1 || \
	docker compose -f "$$PG_COMPOSE" --env-file "$$PG_ENV_FILE" exec -T postgres \
		psql -U "$$PG_USER" -d postgres -c "CREATE DATABASE vedo_versioning" 2>&1; \
	export PG_TEST_DATABASE_URL="$$PG_URL"; \
	export DATABASE_URL="$$PG_URL"; \
	RESULT=0; \
	printf "$(C_CYAN)[Versioning]$(C_RESET) Running unit tests...\n"; \
	cd "$(ROOT)/apps/services/versioning-service" && cargo test --lib 2>&1 || RESULT=1; \
	printf "$(C_CYAN)[Versioning]$(C_RESET) Running integration tests...\n"; \
	cd "$(ROOT)/apps/services/versioning-service" && cargo test --test '*' -- --test-threads=1 --nocapture 2>&1 || RESULT=1; \
	if [ -n "$$PG_STARTED" ]; then \
		printf "$(C_YELLOW)[Versioning]$(C_RESET) Stopping auto-started PostgreSQL...\n"; \
		cd "$(ROOT)" && docker compose -f "$$PG_COMPOSE" --env-file "$$PG_ENV_FILE" stop postgres 2>&1; \
	fi; \
	if [ $$RESULT -ne 0 ]; then \
		printf "$(C_RED)[FAIL]$(C_RESET) Versioning tests failed\n"; \
		exit 1; \
	fi; \
	printf "$(C_GREEN)[PASS]$(C_RESET) All versioning tests passed\n"

##@ Test — E2E (requires Docker test stack)

.PHONY: test-e2e-fast test-e2e-full
test-e2e-fast: ## E2E — fail-fast (API + GUI)
	@printf "$(C_CYAN)[E2E]$(C_RESET) fail-fast mode\n"
	@$(MAKE) test-e2e-api-fast
	@$(MAKE) test-e2e-gui-fast
	@printf "$(C_GREEN)[PASS]$(C_RESET) All E2E tests passed\n"

test-e2e-full: ## E2E — full statistics (API + GUI)
	@failed=0; \
	printf "$(C_CYAN)[E2E]$(C_RESET) full statistics mode\n"; \
	$(MAKE) test-e2e-api-full 2>&1 || failed=1; \
	$(MAKE) test-e2e-gui-full 2>&1 || failed=1; \
	if [ $$failed -ne 0 ]; then \
		printf "$(C_RED)[FAIL]$(C_RESET) Some E2E tests failed\n"; \
		exit 1; \
	fi; \
	printf "$(C_GREEN)[PASS]$(C_RESET) All E2E tests passed\n"

.PHONY: test-e2e-api-fast test-e2e-api-full
test-e2e-api-fast: ## E2E API tests — fail-fast (max-failures=1)
	@printf "$(C_CYAN)[E2E]$(C_RESET) installing dependencies...\n"
	@cd "$(ROOT)/tests/e2e" && CI=true pnpm install --frozen-lockfile 2>&1 || CI=true pnpm install 2>&1
	@cd "$(ROOT)/tests/e2e" && CI=true npx --yes playwright install chromium 2>&1 || true
	@printf "$(C_CYAN)[E2E]$(C_RESET) running API tests (fail-fast)...\n"
	@cd "$(ROOT)/tests/e2e" && pnpm exec playwright test --config=config/playwright.api.config.ts --max-failures=1 && \
		printf "$(C_GREEN)[PASS]$(C_RESET) E2E API tests passed\n" || \
		{ printf "$(C_RED)[FAIL]$(C_RESET) E2E API tests failed\n"; exit 1; }

test-e2e-api-full: ## E2E API tests — full run (with retries, no max-failures)
	@printf "$(C_CYAN)[E2E]$(C_RESET) installing dependencies...\n"
	@cd "$(ROOT)/tests/e2e" && CI=true pnpm install --frozen-lockfile 2>&1 || CI=true pnpm install 2>&1
	@cd "$(ROOT)/tests/e2e" && CI=true npx --yes playwright install chromium 2>&1 || true
	@printf "$(C_CYAN)[E2E]$(C_RESET) running API tests (full)...\n"
	@cd "$(ROOT)/tests/e2e" && pnpm exec playwright test --config=config/playwright.api.config.ts && \
		printf "$(C_GREEN)[PASS]$(C_RESET) E2E API tests passed\n" || \
		{ printf "$(C_RED)[FAIL]$(C_RESET) E2E API tests failed\n"; exit 1; }

.PHONY: test-e2e-gui-fast test-e2e-gui-full
test-e2e-gui-fast: ## E2E GUI tests — fail-fast (maxFailures=1 in config)
	@printf "$(C_CYAN)[E2E]$(C_RESET) installing dependencies...\n"
	@cd "$(ROOT)/tests/e2e" && CI=true pnpm install --frozen-lockfile 2>&1 || CI=true pnpm install 2>&1
	@cd "$(ROOT)/tests/e2e" && CI=true npx --yes playwright install chromium 2>&1 || true
	@printf "$(C_CYAN)[E2E]$(C_RESET) running GUI tests (fail-fast)...\n"
	@cd "$(ROOT)/tests/e2e" && pnpm exec playwright test --config=config/playwright.gui.config.ts && \
		printf "$(C_GREEN)[PASS]$(C_RESET) E2E GUI tests passed\n" || \
		{ printf "$(C_RED)[FAIL]$(C_RESET) E2E GUI tests failed\n"; exit 1; }

test-e2e-gui-full: ## E2E GUI tests — full run (no max-failures)
	@printf "$(C_CYAN)[E2E]$(C_RESET) installing dependencies...\n"
	@cd "$(ROOT)/tests/e2e" && CI=true pnpm install --frozen-lockfile 2>&1 || CI=true pnpm install 2>&1
	@cd "$(ROOT)/tests/e2e" && CI=true npx --yes playwright install chromium 2>&1 || true
	@printf "$(C_CYAN)[E2E]$(C_RESET) running GUI tests (full)...\n"
	@cd "$(ROOT)/tests/e2e" && pnpm exec playwright test --config=config/playwright.gui.config.ts --max-failures=0 && \
		printf "$(C_GREEN)[PASS]$(C_RESET) E2E GUI tests passed\n" || \
		{ printf "$(C_RED)[FAIL]$(C_RESET) E2E GUI tests failed\n"; exit 1; }

.PHONY: test-gates-fast test-gates-full test-gates-security-fast test-gates-security-full
test-gates-fast: ## Gate tests — fail-fast (unit + static, no infra needed)
		@printf "$(C_CYAN)[Gates]$(C_RESET) fail-fast mode\n"
		@printf "$(C_CYAN)[Gates]$(C_RESET) shell script syntax (bash -n)...\n"
		@bash $(ROOT)/tests/gates/test_shell_syntax.sh || { printf "$(C_RED)[FAIL]$(C_RESET) shell-syntax gate failed\n"; exit 1; }
		@printf "$(C_CYAN)[Gates]$(C_RESET) contract tests (Go unit + build checks)...\n"
		@bash $(ROOT)/tests/gates/test_contract_gate.sh || { printf "$(C_RED)[FAIL]$(C_RESET) contract gate failed\n"; exit 1; }
		@printf "$(C_CYAN)[Gates]$(C_RESET) BOLA/BFLA unit tests (auth middleware)...\n"
		@bash $(ROOT)/tests/gates/test_bola_bfla_gate.sh || { printf "$(C_RED)[FAIL]$(C_RESET) bola-bfla gate failed\n"; exit 1; }
		@printf "$(C_CYAN)[Gates]$(C_RESET) Python service manifest validation...\n"
		@bash $(ROOT)/tests/gates/test_python_manifests.sh || { printf "$(C_RED)[FAIL]$(C_RESET) python-manifests gate failed\n"; exit 1; }
		@printf "$(C_CYAN)[Gates]$(C_RESET) static test quality (anti-patterns + TQS)...\n"
		@bash $(ROOT)/tools/scripts/test-quality-gate.sh --score $(ROOT)/apps/services || { printf "$(C_RED)[FAIL]$(C_RESET) test-quality gate failed\n"; exit 1; }
		@printf "$(C_CYAN)[Gates]$(C_RESET) traceability (TTL → RCS)...\n"
		@bash $(ROOT)/tools/scripts/traceability-validator.sh || { printf "$(C_RED)[FAIL]$(C_RESET) traceability gate failed\n"; exit 1; }
		@printf "$(C_GREEN)[PASS]$(C_RESET) all gates + quality checks passed\n"

test-gates-full: ## Gate tests — full statistics (unit + static, collect all failures)
		@gate_failures=""; \
		printf "$(C_CYAN)[Gates]$(C_RESET) full statistics mode\n"; \
		printf "$(C_CYAN)[Gates]$(C_RESET) shell script syntax...\n"; \
		bash $(ROOT)/tests/gates/test_shell_syntax.sh 2>&1 || gate_failures="$${gate_failures} shell-syntax"; \
		printf "$(C_CYAN)[Gates]$(C_RESET) contract tests...\n"; \
		bash $(ROOT)/tests/gates/test_contract_gate.sh 2>&1 || gate_failures="$${gate_failures} contract"; \
		printf "$(C_CYAN)[Gates]$(C_RESET) BOLA/BFLA unit tests...\n"; \
		bash $(ROOT)/tests/gates/test_bola_bfla_gate.sh 2>&1 || gate_failures="$${gate_failures} bola-bfla"; \
		printf "$(C_CYAN)[Gates]$(C_RESET) Python service manifest validation...\n"; \
		bash $(ROOT)/tests/gates/test_python_manifests.sh 2>&1 || gate_failures="$${gate_failures} python-manifests"; \
		printf "$(C_CYAN)[Gates]$(C_RESET) static test quality...\n"; \
		bash $(ROOT)/tools/scripts/test-quality-gate.sh --score $(ROOT)/apps/services 2>&1 || gate_failures="$${gate_failures} test-quality"; \
		printf "$(C_CYAN)[Gates]$(C_RESET) traceability...\n"; \
		bash $(ROOT)/tools/scripts/traceability-validator.sh 2>&1 || gate_failures="$${gate_failures} traceability"; \
		if [ -n "$$gate_failures" ]; then \
			printf "$(C_RED)[FAIL]$(C_RESET) Gates FAILED:$$gate_failures\n"; \
			exit 1; \
		fi; \
		printf "$(C_GREEN)[PASS]$(C_RESET) all gates + quality checks passed\n"

test-gates-security-fast: ## Security integration tests — fail-fast (requires Docker test stack, auto-starts if missing)
		@printf "$(C_CYAN)[Security]$(C_RESET) checking Docker test stack availability...\n"
		@rm -f /tmp/vedo-test-stack
		@if curl -sf http://localhost:8080/api/v1/health >/dev/null 2>&1; then \
			printf "$(C_GREEN)[Security]$(C_RESET) API Gateway is available\n"; \
		else \
			printf "$(C_YELLOW)[Security]$(C_RESET) API Gateway not available — starting Docker test stack...\n"; \
			cd "$(ROOT)" && $(MAKE) docker-up-test 2>&1; \
			touch /tmp/vedo-test-stack; \
			printf "$(C_YELLOW)[Security]$(C_RESET) Waiting for API Gateway readiness...\n"; \
			i=0; \
			while [ $$i -lt 120 ]; do \
				if curl -sf http://localhost:8080/api/v1/health >/dev/null 2>&1; then \
					printf "$(C_GREEN)[Security]$(C_RESET) API Gateway is ready!\n"; \
					break; \
				fi; \
				sleep 5; \
				i=$$((i + 1)); \
			done; \
			if [ $$i -ge 120 ]; then \
				printf "$(C_RED)[Security]$(C_RESET) ERROR: API Gateway did not become ready within 10 minutes\n"; \
				exit 1; \
			fi; \
		fi
		@printf "$(C_CYAN)[Security]$(C_RESET) running BOLA/BFLA/RBAC full-stack integration tests...\n"
		@RESULT=0; \
		bash $(ROOT)/tests/gates/test_security_integration.sh 2>&1 || RESULT=1; \
		if [ -f /tmp/vedo-test-stack ]; then \
			printf "$(C_YELLOW)[Security]$(C_RESET) Stopping auto-started Docker test stack...\n"; \
			cd "$(ROOT)" && docker compose -f deploy/docker-compose.test.yml --env-file config/.env.test down 2>&1; \
			rm -f /tmp/vedo-test-stack; \
		fi; \
		if [ $$RESULT -ne 0 ]; then \
			printf "$(C_RED)[FAIL]$(C_RESET) Security integration tests failed\n"; \
			exit 1; \
		fi; \
		printf "$(C_GREEN)[PASS]$(C_RESET) All security integration tests passed\n"

test-gates-security-full: ## Security integration tests — full statistics (requires Docker test stack, auto-starts if missing)
		@printf "$(C_CYAN)[Security]$(C_RESET) checking Docker test stack availability...\n"
		@rm -f /tmp/vedo-test-stack
		@if curl -sf http://localhost:8080/api/v1/health >/dev/null 2>&1; then \
			printf "$(C_GREEN)[Security]$(C_RESET) API Gateway is available\n"; \
		else \
			printf "$(C_YELLOW)[Security]$(C_RESET) API Gateway not available — starting Docker test stack...\n"; \
			cd "$(ROOT)" && $(MAKE) docker-up-test 2>&1; \
			touch /tmp/vedo-test-stack; \
			printf "$(C_YELLOW)[Security]$(C_RESET) Waiting for API Gateway readiness...\n"; \
			i=0; \
			while [ $$i -lt 120 ]; do \
				if curl -sf http://localhost:8080/api/v1/health >/dev/null 2>&1; then \
					printf "$(C_GREEN)[Security]$(C_RESET) API Gateway is ready!\n"; \
					break; \
				fi; \
				sleep 5; \
				i=$$((i + 1)); \
			done; \
			if [ $$i -ge 120 ]; then \
				printf "$(C_RED)[Security]$(C_RESET) ERROR: API Gateway did not become ready within 10 minutes\n"; \
				exit 1; \
			fi; \
		fi
		@printf "$(C_CYAN)[Security]$(C_RESET) running BOLA/BFLA/RBAC full-stack integration tests...\n"
		@RESULT=0; \
		bash $(ROOT)/tests/gates/test_security_integration.sh 2>&1 || RESULT=1; \
		if [ -f /tmp/vedo-test-stack ]; then \
			printf "$(C_YELLOW)[Security]$(C_RESET) Stopping auto-started Docker test stack...\n"; \
			cd "$(ROOT)" && docker compose -f deploy/docker-compose.test.yml --env-file config/.env.test down 2>&1; \
			rm -f /tmp/vedo-test-stack; \
		fi; \
		if [ $$RESULT -ne 0 ]; then \
			printf "$(C_RED)[FAIL]$(C_RESET) Security integration tests failed\n"; \
			exit 1; \
		fi; \
		printf "$(C_GREEN)[PASS]$(C_RESET) All security integration tests passed\n"

.PHONY: coverage
coverage: ## Run tests with coverage (Go only)
	@printf "$(C_CYAN)[Coverage]$(C_RESET) running tests with coverage...\n"
	@if [ -n "$(GO_DIRS)" ]; then \
		for dir in $(GO_DIRS); do \
			printf "$(C_CYAN)[Go]$(C_RESET) coverage for $$(basename $$dir)\n"; \
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

# ENV=test uses docker-compose.test.yml (isolated project + JWT dev keys for Playwright)
ifeq ($(ENV),test)
COMPOSE_FILE := deploy/docker-compose.test.yml
endif

.PHONY: docker-up docker-down docker-logs docker-ps docker-shell docker-config
.PHONY: docker-up-dev docker-up-test docker-up-staging
.PHONY: docker-down-dev docker-down-test docker-down-staging

docker-up: ## Start all services via Docker Compose (usage: make docker-up [ENV=dev|test|staging])
	SERVICE_VERSION=$(VERSION) docker compose \
		--env-file $(ROOT)/config/.env.$(ENV) \
		-f $(ROOT)/$(COMPOSE_FILE) \
		$(if $(COMPOSE_PROFILE),--profile $(COMPOSE_PROFILE)) \
		up -d

docker-env-check: ## Show resolved environment values for current ENV
	@echo "ENV=$(ENV)"
	@echo "COMPOSE_FILE=$(COMPOSE_FILE)"
	@echo "SERVICE_VERSION=$(VERSION)"
	@SERVICE_VERSION=$(VERSION) docker compose --env-file $(ROOT)/config/.env.$(ENV) -f $(ROOT)/$(COMPOSE_FILE) config 2>/dev/null | grep -E "KC_HOSTNAME_PORT|VEDO_KEYCLOAK_URL|SERVICE_VERSION|FRONTEND_PORT|API_GATEWAY_PORT" || echo "(no matches — compose config may be invalid)"

docker-up-dev: ## Start dev environment (alias for make docker-up ENV=dev)
	$(MAKE) docker-up ENV=dev

docker-up-test: ## Start test environment (docker-compose.test.yml — isolated project + JWT dev keys for Playwright)
	$(MAKE) docker-up ENV=test COMPOSE_FILE=deploy/docker-compose.test.yml

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
	$(MAKE) docker-down ENV=test COMPOSE_FILE=deploy/docker-compose.test.yml

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
	@bash $(ROOT)/tools/scripts/test-quality-gate.sh $(ROOT)/apps/services && \
		printf "$(C_GREEN)[PASS]$(C_RESET) test quality gate passed\n" || \
		{ printf "$(C_RED)[FAIL]$(C_RESET) test quality gate failed\n"; exit 1; }

.PHONY: docs-lint
docs-lint: ## Verify Antora documentation builds without errors
	@printf "$(C_CYAN)[Docs]$(C_RESET) verifying Antora documentation build...\n"
	@if command -v npx &>/dev/null; then \
		cd $(ROOT)/docs/antora && npx antora --fetch antora-playbook.yml 2>&1 | tail -5 || \
		printf "$(C_YELLOW)[Docs]$(C_RESET) WARN: antora build failed — check antora-playbook.yml\n"; \
	else \
		printf "$(C_YELLOW)[Docs]$(C_RESET) SKIP: npx not available\n"; \
	fi

.PHONY: npm-audit
npm-audit: ## Run npm/pnpm audit on frontend dependencies
	@printf "$(C_CYAN)[Security]$(C_RESET) auditing frontend dependencies...\n"
	@if [ -f "$(ROOT)/$(SERVICE_DIR)/frontend/package.json" ]; then \
		cd $(ROOT)/$(SERVICE_DIR)/frontend && \
		if command -v pnpm &>/dev/null; then \
			pnpm audit --audit-level=high 2>&1 || printf "$(C_YELLOW)[Security]$(C_RESET) WARN: audit found vulnerabilities\n"; \
		elif command -v npm &>/dev/null; then \
			npm audit --audit-level=high 2>&1 || printf "$(C_YELLOW)[Security]$(C_RESET) WARN: audit found vulnerabilities\n"; \
		else \
			printf "$(C_YELLOW)[Security]$(C_RESET) SKIP: neither pnpm nor npm available\n"; \
		fi; \
	else \
		echo "[Security] SKIP: frontend/package.json not found"; \
	fi

# ==============================================================================
# CI — Full pipeline aggregate target
# ==============================================================================

##@ CI

.PHONY: ci-fast

ci-fast: ## CI pipeline — fail-fast (proto + build + lint + test-fast + typecheck)
	@printf "$(C_BOLD)$(C_CYAN)=== CI Pipeline Started ===$(C_RESET)\n"
	@$(MAKE) proto-all
	@$(MAKE) build
	@$(MAKE) lint
	@$(MAKE) test-fast
	@$(MAKE) typecheck-typescript
	@printf "$(C_GREEN)=== CI pipeline passed (proto + build + lint + test-fast + typecheck) ===$(C_RESET)\n"

ci-full: ## Full CI pipeline — fail-fast (vendor-go + ci-fast + quality + integration + e2e + gates + security)
	@printf "$(C_BOLD)$(C_CYAN)=== Full CI Pipeline Started ===$(C_RESET)\n"
	@$(MAKE) vendor-go
	@$(MAKE) ci-fast
	@printf "$(C_CYAN)[ci-full]$(C_RESET) static test quality gate...\n"
	@bash $(ROOT)/tools/scripts/test-quality-gate.sh $(ROOT)/apps/services
	@$(MAKE) docs-lint
	@printf "$(C_CYAN)[ci-full]$(C_RESET) Ensuring Docker test stack is up...\n"
	@$(MAKE) docker-up-test 2>/dev/null || true
	@$(MAKE) test-integration-fast
	@$(MAKE) test-e2e-fast
	@$(MAKE) test-gates-fast
	@$(MAKE) npm-audit
	@printf "$(C_CYAN)[ci-full]$(C_RESET) saving quality trend snapshot...\n"
	@mkdir -p $(ROOT)/.quality-trends
	@bash $(ROOT)/tools/scripts/test-quality-gate.sh --json $(ROOT)/apps/services 2>/dev/null > $(ROOT)/.quality-trends/$$(date +%Y-%m-%d).json || true
	@printf "$(C_GREEN)=== Full CI pipeline passed (vendor-go + ci-fast + quality + integration + E2E + gates + security) ===$(C_RESET)\n"

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
