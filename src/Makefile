# @ctx: Root build orchestrator for VEDO Core — PLAT-LOCAL-002
# Use Git Bash on Windows, /bin/bash on Unix
ifeq ($(OS),Windows_NT)
  SHELL := C:/Program Files/Git/bin/bash.exe
else
  SHELL := /bin/bash
endif
.ONESHELL:

ROOT := $(realpath $(dir $(lastword $(MAKEFILE_LIST))))

include $(ROOT)/build/rust.mk
include $(ROOT)/build/go.mk
include $(ROOT)/build/python.mk
include $(ROOT)/build/typescript.mk
include $(ROOT)/build/docker.mk
include $(ROOT)/build/yaml.mk

.PHONY: build lint test typecheck docker-build clean docker-build-stubs

STUB_SERVICES := api-gateway auth-service ontology-service versioning-service metrics-service publisher-service public-browse-api commenting-service ticket-api ticket-classifier ticket-telemetry-listener ticket-notifier frontend publish-browse-ui

build: build-rust build-go build-python build-typescript

lint: lint-rust lint-go lint-python lint-typescript lint-yaml

test: test-rust test-go test-python test-typescript

typecheck: typecheck-typescript

docker-build: docker-build-rust docker-build-go docker-build-python docker-build-typescript

docker-build-stubs: $(addprefix docker-build-,$(STUB_SERVICES))

docker-build-%:
	@if echo "$(STUB_SERVICES)" | grep -Eq "(^| )$*( |$$)"; then \
		echo "[Docker] building stub image vedo-core/$*:latest"; \
		docker build -f "$(ROOT)/services/$*/Dockerfile" -t "vedo-core/$*:latest" "$(ROOT)"; \
	else \
		echo "BUILD_FAILED: unknown service '$*'"; \
		exit 1; \
	fi

clean: clean-rust clean-go clean-python clean-typescript
