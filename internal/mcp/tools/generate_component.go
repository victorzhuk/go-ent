package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/generation"
)

type GenerateComponentInput struct {
	SpecPath      string `json:"spec_path"`
	ComponentName string `json:"component_name,omitempty"`
	OutputDir     string `json:"output_dir,omitempty"`
	ProjectRoot   string `json:"project_root,omitempty"`
}

func registerGenerateComponent(s *mcp.Server) {
	tool := &mcp.Tool{
		Name:        "generate_component",
		Description: "Generate a component scaffold from a spec file. Analyzes the spec, selects templates, and marks extension points for AI completion.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"spec_path": map[string]any{
					"type":        "string",
					"description": "Path to the spec file (relative or absolute)",
				},
				"component_name": map[string]any{
					"type":        "string",
					"description": "Component name (defaults to spec file name)",
				},
				"output_dir": map[string]any{
					"type":        "string",
					"description": "Output directory for generated code (defaults to current directory)",
				},
				"project_root": map[string]any{
					"type":        "string",
					"description": "Project root directory (defaults to current directory)",
				},
			},
			"required": []string{"spec_path"},
		},
	}

	mcp.AddTool(s, tool, generateComponentHandler)
}

func generateComponentHandler(ctx context.Context, req *mcp.CallToolRequest, input GenerateComponentInput) (*mcp.CallToolResult, any, error) {
	if input.SpecPath == "" {
		return nil, nil, fmt.Errorf("spec_path is required")
	}

	// Defaults
	projectRoot := input.ProjectRoot
	if projectRoot == "" {
		projectRoot = "."
	}
	outputDir := input.OutputDir
	if outputDir == "" {
		outputDir = "."
	}

	// Resolve spec path
	specPath := input.SpecPath
	if !filepath.IsAbs(specPath) {
		specPath = filepath.Join(projectRoot, specPath)
	}

	// Check if spec exists
	if _, err := os.Stat(specPath); os.IsNotExist(err) {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{
				Text: fmt.Sprintf("Spec file not found: %s", specPath),
			}},
		}, nil, nil
	}

	// Load config
	cfg, err := generation.LoadConfig(projectRoot)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{
				Text: fmt.Sprintf("Error loading config: %v", err),
			}},
		}, nil, nil
	}

	// Analyze spec
	analysis, err := generation.AnalyzeSpec(specPath)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{
				Text: fmt.Sprintf("Error analyzing spec: %v", err),
			}},
		}, nil, nil
	}

	// Enrich analysis with archetype selection
	if err := generation.EnrichAnalysisWithArchetype(analysis, cfg, ""); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{
				Text: fmt.Sprintf("Error selecting archetype: %v", err),
			}},
		}, nil, nil
	}

	// Get component name
	componentName := input.ComponentName
	if componentName == "" {
		base := filepath.Base(specPath)
		componentName = base[:len(base)-len(filepath.Ext(base))]
	}

	// Format analysis result
	analysisJSON, err := json.MarshalIndent(analysis, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{
				Text: fmt.Sprintf("Error formatting analysis: %v", err),
			}},
		}, nil, nil
	}

	msg := "# Component Generation Analysis\n\n"
	msg += fmt.Sprintf("**Spec:** %s\n", input.SpecPath)
	msg += fmt.Sprintf("**Component:** %s\n", componentName)
	msg += fmt.Sprintf("**Recommended Archetype:** %s (confidence: %.2f)\n\n", analysis.Archetype, analysis.Confidence)

	msg += "## Spec Analysis\n\n"
	msg += "```json\n" + string(analysisJSON) + "\n```\n\n"

	msg += "## Next Steps\n\n"
	msg += "1. Review the analysis and adjust archetype if needed in `generation.yaml`\n"
	msg += "2. Use `generate` to scaffold the project with the selected archetype\n"
	msg += "3. Extension points will be marked with `@generate:` comments\n"
	msg += "4. Use AI prompts to fill in business logic at extension points\n\n"

	msg += "## Example generation.yaml Override\n\n"
	msg += "```yaml\n"
	msg += "components:\n"
	msg += fmt.Sprintf("  - name: %s\n", componentName)
	msg += fmt.Sprintf("    spec: %s\n", input.SpecPath)
	msg += fmt.Sprintf("    archetype: %s  # Change if needed\n", analysis.Archetype)
	msg += fmt.Sprintf("    output: %s\n", outputDir)
	msg += "```\n"

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
	}, nil, nil
}
