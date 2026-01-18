# Tasks: Complete Execution Engine (v2 Features)

## 1. Unit Tests for Sandbox and Code-Mode

### 1.1 Sandbox Resource Limits
- [ ] 1.1.1 Test memory limit enforcement
- [ ] 1.1.2 Test CPU limit enforcement
- [ ] 1.1.3 Test timeout enforcement
- [ ] 1.1.4 Test concurrent sandbox isolation

### 1.2 Sandbox Error Handling
- [ ] 1.2.1 Test panic recovery in sandbox
- [ ] 1.2.2 Test resource exhaustion errors
- [ ] 1.2.3 Test timeout errors
- [ ] 1.2.4 Test sandbox cleanup on error

### 1.3 Code-Mode VM Integration
- [ ] 1.3.1 Test JavaScript VM initialization (goja)
- [ ] 1.3.2 Test code execution in VM
- [ ] 1.3.3 Test VM memory limits
- [ ] 1.3.4 Test VM cleanup

### 1.4 Safe API Surface
- [ ] 1.4.1 Test allowed function exposure
- [ ] 1.4.2 Test blocked function access
- [ ] 1.4.3 Test function argument validation
- [ ] 1.4.4 Test return value handling

## 2. Context Summarization

### 2.1 LLM Integration
- [ ] 2.1.1 Add LLM client to execution package
- [ ] 2.1.2 Implement `SummarizeContext()` function
- [ ] 2.1.3 Add summarization prompt templates
- [ ] 2.1.4 Test summarization accuracy

### 2.2 Context Triggers
- [ ] 2.2.1 Detect context window limit approach (80% threshold)
- [ ] 2.2.2 Track token usage during execution
- [ ] 2.2.3 Trigger summarization before limit
- [ ] 2.2.4 Test trigger timing

### 2.3 Context Management
- [ ] 2.3.1 Store summarized context
- [ ] 2.3.2 Preserve critical information in summary
- [ ] 2.3.3 Maintain context chain (original → summary1 → summary2)
- [ ] 2.3.4 Test multi-level summarization

## 3. Context Limit Handling

### 3.1 Limit Detection
- [ ] 3.1.1 Calculate current context token count
- [ ] 3.1.2 Get model context window size
- [ ] 3.1.3 Determine safe threshold (e.g., 80%)
- [ ] 3.1.4 Test limit calculation

### 3.2 Automatic Summarization
- [ ] 3.2.1 Automatically summarize when approaching limit
- [ ] 3.2.2 Replace old context with summary
- [ ] 3.2.3 Log summarization events
- [ ] 3.2.4 Test automatic workflow

### 3.3 User Control
- [ ] 3.3.1 Allow manual summarization trigger
- [ ] 3.3.2 Configure summarization threshold
- [ ] 3.3.3 Configure model for summarization
- [ ] 3.3.4 Test manual triggers

## 4. Full Execution State Persistence

### 4.1 State Model
- [ ] 4.1.1 Define `ExecutionState` struct
- [ ] 4.1.2 Include context, results, metadata
- [ ] 4.1.3 Add timestamps and versioning
- [ ] 4.1.4 Test state serialization

### 4.2 Storage Layer
- [ ] 4.2.1 Create `.go-ent/executions/` directory
- [ ] 4.2.2 Save state to JSON files
- [ ] 4.2.3 Load state from files
- [ ] 4.2.4 Test storage/retrieval

### 4.3 Checkpointing
- [ ] 4.3.1 Auto-save state on task completion
- [ ] 4.3.2 Manual checkpoint option
- [ ] 4.3.3 Clean up old checkpoints
- [ ] 4.3.4 Test checkpoint frequency

### 4.4 State Recovery
- [ ] 4.4.1 Restore execution from saved state
- [ ] 4.4.2 Validate state integrity
- [ ] 4.4.3 Handle corrupted state files
- [ ] 4.4.4 Test recovery scenarios

## 5. Interrupt/Resume Functionality

### 5.1 Interrupt Mechanism
- [ ] 5.1.1 Implement `engine_interrupt` tool fully
- [ ] 5.1.2 Send interrupt signal to execution
- [ ] 5.1.3 Gracefully stop current task
- [ ] 5.1.4 Test interrupt at various stages

### 5.2 Resume Mechanism
- [ ] 5.2.1 Implement `engine_resume` tool
- [ ] 5.2.2 Load saved execution state
- [ ] 5.2.3 Continue from checkpoint
- [ ] 5.2.4 Test resume scenarios

### 5.3 State Validation
- [ ] 5.3.1 Validate state before resume
- [ ] 5.3.2 Check environment compatibility
- [ ] 5.3.3 Handle version mismatches
- [ ] 5.3.4 Test validation logic

### 5.4 Error Handling
- [ ] 5.4.1 Handle interrupt failures
- [ ] 5.4.2 Handle resume failures
- [ ] 5.4.3 Provide clear error messages
- [ ] 5.4.4 Test error paths

## 6. Execution ID Tracking

### 6.1 ID Generation
- [ ] 6.1.1 Generate unique execution IDs (UUID)
- [ ] 6.1.2 Include in execution state
- [ ] 6.1.3 Display in status output
- [ ] 6.1.4 Test uniqueness

### 6.2 ID Storage
- [ ] 6.2.1 Store ID in execution state files
- [ ] 6.2.2 Index executions by ID
- [ ] 6.2.3 Support listing by ID
- [ ] 6.2.4 Test indexing

### 6.3 ID Lookup
- [ ] 6.3.1 Find execution by ID
- [ ] 6.3.2 Query execution history by ID
- [ ] 6.3.3 Handle missing IDs
- [ ] 6.3.4 Test lookup performance

### 6.4 ID Lifecycle
- [ ] 6.4.1 Track execution status (running, interrupted, completed)
- [ ] 6.4.2 Update status on state changes
- [ ] 6.4.3 Clean up old IDs
- [ ] 6.4.4 Test lifecycle management

## 7. Integration Tests

### 7.1 Interrupt/Resume Workflow
- [ ] 7.1.1 Test interrupt long-running execution
- [ ] 7.1.2 Test resume after interrupt
- [ ] 7.1.3 Test multiple interrupt/resume cycles
- [ ] 7.1.4 Test resume after process restart

### 7.2 Context Summarization Workflow
- [ ] 7.2.1 Test long execution with summarization
- [ ] 7.2.2 Verify context size stays within limits
- [ ] 7.2.3 Verify critical info preserved
- [ ] 7.2.4 Test multi-level summarization

### 7.3 End-to-End Scenarios
- [ ] 7.3.1 Test complete workflow with all v2 features
- [ ] 7.3.2 Test error recovery
- [ ] 7.3.3 Test edge cases
- [ ] 7.3.4 Performance benchmarks
