# @ctx: Docker build rules — PLAT-LOCAL-002, DEPLOY-OPS-001
# Builds Docker images for all service types using multi-stage Dockerfiles
# Images tagged: vedo-core/<service>:latest

-include $(ROOT)/.env
export

DOCKER_DIR := $(ROOT)/docker

# Build context: project root (aligns with docker-compose.yml)
PROJECT_ROOT := $(realpath $(ROOT)/..)

# ── Docker image defaults ────────────────────────────────────────────────────────
# Each variable matches the corresponding ARG in the language-specific Dockerfile.
# Override via .env or environment variables (e.g. RUST_BASE_IMAGE=rust:1.88).

CARGO_CHEF_IMAGE   ?= rust:1.89
RUST_BASE_IMAGE    ?= rust:1.89
RUST_RUNTIME_IMAGE ?= debian:bookworm-slim

GO_BASE_IMAGE    ?= golang:1.24-alpine
GO_RUNTIME_IMAGE ?= alpine:3.19

PYTHON_IMAGE          ?= python:3.12-slim
PYTHON_RUNTIME_IMAGE ?= python:3.12-slim

NODE_IMAGE  ?= node:20-alpine
NGINX_IMAGE ?= nginx:1.25-alpine

.PHONY: docker-build-rust
docker-build-rust:
	@if [ -z "$(RUST_DIRS)" ]; then echo "No Rust services to dockerize"; exit 0; fi
	@for dir in $(RUST_DIRS); do \
		svc=$$(basename $$dir); \
		if [ "$$svc" = "shared" ]; then continue; fi; \
		echo "[Docker] building Rust image: vedo-core/$$svc:latest"; \
		docker build \
			--build-arg BINARY_NAME=$$svc \
			--build-arg SERVICE_DIR=src/services \
			--build-arg CARGO_CHEF_IMAGE=$(CARGO_CHEF_IMAGE) \
			--build-arg RUST_BASE_IMAGE=$(RUST_BASE_IMAGE) \
			--build-arg RUST_RUNTIME_IMAGE=$(RUST_RUNTIME_IMAGE) \
			-f $(DOCKER_DIR)/Dockerfile.rust \
			-t vedo-core/$$svc:latest \
			$(PROJECT_ROOT) 2>&1 || { echo "BUILD_FAILED: docker build for $$svc"; exit 1; }; \
	done

.PHONY: docker-build-go
docker-build-go:
	@if [ -z "$(GO_DIRS)" ]; then echo "No Go services to dockerize"; exit 0; fi
	@for dir in $(GO_DIRS); do \
		svc=$$(basename $$dir); \
		case "$$dir" in cli|services/*/*) continue;; esac; \
		if [ ! -d "$(ROOT)/$$dir/vendor" ]; then \
			echo "[Docker] skipping $$svc — no vendor directory (run 'make vendor-go' first)"; \
			continue; \
		fi; \
		echo "[Docker] building Go image: vedo-core/$$svc:latest"; \
		docker build \
			--build-arg BINARY_NAME=$$svc \
			--build-arg SERVICE_DIR=src/services \
			--build-arg GO_BASE_IMAGE=$(GO_BASE_IMAGE) \
			--build-arg GO_RUNTIME_IMAGE=$(GO_RUNTIME_IMAGE) \
			-f $(DOCKER_DIR)/Dockerfile.go \
			-t vedo-core/$$svc:latest \
			$(PROJECT_ROOT) 2>&1 || { echo "BUILD_FAILED: docker build for $$svc"; exit 1; }; \
	done

.PHONY: docker-build-python
docker-build-python:
	@if [ -z "$(PYTHON_DIRS)" ]; then echo "No Python services to dockerize"; exit 0; fi
	@for dir in $(PYTHON_DIRS); do \
		svc=$$(basename $$dir); \
		rel_dir="src/services/$$svc" && \
		echo "[Docker] building Python image: vedo-core/$$svc:latest"; \
		docker build \
			--build-arg SERVICE_DIR=$$rel_dir \
			--build-arg PYTHON_IMAGE=$(PYTHON_IMAGE) \
			--build-arg PYTHON_RUNTIME_IMAGE=$(PYTHON_RUNTIME_IMAGE) \
			-f $(DOCKER_DIR)/Dockerfile.python \
			-t vedo-core/$$svc:latest \
			$(PROJECT_ROOT) 2>&1 || { echo "BUILD_FAILED: docker build for $$svc"; exit 1; }; \
	done

.PHONY: docker-build-typescript
docker-build-typescript:
	@if [ -z "$(TS_DIRS)" ]; then echo "No TypeScript services to dockerize"; exit 0; fi
	@for dir in $(TS_DIRS); do \
		svc=$$(basename $$dir); \
		echo "[Docker] building TypeScript image: vedo-core/$$svc:latest"; \
		docker build \
			--build-arg SERVICE_DIR=src/$$dir \
			--build-arg NODE_IMAGE=$(NODE_IMAGE) \
			--build-arg NGINX_IMAGE=$(NGINX_IMAGE) \
			-f $(DOCKER_DIR)/Dockerfile.typescript \
			-t vedo-core/$$svc:latest \
			$(PROJECT_ROOT) 2>&1 || { echo "BUILD_FAILED: docker build for $$svc"; exit 1; }; \
	done
