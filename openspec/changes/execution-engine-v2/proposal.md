# Proposal: Complete Execution Engine (v2 Features)

## Overview

Implement v2 features for the execution engine to enable long-running executions, state persistence, and advanced context management. This completes the execution engine with capabilities for interrupting and resuming executions, handling large contexts via LLM summarization, and persistent state tracking.

## Rationale

### Problem
- No ability to interrupt and resume long-running executions
- Context window limits prevent handling long conversations
- Execution state is lost on process termination
- Code-mode and sandbox features lack unit tests
- No tracking mechanism for interrupted executions

### Solution
- **Context summarization**: Use LLM to summarize long execution contexts when approaching limits
- **State persistence**: Persist full execution state to disk for recovery
- **Interrupt/resume**: Track execution IDs and enable graceful interruption with resume capability
- **Unit tests**: Comprehensive test coverage for sandbox and code-mode features

## Key Components

1. `internal/execution/state.go` - Execution state persistence layer
2. `internal/execution/context.go` - Context summarization and limit handling
3. `internal/execution/interrupt.go` - Execution ID tracking and interrupt handling
4. `internal/execution/sandbox_test.go` - Unit tests for sandbox
5. `internal/execution/codemode_test.go` - Unit tests for code-mode

## Dependencies

- Requires: `add-execution-engine` (v1 features)
- Blocks: P7 (long-running-workflows), P8 (advanced-orchestration)

## Success Criteria

- [ ] Context summarization reduces context size while preserving critical information
- [ ] Context limit handling triggers summarization before hitting token limits
- [ ] Full execution state persists to `.go-ent/executions/` directory
- [ ] Execution can be interrupted via `engine_interrupt` tool
- [ ] Interrupted executions can be resumed from saved state
- [ ] Execution IDs uniquely track and reference executions
- [ ] Sandbox unit tests cover resource limits and timeout scenarios
- [ ] Code-mode unit tests cover VM integration and safe API surface
- [ ] Integration tests pass for interrupt/resume workflow

## Implementation Status

**Not Started**

## Impact

**Performance**:
- Context summarization adds LLM call overhead but enables longer workflows
- State persistence adds disk I/O for checkpointing
- Interrupt tracking adds minimal overhead

**Architecture**:
- State persistence enables fault-tolerant long-running executions
- Context summarization enables unlimited workflow length
- Interrupt/resume enables user control over long processes
- Improved test coverage ensures reliability of sandbox and code-mode features
