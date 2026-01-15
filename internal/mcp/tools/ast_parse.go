package tools

import (
	"context"
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	astpkg "github.com/victorzhuk/go-ent/internal/ast"
)

type ASTParseInput struct {
	File             string `json:"file"`
	Package          string `json:"package,omitempty"`
	IncludePositions bool   `json:"include_positions,omitempty"`
}

type parsedSymbol struct {
	Name     string `json:"name"`
	Kind     string `json:"kind"`
	Type     string `json:"type,omitempty"`
	File     string `json:"file,omitempty"`
	Line     int    `json:"line,omitempty"`
	Column   int    `json:"column,omitempty"`
	Exported bool   `json:"exported,omitempty"`
}

type parseResult struct {
	Package   string         `json:"package"`
	Imports   []string       `json:"imports"`
	Functions []parsedSymbol `json:"functions"`
	Types     []parsedSymbol `json:"types"`
	Variables []parsedSymbol `json:"variables"`
	Constants []parsedSymbol `json:"constants"`
	Methods   []parsedSymbol `json:"methods"`
	Fields    []parsedSymbol `json:"fields"`
	Errors    []string       `json:"errors,omitempty"`
}

func registerASTParse(s *mcp.Server) {
	tool := &mcp.Tool{
		Name:        "go_ent_ast_parse",
		Description: "Parse Go source files into AST and return structured information about functions, types, imports, and symbols",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"file": map[string]any{
					"type":        "string",
					"description": "Path to the Go file to parse",
				},
				"package": map[string]any{
					"type":        "string",
					"description": "Package path to parse all files in the package (alternative to file)",
				},
				"include_positions": map[string]any{
					"type":        "boolean",
					"description": "Include line and column numbers for each symbol",
				},
			},
			"oneOf": []map[string]any{
				{"required": []string{"file"}},
				{"required": []string{"package"}},
			},
		},
	}

	mcp.AddTool(s, tool, astParseHandler)
}

func astParseHandler(ctx context.Context, req *mcp.CallToolRequest, input ASTParseInput) (*mcp.CallToolResult, any, error) {
	parser := astpkg.NewParser()

	var results []*parseResult

	switch {
	case input.File != "":
		result, err := parseFile(parser, input.File, input.IncludePositions)
		if err != nil {
			return errorResult(fmt.Errorf("parse file %s: %w", input.File, err)), nil, nil
		}
		results = append(results, result)
	case input.Package != "":
		pkgResults, err := parsePackage(parser, input.Package, input.IncludePositions)
		if err != nil {
			return errorResult(fmt.Errorf("parse package %s: %w", input.Package, err)), nil, nil
		}
		results = append(results, pkgResults...)
	default:
		return errorResult(fmt.Errorf("either file or package must be specified")), nil, nil
	}

	output, err := formatParseResults(results)
	if err != nil {
		return errorResult(fmt.Errorf("format results: %w", err)), nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: output}},
	}, nil, nil
}

func parseFile(parser *astpkg.Parser, filePath string, includePositions bool) (*parseResult, error) {
	f, err := parser.ParseFile(filePath)
	if err != nil {
		if os.IsNotExist(err) || strings.Contains(err.Error(), "no such file") || strings.Contains(err.Error(), "open:") {
			return &parseResult{
				Errors: []string{fmt.Sprintf("file not found: %s", filePath)},
			}, nil
		}
		return &parseResult{
			Errors: []string{fmt.Sprintf("syntax error: %v", err)},
		}, nil
	}

	builder := astpkg.NewBuilder(parser.FileSet())
	scope, err := builder.BuildFile(f)
	if err != nil {
		return &parseResult{
			Errors: []string{fmt.Sprintf("build scope: %v", err)},
		}, nil
	}

	return buildParseResult(parser.FileSet(), f, scope, filePath, includePositions)
}

func parsePackage(parser *astpkg.Parser, pkgPath string, includePositions bool) ([]*parseResult, error) {
	entries, err := os.ReadDir(pkgPath)
	if err != nil {
		return nil, fmt.Errorf("read package directory: %w", err)
	}

	var results []*parseResult

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}

		filePath := fmt.Sprintf("%s/%s", pkgPath, entry.Name())
		result, err := parseFile(parser, filePath, includePositions)
		if err != nil {
			results = append(results, &parseResult{
				Errors: []string{fmt.Sprintf("parse %s: %v", filePath, err)},
			})
			continue
		}
		results = append(results, result)
	}

	return results, nil
}

func buildParseResult(fset *token.FileSet, f *ast.File, scope *astpkg.Scope, filePath string, includePositions bool) (*parseResult, error) {
	result := &parseResult{
		Package: f.Name.Name,
		Imports: extractImports(f),
	}

	for _, sym := range scope.Symbols {
		psym := toParsedSymbol(fset, sym, filePath, includePositions)
		if psym == nil {
			continue
		}

		switch sym.Kind {
		case astpkg.SymbolFunction:
			result.Functions = append(result.Functions, *psym)
		case astpkg.SymbolType:
			result.Types = append(result.Types, *psym)
		case astpkg.SymbolVariable:
			result.Variables = append(result.Variables, *psym)
		case astpkg.SymbolConstant:
			result.Constants = append(result.Constants, *psym)
		case astpkg.SymbolMethod:
			result.Methods = append(result.Methods, *psym)
		case astpkg.SymbolField:
			result.Fields = append(result.Fields, *psym)
		}
	}

	return result, nil
}

func toParsedSymbol(fset *token.FileSet, sym *astpkg.Symbol, filePath string, includePositions bool) *parsedSymbol {
	if sym == nil {
		return nil
	}

	psym := &parsedSymbol{
		Name: sym.Name,
		Kind: sym.Kind.String(),
	}

	if includePositions {
		pos := fset.Position(sym.Pos)
		psym.File = filePath
		psym.Line = pos.Line
		psym.Column = pos.Column
	}

	if sym.Type != nil {
		psym.Type = formatExpr(sym.Type)
	}

	exported := false
	if sym.Kind == astpkg.SymbolType || sym.Kind == astpkg.SymbolFunction || sym.Kind == astpkg.SymbolMethod || sym.Kind == astpkg.SymbolConstant || sym.Kind == astpkg.SymbolField {
		exported = isExported(sym.Name)
	}
	psym.Exported = exported

	return psym
}

func formatExpr(e ast.Expr) string {
	if e == nil {
		return ""
	}

	switch t := e.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", formatExpr(t.X), t.Sel.Name)
	case *ast.StarExpr:
		return fmt.Sprintf("*%s", formatExpr(t.X))
	case *ast.ArrayType:
		return fmt.Sprintf("[]%s", formatExpr(t.Elt))
	case *ast.Ellipsis:
		return fmt.Sprintf("...%s", formatExpr(t.Elt))
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", formatExpr(t.Key), formatExpr(t.Value))
	case *ast.FuncType:
		return formatFuncType(t)
	case *ast.InterfaceType:
		return "any"
	case *ast.ChanType:
		return formatChanType(t)
	case *ast.StructType:
		return "struct{...}"
	default:
		return ""
	}
}

func formatFuncType(ft *ast.FuncType) string {
	if ft == nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("(")

	for i, p := range ft.Params.List {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(formatExpr(p.Type))
	}

	sb.WriteString(")")

	if ft.Results != nil && len(ft.Results.List) > 0 {
		sb.WriteString(" ")
		if len(ft.Results.List) > 1 {
			sb.WriteString("(")
		}

		for i, r := range ft.Results.List {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(formatExpr(r.Type))
		}

		if len(ft.Results.List) > 1 {
			sb.WriteString(")")
		}
	}

	return sb.String()
}

func formatChanType(ct *ast.ChanType) string {
	if ct == nil {
		return ""
	}

	var dir string
	switch ct.Dir {
	case ast.SEND:
		dir = "chan<- "
	case ast.RECV:
		dir = "<-chan "
	default:
		dir = "chan "
	}
	return dir + formatExpr(ct.Value)
}

func extractImports(f *ast.File) []string {
	var imports []string
	for _, imp := range f.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		imports = append(imports, path)
	}
	return imports
}

func isExported(name string) bool {
	if len(name) == 0 {
		return false
	}
	return name[0] >= 'A' && name[0] <= 'Z'
}

func formatParseResults(results []*parseResult) (string, error) {
	var sb strings.Builder

	if len(results) == 1 {
		sb.WriteString(formatSingleResult(results[0]))
	} else {
		for i, result := range results {
			sb.WriteString(fmt.Sprintf("File %d:\n", i+1))
			sb.WriteString(formatSingleResult(result))
			if i < len(results)-1 {
				sb.WriteString("\n")
			}
		}
	}

	return sb.String(), nil
}

func formatSingleResult(result *parseResult) string {
	var sb strings.Builder

	if len(result.Errors) > 0 {
		sb.WriteString("Errors:\n")
		for _, err := range result.Errors {
			sb.WriteString(fmt.Sprintf("  - %s\n", err))
		}
	}

	sb.WriteString(fmt.Sprintf("Package: %s\n", result.Package))

	if len(result.Imports) > 0 {
		sb.WriteString("Imports:\n")
		for _, imp := range result.Imports {
			sb.WriteString(fmt.Sprintf("  - %s\n", imp))
		}
	}

	if len(result.Functions) > 0 {
		sb.WriteString("Functions:\n")
		for _, fn := range result.Functions {
			sb.WriteString(formatSymbol(fn))
		}
	}

	if len(result.Types) > 0 {
		sb.WriteString("Types:\n")
		for _, typ := range result.Types {
			sb.WriteString(formatSymbol(typ))
		}
	}

	if len(result.Methods) > 0 {
		sb.WriteString("Methods:\n")
		for _, meth := range result.Methods {
			sb.WriteString(formatSymbol(meth))
		}
	}

	if len(result.Variables) > 0 {
		sb.WriteString("Variables:\n")
		for _, v := range result.Variables {
			sb.WriteString(formatSymbol(v))
		}
	}

	if len(result.Constants) > 0 {
		sb.WriteString("Constants:\n")
		for _, c := range result.Constants {
			sb.WriteString(formatSymbol(c))
		}
	}

	if len(result.Fields) > 0 {
		sb.WriteString("Fields:\n")
		for _, f := range result.Fields {
			sb.WriteString(formatSymbol(f))
		}
	}

	return sb.String()
}

func formatSymbol(sym parsedSymbol) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("  - %s", sym.Name))

	if sym.Type != "" {
		sb.WriteString(fmt.Sprintf(" %s", sym.Type))
	}

	if sym.Exported {
		sb.WriteString(" (exported)")
	}

	if sym.Line > 0 {
		sb.WriteString(fmt.Sprintf(" [line: %d, column: %d]", sym.Line, sym.Column))
	}

	sb.WriteString("\n")
	return sb.String()
}

func errorResult(err error) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error: %v", err)}},
	}
}
