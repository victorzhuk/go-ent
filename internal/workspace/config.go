package workspace

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type WorkspaceConfig struct {
	Version string            `yaml:"version"`
	Models  map[string]string `yaml:"models,omitempty"`
	Skills  SkillsOverride    `yaml:"skills,omitempty"`
}

type SkillsOverride struct {
	Enabled  []string `yaml:"enabled,omitempty"`
	Disabled []string `yaml:"disabled,omitempty"`
}

func DefaultWorkspaceConfig() *WorkspaceConfig {
	return &WorkspaceConfig{
		Version: "1.0",
	}
}

func LoadWorkspaceConfig(name string) (*WorkspaceConfig, error) {
	path := filepath.Join(workspaceDataDir(name), "config.yaml")

	data, err := os.ReadFile(path) // #nosec G304
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultWorkspaceConfig(), nil
		}
		return nil, fmt.Errorf("read workspace config: %w", err)
	}

	var cfg WorkspaceConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse workspace config: %w", err)
	}

	return &cfg, nil
}

func SaveWorkspaceConfig(name string, cfg *WorkspaceConfig) error {
	dir := workspaceDataDir(name)

	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("create workspace data dir: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal workspace config: %w", err)
	}

	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write workspace config: %w", err)
	}

	return nil
}
