package tools

import (
	"context"
	"fmt"
	"go/ast"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	astpkg "github.com/victorzhuk/go-ent/internal/ast"
)

type ASTGenerateInput struct {
	Type     string `json:"type"`
	File     string `json:"file"`
	Function string `json:"function"`
}

type generateResult struct {
	Generated string `json:"generated"`
	File      string `json:"file"`
}

func registerASTGenerate(s *mcp.Server) {
	tool := &mcp.Tool{
		Name:        "go_ent_ast_generate",
		Description: "Generate code scaffolds from existing code. Supports generating test scaffolds from function signatures.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"type": map[string]any{
					"type":        "string",
					"description": "Type of code to generate (e.g., 'test' for test scaffolds)",
				},
				"file": map[string]any{
					"type":        "string",
					"description": "Path to the Go file containing the code to generate from",
				},
				"function": map[string]any{
					"type":        "string",
					"description": "Name of the function to generate test for",
				},
			},
			"required": []string{"type", "file", "function"},
		},
	}

	mcp.AddTool(s, tool, astGenerateHandler)
}

func astGenerateHandler(ctx context.Context, req *mcp.CallToolRequest, input ASTGenerateInput) (*mcp.CallToolResult, any, error) {
	result, err := generateCode(input)
	if err != nil {
		return errorResult(fmt.Errorf("generate: %w", err)), nil, nil
	}

	output := formatGenerateResult(result)
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: output}},
	}, nil, nil
}

func generateCode(input ASTGenerateInput) (*generateResult, error) {
	if input.Type == "" {
		return nil, fmt.Errorf("type is required")
	}
	if input.File == "" {
		return nil, fmt.Errorf("file path is required")
	}
	if input.Function == "" {
		return nil, fmt.Errorf("function name is required")
	}

	parser := astpkg.NewParser()
	f, err := parser.ParseFile(input.File)
	if err != nil {
		return nil, fmt.Errorf("parse file %s: %w", input.File, err)
	}

	var funcDecl *ast.FuncDecl
	ast.Inspect(f, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok && fn.Name != nil {
			if fn.Name.Name == input.Function {
				funcDecl = fn
				return false
			}
		}
		return true
	})

	if funcDecl == nil {
		return nil, fmt.Errorf("function %s not found in file %s", input.Function, input.File)
	}

	var generated ast.Node
	var outputFileName string

	switch strings.ToLower(input.Type) {
	case "test":
		testNode, err := astpkg.GenerateTestScaffold(funcDecl)
		if err != nil {
			return nil, fmt.Errorf("generate test scaffold: %w", err)
		}
		generated = testNode
		outputFileName = getTestFileName(input.File)
	default:
		return nil, fmt.Errorf("unsupported type: %s (supported: 'test')", input.Type)
	}

	printer := astpkg.NewPrinter(parser.FileSet())
	code, err := printer.PrintNode(generated)
	if err != nil {
		return nil, fmt.Errorf("print generated code: %w", err)
	}

	return &generateResult{
		Generated: code,
		File:      outputFileName,
	}, nil
}

func getTestFileName(filePath string) string {
	if strings.HasSuffix(filePath, ".go") {
		return filePath[:len(filePath)-3] + "_test.go"
	}
	return filePath + "_test.go"
}

func formatGenerateResult(result *generateResult) string {
	var sb strings.Builder

	sb.WriteString("Generated Code:\n")
	sb.WriteString("==============\n\n")
	sb.WriteString(result.Generated)
	sb.WriteString("\n\n")
	sb.WriteString(fmt.Sprintf("Save to file: %s\n", result.File))

	return sb.String()
}
