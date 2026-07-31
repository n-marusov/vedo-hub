# @ctx: Rust build/lint/test rules — PLAT-LOCAL-002

RUST_DIRS := $(shell find $(ROOT) -maxdepth 4 -name Cargo.toml -not -path "*/target/*" -not -path "*/templates/*" -not -path "*/scaffolds/*" -not -path "*/src/services/*" -not -path "*/services/Cargo.toml" -exec dirname {} \; 2>/dev/null)

.PHONY: build-rust
build-rust:
	@if [ -z "$(RUST_DIRS)" ]; then echo "No Rust services found"; exit 0; fi
	@printf "$(C_CYAN)[Rust]$(C_RESET) Proto compilation is automatic via tonic_build/build.rs\n"
	@for dir in $(RUST_DIRS); do \
		printf "$(C_CYAN)[Rust]$(C_RESET) building $$(basename $$dir)\n"; \
		cd $$dir && cargo build --release 2>&1 || { printf "$(C_RED)BUILD_FAILED:$(C_RESET) cargo build failed in $$dir\n"; exit 1; }; \
	done

.PHONY: lint-rust
lint-rust:
	@if [ -z "$(RUST_DIRS)" ]; then echo "No Rust services found"; exit 0; fi
	@for dir in $(RUST_DIRS); do \
		printf "$(C_CYAN)[Rust]$(C_RESET) linting $$(basename $$dir)\n"; \
		cd $$dir && cargo clippy -- -D warnings 2>&1 || { printf "$(C_RED)LINT_FAILED:$(C_RESET) cargo clippy failed in $$dir\n"; exit 1; }; \
	done

.PHONY: test-rust-fast test-rust-fast-unit test-rust-full test-rust-full-unit

test-rust-fast: ## Rust tests — fail-fast (cargo test, stops at first failure)
	@if [ -z "$(RUST_DIRS)" ]; then echo "No Rust services found"; exit 0; fi
	@for dir in $(RUST_DIRS); do \
		printf "$(C_CYAN)[Rust]$(C_RESET) testing $$(basename $$dir)\n"; \
		cd $$dir && cargo test 2>&1 || { printf "$(C_RED)[FAIL]$(C_RESET) cargo test failed in $$dir\n"; exit 1; }; \
	done; \
	printf "$(C_GREEN)[PASS]$(C_RESET) All Rust tests passed\n"

test-rust-fast-unit: ## Rust unit tests — fail-fast (cargo test --lib, stops at first failure)
	@if [ -z "$(RUST_DIRS)" ]; then echo "No Rust services found"; exit 0; fi
	@for dir in $(RUST_DIRS); do \
		printf "$(C_CYAN)[Rust]$(C_RESET) unit testing $$(basename $$dir)\n"; \
		cd $$dir && cargo test --lib 2>&1 || { printf "$(C_RED)[FAIL]$(C_RESET) cargo test --lib failed in $$dir\n"; exit 1; }; \
	done; \
	printf "$(C_GREEN)[PASS]$(C_RESET) All Rust unit tests passed\n"

test-rust-full: ## Rust tests — full statistics (collect all failures)
	@if [ -z "$(RUST_DIRS)" ]; then echo "No Rust services found"; exit 0; fi
	@failed=""; \
	for dir in $(RUST_DIRS); do \
		printf "$(C_CYAN)[Rust]$(C_RESET) testing $$(basename $$dir)\n"; \
		cd $$dir && cargo test 2>&1 || failed="$$failed $$(basename $$dir)"; \
	done; \
	if [ -n "$$failed" ]; then \
		printf "$(C_RED)[FAIL]$(C_RESET) cargo test failed in:$$failed\n"; \
		exit 1; \
	fi; \
	printf "$(C_GREEN)[PASS]$(C_RESET) All Rust tests passed\n"

test-rust-full-unit: ## Rust unit tests — full statistics (collect all failures)
	@if [ -z "$(RUST_DIRS)" ]; then echo "No Rust services found"; exit 0; fi
	@failed=""; \
	for dir in $(RUST_DIRS); do \
		printf "$(C_CYAN)[Rust]$(C_RESET) unit testing $$(basename $$dir)\n"; \
		cd $$dir && cargo test --lib 2>&1 || failed="$$failed $$(basename $$dir)"; \
	done; \
	if [ -n "$$failed" ]; then \
		printf "$(C_RED)[FAIL]$(C_RESET) rust unit tests failed in:$$failed\n"; \
		exit 1; \
	fi; \
	printf "$(C_GREEN)[PASS]$(C_RESET) All Rust unit tests passed\n"

.PHONY: clean-rust
clean-rust:
	@if [ -z "$(RUST_DIRS)" ]; then exit 0; fi
	@for dir in $(RUST_DIRS); do cd $$dir && cargo clean; done
