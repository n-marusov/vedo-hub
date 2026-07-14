# Fix Plan: AI Factory Validation Guidance

**Created:** 2026-07-14
**Branch:** `feature/ontology-core-engine`
**Mode:** fast / current branch
**Objective:** Fix the `$aif-review` blocking finding and clarify traceability scope. Auto-generated `.ai-factory/evolutions/*` and `.ai-factory/skill-context/*` artifacts are internal AI Factory artifacts, not product artifacts, and do not require mandatory entries in `.ai-factory/traceability/traceability.ttl` unless they directly define product behavior.

## Settings

- **Testing:** yes — perform lightweight validation of the updated markdown guidance and verify command wording is directory-aware.
- **Logging:** standard — runtime logging is not applicable because this change only updates AI Factory guidance documents.
- **Docs:** no — no product documentation checkpoint is required for internal AI Factory guidance.

## Roadmap Linkage

- **Milestone:** none
- **Rationale:** Skipped by user; this is an internal agent-workflow correction, not a product milestone item.

## Scope

Modify only AI Factory artifacts:

- `.ai-factory/skill-context/aif-implement/SKILL.md`
- `.ai-factory/RULES.md` or the resolved rules artifact that owns the traceability rule

Do not modify product/service code.

## Tasks

### Phase 1 — Fix Go validation guidance

- [x] **Task 1: Make Go validation commands module-aware**
  - **Files:** `.ai-factory/skill-context/aif-implement/SKILL.md`
  - **Deliverable:** Replace generic `go vet ./...` / `go test ./...` wording with guidance that respects the multi-module monorepo layout.
  - **Expected behavior:** Implementers must run Go checks from each affected Go module directory, for example:
    - `cd src/services/api-gateway` then `go test ./...`
    - `cd src/services/ticket-api` then `go vet ./...`
    - `cd src/cli` then `go test ./...`
  - **Logging requirements:** Runtime logging is not applicable; do not add application logging requirements for this markdown-only update.
  - **Dependency notes:** No dependencies.

### Phase 2 — Clarify Rust validation guidance

- [x] **Task 2: Separate Rust type/lint gates from build gates**
  - **Files:** `.ai-factory/skill-context/aif-implement/SKILL.md`
  - **Deliverable:** Replace wording like `cargo check` or `cargo build` before merge with stricter validation guidance.
  - **Expected behavior:** The rule must state that:
    - `cargo check` and/or `cargo clippy` are the Rust validation/type/lint gates;
    - `cargo build` is a build gate and does not replace `cargo check` / `cargo clippy`;
    - commands must run from the affected Rust crate or workspace directory.
  - **Logging requirements:** Runtime logging is not applicable; do not add application logging requirements for this markdown-only update.
  - **Dependency notes:** Prefer doing this after Task 1 so validation guidance remains consistent across languages.

### Phase 3 — Scope GraphQL/E2E validation commands

- [x] **Task 3: Make GraphQL end-to-end commands directory-aware**
  - **Files:** `.ai-factory/skill-context/aif-implement/SKILL.md`
  - **Deliverable:** Clarify that commands such as `cargo test -p ontology-service --lib` and Go tests must be run from the correct workspace/module directory.
  - **Expected behavior:** The rule must avoid generic commands without a working directory. It should state:
    - Rust ontology tests run from the Rust workspace/crate directory where package `ontology-service` is available;
    - Go gateway/API tests run from the concrete affected Go module directory.
  - **Logging requirements:** Runtime logging is not applicable; if the guidance mentions debugging, keep it limited to implementation-time runtime code, not this markdown-only update.
  - **Dependency notes:** Depends on Tasks 1 and 2 for consistent validation wording.

### Phase 4 — Clarify traceability scope

- [x] **Task 4: Exclude internal AI Factory artifacts from mandatory product traceability**
  - **Files:** `.ai-factory/RULES.md` or the resolved rules artifact that owns the traceability rule
  - **Deliverable:** Add an explicit clarification that mandatory traceability applies to product/service artifacts, not internal agent workflow artifacts.
  - **Expected behavior:** The rule must state that these internal AI Factory artifacts do not require mandatory entries in `.ai-factory/traceability/traceability.ttl` unless they directly define product behavior:
    - `.ai-factory/evolutions/*`
    - `.ai-factory/skill-context/*`
    - agent plan/review/fix metadata
  - **Logging requirements:** Runtime logging is not applicable.
  - **Dependency notes:** Independent of Tasks 1–3.

### Phase 5 — Validate the corrected guidance

- [x] **Task 5: Verify updated rules remove ambiguous command guidance**
  - **Files:** `.ai-factory/skill-context/aif-implement/SKILL.md`, `.ai-factory/RULES.md`
  - **Deliverable:** Inspect the changed markdown and confirm no misleading validation or traceability wording remains.
  - **Expected behavior:** The changed guidance must no longer contain:
    - `go vet ./...` as a standalone recommendation without a Go module working directory;
    - `go test ./...` as a standalone recommendation without a Go module working directory;
    - `cargo build` presented as a replacement for `cargo check` / `cargo clippy`;
    - mandatory product traceability requirements for `.ai-factory/evolutions/*` or `.ai-factory/skill-context/*`.
  - **Logging requirements:** Runtime logging is not applicable; record validation results in the final implementation summary.
  - **Dependency notes:** Depends on Tasks 1–4.

## Acceptance Criteria

- [x] Go validation guidance no longer implies root-level `go vet ./...` / `go test ./...` in a monorepo without root `go.mod`.
- [x] Rust guidance clearly separates `cargo check` / `cargo clippy` from `cargo build`.
- [x] GraphQL/E2E validation guidance identifies the correct workspace/module working directory.
- [x] Traceability rule is explicitly scoped to product/service artifacts.
- [x] `.ai-factory/evolutions/*` and `.ai-factory/skill-context/*` are documented as internal AI Factory artifacts that do not require mandatory traceability entries unless they directly define product behavior.
- [x] Changes remain on the current branch; no branch/worktree is created.

## Progress Tracking

Update task checkboxes in this file as implementation progresses:

- `[ ]` not started
- `[~]` in progress
- `[x]` complete
- `[!]` blocked

## Commit Plan

Single commit after all tasks are complete:

```text
Fix AI Factory validation guidance
```

## Next Steps

To start implementation, run:

```text
$aif-implement
```
