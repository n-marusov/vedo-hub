## Change Summary

**Commits:** N/A (specification-driven scope, not a git diff)
**Changed files:** N/A (source = `specs/` RBAC-related ADRs and requirements)
**Risk level:** 🔴 High

---

### What Changed

This change summary covers the **full RBAC perimeter** of VEDO Core as defined in the project specifications. It is not tied to a specific git branch diff; instead, the "change" under test is the complete RBAC implementation surface that must be verified before production release.

The RBAC system is built on a GitLab-like organization model (`Group`, `Project`, `Ontology` 1:1, members, visibility, ABAC policies) with four base roles (`Viewer`/`Editor`/`Maintainer`/`Owner` mapped to GitLab access levels `Guest=10`/`Developer=30`/`Maintainer=40`/`Owner=50`), role inheritance across the `Group/Subgroup/Project` hierarchy, visibility levels (`Private`/`Internal`/`Public`), attribute-based access control via semantic graph patterns, protected `main` branch with Ontology Merge Request workflow, MFA for destructive operations, JIT/PAM for privileged access, break-glass emergency admin, and mandatory BOLA/BFLA negative authorization tests as a blocking release gate.

---

### Affected Areas

| Component                          | Change type | Description                                                                                          |
|------------------------------------|-------------|------------------------------------------------------------------------------------------------------|
| Organization model                 | Spec        | `Group`/`Project`/`Ontology` 1:1, membership boundary on `Owner`, role inheritance, visibility      |
| REST organization endpoints        | Spec        | Canonical `/api/v1/groups/*` and `/api/v1/projects/*` paths, members/visibility/policies endpoints  |
| Role hierarchy                     | Spec        | `Viewer`→`Editor`→`Maintainer`→`Owner`; GitLab access levels 10/20/30/40/50; max-role-wins inheritance |
| Visibility levels                  | Spec        | `Private`/`Internal`/`Public` stored on `Project`, inherited by `Ontology` via 1:1                  |
| ABAC policies                      | Spec        | Semantic graph patterns (URI prefix, annotation, parent class, relationship); most-specific-wins    |
| Protected `main` + OMR workflow    | Spec        | `write_with_approval` → proposal branch → OMR → Maintainer review → merge/reject                    |
| Fork semantics                    | Spec        | `POST /projects/{id}/fork` — read access sufficient, Guest can fork public, 403 (never 404) on denial |
| Idempotency                        | Spec        | `Idempotency-Key` required on members/visibility/policies write endpoints → 400 if missing          |
| BOLA/BFLA negative tests           | Spec        | 4 test types (A cross-tenant, B cross-object, C BFLA, D IDOR); 100% P0 coverage; 403 not 404/500    |
| Authorization regression gate      | Spec        | Blocking release gate: 0 critical defects, 100% negative test coverage, signed policy bundles       |
| MFA for critical ops               | Spec        | Categories A/B mandatory non-cacheable MFA; C default-on; D none; service-account scope-limited     |
| JIT/PAM                            | Spec        | No standing access; TTL ≤ 60 min; 100% session recording; phishing-resistant MFA (FIDO2/WebAuthn)   |
| Break-glass Emergency Admin        | Spec        | Shamir M-of-N; `/emergency/login` independent of Keycloak; 30-min TTL; immutable audit              |
| Destructive command guardrails     | Spec        | G1–G4 levels; typed confirmation, environment guard, MFA/second-admin, cool-down + verification     |
| Audit logging                      | Spec        | Every write endpoint emits structured `audit_events` with `event`/`user_id`/`object_id`/`trace_id`  |
| Public ontology access             | Spec        | Browse UI + Public Browse API, read-only, no auth, rate-limited, isolated Neo4j                    |

---

### Evidence

| Finding                                              | Evidence                                                                                                                              |
|------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------|
| GitLab-like org model with 4 roles                   | `specs/adr/ADR-DES.SECURITY.gitlab-like-organization-model.md`, `specs/requirements/REQ-NFR.SECURITY.organization-access-model.md`     |
| Canonical REST contract for org endpoints            | `specs/adr/ADR-DES.API.organization-rest-endpoints.md` § "Канонические пути", "RBAC", "Audit"                                          |
| Access levels 10/20/30/40/50                         | `specs/adr/ADR-DES.API.organization-rest-endpoints.md` § "RBAC"; `.ai-factory/references/gitlab-projects-groups-api.md` § "Access Levels Reference" |
| Visibility Private/Internal/Public on Project        | `specs/requirements/REQ-FUN.DATA.ontology-visibility-levels.md`                                                                       |
| BOLA/BFLA 4 test types, 403 not 404/500              | `specs/adr/ADR-IMPL.SECURITY.bola-bfla-negative-tests-mandate.md`, `specs/requirements/REQ-NFR.SECURITY.bola-bfla-negative-tests.md`  |
| Authorization regression gate (blocking)             | `specs/adr/ADR-DES.SECURITY.authorization-policy-gates-strategy.md`, `specs/requirements/REQ-NFR.SECURITY.authorization-regression-gates.md` |
| MFA categories A/B/C/D                               | `specs/adr/ADR-DES.SECURITY.mfa-critical-ops-mandate.md`                                                                              |
| JIT/PAM, no standing access, TTL ≤ 60 min           | `specs/adr/ADR-DES.SECURITY.privileged-access-jit-pam-strategy.md`, `specs/requirements/REQ-NFR.SECURITY.privileged-access-control.md` |
| Break-glass Shamir M-of-N, `/emergency/login`       | `specs/adr/ADR-DES.SECURITY.break-glass-access-strategy.md`, `specs/requirements/REQ-NFR.SECURITY.security-requirements.md` § "Emergency Admin" |
| Destructive guardrails G1–G4                         | `specs/adr/ADR-DES.SECURITY.destructive-command-guardrails.md`                                                                        |
| Fork: read access sufficient, 403 never 404          | `specs/adr/ADR-DES.API.organization-rest-endpoints.md` § "Fork-семантика"                                                              |
| Idempotency-Key required on members/visibility/policies | `specs/adr/ADR-DES.API.organization-rest-endpoints.md` § "Idempotency"                                                              |
| Public browse: no auth, read-only, rate-limited      | `specs/adr/ADR-DES.SECURITY.public-ontology-access.md`                                                                                |
| OMR workflow: proposal branch → Maintainer review    | `specs/adr/ADR-DES.PROCESS.merge-request-strategy.md`, `specs/requirements/REQ-NFR.SECURITY.organization-access-model.md` § "Protected main" |

---

### Risks

🔴 **Critical** (must verify):

- **Cross-tenant BOLA** — подмена `tenantId`/`ontologyId` даёт доступ к чужому tenant. BOLA — #1 OWASP API Security Top 10. Multi-tenant SaaS, каждый запрос содержит объектные ID.
- **BFLA privilege escalation** — пользователь с ролью `Viewer`/`Guest` выполняет admin/destructive операцию (delete ontology, manage members, change visibility).
- **Membership boundary bypass** — `Maintainer` или `Editor` может добавить/удалить члена или изменить роль (должен только `Owner`).
- **IDOR через guessable IDs** — предсказуемые ID (numeric+1, UUID variant) возвращают 404 вместо 403, раскрывая существование объекта.
- **Visibility bypass** — `Private` Project доступен не-члену; `Internal` доступен неаутентифицированному; `Public` раскрывает служебные данные.
- **Fork RBAC bypass** — fork приватного Project без read-доступа возвращает 404 вместо 403; Guest не может fork публичного.
- **Role inheritance failure** — max-role-wins не срабатывает; пользователь с `Viewer` в subgroup и `Owner` в parent получает `Viewer` вместо `Owner`.
- **MFA bypass для категорий A/B** — destructive операция (tenant purge, backup delete, restore) выполняется без MFA или с кэшированным MFA.
- **Break-glass abuse** — Emergency Admin используется вне P0 инцидента; TTL превышает 30 мин; audit log не immutable.
- **Audit log missing** — write-endpoint не эмиттит `audit_events` или запись не содержит обязательные поля (`event`, `user_id`, `object_id`, `trace_id`).

🟡 **Medium** (should verify):

- **Idempotency-Key missing** — write-endpoint members/visibility/policies выполняется без `Idempotency-Key` (должен 400 `INVALID_IDEMPOTENCY_KEY`).
- **ABAC most-specific-wins failure** — более специфичное правило на классе/свойстве не ограничивает общее право на уровне Project.
- **OMR workflow bypass** — Editor коммитит напрямую в protected `main` при `write_with_approval` без proposal branch.
- **JIT/PAM standing access** — привилегированная сессия превышает TTL 60 мин; session recording < 100%.
- **Destructive guardrails G1–G4** — destructive-команда выполняется без соответствующего уровня подтверждения (typed confirmation, environment guard, MFA/second-admin, cool-down).
- **Service-account scope overflow** — service-account токен выполняет операции категории A/B (запрещены) или превышает allow-list scope.
- **Public Browse API abuse** — отсутствие rate limiting; произвольный Cypher/SPARQL принимается; Public Neo4j содержит служебные данные.

🟢 **Low** (nice to verify):

- **Legacy scope alias `"ontology/" + id`** — принимается вне окна миграции или не удаляется после cleanup.
- **GraphQL non-graph queries** — GraphQL обслуживает не-графовые запросы (members, groups) вместо REST после tightening boundary.
- **Pagination** — `per_page` > 100 не отклоняется; cursor-based pagination нарушен.

---

### Testing Recommendations

**First priority:**

- [ ] Cross-tenant BOLA для всех P0 endpoint-классов (`/ontologies/{id}`, `/branches/{id}`, `/commits/{id}`, GraphQL mutations/queries, WebSocket JoinRoom)
- [ ] BFLA privilege escalation: `Viewer`/`Guest` → admin/destructive endpoints (delete ontology, manage members, change visibility, manage policies)
- [ ] Membership boundary: только `Owner` управляет members/visibility/policies; `Maintainer`/`Editor`/`Viewer` получают 403
- [ ] IDOR: guessable IDs возвращают 403, никогда 404/500
- [ ] Visibility enforcement: `Private`/`Internal`/`Public` доступ строго по модели
- [ ] Fork RBAC: read access sufficient для public/internal; 403 (never 404) для private без доступа
- [ ] MFA для категорий A/B: non-cacheable, обязательный; service-account не может выполнять A/B
- [ ] Audit log: каждый write-endpoint эмиттит `audit_events` со всеми обязательными полями

**Regression:**

- [ ] Role inheritance: max-role-wins по иерархии Group/Subgroup/Project
- [ ] Idempotency-Key: 400 `INVALID_IDEMPOTENCY_KEY` при отсутствии на members/visibility/policies
- [ ] OMR workflow: `write_with_approval` → proposal branch → Maintainer review → merge/reject
- [ ] Break-glass: `/emergency/login` независим от Keycloak; 30-min TTL; immutable audit; Shamir M-of-N
- [ ] JIT/PAM: TTL ≤ 60 min; 100% session recording; no standing access
- [ ] Destructive guardrails G1–G4: соответствующий уровень подтверждения для каждой destructive-команды
- [ ] Public Browse API: rate limiting; allowlisted endpoints; isolated Neo4j; no arbitrary Cypher/SPARQL
