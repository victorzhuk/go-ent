package tools

import (
	"context"
	"fmt"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	astpkg "github.com/victorzhuk/go-ent/internal/ast"
)

type ASTRefsInput struct {
	File         string `json:"file"`
	Line         int    `json:"line"`
	Column       int    `json:"column"`
	IncludeTests bool   `json:"include_tests,omitempty"`
}

func registerASTRefs(s *mcp.Server) {
	tool := &mcp.Tool{
		Name:        "go_ent_ast_refs",
		Description: "Find all references to a Go symbol (function, variable, type, etc.) in the codebase",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"file": map[string]any{
					"type":        "string",
					"description": "Path to the Go file containing the symbol",
				},
				"line": map[string]any{
					"type":        "integer",
					"description": "Line number of the symbol reference",
				},
				"column": map[string]any{
					"type":        "integer",
					"description": "Column number of the symbol reference",
				},
				"include_tests": map[string]any{
					"type":        "boolean",
					"description": "Include references in test files (default: true)",
				},
			},
			"required": []string{"file", "line", "column"},
		},
	}

	mcp.AddTool(s, tool, astRefsHandler)
}

func astRefsHandler(ctx context.Context, req *mcp.CallToolRequest, input ASTRefsInput) (*mcp.CallToolResult, any, error) {
	result, err := findReferences(input)
	if err != nil {
		return errorResult(fmt.Errorf("find references: %w", err)), nil, nil
	}

	output := formatRefsResult(result)
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: output}},
	}, nil, nil
}

type refsResult struct {
	SymbolName string
	SymbolKind string
	File       string
	Line       int
	Column     int
	References []refLocation
}

type refLocation struct {
	File   string
	Line   int
	Column int
	Kind   string
}

func findReferences(input ASTRefsInput) (*refsResult, error) {
	if input.File == "" {
		return nil, fmt.Errorf("file path is required")
	}
	if input.Line <= 0 {
		return nil, fmt.Errorf("line must be greater than 0")
	}
	if input.Column <= 0 {
		return nil, fmt.Errorf("column must be greater than 0")
	}

	parser := astpkg.NewParser()
	f, err := parser.ParseFile(input.File)
	if err != nil {
		if os.IsNotExist(err) || strings.Contains(err.Error(), "no such file") {
			return nil, fmt.Errorf("file not found: %s", input.File)
		}
		return nil, fmt.Errorf("parse file: %w", err)
	}

	builder := astpkg.NewBuilder(parser.FileSet())
	_, err = builder.BuildFile(f)
	if err != nil {
		return nil, fmt.Errorf("build symbol table: %w", err)
	}

	file := parser.FileSet().File(f.Pos())
	pos := file.LineStart(input.Line) + token.Pos(input.Column-1)

	ident := findIdentifierAtPos(parser.FileSet(), f, pos)
	if ident == nil {
		return nil, fmt.Errorf("symbol not found at %s:%d:%d", input.File, input.Line, input.Column)
	}

	targetSym := builder.FindSymbol(ident.Name, pos)
	if targetSym == nil {
		return nil, fmt.Errorf("symbol not found in scope")
	}

	posData := parser.FileSet().Position(targetSym.Pos)

	refs := builder.FindReferences(ident.Name, pos, input.IncludeTests)
	var locations []refLocation
	for _, ref := range refs {
		refPos := parser.FileSet().Position(ref.Pos)
		location := refLocation{
			File:   refPos.Filename,
			Line:   refPos.Line,
			Column: refPos.Column,
			Kind:   ref.Kind.String(),
		}
		locations = append(locations, location)
	}

	return &refsResult{
		SymbolName: targetSym.Name,
		SymbolKind: targetSym.Kind.String(),
		File:       posData.Filename,
		Line:       posData.Line,
		Column:     posData.Column,
		References: locations,
	}, nil
}

func formatRefsResult(result *refsResult) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Symbol: %s (%s)\n", result.SymbolName, result.SymbolKind))
	sb.WriteString(fmt.Sprintf("Definition: %s:%d:%d\n\n", filepath.Base(result.File), result.Line, result.Column))

	if len(result.References) == 0 {
		sb.WriteString("No references found\n")
		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("Found %d reference(s):\n", len(result.References)))

	for i, ref := range result.References {
		if ref.Kind == "definition" {
			sb.WriteString(fmt.Sprintf("  %d. [definition] %s:%d:%d\n", i+1, filepath.Base(ref.File), ref.Line, ref.Column))
		} else {
			sb.WriteString(fmt.Sprintf("  %d. [%s] %s:%d:%d\n", i+1, ref.Kind, filepath.Base(ref.File), ref.Line, ref.Column))
		}
	}

	return sb.String()
}
