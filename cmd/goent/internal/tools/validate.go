package tools

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/cmd/goent/internal/spec"
)

type ValidateInput struct {
	Path   string `json:"path"`
	Type   string `json:"type,omitempty"`
	ID     string `json:"id,omitempty"`
	Strict bool   `json:"strict,omitempty"`
}

func registerValidate(s *mcp.Server) {
	tool := &mcp.Tool{
		Name:        "goent_spec_validate",
		Description: "Validate OpenSpec files. Type can be 'spec', 'change', or 'all'. Use strict mode for comprehensive validation.",
	}

	mcp.AddTool(s, tool, validateHandler)
}

func validateHandler(ctx context.Context, req *mcp.CallToolRequest, input ValidateInput) (*mcp.CallToolResult, any, error) {
	if input.Path == "" {
		return nil, nil, fmt.Errorf("path is required")
	}

	store := spec.NewStore(input.Path)

	exists, err := store.Exists()
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error checking spec folder: %v", err)}},
		}, nil, nil
	}

	if !exists {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "No openspec folder found. Run goent_spec_init first."}},
		}, nil, nil
	}

	validator := spec.NewValidator()

	// Default type
	if input.Type == "" {
		if input.ID != "" {
			input.Type = "change"
		} else {
			input.Type = "all"
		}
	}

	var result *spec.ValidationResult
	var targetDesc string

	switch input.Type {
	case "spec":
		if input.ID == "" {
			// Validate all specs
			specsPath := filepath.Join(store.SpecPath(), "specs")
			result, err = validator.ValidateAllSpecs(specsPath, input.Strict)
			targetDesc = "all specs"
		} else {
			// Validate specific spec
			specPath := filepath.Join(store.SpecPath(), "specs", input.ID, "spec.md")
			result, err = validator.ValidateSpec(specPath, input.Strict)
			targetDesc = fmt.Sprintf("spec '%s'", input.ID)
		}

	case "change":
		if input.ID == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "ID is required when type is 'change'"}},
			}, nil, nil
		}
		changePath := filepath.Join(store.SpecPath(), "changes", input.ID)
		result, err = validator.ValidateChange(changePath, input.Strict)
		targetDesc = fmt.Sprintf("change '%s'", input.ID)

	case "all":
		// Validate all specs and changes
		var allIssues []spec.ValidationIssue

		// Validate specs
		specsPath := filepath.Join(store.SpecPath(), "specs")
		specsResult, specsErr := validator.ValidateAllSpecs(specsPath, input.Strict)
		if specsErr == nil {
			allIssues = append(allIssues, specsResult.Issues...)
		}

		// Validate active changes
		changesPath := filepath.Join(store.SpecPath(), "changes")
		changes, _ := store.ListChanges("")
		for _, change := range changes {
			if change.Status == "archived" {
				continue
			}
			changePath := filepath.Join(changesPath, change.ID)
			changeResult, changeErr := validator.ValidateChange(changePath, input.Strict)
			if changeErr == nil {
				allIssues = append(allIssues, changeResult.Issues...)
			}
		}

		result = &spec.ValidationResult{
			Issues: allIssues,
		}
		if input.Strict {
			result.Valid = len(allIssues) == 0
		} else {
			result.Valid = result.ErrorCount() == 0
		}
		result.Summary = buildValidationSummary(result)
		targetDesc = "all specs and changes"

	default:
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid type: %s. Must be 'spec', 'change', or 'all'", input.Type)}},
		}, nil, nil
	}

	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error validating: %v", err)}},
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: formatValidationResult(result, targetDesc, input.Strict)}},
	}, nil, nil
}

func formatValidationResult(result *spec.ValidationResult, target string, strict bool) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Validation of %s\n", target))
	sb.WriteString(fmt.Sprintf("Mode: %s\n\n", modeString(strict)))
	sb.WriteString(result.Summary)
	sb.WriteString("\n")

	if len(result.Issues) > 0 {
		sb.WriteString("\nIssues:\n")
		for _, issue := range result.Issues {
			sb.WriteString(fmt.Sprintf("  %s\n", issue.String()))
		}
	}

	return sb.String()
}

func modeString(strict bool) string {
	if strict {
		return "strict (warnings are errors)"
	}
	return "normal"
}

func buildValidationSummary(result *spec.ValidationResult) string {
	errors := result.ErrorCount()
	warnings := result.WarningCount()

	if errors == 0 && warnings == 0 {
		return "✅ Validation passed with no issues"
	}

	var parts []string
	if errors > 0 {
		parts = append(parts, fmt.Sprintf("%d error(s)", errors))
	}
	if warnings > 0 {
		parts = append(parts, fmt.Sprintf("%d warning(s)", warnings))
	}

	status := "✅ Passed"
	if !result.Valid {
		status = "❌ Failed"
	}

	return fmt.Sprintf("%s: %s", status, strings.Join(parts, ", "))
}
