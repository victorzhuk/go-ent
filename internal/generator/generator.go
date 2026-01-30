package generator

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/victorzhuk/go-ent/internal/genconfig"
)

// Generator orchestrates agent generation for multiple targets
type Generator struct {
	SrcDir  string
	Targets []Target
	Config  *genconfig.Config
}

// New creates a new Generator
func New(srcDir string, cfg *genconfig.Config, targets ...Target) *Generator {
	return &Generator{
		SrcDir:  srcDir,
		Targets: targets,
		Config:  cfg,
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
	// Load from meta format
	metaAgent, prompts, err := LoadAgentMetaSource(g.SrcDir, name)
	if err != nil {
		return fmt.Errorf("load source: %w", err)
	}

	// Convert to old format for generation
	agent := ConvertMetaToSource(metaAgent)

	// Resolve model aliases
	if g.Config != nil {
		agent.Model.Claude = g.ResolveModel(agent.Model.Claude, "claude")
		agent.Model.OpenCode = g.ResolveModel(agent.Model.OpenCode, "opencode")
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

// ResolveModel resolves a model alias to a tool-specific model ID
func (g *Generator) ResolveModel(alias, tool string) string {
	if g.Config == nil {
		return ""
	}

	var models genconfig.ToolModels
	switch alias {
	case "fast":
		models = g.Config.Models.Fast
	case "main":
		models = g.Config.Models.Main
	case "heavy":
		models = g.Config.Models.Heavy
	default:
		return alias
	}

	switch tool {
	case "claude":
		return models.Claude
	case "opencode":
		return models.OpenCode
	default:
		return ""
	}
}
