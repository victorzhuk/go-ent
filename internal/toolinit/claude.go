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
)

// ClaudeAdapter implements the Adapter interface for Claude Code
type ClaudeAdapter struct {
	cfg *GenerateConfig
}

// NewClaudeAdapter creates a new Claude Code adapter
func NewClaudeAdapter() *ClaudeAdapter {
	return &ClaudeAdapter{}
}

// Name returns the tool name
func (a *ClaudeAdapter) Name() string {
	return "claude"
}

// TargetDir returns the configuration directory name
func (a *ClaudeAdapter) TargetDir() string {
	return ".claude"
}

// Generate creates the Claude Code configuration
func (a *ClaudeAdapter) Generate(ctx context.Context, cfg *GenerateConfig) error {
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

	// Generate commands (only plan.md for Claude)
	commandOps, err := a.generateCommands(cfg)
	if err != nil {
		return fmt.Errorf("generate commands: %w", err)
	}
	ops = append(ops, commandOps...)

	// Generate agents (only planning agents for Claude)
	agentOps, err := a.generateAgents(cfg)
	if err != nil {
		return fmt.Errorf("generate agents: %w", err)
	}
	ops = append(ops, agentOps...)

	// Generate skills (all skills, preserve categories)
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

// generateCommands generates command files for Claude Code
func (a *ClaudeAdapter) generateCommands(cfg *GenerateConfig) ([]FileOperation, error) {
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

		ops = append(ops, FileOperation{
			Path:    filepath.Join("commands", "ent", filename),
			Content: transformed,
			Mode:    0644,
		})

		return nil
	})

	return ops, err
}

// generateAgents generates agent files for Claude Code
func (a *ClaudeAdapter) generateAgents(cfg *GenerateConfig) ([]FileOperation, error) {
	var ops []FileOperation

	// Process split-format agents (meta/*.yaml)
	metaEntries, err := fs.ReadDir(cfg.PluginFS, "plugins/go-ent/agents/meta")
	if err == nil {
		for _, entry := range metaEntries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
				continue
			}

			agentName := FileNameWithoutExt(entry.Name())

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

			// Read metadata YAML
			metaPath := filepath.Join("plugins/go-ent/agents/meta", entry.Name())
			metaContent, err := fs.ReadFile(cfg.PluginFS, metaPath)
			if err != nil {
				return nil, fmt.Errorf("read %s: %w", metaPath, err)
			}

			// Parse metadata
			meta, err := ParseAgentMetaYAML(string(metaContent), metaPath)
			if err != nil {
				return nil, fmt.Errorf("parse metadata %s: %w", metaPath, err)
			}

			// Set prompt paths for composer
			meta.Prompts.Main = filepath.Join("plugins/go-ent/agents/prompts/agents", agentName+".md")
			meta.Prompts.Shared = []string{
				"_tooling",
				"_conventions",
				"_handoffs",
				"_openspec",
			}

			// Apply model overrides if configured
			if len(cfg.ModelOverrides) > 0 {
				resolver := NewModelResolver(cfg.ModelOverrides)
				meta.Model = resolver.Resolve(meta)
			}

			// Transform using composer + template
			transformed, err := a.TransformAgentWithComposer(cfg.PluginFS, meta)
			if err != nil {
				return nil, fmt.Errorf("transform split-format agent %s: %w", agentName, err)
			}

			ops = append(ops, FileOperation{
				Path:    filepath.Join("agents", "ent", agentName+".md"),
				Content: transformed,
				Mode:    0644,
			})
		}
	}

	return ops, nil
}

// generateSkills generates skill files for Claude Code
func (a *ClaudeAdapter) generateSkills(cfg *GenerateConfig) ([]FileOperation, error) {
	var ops []FileOperation

	// Claude Code gets all skills, preserving category structure
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

		// Preserve category structure: skills/ent/{category}/{skill}/SKILL.md
		ops = append(ops, FileOperation{
			Path:    filepath.Join("skills", "ent", category, skillName, "SKILL.md"),
			Content: transformed,
			Mode:    0644,
		})

		return nil
	})

	return ops, err
}

// TransformAgent transforms an agent file for Claude Code (legacy format)
func (a *ClaudeAdapter) TransformAgent(meta *AgentMeta) (string, error) {
	// Claude Code agent frontmatter format
	metadata := make(map[string]interface{})

	// Add ent- prefix to agent name
	if meta.Name != "" {
		metadata["name"] = "ent-" + meta.Name
	}
	if meta.Description != "" {
		metadata["description"] = meta.Description
	}
	if meta.Model != "" {
		// Use model resolver to map categories to Claude API model IDs
		var modelConfig *model.Config
		if a.cfg != nil {
			modelConfig = a.cfg.ModelConfig
		}
		resolver := model.NewResolver(modelConfig, "claude")
		metadata["model"] = resolver.ResolveAgent(meta.Model)
	}
	if meta.Color != "" {
		metadata["color"] = meta.Color
	}
	if len(meta.Skills) > 0 {
		metadata["skills"] = meta.Skills
	}
	if len(meta.Tools) > 0 {
		// Tools in Claude Code format (map from internal names)
		metadata["tools"] = meta.Tools
	}

	frontmatter := GenerateFrontmatter(metadata)
	return frontmatter + "\n" + meta.Body, nil
}

// TransformAgentWithComposer transforms an agent using composer and template (new format)
func (a *ClaudeAdapter) TransformAgentWithComposer(pluginFS fs.FS, meta *AgentMeta) (string, error) {
	// Load and compose prompt
	composer := NewPromptComposer(pluginFS)
	body, err := composer.Compose(meta)
	if err != nil {
		return "", fmt.Errorf("compose prompt: %w", err)
	}

	// Prepare template data
	data := struct {
		Name         string
		Description  string
		Model        string
		Color        string
		Skills       []string
		Tools        []string
		Dependencies []string
		Role         string
		Complexity   string
	}{
		Name:         "ent-" + meta.Name,
		Description:  meta.Description,
		Color:        meta.Color,
		Skills:       meta.Skills,
		Tools:        meta.Tools,
		Dependencies: meta.Dependencies,
		Role:         meta.Tags.Role,
		Complexity:   meta.Tags.Complexity,
	}

	// Resolve model
	if meta.Model != "" {
		var modelConfig *model.Config
		if a.cfg != nil {
			modelConfig = a.cfg.ModelConfig
		}
		resolver := model.NewResolver(modelConfig, "claude")
		data.Model = resolver.ResolveAgent(meta.Model)
	}

	// Load template
	tmplContent, err := fs.ReadFile(pluginFS, "plugins/go-ent/agents/templates/claude.yaml.tmpl")
	if err != nil {
		return "", fmt.Errorf("load template: %w", err)
	}

	// Parse template
	tmpl, err := template.New("claude").Parse(string(tmplContent))
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	// Execute template
	var frontmatterBuf strings.Builder
	if err := tmpl.Execute(&frontmatterBuf, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return frontmatterBuf.String() + "\n" + body, nil
}

// TransformCommand transforms a command file for Claude Code
func (a *ClaudeAdapter) TransformCommand(meta *CommandMeta) (string, error) {
	// Process include patterns in command body
	body := meta.Body
	if a.cfg != nil {
		body = processIncludes(meta.Body, a.cfg.PluginFS)
	}

	// Claude Code command frontmatter format
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

	frontmatter := GenerateFrontmatter(metadata)
	return frontmatter + "\n" + body, nil
}

// TransformSkill transforms a skill file for Claude Code
func (a *ClaudeAdapter) TransformSkill(meta *SkillMeta) (string, error) {
	// Claude Code skill frontmatter format
	metadata := make(map[string]interface{})

	if meta.Name != "" {
		metadata["name"] = meta.Name
	}
	if meta.Description != "" {
		metadata["description"] = meta.Description
	}

	// Add version if not present
	metadata["version"] = "1.0.0"

	frontmatter := GenerateFrontmatter(metadata)
	return frontmatter + "\n" + meta.Body, nil
}
