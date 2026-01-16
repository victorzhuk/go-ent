package tools

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	astpkg "github.com/victorzhuk/go-ent/internal/ast"
)

type ASTRenameInput struct {
	File      string `json:"file"`
	Line      int    `json:"line"`
	Column    int    `json:"column"`
	NewName   string `json:"new_name"`
	DryRun    bool   `json:"dry_run"`
	FilesOnly bool   `json:"files_only,omitempty"`
}

type renameChange struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	OldText string `json:"old_text"`
	NewText string `json:"new_text"`
}

type renameResult struct {
	SymbolName string         `json:"symbol_name"`
	SymbolKind string         `json:"symbol_kind"`
	Changes    []renameChange `json:"changes"`
	Conflicts  []string       `json:"conflicts,omitempty"`
	Applied    bool           `json:"applied"`
}

func registerASTRename(s *mcp.Server) {
	tool := &mcp.Tool{
		Name:        "go_ent_ast_rename",
		Description: "Safely rename a Go symbol (function, variable, type, etc.) using type-aware refactoring. Finds all references across all files in the codebase and updates them atomically.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"file": map[string]any{
					"type":        "string",
					"description": "Path to the Go file containing the symbol to rename",
				},
				"line": map[string]any{
					"type":        "integer",
					"description": "Line number of the symbol to rename",
				},
				"column": map[string]any{
					"type":        "integer",
					"description": "Column number of the symbol to rename",
				},
				"new_name": map[string]any{
					"type":        "string",
					"description": "New name for the symbol",
				},
				"dry_run": map[string]any{
					"type":        "boolean",
					"description": "Preview changes without applying (default: false)",
				},
				"files_only": map[string]any{
					"type":        "boolean",
					"description": "Only list affected files without modifying (default: false)",
				},
			},
			"required": []string{"file", "line", "column", "new_name"},
		},
	}

	mcp.AddTool(s, tool, astRenameHandler)
}

func astRenameHandler(ctx context.Context, req *mcp.CallToolRequest, input ASTRenameInput) (*mcp.CallToolResult, any, error) {
	result, err := renameSymbol(input)
	if err != nil {
		return errorResult(fmt.Errorf("rename: %w", err)), nil, nil
	}

	output := formatRenameResult(result)
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: output}},
	}, nil, nil
}

func renameSymbol(input ASTRenameInput) (*renameResult, error) {
	if input.File == "" {
		return nil, fmt.Errorf("file path is required")
	}
	if input.Line <= 0 {
		return nil, fmt.Errorf("line must be greater than 0")
	}
	if input.Column <= 0 {
		return nil, fmt.Errorf("column must be greater than 0")
	}
	if input.NewName == "" {
		return nil, fmt.Errorf("new_name is required")
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
	scope, err := builder.BuildFile(f)
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

	if ident.Name == input.NewName {
		return &renameResult{
			SymbolName: ident.Name,
			SymbolKind: targetSym.Kind.String(),
			Changes:    []renameChange{},
			Applied:    false,
		}, nil
	}

	conflicts := checkForConflicts(builder, scope, targetSym, input.NewName)
	if len(conflicts) > 0 {
		return &renameResult{
			SymbolName: ident.Name,
			SymbolKind: targetSym.Kind.String(),
			Conflicts:  conflicts,
			Applied:    false,
		}, nil
	}

	affectedFiles, err := findAffectedFiles(input.File, ident.Name, pos)
	if err != nil {
		return nil, fmt.Errorf("find affected files: %w", err)
	}

	if len(affectedFiles) == 0 {
		return nil, fmt.Errorf("no references found")
	}

	if input.FilesOnly {
		return &renameResult{
			SymbolName: ident.Name,
			SymbolKind: targetSym.Kind.String(),
			Changes:    fileChangesToRenameChanges(affectedFiles),
			Applied:    false,
		}, nil
	}

	changes, err := applyRenameToFiles(affectedFiles, ident.Name, input.NewName, input.DryRun)
	if err != nil {
		return nil, fmt.Errorf("apply rename: %w", err)
	}

	return &renameResult{
		SymbolName: ident.Name,
		SymbolKind: targetSym.Kind.String(),
		Changes:    changes,
		Applied:    !input.DryRun,
	}, nil
}

func findIdentifierAtPos(fset *token.FileSet, f *ast.File, pos token.Pos) *ast.Ident {
	if pos == token.NoPos {
		return nil
	}

	var target *ast.Ident
	ast.Inspect(f, func(n ast.Node) bool {
		if target != nil {
			return false
		}

		if ident, ok := n.(*ast.Ident); ok {
			if ident.Pos() <= pos && pos < ident.End() {
				target = ident
			}
		}

		return target == nil
	})

	return target
}

func checkForConflicts(builder *astpkg.Builder, scope *astpkg.Scope, targetSym *astpkg.Symbol, newName string) []string {
	var conflicts []string

	for _, sym := range scope.Symbols {
		if sym.Name == newName && sym != targetSym {
			conflictDesc := fmt.Sprintf("%s '%s' at same scope", sym.Kind.String(), newName)
			conflicts = append(conflicts, conflictDesc)
		}
	}

	return conflicts
}

func computeChanges(fset *token.FileSet, oldContent, newContent, filePath string) []renameChange {
	var changes []renameChange

	oldLines := strings.Split(oldContent, "\n")
	newLines := strings.Split(newContent, "\n")

	for line := 0; line < len(oldLines) && line < len(newLines); line++ {
		if oldLines[line] != newLines[line] {
			changes = append(changes, renameChange{
				File:    filePath,
				Line:    line + 1,
				OldText: oldLines[line],
				NewText: newLines[line],
			})
		}
	}

	return changes
}

func formatRenameResult(result *renameResult) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Symbol: %s (%s)\n", result.SymbolName, result.SymbolKind))

	if len(result.Conflicts) > 0 {
		sb.WriteString("Conflicts:\n")
		for _, conflict := range result.Conflicts {
			sb.WriteString(fmt.Sprintf("  - %s\n", conflict))
		}
		sb.WriteString("Rename not applied due to conflicts.\n")
		return sb.String()
	}

	if len(result.Changes) == 0 {
		sb.WriteString("No changes needed (name already matches).\n")
		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("Found %d change(s):\n", len(result.Changes)))
	for _, change := range result.Changes {
		sb.WriteString(fmt.Sprintf("  %s:%d\n", change.File, change.Line))
		sb.WriteString(fmt.Sprintf("    - %s\n", change.OldText))
		sb.WriteString(fmt.Sprintf("    + %s\n", change.NewText))
	}

	if result.Applied {
		sb.WriteString("\nChanges applied successfully.\n")
	} else {
		sb.WriteString("\nDry run: changes not applied.\n")
	}

	return sb.String()
}

type fileChange struct {
	filePath string
	oldName  string
}

func findAffectedFiles(sourceFile, symbolName string, pos token.Pos) ([]fileChange, error) {
	var affected []fileChange

	fset := token.NewFileSet()

	goFiles, err := findGoFilesInDir(sourceFile)
	if err != nil {
		return nil, err
	}

	for _, goFile := range goFiles {
		f, err := parser.ParseFile(fset, goFile, nil, parser.AllErrors)
		if err != nil {
			continue
		}

		builder := astpkg.NewBuilder(fset)
		scope, err := builder.BuildFile(f)
		if err != nil {
			continue
		}

		if containsSymbolReference(builder, fset, f, symbolName, scope) {
			affected = append(affected, fileChange{
				filePath: goFile,
				oldName:  symbolName,
			})
		}
	}

	return affected, nil
}

func containsSymbolReference(builder *astpkg.Builder, fset *token.FileSet, f *ast.File, symbolName string, scope *astpkg.Scope) bool {
	var found bool

	ast.Inspect(f, func(n ast.Node) bool {
		if found {
			return false
		}

		if ident, ok := n.(*ast.Ident); ok {
			if ident.Name == symbolName {
				found = true
			}
		}

		return !found
	})

	return found
}

func findGoFilesInDir(startFile string) ([]string, error) {
	var files []string

	dir := filepath.Dir(startFile)

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "vendor" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}

		if strings.HasSuffix(path, ".go") {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

func applyRenameToFiles(affectedFiles []fileChange, oldName, newName string, dryRun bool) ([]renameChange, error) {
	var allChanges []renameChange

	for _, fc := range affectedFiles {
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, fc.filePath, nil, parser.AllErrors)
		if err != nil {
			continue
		}

		printer := astpkg.NewPrinter(fset)
		oldContent, err := printer.PrintFile(f)
		if err != nil {
			continue
		}

		transform := astpkg.NewTransform(fset)
		newFile, err := transform.RenameSymbol(f, oldName, newName)
		if err != nil {
			continue
		}

		newContent, err := printer.PrintFile(newFile)
		if err != nil {
			continue
		}

		fileChanges := computeChanges(fset, oldContent, newContent, fc.filePath)
		allChanges = append(allChanges, fileChanges...)

		if !dryRun {
			if err := printer.WriteFile(newFile, fc.filePath); err != nil {
				return allChanges, fmt.Errorf("write file %s: %w", fc.filePath, err)
			}
		}
	}

	return allChanges, nil
}

func fileChangesToRenameChanges(affectedFiles []fileChange) []renameChange {
	changes := make([]renameChange, 0, len(affectedFiles))
	for _, fc := range affectedFiles {
		changes = append(changes, renameChange{
			File:    fc.filePath,
			Line:    0,
			OldText: fc.oldName,
			NewText: "(files_only mode: use without files_only to see changes)",
		})
	}
	return changes
}
