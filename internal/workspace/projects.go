package workspace

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func LoadProjectsRegistry(name string) (*ProjectsRegistry, error) {
	path := filepath.Join(workspaceDataDir(name), "projects.yaml")

	data, err := os.ReadFile(path) // #nosec G304
	if err != nil {
		if os.IsNotExist(err) {
			return &ProjectsRegistry{}, nil
		}
		return nil, fmt.Errorf("read projects registry: %w", err)
	}

	var reg ProjectsRegistry
	if err := yaml.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("parse projects registry: %w", err)
	}

	return &reg, nil
}

func SaveProjectsRegistry(name string, reg *ProjectsRegistry) error {
	dir := workspaceDataDir(name)

	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("create workspace data dir: %w", err)
	}

	data, err := yaml.Marshal(reg)
	if err != nil {
		return fmt.Errorf("marshal projects registry: %w", err)
	}

	path := filepath.Join(dir, "projects.yaml")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write projects registry: %w", err)
	}

	return nil
}
