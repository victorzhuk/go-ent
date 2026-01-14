package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// yamlMeta represents the structure of agent metadata YAML files.
type yamlMeta struct {
	Name         string   `yaml:"name"`
	Description  string   `yaml:"description"`
	Model        string   `yaml:"model"`
	Color        string   `yaml:"color"`
	Skills       []string `yaml:"skills"`
	Tools        []string `yaml:"tools"`
	Dependencies []string `yaml:"dependencies,omitempty"`
}

// MetaLoader handles loading and parsing agent metadata YAML files.
type MetaLoader struct{}

// NewMetaLoader creates a new meta loader.
func NewMetaLoader() *MetaLoader {
	return &MetaLoader{}
}

// LoadMetaFiles scans a directory for *.yaml files and loads agent metadata.
func (l *MetaLoader) LoadMetaFiles(metaDir string) (map[string]AgentMeta, error) {
	entries, err := os.ReadDir(metaDir)
	if err != nil {
		return nil, fmt.Errorf("read dir: %w", err)
	}

	metas := make(map[string]AgentMeta)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			continue
		}

		path := filepath.Join(metaDir, name)
		meta, err := l.parseMetaFile(path)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", name, err)
		}

		metas[meta.Name] = *meta
	}

	return metas, nil
}

// parseMetaFile parses a single agent metadata YAML file.
func (l *MetaLoader) parseMetaFile(path string) (*AgentMeta, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var ym yamlMeta
	if err := yaml.Unmarshal(data, &ym); err != nil {
		return nil, fmt.Errorf("unmarshal yaml: %w", err)
	}

	if ym.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	tools := make(map[string]bool)
	for _, tool := range ym.Tools {
		tools[tool] = true
	}

	return &AgentMeta{
		Name:        ym.Name,
		Description: ym.Description,
		Model:       ym.Model,
		Color:       ym.Color,
		Skills:      ym.Skills,
		Tools:       tools,
		FilePath:    path,
	}, nil
}

// BuildDependencyGraph builds a dependency graph from agent metadata.
func (l *MetaLoader) BuildDependencyGraph(metas map[string]AgentMeta) (*DependencyGraph, error) {
	graph := NewDependencyGraph()

	for name, meta := range metas {
		graph.AddNode(name, meta)
	}

	for name, meta := range metas {
		for _, depName := range meta.Dependencies {
			if !graph.HasNode(depName) {
				return nil, fmt.Errorf("dependency not found: %s depends on %s", name, depName)
			}
			graph.AddEdge(name, depName)
		}
	}

	return graph, nil
}
