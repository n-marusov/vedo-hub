# Implementation Plan: M5 Closing — Minimal ABox CRUD GUI E2E Tests

Branch: feature/m5-abox-crud-e2e
Created: 2026-08-02

## Settings
- Testing: yes — write GUI e2e tests for ABox CRUD (feature already exists)
- Logging: standard
- Docs: no — test-only change, no user-facing docs needed

## Roadmap Linkage
Milestone: "M5: MVP Scope Gap Closure"
Rationale: This plan closes the last remaining M5 gap — Minimal ABox CRUD GUI e2e coverage. REST CRUD for `/individuals` exists, `CreateIndividualDialog.vue` exists, ABox view exists. Only e2e tests are missing.

## Verified Baseline (2026-08-02)
- REST CRUD: `POST/PUT/DELETE /api/v1/ontologies/{id}/individuals` — backend exists
- Frontend: `CreateIndividualDialog.vue` + REST client `createIndividual()` — implemented
- ABox view: individuals table in `OntologyWorkspace.vue` three-view switcher — implemented (Q3)
- GUI e2e: `editor.spec.ts` (all 13 tests skipped, empty TODO stubs, post-MVP scope)
- GUI e2e: `browse.spec.ts` (all 8 tests skipped, empty TODO stubs, M8/M11 scope)

## Tasks

### Phase 1: ABox CRUD E2E Tests

- [x] **A1. Add page object methods for ABox operations**
  - Added `switchToABoxView()`, `getABoxIndividuals()`, `injectABoxIndividuals()`, `deleteIndividual()`, `updateIndividual()` to `OntologyWorkspacePage`.
  - **Files:** `tests/e2e/pages/ontology-workspace.page.ts`
  - **Acceptance:**
    - [x] All new methods accessible from test specs
    - [x] Methods follow existing patterns (DOM injection + locators)

- [x] **A2. Create abox-crud.spec.ts with full CRUD tests**
  - Create `tests/e2e/specs/gui/flows/abox-crud.spec.ts`
  - Tests:
    1. `should create individual and verify in ABox view when class and name provided` — POST /individuals → verify appears in ABox table
    2. `should show validation error when individual name is empty` — empty name → inline error, no request
    3. `should list individuals by class in ABox view` — switch to ABox view → verify list renders
    4. `should delete individual and remove from ABox view` — delete → verify removed
    5. `should update individual label and show new label in ABox view` — PUT → verify updated
  - Mock REST endpoints for GET/PUT/DELETE `/api/v1/ontologies/*/individuals/*`
  - **Files:** `tests/e2e/specs/gui/flows/abox-crud.spec.ts` (created)
  - **Acceptance:**
    - [x] 5 tests pass against mocked stack
    - [x] BDD naming: `'should <expected> when <condition>'` (TS convention)

### Phase 2: Closure

- [x] **C1. Update ROADMAP.md — mark Minimal ABox CRUD as [x]**
  - Change `[~]` to `[x]` on Minimal ABox CRUD sub-item
  - Update M5 milestone header: remove `[~]`, mark as `[x]` with verification note
  - Add M5 to Completed table with date 2026-08-02

## Acceptance Criteria (plan-level)
- [x] All ABox CRUD e2e tests pass against the test stack
- [x] M5 all sub-items `[x]`, M5 marked Completed in ROADMAP.md
- [ ] GUI e2e run: existing tests unaffected (119 pass / 78 skip → 124+ pass)

## Notes
- Minimal scope: basic CRUD only. Batch operations (M9), inline editing (M9), search/filter (M8) are out of scope.
- Uses same mock pattern as `ontology-lifecycle.spec.ts` — REST endpoint mocks for individuals.
- Tests run with the existing Playwright GUI config.
