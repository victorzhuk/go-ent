package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type ModelTiers struct {
	Fast   string            `yaml:"fast"`
	Main   string            `yaml:"main"`
	Heavy  string            `yaml:"heavy"`
	Agents map[string]string `yaml:"agents,omitempty"`
}

type ToolRuntimeConfig struct {
	Claude   ModelTiers `yaml:"claude"`
	OpenCode ModelTiers `yaml:"opencode"`
}

func (m *ModelTiers) Resolve(tier string) string {
	switch tier {
	case "fast":
		return m.Fast
	case "main":
		return m.Main
	case "heavy":
		return m.Heavy
	default:
		return tier
	}
}

func (m *ModelTiers) ResolveForAgent(agentName, defaultTier string) string {
	if m.Agents != nil {
		if override, ok := m.Agents[agentName]; ok {
			return m.Resolve(override)
		}
	}
	return m.Resolve(defaultTier)
}

func LoadToolRuntimeConfig(projectDir, runtime string) (*ToolRuntimeConfig, error) {
	configPath, err := RuntimeConfigPath(runtime, projectDir)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config not found at %s: run 'ent config init --runtime=%s': %w", configPath, runtime, os.ErrNotExist)
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg ToolRuntimeConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return &cfg, nil
}

func ValidateForRuntime(cfg *ToolRuntimeConfig, runtime string) error {
	var missing []string
	switch runtime {
	case "claude":
		if cfg.Claude.Fast == "" {
			missing = append(missing, "claude.fast")
		}
		if cfg.Claude.Main == "" {
			missing = append(missing, "claude.main")
		}
		if cfg.Claude.Heavy == "" {
			missing = append(missing, "claude.heavy")
		}
	case "opencode":
		if cfg.OpenCode.Fast == "" {
			missing = append(missing, "opencode.fast")
		}
		if cfg.OpenCode.Main == "" {
			missing = append(missing, "opencode.main")
		}
		if cfg.OpenCode.Heavy == "" {
			missing = append(missing, "opencode.heavy")
		}
	default:
		return fmt.Errorf("unknown runtime: %s", runtime)
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required config values: %s", strings.Join(missing, ", "))
	}
	return nil
}

func ApplyKey(cfg *ToolRuntimeConfig, key, value string) error {
	parts := strings.SplitN(key, ".", 3)
	switch {
	case len(parts) == 2 && parts[0] == "claude":
		switch parts[1] {
		case "fast":
			cfg.Claude.Fast = value
		case "main":
			cfg.Claude.Main = value
		case "heavy":
			cfg.Claude.Heavy = value
		default:
			return fmt.Errorf("unknown config key: %s (use claude.fast, claude.main, claude.heavy, claude.agents.<name>)", key)
		}
	case len(parts) == 3 && parts[0] == "claude" && parts[1] == "agents":
		if !isValidTier(value) {
			return fmt.Errorf("invalid tier %q: use fast, main, or heavy", value)
		}
		if cfg.Claude.Agents == nil {
			cfg.Claude.Agents = make(map[string]string)
		}
		cfg.Claude.Agents[parts[2]] = value
	case len(parts) == 2 && parts[0] == "opencode":
		switch parts[1] {
		case "fast":
			cfg.OpenCode.Fast = value
		case "main":
			cfg.OpenCode.Main = value
		case "heavy":
			cfg.OpenCode.Heavy = value
		default:
			return fmt.Errorf("unknown config key: %s (use opencode.fast, opencode.main, opencode.heavy, opencode.agents.<name>)", key)
		}
	case len(parts) == 3 && parts[0] == "opencode" && parts[1] == "agents":
		if !isValidTier(value) {
			return fmt.Errorf("invalid tier %q: use fast, main, or heavy", value)
		}
		if cfg.OpenCode.Agents == nil {
			cfg.OpenCode.Agents = make(map[string]string)
		}
		cfg.OpenCode.Agents[parts[2]] = value
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}
	return nil
}

func RuntimeConfigPath(runtime, projectRoot string) (string, error) {
	switch runtime {
	case "claude":
		return filepath.Join(projectRoot, ".claude", "ent.yaml"), nil
	case "opencode":
		return filepath.Join(projectRoot, ".opencode", "ent.yaml"), nil
	default:
		return "", fmt.Errorf("unknown runtime: %s", runtime)
	}
}

func LoadCombinedRuntimeConfig(projectDir string, tools []string) (*ToolRuntimeConfig, error) {
	var combined ToolRuntimeConfig
	for _, tool := range tools {
		toolCfg, err := LoadToolRuntimeConfig(projectDir, tool)
		if err != nil {
			return nil, fmt.Errorf("load %s config: %w", tool, err)
		}
		switch tool {
		case "claude":
			combined.Claude = toolCfg.Claude
		case "opencode":
			combined.OpenCode = toolCfg.OpenCode
		}
	}
	return &combined, nil
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

func isValidTier(s string) bool {
	return s == "fast" || s == "main" || s == "heavy"
}
