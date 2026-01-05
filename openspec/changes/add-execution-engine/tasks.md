# Tasks: Add Execution Engine

## 1. Runner Interface
- [ ] Create `internal/execution/runner.go`
- [ ] Define `Runner` interface
- [ ] Define `RunRequest`, `RunResult` types
- [ ] Add runner factory

## 2. Claude Code Runner
- [ ] Create `internal/execution/claude.go`
- [ ] Implement MCP protocol integration
- [ ] Add prompt building with role + skills
- [ ] Unit tests with mock MCP client

## 3. OpenCode Runner (Required)
- [ ] Create `internal/execution/opencode.go`
- [ ] Implement native OpenCode API integration
- [ ] Add authentication handling
- [ ] Unit tests with mock API client

## 4. CLI Runner
- [ ] Create `internal/execution/cli.go`
- [ ] Implement standalone execution
- [ ] Add prompt template rendering
- [ ] Unit tests

## 5. Execution Engine
- [ ] Create `internal/execution/engine.go`
- [ ] Implement `Execute(ctx, req) (*Result, error)`
- [ ] Add runtime selection logic
- [ ] Integration tests

## 6. Execution Strategies
- [ ] Create `internal/execution/strategy.go`
- [ ] Implement Single strategy
- [ ] Implement Multi strategy (agent handoff)
- [ ] Implement Parallel strategy with dependency graph
- [ ] Unit tests for each strategy

## 7. Budget Tracking
- [ ] Create `internal/execution/budget.go`
- [ ] Implement budget tracker
- [ ] Add cost calculation per model
- [ ] Add spending limits enforcement

## 8. Integration
- [ ] Update `internal/spec/workflow.go` with ExecutionHistory
- [ ] Update `internal/spec/loop.go` to use execution engine
