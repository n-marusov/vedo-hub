## Test Cases: Full RBAC Testing

**Branch / Version:** rbac-full (specification-driven scope)
**Environment:** staging / local docker-compose (all-services)
**Date:** 2026-07-23

### Test Data Conventions

| Alias | Role / Tenant | Notes |
|-------|---------------|-------|
| `u_viewer_A` | Viewer (Guest=10), tenant_A | member of `proj_private_A` |
| `u_reporter_A` | Reporter (20), tenant_A | member of `proj_private_A` |
| `u_editor_A` | Editor (Developer=30), tenant_A | member of `proj_private_A` |
| `u_maintainer_A` | Maintainer (40), tenant_A | member of `proj_private_A` |
| `u_owner_A` | Owner (50), tenant_A | member of `proj_private_A` |
| `u_outsider_A` | no role, tenant_A | authenticated, not a member |
| `u_viewer_B` | Viewer (Guest=10), tenant_B | cross-tenant |
| `u_owner_B` | Owner (50), tenant_B | cross-tenant |
| `u_anon` | unauthenticated | no JWT |
| `proj_private_A` | Private, tenant_A | contains `ont_A` |
| `proj_internal_A` | Internal, tenant_A | contains `ont_internal_A` |
| `proj_public_A` | Public, tenant_A | contains `ont_public_A` |
| `proj_private_B` | Private, tenant_B | contains `ont_B` |
| `proj_private_A2` | Private, tenant_A | second project, u_viewer_A is NOT a member |

All write requests include header `Idempotency-Key: <uuid>` unless the TC explicitly omits it. JWT obtained via Keycloak OIDC `password` grant for test users.

---

### A. Cross-Tenant BOLA (Type A)

#### TC-001: Cross-tenant read of private ontology via REST

**Priority:** High
**Type:** Negative

**Precondition:** `u_viewer_A` authenticated; `proj_private_B` / `ont_B` exist in `tenant_B`.

**Steps:**

1. Obtain JWT for `u_viewer_A` (tenant_A).
2. `GET /api/v1/ontologies/ont_B` with `Authorization: Bearer <jwt_u_viewer_A>`.
3. Inspect HTTP status code and response body.
4. Query `audit_events` table for an entry with `object_id = ont_B` and `user_id = u_viewer_A` within the last 5 seconds.

**Expected result:**

- HTTP status `403`.
- Response body contains `error_code: FORBIDDEN_CROSS_TENANT_ACCESS`.
- No `404` or `500` returned.
- An `audit_events` row exists with `event` containing the denied access.

**Test data:**

```
GET /api/v1/ontologies/ont_B
Authorization: Bearer <jwt_u_viewer_A>
```

---

#### TC-002: Cross-tenant create ontology in foreign tenant

**Priority:** High
**Type:** Negative

**Precondition:** `u_owner_A` authenticated; `tenant_B` exists.

**Steps:**

1. Obtain JWT for `u_owner_A` (tenant_A).
2. `POST /api/v1/ontologies` with body `{ "tenantId": "tenant_B", "name": "stolen-ont" }` and `Idempotency-Key: <uuid>`.
3. Inspect status code and body.

**Expected result:**

- HTTP `403` with `error_code: FORBIDDEN_CROSS_TENANT_ACCESS`.
- No ontology row created in `tenant_B`.
- `audit_events` records the denied attempt.

**Test data:**

```
POST /api/v1/ontologies
{ "tenantId": "tenant_B", "name": "stolen-ont" }
Idempotency-Key: 11111111-1111-1111-1111-111111111111
Authorization: Bearer <jwt_u_owner_A>
```

---

#### TC-003: Cross-tenant access to branch of foreign ontology

**Priority:** High
**Type:** Negative

**Precondition:** `u_editor_A` authenticated; `branch_feature_B` exists in `ont_B` (`tenant_B`).

**Steps:**

1. Obtain JWT for `u_editor_A` (tenant_A).
2. `GET /api/v1/branches/branch_feature_B` with the JWT.
3. Inspect status code.

**Expected result:**

- HTTP `403` with `error_code: FORBIDDEN_CROSS_TENANT_ACCESS`.
- Never `404` or `500`.

**Test data:**

```
GET /api/v1/branches/branch_feature_B
Authorization: Bearer <jwt_u_editor_A>
```

---

#### TC-004: Cross-tenant WebSocket JoinRoom for foreign ontology

**Priority:** High
**Type:** Negative

**Precondition:** `u_editor_A` authenticated; WebSocket endpoint reachable; `ont_B` exists in `tenant_B`.

**Steps:**

1. Open WebSocket connection to collaboration endpoint with `u_editor_A` JWT.
2. Send `JoinRoom` message: `{ "type": "join", "ontologyId": "ont_B" }`.
3. Inspect the response frame.

**Expected result:**

- WebSocket server replies with an error frame: `{ "type": "error", "code": "FORBIDDEN_CROSS_TENANT_ACCESS" }`.
- Connection is **not** joined to the `ont_B` room.
- `audit_events` contains a denied-join record.

**Test data:**

```
WS /ws/collaboration
Authorization: Bearer <jwt_u_editor_A>
{ "type": "join", "ontologyId": "ont_B" }
```

---

### B. Cross-Object BOLA (Type B — same tenant, no role)

#### TC-005: Authenticated user without role reads restricted project

**Priority:** High
**Type:** Negative

**Precondition:** `u_outsider_A` authenticated (tenant_A) but NOT a member of `proj_private_A`.

**Steps:**

1. Obtain JWT for `u_outsider_A`.
2. `GET /api/v1/projects/proj_private_A` with the JWT.
3. Inspect status code and body.

**Expected result:**

- HTTP `403` with `error_code: FORBIDDEN_INSUFFICIENT_ROLE`.
- No project details leaked.

**Test data:**

```
GET /api/v1/projects/proj_private_A
Authorization: Bearer <jwt_u_outsider_A>
```

---

#### TC-006: Member of one project cannot read another project in same tenant

**Priority:** High
**Type:** Negative

**Precondition:** `u_viewer_A` is a member of `proj_private_A` but NOT of `proj_private_A2` (both in `tenant_A`).

**Steps:**

1. Obtain JWT for `u_viewer_A`.
2. `GET /api/v1/projects/proj_private_A2` with the JWT.
3. Inspect status code.

**Expected result:**

- HTTP `403` with `error_code: FORBIDDEN_INSUFFICIENT_ROLE`.

**Test data:**

```
GET /api/v1/projects/proj_private_A2
Authorization: Bearer <jwt_u_viewer_A>
```

---

### C. BFLA — Privilege Escalation (Type C)

#### TC-007: Viewer attempts to delete ontology

**Priority:** High
**Type:** Negative

**Precondition:** `u_viewer_A` authenticated; `ont_A` exists in `proj_private_A`.

**Steps:**

1. Obtain JWT for `u_viewer_A`.
2. `DELETE /api/v1/ontologies/ont_A` with `Idempotency-Key: <uuid>`.
3. Inspect status code.
4. Verify `ont_A` still exists via `GET` as `u_owner_A`.

**Expected result:**

- HTTP `403` with `error_code: FORBIDDEN_INSUFFICIENT_ROLE`.
- `ont_A` is not deleted (verified by successful `GET` as owner).

**Test data:**

```
DELETE /api/v1/ontologies/ont_A
Idempotency-Key: 22222222-2222-2222-2222-222222222222
Authorization: Bearer <jwt_u_viewer_A>
```

---

#### TC-008: Editor attempts to add a project member

**Priority:** High
**Type:** Negative

**Precondition:** `u_editor_A` authenticated; `proj_private_A` exists; `u_outsider_A` is NOT a member.

**Steps:**

1. Obtain JWT for `u_editor_A`.
2. `POST /api/v1/projects/proj_private_A/members` with body `{ "userId": "u_outsider_A", "role": "Viewer" }` and `Idempotency-Key`.
3. Inspect status code.
4. Verify `u_outsider_A` was NOT added by listing members as `u_owner_A`.

**Expected result:**

- HTTP `403` with `error_code: FORBIDDEN_INSUFFICIENT_ROLE`.
- `u_outsider_A` NOT present in `GET /api/v1/projects/proj_private_A/members`.

**Test data:**

```
POST /api/v1/projects/proj_private_A/members
{ "userId": "u_outsider_A", "role": "Viewer" }
Idempotency-Key: 33333333-3333-3333-3333-333333333333
Authorization: Bearer <jwt_u_editor_A>
```

---

#### TC-009: Maintainer attempts to change project visibility

**Priority:** High
**Type:** Negative

**Precondition:** `u_maintainer_A` authenticated; `proj_private_A` exists (visibility = `private`).

**Steps:**

1. Obtain JWT for `u_maintainer_A`.
2. `PUT /api/v1/projects/proj_private_A/visibility` with body `{ "visibility": "internal" }` and `Idempotency-Key`.
3. Inspect status code.
4. Verify visibility unchanged via `GET` as `u_owner_A`.

**Expected result:**

- HTTP `403` with `error_code: FORBIDDEN_INSUFFICIENT_ROLE`.
- Visibility remains `private`.

**Test data:**

```
PUT /api/v1/projects/proj_private_A/visibility
{ "visibility": "internal" }
Idempotency-Key: 44444444-4444-4444-4444-444444444444
Authorization: Bearer <jwt_u_maintainer_A>
```

---

#### TC-010: Maintainer attempts to create ABAC policy

**Priority:** High
**Type:** Negative

**Precondition:** `u_maintainer_A` authenticated; `proj_private_A` exists.

**Steps:**

1. Obtain JWT for `u_maintainer_A`.
2. `POST /api/v1/projects/proj_private_A/policies` with `{ "pattern": { "uriPrefix": "http://vedo/example/" }, "effect": "deny", "action": "delete" }` and `Idempotency-Key`.
3. Inspect status code.
4. Verify policy NOT created via `GET` as `u_owner_A`.

**Expected result:**

- HTTP `403` with `error_code: FORBIDDEN_INSUFFICIENT_ROLE`.
- Policy list unchanged.

**Test data:**

```
POST /api/v1/projects/proj_private_A/policies
{ "pattern": { "uriPrefix": "http://vedo/example/" }, "effect": "deny", "action": "delete" }
Idempotency-Key: 55555555-5555-5555-5555-555555555555
Authorization: Bearer <jwt_u_maintainer_A>
```

---

### D. IDOR — Guessable IDs (Type D)

#### TC-011: Sequential numeric ID returns 403, never 404

**Priority:** High
**Type:** Negative

**Precondition:** `u_viewer_A` authenticated; an ontology with numeric id `1001` exists in `tenant_B`; `1002` does not exist.

**Steps:**

1. Obtain JWT for `u_viewer_A`.
2. `GET /api/v1/ontologies/1001` (existing foreign object).
3. `GET /api/v1/ontologies/1002` (non-existent foreign object).
4. Inspect both status codes.

**Expected result:**

- Both requests return `403` with `error_code: FORBIDDEN_OBJECT_NOT_FOUND_OR_ACCESS_DENIED`.
- Neither returns `404` or `500`.
- Existence of `1001` is not leaked.

**Test data:**

```
GET /api/v1/ontologies/1001
GET /api/v1/ontologies/1002
Authorization: Bearer <jwt_u_viewer_A>
```

---

#### TC-012: Invalid UUID variant returns 403, never 404

**Priority:** High
**Type:** Negative

**Precondition:** `u_viewer_A` authenticated; `ont_B` exists in `tenant_B`.

**Steps:**

1. Obtain JWT for `u_viewer_A`.
2. Mutate one hex char of `ont_B` to produce `ont_B_variant` (invalid UUID, not in DB).
3. `GET /api/v1/ontologies/ont_B_variant`.
4. Inspect status code.

**Expected result:**

- HTTP `403` with `error_code: FORBIDDEN_OBJECT_NOT_FOUND_OR_ACCESS_DENIED`.
- Never `404`.

**Test data:**

```
GET /api/v1/ontologies/<ont_B with last hex char changed>
Authorization: Bearer <jwt_u_viewer_A>
```

---

### E. Membership Boundary (Owner-only)

#### TC-013: Owner adds a new member (positive)

**Priority:** High
**Type:** Positive

**Precondition:** `u_owner_A` authenticated; `u_outsider_A` is NOT a member of `proj_private_A`.

**Steps:**

1. Obtain JWT for `u_owner_A`.
2. `POST /api/v1/projects/proj_private_A/members` with `{ "userId": "u_outsider_A", "role": "Viewer" }` and `Idempotency-Key`.
3. Inspect status code.
4. `GET /api/v1/projects/proj_private_A/members` as `u_owner_A`.

**Expected result:**

- HTTP `201` with the new membership.
- `u_outsider_A` appears in the member list with role `Viewer`.
- Audit row with `event: member.added`, `object_id: proj_private_A`.

**Test data:**

```
POST /api/v1/projects/proj_private_A/members
{ "userId": "u_outsider_A", "role": "Viewer" }
Idempotency-Key: 66666666-6666-6666-6666-666666666666
Authorization: Bearer <jwt_u_owner_A>
```

---

#### TC-014: Owner changes member role (positive)

**Priority:** High
**Type:** Positive

**Precondition:** `u_owner_A` authenticated; `u_viewer_A` is a member of `proj_private_A` with role `Viewer`.

**Steps:**

1. Obtain JWT for `u_owner_A`.
2. `PUT /api/v1/projects/proj_private_A/members/u_viewer_A` with `{ "role": "Editor" }` and `Idempotency-Key`.
3. Inspect status code.
4. Verify via member list.

**Expected result:**

- HTTP `200`.
- `u_viewer_A` role is now `Editor`.
- Audit row with `event: member.role_changed`.

**Test data:**

```
PUT /api/v1/projects/proj_private_A/members/u_viewer_A
{ "role": "Editor" }
Idempotency-Key: 77777777-7777-7777-7777-777777777777
Authorization: Bearer <jwt_u_owner_A>
```

---

#### TC-015: Owner removes a member (positive)

**Priority:** High
**Type:** Positive

**Precondition:** `u_owner_A` authenticated; `u_editor_A` is a member of `proj_private_A`.

**Steps:**

1. Obtain JWT for `u_owner_A`.
2. `DELETE /api/v1/projects/proj_private_A/members/u_editor_A` with `Idempotency-Key`.
3. Inspect status code.
4. Verify `u_editor_A` no longer in member list.

**Expected result:**

- HTTP `200` or `204`.
- `u_editor_A` not in member list.
- Audit row with `event: member.removed`.

**Test data:**

```
DELETE /api/v1/projects/proj_private_A/members/u_editor_A
Idempotency-Key: 88888888-8888-8888-8888-888888888888
Authorization: Bearer <jwt_u_owner_A>
```

---

#### TC-016: Viewer attempts to remove a member (negative)

**Priority:** High
**Type:** Negative

**Precondition:** `u_viewer_A` authenticated; `u_editor_A` is a member of `proj_private_A`.

**Steps:**

1. Obtain JWT for `u_viewer_A`.
2. `DELETE /api/v1/projects/proj_private_A/members/u_editor_A` with `Idempotency-Key`.
3. Inspect status code.
4. Verify `u_editor_A` still a member.

**Expected result:**

- HTTP `403` with `error_code: FORBIDDEN_INSUFFICIENT_ROLE`.
- `u_editor_A` remains in member list.

**Test data:**

```
DELETE /api/v1/projects/proj_private_A/members/u_editor_A
Idempotency-Key: 99999999-9999-9999-9999-999999999999
Authorization: Bearer <jwt_u_viewer_A>
```

---

### F. Role Inheritance (Max-Role-Wins)

#### TC-017: User with Owner in parent Group and Viewer in Subgroup gets Owner via max-role-wins

**Priority:** High
**Type:** Positive

**Precondition:** Group hierarchy `group_parent` → `subgroup_child` → `proj_in_child`. `u_user` has `Owner` on `group_parent` and `Viewer` on `subgroup_child`.

**Steps:**

1. Obtain JWT for `u_user`.
2. `GET /api/v1/projects/proj_in_child` (basic read).
3. `POST /api/v1/projects/proj_in_child/members` with `{ "userId": "new_member", "role": "Viewer" }` and `Idempotency-Key` (Owner-only operation).
4. Inspect both status codes.

**Expected result:**

- Read returns `200`.
- Member-add returns `201`/`200` (max-role-wins → effective role `Owner`).
- Audit row with `event: member.added`.

**Test data:**

```
u_user: Owner on group_parent, Viewer on subgroup_child
POST /api/v1/projects/proj_in_child/members
{ "userId": "new_member", "role": "Viewer" }
Idempotency-Key: aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa
Authorization: Bearer <jwt_u_user>
```

---

#### TC-018: Viewer role inheritance from parent Group applies to Subgroup Project

**Priority:** High
**Type:** Positive

**Precondition:** `u_viewer_A` has `Viewer` on `group_parent` only (no explicit role on `subgroup_child` or `proj_in_child`).

**Steps:**

1. Obtain JWT for `u_viewer_A`.
2. `GET /api/v1/projects/proj_in_child`.
3. `POST /api/v1/ontologies/ont_in_child/classes` (write attempt).
4. Inspect both status codes.

**Expected result:**

- Read returns `200` (inherited Viewer grants read access).
- Write returns `403` (inherited Viewer does not grant Editor permission).

**Test data:**

```
u_viewer_A: Viewer on group_parent, no explicit role on subgroup_child
GET /api/v1/projects/proj_in_child
POST /api/v1/ontologies/ont_in_child/classes { "name": "TestClass" }
Authorization: Bearer <jwt_u_viewer_A>
```

---

### G. Visibility Levels

#### TC-019: Private Project — non-member gets 403

**Priority:** High
**Type:** Negative

**Precondition:** `proj_private_A` exists with visibility `private`; `u_outsider_A` is authenticated but NOT a member.

**Steps:**

1. Obtain JWT for `u_outsider_A`.
2. `GET /api/v1/projects/proj_private_A`.
3. Inspect status code.

**Expected result:**

- HTTP `403` with `error_code: FORBIDDEN_INSUFFICIENT_ROLE`.
- No project details or ontology content returned.

**Test data:**

```
GET /api/v1/projects/proj_private_A
Authorization: Bearer <jwt_u_outsider_A>
```

---

#### TC-020: Internal Project — unauthenticated user gets 401

**Priority:** High
**Type:** Negative

**Precondition:** `proj_internal_A` exists with visibility `internal`; `u_anon` is unauthenticated (no JWT).

**Steps:**

1. `GET /api/v1/projects/proj_internal_A` without any `Authorization` header.
2. Inspect status code.

**Expected result:**

- HTTP `401` (unauthorized).
- No project details returned.

**Test data:**

```
GET /api/v1/projects/proj_internal_A
(no Authorization header)
```

---

#### TC-021: Internal Project — authenticated user with no role gets 200

**Priority:** High
**Type:** Positive

**Precondition:** `proj_internal_A` exists with visibility `internal`; `u_outsider_A` is authenticated (tenant_A) but NOT a member.

**Steps:**

1. Obtain JWT for `u_outsider_A`.
2. `GET /api/v1/projects/proj_internal_A`.
3. Inspect status code and body.

**Expected result:**

- HTTP `200`.
- Project details returned.
- Ontology content is readable (Internal = all authenticated users).

**Test data:**

```
GET /api/v1/projects/proj_internal_A
Authorization: Bearer <jwt_u_outsider_A>
```

---

#### TC-022: Public Project — anonymous user gets 200

**Priority:** High
**Type:** Positive

**Precondition:** `proj_public_A` exists with visibility `public`; `u_anon` is unauthenticated.

**Steps:**

1. `GET /api/v1/projects/proj_public_A` without any `Authorization` header.
2. `GET /api/v1/ontologies/ont_public_A/classes` without any `Authorization` header.
3. Inspect both status codes.

**Expected result:**

- Both requests return HTTP `200`.
- Ontology content is fully readable by anyone.

**Test data:**

```
GET /api/v1/projects/proj_public_A
GET /api/v1/ontologies/ont_public_A/classes
(no Authorization header)
```

---

#### TC-023: Change visibility without Owner role returns 403

**Priority:** High
**Type:** Negative

**Precondition:** `proj_private_A` exists; `u_editor_A` is a member but not Owner.

**Steps:**

1. Obtain JWT for `u_editor_A`.
2. `PUT /api/v1/projects/proj_private_A/visibility` with `{ "visibility": "public" }` and `Idempotency-Key`.
3. Inspect status code.
4. Verify visibility unchanged via `GET` as `u_owner_A`.

**Expected result:**

- HTTP `403` with `error_code: FORBIDDEN_INSUFFICIENT_ROLE`.
- Visibility remains the original value.

**Test data:**

```
PUT /api/v1/projects/proj_private_A/visibility
{ "visibility": "public" }
Idempotency-Key: bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb
Authorization: Bearer <jwt_u_editor_A>
```

---

### H. Fork RBAC

#### TC-024: Valid user forks a public project (positive)

**Priority:** High
**Type:** Positive

**Precondition:** `proj_public_A` exists with visibility `public`; `u_outsider_A` authenticated.

**Steps:**

1. Obtain JWT for `u_outsider_A`.
2. `POST /api/v1/projects/proj_public_A/fork` with `Idempotency-Key`.
3. Inspect status code and body.

**Expected result:**

- HTTP `201`.
- Response body contains `{ "project_id": "<new_uuid>", "ontology_id": "<new_uuid>", "upstream_project_id": "proj_public_A" }`.
- New fork is created with `visibility = private`, caller becomes `Owner`.

**Test data:**

```
POST /api/v1/projects/proj_public_A/fork
Idempotency-Key: cccccccc-cccc-cccc-cccc-cccccccccccc
Authorization: Bearer <jwt_u_outsider_A>
```

---

#### TC-025: Guest forks a public project (positive)

**Priority:** High
**Type:** Positive

**Precondition:** `proj_public_A` exists; `u_viewer_A` (Guest=10) authenticated.

**Steps:**

1. Obtain JWT for `u_viewer_A`.
2. `POST /api/v1/projects/proj_public_A/fork` with `Idempotency-Key`.
3. Inspect status code.

**Expected result:**

- HTTP `201`. Guest can fork public projects per spec "read access sufficient".

**Test data:**

```
POST /api/v1/projects/proj_public_A/fork
Idempotency-Key: dddddddd-dddd-dddd-dddd-dddddddddddd
Authorization: Bearer <jwt_u_viewer_A>
```

---

#### TC-026: Fork private project without read access returns 403, never 404

**Priority:** High
**Type:** Negative

**Precondition:** `proj_private_A` exists with visibility `private`; `u_outsider_A` is NOT a member; `proj_private_nonexistent` does not exist.

**Steps:**

1. Obtain JWT for `u_outsider_A`.
2. `POST /api/v1/projects/proj_private_A/fork` with `Idempotency-Key` (existing, no access).
3. `POST /api/v1/projects/proj_private_nonexistent/fork` with `Idempotency-Key` (does not exist).
4. Inspect both status codes.

**Expected result:**

- Both return HTTP `403` with `error_code: FORBIDDEN_OBJECT_NOT_FOUND_OR_ACCESS_DENIED`.
- Neither returns `404`.

**Test data:**

```
POST /api/v1/projects/proj_private_A/fork
POST /api/v1/projects/proj_private_nonexistent/fork
Idempotency-Key: eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee
Authorization: Bearer <jwt_u_outsider_A>
```

---

### I. Idempotency

#### TC-027: Write endpoint without Idempotency-Key returns 400

**Priority:** High
**Type:** Negative

**Precondition:** `u_owner_A` authenticated; `proj_private_A` exists.

**Steps:**

1. Obtain JWT for `u_owner_A`.
2. `POST /api/v1/projects/proj_private_A/members` with `{ "userId": "u_viewer_A", "role": "Viewer" }` — **without** `Idempotency-Key` header.
3. Inspect status code.

**Expected result:**

- HTTP `400` with `error_code: INVALID_IDEMPOTENCY_KEY`.
- No member added.

**Test data:**

```
POST /api/v1/projects/proj_private_A/members
{ "userId": "u_viewer_A", "role": "Viewer" }
(no Idempotency-Key header)
Authorization: Bearer <jwt_u_owner_A>
```

---

#### TC-028: Duplicate Idempotency-Key returns same result (idempotent)

**Priority:** Medium
**Type:** Positive

**Precondition:** `u_owner_A` authenticated; `proj_private_A` exists; `u_outsider_A` is NOT a member.

**Steps:**

1. Obtain JWT for `u_owner_A`.
2. Generate a fixed `Idempotency-Key: ffffffff-ffff-ffff-ffff-ffffffffffff`.
3. `POST /api/v1/projects/proj_private_A/members` with `{ "userId": "u_outsider_A", "role": "Viewer" }` and the fixed key.
4. Record first response status and body.
5. Send **identical** request again with the same `Idempotency-Key`.
6. Compare both responses.

**Expected result:**

- First request: HTTP `201`.
- Second request: HTTP `200` (or `201`, same body). No duplicate member created.
- Only one membership row for `u_outsider_A` exists.

**Test data:**

```
First:  POST /api/v1/projects/proj_private_A/members { "userId": "u_outsider_A", "role": "Viewer" }
Second: POST /api/v1/projects/proj_private_A/members { "userId": "u_outsider_A", "role": "Viewer" }
Both:   Idempotency-Key: ffffffff-ffff-ffff-ffff-ffffffffffff
        Authorization: Bearer <jwt_u_owner_A>
```

---

### J. Audit Logging

#### TC-029: Every write endpoint emits structured audit_events

**Priority:** High
**Type:** Positive

**Precondition:** `u_owner_A` authenticated; audit store accessible (PostgreSQL `audit_events` table or equivalent).

**Steps:**

1. Obtain JWT for `u_owner_A`.
2. Execute each of the following operations with unique `Idempotency-Key`:
   - `POST /api/v1/projects/proj_private_A/members` (add member)
   - `PUT /api/v1/projects/proj_private_A/members/{userId}` (change role)
   - `DELETE /api/v1/projects/proj_private_A/members/{userId}` (remove member)
   - `PUT /api/v1/projects/proj_private_A/visibility` (change visibility)
   - `POST /api/v1/projects/proj_private_A/policies` (create policy)
   - `DELETE /api/v1/projects/proj_private_A/policies/{policyId}` (delete policy)
   - `DELETE /api/v1/ontologies/{ontId}` (delete ontology)
3. Query `audit_events` for each operation.

**Expected result:**

Each audit row contains:

| Field | Value example |
|-------|---------------|
| `event` | `member.added`, `member.role_changed`, `member.removed`, `visibility.changed`, `policy.created`, `policy.deleted`, `project.deleted` |
| `reason` | (optional, can be null) |
| `user_id` | `u_owner_A` |
| `object_type` | `project` or `group` |
| `object_id` | the affected Project ID |
| `source_ip` | valid IP string |
| `trace_id` | non-empty OpenTelemetry trace ID |
| `timestamp` | ISO 8601 UTC, within 5 seconds of operation |

**Test data:**

```
Execute 7 write operations; query audit_events.assert each has all 8 fields with valid values.
```

---

### K. OMR Workflow (Protected main)

#### TC-030: Editor cannot commit directly to protected main under write_with_approval

**Priority:** Medium
**Type:** Negative

**Precondition:** `proj_private_A` has `main` protected with `write_with_approval`; `u_editor_A` authenticated.

**Steps:**

1. Obtain JWT for `u_editor_A`.
2. `POST /api/v1/ontologies/ont_A/commits` with body `{ "branch": "main", "message": "direct commit" }` and `Idempotency-Key`.
3. Inspect status code.

**Expected result:**

- HTTP `403` or endpoint redirects to OMR creation flow.
- Commit is NOT created on `main`.
- A proposal branch `proposal/...` may be created automatically.

**Test data:**

```
POST /api/v1/ontologies/ont_A/commits
{ "branch": "main", "message": "direct commit" }
Idempotency-Key: a1a1a1a1-a1a1-a1a1-a1a1-a1a1a1a1a1a1
Authorization: Bearer <jwt_u_editor_A>
```

---

#### TC-031: Maintainer can approve and merge OMR

**Priority:** Medium
**Type:** Positive

**Precondition:** An open OMR exists on `proj_private_A` awaiting review; `u_maintainer_A` authenticated.

**Steps:**

1. Obtain JWT for `u_maintainer_A`.
2. `GET /api/v1/projects/proj_private_A/merge-requests/{omrId}`.
3. Review the semantic diff.
4. `POST /api/v1/projects/proj_private_A/merge-requests/{omrId}/approve`.
5. `POST /api/v1/projects/proj_private_A/merge-requests/{omrId}/merge`.
6. Inspect all status codes.

**Expected result:**

- Read returns `200` with OMR details and diff.
- Approve returns `200`.
- Merge returns `200`. The feature branch is merged into `main`.
- Proposed changes are visible on `main`.

**Test data:**

```
GET /api/v1/projects/proj_private_A/merge-requests/{omrId}
POST /api/v1/projects/proj_private_A/merge-requests/{omrId}/approve
POST /api/v1/projects/proj_private_A/merge-requests/{omrId}/merge
Authorization: Bearer <jwt_u_maintainer_A>
```

---

### L. GraphQL RBAC

#### TC-032: GraphQL mutation — cross-tenant BOLA on ontology mutation

**Priority:** High
**Type:** Negative

**Precondition:** `u_editor_A` authenticated; `ont_B` exists in `tenant_B`.

**Steps:**

1. Obtain JWT for `u_editor_A`.
2. Send GraphQL mutation:
   ```graphql
   mutation {
     updateOntology(id: "ont_B", name: "hacked-name") {
       id
     }
   }
   ```
3. Inspect response.

**Expected result:**

- Response contains `errors` array with `extensions.code: FORBIDDEN_CROSS_TENANT_ACCESS`.
- `data` is `null` or `partial`.
- No changes made to `ont_B`.

**Test data:**

```
POST /api/v1/graphql
Authorization: Bearer <jwt_u_editor_A>
mutation { updateOntology(id: "ont_B", name: "hacked-name") { id } }
```

---

#### TC-033: GraphQL query — Viewer cannot query restricted ontology

**Priority:** High
**Type:** Negative

**Precondition:** `u_viewer_A` authenticated; `ont_B` exists in `tenant_B`.

**Steps:**

1. Obtain JWT for `u_viewer_A`.
2. Send GraphQL query:
   ```graphql
   query {
     ontology(id: "ont_B") {
       id
       name
       classes { name }
     }
   }
   ```
3. Inspect response.

**Expected result:**

- Response contains `errors` array with `extensions.code: FORBIDDEN_CROSS_TENANT_ACCESS`.
- No ontology data leaked.

**Test data:**

```
POST /api/v1/graphql
Authorization: Bearer <jwt_u_viewer_A>
query { ontology(id: "ont_B") { id name classes { name } } }
```

---

### M. SPARQL RBAC

#### TC-034: Cross-tenant SPARQL query returns 403

**Priority:** Medium
**Type:** Negative

**Precondition:** `u_viewer_A` authenticated; `ont_B` exists in `tenant_B`.

**Steps:**

1. Obtain JWT for `u_viewer_A`.
2. `POST /api/v1/sparql` with body `{ "query": "SELECT ?s ?p ?o WHERE { ?s ?p ?o }", "default-graph-uri": "ont_B" }`.
3. Inspect status code.

**Expected result:**

- HTTP `403` (or error in SPARQL response body).
- No ontology data from `ont_B` returned.

**Test data:**

```
POST /api/v1/sparql
{ "query": "SELECT ?s ?p ?o WHERE { ?s ?p ?o }", "default-graph-uri": "ont_B" }
Authorization: Bearer <jwt_u_viewer_A>
```

---

### N. MFA Categories

#### TC-035: Category A operation (tenant purge) without MFA returns 403

**Priority:** High
**Type:** Negative

**Precondition:** Authenticated as user with authority to perform category A operations; a test tenant exists that can be safely purged in test environment.

**Steps:**

1. Authenticate with password only (no TOTP).
2. Attempt to execute a category A destructive command (e.g., `DELETE /api/v1/tenants/{tenantId}` or `vedo-cli tenant purge`).
3. Inspect response.

**Expected result:**

- HTTP `403` or CLI error: MFA required.
- Operation is rejected.
- Audit log records denied attempt with reason `MFA_REQUIRED`.

**Test data:**

```
DELETE /api/v1/tenants/test-tenant-to-purge
(no TOTP / no second factor)
```

---

#### TC-036: Category B operation (restore) with cached MFA still requires fresh MFA

**Priority:** High
**Type:** Negative

**Precondition:** User authenticated with MFA for a session; session is still active.

**Steps:**

1. Authenticate with password + valid TOTP. Obtain session.
2. Perform a non-destructive operation (successful).
3. Attempt a category B operation (e.g., production restore) within the same session.
4. Inspect response.

**Expected result:**

- Operation is rejected with prompt for fresh MFA.
- MFA is **not** cached; each category A/B operation requires a new TOTP challenge.

---

#### TC-037: Service-account performs category C operation (positive)

**Priority:** High
**Type:** Positive

**Precondition:** Service-account token `sa_cat_C` with scope `["ontology:delete"]` and TTL 30 min.

**Steps:**

1. Authenticate as service-account `sa_cat_C` (no interactive MFA).
2. `DELETE /api/v1/ontologies/ont_A` with `Idempotency-Key`.
3. Inspect status code.

**Expected result:**

- HTTP `200` or `204`. Service-account with `ontology:delete` scope can perform category C operations.
- Audit log records `actor: sa_cat_C`, `event: ontology.deleted`.

**Test data:**

```
DELETE /api/v1/ontologies/ont_A
Authorization: Bearer <sa_cat_C_token>
Idempotency-Key: b2b2b2b2-b2b2-b2b2-b2b2-b2b2b2b2b2b2
```

---

### O. JIT/PAM — Privileged Access

#### TC-038: No standing access; privileged session expires after TTL

**Priority:** Medium
**Type:** Positive

**Precondition:** `Support Engineer` role configured with JIT/PAM; TTL = 60 min.

**Steps:**

1. Request privileged access via PAM workflow with a valid ticket ID.
2. Receive temporary credentials with TTL ≤ 60 min.
3. Execute a privileged operation (e.g., `vedo-cli support tenant-info --id tenant-456`).
4. Wait for TTL to expire (or attempt to use token after setting system clock beyond TTL).
5. Attempt the same operation again.

**Expected result:**

- First attempt: operation succeeds (within TTL).
- Second attempt (post-TTL): HTTP `401` or `403` — token expired or session revoked.
- Session recording captured 100% of the privileged session.
- Audit log contains ticket ID, scope, and operation timestamps.

---

### P. Break-Glass Emergency Admin

#### TC-039: Emergency Admin login works with Keycloak down

**Priority:** High
**Type:** Positive

**Precondition:** Shamir parts combined to produce valid Emergency Admin password; Keycloak service is stopped or unreachable.

**Steps:**

1. Simulate Keycloak outage (stop Keycloak container or block its port).
2. `POST /api/v1/emergency/login` with `{ "username": "emergency-admin", "password": "<reconstructed-shamir-password>" }`.
3. Inspect status code.
4. `GET /api/v1/health` (emergency read-only).
5. `POST /api/v1/restore` (emergency restore).

**Expected result:**

- Login: HTTP `200` with a temporary JWT.
- Emergency read-only and restore: HTTP `200`.
- All operations logged in immutable audit log.
- Security team notified (check notification channel).

**Test data:**

```
POST /api/v1/emergency/login
{ "username": "emergency-admin", "password": "<shamir-password>" }
```

---

#### TC-040: Emergency Admin TTL enforced — session expires after 30 min

**Priority:** High
**Type:** Negative

**Precondition:** Emergency Admin logged in via `/emergency/login`.

**Steps:**

1. Log in as Emergency Admin. Record `expires_at`.
2. Perform one valid operation (success).
3. Wait 31 minutes (or fast-forward if environment supports it).
4. Attempt the same operation again.

**Expected result:**

- Operation after 30 min returns HTTP `401`.
- Session auto-terminated.
- Audit log records session expiry.

---

### Q. Destructive Command Guardrails

#### TC-041: G1 (typed confirmation) — branch delete requires exact confirmation input

**Priority:** Medium
**Type:** Negative

**Precondition:** CLI accessible; test branch `branch_to_delete` exists.

**Steps:**

1. Run `vedo-cli ontology branch delete --id branch_to_delete`.
2. When prompted for typed confirmation, enter an incorrect value (e.g., wrong branch name).
3. Run the command again with the correct typed confirmation.

**Expected result:**

- First attempt: operation aborted, error message "Confirmation does not match. Operation cancelled."
- Second attempt: operation succeeds (branch deleted).
- G2+ guardrails additionally check environment name and network.

---

### R. Public Browse API

#### TC-042: Public Browse API serves published ontology without auth

**Priority:** Medium
**Type:** Positive

**Precondition:** `ont_public_A` has a published snapshot; Public Browse API is running.

**Steps:**

1. Navigate to Public Browse URL (without any JWT).
2. Navigate through classes, properties, and class hierarchy.
3. Attempt to perform a write operation (e.g., `POST` to any path under `/browse/`).
4. Send high-frequency requests to test rate limiting.

**Expected result:**

- Navigation loads ontology content with HTTP `200`.
- Write operation returns HTTP `405` (method not allowed) or `403`.
- After exceeding rate limit, subsequent requests return HTTP `429`.
- No arbitrary Cypher/SPARQL accepted (only allowlisted templates).

---

### S. ABAC Most-Specific-Wins

#### TC-043: Specific class policy overrides permissive Project-level permission

**Priority:** Medium
**Type:** Positive

**Precondition:** `proj_private_A` has a Project-level permission allowing `Editor` to delete all classes; an ABAC policy exists on class `RestrictedClass` with `effect: deny, action: delete`.

**Steps:**

1. Obtain JWT for `u_editor_A`.
2. `DELETE /api/v1/ontologies/ont_A/classes/RestrictedClass` (class with specific deny policy).
3. `DELETE /api/v1/ontologies/ont_A/classes/RegularClass` (no specific deny policy).
4. Inspect both status codes.

**Expected result:**

- RestrictedClass: HTTP `403` (specific deny wins over general allow).
- RegularClass: HTTP `200` or `204` (general allow applies).
- Most-specific-wins rule enforced.

**Test data:**

```
DELETE /api/v1/ontologies/ont_A/classes/RestrictedClass
DELETE /api/v1/ontologies/ont_A/classes/RegularClass
Idempotency-Key: c3c3c3c3-c3c3-c3c3-c3c3-c3c3c3c3c3c3
Authorization: Bearer <jwt_u_editor_A>
```

---

### T. Service-Account Restrictions

#### TC-044: Service-account with scope-limited token cannot perform category A operation

**Priority:** High
**Type:** Negative

**Precondition:** Service-account token `sa_cat_D` with scope `["export", "diagnostics"]` (category D only).

**Steps:**

1. Authenticate as `sa_cat_D`.
2. Attempt `DELETE /api/v1/backups/some-backup` (category A — backup delete).
3. Attempt `POST /api/v1/restore` (category B — production restore).
4. Inspect both status codes.

**Expected result:**

- Both requests return HTTP `403` with `error_code: FORBIDDEN_INSUFFICIENT_SCOPE`.
- Service-account cannot escalate beyond its allowlist scope.

**Test data:**

```
DELETE /api/v1/backups/some-backup
POST /api/v1/restore
Authorization: Bearer <sa_cat_D_token>
```

---

### U. Pagination Compliance

#### TC-045: per_page > 100 is rejected

**Priority:** Low
**Type:** Negative

**Precondition:** `u_viewer_A` authenticated; sufficient test data exists.

**Steps:**

1. Obtain JWT for `u_viewer_A`.
2. `GET /api/v1/projects?per_page=200`.
3. Inspect status code.

**Expected result:**

- HTTP `400` or `422` with error message indicating `per_page` exceeds maximum (100).
- Documentation in `.ai-factory/references/gitlab-projects-groups-api.md` § "Pagination".

**Test data:**

```
GET /api/v1/projects?per_page=200
Authorization: Bearer <jwt_u_viewer_A>
```

---
