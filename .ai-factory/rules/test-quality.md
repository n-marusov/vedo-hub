# Test Quality Rules

> Enforced by `$aif-test-quality` skill. Referenced by `$aif-plan`, `$aif-implement`, `$aif-verify`.

## Quality Thresholds

| Grade | TQS | RCS | Gate |
|-------|-----|-----|------|
| 🟢 gold | ≥ 9.5 | ≥ 9.0 | Exceeds expectations |
| 🔵 silver | ≥ 8.0 | ≥ 7.0 | Meets high standards |
| 🟡 bronze | ≥ 6.0 | ≥ 5.0 | Minimum acceptable |
| 🔴 fail | < 6.0 | < 5.0 | BLOCK — must improve |
| ⚪ na | N/A | N/A | No tests found (informational) |

**Minimum gate for `$aif-verify`:** TQS ≥ bronze (6.0) AND no B1–B7 violations.

## Blocking Anti-Patterns (B1–B7)

These patterns **instantly BLOCK** a test file (TQS = 0):

| ID | Pattern | Languages |
|----|---------|-----------|
| B1 | Tautology assertion (`assertTrue(true)`, `assert!(true)`) | All |
| B2 | `sleep`/`thread::sleep` in test body | All |
| B3 | Empty catch block (`except: pass`, empty `if err != nil {}`) | All |
| B4 | Process exit in test (`os.Exit`, `sys.exit`, `process.exit`) | All |
| B5 | Test function with zero assertions | All |
| B6 | Test code in production source file (non-idiomatic) | Go, TS, Python |
| B7 | `__main__` + assert pattern (Python) | Python |

**Exception:** Rust `#[cfg(test)]` modules are idiomatic — B6 does NOT apply to Rust.

## BDD Naming Convention

All test functions MUST follow BDD-style naming:

- **Go / Rust / Python:** `[Condition]_[Action]_[ExpectedResult]`
  ```
  TestCreateClass_ValidInput_ReturnsClassID
  test_create_class_valid_input_returns_class_id
  ```
- **TypeScript:** `'<expected> when <condition>'` or `'should <expected> when <condition>'`
  ```
  it('should return class ID when input is valid')
  ```

## Test Annotations

Every test that validates a requirement MUST include a traceability annotation:

```go
// Go
// Validates: REQ-FUN-API-create-class
func TestCreateClass_ValidInput_ReturnsClassID(t *testing.T) { ... }
```

```rust
// Rust
// Validates: REQ-FUN-API-create-class
#[test]
fn should_return_class_id_when_input_is_valid() { ... }
```

```python
# Python
# Validates: REQ-FUN-API-create-class
def test_create_class_valid_input_returns_class_id(): ...
```

```typescript
// TypeScript
// Validates: REQ-FUN-API-create-class
it('should return class ID when input is valid', () => { ... })
```

## Test Structure (given/when/then)

All tests MUST follow the arrange-act-assert pattern with comments:

```go
func TestCreateClass_ValidInput_ReturnsClassID(t *testing.T) {
    // given: a valid ontology class definition
    req := &ontology.CreateClassRequest{Name: "Person"}

    // when: the class is created
    result, err := svc.CreateClass(ctx, req)

    // then: a class ID is returned without error
    require.NoError(t, err)
    assert.NotEmpty(t, result.ClassID)
}
```

## traceability.ttl Requirements

1. **Every test file** must have a corresponding `vdo:TestSuite` entry in `.ai-factory/traceability/traceability.ttl`
2. **Every `// Validates:` annotation** must have a matching `vdo:validates` triple in the TTL
3. **P0 requirements** MUST have at least one test covering them (orphan P0 = WARN in quality report)
4. **When creating/modifying tests** → update `traceability.ttl` immediately (per RULES.md Traceability §1)

## Service-Level Expectations

| Service | Current TQS | Expected | Notes |
|---------|------------|----------|-------|
| api-gateway | ~7.25 (silver) | ≥ silver | 6 test files, good baseline |
| ontology-service | TBD | ≥ bronze | Rust #[cfg(test)] analyzed |
| versioning-service | TBD | ≥ bronze | Rust #[cfg(test)] analyzed |
| frontend | N/A (0 tests) | N/A | No BLOCK, but new tests must reach bronze |
| document-extractor | N/A (0 tests) | N/A | No BLOCK, but new tests must reach bronze |

## Integration in AI Factory Pipeline

### $aif-plan

When planning features with tests:
- Add TQS gate to **acceptance criteria** (if test files are in the plan)
- Add "Update `traceability.ttl`" to implementation **steps**
- Reference this rules file: `.ai-factory/rules/test-quality.md`

### $aif-implement

When implementing tests:
- Use BDD naming (see above)
- Add `// Validates: REQ-<id>` annotations
- Include given/when/then comments
- Avoid all B1–B7 anti-patterns
- Update `traceability.ttl` after test creation

### $aif-verify

Test Quality Gate runs as **Step 1** (before task audit, code quality, consistency checks):
- Fail-fast: bad tests → BLOCK immediately
- TQS and RCS appear in the verification report header
- BLOCK verdict prevents moving to subsequent steps
