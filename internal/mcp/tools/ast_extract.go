package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	astpkg "github.com/victorzhuk/go-ent/internal/ast"
)

type ASTExtractInput struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	EndLine int    `json:"end_line"`
	Name    string `json:"name"`
	DryRun  bool   `json:"dry_run"`
}

type extractChange struct {
	File      string `json:"file"`
	Line      int    `json:"line"`
	OldText   string `json:"old_text,omitempty"`
	NewText   string `json:"new_text,omitempty"`
	Extracted string `json:"extracted,omitempty"`
}

type extractResult struct {
	FuncName  string          `json:"func_name"`
	StartLine int             `json:"start_line"`
	EndLine   int             `json:"end_line"`
	Changes   []extractChange `json:"changes"`
	Applied   bool            `json:"applied"`
}

func registerASTExtract(s *mcp.Server) {
	tool := &mcp.Tool{
		Name:        "go_ent_ast_extract",
		Description: "Extract a range of code into a new function. The selected statements will be replaced with a call to the new function.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"file": map[string]any{
					"type":        "string",
					"description": "Path to the Go file containing the code to extract",
				},
				"line": map[string]any{
					"type":        "integer",
					"description": "Line number of the start of the code to extract",
				},
				"end_line": map[string]any{
					"type":        "integer",
					"description": "Line number of the end of the code to extract",
				},
				"name": map[string]any{
					"type":        "string",
					"description": "Name for the extracted function",
				},
				"dry_run": map[string]any{
					"type":        "boolean",
					"description": "Preview changes without applying (default: false)",
				},
			},
			"required": []string{"file", "line", "end_line", "name"},
		},
	}

	mcp.AddTool(s, tool, astExtractHandler)
}

func astExtractHandler(ctx context.Context, req *mcp.CallToolRequest, input ASTExtractInput) (*mcp.CallToolResult, any, error) {
	result, err := extractFunc(input)
	if err != nil {
		return errorResult(fmt.Errorf("extract: %w", err)), nil, nil
	}

	output := formatExtractResult(result)
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: output}},
	}, nil, nil
}

func extractFunc(input ASTExtractInput) (*extractResult, error) {
	if input.File == "" {
		return nil, fmt.Errorf("file path is required")
	}
	if input.Line <= 0 {
		return nil, fmt.Errorf("line must be greater than 0")
	}
	if input.EndLine <= 0 {
		return nil, fmt.Errorf("end_line must be greater than 0")
	}
	if input.EndLine < input.Line {
		return nil, fmt.Errorf("end_line must be >= line")
	}
	if input.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	parser := astpkg.NewParser()
	f, err := parser.ParseFile(input.File)
	if err != nil {
		if strings.Contains(err.Error(), "no such file") {
			return nil, fmt.Errorf("file not found: %s", input.File)
		}
		return nil, fmt.Errorf("parse file: %w", err)
	}

	transform := astpkg.NewTransform(parser.FileSet())
	newFile, err := transform.ExtractFunc(f, input.Line, input.EndLine, input.Name)
	if err != nil {
		return nil, fmt.Errorf("extract function: %w", err)
	}

	printer := astpkg.NewPrinter(parser.FileSet())
	oldContent, err := printer.PrintFile(f)
	if err != nil {
		return nil, fmt.Errorf("print original file: %w", err)
	}

	newContent, err := printer.PrintFile(newFile)
	if err != nil {
		return nil, fmt.Errorf("print extracted file: %w", err)
	}

	changes := computeExtractChanges(oldContent, newContent, input)

	if !input.DryRun {
		if err := printer.WriteFile(newFile, input.File); err != nil {
			return nil, fmt.Errorf("write file: %w", err)
		}
	}

	return &extractResult{
		FuncName:  input.Name,
		StartLine: input.Line,
		EndLine:   input.EndLine,
		Changes:   changes,
		Applied:   !input.DryRun,
	}, nil
}

func computeExtractChanges(oldContent, newContent string, input ASTExtractInput) []extractChange {
	var changes []extractChange

	oldLines := strings.Split(oldContent, "\n")
	newLines := strings.Split(newContent, "\n")

	startLine := input.Line - 1
	endLine := input.EndLine - 1

	if startLine < 0 || endLine >= len(oldLines) {
		return changes
	}

	var extracted []string
	for i := startLine; i <= endLine && i < len(oldLines); i++ {
		extracted = append(extracted, oldLines[i])
	}

	changes = append(changes, extractChange{
		File:      input.File,
		Line:      input.Line,
		Extracted: strings.Join(extracted, "\n"),
	})

	for line := 0; line < len(oldLines) && line < len(newLines); line++ {
		if oldLines[line] != newLines[line] {
			changes = append(changes, extractChange{
				File:    input.File,
				Line:    line + 1,
				OldText: oldLines[line],
				NewText: newLines[line],
			})
		}
	}

	return changes
}

func formatExtractResult(result *extractResult) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Extracted function: %s\n", result.FuncName))
	sb.WriteString(fmt.Sprintf("Line range: %d-%d\n\n", result.StartLine, result.EndLine))

	if len(result.Changes) > 0 && result.Changes[0].Extracted != "" {
		sb.WriteString("Extracted code:\n")
		for _, line := range strings.Split(result.Changes[0].Extracted, "\n") {
			sb.WriteString(fmt.Sprintf("  %s\n", line))
		}
		sb.WriteString("\n")
	}

	if len(result.Changes) > 1 {
		sb.WriteString(fmt.Sprintf("Modified %d line(s):\n", len(result.Changes)-1))
		for i, change := range result.Changes {
			if i == 0 {
				continue
			}
			sb.WriteString(fmt.Sprintf("  %s:%d\n", change.File, change.Line))
			if change.OldText != "" {
				sb.WriteString(fmt.Sprintf("    - %s\n", change.OldText))
			}
			if change.NewText != "" {
				sb.WriteString(fmt.Sprintf("    + %s\n", change.NewText))
			}
		}
		sb.WriteString("\n")
	}

	if result.Applied {
		sb.WriteString("Extraction applied successfully.\n")
	} else {
		sb.WriteString("Dry run: changes not applied.\n")
	}

	return sb.String()
}
