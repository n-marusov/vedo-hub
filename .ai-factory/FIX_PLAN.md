# Fix Plan: M2 AI Ontology Creation — Critical Issues

**Problem:** 15 blocking critical issues found in `$aif-review` of branch `feature/ai-ontology-creation` (17 commits, +36,094/-555 lines). Issues span: gRPC lifecycle bug, data integrity violations, security bypasses, dead code, and architectural misalignment.
**Created:** 2026-07-16 16:30

## Analysis

Root causes grouped by category:

1. **gRPC lifecycle** (`routes.go:43`): `defer grpcPool.Close()` inside `RegisterRoutes()` fires before `srv.ListenAndServe()`. All gRPC calls from handlers silently fail. HTTP proxy fallback masked the bug — only new gRPC clients are affected.

2. **ApplySequence atomicity** (`apply_sequence.rs:136-137`): Condition `steps_failed > 0 && steps_applied == 0` commits partial failures when at least one step succeeded. Data corruption.

3. **Cross-ontology scope** (`sequence_repo.rs:376-423`): `set_domain`/`set_range` MATCH without `ontology_id` filter silently modifies entities in unrelated ontologies.

4. **LLM Policy bypass**: `X-Admin-Override` (no JWT verification) and `X-Ontology-Visibility` (client-supplied, no server check) allow privilege escalation. Any user can exfiltrate private ontology data to external LLM providers.

5. **Prompt injection dead code**: `ValidateLLMOutput` and `containsPromptInjectionPatterns` defined but never called. `SanitizeLLMInput` bypassable via Unicode homoglyphs. No input length limits.

6. **gRPC auth**: No `PerRPCCredentials`/`UnaryInterceptor` — all downstream calls unauthenticated. No TLS.

7. **Silent failures**: `aiProvider` nil → panic on first AI request. `HandleApplyTemplate` returns 200 but does nothing. Document-extractor routes proxy to wrong service.

## Fix Steps

### Group 1: Critical Infrastructure (blocks everything — fix first)

1. [ ] **Fix `grpcPool.Close()` lifecycle.** Move `defer grpcPool.Close()` from `RegisterRoutes()` to `main()` graceful shutdown path.
   - File: `src/services/api-gateway/routes.go:43`
   - Also: `src/services/api-gateway/main.go` — add pool.Close() in shutdown handler
   - Verify: start gateway, call a gRPC-backed endpoint, confirm connection alive

2. [ ] **Fix document-extractor routes.** Redirect `/ontologies/:id/documents/extract` and `/documents/extract/batch` from `ontologyProxy` to a dedicated `documentExtractorProxy`.
   - File: `src/services/api-gateway/routes.go:149-150`
   - Create: `mustNewProxy(getEnv("DOCUMENT_EXTRACTOR_URL", "http://localhost:8092"), "document-extractor")`

3. [ ] **Fix nil `aiProvider` panic.** Add nil-guard in every AI handler: `if h.provider == nil { c.JSON(503, ...); return }`.
   - Files: `nl_to_owl_handler.go`, `ai_completion_handler.go`, `refinement_handler.go`, `template_handler.go`
   - Alternative: return fatal error at startup if AI routes registered but provider unavailable

### Group 2: Security — LLM Policy Bypass

4. [ ] **Fix `X-Admin-Override` privilege escalation.** Remove header-based override or verify against JWT claims (admin role).
   - File: `src/services/api-gateway/middleware/llm_policy_router.go:209`
   - Extract JWT claims from Gin context (set by auth middleware), check `auth_roles` for admin

5. [ ] **Fix `X-Ontology-Visibility` forgery.** Resolve visibility from ontology-service gRPC, not from untrusted client header.
   - File: `src/services/api-gateway/middleware/llm_policy_router.go:149-153`
   - Call `ontologyGrpc.GetOntology(ontologyID)` and read actual visibility field
   - Cache visibility in Redis (existing Redis infrastructure) with 60s TTL

### Group 3: Security — Prompt Injection

6. [ ] **Activate `ValidateLLMOutput`.** Call on every `completion.Text` before parsing/returning. Remove dead-code status.
   - Files: `nl_to_owl_handler.go`, `ai_completion_handler.go`, `refinement_handler.go`
   - Add call after LLM response: `if err := ValidateLLMOutput(completion.Text); err != nil { ... }`

7. [ ] **Harden `SanitizeLLMInput` against Unicode bypass.** Add NFC/NFKD normalization and strip zero-width characters before pattern matching.
   - File: `src/services/api-gateway/handlers/prompt_injection.go:12-20`
   - Use `golang.org/x/text/unicode/norm` for normalization
   - Strip: U+200B (ZWSP), U+200C (ZWNJ), U+200D (ZWJ), U+FEFF (BOM)

8. [ ] **Sanitize user identifiers in AI completion prompts.** Apply `SanitizeLLMInput` to `req.SourceClassID`, `req.TargetClassID`, `req.ClassID` before prompt interpolation.
   - File: `src/services/api-gateway/handlers/ai_completion_handler.go:221-222,71,151`

9. [ ] **Add input length limits.** Add `MaxTextLength` (50KB) and `MaxFeedbackLength` (10KB) constants, enforce before LLM processing.
   - Files: `nl_to_owl_handler.go:51-53`, `refinement_handler.go:58-62`
   - Return 400 with `PAYLOAD-TOO-LARGE` error code

### Group 4: Data Integrity — ApplySequence

10. [ ] **Fix ApplySequence atomicity.** Change condition from `steps_failed > 0 && steps_applied == 0` to `steps_failed > 0` → rollback on ANY failure.
    - File: `src/services/ontology-service/src/services/apply_sequence.rs:136-137`
    - Verify: unit test with mixed success/failure steps — transaction must roll back entirely

11. [ ] **Fix cross-ontology `set_domain`/`set_range`.** Add `ontology_id` filter to MATCH clauses.
    - File: `src/services/ontology-service/src/repositories/sequence_repo.rs:376-423`
    - Add parameter `ontology_id: &str` to both functions
    - Add `AND p.ontology_id = $ontology_id` to all MATCH clauses

12. [ ] **Implement `HandleApplyTemplate` actual gRPC call.** Replace stub with `ontologyGrpc.ApplySequence()` call using template steps.
    - File: `src/services/api-gateway/handlers/template_handler.go:156-220`
    - Convert `tpl.Steps` to proto `SequenceStep` messages, call `ApplySequence`
    - Return commit ID and entity count from actual gRPC response

### Group 5: gRPC Security — Auth & TLS

13. [ ] **Implement gRPC auth token propagation.** Add `grpc.UnaryClientInterceptor` that reads JWT from `context.Context` and adds `authorization` metadata.
    - File: `src/services/api-gateway/proxy/grpc_client.go:54-58`
    - Auth middleware stores JWT token in `context.Context` via `context.WithValue`
    - Interceptor extracts and adds `authorization: Bearer <token>` to gRPC metadata
    - Related: `src/services/api-gateway/auth/auth.go:265-266` — store in context, not just headers

14. [ ] **Enable TLS for gRPC.** Replace `insecure.NewCredentials()` with `credentials.NewTLS()`.
    - File: `src/services/api-gateway/proxy/grpc_client.go:55`
    - Phase 1: server cert verification only (dev: self-signed, prod: proper CA)
    - Add `GRPC_TLS_ENABLED` feature flag for backward compatibility

### Group 6: Code Quality & Cleanup

15. [ ] **Remove `ontology_context` dead parameter** from document-extractor routes. Not in plan, not used, potential injection vector.
    - File: `src/services/document-extractor/api/routes.py:126`
    - Remove `ontology_context: str | None = Form(None)` from function signature

16. [ ] **Fix `llm_policy_router.go` route matching scope.** Restrict `isAIRoute` to exact AI prefixes, not the broad `/api/v1/ontologies/` prefix.
    - File: `src/services/api-gateway/middleware/llm_policy_router.go:104-118`
    - Match only: `generate-from-text`, `ai/suggest-*`, `ai/refine`, `documents/extract`

17. [ ] **Fix LLM library `context.Context` propagation.** Use `http.NewRequestWithContext(ctx.BaseCtx, ...)` in both providers.
    - Files: `src/services/shared/llm/providers/anthropic.go:107`, `src/services/shared/llm/providers/openai_compatible.go:112`
    - Add response body size limit: wrap `resp.Body` with `io.LimitReader`

## Files to Modify

- `src/services/api-gateway/routes.go` — grpcPool lifecycle, doc-extractor proxy, nil provider guard
- `src/services/api-gateway/main.go` — grpcPool.Close() in shutdown
- `src/services/api-gateway/middleware/llm_policy_router.go` — X-Admin-Override, X-Ontology-Visibility, route matching
- `src/services/api-gateway/handlers/prompt_injection.go` — Unicode normalization, activation
- `src/services/api-gateway/handlers/nl_to_owl_handler.go` — ValidateLLMOutput call, input limits, nil guard
- `src/services/api-gateway/handlers/ai_completion_handler.go` — sanitize IDs, nil guard, ValidateLLMOutput
- `src/services/api-gateway/handlers/refinement_handler.go` — input limits, nil guard
- `src/services/api-gateway/handlers/template_handler.go` — actual ApplySequence call, nil guard
- `src/services/ontology-service/src/services/apply_sequence.rs` — atomicity fix
- `src/services/ontology-service/src/repositories/sequence_repo.rs` — ontology_id filter, Neo4j index
- `src/services/api-gateway/proxy/grpc_client.go` — TLS, auth interceptor, keepalive
- `src/services/api-gateway/auth/auth.go` — store JWT in context.Context
- `src/services/document-extractor/api/routes.py` — remove ontology_context
- `src/services/shared/llm/providers/anthropic.go` — context propagation, body size limit
- `src/services/shared/llm/providers/openai_compatible.go` — context propagation, body size limit

## Risks & Considerations

- **grpcPool lifecycle fix** must be verified with integration test — gRPC connection alive after startup
- **ApplySequence atomicity fix** changes behavior: previously partially-applied sequences will now fully roll back. Existing data in dev/staging may need cleanup.
- **X-Admin-Override removal** may break admin workflows that rely on the header. Ensure JWT-based alternative works.
- **TLS for gRPC** adds operational complexity (cert management). Keep `GRPC_TLS_ENABLED=false` default for dev.
- **Prompt injection Unicode hardening** may have false positives on legitimate Unicode content. Test with real multilingual ontology terms.

## Test Coverage

- Integration test: gRPC pool alive after gateway startup
- Unit test: ApplySequence rollback on any step failure
- Unit test: set_domain/set_range with cross-ontology entities (should NOT modify)
- Security test: X-Admin-Override rejected without admin JWT
- Security test: X-Ontology-Visibility ignored, server-side visibility used
- Unit test: SanitizeLLMInput with Unicode homoglyphs (İgnore, zero-width chars)
- Unit test: nil aiProvider returns 503, not panic
- Unit test: ValidateLLMOutput catches embedded <script> in LLM response
