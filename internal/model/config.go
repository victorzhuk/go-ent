package model

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Version  string             `yaml:"version"`
	Runtimes map[string]Mapping `yaml:"runtimes"`
	Aliases  map[string]string  `yaml:"aliases"`
}

type Mapping struct {
	Fast  string `yaml:"fast"`
	Main  string `yaml:"main"`
	Heavy string `yaml:"heavy"`
}

func (m Mapping) Get(cat Category) string {
	switch cat {
	case Fast:
		return m.Fast
	case Main:
		return m.Main
	case Heavy:
		return m.Heavy
	default:
		return m.Main
	}
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return &cfg, nil
}

func LoadGlobal() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("get home dir: %w", err)
	}
	return Load(filepath.Join(home, ".go-ent", "models.yaml"))
}

func LoadProject(projectPath string) (*Config, error) {
	return Load(filepath.Join(projectPath, ".go-ent", "models.yaml"))
}

func Merge(global, project *Config) *Config {
	if global == nil && project == nil {
		return DefaultConfig()
	}
	if global == nil {
		return project
	}
	if project == nil {
		return global
	}

	merged := &Config{
		Version:  project.Version,
		Runtimes: make(map[string]Mapping),
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

func Save(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

func SaveGlobal(cfg *Config) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home dir: %w", err)
	}
	return Save(filepath.Join(home, ".go-ent", "models.yaml"), cfg)
}

func SaveProject(projectPath string, cfg *Config) error {
	return Save(filepath.Join(projectPath, ".go-ent", "models.yaml"), cfg)
}
