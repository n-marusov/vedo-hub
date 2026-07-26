# @ctx: Rust build/lint/test rules — PLAT-LOCAL-002

RUST_DIRS := $(shell find $(ROOT) -maxdepth 4 -name Cargo.toml -not -path "*/target/*" -not -path "*/templates/*" -not -path "*/scaffolds/*" -not -path "*/src/services/*" -not -path "*/services/Cargo.toml" -exec dirname {} \; 2>/dev/null)

.PHONY: build-rust
build-rust:
	@if [ -z "$(RUST_DIRS)" ]; then echo "No Rust services found"; exit 0; fi
	@echo "[Rust] Proto compilation is automatic via tonic_build/build.rs"
	@for dir in $(RUST_DIRS); do \
		echo "[Rust] building $$(basename $$dir)"; \
		cd $$dir && cargo build --release 2>&1 || { echo "BUILD_FAILED: cargo build failed in $$dir"; exit 1; }; \
	done

.PHONY: lint-rust
lint-rust:
	@if [ -z "$(RUST_DIRS)" ]; then echo "No Rust services found"; exit 0; fi
	@for dir in $(RUST_DIRS); do \
		echo "[Rust] linting $$(basename $$dir)"; \
		cd $$dir && cargo clippy -- -D warnings 2>&1 || { echo "LINT_FAILED: cargo clippy failed in $$dir"; exit 1; }; \
	done

.PHONY: test-rust
test-rust:
	@if [ -z "$(RUST_DIRS)" ]; then echo "No Rust services found"; exit 0; fi
	@failed=""; \
	for dir in $(RUST_DIRS); do \
		echo "[Rust] testing $$(basename $$dir)"; \
		cd $$dir && cargo test 2>&1 || failed="$$failed $$(basename $$dir)"; \
	done; \
	if [ -n "$$failed" ]; then echo "TEST_FAILED: cargo test failed in:$$failed"; exit 1; fi

.PHONY: test-rust-unit
test-rust-unit:
	@if [ -z "$(RUST_DIRS)" ]; then echo "No Rust services found"; exit 0; fi
	@failed=""; \
	for dir in $(RUST_DIRS); do \
		echo "[Rust] unit testing $$(basename $$dir)"; \
		cd $$dir && cargo test --lib 2>&1 || failed="$$failed $$(basename $$dir)"; \
	done; \
	if [ -n "$$failed" ]; then echo "TEST_FAILED: rust unit tests failed in:$$failed"; exit 1; fi

.PHONY: clean-rust
clean-rust:
	@if [ -z "$(RUST_DIRS)" ]; then exit 0; fi
	@for dir in $(RUST_DIRS); do cd $$dir && cargo clean; done
