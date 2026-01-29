package genconfig

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents ent.yaml configuration
type Config struct {
	Tools    []string       `yaml:"tools"`
	Models   ModelAliases   `yaml:"models"`
	OpenSpec OpenSpecConfig `yaml:"openspec,omitempty"`
}

// ModelAliases maps alias names to tool-specific model IDs
type ModelAliases struct {
	Fast  ToolModels `yaml:"fast"`
	Main  ToolModels `yaml:"main"`
	Heavy ToolModels `yaml:"heavy"`
}

// ToolModels maps tool name to model ID
type ToolModels struct {
	Claude   string `yaml:"claude"`
	OpenCode string `yaml:"opencode"`
}

// OpenSpecConfig holds OpenSpec customization
type OpenSpecConfig struct {
	Schema  string `yaml:"schema"`
	Context string `yaml:"context,omitempty"`
}

// Load loads config from ent.yaml
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &cfg, nil
}

// Default returns default configuration
func Default() *Config {
	return &Config{
		Tools: []string{"claude"},
		Models: ModelAliases{
			Fast: ToolModels{
				Claude:   "haiku",
				OpenCode: "zai-coding-plan/glm-4.5-air",
			},
			Main: ToolModels{
				Claude:   "sonnet",
				OpenCode: "zai-coding-plan/glm-4.7",
			},
			Heavy: ToolModels{
				Claude:   "opus",
				OpenCode: "kimi-for-coding/k2p5",
			},
		},
		OpenSpec: OpenSpecConfig{
			Schema: "spec-driven",
		},
	}
}

// Save writes config to file
func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}
