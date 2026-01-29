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

// LoadAgentSource loads an agent from src/agents/{name}.yaml and src/agents/{name}.prompt.md
func LoadAgentSource(srcDir, name string) (*AgentSource, *PromptContent, error) {
	// Load agent metadata
	metaPath := filepath.Join(srcDir, "agents", name+".yaml")
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

// LoadPrompts loads all referenced prompt files
func LoadPrompts(srcDir string, cfg PromptsConfig) (*PromptContent, error) {
	pc := &PromptContent{
		Shared: make(map[string]string),
	}

	// Load shared prompts
	for _, name := range cfg.Shared {
		path := filepath.Join(srcDir, "prompts", name+".md")
		content, err := readFile(path)
		if err != nil {
			return nil, fmt.Errorf("read shared prompt %s: %w", name, err)
		}
		pc.Shared[name] = string(content)
	}

	// Load main agent prompt
	mainPath := filepath.Join(srcDir, "agents", cfg.Main+".prompt.md")
	content, err := readFile(mainPath)
	if err != nil {
		return nil, fmt.Errorf("read main prompt %s: %w", cfg.Main, err)
	}
	pc.Main = string(content)

	return pc, nil
}

// ListAgents returns all agent names in src/agents/
func ListAgents(srcDir string) ([]string, error) {
	agentsDir := filepath.Join(srcDir, "agents")
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
