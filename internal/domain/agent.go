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
	AgentRoleProduct    AgentRole = "product"
	AgentRoleArchitect  AgentRole = "architect"
	AgentRoleSenior     AgentRole = "senior"
	AgentRoleDeveloper  AgentRole = "developer"
	AgentRoleReviewer   AgentRole = "reviewer"
	AgentRoleOps        AgentRole = "ops"
	AgentRoleCoder      AgentRole = "coder"
	AgentRolePlanner    AgentRole = "planner"
	AgentRoleTester     AgentRole = "tester"
	AgentRoleDebugger   AgentRole = "debugger"
	AgentRoleResearcher AgentRole = "researcher"
	AgentRoleAcceptor   AgentRole = "acceptor"
	AgentRoleReproducer AgentRole = "reproducer"
	AgentRoleDecomposer AgentRole = "decomposer"
	AgentRoleTask       AgentRole = "task"
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
