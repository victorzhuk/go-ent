package tools

import (
	"context"
	"fmt"
	"go/ast"
	"os"
	"path/filepath"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	astpkg "github.com/victorzhuk/go-ent/internal/ast"
)

type ASTQueryInput struct {
	File      string `json:"file,omitempty"`
	Package   string `json:"package,omitempty"`
	Type      string `json:"type"`
	Pattern   string `json:"pattern,omitempty"`
	Signature string `json:"signature,omitempty"`
	Interface string `json:"interface,omitempty"`
	FieldType string `json:"field_type,omitempty"`
}

func registerASTQuery(s *mcp.Server) {
	tool := &mcp.Tool{
		Name:        "go_ent_ast_query",
		Description: "Query Go AST to find functions, types, and interfaces by pattern, signature, interface implementation, or struct field type",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"file": map[string]any{
					"type":        "string",
					"description": "Path to the Go file to query (alternative to package)",
				},
				"package": map[string]any{
					"type":        "string",
					"description": "Package path to query (use '...' for recursive search)",
				},
				"type": map[string]any{
					"type":        "string",
					"description": "Query type: function, implements, struct_field",
					"enum":        []string{"function", "implements", "struct_field"},
				},
				"pattern": map[string]any{
					"type":        "string",
					"description": "Function name pattern (supports wildcard '*')",
				},
				"signature": map[string]any{
					"type":        "string",
					"description": "Function signature pattern (e.g., '(context.Context) error')",
				},
				"interface": map[string]any{
					"type":        "string",
					"description": "Interface name to find implementations (e.g., 'io.Reader')",
				},
				"field_type": map[string]any{
					"type":        "string",
					"description": "Struct field type to search (e.g., 'string', '*http.Client', '[]string')",
				},
			},
			"oneOf": []map[string]any{
				{"required": []string{"file", "type"}},
				{"required": []string{"package", "type"}},
			},
		},
	}

	mcp.AddTool(s, tool, astQueryHandler)
}

func astQueryHandler(ctx context.Context, req *mcp.CallToolRequest, input ASTQueryInput) (*mcp.CallToolResult, any, error) {
	parser := astpkg.NewParser()

	var results []astpkg.Result

	switch {
	case input.File != "":
		fileResults, err := queryFile(parser, input.File, input)
		if err != nil {
			return errorResult(fmt.Errorf("query file %s: %w", input.File, err)), nil, nil
		}
		results = append(results, fileResults...)
	case input.Package != "":
		pkgResults, err := queryPackage(parser, input.Package, input)
		if err != nil {
			return errorResult(fmt.Errorf("query package %s: %w", input.Package, err)), nil, nil
		}
		results = append(results, pkgResults...)
	default:
		return errorResult(fmt.Errorf("either file or package must be specified")), nil, nil
	}

	output := formatQueryResults(results)
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: output}},
	}, nil, nil
}

func queryFile(parser *astpkg.Parser, filePath string, input ASTQueryInput) ([]astpkg.Result, error) {
	f, err := parser.ParseFile(filePath)
	if err != nil {
		if os.IsNotExist(err) || strings.Contains(err.Error(), "no such file") {
			return []astpkg.Result{}, nil
		}
		return nil, err
	}

	query := astpkg.NewQuery(parser.FileSet())
	files := map[string]*ast.File{filePath: f}

	return executeQuery(query, files, input)
}

func queryPackage(parser *astpkg.Parser, pkgPath string, input ASTQueryInput) ([]astpkg.Result, error) {
	paths, err := findGoFiles(pkgPath)
	if err != nil {
		return nil, err
	}

	if len(paths) == 0 {
		return []astpkg.Result{}, nil
	}

	query := astpkg.NewQuery(parser.FileSet())
	files := make(map[string]*ast.File)

	for _, path := range paths {
		f, err := parser.ParseFile(path)
		if err != nil {
			continue
		}
		files[path] = f
	}

	return executeQuery(query, files, input)
}

func findGoFiles(pkgPath string) ([]string, error) {
	var paths []string

	isRecursive := strings.HasSuffix(pkgPath, "/...") || strings.HasSuffix(pkgPath, `\...`)
	searchDir := strings.TrimSuffix(pkgPath, "/...")
	searchDir = strings.TrimSuffix(searchDir, `\...`)

	info, err := os.Stat(searchDir)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		if strings.HasSuffix(searchDir, ".go") {
			if _, err := os.Stat(searchDir); err == nil {
				return []string{searchDir}, nil
			}
		}
		return nil, nil
	}

	if isRecursive {
		err := filepath.WalkDir(searchDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && strings.HasSuffix(path, ".go") {
				paths = append(paths, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		entries, err := os.ReadDir(searchDir)
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			if strings.HasSuffix(entry.Name(), ".go") {
				paths = append(paths, filepath.Join(searchDir, entry.Name()))
			}
		}
	}

	return paths, nil
}

func executeQuery(query *astpkg.Query, files map[string]*ast.File, input ASTQueryInput) ([]astpkg.Result, error) {
	switch input.Type {
	case "function":
		return queryFunctions(query, files, input)
	case "implements":
		return queryImplementations(query, files, input)
	case "struct_field":
		return queryStructFields(query, files, input)
	default:
		return nil, fmt.Errorf("invalid query type: %s", input.Type)
	}
}

func queryFunctions(query *astpkg.Query, files map[string]*ast.File, input ASTQueryInput) ([]astpkg.Result, error) {
	if input.Signature != "" {
		return query.FindBySignature(files, input.Signature), nil
	}
	if input.Pattern != "" {
		return query.FindFunctions(files, input.Pattern), nil
	}
	return query.FindFunctions(files, "*"), nil
}

func queryImplementations(query *astpkg.Query, files map[string]*ast.File, input ASTQueryInput) ([]astpkg.Result, error) {
	if input.Interface == "" {
		return nil, fmt.Errorf("interface name required for implements query")
	}
	return query.FindImplementations(files, input.Interface), nil
}

func queryStructFields(query *astpkg.Query, files map[string]*ast.File, input ASTQueryInput) ([]astpkg.Result, error) {
	if input.FieldType == "" {
		return nil, fmt.Errorf("field type required for struct_field query")
	}
	return query.FindStructsByFieldType(files, input.FieldType), nil
}

func formatQueryResults(results []astpkg.Result) string {
	if len(results) == 0 {
		return "No matches found\n"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d match%s:\n", len(results), pluralize(len(results))))

	for _, r := range results {
		switch r.Type {
		case "implementation":
			sb.WriteString(fmt.Sprintf("  - %s implements %s (%s:%d)\n", r.Name, r.Signature, filepath.Base(r.File), r.Line))
		case "struct_field":
			sb.WriteString(fmt.Sprintf("  - %s has field %s (%s:%d)\n", r.Name, r.Signature, filepath.Base(r.File), r.Line))
		default:
			sb.WriteString(fmt.Sprintf("  - %s (%s:%d)\n", r.Name, filepath.Base(r.File), r.Line))
		}
	}

	return sb.String()
}

func pluralize(n int) string {
	if n == 1 {
		return ""
	}
	return "es"
}
