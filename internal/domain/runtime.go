package domain

// Runtime defines the execution environment for agents.
type Runtime string

// Runtime constants define the available execution environments.
const (
	// RuntimeClaudeCode represents the Claude Code CLI environment.
	// Provides full filesystem access and interactive chat capabilities.
	RuntimeClaudeCode Runtime = "claude-code"

	// RuntimeOpenCode represents the OpenCode environment.
	// Alternative IDE integration with similar capabilities to Claude Code.
	RuntimeOpenCode Runtime = "open-code"

	// RuntimeCLI represents the standalone CLI environment.
	// Headless execution suitable for automation and scripting.
	RuntimeCLI Runtime = "cli"
)

// String returns the string representation of the runtime.
func (r Runtime) String() string {
	return string(r)
}

// Valid returns true if the runtime is valid.
func (r Runtime) Valid() bool {
	switch r {
	case RuntimeClaudeCode, RuntimeOpenCode, RuntimeCLI:
		return true
	default:
		return false
	}
}

// RuntimeCapability describes the features and limitations of a runtime environment.
type RuntimeCapability struct {
	// Runtime is the runtime this capability describes.
	Runtime Runtime

	// SupportsInteractive indicates whether the runtime supports interactive chat.
	SupportsInteractive bool

	// SupportsFileSystem indicates whether the runtime has filesystem access.
	SupportsFileSystem bool

	// SupportsTools indicates whether the runtime can invoke tools.
	SupportsTools bool

	// SupportsSkills indicates whether the runtime supports skill execution.
	SupportsSkills bool

	// MaxConcurrentAgents is the maximum number of agents that can run concurrently (0 = unlimited).
	MaxConcurrentAgents int

	// Description provides human-readable information about the runtime.
	Description string
}

// NewRuntimeCapability creates a RuntimeCapability for the given runtime with default values.
func NewRuntimeCapability(r Runtime) RuntimeCapability {
	switch r {
	case RuntimeClaudeCode, RuntimeOpenCode:
		return RuntimeCapability{
			Runtime:             r,
			SupportsInteractive: true,
			SupportsFileSystem:  true,
			SupportsTools:       true,
			SupportsSkills:      true,
			MaxConcurrentAgents: 0, // unlimited
			Description:         "Full-featured IDE integration with interactive chat and filesystem access",
		}
	case RuntimeCLI:
		return RuntimeCapability{
			Runtime:             r,
			SupportsInteractive: false,
			SupportsFileSystem:  true,
			SupportsTools:       true,
			SupportsSkills:      true,
			MaxConcurrentAgents: 1,
			Description:         "Headless CLI environment for automation and scripting",
		}
	default:
		return RuntimeCapability{
			Runtime:     r,
			Description: "Unknown runtime",
		}
	}
}

// CanRunAgent checks if the runtime can execute an agent with the given requirements.
func (rc *RuntimeCapability) CanRunAgent(requiresInteractive, requiresFileSystem bool) bool {
	if requiresInteractive && !rc.SupportsInteractive {
		return false
	}
	if requiresFileSystem && !rc.SupportsFileSystem {
		return false
	}
	return true
}

// HasFeature checks if a specific feature is supported.
func (rc *RuntimeCapability) HasFeature(feature string) bool {
	switch feature {
	case "interactive":
		return rc.SupportsInteractive
	case "filesystem":
		return rc.SupportsFileSystem
	case "tools":
		return rc.SupportsTools
	case "skills":
		return rc.SupportsSkills
	default:
		return false
	}
}
