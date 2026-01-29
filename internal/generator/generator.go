package generator

import (
	"fmt"
	"os"
	"path/filepath"
)

// Generator orchestrates agent generation for multiple targets
type Generator struct {
	SrcDir  string
	Targets []Target
}

// New creates a new Generator
func New(srcDir string, targets ...Target) *Generator {
	return &Generator{
		SrcDir:  srcDir,
		Targets: targets,
	}
}

// GenerateAll generates all agents for all targets
func (g *Generator) GenerateAll() error {
	agents, err := ListAgents(g.SrcDir)
	if err != nil {
		return fmt.Errorf("list agents: %w", err)
	}

	for _, agentName := range agents {
		if err := g.GenerateAgent(agentName); err != nil {
			return fmt.Errorf("generate agent %s: %w", agentName, err)
		}
	}

	return nil
}

// GenerateAgent generates a single agent for all targets
func (g *Generator) GenerateAgent(name string) error {
	// Load source
	agent, prompts, err := LoadAgentSource(g.SrcDir, name)
	if err != nil {
		return fmt.Errorf("load source: %w", err)
	}

	// Generate for each target
	for _, target := range g.Targets {
		output, err := target.Generate(agent, prompts)
		if err != nil {
			return fmt.Errorf("generate %s target: %w", target.Name(), err)
		}

		outputPath := target.OutputPath(name)
		if err := g.writeOutput(outputPath, output); err != nil {
			return fmt.Errorf("write %s output: %w", target.Name(), err)
		}

		fmt.Printf("Generated %s → %s\n", name, outputPath)
	}

	return nil
}

// writeOutput writes output to file, creating directories as needed
func (g *Generator) writeOutput(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}
