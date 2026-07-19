## Iteration 1 Plan

**Focus:** Run all verifiable acceptance criteria checks against the auth-service/org implementation. Identify which criteria pass, which fail, and which require infrastructure not available in this environment.

### Execution:
1. Run all `go test` suites in auth-service/org that use MemStore (no PostgreSQL)
2. Check MVP role vocabulary tests (org_roles_test.go)
3. Check hierarchy depth and move scope tests (org_hierarchy_test.go)  
4. Check existing gRPC contract tests (org_grpc_test.go)
5. Check visibility, cross-tenant, cycle detection, last-owner protection, role inheritance via existing tests
6. Check audit logging via code review
7. Check idempotency middleware code
8. Check traceability.ttl updates
9. Check build/lint

### After evaluation:
- If critical (fail) criteria fail → CRITIQUE → REFINE (fix code)
- If all Phase A criteria pass → advance to Phase B (traceability, docs)
