# MVP Gap Closure — Verification & Fixes Report (artifact.md)

Run: mvp-gap-closure-verify-20260802-000000 — completed 2026-08-02 (threshold_reached, score 1.0)

## Result

All tasks in `feature-mvp-gap-closure.md` (T1–T5, Q1–Q4, F1–F2, Step 22) are implemented and
verified green by unit/integration/E2E tests. No deviations from specs/ADR.

## Test evidence (2026-08-02)

| Suite | Result |
|---|---|
| Go unit (10 modules) | 727 pass |
| Rust unit | ontology 122/122, versioning 80/80, clippy clean |
| Versioning integration | transactional 5/5, full green |
| Python | doc-extractor 61 pass/4 skip, metrics-service 5 pass |
| Frontend vitest | frontend 256 pass/67 todo, publish-browse-ui 9 pass |
| Playwright API E2E | 66/66 (stable x2) |
| Playwright GUI E2E | 119 pass / 78 skip (all milestone-annotated) / 0 fail (stable x2) |
| Security RBAC | 45/49; 4 RED (TC-001/003/011/012 BOLA «never 404») tracked to M7, fail on assertions not env dial |
| Gates | all pass (shell, contract, bola-bfla, python manifests, test-quality TQS 78.0 bronze, traceability VALID) |
| Specs validation | tests/specs PASS |

## Changes made in this run

### F2 — GUI specs (un-skip / fix / annotate)
- Un-skipped 9 AI tests (ai-completion, ai-property-suggestions, iterative-refinement) —
  features implemented in M4 Д1-Д3; rewritten to the real UI/API contract.
- Fixed 4 stale specs against the Q3 frontend rebuild:
  - `ontology-lifecycle.spec.ts` — GraphQL ClassTree mock registered before navigation; stateful created-classes; `injectClassTree` injects real `.class-tree__node`.
  - `graph-visualization.spec.ts` — TBox Graph tab (5/5).
  - `org-lifecycle.spec.ts` — `auth.fixture.ts` sets `sessionStorage.vedo_session` (SKIP_AUTH initSession overwrote the JWT with skip-auth-token → 401); locale ru-RU.
  - `versioning-tabs.spec.ts` — `graphql-fixtures.ts` route for GitLab-aligned `/api/v1/projects/*/repository/*` (ADR-DES.API.rest-gitlab-alignment).
- Page object `ontology-workspace.page.ts`: getClassTree/selectClass/injectClassTree support the real ClassTree.
- Annotated all 78 remaining skips with milestone + backlog reference.

### Product bugs found & fixed (spec-compliant)
- `OntologyWorkspace.vue`: `maxRefinementRounds` 5 → 3 per REQ-FUN.API.max-refinement-iterations (≤ 3 cycles);
  added limit warning + disabled Refine button at limit; new i18n keys en/ru.
- `OntologyWorkspace.vue`: class tree refetches after class creation (`refetchClassTree`).

### Step 22 — traceability + ROADMAP
- `traceability.ttl`: added TestSuite entries for metrics `tests/test_metrics.py`,
  `OntologyWorkspaceViews.spec.ts`, `CommentsPage.spec.ts`, `LandingPage.spec.ts`
  (+ `vdo:validates` REQ-USR.UI.gui-implementation). Validator PASS.
- `ROADMAP.md`: M5 items marked `[x]` (merge blocking, SPARQL audit, graph views, comment wiring,
  F11.1, forks security, versioning integration); notes updated to verified-green state.
- Plan file: F2 + Step 22 marked done; all plan-level acceptance criteria `[x]`; 0 unchecked.
