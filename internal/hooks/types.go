package hooks

// HookType specifies how a hook should be executed.
type HookType string

const (
	// HookTypeCommand executes a shell command.
	HookTypeCommand HookType = "command"
	// HookTypeAgent logs an agent suggestion (no auto-invocation).
	// Example output: "💡 Suggestion: Run /ent:reviewer to validate before archiving"
	HookTypeAgent HookType = "agent"
)

// Hook represents a single hook action to be executed.
type Hook struct {
	Type    HookType `yaml:"type" json:"type"`
	Command string   `yaml:"command,omitempty" json:"command,omitempty"` // For type=command
	Agent   string   `yaml:"agent,omitempty" json:"agent,omitempty"`     // For type=agent
	Prompt  string   `yaml:"prompt,omitempty" json:"prompt,omitempty"`   // For type=agent
}

// HookMatcher matches tool names against a pattern and executes hooks.
type HookMatcher struct {
	Matcher string `yaml:"matcher" json:"matcher"` // Tool name pattern (regex)
	Hooks   []Hook `yaml:"hooks" json:"hooks"`
}

// ToolHooks defines hooks for MCP tool lifecycle events.
type ToolHooks struct {
	PreToolUse  []HookMatcher `yaml:"PreToolUse,omitempty" json:"PreToolUse,omitempty"`
	PostToolUse []HookMatcher `yaml:"PostToolUse,omitempty" json:"PostToolUse,omitempty"`
	Stop        []HookMatcher `yaml:"Stop,omitempty" json:"Stop,omitempty"`
}

// OpenSpecHooks defines hooks for OpenSpec workflow lifecycle events.
type OpenSpecHooks struct {
	OnChangeCreated Hook `yaml:"onChangeCreated" json:"onChangeCreated"`
	OnTasksReady    Hook `yaml:"onTasksReady" json:"onTasksReady"`
	OnTaskStarted   Hook `yaml:"onTaskStarted" json:"onTaskStarted"`
	OnTaskCompleted Hook `yaml:"onTaskCompleted" json:"onTaskCompleted"`
	BeforeArchive   Hook `yaml:"beforeArchive" json:"beforeArchive"`
	AfterArchive    Hook `yaml:"afterArchive" json:"afterArchive"`
}

// HooksConfig is the root configuration for all hooks.
type HooksConfig struct {
	Tool     ToolHooks     `yaml:"tool,omitempty" json:"hooks,omitempty"`
	OpenSpec OpenSpecHooks `yaml:"openspec,omitempty" json:"openspec,omitempty"`
}
