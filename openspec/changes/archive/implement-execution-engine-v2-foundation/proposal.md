# Proposal: Execution Engine v2 Foundation (Phase 1)

## Status: pending

## Why

**Current Problem:**

The execution engine v2 features (sandbox resource limits, code-mode VM integration, safe API surface) are implemented but lack comprehensive unit test coverage. Without tests, these critical features are at risk of regression and cannot be safely extended.

**Quantified Impact:**
- 16 test files need creation in `internal/execution/`
- ~28 test scenarios covering sandbox limits, error handling, VM integration, and API surface
- 0% current test coverage for v2 features (estimated)
- Risk: High - production features without tests

**From Registry Analysis:**
The `execution-engine-v2` change has 108 total tasks across 4 phases. Phase 1 (Foundation) with 28 tasks is marked "ready" and focuses exclusively on unit tests for existing v2 features.

## What Changes

### Before

```
internal/execution/
├── sandbox.go          # Implemented, no tests
├── codemode.go         # Implemented, no tests
├── limits.go           # Implemented, no tests
└── safeapi.go          # Implemented, no tests
```

Test coverage: 0% for v2 features

### After

```
internal/execution/
├── sandbox.go
├── sandbox_test.go     # NEW - 8 test cases
├── codemode.go
├── codemode_test.go    # NEW - 8 test cases
├── limits.go
├── limits_test.go      # NEW - 4 test cases
├── safeapi.go
├── safeapi_test.go     # NEW - 8 test cases
└── testdata/           # NEW - Test fixtures
    ├── sandbox/
    ├── codemode/
    └── safeapi/
```

Test coverage: >80% for v2 features

### Key Components

| File | Description |
|------|-------------|
| `internal/execution/sandbox_test.go` | Unit tests for sandbox resource limits and isolation |
| `internal/execution/codemode_test.go` | Unit tests for JavaScript VM integration |
| `internal/execution/limits_test.go` | Unit tests for resource limit enforcement |
| `internal/execution/safeapi_test.go` | Unit tests for safe API surface |
| `internal/execution/testdata/sandbox/` | Test fixtures for sandbox tests |
| `internal/execution/testdata/codemode/` | Test fixtures for code-mode tests |
| `internal/execution/testdata/safeapi/` | Test fixtures for safe API tests |

**Test Categories:**

1. **Sandbox Resource Limits (8 tasks)**
   - Memory limit enforcement
   - CPU limit enforcement
   - Timeout enforcement
   - Concurrent sandbox isolation

2. **Sandbox Error Handling (4 tasks)**
   - Panic recovery in sandbox
   - Resource exhaustion errors
   - Timeout errors
   - Sandbox cleanup on error

3. **Code-Mode VM Integration (8 tasks)**
   - JavaScript VM initialization (goja)
   - Code execution in VM
   - VM memory limits
   - VM cleanup

4. **Safe API Surface (8 tasks)**
   - Allowed function exposure
   - Blocked function access
   - Function argument validation
   - Return value handling

## Impact

**Breaking Changes:** None - adding tests only

**Performance Impact:** None at runtime (tests don't affect production)

**Benefits:**
- Enables safe refactoring of execution engine
- Documents expected behavior through tests
- Catches regressions before they reach production
- Required foundation for Phases 2-4

**Dependencies:**
- Requires: `add-execution-engine` (v1 features) - COMPLETED
- Enables: Phase 2 (Context Management), Phase 3 (State Persistence), Phase 4 (Integration)

## Success Criteria

- [ ] All 28 Phase 1 tasks completed
- [ ] Sandbox resource limit tests pass
- [ ] Sandbox error handling tests pass
- [ ] Code-mode VM integration tests pass
- [ ] Safe API surface tests pass
- [ ] Test coverage >80% for v2 features
- [ ] `go test ./internal/execution/... -v` passes
- [ ] `make test` passes with new tests
- [ ] No regressions in existing tests

## Risk Assessment

| Risk | Severity | Mitigation |
|------|----------|------------|
| Tests reveal existing bugs | Medium | Document bugs, fix critical ones, mark non-critical as known issues |
| VM testing complexity | Medium | Use goja test helpers, mock where appropriate |
| Resource limit testing flaky | Low | Use generous timeouts, retry logic in tests |
| Test fixtures maintenance | Low | Keep fixtures minimal, document purpose |

## Alternatives Considered

1. **Skip Phase 1, start with Phase 2** - Rejected: Building on untested foundation is risky
2. **Write integration tests only** - Rejected: Unit tests catch issues faster and are more maintainable
3. **Use mocks for VM** - Rejected: Testing actual goja behavior is more valuable

## Phase Context

This proposal covers **Phase 1 only** (Foundation - 28 tasks). The full execution-engine-v2 change has:

- **Phase 1: Foundation** (28 tasks) - Unit tests for existing v2 features ⭐ THIS PROPOSAL
- **Phase 2: Context Management** (28 tasks) - Context summarization and limit handling
- **Phase 3: State Persistence** (36 tasks) - Execution state persistence and interrupt/resume
- **Phase 4: Integration** (16 tasks) - End-to-end testing and validation

**Critical Path:** Phases 1→2→4 (Phase 3 can use file fallback if add-boltdb-state-system blocked)

## Related Documentation

- `openspec/changes/archive/2026-01-26-execution-engine-v2/` - Original v2 proposal (archived)
- `internal/execution/` - Target package for tests
- `openspec/registry.yaml` - Change tracking registry
