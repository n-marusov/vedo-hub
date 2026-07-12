# @ctx: Rust build/lint/test rules — PLAT-LOCAL-002

RUST_DIRS := $(shell find $(ROOT) -maxdepth 2 -name Cargo.toml -not -path "*/target/*" -exec dirname {} \; 2>/dev/null)

.PHONY: build-rust
build-rust:
	@if [ -z "$(RUST_DIRS)" ]; then echo "No Rust services found"; exit 0; fi
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
	@for dir in $(RUST_DIRS); do \
		echo "[Rust] testing $$(basename $$dir)"; \
		cd $$dir && cargo test 2>&1 || { echo "TEST_FAILED: cargo test failed in $$dir"; exit 1; }; \
	done

.PHONY: clean-rust
clean-rust:
	@if [ -z "$(RUST_DIRS)" ]; then exit 0; fi
	@for dir in $(RUST_DIRS); do cd $$dir && cargo clean; done
