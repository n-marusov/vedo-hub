# @ctx: Python build/lint/test rules — PLAT-LOCAL-002

PYTHON_DIRS := $(shell find $(ROOT) -maxdepth 4 -name pyproject.toml -not -path "*/node_modules/*" -not -path "*/templates/*" -exec dirname {} \; 2>/dev/null | sort -u)

.PHONY: build-python
build-python:
	@if [ -z "$(PYTHON_DIRS)" ]; then echo "No Python services found"; exit 0; fi
	@for dir in $(PYTHON_DIRS); do \
		echo "[Python] installing deps for $$(basename $$dir)"; \
		cd $$dir && pip install -e . 2>&1 || { echo "BUILD_FAILED: pip install failed in $$dir"; exit 1; }; \
	done

.PHONY: lint-python
lint-python:
	@if [ -z "$(PYTHON_DIRS)" ]; then echo "No Python services found"; exit 0; fi
	@for dir in $(PYTHON_DIRS); do \
		echo "[Python] linting $$(basename $$dir)"; \
		cd $$dir && ruff check . 2>&1 || { echo "LINT_FAILED: ruff check failed in $$dir"; exit 1; }; \
	done

.PHONY: test-python
test-python:
	@if [ -z "$(PYTHON_DIRS)" ]; then echo "No Python services found"; exit 0; fi
	@for dir in $(PYTHON_DIRS); do \
		echo "[Python] testing $$(basename $$dir)"; \
		cd $$dir && python -m pytest 2>&1 || { echo "TEST_FAILED: pytest failed in $$dir"; exit 1; }; \
	done

.PHONY: clean-python
clean-python:
	@if [ -z "$(PYTHON_DIRS)" ]; then exit 0; fi
	@for dir in $(PYTHON_DIRS); do \
		cd $$dir && rm -rf __pycache__ .pytest_cache .ruff_cache *.egg-info; \
	done
