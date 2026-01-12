package toolinit

import (
	"context"
	"io/fs"

	"github.com/victorzhuk/go-ent/internal/model"
)

// Adapter generates tool-specific configurations from embedded plugin resources
type Adapter interface {
	// Name returns the tool name (claude, opencode, cursor)
	Name() string

	// TargetDir returns the configuration directory name (.claude, .opencode, .cursor)
	TargetDir() string

	// Generate creates the tool configuration in the specified path
	Generate(ctx context.Context, cfg *GenerateConfig) error

	// TransformAgent transforms an agent file for the target tool
	TransformAgent(meta *AgentMeta) (string, error)

	// TransformCommand transforms a command file for the target tool
	TransformCommand(meta *CommandMeta) (string, error)

	// TransformSkill transforms a skill file for the target tool
	TransformSkill(meta *SkillMeta) (string, error)
}

// GenerateConfig configures the generation process
type GenerateConfig struct {
	// Path is the target project directory
	Path string

	// PluginFS is the embedded filesystem containing plugin resources
	PluginFS fs.FS

	// Agents is the list of agents to install (nil = all)
	Agents []string

	// Commands is the list of commands to install (nil = all)
	Commands []string

	// Skills is the list of skills to install (nil = all)
	Skills []string

	// Force overwrites existing files
	Force bool

	// DryRun previews changes without writing
	DryRun bool

	// MCPBinary is the path to the MCP server binary
	MCPBinary string

	// ModelOverrides maps tag patterns to model names (e.g., "heavy" -> "opus")
	ModelOverrides map[string]string

	// ModelConfig contains model category configuration
	ModelConfig *model.Config
}

// AgentTags categorizes agents by role and complexity
type AgentTags struct {
	Role       string `yaml:"role"`       // planning, execution, review, debug, test
	Complexity string `yaml:"complexity"` // light, standard, heavy
}

// AgentMeta contains parsed agent metadata
type AgentMeta struct {
	Name        string
	Description string
	Model       string
	Color       string
	Skills      []string
	Tools       []string
	Tags        AgentTags
	Body        string
	FilePath    string
}

// CommandMeta contains parsed command metadata
type CommandMeta struct {
	Name         string
	Description  string
	ArgumentHint string
	AllowedTools []string
	Body         string
	FilePath     string
}

// SkillMeta contains parsed skill metadata
type SkillMeta struct {
	Name        string
	Description string
	Body        string
	FilePath    string
}

// FileOperation represents a file write operation
type FileOperation struct {
	Path    string
	Content string
	Mode    fs.FileMode
}
