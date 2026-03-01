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

type RuntimeConfig struct {
	Models ModelTiers `yaml:"models"`
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

func LoadRuntimeConfig(projectDir, runtime string) (*RuntimeConfig, error) {
	configPath, err := RuntimeConfigPath(runtime, projectDir)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath) // #nosec G304
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config not found at %s: run 'ent config init --runtime=%s': %w", configPath, runtime, os.ErrNotExist)
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg RuntimeConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return &cfg, nil
}

func (cfg *RuntimeConfig) Validate() error {
	var missing []string
	if cfg.Models.Fast == "" {
		missing = append(missing, "models.fast")
	}
	if cfg.Models.Main == "" {
		missing = append(missing, "models.main")
	}
	if cfg.Models.Heavy == "" {
		missing = append(missing, "models.heavy")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required config values: %s", strings.Join(missing, ", "))
	}
	return nil
}

func ApplyKey(cfg *RuntimeConfig, key, value string) error {
	parts := strings.SplitN(key, ".", 3)
	switch {
	case len(parts) == 2 && parts[0] == "models":
		switch parts[1] {
		case "fast":
			cfg.Models.Fast = value
		case "main":
			cfg.Models.Main = value
		case "heavy":
			cfg.Models.Heavy = value
		default:
			return fmt.Errorf("unknown config key: %s (use models.fast, models.main, models.heavy, models.agents.<name>)", key)
		}
	case len(parts) == 3 && parts[0] == "models" && parts[1] == "agents":
		if !isValidTier(value) {
			return fmt.Errorf("invalid tier %q: use fast, main, or heavy", value)
		}
		if cfg.Models.Agents == nil {
			cfg.Models.Agents = make(map[string]string)
		}
		cfg.Models.Agents[parts[2]] = value
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
