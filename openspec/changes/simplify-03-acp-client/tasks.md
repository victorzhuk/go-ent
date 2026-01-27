# Tasks: Create ACP Client for OpenCode (Phase 3)

## Task Breakdown

### 1. Create internal/acp/client.go - HTTP client basics
**Priority:** High
**Estimated Complexity:** Medium
**Dependencies:** None

**Steps:**
1. Create `internal/acp/client.go`
2. Define `Client` struct with HTTP client
3. Implement `NewClient(baseURL string) *Client`
4. Implement `Connect(ctx context.Context) error` - verify ACP server is running
5. Add basic HTTP error handling
6. Write unit tests for client creation

**Validation:**
- [ ] Client struct defined
- [ ] Can create new client
- [ ] Connection verification works
- [ ] Unit tests pass

**Files Modified:**
- `internal/acp/client.go` (new)
- `internal/acp/client_test.go` (new)

---

### 2. Create internal/acp/protocol.go - Message types
**Priority:** High
**Estimated Complexity:** Low
**Dependencies:** None

**Steps:**
1. Create `internal/acp/protocol.go`
2. Define `Task` struct with ID, Description, Skills, Context
3. Define `Result` struct with TaskID, Status, Output, Error, Duration
4. Define `Status` struct with TaskID, State, Progress, Output, timestamps
5. Define JSON-RPC request/response wrappers
6. Add JSON tags for marshaling
7. Write unit tests for marshaling/unmarshaling

**Validation:**
- [ ] All message types defined
- [ ] JSON marshaling works correctly
- [ ] Unit tests pass

**Files Modified:**
- `internal/acp/protocol.go` (new)
- `internal/acp/protocol_test.go` (new)

---

### 3. Implement Execute() - Single task execution
**Priority:** High
**Estimated Complexity:** Medium
**Dependencies:** Tasks 1, 2

**Steps:**
1. Add `Execute(ctx context.Context, task Task) (*Result, error)` to client.go
2. Marshal task to JSON-RPC request
3. POST to `/execute` endpoint
4. Parse JSON-RPC response
5. Return Result or error
6. Add timeout handling (use context)
7. Write integration test with mock server

**Validation:**
- [ ] Can execute single task
- [ ] Timeout handling works
- [ ] Error handling works
- [ ] Integration test passes

**Files Modified:**
- `internal/acp/client.go`
- `internal/acp/client_test.go`

---

### 4. Implement Status() - Task status checking
**Priority:** High
**Estimated Complexity:** Low
**Dependencies:** Tasks 1, 2

**Steps:**
1. Add `Status(ctx context.Context, taskID string) (*Status, error)` to client.go
2. GET from `/status/:id` endpoint
3. Parse JSON-RPC response
4. Return Status or error
5. Write unit test

**Validation:**
- [ ] Can get task status
- [ ] Status parsing works
- [ ] Unit test passes

**Files Modified:**
- `internal/acp/client.go`
- `internal/acp/client_test.go`

---

### 5. Implement Cancel() - Task cancellation
**Priority:** High
**Estimated Complexity:** Low
**Dependencies:** Tasks 1, 2

**Steps:**
1. Add `Cancel(ctx context.Context, taskID string) error` to client.go
2. POST to `/cancel/:id` endpoint
3. Handle response
4. Write unit test

**Validation:**
- [ ] Can cancel task
- [ ] Error handling works
- [ ] Unit test passes

**Files Modified:**
- `internal/acp/client.go`
- `internal/acp/client_test.go`

---

### 6. Implement StreamOutput() - Output streaming
**Priority:** Medium
**Estimated Complexity:** Medium
**Dependencies:** Tasks 1, 2

**Steps:**
1. Add `StreamOutput(ctx context.Context, taskID string) (<-chan string, error)` to client.go
2. Open WebSocket connection to `/stream/:id`
3. Read messages from WebSocket
4. Send to output channel
5. Handle WebSocket errors and closure
6. Write integration test

**Validation:**
- [ ] Can stream task output
- [ ] Channel receives messages
- [ ] Cleanup on context cancellation
- [ ] Integration test passes

**Files Modified:**
- `internal/acp/client.go`
- `internal/acp/client_test.go`

---

### 7. Implement ExecuteParallel() - Parallel execution
**Priority:** High
**Estimated Complexity:** Medium
**Dependencies:** Tasks 3

**Steps:**
1. Add `ExecuteParallel(ctx context.Context, tasks []Task) ([]Result, error)` to client.go
2. Use goroutines to execute tasks concurrently
3. Collect results in order
4. Use errgroup for error handling
5. Handle partial failures
6. Write integration test with multiple tasks

**Validation:**
- [ ] Can execute multiple tasks in parallel
- [ ] Results returned in correct order
- [ ] Error handling works
- [ ] Integration test passes

**Files Modified:**
- `internal/acp/client.go`
- `internal/acp/client_test.go`

---

### 8. Create internal/acp/executor.go - Domain integration
**Priority:** High
**Estimated Complexity:** Medium
**Dependencies:** Tasks 1-7

**Steps:**
1. Create `internal/acp/executor.go`
2. Define `Executor` struct wrapping Client
3. Implement `NewExecutor(baseURL string) (*Executor, error)`
4. Implement `ExecuteTask(ctx, *domain.Task) error` - bridge to domain types
5. Implement `ExecuteTasks(ctx, []*domain.Task) error` - parallel execution
6. Implement `GetTaskStatus(ctx, taskID string) (string, error)`
7. Implement `CancelTask(ctx, taskID string) error`
8. Add conversion between domain.Task and acp.Task

**Validation:**
- [ ] Executor created successfully
- [ ] Domain task conversion works
- [ ] Integration with domain types works
- [ ] Tests pass

**Files Modified:**
- `internal/acp/executor.go` (new)
- `internal/acp/executor_test.go` (new)

---

### 9. Add progress/output streaming to Executor
**Priority:** Medium
**Estimated Complexity:** Medium
**Dependencies:** Task 8

**Steps:**
1. Add callback function parameter to ExecuteTask for progress updates
2. Use StreamOutput internally to get task output
3. Parse output for progress indicators
4. Call callback with progress updates
5. Write integration test

**Validation:**
- [ ] Progress callbacks work
- [ ] Output streaming integrated
- [ ] Integration test passes

**Files Modified:**
- `internal/acp/executor.go`
- `internal/acp/executor_test.go`

---

### 10. Integration tests with OpenCode
**Priority:** High
**Estimated Complexity:** High
**Dependencies:** All previous tasks

**Steps:**
1. Create `internal/acp/integration_test.go`
2. Set up test that starts OpenCode ACP server (or use testcontainers)
3. Test full workflow:
   - Connect to server
   - Execute task
   - Monitor status
   - Verify completion
4. Test parallel execution with 3+ tasks
5. Test cancellation mid-execution
6. Add skip flag if OpenCode not available

**Validation:**
- [ ] Integration test passes with real OpenCode
- [ ] Parallel execution works
- [ ] Cancellation works
- [ ] Can skip if OpenCode unavailable

**Files Modified:**
- `internal/acp/integration_test.go` (new)

---

### 11. Add documentation and examples
**Priority:** Medium
**Estimated Complexity:** Low
**Dependencies:** All previous tasks

**Steps:**
1. Add package documentation to client.go
2. Add examples to GoDoc comments
3. Create `internal/acp/README.md` with usage guide
4. Document ACP protocol details
5. Add troubleshooting section

**Validation:**
- [ ] Package documentation complete
- [ ] Examples clear
- [ ] README created

**Files Modified:**
- `internal/acp/client.go`
- `internal/acp/README.md` (new)

---

### 12. Final verification
**Priority:** High
**Estimated Complexity:** Low
**Dependencies:** All previous tasks

**Steps:**
1. Run `make build` - must succeed
2. Run `make test` - all tests must pass
3. Run `make lint` - no lint errors
4. Manual test: start OpenCode, execute task via ACP
5. Verify parallel execution with 3 tasks
6. Verify cancellation works

**Validation:**
- [ ] Build succeeds
- [ ] All tests pass
- [ ] Lint clean
- [ ] Manual tests work
- [ ] Documentation complete

**Files Modified:**
- None (verification only)

---

## Task Order

**Sequential:**
- Tasks 1-2 can be done in parallel (client and protocol)
- Tasks 3-7 must be done after 1-2
- Task 8 must be done after 3-7
- Task 9 must be done after 8
- Task 10 must be done after 9
- Tasks 11-12 can be done after 10

## Estimated Total Time

- Tasks 1-2: 2 hours (client and protocol basics)
- Tasks 3-7: 4 hours (core functionality)
- Task 8: 2 hours (domain integration)
- Task 9: 1.5 hours (progress streaming)
- Task 10: 2 hours (integration tests)
- Task 11: 1 hour (documentation)
- Task 12: 30 minutes (verification)
- **Total:** ~13 hours

## Testing Strategy

1. **Unit Tests:** For protocol marshaling, client methods
2. **Integration Tests:** With mock HTTP server
3. **End-to-End Tests:** With real OpenCode (optional/skip if unavailable)
4. **Parallel Execution:** Test with 3+ concurrent tasks
5. **Cancellation:** Test mid-execution cancellation
6. **Error Handling:** Test connection failures, timeouts, invalid responses
