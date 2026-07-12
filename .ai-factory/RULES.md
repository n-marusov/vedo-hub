# Project Rules

> This file declares top-level project rules, constraints, and axioms that guide AI agent behavior.
> More specific rules by area (API, frontend, backend, database) may be created in `.ai-factory/rules/`.

## Working with the Repository

1. **Shell commands must be decomposed** — do not chain `cd` with `&&`. Use separate steps.
2. **Read before edit** — understand full context before modifying files.
3. **Run tests after changes** — verify the relevant test suite passes.
4. **Use `$aif-plan` for new features** — plan before implementing.
5. **Use `$aif-fix` for bugs** — structured bug fix workflow.
6. **Keep AGENTS.md updated** — structural changes require AGENTS.md updates.

## Code Quality

1. **Structured logging** — all services must use structured JSON logging with trace context.
2. **Health endpoints** — every service must expose `/health` and `/ready` endpoints.
3. **OpenTelemetry** — trace context must propagate across all inter-service calls.
4. **Error codes** — use domain-prefixed error codes (e.g., `CLI-TICKET-NOT-FOUND`).
5. **No stub markers** — production code must not contain `[STUB]` prefixes or stubbed versions.

## Reference Files

- `.ai-factory/DESCRIPTION.md` — Project specification
- `.ai-factory/ARCHITECTURE.md` — Architecture guidelines
- `.ai-factory/rules/base.md` — Codebase conventions
- `.ai-factory/traceability/traceability.ttl` — Artifact traceability graph (RDF Turtle)
- `AGENTS.md` — Project structure map for AI agents

## Traceability

1. **Keep traceability.ttl up to date** — every new or modified artifact (service, command, test suite, security boundary, error code, deployment config) must be reflected in `.ai-factory/traceability/traceability.ttl`.
2. **Always run change impact analysis via traceability.ttl** — before modifying any artifact, query the traceability graph to identify all dependent artifacts that may be affected.
3. **Use the VEDO ontology vocabulary** — all traceability relationships (implements, validates, satisfies, constrains, deploys, monitors, affectedBy) must use the vdo: prefix defined in traceability.ttl.

## Documentation

1. **User-facing documentation must use Antora** — all user, developer, administrator, and operator documentation must be authored in AsciiDoc format under `src/docs/antora/` and organized as Antora component modules. This documentation is intended for publication via Antora site generation.
2. **System specification is maintained in the `specs/` submodule** — the `specs/` git submodule contains requirements, ADRs, use cases, user stories, and architectural specifications. This is the authoritative source for system design and must be consulted during development. Changes to specifications require updating the submodule separately.
