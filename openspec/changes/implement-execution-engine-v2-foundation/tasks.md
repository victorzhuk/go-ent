# Tasks: Execution Engine v2 Foundation (Phase 1)

## Status: pending

## Dependencies

```
T1.1 ─┬→ T1.2 ─┬→ T2.1 ─┬→ T2.2 ─┬→ T3.1 ─┬→ T3.2 ─┬→ T4.1
      │        │        │        │        │        └→ T4.2
      │        │        │        │        └→ T3.3 ─┬→ T4.3
      │        │        │        │                 └→ T4.4
      │        │        │        └→ T3.4
      │        │        └→ T2.3
      │        └→ T2.4
      └→ T1.3 ─┬→ T1.4
               └→ T1.5
```

- **Phase 1** (Foundation): No dependencies - can start immediately
- Tasks within Phase 1 have internal dependencies as shown above
- Phase 1 completion enables Phase 2 (Context Management)

## Phase 1: Foundation - Unit Tests (28 tasks)

**Objective**: Test existing v2 features (sandbox, code-mode) that are already implemented
**Effort**: 10-12 hours | **Dependencies**: None | **Status**: Ready to start

---

### T1.1: Sandbox Resource Limits - Memory and CPU
**Files:** `internal/execution/sandbox_test.go`
**Dependencies:** None
**Parallel with:** T1.2, T1.3

Steps:
- [ ] 1.1.1 Create `internal/execution/sandbox_test.go` file
- [ ] 1.1.2 Implement `TestSandbox_MemoryLimit` - Verify memory limit enforcement
- [ ] 1.1.3 Implement `TestSandbox_CPULimit` - Verify CPU limit enforcement
- [ ] 1.1.4 Add test fixtures in `internal/execution/testdata/sandbox/`

Validation:
- [ ] Tests compile: `go build ./internal/execution/...`
- [ ] Tests pass: `go test ./internal/execution -run TestSandbox_MemoryLimit -v`
- [ ] Tests pass: `go test ./internal/execution -run TestSandbox_CPULimit -v`

---

### T1.2: Sandbox Resource Limits - Timeout and Isolation
**Files:** `internal/execution/sandbox_test.go`
**Dependencies:** T1.1
**Parallel with:** None

Steps:
- [ ] 1.2.1 Implement `TestSandbox_TimeoutEnforcement` - Verify timeout handling
- [ ] 1.2.2 Implement `TestSandbox_ConcurrentIsolation` - Verify concurrent sandbox isolation
- [ ] 1.2.3 Add helper functions for sandbox test setup

Validation:
- [ ] Tests pass: `go test ./internal/execution -run TestSandbox_Timeout -v`
- [ ] Tests pass: `go test ./internal/execution -run TestSandbox_Concurrent -v`

---

### T1.3: Sandbox Error Handling
**Files:** `internal/execution/sandbox_test.go`
**Dependencies:** T1.1
**Parallel with:** T1.2

Steps:
- [ ] 1.3.1 Implement `TestSandbox_PanicRecovery` - Verify panic recovery in sandbox
- [ ] 1.3.2 Implement `TestSandbox_ResourceExhaustion` - Verify resource exhaustion errors
- [ ] 1.3.3 Implement `TestSandbox_TimeoutErrors` - Verify timeout error handling
- [ ] 1.3.4 Implement `TestSandbox_CleanupOnError` - Verify sandbox cleanup on error

Validation:
- [ ] All sandbox error tests pass
- [ ] No goroutine leaks in tests
- [ ] Cleanup verified with resource monitoring

---

### T1.4: Code-Mode VM Integration - Initialization and Execution
**Files:** `internal/execution/codemode_test.go`
**Dependencies:** T1.2, T1.3
**Parallel with:** None

Steps:
- [ ] 1.4.1 Create `internal/execution/codemode_test.go` file
- [ ] 1.4.2 Implement `TestCodeMode_VMInitialization` - Test JavaScript VM initialization (goja)
- [ ] 1.4.3 Implement `TestCodeMode_CodeExecution` - Test code execution in VM
- [ ] 1.4.4 Add test fixtures in `internal/execution/testdata/codemode/`

Validation:
- [ ] Tests compile: `go build ./internal/execution/...`
- [ ] VM initialization tests pass
- [ ] Code execution tests pass

---

### T1.5: Code-Mode VM Integration - Limits and Cleanup
**Files:** `internal/execution/codemode_test.go`
**Dependencies:** T1.4
**Parallel with:** None

Steps:
- [ ] 1.5.1 Implement `TestCodeMode_VMMemoryLimits` - Test VM memory limits
- [ ] 1.5.2 Implement `TestCodeMode_VMCleanup` - Test VM cleanup
- [ ] 1.5.3 Add helper functions for VM test setup

Validation:
- [ ] Memory limit tests pass
- [ ] Cleanup tests pass (verify no memory leaks)

---

### T2.1: Safe API Surface - Allowed Functions
**Files:** `internal/execution/safeapi_test.go`
**Dependencies:** T1.5
**Parallel with:** None

Steps:
- [ ] 2.1.1 Create `internal/execution/safeapi_test.go` file
- [ ] 2.1.2 Implement `TestSafeAPI_AllowedFunctions` - Test allowed function exposure
- [ ] 2.1.3 Implement `TestSafeAPI_BlockedFunctions` - Test blocked function access
- [ ] 2.1.4 Add test fixtures in `internal/execution/testdata/safeapi/`

Validation:
- [ ] Tests compile: `go build ./internal/execution/...`
- [ ] Allowed function tests pass
- [ ] Blocked function tests pass (verify access denied)

---

### T2.2: Safe API Surface - Validation and Returns
**Files:** `internal/execution/safeapi_test.go`
**Dependencies:** T2.1
**Parallel with:** None

Steps:
- [ ] 2.2.1 Implement `TestSafeAPI_ArgumentValidation` - Test function argument validation
- [ ] 2.2.2 Implement `TestSafeAPI_ReturnValueHandling` - Test return value handling
- [ ] 2.2.3 Add helper functions for safe API test setup

Validation:
- [ ] Argument validation tests pass
- [ ] Return value tests pass

---

### T2.3: Resource Limits - Core Functionality
**Files:** `internal/execution/limits_test.go`
**Dependencies:** T1.5
**Parallel with:** T2.1

Steps:
- [ ] 2.3.1 Create `internal/execution/limits_test.go` file
- [ ] 2.3.2 Implement `TestLimits_MemoryTracking` - Test memory usage tracking
- [ ] 2.3.3 Implement `TestLimits_CPUTracking` - Test CPU usage tracking

Validation:
- [ ] Tests compile: `go build ./internal/execution/...`
- [ ] Memory tracking tests pass
- [ ] CPU tracking tests pass

---

### T2.4: Resource Limits - Enforcement and Errors
**Files:** `internal/execution/limits_test.go`
**Dependencies:** T2.3
**Parallel with:** T2.2

Steps:
- [ ] 2.4.1 Implement `TestLimits_Enforcement` - Test limit enforcement
- [ ] 2.4.2 Implement `TestLimits_ErrorHandling` - Test limit error handling

Validation:
- [ ] Limit enforcement tests pass
- [ ] Error handling tests pass

---

### T3.1: Sandbox Integration Tests - Resource Scenarios
**Files:** `internal/execution/sandbox_integration_test.go`
**Dependencies:** T2.2, T2.4
**Parallel with:** None

Steps:
- [ ] 3.1.1 Create `internal/execution/sandbox_integration_test.go` file
- [ ] 3.1.2 Implement `TestSandboxIntegration_MemoryPressure` - Test under memory pressure
- [ ] 3.1.3 Implement `TestSandboxIntegration_CPUPressure` - Test under CPU pressure

Validation:
- [ ] Integration tests compile
- [ ] Memory pressure tests pass
- [ ] CPU pressure tests pass

---

### T3.2: Sandbox Integration Tests - Concurrent Scenarios
**Files:** `internal/execution/sandbox_integration_test.go`
**Dependencies:** T3.1
**Parallel with:** None

Steps:
- [ ] 3.2.1 Implement `TestSandboxIntegration_ConcurrentExecution` - Test concurrent sandbox execution
- [ ] 3.2.2 Implement `TestSandboxIntegration_ResourceSharing` - Test resource sharing between sandboxes

Validation:
- [ ] Concurrent execution tests pass
- [ ] Resource sharing tests pass

---

### T3.3: Code-Mode Integration Tests - VM Scenarios
**Files:** `internal/execution/codemode_integration_test.go`
**Dependencies:** T3.1
**Parallel with:** T3.2

Steps:
- [ ] 3.3.1 Create `internal/execution/codemode_integration_test.go` file
- [ ] 3.3.2 Implement `TestCodeModeIntegration_ComplexScript` - Test complex script execution
- [ ] 3.3.3 Implement `TestCodeModeIntegration_ErrorRecovery` - Test error recovery in VM

Validation:
- [ ] Complex script tests pass
- [ ] Error recovery tests pass

---

### T3.4: Code-Mode Integration Tests - API Scenarios
**Files:** `internal/execution/codemode_integration_test.go`
**Dependencies:** T3.3
**Parallel with:** None

Steps:
- [ ] 3.4.1 Implement `TestCodeModeIntegration_APICalls` - Test API calls from VM
- [ ] 3.4.2 Implement `TestCodeModeIntegration_SandboxInteraction` - Test VM-sandbox interaction

Validation:
- [ ] API call tests pass
- [ ] Sandbox interaction tests pass

---

### T4.1: Test Infrastructure - Helpers and Fixtures
**Files:** `internal/execution/test_helper.go`, `internal/execution/testdata/`
**Dependencies:** T3.2, T3.4
**Parallel with:** T4.2

Steps:
- [ ] 4.1.1 Create `internal/execution/test_helper.go` with shared test utilities
- [ ] 4.1.2 Add `NewTestSandbox()` helper function
- [ ] 4.1.3 Add `NewTestVM()` helper function
- [ ] 4.1.4 Create comprehensive test fixtures

Validation:
- [ ] Helper functions compile
- [ ] Helpers are used by existing tests
- [ ] Fixtures load correctly

---

### T4.2: Test Infrastructure - Benchmarks
**Files:** `internal/execution/benchmark_test.go`
**Dependencies:** T4.1
**Parallel with:** None

Steps:
- [ ] 4.2.1 Create `internal/execution/benchmark_test.go` file
- [ ] 4.2.2 Implement `BenchmarkSandbox_Execution` - Benchmark sandbox execution
- [ ] 4.2.3 Implement `BenchmarkCodeMode_Execution` - Benchmark code-mode execution

Validation:
- [ ] Benchmarks compile: `go build ./internal/execution/...`
- [ ] Benchmarks run: `go test -bench=. ./internal/execution/`

---

### T4.3: Test Infrastructure - Coverage Analysis
**Files:** `coverage.out`, test reports
**Dependencies:** T4.1
**Parallel with:** T4.2

Steps:
- [ ] 4.3.1 Run coverage analysis: `go test -coverprofile=coverage.out ./internal/execution/...`
- [ ] 4.3.2 Verify coverage >80% for v2 features
- [ ] 4.3.3 Identify and document uncovered code paths

Validation:
- [ ] Coverage report generated
- [ ] Coverage threshold met (>80%)
- [ ] Uncovered paths documented

---

### T4.4: Final Verification and Documentation
**Files:** `internal/execution/README.md` (test documentation)
**Dependencies:** T4.2, T4.3
**Parallel with:** None

Steps:
- [ ] 4.4.1 Run full test suite: `go test ./internal/execution/... -v`
- [ ] 4.4.2 Verify all tests pass
- [ ] 4.4.3 Run race detector: `go test -race ./internal/execution/...`
- [ ] 4.4.4 Document test patterns in README

Validation:
- [ ] All tests pass
- [ ] Race detector clean
- [ ] Documentation updated

## Phase 1 Checkpoint
- [ ] All 28 tasks completed
- [ ] Test coverage >80% for v2 features
- [ ] All tests pass: `go test ./internal/execution/...`
- [ ] Race detector clean
- [ ] No regressions in existing tests

## Completion Summary

**Date:** TBD
**Tasks Completed:** 28/28

### Final Checks
- [ ] `make build` - PASS
- [ ] `make test` - PASS (including new tests)
- [ ] `make lint` - PASS
- [ ] Test coverage >80% - PASS
- [ ] Race detector clean - PASS

### Phase 1 Success Criteria Met
- [ ] Sandbox resource limit tests pass
- [ ] Sandbox error handling tests pass
- [ ] Code-mode VM integration tests pass
- [ ] Safe API surface tests pass
- [ ] Integration tests pass
- [ ] Benchmarks functional

**Next Phase:** Phase 2 (Context Management) can begin after Phase 1 completion
