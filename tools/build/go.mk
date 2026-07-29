# @ctx: Go build/lint/test rules — PLAT-LOCAL-002

GO_DIRS := $(shell cd $(ROOT) && find . -maxdepth 5 -name go.mod -not -path "*/templates/*" -not -path "*/scaffolds/*" -not -path "*/src/*" -not -path "*/tests/*" -not -path "*/shared/proto/*" -exec dirname {} \; 2>/dev/null | sed 's|^\./||')

.PHONY: build-go
build-go:
	@if [ -z "$(GO_DIRS)" ]; then echo "No Go services found"; exit 0; fi
	@printf "$(C_CYAN)[Go]$(C_RESET) Proto generation: run 'make proto-generate' first\n"
	@for dir in $(GO_DIRS); do \
		printf "$(C_CYAN)[Go]$(C_RESET) building $$(basename $$dir)\n"; \
		cd $(ROOT)/$$dir && go build ./... 2>&1 || { printf "$(C_RED)BUILD_FAILED:$(C_RESET) go build failed in $$dir\n"; exit 1; }; \
	done

.PHONY: lint-go
lint-go:
	@if [ -z "$(GO_DIRS)" ]; then echo "No Go services found"; exit 0; fi
	@for dir in $(GO_DIRS); do \
		printf "$(C_CYAN)[Go]$(C_RESET) linting $$(basename $$dir)\n"; \
		cd $(ROOT)/$$dir && golangci-lint run ./... 2>&1 || { printf "$(C_RED)LINT_FAILED:$(C_RESET) golangci-lint failed in $$dir\n"; exit 1; }; \
	done

.PHONY: test-go-fast test-go-full

test-go-fast: ## Go tests — fail-fast (go test, stops at first failure)
	@if [ -z "$(GO_DIRS)" ]; then echo "No Go services found"; exit 0; fi
	@for dir in $(GO_DIRS); do \
		printf "$(C_CYAN)[Go]$(C_RESET) testing $$(basename $$dir)\n"; \
		cd $(ROOT)/$$dir && go test ./... 2>&1; \
	done

test-go-full: ## Go tests — full statistics (collect all failures)
	@if [ -z "$(GO_DIRS)" ]; then echo "No Go services found"; exit 0; fi
	@failed=""; \
	for dir in $(GO_DIRS); do \
		printf "$(C_CYAN)[Go]$(C_RESET) testing $$(basename $$dir)\n"; \
		cd $(ROOT)/$$dir && go test ./... 2>&1 || failed="$$failed $$(basename $$dir)"; \
	done; \
	if [ -n "$$failed" ]; then \
		printf "$(C_RED)TEST_FAILED:$(C_RESET) go test failed in:$$failed\n"; \
		exit 1; \
	fi

.PHONY: vendor-go
vendor-go: ## Populate vendor directories for all Go services (enables offline Docker builds)
	@if [ -z "$(GO_DIRS)" ]; then echo "No Go services found"; exit 0; fi
	@echo "[Go] populating vendor directories..."
	@for dir in $(GO_DIRS); do \
		echo "[Go] vendoring $$(basename $$dir)"; \
		cd $(ROOT)/$$dir && go mod vendor 2>&1 || echo "WARN: go mod vendor failed in $$dir"; \
	done
	@echo "[Go] vendor directories populated"

.PHONY: clean-go
clean-go:
	@if [ -z "$(GO_DIRS)" ]; then exit 0; fi
	@for dir in $(GO_DIRS); do cd $(ROOT)/$$dir && go clean; done
	@echo "[Go] to remove vendor directories, run: rm -rf $$(find $(ROOT) -type d -name vendor | grep -v node_modules)"
