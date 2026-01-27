# Tasks: Complete Execution Engine (v2 Features)

## 🎯 Implementation Strategy - Phased Approach

**Total Tasks**: 100 | **Phases**: 4 | **Estimated Effort**: 40-50 hours

### Phase Dependencies
- **Phase 1** (Foundation): No dependencies - can start immediately
- **Phase 2** (Context Management): Requires Phase 1 completion
- **Phase 3** (State Persistence): Requires Phase 2 completion + add-boltdb-state-system (partial)
- **Phase 4** (Integration): Requires all previous phases

### 🚀 Quick Start
- **Start with Phase 1** - Unit tests for existing features (28 tasks)
- **Phase 2** can begin immediately after Phase 1
- **Phase 3** may be delayed if add-boltdb-state-system blockers persist
- **Phase 4** brings everything together with integration tests

---

## Phase 1: Foundation - Unit Tests (28 tasks) ⭐ START HERE

**Objective**: Test existing v2 features (sandbox, code-mode) that are already implemented
**Effort**: 10-12 hours | **Dependencies**: None | **Status**: Ready to start

### 1.1 Sandbox Resource Limits (4 tasks)

## 1. Unit Tests for Sandbox and Code-Mode

### 1.1 Sandbox Resource Limits
- [x] 1.1.1 Test memory limit enforcement
- [x] 1.1.2 Test CPU limit enforcement
- [x] 1.1.3 Test timeout enforcement
- [x] 1.1.4 Test concurrent sandbox isolation

### 1.2 Sandbox Error Handling
- [x] 1.2.1 Test panic recovery in sandbox
- [x] 1.2.2 Test resource exhaustion errors
- [x] 1.2.3 Test timeout errors
- [x] 1.2.4 Test sandbox cleanup on error

### 1.3 Code-Mode VM Integration
- [x] 1.3.1 Test JavaScript VM initialization (goja)
- [x] 1.3.2 Test code execution in VM
- [x] 1.3.3 Test VM memory limits
- [x] 1.3.4 Test VM cleanup

### 1.4 Safe API Surface (4 tasks)
- [x] 1.4.1 Test allowed function exposure
- [x] 1.4.2 Test blocked function access
- [x] 1.4.3 Test function argument validation
- [x] 1.4.4 Test return value handling

---

## Phase 2: Context Management (28 tasks)

**Objective**: Implement context summarization and limit handling
**Effort**: 12-15 hours | **Dependencies**: Phase 1 | **Status**: Can start after Phase 1

## 2. Context Summarization

### 2.1 LLM Integration
- [x] 2.1.1 Add LLM client to execution package
- [x] 2.1.2 Implement `SummarizeContext()` function
- [x] 2.1.3 Add summarization prompt templates
- [x] 2.1.4 Test summarization accuracy

### 2.2 Context Triggers
- [x] 2.2.1 Detect context window limit approach (80% threshold)
- [x] 2.2.2 Track token usage during execution
- [x] 2.2.3 Trigger summarization before limit
- [x] 2.2.4 Test trigger timing

### 2.3 Context Management (4 tasks)
- [x] 2.3.1 Store summarized context
- [x] 2.3.2 Preserve critical information in summary
- [x] 2.3.3 Maintain context chain (original → summary1 → summary2)
- [x] 2.3.4 Test multi-level summarization

## 3. Context Limit Handling (12 tasks)

### 3.1 Limit Detection
- [x] 3.1.1 Calculate current context token count
- [x] 3.1.2 Get model context window size
- [x] 3.1.3 Determine safe threshold (e.g., 80%)
- [x] 3.1.4 Test limit calculation

### 3.2 Automatic Summarization
- [x] 3.2.1 Automatically summarize when approaching limit
- [x] 3.2.2 Replace old context with summary
- [x] 3.2.3 Log summarization events
- [x] 3.2.4 Test automatic workflow

### 3.3 User Control (4 tasks)
- [x] 3.3.1 Allow manual summarization trigger
- [x] 3.3.2 Configure summarization threshold
- [x] 3.3.3 Configure model for summarization
- [x] 3.3.4 Test manual triggers

---

## Phase 3: State Persistence (36 tasks)

**Objective**: Implement execution state persistence and interrupt/resume
**Effort**: 12-15 hours | **Dependencies**: Phase 2 + add-boltdb-state-system* | **Status**: *May use file fallback

*Note: If add-boltdb-state-system remains blocked, implement file-based persistence as fallback*

## 4. Full Execution State Persistence (16 tasks)

### 4.1 State Model
- [x] 4.1.1 Define `ExecutionState` struct
- [x] 4.1.2 Include context, results, metadata
- [x] 4.1.3 Add timestamps and versioning
- [x] 4.1.4 Test state serialization

### 4.2 Storage Layer
- [x] 4.2.1 Create `.go-ent/executions/` directory
- [x] 4.2.2 Save state to JSON files
- [x] 4.2.3 Load state from files
- [x] 4.2.4 Test storage/retrieval

### 4.3 Checkpointing
- [x] 4.3.1 Auto-save state on task completion
- [x] 4.3.2 Manual checkpoint option
- [x] 4.3.3 Clean up old checkpoints
- [x] 4.3.4 Test checkpoint frequency

### 4.4 State Recovery (4 tasks)
- [x] 4.4.1 Restore execution from saved state
- [x] 4.4.2 Validate state integrity
- [x] 4.4.3 Handle corrupted state files
- [x] 4.4.4 Test recovery scenarios

## 5. Interrupt/Resume Functionality (20 tasks)

### 5.1 Interrupt Mechanism
- [x] 5.1.1 Implement `engine_interrupt` tool fully
- [x] 5.1.2 Send interrupt signal to execution
- [x] 5.1.3 Gracefully stop current task
- [x] 5.1.4 Test interrupt at various stages

### 5.2 Resume Mechanism
- [x] 5.2.1 Implement `engine_resume` tool
- [x] 5.2.2 Load saved execution state
- [x] 5.2.3 Continue from checkpoint
- [x] 5.2.4 Test resume scenarios

### 5.3 State Validation
- [x] 5.3.1 Validate state before resume
- [x] 5.3.2 Check environment compatibility
- [x] 5.3.3 Handle version mismatches
- [x] 5.3.4 Test validation logic

### 5.4 Error Handling (4 tasks)
- [x] 5.4.1 Handle interrupt failures
- [x] 5.4.2 Handle resume failures
- [x] 5.4.3 Provide clear error messages
- [x] 5.4.4 Test error paths

## 6. Execution ID Tracking (12 tasks)

### 6.1 ID Generation
- [x] 6.1.1 Generate unique execution IDs (UUID)
- [x] 6.1.2 Include in execution state
- [x] 6.1.3 Display in status output
- [x] 6.1.4 Test uniqueness

### 6.2 ID Storage
- [x] 6.2.1 Store ID in execution state files
- [x] 6.2.2 Index executions by ID
- [x] 6.2.3 Support listing by ID
- [x] 6.2.4 Test indexing

### 6.3 ID Lookup
- [x] 6.3.1 Find execution by ID
- [x] 6.3.2 Query execution history by ID
- [x] 6.3.3 Handle missing IDs
- [x] 6.3.4 Test lookup performance

### 6.4 ID Lifecycle
- [x] 6.4.1 Track execution status (running, interrupted, completed, failed)
- [x] 6.4.2 Update status on state changes
- [x] 6.4.3 Clean up old IDs
- [x] 6.4.4 Test lifecycle management

## Phase 4: Integration & Testing (16 tasks)

**Objective**: End-to-end testing and performance validation
**Effort**: 6-8 hours | **Dependencies**: Phases 1-3 | **Status**: Final phase

### 7.1 Interrupt/Resume Workflow (4 tasks)
- [x] 7.1.1 Test interrupt long-running execution
- [x] 7.1.2 Test resume after interrupt
- [x] 7.1.3 Test multiple interrupt/resume cycles
- [x] 7.1.4 Test resume after process restart

### 7.2 Context Summarization Workflow (4 tasks)
- [x] 7.2.1 Test long execution with summarization
- [x] 7.2.2 Verify context size stays within limits
- [x] 7.2.3 Verify critical info preserved
- [x] 7.2.4 Test multi-level summarization

### 7.3 End-to-End Scenarios (4 tasks)
- [x] 7.3.1 Test complete workflow with all v2 features
- [x] 7.3.2 Test error recovery
- [x] 7.3.3 Test edge cases
- [x] 7.3.4 Performance benchmarks

### 7.4 Documentation & Examples (4 tasks)
- [x] 7.4.1 Document v2 feature usage
- [x] 7.4.2 Create example workflows
- [x] 7.4.3 Update API documentation
- [x] 7.4.4 Add troubleshooting guide

---

## 🎯 Implementation Roadmap

### Week 1: Foundation (Phase 1)
```bash
# Start with unit tests for existing features
go test ./internal/execution/... -v

# Target: All Phase 1 tests passing
```

### Week 2: Context Management (Phase 2)  
```bash
# Implement context summarization
go test ./internal/execution/context_test.go -v

# Target: Context limits and summarization working
```

### Week 3: State Persistence (Phase 3)
```bash
# Note: May be delayed if add-boltdb-state-system blockers persist
# Can implement file-based state as fallback
go test ./internal/execution/state_test.go -v

# Target: Execution state persistence working
```

### Week 4: Integration (Phase 4)
```bash
# Full integration testing
go test ./test/integration/... -v

# Target: All v2 features integrated and tested
```

---

## 🔗 Cross-Phase Dependencies

```
Phase 1 (Foundation) 
    ↓ (depends on)
Phase 2 (Context Management)
    ↓ (depends on) 
Phase 3 (State Persistence) ←── add-boltdb-state-system
    ↓ (depends on)
Phase 4 (Integration & Testing)
```

## ⚠️ Risk Mitigation

### If add-boltdb-state-system remains blocked:
- **Phase 3**: Use file-based state persistence as fallback
- **Impact**: Slightly less robust but fully functional
- **Timeline**: No delay to overall project

### If LLM summarization proves complex:
- **Phase 2**: Start with simple truncation + key points preservation
- **Enhancement**: Add sophisticated summarization in Phase 4
- **Impact**: Minimal - core functionality preserved

---

## ✅ Success Criteria by Phase

### Phase 1 Success
- [x] All sandbox unit tests pass
- [x] All code-mode unit tests pass
- [x] Resource limit enforcement verified
- [x] Error handling tested

### Phase 2 Success
- [ ] Context summarization working
- [ ] Token limit detection accurate
- [ ] Automatic summarization triggers correctly
- [ ] Manual controls functional

### Phase 3 Success
- [x] Execution state persistence working
- [x] Checkpoint/restore functional
- [x] State recovery handles errors
- [x] Execution ID tracking complete

### Phase 4 Success
- [x] All integration tests pass
- [x] End-to-end workflows verified
- [x] Performance benchmarks meet targets
- [x] Documentation complete

**Overall**: ✅ v2 features ready for production use, unblocking add-acp-agent-mode and other dependent changes.
