package toolinit

import (
	"bytes"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

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

// ParseAgentFile parses an agent markdown file
func ParseAgentFile(content, filePath string) (*AgentMeta, error) {
	metadata, body, err := ParseFrontmatter(content)
	if err != nil {
		return nil, err
	}

	meta := &AgentMeta{
		Body:     body,
		FilePath: filePath,
	}

	if name, ok := metadata["name"].(string); ok {
		meta.Name = name
	}
	if desc, ok := metadata["description"].(string); ok {
		meta.Description = desc
	}
	if model, ok := metadata["model"].(string); ok {
		meta.Model = model
	}
	if color, ok := metadata["color"].(string); ok {
		meta.Color = color
	}

	if skills, ok := metadata["skills"].([]interface{}); ok {
		for _, s := range skills {
			if str, ok := s.(string); ok {
				meta.Skills = append(meta.Skills, str)
			}
		}
	}

	if tools, ok := metadata["tools"].([]interface{}); ok {
		for _, t := range tools {
			if str, ok := t.(string); ok {
				meta.Tools = append(meta.Tools, str)
			}
		}
	}

	// Parse tags array format: ["role:planning", "complexity:heavy"]
	if tags, ok := metadata["tags"].([]interface{}); ok {
		for _, tag := range tags {
			if tagStr, ok := tag.(string); ok {
				parts := strings.SplitN(tagStr, ":", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])
					switch key {
					case "role":
						meta.Tags.Role = value
					case "complexity":
						meta.Tags.Complexity = value
					}
				}
			}
		}
	}

	return meta, nil
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
