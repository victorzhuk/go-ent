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

## 3. OpenCode Runner (Subprocess)
- [ ] Create `internal/execution/opencode.go`
- [ ] Implement subprocess execution (spawn opencode CLI)
- [ ] Add stdin/stdout communication
- [ ] Parse CLI output for results
- [ ] Unit tests with mock subprocess

## 4. CLI Runner
- [ ] Create `internal/execution/cli.go`
- [ ] Implement standalone execution
- [ ] Add prompt template rendering
- [ ] Unit tests

## 5. Execution Engine
- [ ] Create `internal/execution/engine.go`
- [ ] Implement `Execute(ctx, req) (*Result, error)`
- [ ] Add runtime selection logic
- [ ] Create `internal/execution/fallback.go`
- [ ] Implement same-family fallback (MCP: claude-code ↔ open-code, CLI: isolated)
- [ ] Integration tests

## 6. Execution Strategies
- [ ] Create `internal/execution/strategy.go`
- [ ] Implement Single strategy
- [ ] Implement Multi strategy (agent handoff)
- [ ] Implement Parallel strategy with dependency graph
- [ ] Unit tests for each strategy

## 7. Budget Tracking
- [ ] Create `internal/execution/budget.go`
- [ ] Implement budget tracker with mode detection (MCP vs CLI)
- [ ] Add cost calculation per model (Opus $15/$75, Sonnet $3/$15, Haiku $0.25/$1.25)
- [ ] MCP mode: auto-proceed with warning log
- [ ] CLI mode: prompt user for approval
- [ ] Unit tests for both modes

## 8. Code-Mode and Tool Composition
- [ ] Create `internal/execution/sandbox.go`
- [ ] Implement resource limits (memory, CPU, timeout)
- [ ] Create `internal/execution/codemode.go`
- [ ] Integrate goja JavaScript VM
- [ ] Implement safe API surface
- [ ] Create `internal/tool/composer.go`
- [ ] Implement composed tool registry
- [ ] Add persistence to `.go-ent/composed-tools/`
- [ ] Unit tests for sandbox and code-mode

## 9. MCP Integration
- [ ] Create `internal/mcp/tools/execution.go`
- [ ] Register `engine_execute` tool
- [ ] Register `engine_status` tool
- [ ] Register `engine_budget` tool
- [ ] Register `engine_interrupt` tool
- [ ] Update `internal/mcp/tools/register.go` to wire engine

## 10. Integration
- [ ] Update `internal/spec/workflow.go` with ExecutionHistory
- [ ] Update `internal/spec/loop.go` to use execution engine
- [ ] Add goja dependency to go.mod
