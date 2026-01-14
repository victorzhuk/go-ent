package toolinit

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/victorzhuk/go-ent/internal/model"
	"github.com/victorzhuk/go-ent/internal/version"
	"gopkg.in/yaml.v3"
)

// OpenCodeAdapter implements the Adapter interface for OpenCode
type OpenCodeAdapter struct {
	cfg *GenerateConfig
}

// NewOpenCodeAdapter creates a new OpenCode adapter
func NewOpenCodeAdapter() *OpenCodeAdapter {
	return &OpenCodeAdapter{}
}

// Name returns the tool name
func (a *OpenCodeAdapter) Name() string {
	return "opencode"
}

// TargetDir returns the configuration directory name
func (a *OpenCodeAdapter) TargetDir() string {
	return ".opencode"
}

// Generate creates the OpenCode configuration
func (a *OpenCodeAdapter) Generate(ctx context.Context, cfg *GenerateConfig) error {
	a.cfg = cfg
	targetDir := filepath.Join(cfg.Path, a.TargetDir())

	// Check if target directory exists
	if !cfg.Force {
		if _, err := os.Stat(targetDir); err == nil {
			return fmt.Errorf("%s already exists (use --force to overwrite)", targetDir)
		}
	}

	// Collect file operations
	var ops []FileOperation

	// Generate commands (task.md, bug.md for OpenCode)
	commandOps, err := a.generateCommands(cfg)
	if err != nil {
		return fmt.Errorf("generate commands: %w", err)
	}
	ops = append(ops, commandOps...)

	// Generate agents (only execution agents for OpenCode)
	agentOps, err := a.generateAgents(cfg)
	if err != nil {
		return fmt.Errorf("generate agents: %w", err)
	}
	ops = append(ops, agentOps...)

	// Generate skills (all skills, flatten to skill/ directory)
	skillOps, err := a.generateSkills(cfg)
	if err != nil {
		return fmt.Errorf("generate skills: %w", err)
	}
	ops = append(ops, skillOps...)

	// Execute file operations
	if cfg.DryRun {
		fmt.Println("DRY RUN - would create:")
		for _, op := range ops {
			fmt.Printf("  %s\n", op.Path)
		}
		return nil
	}

	for _, op := range ops {
		targetPath := filepath.Join(targetDir, op.Path)
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return fmt.Errorf("create directory %s: %w", filepath.Dir(targetPath), err)
		}
		if err := os.WriteFile(targetPath, []byte(op.Content), op.Mode); err != nil {
			return fmt.Errorf("write file %s: %w", targetPath, err)
		}
	}

	// Save version info
	v := version.Get()
	info := &EntInfo{
		Version:     v.Version,
		VCSRef:      v.VCSRef,
		InstalledAt: time.Now(),
		Components:  BuildComponentManifest(ops),
	}

	if err := SaveEntInfo(targetDir, info); err != nil {
		return fmt.Errorf("save version info: %w", err)
	}

	return nil
}

// generateCommands generates command files for OpenCode (SINGULAR: command/)
func (a *OpenCodeAdapter) generateCommands(cfg *GenerateConfig) ([]FileOperation, error) {
	var ops []FileOperation

	err := fs.WalkDir(cfg.PluginFS, "plugins/go-ent/commands", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		filename := filepath.Base(path)

		// Check if this command is in the filter list (if provided)
		if len(cfg.Commands) > 0 {
			cmdName := FileNameWithoutExt(filename)
			found := false
			for _, c := range cfg.Commands {
				if c == cmdName {
					found = true
					break
				}
			}
			if !found {
				return nil
			}
		}

		// Read file content
		content, err := fs.ReadFile(cfg.PluginFS, path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}

		// Parse and transform
		meta, err := ParseCommandFile(string(content), path)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}

		transformed, err := a.TransformCommand(meta)
		if err != nil {
			return fmt.Errorf("transform %s: %w", path, err)
		}

		// OpenCode uses SINGULAR: command/ (not commands/)
		ops = append(ops, FileOperation{
			Path:    filepath.Join("command", "ent", filename),
			Content: transformed,
			Mode:    0644,
		})

		return nil
	})

	return ops, err
}

// generateAgents generates agent files for OpenCode (SINGULAR: agent/)
func (a *OpenCodeAdapter) generateAgents(cfg *GenerateConfig) ([]FileOperation, error) {
	var ops []FileOperation

	// First, check for agents in split format (meta/*.yaml)
	metaFiles, err := fs.Glob(cfg.PluginFS, "plugins/go-ent/agents/meta/*.yaml")
	if err != nil {
		return nil, fmt.Errorf("list meta files: %w", err)
	}

	processedAgents := make(map[string]bool)

	// Process split format agents
	for _, metaPath := range metaFiles {
		filename := filepath.Base(metaPath)
		agentName := FileNameWithoutExt(filename)

		// Check if this agent is in the filter list (if provided)
		if len(cfg.Agents) > 0 {
			found := false
			for _, a := range cfg.Agents {
				if a == agentName {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		// Load metadata
		meta, err := a.loadAgentMetadata(agentName)
		if err != nil {
			return nil, fmt.Errorf("load metadata for %s: %w", agentName, err)
		}

		// Apply model overrides if configured
		if len(cfg.ModelOverrides) > 0 {
			resolver := NewModelResolver(cfg.ModelOverrides)
			meta.Model = resolver.Resolve(meta)
		}

		transformed, err := a.TransformAgent(meta)
		if err != nil {
			return nil, fmt.Errorf("transform %s: %w", agentName, err)
		}

		// OpenCode uses SINGULAR: agent/ (not agents/)
		outputFilename := agentName + ".md"
		ops = append(ops, FileOperation{
			Path:    filepath.Join("agent", "ent", outputFilename),
			Content: transformed,
			Mode:    0644,
		})

		processedAgents[agentName] = true
	}

	// Process single-file format agents (backward compatibility)
	err = fs.WalkDir(cfg.PluginFS, "plugins/go-ent/agents", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-markdown files
		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		// Skip meta and templates directories
		if strings.Contains(path, "/meta/") || strings.Contains(path, "/templates/") || strings.Contains(path, "/prompts/") {
			return nil
		}

		filename := filepath.Base(path)
		agentName := FileNameWithoutExt(filename)

		// Skip if already processed in split format
		if processedAgents[agentName] {
			return nil
		}

		// Check if this agent is in the filter list (if provided)
		if len(cfg.Agents) > 0 {
			found := false
			for _, a := range cfg.Agents {
				if a == agentName {
					found = true
					break
				}
			}
			if !found {
				return nil
			}
		}

		// Read file content
		content, err := fs.ReadFile(cfg.PluginFS, path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}

		// Parse and transform (single-file format)
		meta, err := ParseAgentFile(string(content), path)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}

		// Apply model overrides if configured
		if len(cfg.ModelOverrides) > 0 {
			resolver := NewModelResolver(cfg.ModelOverrides)
			meta.Model = resolver.Resolve(meta)
		}

		transformed, err := a.TransformAgent(meta)
		if err != nil {
			return fmt.Errorf("transform %s: %w", path, err)
		}

		// OpenCode uses SINGULAR: agent/ (not agents/)
		ops = append(ops, FileOperation{
			Path:    filepath.Join("agent", "ent", filename),
			Content: transformed,
			Mode:    0644,
		})

		return nil
	})

	return ops, err
}

// generateSkills generates skill files for OpenCode (SINGULAR: skill/)
func (a *OpenCodeAdapter) generateSkills(cfg *GenerateConfig) ([]FileOperation, error) {
	var ops []FileOperation

	// OpenCode gets all skills, but flattened structure
	// Can also read from .claude/skills/ for compatibility, so we can skip this
	// or create flattened structure with category prefix in name

	err := fs.WalkDir(cfg.PluginFS, "plugins/go-ent/skills", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, "SKILL.md") {
			return nil
		}

		// Extract category and skill name from path
		// Path format: plugins/go-ent/skills/{category}/{skill}/SKILL.md
		parts := strings.Split(path, "/")
		if len(parts) < 5 {
			return fmt.Errorf("invalid skill path: %s", path)
		}
		category := parts[3]
		skillName := parts[4]

		// Check if this skill is in the filter list (if provided)
		if len(cfg.Skills) > 0 {
			found := false
			for _, s := range cfg.Skills {
				if s == skillName || s == category+"/"+skillName {
					found = true
					break
				}
			}
			if !found {
				return nil
			}
		}

		// Read file content
		content, err := fs.ReadFile(cfg.PluginFS, path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}

		// Parse and transform
		meta, err := ParseSkillFile(string(content), path)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}

		transformed, err := a.TransformSkill(meta)
		if err != nil {
			return fmt.Errorf("transform %s: %w", path, err)
		}

		// OpenCode uses SINGULAR flat structure: skill/ent/{name}/SKILL.md
		// Use category-prefixed name for uniqueness: skill/ent/core-arch-core/SKILL.md
		flatName := category + "-" + skillName
		ops = append(ops, FileOperation{
			Path:    filepath.Join("skill", "ent", flatName, "SKILL.md"),
			Content: transformed,
			Mode:    0644,
		})

		return nil
	})

	return ops, err
}

// TransformAgent transforms an agent file for OpenCode
func (a *OpenCodeAdapter) TransformAgent(meta *AgentMeta) (string, error) {
	// Check if agent uses split format (has prompts configured)
	if meta.Prompts.Main != "" {
		return a.transformAgentSplit(meta)
	}

	// Single-file format (backward compatibility)
	return a.transformAgentSingle(meta)
}

// transformAgentSplit transforms a split-format agent using composer and template
func (a *OpenCodeAdapter) transformAgentSplit(meta *AgentMeta) (string, error) {
	// Add ent- prefix to agent name
	meta.Name = "ent-" + meta.Name

	// Compose prompt from shared sections and agent-specific prompt
	body, err := a.composePrompt(meta)
	if err != nil {
		return "", fmt.Errorf("compose prompt: %w", err)
	}

	// Apply OpenCode template to generate frontmatter
	frontmatter, err := a.applyTemplate(meta)
	if err != nil {
		return "", fmt.Errorf("apply template: %w", err)
	}

	// Combine template output (frontmatter) with composed prompt (body)
	return "---\n" + frontmatter + "---\n\n" + body, nil
}

// transformAgentSingle transforms a single-file agent (backward compatibility)
func (a *OpenCodeAdapter) transformAgentSingle(meta *AgentMeta) (string, error) {
	// OpenCode agent frontmatter format
	metadata := make(map[string]interface{})

	// Add ent- prefix to agent name
	if meta.Name != "" {
		metadata["name"] = "ent-" + meta.Name
	}
	if meta.Description != "" {
		metadata["description"] = meta.Description
	}

	// Mode: all agents are subagents in OpenCode
	metadata["mode"] = "subagent"

	// Use model resolver to map categories to OpenCode model IDs
	if meta.Model != "" {
		var modelConfig *model.Config
		if a.cfg != nil {
			modelConfig = a.cfg.ModelConfig
		}
		resolver := model.NewResolver(modelConfig, "opencode")
		metadata["model"] = resolver.ResolveAgent(meta.Model)
	}

	// Temperature: 0.0 for deterministic code generation
	metadata["temperature"] = 0.0

	// Tools: disable write/edit by default (command controls this)
	if len(meta.Tools) > 0 {
		tools := make(map[string]bool)
		for _, tool := range meta.Tools {
			// Map tool names to OpenCode format
			switch tool {
			case "read", "glob", "grep", "bash":
				tools[tool] = true
			case "write", "edit":
				tools[tool] = false // Disabled by default
			}
		}
		if len(tools) > 0 {
			metadata["tools"] = tools
		}
	}

	// Permission: skill patterns
	if len(meta.Skills) > 0 {
		permission := make(map[string]interface{})
		skillPerms := make(map[string]string)
		for _, skill := range meta.Skills {
			// Allow specific skills
			skillPerms[skill] = "allow"
		}
		permission["skill"] = skillPerms
		metadata["permission"] = permission
	}

	frontmatter := GenerateFrontmatter(metadata)
	return frontmatter + "\n" + meta.Body, nil
}

// TransformCommand transforms a command file for OpenCode
func (a *OpenCodeAdapter) TransformCommand(meta *CommandMeta) (string, error) {
	// OpenCode command frontmatter format
	metadata := make(map[string]interface{})

	// Derive command name from filepath if not in meta
	if meta.Name == "" {
		meta.Name = FileBaseName(meta.FilePath)
	}

	// Add ent: prefix to command name
	metadata["name"] = "ent:" + meta.Name
	if meta.Description != "" {
		metadata["description"] = meta.Description
	}

	// Commands in OpenCode are templates with $ARGUMENTS
	frontmatter := GenerateFrontmatter(metadata)
	return frontmatter + "\n" + meta.Body, nil
}

// TransformSkill transforms a skill file for OpenCode
func (a *OpenCodeAdapter) TransformSkill(meta *SkillMeta) (string, error) {
	// OpenCode skill frontmatter format (same as Claude Code)
	metadata := make(map[string]interface{})

	if meta.Name != "" {
		metadata["name"] = meta.Name
	}
	if meta.Description != "" {
		metadata["description"] = meta.Description
	}

	frontmatter := GenerateFrontmatter(metadata)
	return frontmatter + "\n" + meta.Body, nil
}

// composePrompt composes the agent prompt from shared sections and agent-specific prompt
func (a *OpenCodeAdapter) composePrompt(meta *AgentMeta) (string, error) {
	var result strings.Builder

	// Add shared prompt sections
	for _, name := range meta.Prompts.Shared {
		path := fmt.Sprintf("plugins/go-ent/agents/prompts/shared/%s.md", name)
		content, err := fs.ReadFile(a.cfg.PluginFS, path)
		if err != nil {
			return "", fmt.Errorf("shared section not found: %s", name)
		}
		result.WriteString(string(content))
		result.WriteString("\n\n")
	}

	// Add agent-specific prompt
	agentPrompt, err := fs.ReadFile(a.cfg.PluginFS, meta.Prompts.Main)
	if err != nil {
		return "", fmt.Errorf("read agent prompt for %s: %w", meta.Name, err)
	}

	result.WriteString(string(agentPrompt))

	return result.String(), nil
}

// loadAgentMetadata loads agent metadata from split format YAML file
func (a *OpenCodeAdapter) loadAgentMetadata(agentName string) (*AgentMeta, error) {
	metaPath := fmt.Sprintf("plugins/go-ent/agents/meta/%s.yaml", agentName)
	content, err := fs.ReadFile(a.cfg.PluginFS, metaPath)
	if err != nil {
		return nil, fmt.Errorf("read agent metadata: %w", err)
	}

	var meta AgentMeta
	if err := yaml.Unmarshal(content, &meta); err != nil {
		return nil, fmt.Errorf("parse agent metadata: %w", err)
	}

	return &meta, nil
}

// applyTemplate applies the OpenCode template to generate frontmatter
func (a *OpenCodeAdapter) applyTemplate(meta *AgentMeta) (string, error) {
	tmplPath := "plugins/go-ent/agents/templates/opencode.yaml.tmpl"
	tmplContent, err := fs.ReadFile(a.cfg.PluginFS, tmplPath)
	if err != nil {
		return "", fmt.Errorf("read template: %w", err)
	}

	tmpl, err := template.New("opencode").Parse(string(tmplContent))
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, meta); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return buf.String(), nil
}
