# Tasks: Add Domain Types

## 1. Create Domain Package Structure

### 1.1 Create domain directory
- [x] Create `internal/domain/` directory
- [x] Add package documentation

### 1.2 Create domain type files
- [x] Create `internal/domain/agent.go`
- [x] Create `internal/domain/runtime.go`
- [x] Create `internal/domain/action.go`
- [x] Create `internal/domain/execution.go`
- [x] Create `internal/domain/skill.go`
- [x] Create `internal/domain/errors.go`

## 2. Implement Agent Types

### 2.1 Define AgentRole enum
- [x] Define `type AgentRole string`
- [x] Add constants: Product, Architect, Senior, Developer, Reviewer, Ops
- [x] Add `String() string` method
- [x] Add `Valid() bool` method
- [x] Add doc comments explaining each role

### 2.2 Define AgentConfig struct
- [x] Define struct with Role, Model, Skills, Tools fields
- [x] Add BudgetLimit, Priority fields
- [x] Add validation methods
- [x] Add doc comments

### 2.3 Define AgentCapability
- [x] Define capability flags/enum
- [x] Add methods to check capabilities
- [x] Add doc comments

## 3. Implement Runtime Types

### 3.1 Define Runtime enum
- [x] Define `type Runtime string`
- [x] Add constants: ClaudeCode, OpenCode, CLI
- [x] Add `String() string` method
- [x] Add `Valid() bool` method
- [x] Add doc comments

### 3.2 Define RuntimeCapability
- [x] Define capability struct
- [x] Add methods to query runtime features
- [x] Add doc comments

## 4. Implement Action Types

### 4.1 Define SpecAction enum
- [x] Define `type SpecAction string`
- [x] Add Discovery actions: Research, Analyze, Retrofit
- [x] Add Planning actions: Proposal, Plan, Design, Split
- [x] Add Execution actions: Implement, Execute, Scaffold
- [x] Add Validation actions: Review, Verify, Debug, Lint
- [x] Add Lifecycle actions: Approve, Archive, Status
- [x] Add `String() string` method
- [x] Add `Valid() bool` method

### 4.2 Add action classification
- [x] Add `Phase() ActionPhase` method
- [x] Define ActionPhase enum (Discovery, Planning, Execution, Validation, Lifecycle)
- [x] Add doc comments explaining each action

## 5. Implement Execution Types

### 5.1 Define ExecutionStrategy enum
- [x] Define `type ExecutionStrategy string`
- [x] Add constants: Single, Multi, Parallel
- [x] Add `String() string` method
- [x] Add `Valid() bool` method
- [x] Add doc comments

### 5.2 Define ExecutionContext struct
- [x] Define struct with Runtime, Agent, Strategy fields
- [x] Add ChangeID, TaskID, Budget fields
- [x] Add metadata fields
- [x] Add doc comments

### 5.3 Define ExecutionResult struct
- [x] Define struct with Success, Output, Error fields
- [x] Add Tokens, Cost, Duration fields
- [x] Add doc comments

## 6. Implement Skill Types

### 6.1 Define Skill interface
- [x] Define interface with Name(), Description() methods
- [x] Add CanHandle(ctx SkillContext) bool method
- [x] Add Execute(ctx context.Context, req SkillRequest) (SkillResult, error) method
- [x] Add doc comments

### 6.2 Define supporting types
- [x] Define SkillMetadata struct
- [x] Define SkillContext struct
- [x] Define SkillRequest struct
- [x] Define SkillResult struct
- [x] Add doc comments for each

## 7. Implement Domain Errors

### 7.1 Define error types
- [x] Define `ErrAgentNotFound`
- [x] Define `ErrRuntimeUnavailable`
- [x] Define `ErrInvalidAction`
- [x] Define `ErrInvalidStrategy`
- [x] Define `ErrSkillNotFound`
- [x] Add error wrapping helpers

### 7.2 Add error helpers
- [x] Add `IsAgentError(err error) bool` type check
- [x] Add `IsRuntimeError(err error) bool` type check
- [x] Add doc comments

## 8. Integration with Existing Code

### 8.1 Update internal/spec/domain.go
- [x] Import `internal/domain` package
- [x] Document relationship between spec and agent domains
- [x] Verify no circular dependencies

### 8.2 Update internal/spec/workflow.go
- [x] Add `AgentRole` field to WorkflowState
- [x] Add import for `internal/domain`
- [x] Update doc comments

## 9. Testing

### 9.1 Unit tests for agent types
- [x] Test AgentRole validation
- [x] Test AgentConfig validation
- [x] Test AgentCapability methods

### 9.2 Unit tests for runtime types
- [x] Test Runtime validation
- [x] Test RuntimeCapability methods

### 9.3 Unit tests for action types
- [x] Test SpecAction validation
- [x] Test Phase() classification
- [x] Test all action constants

### 9.4 Unit tests for execution types
- [x] Test ExecutionStrategy validation
- [x] Test ExecutionContext creation
- [x] Test ExecutionResult marshaling

### 9.5 Unit tests for skill types
- [x] Test Skill interface compliance (mock implementation)
- [x] Test SkillContext usage
- [x] Test SkillResult handling

### 9.6 Unit tests for errors
- [x] Test error creation
- [x] Test error wrapping
- [x] Test error type checks

## 10. Documentation

### 10.1 Add package documentation
- [x] Write comprehensive package doc comment
- [x] Add examples for common usage
- [x] Document design decisions

### 10.2 Add godoc examples
- [x] Add example for AgentRole usage
- [x] Add example for SpecAction classification
- [x] Add example for Skill implementation

## 11. Verification

### 11.1 Build verification
- [x] Run `go build ./internal/domain`
- [x] Verify no compilation errors
- [x] Check for unused imports

### 11.2 Test verification
- [x] Run `go test ./internal/domain`
- [x] Verify >80% coverage
- [x] Check for race conditions with `go test -race`

### 11.3 Dependency verification
- [x] Run `go mod graph | grep internal/domain`
- [x] Verify zero external dependencies
- [x] Verify no circular imports
