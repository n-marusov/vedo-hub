# @ctx: Docker build rules — PLAT-LOCAL-002, DEPLOY-OPS-001
# Builds Docker images for all service types using multi-stage Dockerfiles
# Images tagged: vedo-core/<service>:latest

-include $(ROOT)/.env
export

DOCKER_DIR := $(ROOT)/docker

.PHONY: docker-build-rust
docker-build-rust:
	@if [ -z "$(RUST_DIRS)" ]; then echo "No Rust services to dockerize"; exit 0; fi
	@for dir in $(RUST_DIRS); do \
		svc=$$(basename $$dir); \
		echo "[Docker] building Rust image: vedo-core/$$svc:latest"; \
		docker build \
			--build-arg BINARY_NAME=$$svc \
			--build-arg SERVICE_DIR=$$dir \
			--build-arg CARGO_CHEF_IMAGE=$(CARGO_CHEF_IMAGE) \
			--build-arg RUST_BASE_IMAGE=$(RUST_BASE_IMAGE) \
			--build-arg RUST_RUNTIME_IMAGE=$(RUST_RUNTIME_IMAGE) \
			-f $(DOCKER_DIR)/Dockerfile.rust \
			-t vedo-core/$$svc:latest \
			. 2>&1 || { echo "BUILD_FAILED: docker build for $$svc"; exit 1; }; \
	done

.PHONY: docker-build-go
docker-build-go:
	@if [ -z "$(GO_DIRS)" ]; then echo "No Go services to dockerize"; exit 0; fi
	@for dir in $(GO_DIRS); do \
		svc=$$(basename $$dir); \
		echo "[Docker] building Go image: vedo-core/$$svc:latest"; \
		docker build \
			--build-arg BINARY_NAME=$$svc \
			--build-arg SERVICE_DIR=$$dir \
			--build-arg GO_BASE_IMAGE=$(GO_BASE_IMAGE) \
			--build-arg GO_RUNTIME_IMAGE=$(GO_RUNTIME_IMAGE) \
			-f $(DOCKER_DIR)/Dockerfile.go \
			-t vedo-core/$$svc:latest \
			. 2>&1 || { echo "BUILD_FAILED: docker build for $$svc"; exit 1; }; \
	done

.PHONY: docker-build-python
docker-build-python:
	@if [ -z "$(PYTHON_DIRS)" ]; then echo "No Python services to dockerize"; exit 0; fi
	@for dir in $(PYTHON_DIRS); do \
		svc=$$(basename $$dir); \
		echo "[Docker] building Python image: vedo-core/$$svc:latest"; \
		docker build \
			--build-arg SERVICE_DIR=$$dir \
			--build-arg PYTHON_BASE_IMAGE=$(PYTHON_BASE_IMAGE) \
			--build-arg PYTHON_RUNTIME_IMAGE=$(PYTHON_RUNTIME_IMAGE) \
			-f $(DOCKER_DIR)/Dockerfile.python \
			-t vedo-core/$$svc:latest \
			. 2>&1 || { echo "BUILD_FAILED: docker build for $$svc"; exit 1; }; \
	done

.PHONY: docker-build-typescript
docker-build-typescript:
	@if [ -z "$(TS_DIRS)" ]; then echo "No TypeScript services to dockerize"; exit 0; fi
	@for dir in $(TS_DIRS); do \
		svc=$$(basename $$dir); \
		echo "[Docker] building TypeScript image: vedo-core/$$svc:latest"; \
		docker build \
			--build-arg SERVICE_DIR=$$dir \
			--build-arg NODE_BASE_IMAGE=$(NODE_BASE_IMAGE) \
			--build-arg NGINX_IMAGE=$(NGINX_IMAGE) \
			-f $(DOCKER_DIR)/Dockerfile.typescript \
			-t vedo-core/$$svc:latest \
			. 2>&1 || { echo "BUILD_FAILED: docker build for $$svc"; exit 1; }; \
	done
