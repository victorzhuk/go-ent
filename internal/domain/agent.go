package domain

import (
	"fmt"
	"strings"
)

// AgentRole defines the specialization of an agent in the multi-agent system.
type AgentRole string

// Agent role constants define the available specializations in the multi-agent system.
// Each role represents a distinct area of expertise and responsibility.
//
// Plugin Agent Mapping:
//   - ent:coder        → AgentRoleCoder (or AgentRoleDeveloper)
//   - ent:planner      → AgentRolePlanner
//   - ent:planner-fast → AgentRolePlanner
//   - ent:planner-heavy→ AgentRolePlanner
//   - ent:architect    → AgentRoleArchitect
//   - ent:tester       → AgentRoleTester
//   - ent:reviewer     → AgentRoleReviewer
//   - ent:debugger     → AgentRoleDebugger
//   - ent:debugger-fast→ AgentRoleDebugger
//   - ent:debugger-heavy→ AgentRoleDebugger
//   - ent:researcher   → AgentRoleResearcher
//   - ent:acceptor     → AgentRoleAcceptor
//   - ent:reproducer   → AgentRoleReproducer
//   - ent:decomposer   → AgentRoleDecomposer
//   - ent:task-fast    → AgentRoleTask
//   - ent:task-heavy   → AgentRoleTask
const (
	// AgentRoleProduct handles user needs, requirements, and product decisions.
	// Focuses on understanding what needs to be built and why.
	AgentRoleProduct AgentRole = "product"

	// AgentRoleArchitect handles system design, architecture, and technical decisions.
	// Responsible for high-level design, technology choices, and architectural patterns.
	AgentRoleArchitect AgentRole = "architect"

	// AgentRoleSenior handles complex implementation, debugging, and code review.
	// Takes on challenging technical problems requiring deep expertise.
	AgentRoleSenior AgentRole = "senior"

	// AgentRoleDeveloper handles standard implementation and testing.
	// Executes well-defined tasks and writes tests for new functionality.
	AgentRoleDeveloper AgentRole = "developer"

	// AgentRoleReviewer handles code quality and standards enforcement.
	// Reviews code for correctness, style, security, and best practices.
	AgentRoleReviewer AgentRole = "reviewer"

	// AgentRoleOps handles deployment, monitoring, and production issues.
	// Manages infrastructure, observability, and operational concerns.
	AgentRoleOps AgentRole = "ops"

	// AgentRoleCoder is an alias for AgentRoleDeveloper.
	// Used by the plugin system (ent:coder) for code implementation.
	AgentRoleCoder AgentRole = "coder"

	// AgentRolePlanner handles task planning and breakdown.
	// Breaks features into actionable, dependency-aware tasks.
	AgentRolePlanner AgentRole = "planner"

	// AgentRoleTester handles test coverage and TDD cycles.
	// Writes tests, analyzes failures, ensures quality.
	AgentRoleTester AgentRole = "tester"

	// AgentRoleDebugger handles bug investigation and resolution.
	// Systematic debugging across multiple files and components.
	AgentRoleDebugger AgentRole = "debugger"

	// AgentRoleResearcher handles deep code analysis and investigation.
	// Investigates root causes, analyzes patterns, researches solutions.
	AgentRoleResearcher AgentRole = "researcher"

	// AgentRoleAcceptor validates acceptance criteria and spec compliance.
	// Ensures implementations meet requirements before completion.
	AgentRoleAcceptor AgentRole = "acceptor"

	// AgentRoleReproducer creates minimal bug reproductions.
	// Writes failing tests first to establish baseline.
	AgentRoleReproducer AgentRole = "reproducer"

	// AgentRoleDecomposer breaks down complex tasks into implementable graphs.
	// Creates dependency-aware task breakdowns.
	AgentRoleDecomposer AgentRole = "decomposer"

	// AgentRoleTask handles task assessment and routing.
	// Fast or heavy variants for complexity-based routing.
	AgentRoleTask AgentRole = "task"
)

// String returns the string representation of the agent role.
func (r AgentRole) String() string {
	return string(r)
}

// Valid returns true if the agent role is valid.
func (r AgentRole) Valid() bool {
	switch r {
	case AgentRoleProduct, AgentRoleArchitect, AgentRoleSenior,
		AgentRoleDeveloper, AgentRoleReviewer, AgentRoleOps,
		AgentRoleCoder, AgentRolePlanner, AgentRoleTester,
		AgentRoleDebugger, AgentRoleResearcher, AgentRoleAcceptor,
		AgentRoleReproducer, AgentRoleDecomposer, AgentRoleTask:
		return true
	default:
		return false
	}
}

// ParseAgentRole parses a plugin agent name (e.g., "ent:coder") into an AgentRole.
// It handles the "ent:" prefix and maps variant names to their base roles.
func ParseAgentRole(name string) (AgentRole, error) {
	// Strip "ent:" prefix if present
	name = strings.TrimPrefix(name, "ent:")

	switch name {
	case "coder", "dev":
		return AgentRoleCoder, nil
	case "developer":
		return AgentRoleDeveloper, nil
	case "planner", "planner-fast", "planner-heavy":
		return AgentRolePlanner, nil
	case "architect":
		return AgentRoleArchitect, nil
	case "tester":
		return AgentRoleTester, nil
	case "reviewer":
		return AgentRoleReviewer, nil
	case "debugger", "debugger-fast", "debugger-heavy":
		return AgentRoleDebugger, nil
	case "researcher":
		return AgentRoleResearcher, nil
	case "acceptor":
		return AgentRoleAcceptor, nil
	case "reproducer":
		return AgentRoleReproducer, nil
	case "decomposer":
		return AgentRoleDecomposer, nil
	case "task", "task-fast", "task-heavy":
		return AgentRoleTask, nil
	case "product":
		return AgentRoleProduct, nil
	case "senior":
		return AgentRoleSenior, nil
	case "ops":
		return AgentRoleOps, nil
	default:
		return "", fmt.Errorf("unknown agent role: %s", name)
	}
}

// AgentConfig defines the configuration for an agent instance.
// It specifies the agent's role, model, available skills, and tools.
type AgentConfig struct {
	// Role specifies the agent's specialization and responsibility area.
	Role AgentRole

	// Model is the AI model to use for this agent (e.g., "sonnet", "opus", "haiku").
	Model string

	// Skills is a list of skill names available to this agent.
	Skills []string

	// Tools is a list of tool names available to this agent.
	Tools []string

	// BudgetLimit is the maximum token budget for this agent (0 means unlimited).
	BudgetLimit int

	// Priority is the execution priority for this agent (higher = more important).
	Priority int
}

// Valid returns true if the agent configuration is valid.
func (c *AgentConfig) Valid() bool {
	if !c.Role.Valid() {
		return false
	}
	if c.Model == "" {
		return false
	}
	if c.BudgetLimit < 0 {
		return false
	}
	return true
}

// Validate returns an error if the agent configuration is invalid.
func (c *AgentConfig) Validate() error {
	if !c.Role.Valid() {
		return ErrInvalidAgentConfig
	}
	if c.Model == "" {
		return ErrInvalidAgentConfig
	}
	if c.BudgetLimit < 0 {
		return ErrInvalidAgentConfig
	}
	return nil
}

// AgentCapability defines capability flags for agents.
// Capabilities indicate what features or operations an agent supports.
type AgentCapability uint32

// Agent capability flags define specific features or operations an agent can perform.
const (
	// CapabilityCodeGeneration indicates the agent can generate code.
	CapabilityCodeGeneration AgentCapability = 1 << iota

	// CapabilityCodeReview indicates the agent can review code.
	CapabilityCodeReview

	// CapabilityArchitecture indicates the agent can make architectural decisions.
	CapabilityArchitecture

	// CapabilityTesting indicates the agent can write and run tests.
	CapabilityTesting

	// CapabilityDebugging indicates the agent can debug and fix issues.
	CapabilityDebugging

	// CapabilityDocumentation indicates the agent can write documentation.
	CapabilityDocumentation

	// CapabilityRefactoring indicates the agent can refactor code.
	CapabilityRefactoring

	// CapabilityDeployment indicates the agent can handle deployment tasks.
	CapabilityDeployment
)

// Has checks if the capability set includes the specified capability.
func (c AgentCapability) Has(capability AgentCapability) bool {
	return c&capability != 0
}

// Add returns a new capability set with the specified capability added.
func (c AgentCapability) Add(capability AgentCapability) AgentCapability {
	return c | capability
}

// Remove returns a new capability set with the specified capability removed.
func (c AgentCapability) Remove(capability AgentCapability) AgentCapability {
	return c &^ capability
}

// String returns a string representation of the capability set.
func (c AgentCapability) String() string {
	if c == 0 {
		return "none"
	}

	caps := []string{}
	if c.Has(CapabilityCodeGeneration) {
		caps = append(caps, "code-generation")
	}
	if c.Has(CapabilityCodeReview) {
		caps = append(caps, "code-review")
	}
	if c.Has(CapabilityArchitecture) {
		caps = append(caps, "architecture")
	}
	if c.Has(CapabilityTesting) {
		caps = append(caps, "testing")
	}
	if c.Has(CapabilityDebugging) {
		caps = append(caps, "debugging")
	}
	if c.Has(CapabilityDocumentation) {
		caps = append(caps, "documentation")
	}
	if c.Has(CapabilityRefactoring) {
		caps = append(caps, "refactoring")
	}
	if c.Has(CapabilityDeployment) {
		caps = append(caps, "deployment")
	}

	return fmt.Sprintf("[%s]", strings.Join(caps, ", "))
}
