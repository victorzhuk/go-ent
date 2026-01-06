// Package domain defines core domain types for the multi-agent orchestration system.
//
// # Overview
//
// This package establishes the vocabulary and type system for agent-based workflows,
// providing the foundational types that enable multi-agent collaboration, task execution,
// and skill-based capabilities.
//
// # Core Types
//
// The package is organized around five primary concepts:
//
//   - AgentRole: Defines agent specializations and their responsibilities
//   - Runtime: Identifies execution environments where agents operate
//   - SpecAction: Taxonomy of actions agents can perform across workflow phases
//   - ExecutionStrategy: Models execution patterns (single agent, multi-agent, parallel)
//   - Skill: Interface for reusable agent capabilities
//
// # AgentRole Hierarchy
//
// Each role represents a specialized agent with specific capabilities:
//
//	Product      - User needs, requirements, product decisions
//	Architect    - System design, architecture, technical decisions
//	Senior       - Complex implementation, debugging, code review
//	Developer    - Standard implementation, testing
//	Reviewer     - Code quality, standards enforcement
//	Ops          - Deployment, monitoring, production issues
//
// Example usage:
//
//	role := AgentRoleDeveloper
//	if !role.Valid() {
//	    return fmt.Errorf("invalid role: %s", role)
//	}
//
//	// Check capabilities
//	if role.CanImplement() {
//	    // Proceed with implementation task
//	}
//
// # SpecAction Taxonomy
//
// Actions are organized by workflow phase:
//
//	Discovery:   research, analyze, retrofit
//	Planning:    proposal, plan, design, split
//	Execution:   implement, execute, scaffold
//	Validation:  review, verify, debug, lint
//	Lifecycle:   approve, archive, status
//
// Example usage:
//
//	action := SpecActionImplement
//	phase := action.Phase()  // Returns ActionPhaseExecution
//
//	switch phase {
//	case ActionPhaseDiscovery:
//	    // Route to research agent
//	case ActionPhaseExecution:
//	    // Route to developer agent
//	}
//
// # Execution Strategies
//
// Three execution patterns are supported:
//
//   - Single: One agent, sequential execution (simple tasks)
//   - Multi: Multiple agents in conversation/handoff (complex tasks)
//   - Parallel: Independent agents working simultaneously (batch operations)
//
// Example usage:
//
//	ctx := &ExecutionContext{
//	    Runtime:  RuntimeClaudeCode,
//	    Agent:    AgentRoleDeveloper,
//	    Strategy: ExecutionStrategySingle,
//	    ChangeID: "add-feature",
//	}
//
// # Skill Interface
//
// Skills provide reusable capabilities that agents can leverage:
//
//	type MySkill struct {
//	    metadata SkillMetadata
//	}
//
//	func (s *MySkill) Name() string { return "go-code" }
//	func (s *MySkill) Description() string { return "Go code implementation" }
//
//	func (s *MySkill) CanHandle(ctx SkillContext) bool {
//	    return ctx.Language == "go"
//	}
//
//	func (s *MySkill) Execute(ctx context.Context, req SkillRequest) (SkillResult, error) {
//	    // Implementation...
//	}
//
// # Error Handling
//
// The package provides domain-specific error types with context:
//
//	// Wrap with agent context
//	err := &AgentError{
//	    Role: AgentRoleDeveloper,
//	    Err:  ErrInvalidAgentConfig,
//	}
//
//	// Check error types
//	if IsAgentError(err) {
//	    var ae *AgentError
//	    if errors.As(err, &ae) {
//	        log.Printf("agent error for role: %s", ae.Role)
//	    }
//	}
//
//	// Sentinel errors for common conditions
//	if errors.Is(err, ErrAgentNotFound) {
//	    // Handle missing agent
//	}
//
// # Design Decisions
//
// ## Zero External Dependencies
//
// This package has no external dependencies, ensuring it remains a pure domain layer.
// All types use standard library types only.
//
// ## Type-Safe Enums
//
// Custom string types with constants provide type safety without reflection:
//
//	type AgentRole string
//	const AgentRoleDeveloper AgentRole = "developer"
//
// This approach offers:
//   - Type safety at compile time
//   - Automatic JSON marshaling
//   - Easy string conversion
//   - Simple extension without breaking changes
//
// ## Separation from Spec Domain
//
// This package is distinct from internal/spec to avoid circular dependencies
// and maintain clear bounded contexts:
//
//   - internal/spec: OpenSpec change proposals, tasks, validation
//   - internal/domain: Agent roles, execution semantics, skills
//
// ## Skill as Interface
//
// The Skill type is defined as an interface rather than a struct to enable:
//   - Multiple implementations (built-in, custom, plugin-provided)
//   - Easy testing via mocks
//   - Plugin system integration
//   - Dependency inversion principle
//
// # Integration Points
//
// ## With Existing Code
//
//   - internal/spec/workflow.go: Uses AgentRole to track current agent
//   - internal/spec/domain.go: Imports domain types for execution context
//
// ## With Future Systems
//
//   - Config System: Uses AgentRole, Runtime in agent configuration
//   - Agent System: Uses AgentRole, Skill for agent selection and dispatch
//   - Execution Engine: Uses ExecutionStrategy, Runtime for task execution
//
// # Thread Safety
//
// All types in this package are immutable value types (enums, structs with no mutators)
// and are safe for concurrent use. The Skill interface requires implementations
// to be thread-safe if shared across goroutines.
package domain
