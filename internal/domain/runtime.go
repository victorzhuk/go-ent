package domain

// Runtime defines the execution environment for agents.
type Runtime string

// Runtime constants define the available execution environments.
const (
	// RuntimeClaudeCode represents the Claude Code CLI environment.
	RuntimeClaudeCode Runtime = "claude-code"

	// RuntimeOpenCode represents the OpenCode environment.
	RuntimeOpenCode Runtime = "open-code"

	// RuntimeCLI represents the standalone CLI environment.
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
