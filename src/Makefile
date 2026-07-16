# =============================================================================
# VEDO Core — Build Orchestrator
# =============================================================================
# Root Makefile for polyglot microservices monorepo.
# Delegates to language-specific includes in build/*.mk
# =============================================================================

# --- Preamble ---
.SHELLFLAGS := -eu -o pipefail -c
.DELETE_ON_ERROR:
MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules

# Use Git Bash on Windows, /bin/bash on Unix
ifeq ($(OS),Windows_NT)
  SHELL := C:/Program Files/Git/bin/bash.exe
else
  SHELL := /bin/bash
endif
.ONESHELL:

# --- Project Metadata ---
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ' 2>/dev/null || echo "unknown")

ROOT := $(realpath $(dir $(lastword $(MAKEFILE_LIST))))

include $(ROOT)/build/rust.mk
include $(ROOT)/build/go.mk
include $(ROOT)/build/python.mk
include $(ROOT)/build/typescript.mk
include $(ROOT)/build/docker.mk
include $(ROOT)/build/yaml.mk

.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help message
	@awk 'BEGIN {FS = ":.*##"; printf "Usage: make \\033[36m<target>\\033[0m\\n"} \
		/^[a-zA-Z_-]+:.*## / {printf "  \\033[36m%-22s\\033[0m %s\\n", $$1, $$2} \
		/^##@/ {printf "\\n\\033[1m%s\\033[0m\\n", substr($$0, 5)}' $(MAKEFILE_LIST)

STUB_SERVICES := api-gateway auth-service ontology-service versioning-service metrics-service publisher-service public-browse-api commenting-service ticket-api ticket-classifier ticket-telemetry-listener ticket-notifier frontend publish-browse-ui


##@ Proto — gRPC code generation

PROTO_DIR := $(ROOT)/services/shared/proto

.PHONY: proto-generate
proto-generate: ## Generate gRPC code from .proto files (Go + Rust)
	@echo "[Proto] generating Go code..."
	@cd $(PROTO_DIR) && buf generate
	@echo "[Proto] Go code generated"
	@echo "[Proto] Rust code is generated at build-time via tonic_build"

.PHONY: proto-lint
proto-lint: ## Lint .proto files with buf
	@echo "[Proto] linting..."
	@cd $(PROTO_DIR) && buf lint

.PHONY: proto-breaking
proto-breaking: ## Check .proto breaking changes with buf
	@echo "[Proto] checking breaking changes..."
	@cd $(PROTO_DIR) && buf breaking --against '.git#branch=main'

.PHONY: proto-all
proto-all: proto-lint proto-generate ## Run all proto checks (lint + generate)

##@ Build
build: build-rust build-go build-python build-typescript ## Build all services (Rust, Go, Python, TypeScript)

##@ Lint
lint: lint-rust lint-go lint-python lint-typescript lint-yaml ## Run all linters

##@ Format
.PHONY: fmt-rust fmt-go fmt-python fmt-typescript fmt

fmt-rust:
	@if [ -z "$(RUST_DIRS)" ]; then echo "No Rust services found"; exit 0; fi
	@for dir in $(RUST_DIRS); do \
		echo "[Rust] formatting $$(basename $$dir)"; \
		cd $$dir && cargo fmt 2>&1 || true; \
	done

fmt-go:
	@if [ -z "$(GO_DIRS)" ]; then echo "No Go services found"; exit 0; fi
	@for dir in $(GO_DIRS); do \
		echo "[Go] formatting $$(basename $$dir)"; \
		cd $(ROOT)/$$dir && gofmt -l -w . 2>&1 || true; \
	done

fmt-python:
	@if [ -z "$(PYTHON_DIRS)" ]; then echo "No Python services found"; exit 0; fi
	@for dir in $(PYTHON_DIRS); do \
		echo "[Python] formatting $$(basename $$dir)"; \
		cd $$dir && ruff format . 2>&1 || true; \
	done

fmt-typescript:
	@if [ -z "$(TS_DIRS)" ]; then echo "No TypeScript services found"; exit 0; fi
	@for dir in $(TS_DIRS); do \
		echo "[TypeScript] formatting $$(basename $$dir)"; \
		cd $(ROOT)/$$dir && npx biome check --write . 2>&1 || true; \
	done

fmt: fmt-rust fmt-go fmt-python fmt-typescript ## Format all source code

##@ Test
test: test-rust test-go test-python test-typescript ## Run all tests

.PHONY: coverage
coverage: ## Run tests with coverage
	@echo "[Coverage] running tests with coverage..."
	@if [ -n "$(GO_DIRS)" ]; then \
		for dir in $(GO_DIRS); do \
			echo "[Go] coverage for $$(basename $$dir)"; \
			cd $(ROOT)/$$dir && go test -coverprofile=cover.out ./... 2>&1 || true; \
		done; \
	fi

##@ TypeScript
typecheck: typecheck-typescript ## Run TypeScript type checking

##@ CI
ci: build lint test typecheck ## Run all CI checks (build + lint + test + typecheck)

##@ Docker — Build
docker-build: docker-build-rust docker-build-go docker-build-python docker-build-typescript ## Build Docker images for all services

docker-build-stubs: $(addprefix docker-build-,$(STUB_SERVICES)) ## Build stub service images

docker-build-%:
	@if echo "$(STUB_SERVICES)" | grep -Eq "(^| )$*( |$$)"; then \
		echo "[Docker] building stub image vedo-core/$*:latest"; \
		docker build -f "$(ROOT)/services/$*/Dockerfile" -t "vedo-core/$*:latest" "$(ROOT)"; \
	else \
		echo "BUILD_FAILED: unknown service '$*'"; \
		exit 1; \
	fi

##@ Docker — Compose Lifecycle
COMPOSE_FILE ?= ../deploy/docker-compose.yml
COMPOSE_PROFILE ?=

.PHONY: docker-up docker-down docker-logs docker-ps docker-shell

docker-up: ## Start all services via Docker Compose
	docker compose -f $(COMPOSE_FILE) $(if $(COMPOSE_PROFILE),--profile $(COMPOSE_PROFILE)) up -d

docker-down: ## Stop all services
	docker compose -f $(COMPOSE_FILE) down

docker-logs: ## Tail logs from all services
	docker compose -f $(COMPOSE_FILE) logs -f

docker-ps: ## List running service containers
	docker compose -f $(COMPOSE_FILE) ps

docker-shell: ## Open shell in a service container (usage: make docker-shell SVC=<name>)
	@if [ -z "$(SVC)" ]; then echo "Usage: make docker-shell SVC=<service-name>"; exit 1; fi
	docker compose -f $(COMPOSE_FILE) exec $(SVC) sh

##@ Docker — Status
.PHONY: docker-status docker-status-full

docker-status: ## Brief health table (Name, Health, Ports)
	@COMPOSE_FILE="$(COMPOSE_FILE)"
	docker compose -f "$$COMPOSE_FILE" ps --format "table {{.Name}}\t{{.Health}}\t{{.Ports}}" 2>/dev/null

docker-status-full: ## Full machine-readable JSON with all container details
	@COMPOSE_FILE="$(COMPOSE_FILE)"
	docker compose -f "$$COMPOSE_FILE" ps --format json 2>/dev/null || echo '[]'

##@ Development
.PHONY: dev dev-frontend dev-api

dev: ## Start development environment (hint: use docker-up or run services individually)
	@echo "Use 'make docker-up' to start all services via Docker Compose."
	@echo "Or run individual services:"
	@echo "  make dev-frontend  - Start frontend dev server"
	@echo "  make dev-api       - Start API gateway"

dev-frontend: ## Start frontend dev server with hot reload
	cd $(ROOT)/services/frontend && pnpm dev

dev-api: ## Start API gateway in development mode
	cd $(ROOT)/services/api-gateway && go run .

##@ Cleanup
clean: clean-rust clean-go clean-python clean-typescript ## Clean all build artifacts
