package generator

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/victorzhuk/go-ent/pkg"
	"gopkg.in/yaml.v3"
)

// readFile tries local filesystem first, falls back to embedded
func readFile(path string) ([]byte, error) {
	// Try local filesystem first (allows project customization)
	// #nosec G304 - path is validated by caller
	if data, err := os.ReadFile(path); err == nil {
		return data, nil
	}
	// Fall back to embedded sources
	return pkg.FS.ReadFile(path)
}

// listDir lists directory from local or embedded FS
func listDir(path string) ([]fs.DirEntry, error) {
	// Try local first
	if entries, err := os.ReadDir(path); err == nil {
		return entries, nil
	}
	// Fall back to embedded
	return pkg.FS.ReadDir(path)
}

// LoadAgentSource loads an agent from srcDir/{name}.yaml
// srcDir should be the agents directory (e.g., "agents" or "agents/meta")
func LoadAgentSource(srcDir, name string) (*AgentSource, *PromptContent, error) {
	// Load agent metadata
	metaPath := filepath.Join(srcDir, name+".yaml")
	data, err := readFile(metaPath)
	if err != nil {
		return nil, nil, fmt.Errorf("read agent meta %s: %w", metaPath, err)
	}

	var agent AgentSource
	if err := yaml.Unmarshal(data, &agent); err != nil {
		return nil, nil, fmt.Errorf("unmarshal agent meta: %w", err)
	}

	// Load prompts
	prompts, err := LoadPrompts(srcDir, agent.Prompts)
	if err != nil {
		return nil, nil, fmt.Errorf("load prompts: %w", err)
	}

	return &agent, prompts, nil
}

// LoadAgentMetaSource loads an agent from the meta format (srcDir should be agents/meta)
func LoadAgentMetaSource(srcDir, name string) (*AgentMetaSource, *PromptContent, error) {
	// Load agent metadata
	metaPath := filepath.Join(srcDir, name+".yaml")
	data, err := readFile(metaPath)
	if err != nil {
		return nil, nil, fmt.Errorf("read agent meta %s: %w", metaPath, err)
	}

	var agent AgentMetaSource
	if err := yaml.Unmarshal(data, &agent); err != nil {
		return nil, nil, fmt.Errorf("unmarshal agent meta: %w", err)
	}

	// Load prompts (srcDir is "agents/meta", so we need to go up to "agents" for prompts)
	// Extract base dir by removing last component
	baseDir := filepath.Dir(srcDir)
	prompts, err := LoadPrompts(baseDir, agent.Prompts)
	if err != nil {
		return nil, nil, fmt.Errorf("load prompts: %w", err)
	}

	return &agent, prompts, nil
}

// ConvertMetaToSource converts meta format to old AgentSource format for generation
func ConvertMetaToSource(meta *AgentMetaSource) *AgentSource {
	// Map model names: main->sonnet, fast->haiku, heavy->opus
	modelMap := map[string]string{
		"main":  "sonnet",
		"fast":  "haiku",
		"heavy": "opus",
	}

	claudeModel := modelMap[meta.Model]
	if claudeModel == "" {
		claudeModel = "sonnet" // default
	}

	// Build extended description if whenToUse is provided
	description := meta.Description
	if meta.WhenToUse != "" {
		description = meta.Description + "\n\n" + meta.WhenToUse
	}

	// For old format, both Claude and OpenCode use same model string
	agent := &AgentSource{
		Name:        meta.Name,
		Description: description,
		Model: ModelConfig{
			Claude:   claudeModel,
			OpenCode: meta.Model, // OpenCode uses main/fast/heavy directly
		},
		Skills:          meta.Skills,
		Prompts:         meta.Prompts,
		Color:           meta.Color,
		ComplexityHints: meta.ComplexityHints,
		ModelMapping:    meta.ModelMapping,
	}

	// Convert tool presets to tool configurations
	// This is a simplified conversion - full implementation would map presets to actual tools
	agent.Tools = convertToolPresetsToConfig(meta.ToolPresets, meta.DisallowedToolPresets)

	return agent
}

// convertToolPresetsToConfig converts tool presets to tool configuration
func convertToolPresetsToConfig(presets, disallowedPresets []string) ToolsConfig {
	cfg := ToolsConfig{
		Claude: ClaudeTools{
			Allowed:    []string{},
			Disallowed: []string{},
		},
		OpenCode: make(OpenCodeTools),
	}

	// Map presets to tool lists
	// This is simplified - a full implementation would have complete mappings
	hasEditing := contains(presets, "editing")
	hasReadonly := contains(presets, "readonly")
	hasSerenaEditing := contains(disallowedPresets, "serena-editing")

	if hasEditing {
		cfg.Claude.Allowed = []string{"Read", "Write", "Edit", "Bash", "Glob", "Grep"}
		cfg.OpenCode["read"] = true
		cfg.OpenCode["write"] = true
		cfg.OpenCode["edit"] = true
		cfg.OpenCode["bash"] = true
		cfg.OpenCode["glob"] = true
		cfg.OpenCode["grep"] = true
	}

	if hasReadonly {
		cfg.Claude.Allowed = []string{"Read", "Bash", "Glob", "Grep"}
		cfg.OpenCode["read"] = true
		cfg.OpenCode["bash"] = true
		cfg.OpenCode["glob"] = true
		cfg.OpenCode["grep"] = true
	}

	if hasSerenaEditing {
		cfg.Claude.Disallowed = []string{
			"mcp__plugin_serena_serena__replace_symbol_body",
			"mcp__plugin_serena_serena__insert_after_symbol",
			"mcp__plugin_serena_serena__insert_before_symbol",
			"mcp__plugin_serena_serena__replace_content",
			"mcp__plugin_serena_serena__create_text_file",
		}
	}

	return cfg
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// LoadPrompts loads all referenced prompt files
func LoadPrompts(srcDir string, cfg PromptsConfig) (*PromptContent, error) {
	pc := &PromptContent{
		Shared: make(map[string]string),
	}

	// Load shared prompts
	// Shared prompts may have underscore prefix in meta format
	for _, name := range cfg.Shared {
		// Add underscore prefix if not present (for meta format compatibility)
		promptName := name
		if name[0] != '_' {
			promptName = "_" + name
		}
		path := filepath.Join(srcDir, "prompts", "shared", promptName+".md")
		content, err := readFile(path)
		if err != nil {
			return nil, fmt.Errorf("read shared prompt %s: %w", name, err)
		}
		pc.Shared[name] = string(content)
	}

	// Load main agent prompt
	// cfg.Main can be:
	//   - just a name like "coder" (old format) -> prompts/agents/coder.md
	//   - a path like "agents/coder" (meta format) -> prompts/agents/coder.md
	mainPath := cfg.Main
	if !strings.Contains(mainPath, "/") {
		// Old format: add "agents/" prefix
		mainPath = filepath.Join("agents", mainPath)
	}
	mainPath = filepath.Join(srcDir, "prompts", mainPath+".md")
	content, err := readFile(mainPath)
	if err != nil {
		return nil, fmt.Errorf("read main prompt %s: %w", cfg.Main, err)
	}
	pc.Main = string(content)

	return pc, nil
}

// ListAgents returns all agent names in srcDir
// srcDir should be the agents directory (e.g., "agents" or "agents/meta")
func ListAgents(srcDir string) ([]string, error) {
	agentsDir := srcDir
	entries, err := listDir(agentsDir)
	if err != nil {
		return nil, fmt.Errorf("read agents dir: %w", err)
	}

	var agents []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if filepath.Ext(name) == ".yaml" {
			// Remove .yaml extension
			agentName := name[:len(name)-5]
			agents = append(agents, agentName)
		}
	}

	return agents, nil
}

// LoadSkillSource loads a skill from skills/{category}/{name}/SKILL.md
// Skills are markdown files with YAML frontmatter
func LoadSkillSource(skillsDir, category, name string) (*SkillSource, error) {
	skillPath := filepath.Join(skillsDir, category, name, "SKILL.md")
	data, err := readFile(skillPath)
	if err != nil {
		return nil, fmt.Errorf("read skill %s: %w", skillPath, err)
	}

	// Parse frontmatter and content
	content := string(data)

	// Split frontmatter and content
	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid skill format: missing frontmatter in %s", skillPath)
	}

	frontmatter := parts[1]
	skillContent := strings.TrimSpace(parts[2])

	// Parse YAML frontmatter
	var skill SkillSource
	if err := yaml.Unmarshal([]byte(frontmatter), &skill); err != nil {
		return nil, fmt.Errorf("unmarshal skill frontmatter: %w", err)
	}

	skill.Content = skillContent
	return &skill, nil
}

// ListSkills lists all available skills by category and name
// Returns map[category][]name
func ListSkills(skillsDir string) (map[string][]string, error) {
	categories, err := listDir(skillsDir)
	if err != nil {
		return nil, fmt.Errorf("read skills dir: %w", err)
	}

	result := make(map[string][]string)

	for _, catEntry := range categories {
		if !catEntry.IsDir() {
			continue
		}

		category := catEntry.Name()
		categoryPath := filepath.Join(skillsDir, category)

		skills, err := listDir(categoryPath)
		if err != nil {
			continue // Skip categories we can't read
		}

		var skillNames []string
		for _, skillEntry := range skills {
			if !skillEntry.IsDir() {
				continue
			}

			skillName := skillEntry.Name()
			// Verify SKILL.md exists
			skillFile := filepath.Join(categoryPath, skillName, "SKILL.md")
			if _, err := readFile(skillFile); err == nil {
				skillNames = append(skillNames, skillName)
			}
		}

		if len(skillNames) > 0 {
			result[category] = skillNames
		}
	}

	return result, nil
}

func MergeSkillDirs(primaryDir string, extraDirs ...string) (map[string][]string, error) {
	seen := make(map[string]bool)
	result := make(map[string][]string)

	primary, err := ListSkills(primaryDir)
	if err == nil {
		for cat, names := range primary {
			for _, name := range names {
				key := cat + "/" + name
				if !seen[key] {
					seen[key] = true
					result[cat] = append(result[cat], name)
				}
			}
		}
	}

	for _, dir := range extraDirs {
		extra, err := ListSkills(dir)
		if err != nil {
			continue
		}
		for cat, names := range extra {
			for _, name := range names {
				key := cat + "/" + name
				if !seen[key] {
					seen[key] = true
					result[cat] = append(result[cat], name)
				}
			}
		}
	}

	return result, nil
}
