package toolinit

import (
	"bytes"
	"fmt"
	"io/fs"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// PromptComposer composes agent prompts from shared and agent-specific sections
type PromptComposer struct {
	fs fs.FS
}

func NewPromptComposer(fs fs.FS) *PromptComposer {
	return &PromptComposer{fs: fs}
}

func (c *PromptComposer) Compose(meta *AgentMeta) (string, error) {
	var result strings.Builder

	for _, name := range meta.Prompts.Shared {
		path := fmt.Sprintf("plugins/go-ent/agents/prompts/shared/%s.md", name)
		content, err := fs.ReadFile(c.fs, path)
		if err != nil {
			return "", fmt.Errorf("shared section not found: %s: %w", name, err)
		}
		result.WriteString(string(content))
		result.WriteString("\n\n")
	}

	agentPrompt, err := fs.ReadFile(c.fs, meta.Prompts.Main)
	if err != nil {
		return "", fmt.Errorf("compose prompt for %s: %w", meta.Name, err)
	}

	result.WriteString(string(agentPrompt))

	return result.String(), nil
}

// ParseFrontmatter extracts YAML frontmatter and body from markdown content
func ParseFrontmatter(content string) (map[string]interface{}, string, error) {
	if !strings.HasPrefix(content, "---\n") {
		return nil, content, nil
	}

	parts := strings.SplitN(content[4:], "\n---\n", 2)
	if len(parts) != 2 {
		return nil, content, fmt.Errorf("invalid frontmatter format")
	}

	var metadata map[string]interface{}
	if err := yaml.Unmarshal([]byte(parts[0]), &metadata); err != nil {
		return nil, "", fmt.Errorf("parse frontmatter: %w", err)
	}

	return metadata, strings.TrimSpace(parts[1]), nil
}

// ParseCommandFile parses a command markdown file
func ParseCommandFile(content, filePath string) (*CommandMeta, error) {
	metadata, body, err := ParseFrontmatter(content)
	if err != nil {
		return nil, err
	}

	meta := &CommandMeta{
		Body:     body,
		FilePath: filePath,
	}

	if desc, ok := metadata["description"].(string); ok {
		meta.Description = desc
	}
	if hint, ok := metadata["argument-hint"].(string); ok {
		meta.ArgumentHint = hint
	}

	if tools, ok := metadata["allowed-tools"].([]interface{}); ok {
		for _, t := range tools {
			if str, ok := t.(string); ok {
				meta.AllowedTools = append(meta.AllowedTools, str)
			}
		}
	}

	return meta, nil
}

// ParseSkillFile parses a skill SKILL.md file
func ParseSkillFile(content, filePath string) (*SkillMeta, error) {
	metadata, body, err := ParseFrontmatter(content)
	if err != nil {
		return nil, err
	}

	meta := &SkillMeta{
		Body:     body,
		FilePath: filePath,
	}

	if name, ok := metadata["name"].(string); ok {
		meta.Name = name
	}
	if desc, ok := metadata["description"].(string); ok {
		meta.Description = desc
	}

	return meta, nil
}

// GenerateFrontmatter creates YAML frontmatter from a map
func GenerateFrontmatter(metadata map[string]interface{}) string {
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	encoder.Encode(metadata)
	encoder.Close()

	return "---\n" + buf.String() + "---\n"
}

// FileNameWithoutExt removes the extension from a filename
func FileNameWithoutExt(filename string) string {
	if idx := strings.LastIndex(filename, "."); idx != -1 {
		return filename[:idx]
	}
	return filename
}

// FileBaseName extracts the base name from a path (last component without extension)
func FileBaseName(path string) string {
	parts := strings.Split(path, "/")
	name := parts[len(parts)-1]
	return FileNameWithoutExt(name)
}

// ParseAgentMetaYAML parses agent metadata from a YAML file (split format)
func ParseAgentMetaYAML(content string, filePath string) (*AgentMeta, error) {
	var meta struct {
		Name         string   `yaml:"name"`
		Description  string   `yaml:"description"`
		Model        string   `yaml:"model"`
		Color        string   `yaml:"color"`
		Skills       []string `yaml:"skills"`
		Tools        []string `yaml:"tools"`
		Dependencies []string `yaml:"dependencies"`
		Tags         []string `yaml:"tags"`
	}

	if err := yaml.Unmarshal([]byte(content), &meta); err != nil {
		return nil, fmt.Errorf("parse agent metadata: %w", err)
	}

	result := &AgentMeta{
		Name:         meta.Name,
		Description:  meta.Description,
		Model:        meta.Model,
		Color:        meta.Color,
		Skills:       meta.Skills,
		Tools:        meta.Tools,
		Dependencies: meta.Dependencies,
		FilePath:     filePath,
	}

	// Parse tags: ["role:planning", "complexity:light"]
	for _, tag := range meta.Tags {
		parts := strings.SplitN(tag, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			switch key {
			case "role":
				result.Tags.Role = value
			case "complexity":
				result.Tags.Complexity = value
			}
		}
	}

	return result, nil
}

// processIncludes processes {{include "domains/..."}} patterns in command content
func processIncludes(content string, filesystem fs.FS) string {
	includePattern := regexp.MustCompile(`{{include "domains/([^"]+)"}}`)

	matches := includePattern.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return content
	}

	result := content
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		fullMatch := match[0]
		domainFile := match[1]

		domainPath := fmt.Sprintf("plugins/go-ent/commands/domains/%s", domainFile)
		domainContent, err := fs.ReadFile(filesystem, domainPath)
		if err != nil {
			continue
		}

		result = strings.ReplaceAll(result, fullMatch, string(domainContent))
	}

	return result
}
