package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/victorzhuk/go-ent/internal/config"
)

type Generator struct {
	SrcDir  string
	Targets []Target
	Config  *config.ToolRuntimeConfig
}

func New(srcDir string, cfg *config.ToolRuntimeConfig, targets ...Target) *Generator {
	return &Generator{
		SrcDir:  srcDir,
		Targets: targets,
		Config:  cfg,
	}
}

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

func (g *Generator) GenerateAgent(name string) error {
	metaAgent, prompts, err := LoadAgentMetaSource(g.SrcDir, name)
	if err != nil {
		return fmt.Errorf("load source: %w", err)
	}

	for _, target := range g.Targets {
		agent := ConvertMetaToSource(metaAgent)

		if g.Config != nil {
			switch target.Runtime() {
			case "claude":
				agent.Model.Claude = g.Config.Claude.Resolve(agent.Model.Claude)
			case "opencode":
				agent.Model.OpenCode = g.Config.OpenCode.Resolve(agent.Model.OpenCode)
			}
		}

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

func (g *Generator) writeOutput(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

func wrapWithFrontmatter(frontmatter []byte, content string) []byte {
	var sb strings.Builder
	sb.Grow(len(frontmatter) + len(content) + 8)
	sb.WriteString("---\n")
	sb.Write(frontmatter)
	sb.WriteString("---\n\n")
	sb.WriteString(content)
	return []byte(sb.String())
}
