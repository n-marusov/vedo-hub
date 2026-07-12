# @ctx: YAML linting rules — PLAT-LOCAL-002
# Validates all YAML files in the repository using yamllint

YAML_ROOT := $(ROOT)/../..

.PHONY: lint-yaml
lint-yaml:
	@echo "[YAML] linting all YAML files"
	@cd $(YAML_ROOT) && yamllint -c .yamllint .
