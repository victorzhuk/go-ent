package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/template"
	"github.com/victorzhuk/go-ent/internal/templates"
)

type GenerateInput struct {
	Path        string `json:"path"`
	ModulePath  string `json:"module_path"`
	ProjectName string `json:"project_name,omitempty"`
	ProjectType string `json:"project_type,omitempty"`
	GoVersion   string `json:"go_version,omitempty"`
}

func registerGenerate(s *mcp.Server) {
	tool := &mcp.Tool{
		Name:        "go_ent_generate",
		Description: "Generate a new Go project from templates. Supports 'standard' (web service) and 'mcp' (MCP server) project types.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "Target directory path for the new project",
				},
				"module_path": map[string]any{
					"type":        "string",
					"description": "Go module path (e.g., 'github.com/user/project')",
				},
				"project_name": map[string]any{
					"type":        "string",
					"description": "Project name (defaults to last segment of module_path if omitted)",
				},
				"project_type": map[string]any{
					"type":        "string",
					"description": "Project type: 'standard' (web service) or 'mcp' (MCP server)",
					"enum":        []string{"standard", "mcp"},
					"default":     "standard",
				},
				"go_version": map[string]any{
					"type":        "string",
					"description": "Go version (e.g., '1.24'). Defaults to current runtime version if omitted",
				},
			},
			"required": []string{"path", "module_path"},
		},
	}

	mcp.AddTool(s, tool, generateHandler)
}

func generateHandler(ctx context.Context, req *mcp.CallToolRequest, input GenerateInput) (*mcp.CallToolResult, any, error) {
	if input.Path == "" {
		return nil, nil, fmt.Errorf("path is required")
	}
	if input.ModulePath == "" {
		return nil, nil, fmt.Errorf("module_path is required")
	}

	// Default project name from module path
	if input.ProjectName == "" {
		parts := strings.Split(input.ModulePath, "/")
		input.ProjectName = parts[len(parts)-1]
	}

	// Default project type
	if input.ProjectType == "" {
		input.ProjectType = "standard"
	}

	// Default Go version from runtime
	if input.GoVersion == "" {
		input.GoVersion = strings.TrimPrefix(runtime.Version(), "go")
		// Keep only major.minor
		parts := strings.Split(input.GoVersion, ".")
		if len(parts) >= 2 {
			input.GoVersion = parts[0] + "." + parts[1]
		}
	}

	// Validate project type
	if input.ProjectType != "standard" && input.ProjectType != "mcp" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{
				Text: fmt.Sprintf("Invalid project_type: %s. Must be 'standard' or 'mcp'", input.ProjectType),
			}},
		}, nil, nil
	}

	// Check if target directory exists and is not empty
	if err := validateTargetDir(input.Path); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}},
		}, nil, nil
	}

	// Create template engine
	engine := template.NewEngine(templates.TemplateFS)

	// Prepare template variables
	vars := template.TemplateVars{
		ModulePath:  input.ModulePath,
		ProjectName: input.ProjectName,
		GoVersion:   input.GoVersion,
	}

	// Determine template directory based on project type
	templateDir := "."
	if input.ProjectType == "mcp" {
		templateDir = "mcp"
	}

	// Process all templates
	if err := engine.ProcessAll(templateDir, vars, input.Path); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{
				Text: fmt.Sprintf("Error generating project: %v", err),
			}},
		}, nil, nil
	}

	// Get list of generated files
	files, err := listGeneratedFiles(input.Path)
	if err != nil {
		files = []string{"(unable to list files)"}
	}

	msg := fmt.Sprintf("✅ Generated %s project at %s\n\n", input.ProjectType, input.Path)
	msg += fmt.Sprintf("Module: %s\n", input.ModulePath)
	msg += fmt.Sprintf("Name: %s\n", input.ProjectName)
	msg += fmt.Sprintf("Go Version: %s\n\n", input.GoVersion)
	msg += "Generated files:\n"
	for _, f := range files {
		msg += fmt.Sprintf("  - %s\n", f)
	}
	msg += "\nNext steps:\n"
	msg += fmt.Sprintf("  cd %s\n", input.Path)
	msg += "  go mod tidy\n"
	msg += "  make build\n"

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
	}, nil, nil
}

func validateTargetDir(path string) error {
	info, err := os.Stat(path)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("path exists and is not a directory: %s", path)
		}
		// Check if directory is empty
		entries, err := os.ReadDir(path)
		if err != nil {
			return fmt.Errorf("cannot read directory: %w", err)
		}
		if len(entries) > 0 {
			return fmt.Errorf("directory is not empty: %s", path)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("cannot access path: %w", err)
	}
	return nil
}

func listGeneratedFiles(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			files = append(files, rel)
		}
		return nil
	})
	return files, err
}
