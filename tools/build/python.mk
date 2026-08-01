# @ctx: Python build/lint/test rules — PLAT-LOCAL-002

PYTHON_DIRS := $(shell find $(ROOT) -maxdepth 4 -name pyproject.toml -not -path "*/node_modules/*" -not -path "*/templates/*" -not -path "*/scaffolds/*" -not -path "*/src/services/*" -exec dirname {} \; 2>/dev/null | sort -u)

.PHONY: build-python
build-python:
	@if [ -z "$(PYTHON_DIRS)" ]; then echo "No Python services found"; exit 0; fi
	@for dir in $(PYTHON_DIRS); do \
		printf "$(C_CYAN)[Python]$(C_RESET) installing deps for $$(basename $$dir)\n"; \
		cd $$dir && uv sync 2>&1 || { printf "$(C_RED)BUILD_FAILED:$(C_RESET) uv sync failed in $$dir\n"; exit 1; }; \
done

.PHONY: lint-python
lint-python:
	@if [ -z "$(PYTHON_DIRS)" ]; then echo "No Python services found"; exit 0; fi
	@for dir in $(PYTHON_DIRS); do \
		printf "$(C_CYAN)[Python]$(C_RESET) linting $$(basename $$dir)\n"; \
		cd $$dir && uv run ruff check . 2>&1 || { printf "$(C_RED)LINT_FAILED:$(C_RESET) ruff check failed in $$dir\n"; exit 1; }; \
	done

.PHONY: test-python-fast test-python-full

test-python-fast: ## Python tests — fail-fast (stops at first failure)
	@if [ -z "$(PYTHON_DIRS)" ]; then echo "No Python services found"; exit 0; fi
	@for dir in $(PYTHON_DIRS); do \
		printf "$(C_CYAN)[Python]$(C_RESET) testing $$(basename $$dir)\n"; \
		cd $$dir && (uv run pytest 2>&1 || [ $$? -eq 5 ]) || { printf "$(C_RED)[FAIL]$(C_RESET) pytest failed in $$dir\n"; exit 1; }; \
	done; \
	printf "$(C_GREEN)[PASS]$(C_RESET) All Python tests passed\n"

test-python-full: ## Python tests — full statistics (collect all failures)
	@if [ -z "$(PYTHON_DIRS)" ]; then echo "No Python services found"; exit 0; fi
	@failed=""; \
	for dir in $(PYTHON_DIRS); do \
		printf "$(C_CYAN)[Python]$(C_RESET) testing $$(basename $$dir)\n"; \
		cd $$dir && (uv run pytest 2>&1 || [ $$? -eq 5 ]) || failed="$$failed $$(basename $$dir)"; \
	done; \
	if [ -n "$$failed" ]; then \
		printf "$(C_RED)[FAIL]$(C_RESET) pytest failed in:$$failed\n"; \
		exit 1; \
	fi; \
	printf "$(C_GREEN)[PASS]$(C_RESET) All Python tests passed\n"

.PHONY: clean-python
clean-python:
	@if [ -z "$(PYTHON_DIRS)" ]; then exit 0; fi
	@for dir in $(PYTHON_DIRS); do \
		cd $$dir && rm -rf __pycache__ .pytest_cache .ruff_cache *.egg-info; \
	done
