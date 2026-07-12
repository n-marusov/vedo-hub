# @ctx: TypeScript build/lint/test rules — PLAT-LOCAL-002

TS_DIRS := $(shell cd $(ROOT) && find . -maxdepth 3 -name package.json -not -path "*/node_modules/*" -not -path "*/templates/*" -exec dirname {} \; 2>/dev/null | sed 's|^\./||')

.PHONY: build-typescript
build-typescript:
	@if [ -z "$(TS_DIRS)" ]; then echo "No TypeScript services found"; exit 0; fi
	@for dir in $(TS_DIRS); do \
		echo "[TypeScript] building $$(basename $$dir)"; \
		cd $(ROOT)/$$dir && pnpm install 2>&1 && pnpm build 2>&1 || { echo "BUILD_FAILED: pnpm build failed in $$dir"; exit 1; }; \
	done

.PHONY: lint-typescript
lint-typescript:
	@if [ -z "$(TS_DIRS)" ]; then echo "No TypeScript services found"; exit 0; fi
	@for dir in $(TS_DIRS); do \
		echo "[TypeScript] linting $$(basename $$dir)"; \
		cd $(ROOT)/$$dir && npx biome check . 2>&1 || { echo "LINT_FAILED: lint failed in $$dir"; exit 1; }; \
	done

.PHONY: test-typescript
test-typescript:
	@if [ -z "$(TS_DIRS)" ]; then echo "No TypeScript services found"; exit 0; fi
	@for dir in $(TS_DIRS); do \
		echo "[TypeScript] testing $$(basename $$dir)"; \
		cd $(ROOT)/$$dir && pnpm test 2>&1 || { echo "TEST_FAILED: pnpm test failed in $$dir"; exit 1; }; \
	done

.PHONY: typecheck-typescript
typecheck-typescript:
	@if [ -z "$(TS_DIRS)" ]; then echo "No TypeScript services found"; exit 0; fi
	@for dir in $(TS_DIRS); do \
		echo "[TypeScript] typechecking $$(basename $$dir)"; \
		cd $(ROOT)/$$dir && npx vue-tsc --noEmit 2>&1 || { echo "LINT_FAILED: vue-tsc typecheck failed in $$dir"; exit 1; }; \
	done

.PHONY: clean-typescript
clean-typescript:
	@if [ -z "$(TS_DIRS)" ]; then exit 0; fi
	@for dir in $(TS_DIRS); do \
		rm -rf $(ROOT)/$$dir/node_modules $(ROOT)/$$dir/dist; \
	done
