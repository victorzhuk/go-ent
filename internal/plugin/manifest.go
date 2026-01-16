package plugin

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	ManifestFile = "plugin.yaml"
	PluginDir    = "plugins"
)

type Manifest struct {
	Name        string     `yaml:"name" json:"name"`
	Version     string     `yaml:"version" json:"version"`
	Description string     `yaml:"description" json:"description"`
	Author      string     `yaml:"author" json:"author"`
	Skills      []SkillRef `yaml:"skills,omitempty" json:"skills,omitempty"`
	Agents      []AgentRef `yaml:"agents,omitempty" json:"agents,omitempty"`
	Rules       []RuleRef  `yaml:"rules,omitempty" json:"rules,omitempty"`
	MinVersion  string     `yaml:"min_version,omitempty" json:"min_version,omitempty"`
}

type SkillRef struct {
	Name string `yaml:"name" json:"name"`
	Path string `yaml:"path" json:"path"`
}

type AgentRef struct {
	Name string `yaml:"name" json:"name"`
	Path string `yaml:"path" json:"path"`
}

type RuleRef struct {
	Name string `yaml:"name" json:"name"`
	Path string `yaml:"path" json:"path"`
}

type Plugin struct {
	Manifest  Manifest
	RootPath  string
	Enabled   bool
	Installed bool
}

func ParseManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- controlled config/template file path
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}

	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}

	if err := m.Validate(); err != nil {
		return nil, fmt.Errorf("validate manifest: %w", err)
	}

	return &m, nil
}

func (m *Manifest) Validate() error {
	if m.Name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	if len(m.Name) > 100 {
		return fmt.Errorf("name too long (max 100 characters)")
	}

	if m.Version == "" {
		return fmt.Errorf("version cannot be empty")
	}

	if m.Description == "" {
		return fmt.Errorf("description cannot be empty")
	}

	if len(m.Description) > 500 {
		return fmt.Errorf("description too long (max 500 characters)")
	}

	if m.Author == "" {
		return fmt.Errorf("author cannot be empty")
	}

	for i, ref := range m.Skills {
		if ref.Name == "" {
			return fmt.Errorf("skill[%d]: name cannot be empty", i)
		}
		if ref.Path == "" {
			return fmt.Errorf("skill[%d]: path cannot be empty", i)
		}
	}

	for i, ref := range m.Agents {
		if ref.Name == "" {
			return fmt.Errorf("agent[%d]: name cannot be empty", i)
		}
		if ref.Path == "" {
			return fmt.Errorf("agent[%d]: path cannot be empty", i)
		}
	}

	for i, ref := range m.Rules {
		if ref.Name == "" {
			return fmt.Errorf("rule[%d]: name cannot be empty", i)
		}
		if ref.Path == "" {
			return fmt.Errorf("rule[%d]: path cannot be empty", i)
		}
	}

	return nil
}

func (m *Manifest) GetSkillPath(skillName string) (string, error) {
	for _, ref := range m.Skills {
		if ref.Name == skillName {
			return ref.Path, nil
		}
	}
	return "", fmt.Errorf("skill not found: %s", skillName)
}

func (m *Manifest) GetAgentPath(agentName string) (string, error) {
	for _, ref := range m.Agents {
		if ref.Name == agentName {
			return ref.Path, nil
		}
	}
	return "", fmt.Errorf("agent not found: %s", agentName)
}

func (m *Manifest) GetRulePath(ruleName string) (string, error) {
	for _, ref := range m.Rules {
		if ref.Name == ruleName {
			return ref.Path, nil
		}
	}
	return "", fmt.Errorf("rule not found: %s", ruleName)
}

func (p *Plugin) ResolvePath(relativePath string) (string, error) {
	if p.RootPath == "" {
		return "", fmt.Errorf("plugin root path not set")
	}

	absPath := filepath.Join(p.RootPath, relativePath)
	if _, err := os.Stat(absPath); err != nil {
		return "", fmt.Errorf("resolve path %s: %w", relativePath, err)
	}

	return absPath, nil
}
