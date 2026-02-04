package genspec

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// findSchemaPath looks for the platformspecs directory starting from current dir
func findSchemaPath(tool string) (string, error) {
	// Try current directory first
	path := fmt.Sprintf("platformspecs/%s.yaml", tool)
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}

	// Try from ../../platformspecs (for tests running from internal/genspec)
	path = fmt.Sprintf("../../platformspecs/%s.yaml", tool)
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}

	return "", fmt.Errorf("schema not found: %s.yaml", tool)
}

// LoadToolSpec loads the spec for a given tool (claude, opencode) from the platformspecs/ directory
func LoadToolSpec(tool string) (*ToolSpec, error) {
	path, err := findSchemaPath(tool)
	if err != nil {
		return nil, err
	}

	// #nosec G304 - path is validated by findSchemaPath
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read schema: %w", err)
	}

	var spec ToolSpec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("parse schema: %w", err)
	}

	return &spec, nil
}
