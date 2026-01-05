# Tasks: Add Domain Types

## 1. Create Domain Package Structure

### 1.1 Create domain directory
- [ ] Create `internal/domain/` directory
- [ ] Add package documentation

### 1.2 Create domain type files
- [ ] Create `internal/domain/agent.go`
- [ ] Create `internal/domain/runtime.go`
- [ ] Create `internal/domain/action.go`
- [ ] Create `internal/domain/execution.go`
- [ ] Create `internal/domain/skill.go`
- [ ] Create `internal/domain/errors.go`

## 2. Implement Agent Types

### 2.1 Define AgentRole enum
- [ ] Define `type AgentRole string`
- [ ] Add constants: Product, Architect, Senior, Developer, Reviewer, Ops
- [ ] Add `String() string` method
- [ ] Add `Valid() bool` method
- [ ] Add doc comments explaining each role

### 2.2 Define AgentConfig struct
- [ ] Define struct with Role, Model, Skills, Tools fields
- [ ] Add BudgetLimit, Priority fields
- [ ] Add validation methods
- [ ] Add doc comments

### 2.3 Define AgentCapability
- [ ] Define capability flags/enum
- [ ] Add methods to check capabilities
- [ ] Add doc comments

## 3. Implement Runtime Types

### 3.1 Define Runtime enum
- [ ] Define `type Runtime string`
- [ ] Add constants: ClaudeCode, OpenCode, CLI
- [ ] Add `String() string` method
- [ ] Add `Valid() bool` method
- [ ] Add doc comments

### 3.2 Define RuntimeCapability
- [ ] Define capability struct
- [ ] Add methods to query runtime features
- [ ] Add doc comments

## 4. Implement Action Types

### 4.1 Define SpecAction enum
- [ ] Define `type SpecAction string`
- [ ] Add Discovery actions: Research, Analyze, Retrofit
- [ ] Add Planning actions: Proposal, Plan, Design, Split
- [ ] Add Execution actions: Implement, Execute, Scaffold
- [ ] Add Validation actions: Review, Verify, Debug, Lint
- [ ] Add Lifecycle actions: Approve, Archive, Status
- [ ] Add `String() string` method
- [ ] Add `Valid() bool` method

### 4.2 Add action classification
- [ ] Add `Phase() ActionPhase` method
- [ ] Define ActionPhase enum (Discovery, Planning, Execution, Validation, Lifecycle)
- [ ] Add doc comments explaining each action

## 5. Implement Execution Types

### 5.1 Define ExecutionStrategy enum
- [ ] Define `type ExecutionStrategy string`
- [ ] Add constants: Single, Multi, Parallel
- [ ] Add `String() string` method
- [ ] Add `Valid() bool` method
- [ ] Add doc comments

### 5.2 Define ExecutionContext struct
- [ ] Define struct with Runtime, Agent, Strategy fields
- [ ] Add ChangeID, TaskID, Budget fields
- [ ] Add metadata fields
- [ ] Add doc comments

### 5.3 Define ExecutionResult struct
- [ ] Define struct with Success, Output, Error fields
- [ ] Add Tokens, Cost, Duration fields
- [ ] Add doc comments

## 6. Implement Skill Types

### 6.1 Define Skill interface
- [ ] Define interface with Name(), Description() methods
- [ ] Add CanHandle(ctx SkillContext) bool method
- [ ] Add Execute(ctx context.Context, req SkillRequest) (SkillResult, error) method
- [ ] Add doc comments

### 6.2 Define supporting types
- [ ] Define SkillMetadata struct
- [ ] Define SkillContext struct
- [ ] Define SkillRequest struct
- [ ] Define SkillResult struct
- [ ] Add doc comments for each

## 7. Implement Domain Errors

### 7.1 Define error types
- [ ] Define `ErrAgentNotFound`
- [ ] Define `ErrRuntimeUnavailable`
- [ ] Define `ErrInvalidAction`
- [ ] Define `ErrInvalidStrategy`
- [ ] Define `ErrSkillNotFound`
- [ ] Add error wrapping helpers

### 7.2 Add error helpers
- [ ] Add `IsAgentError(err error) bool` type check
- [ ] Add `IsRuntimeError(err error) bool` type check
- [ ] Add doc comments

## 8. Integration with Existing Code

### 8.1 Update internal/spec/domain.go
- [ ] Import `internal/domain` package
- [ ] Document relationship between spec and agent domains
- [ ] Verify no circular dependencies

### 8.2 Update internal/spec/workflow.go
- [ ] Add `AgentRole` field to WorkflowState
- [ ] Add import for `internal/domain`
- [ ] Update doc comments

## 9. Testing

### 9.1 Unit tests for agent types
- [ ] Test AgentRole validation
- [ ] Test AgentConfig validation
- [ ] Test AgentCapability methods

### 9.2 Unit tests for runtime types
- [ ] Test Runtime validation
- [ ] Test RuntimeCapability methods

### 9.3 Unit tests for action types
- [ ] Test SpecAction validation
- [ ] Test Phase() classification
- [ ] Test all action constants

### 9.4 Unit tests for execution types
- [ ] Test ExecutionStrategy validation
- [ ] Test ExecutionContext creation
- [ ] Test ExecutionResult marshaling

### 9.5 Unit tests for skill types
- [ ] Test Skill interface compliance (mock implementation)
- [ ] Test SkillContext usage
- [ ] Test SkillResult handling

### 9.6 Unit tests for errors
- [ ] Test error creation
- [ ] Test error wrapping
- [ ] Test error type checks

## 10. Documentation

### 10.1 Add package documentation
- [ ] Write comprehensive package doc comment
- [ ] Add examples for common usage
- [ ] Document design decisions

### 10.2 Add godoc examples
- [ ] Add example for AgentRole usage
- [ ] Add example for SpecAction classification
- [ ] Add example for Skill implementation

## 11. Verification

### 11.1 Build verification
- [ ] Run `go build ./internal/domain`
- [ ] Verify no compilation errors
- [ ] Check for unused imports

### 11.2 Test verification
- [ ] Run `go test ./internal/domain`
- [ ] Verify >80% coverage
- [ ] Check for race conditions with `go test -race`

### 11.3 Dependency verification
- [ ] Run `go mod graph | grep internal/domain`
- [ ] Verify zero external dependencies
- [ ] Verify no circular imports
