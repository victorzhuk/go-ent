package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type ToolRuntimeConfig struct {
	Claude   ClaudeModels   `yaml:"claude"`
	OpenCode OpenCodeModels `yaml:"opencode"`
}

type ClaudeModels struct {
	Sonnet string `yaml:"sonnet"`
	Opus   string `yaml:"opus"`
	Haiku  string `yaml:"haiku"`
}

type OpenCodeModels struct {
	Fast  string `yaml:"fast"`
	Main  string `yaml:"main"`
	Heavy string `yaml:"heavy"`
}

func LoadToolRuntimeConfig(projectDir, runtime string) (*ToolRuntimeConfig, error) {
	var configPath string
	switch runtime {
	case "claude":
		configPath = filepath.Join(projectDir, ".claude", "ent.yaml")
	case "opencode":
		configPath = filepath.Join(projectDir, ".opencode", "ent.yaml")
	default:
		return nil, fmt.Errorf("unknown runtime: %s", runtime)
	}

	cfg := DefaultToolRuntimeConfig()

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return cfg, nil
}

func DefaultToolRuntimeConfig() *ToolRuntimeConfig {
	return &ToolRuntimeConfig{
		Claude: ClaudeModels{
			Sonnet: "claude-sonnet-4-5-20250929",
			Opus:   "claude-opus-4-5-20251101",
			Haiku:  "claude-haiku-4-5-20251001",
		},
		OpenCode: OpenCodeModels{
			Fast:  "zai-coding-plan/glm-4.7-flash",
			Main:  "zai-coding-plan/glm-5",
			Heavy: "kimi-for-coding/k2p5",
		},
	}
}

// CategoryToAlias maps a model category (fast/main/heavy) to a Claude model alias (haiku/sonnet/opus).
func CategoryToAlias(category string) string {
	switch category {
	case "fast":
		return "haiku"
	case "main":
		return "sonnet"
	case "heavy":
		return "opus"
	default:
		return category
	}
}

func (c *ClaudeModels) Resolve(alias string) string {
	switch alias {
	case "fast", "haiku":
		if c.Haiku != "" {
			return c.Haiku
		}
		return "claude-haiku-4-5-20251001"
	case "main", "sonnet":
		if c.Sonnet != "" {
			return c.Sonnet
		}
		return "claude-sonnet-4-5-20250929"
	case "heavy", "opus":
		if c.Opus != "" {
			return c.Opus
		}
		return "claude-opus-4-5-20251101"
	default:
		return alias
	}
}

func (c *OpenCodeModels) Resolve(alias string) string {
	switch alias {
	case "fast", "haiku":
		if c.Fast != "" {
			return c.Fast
		}
		return "zai-coding-plan/glm-4.7-flash"
	case "main", "sonnet":
		if c.Main != "" {
			return c.Main
		}
		return "zai-coding-plan/glm-5"
	case "heavy", "opus":
		if c.Heavy != "" {
			return c.Heavy
		}
		return "kimi-for-coding/k2p5"
	default:
		return alias
	}
}

func DetectRuntime(projectDir string) string {
	claudeConfig := filepath.Join(projectDir, ".claude", "ent.yaml")
	if fileExists(claudeConfig) {
		return "claude"
	}

	opencodeConfig := filepath.Join(projectDir, ".opencode", "ent.yaml")
	if fileExists(opencodeConfig) {
		return "opencode"
	}

	if dirExists(filepath.Join(projectDir, ".claude")) {
		return "claude"
	}
	if dirExists(filepath.Join(projectDir, ".opencode")) {
		return "opencode"
	}

	return ""
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func LoadCombinedRuntimeConfig(projectDir string, tools []string) *ToolRuntimeConfig {
	cfg := DefaultToolRuntimeConfig()

	for _, tool := range tools {
		toolCfg, err := LoadToolRuntimeConfig(projectDir, tool)
		if err != nil {
			continue
		}
		switch tool {
		case "claude":
			cfg.Claude = toolCfg.Claude
		case "opencode":
			cfg.OpenCode = toolCfg.OpenCode
		}
	}

	return cfg
}
