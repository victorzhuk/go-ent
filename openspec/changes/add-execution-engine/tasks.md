# Tasks: Add Execution Engine

## 1. Runner Interface ✅
- [x] Create `internal/execution/runner.go`
- [x] Define `Runner` interface
- [x] Define `Request`, `Result` types
- [x] Define `TaskContext`, `BudgetLimit` types

## 2. Claude Code Runner ✅
- [x] Create `internal/execution/claude.go`
- [x] Implement MCP protocol integration
- [x] Add prompt building with role + skills

## 3. OpenCode Runner (Subprocess) ✅
- [x] Create `internal/execution/opencode.go`
- [x] Implement subprocess execution (spawn opencode CLI)
- [x] Use `opencode -p "prompt" -f json -q` command format
- [x] Parse JSON output from stdout

## 4. CLI Runner ✅
- [x] Create `internal/execution/cli.go`
- [x] Implement standalone execution
- [x] Add prompt template rendering

## 5. Execution Engine ✅
- [x] Create `internal/execution/engine.go`
- [x] Implement `Execute(ctx, task) (*Result, error)`
- [x] Add runtime selection logic
- [x] Add `Status(ctx) StatusInfo` method
- [x] Create `internal/execution/fallback.go`
- [x] Implement same-family fallback (MCP: claude-code ↔ open-code, CLI: isolated)
- [x] Integration tests

## 6. Execution Strategies ✅
- [x] Create `internal/execution/strategy.go` (interface)
- [x] Create `internal/execution/single.go` (Single strategy)
- [x] Create `internal/execution/multi.go` (Multi strategy with agent handoff)
- [x] Create `internal/execution/parallel.go` (Parallel strategy with dependency graph)
- [x] Implement topological sort for dependency ordering
- [x] Use errgroup for concurrent execution
- [x] Integration tests for each strategy

## 7. Budget Tracking ✅
- [x] Create `internal/execution/budget.go`
- [x] Implement budget tracker with mode detection (MCP vs CLI)
- [x] Add cost calculation per model (Opus $15/$75, Sonnet $3/$15, Haiku $0.25/$1.25)
- [x] MCP mode: auto-proceed with warning log
- [x] CLI mode: return error for user approval
- [x] Unit tests for budget tracking

## 8. MCP Integration ✅
- [x] Create `internal/mcp/tools/execution.go`
- [x] Register `engine_execute` tool
- [x] Register `engine_status` tool
- [x] Register `engine_budget` tool
- [x] Register `engine_interrupt` tool (stub for v2)
- [x] Update `internal/mcp/tools/register.go` to wire engine

## 9. Workflow Integration ✅
- [x] Update `internal/spec/workflow.go` with ExecutionHistory
- [x] Add `ExecutionRecord` type
- [x] Add `RecordExecution()`, `TotalCost()`, `TotalTokens()` methods
- [x] Add golang.org/x/sync dependency to go.mod

---

## Deferred to v2

### Code-Mode and Tool Composition
- [ ] Create `internal/execution/sandbox.go`
- [ ] Implement resource limits (memory, CPU, timeout)
- [ ] Create `internal/execution/codemode.go`
- [ ] Integrate JavaScript VM (goja or v8go)
- [ ] Implement safe API surface
- [ ] Create `internal/tool/composer.go`
- [ ] Implement composed tool registry
- [ ] Add persistence to `.go-ent/composed-tools/`
- [ ] Unit tests for sandbox and code-mode

### Context Management
- [ ] Context summarization for long executions
- [ ] Context limit handling with LLM-based summarization

### State Management
- [ ] Full execution state persistence
- [ ] Interrupt/resume functionality
- [ ] Execution ID tracking for interrupts
