package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/victorzhuk/go-ent/internal/xdg"
	"gopkg.in/yaml.v3"
)

type ModelConfig struct {
	Version  string                  `yaml:"version"`
	Runtimes map[string]ModelMapping `yaml:"runtimes"`
	Aliases  map[string]string       `yaml:"aliases"`
}

type ModelMapping struct {
	Fast  string `yaml:"fast"`
	Main  string `yaml:"main"`
	Heavy string `yaml:"heavy"`
}

func (m ModelMapping) Get(cat ModelCategory) string {
	switch cat {
	case ModelFast:
		return m.Fast
	case ModelMain:
		return m.Main
	case ModelHeavy:
		return m.Heavy
	default:
		return m.Main
	}
}

func LoadModelConfig(path string) (*ModelConfig, error) {
	data, err := os.ReadFile(path) // #nosec G304
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg ModelConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return &cfg, nil
}

func LoadGlobalModelConfig() (*ModelConfig, error) {
	xdgPath := filepath.Join(xdg.ConfigDir(), "models.yaml")
	if cfg, err := LoadModelConfig(xdgPath); cfg != nil || err != nil {
		return cfg, err
	}

	legacyPath := filepath.Join(xdg.LegacyDir(), "models.yaml")
	return LoadModelConfig(legacyPath)
}

func LoadProjectModelConfig(projectPath string) (*ModelConfig, error) {
	return LoadModelConfig(filepath.Join(projectPath, ".go-ent", "models.yaml"))
}

func MergeModelConfigs(global, project *ModelConfig) *ModelConfig {
	if global == nil && project == nil {
		return DefaultModelConfig()
	}
	if global == nil {
		return project
	}
	if project == nil {
		return global
	}

	merged := &ModelConfig{
		Version:  project.Version,
		Runtimes: make(map[string]ModelMapping),
		Aliases:  make(map[string]string),
	}

	for k, v := range global.Runtimes {
		merged.Runtimes[k] = v
	}
	for k, v := range project.Runtimes {
		merged.Runtimes[k] = v
	}

	for k, v := range global.Aliases {
		merged.Aliases[k] = v
	}
	for k, v := range project.Aliases {
		merged.Aliases[k] = v
	}

	return merged
}

func SaveModelConfig(path string, cfg *ModelConfig) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

func SaveGlobalModelConfig(cfg *ModelConfig) error {
	return SaveModelConfig(filepath.Join(xdg.ConfigDir(), "models.yaml"), cfg)
}

func SaveProjectModelConfig(projectPath string, cfg *ModelConfig) error {
	return SaveModelConfig(filepath.Join(projectPath, ".go-ent", "models.yaml"), cfg)
}
