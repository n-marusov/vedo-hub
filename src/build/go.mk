# @ctx: Go build/lint/test rules — PLAT-LOCAL-002

GO_DIRS := $(shell cd $(ROOT) && find . -maxdepth 3 -name go.mod -not -path "*/templates/*" -exec dirname {} \; 2>/dev/null | sed 's|^\./||')

.PHONY: build-go
build-go:
	@if [ -z "$(GO_DIRS)" ]; then echo "No Go services found"; exit 0; fi
	@for dir in $(GO_DIRS); do \
		echo "[Go] building $$(basename $$dir)"; \
		cd $(ROOT)/$$dir && go build ./... 2>&1 || { echo "BUILD_FAILED: go build failed in $$dir"; exit 1; }; \
	done

.PHONY: lint-go
lint-go:
	@if [ -z "$(GO_DIRS)" ]; then echo "No Go services found"; exit 0; fi
	@for dir in $(GO_DIRS); do \
		echo "[Go] linting $$(basename $$dir)"; \
		cd $(ROOT)/$$dir && golangci-lint run ./... 2>&1 || { echo "LINT_FAILED: golangci-lint failed in $$dir"; exit 1; }; \
	done

.PHONY: test-go
test-go:
	@if [ -z "$(GO_DIRS)" ]; then echo "No Go services found"; exit 0; fi
	@for dir in $(GO_DIRS); do \
		echo "[Go] testing $$(basename $$dir)"; \
		cd $(ROOT)/$$dir && go test ./... 2>&1 || { echo "TEST_FAILED: go test failed in $$dir"; exit 1; }; \
	done

.PHONY: clean-go
clean-go:
	@if [ -z "$(GO_DIRS)" ]; then exit 0; fi
	@for dir in $(GO_DIRS); do cd $(ROOT)/$$dir && go clean; done
