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
6. **Docker quality gate** — the quality gate must verify Docker container builds, docker-compose startup, and health status of all services.

## Domain Terminology

1. **Follow the glossary for naming** — when naming domain entities, types, fields, error codes, and API surfaces in code, use the canonical terms defined in `specs/glossary.md` (e.g., `Ontology`, `Project`, `Group`, `Commit`, `Branch`, `MergeRequest`, `Triple`, `Class`); do not invent synonyms, abbreviations, or translations that diverge from the glossary's canonical English term in parentheses.

## Design-First Frontend Development

1. **Verify Pencil design before frontend work** — before planning or implementing any frontend GUI, the agent must confirm that all necessary and sufficient UI elements (components, screens, layouts) exist in the Pencil design files under `design/`; if the required designs are missing, the implementation plan must include creating the design elements (under human supervision) as a prerequisite before any frontend code is written.

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
4. **Internal AI Factory artifacts are excluded from mandatory traceability** — the following artifacts are internal agent workflow metadata and do not require entries in `.ai-factory/traceability/traceability.ttl` unless they directly define product behavior (e.g., a skill-context rule that mandates a production-visible configuration change):
   - `.ai-factory/evolutions/*` — agent self-improvement evolution logs
   - `.ai-factory/skill-context/*` — project-specific skill rules accumulated by `$aif-evolve`
   - Agent plan/review/fix metadata (plan files, review reports, fix patches)
   
   Product artifacts (service code, tests, configs, deployment specs, API contracts) remain subject to the mandatory traceability requirement in point 1 above. When in doubt, err on the side of adding a traceability entry — the exclusion is for strictly internal agent orchestration files that have no observable product behavior impact.

## Testing (TDD)

1. **Follow TDD methodology** — tests act as the reference for behavior verification; subsequent implementation must conform to the behavior described in tests, alongside requirements.
2. **Plan tests in order via `$aif-plan`** — unit tests per-task first, then integration tests as needed, then E2E tests for the whole feature.
3. **Major-feature plans start with E2E tests** — a plan for a major feature must begin by implementing the E2E tests for that feature.
4. **Plan integration tests for external interfaces** — if a feature touches external interfaces, the plan must include integration tests.
5. **Write unit tests just-in-time per task** — the plan must specify that unit tests are written immediately before each individual task adding new functionality; do not draft all unit tests for a milestone upfront.
6. **Update traceability for tests** — when tests are written or modified, reflect them in the artifact traceability ontology (`.ai-factory/traceability/traceability.ttl`).

## Documentation

1. **User-facing documentation must use Antora** — all user, developer, administrator, and operator documentation must be authored in AsciiDoc format under `src/docs/antora/` and organized as Antora component modules. This documentation is intended for publication via Antora site generation.
2. **System specification is maintained in the `specs/` submodule** — the `specs/` git submodule contains requirements, ADRs, use cases, user stories, and architectural specifications. This is the authoritative source for system design and must be consulted during development. Changes to specifications require updating the submodule separately.
3. **Requirements and ADRs must follow naming conventions** — when generating REQ-* requirement files, follow `specs/requirements/README.md`: format `REQ-<LEVEL>.<AREA>.<semantic-tag>.md` with LEVEL ∈ {BIZ, USR, FUN, NFR, CON} and AREA ∈ {API, DATA, INFRA, SECURITY, UI, PROCESS, INTEGRATION, STACK, OPS, DOC, SUP}. When generating ADR-* files, follow `specs/adr/README.md`: format `ADR-<LEVEL>.<AREA>.<semantic-tag>.md` with LEVEL ∈ {BIZ, DES, IMPL} and the same AREA set. Never invent new LEVEL or AREA values.
