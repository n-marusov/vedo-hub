# @ctx: TypeScript build/lint/test rules — PLAT-LOCAL-002

TS_DIRS := $(shell cd $(ROOT) && find . -maxdepth 5 -name package.json -not -path "*/node_modules/*" -not -path "*/templates/*" -not -path "*/scaffolds/*" -not -path "*/src/services/*" -not -path "*/tests/*" -exec dirname {} \; 2>/dev/null | sed 's|^\./||')

.PHONY: build-typescript
build-typescript:
	@if [ -z "$(TS_DIRS)" ]; then echo "No TypeScript services found"; exit 0; fi
	@for dir in $(TS_DIRS); do \
		printf "$(C_CYAN)[TypeScript]$(C_RESET) building $$(basename $$dir)\n"; \
		cd $(ROOT)/$$dir && CI=true pnpm install 2>&1 && CI=true pnpm build 2>&1 || { printf "$(C_RED)BUILD_FAILED:$(C_RESET) pnpm build failed in $$dir\n"; exit 1; }; \
	done

.PHONY: lint-typescript
lint-typescript:
	@if [ -z "$(TS_DIRS)" ]; then echo "No TypeScript services found"; exit 0; fi
	@for dir in $(TS_DIRS); do \
		printf "$(C_CYAN)[TypeScript]$(C_RESET) linting $$(basename $$dir)\n"; \
		cd $(ROOT)/$$dir && npx biome check . 2>&1 || { printf "$(C_RED)LINT_FAILED:$(C_RESET) lint failed in $$dir\n"; exit 1; }; \
	done

.PHONY: test-typescript-fast test-typescript-full

test-typescript-fast: ## TypeScript tests — fail-fast (stops at first failure)
	@if [ -z "$(TS_DIRS)" ]; then echo "No TypeScript services found"; exit 0; fi
	@for dir in $(TS_DIRS); do \
		printf "$(C_CYAN)[TypeScript]$(C_RESET) testing $$(basename $$dir)\n"; \
		cd $(ROOT)/$$dir && pnpm test 2>&1 || { printf "$(C_RED)[FAIL]$(C_RESET) pnpm test failed in $$dir\n"; exit 1; }; \
	done; \
	printf "$(C_GREEN)[PASS]$(C_RESET) All TypeScript tests passed\n"

test-typescript-full: ## TypeScript tests — full statistics (collect all failures)
	@if [ -z "$(TS_DIRS)" ]; then echo "No TypeScript services found"; exit 0; fi
	@failed=""; \
	for dir in $(TS_DIRS); do \
		printf "$(C_CYAN)[TypeScript]$(C_RESET) testing $$(basename $$dir)\n"; \
		cd $(ROOT)/$$dir && pnpm test 2>&1 || failed="$$failed $$(basename $$dir)"; \
	done; \
	if [ -n "$$failed" ]; then \
		printf "$(C_RED)[FAIL]$(C_RESET) pnpm test failed in:$$failed\n"; \
		exit 1; \
	fi; \
	printf "$(C_GREEN)[PASS]$(C_RESET) All TypeScript tests passed\n"

.PHONY: typecheck-typescript
codegen-check:
	@if [ -z "$(TS_DIRS)" ]; then echo "No TypeScript services found"; exit 0; fi
	@for dir in $(TS_DIRS); do \
		if [ -f "$(ROOT)/$$dir/codegen.ts" ]; then \
			echo "[TypeScript] codegen check $$(basename $$dir)"; \
			cd $(ROOT)/$$dir && pnpm codegen:check 2>&1 || { echo "CODEGEN_FAILED: generated GraphQL types out of date in $$dir (run pnpm codegen)"; exit 1; }; \
		fi; \
	done

typecheck-typescript: codegen-check
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
