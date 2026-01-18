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

- Requires: `add-execution-engine` (v1 features) - ✅ COMPLETED
- Requires: `add-boltdb-state-system` (state persistence) - ⚠️ PARTIAL (12 blocked tasks)
- Blocks: P7 (long-running-workflows), P8 (advanced-orchestration), `add-acp-agent-mode` (worker orchestration)

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

**📋 Phased Implementation Ready** - Broken down into 4 manageable phases:

### Phase 1: Foundation (28 tasks) - ⭐ READY TO START
**Unit tests for existing v2 features** - No dependencies, can start immediately
- Sandbox resource limits and error handling
- Code-mode VM integration and safe API surface

### Phase 2: Context Management (28 tasks) - 🚀 CAN START AFTER PHASE 1  
**Context summarization and limit handling** - Core v2 functionality
- LLM-based context summarization
- Automatic context limit detection and handling

### Phase 3: State Persistence (36 tasks) - ⚠️ DEPENDENT
**Execution state persistence and interrupt/resume** - May use file fallback if add-boltdb-state-system blocked
- Execution state serialization and storage
- Interrupt/resume functionality with execution ID tracking

### Phase 4: Integration (16 tasks) - 🎯 FINAL PHASE
**End-to-end testing and validation** - Requires all previous phases
- Full integration test suite
- Performance benchmarks and documentation

**Estimated Timeline**: 4-6 weeks | **Critical Path**: Phases 1→2→4 (Phase 3 optional with file fallback)

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
