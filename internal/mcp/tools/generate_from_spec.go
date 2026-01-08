package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/generation"
	"github.com/victorzhuk/go-ent/internal/template"
	"github.com/victorzhuk/go-ent/internal/templates"
)

type GenerateFromSpecInput struct {
	SpecPath    string `json:"spec_path"`
	OutputDir   string `json:"output_dir"`
	ModulePath  string `json:"module_path"`
	ProjectName string `json:"project_name,omitempty"`
	ProjectRoot string `json:"project_root,omitempty"`
	GoVersion   string `json:"go_version,omitempty"`
}

func registerGenerateFromSpec(s *mcp.Server) {
	tool := &mcp.Tool{
		Name:        "go_ent_generate_from_spec",
		Description: "Generate a complete project from a spec file. Analyzes spec, selects archetype, generates scaffold, and provides AI prompts for business logic.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"spec_path": map[string]any{
					"type":        "string",
					"description": "Path to the spec file (relative or absolute)",
				},
				"output_dir": map[string]any{
					"type":        "string",
					"description": "Output directory for generated project",
				},
				"module_path": map[string]any{
					"type":        "string",
					"description": "Go module path (e.g., 'github.com/user/project')",
				},
				"project_name": map[string]any{
					"type":        "string",
					"description": "Project name (defaults to last segment of module_path)",
				},
				"project_root": map[string]any{
					"type":        "string",
					"description": "Project root directory for config (defaults to current directory)",
				},
				"go_version": map[string]any{
					"type":        "string",
					"description": "Go version (e.g., '1.25'). Defaults to current runtime version",
				},
			},
			"required": []string{"spec_path", "output_dir", "module_path"},
		},
	}

	mcp.AddTool(s, tool, generateFromSpecHandler)
}

func generateFromSpecHandler(ctx context.Context, req *mcp.CallToolRequest, input GenerateFromSpecInput) (*mcp.CallToolResult, any, error) {
	if input.SpecPath == "" {
		return nil, nil, fmt.Errorf("spec_path is required")
	}
	if input.OutputDir == "" {
		return nil, nil, fmt.Errorf("output_dir is required")
	}
	if input.ModulePath == "" {
		return nil, nil, fmt.Errorf("module_path is required")
	}

	// Defaults
	projectRoot := input.ProjectRoot
	if projectRoot == "" {
		projectRoot = "."
	}

	if input.ProjectName == "" {
		parts := strings.Split(input.ModulePath, "/")
		input.ProjectName = parts[len(parts)-1]
	}

	if input.GoVersion == "" {
		input.GoVersion = strings.TrimPrefix(runtime.Version(), "go")
		parts := strings.Split(input.GoVersion, ".")
		if len(parts) >= 2 {
			input.GoVersion = parts[0] + "." + parts[1]
		}
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

	// Enrich with archetype selection
	if err := generation.EnrichAnalysisWithArchetype(analysis, cfg, ""); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{
				Text: fmt.Sprintf("Error selecting archetype: %v", err),
			}},
		}, nil, nil
	}

	// Validate output directory
	if err := validateTargetDir(input.OutputDir); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
		}, nil, nil
	}

	// Validate archetype exists
	if _, err := generation.GetArchetype(analysis.Archetype, cfg); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{
				Text: fmt.Sprintf("Error getting archetype: %v", err),
			}},
		}, nil, nil
	}

	// Generate project scaffold
	engine := template.NewEngine(templates.TemplateFS)
	vars := template.TemplateVars{
		ModulePath:  input.ModulePath,
		ProjectName: input.ProjectName,
		GoVersion:   input.GoVersion,
	}

	// Determine template directory
	templateDir := "."
	if analysis.Archetype == "mcp" {
		templateDir = "mcp"
	}

	// Process templates
	if err := engine.ProcessAll(templateDir, vars, input.OutputDir); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{
				Text: fmt.Sprintf("Error generating project: %v", err),
			}},
		}, nil, nil
	}

	// Get generated files
	files, err := listGeneratedFiles(input.OutputDir)
	if err != nil {
		files = []string{"(unable to list files)"}
	}

	// Format analysis result
	analysisJSON, _ := json.MarshalIndent(analysis, "", "  ")

	msg := fmt.Sprintf("✅ Generated project from spec at %s\n\n", input.OutputDir)
	msg += fmt.Sprintf("**Spec:** %s\n", input.SpecPath)
	msg += fmt.Sprintf("**Module:** %s\n", input.ModulePath)
	msg += fmt.Sprintf("**Archetype:** %s (confidence: %.2f)\n\n", analysis.Archetype, analysis.Confidence)

	msg += "## Spec Analysis\n\n"
	msg += "```json\n" + string(analysisJSON) + "\n```\n\n"

	msg += "## Generated Files\n\n"
	for _, f := range files {
		msg += fmt.Sprintf("  - %s\n", f)
	}

	msg += "\n## Extension Points\n\n"
	msg += "The generated code contains `@generate:` markers for AI completion:\n"
	msg += "- `@generate:constructor` - Dependency injection\n"
	msg += "- `@generate:methods` - Business logic methods\n"
	msg += "- `@generate:handlers` - API endpoint handlers\n\n"

	msg += "## Next Steps\n\n"
	msg += fmt.Sprintf("  cd %s\n", input.OutputDir)
	msg += "  go mod tidy\n"
	msg += "  # Review extension points and use AI to generate business logic\n"
	msg += "  make build\n\n"

	msg += "## AI Prompt Templates Available\n\n"
	msg += "Use these prompts to fill extension points:\n"
	msg += "- `usecase.md` - Use case implementations\n"
	msg += "- `handler.md` - API handlers\n"
	msg += "- `repository.md` - Repository implementations\n"

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
	}, nil, nil
}
