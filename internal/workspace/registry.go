package workspace

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func LoadRegistry() (*WorkspaceRegistry, error) {
	path := registryPath()

	data, err := os.ReadFile(path) // #nosec G304
	if err != nil {
		if os.IsNotExist(err) {
			return &WorkspaceRegistry{
				Workspaces: make(map[string]string),
			}, nil
		}
		return nil, fmt.Errorf("read workspace registry: %w", err)
	}

	var reg WorkspaceRegistry
	if err := yaml.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("parse workspace registry: %w", err)
	}

	if reg.Workspaces == nil {
		reg.Workspaces = make(map[string]string)
	}

	return &reg, nil
}

func SaveRegistry(reg *WorkspaceRegistry) error {
	path := registryPath()

	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("create registry dir: %w", err)
	}

	data, err := yaml.Marshal(reg)
	if err != nil {
		return fmt.Errorf("marshal workspace registry: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write workspace registry: %w", err)
	}

	return nil
}
