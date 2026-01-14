package ast

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

type Transform struct {
	fset *token.FileSet
}

func NewTransform(fset *token.FileSet) *Transform {
	return &Transform{fset: fset}
}

func (t *Transform) ExtractFunc(f *ast.File, startLine, endLine int, name string) (*ast.File, error) {
	if f == nil {
		return nil, ErrInvalidSource
	}
	if startLine <= 0 || endLine < startLine {
		return nil, fmt.Errorf("invalid line range")
	}
	if name == "" {
		return nil, fmt.Errorf("empty function name")
	}

	newFile := t.copyFile(f)
	extractedStmts, remainingStmts := t.extractStatements(newFile, startLine, endLine)

	if len(extractedStmts) == 0 {
		return nil, fmt.Errorf("no statements in range")
	}

	newFunc := t.createExtractedFunc(name, extractedStmts)

	var newDecls []ast.Decl
	inserted := false
	for _, decl := range newFile.Decls {
		if fd, ok := decl.(*ast.FuncDecl); ok && !inserted {
			if t.containsLine(fd, startLine, endLine) {
				newDecls = append(newDecls, newFunc)
				newDecls = append(newDecls, t.replaceWithCall(fd, remainingStmts, name)...)
				inserted = true
				continue
			}
		}
		newDecls = append(newDecls, decl)
	}

	if !inserted {
		for _, decl := range newFile.Decls {
			if gd, ok := decl.(*ast.GenDecl); ok {
				if t.containsStatementsInRange(gd, startLine, endLine) {
					idx := t.findDeclIndex(newFile.Decls, decl)
					newDecls = t.insertBefore(newDecls, idx, newFunc)
					inserted = true
					break
				}
			}
		}
	}

	if !inserted {
		newFile.Decls = append(newFile.Decls, newFunc)
	} else {
		newFile.Decls = newDecls
	}

	return newFile, nil
}

func (t *Transform) RenameSymbol(f *ast.File, oldName, newName string) (*ast.File, error) {
	if f == nil {
		return nil, ErrInvalidSource
	}
	if oldName == "" || newName == "" {
		return nil, fmt.Errorf("empty name")
	}
	if oldName == newName {
		return f, nil
	}

	newFile := t.copyFile(f)
	renamer := &renamer{oldName: oldName, newName: newName}
	ast.Walk(renamer, newFile)

	return newFile, nil
}

func (t *Transform) InlineVariable(f *ast.File, varName string) (*ast.File, error) {
	if f == nil {
		return nil, ErrInvalidSource
	}
	if varName == "" {
		return nil, fmt.Errorf("empty variable name")
	}

	newFile := t.copyFile(f)
	inliner := &inliner{varName: varName}
	ast.Walk(inliner, newFile)

	return newFile, nil
}

func (t *Transform) copyFile(f *ast.File) *ast.File {
	newFile := &ast.File{
		Name:       ast.NewIdent(f.Name.Name),
		Decls:      make([]ast.Decl, len(f.Decls)),
		Scope:      f.Scope,
		Imports:    f.Imports,
		Unresolved: f.Unresolved,
		Comments:   f.Comments,
	}
	copy(newFile.Decls, f.Decls)
	return newFile
}

func (t *Transform) copyStmt(stmt ast.Stmt) ast.Stmt {
	if stmt == nil {
		return nil
	}
	switch s := stmt.(type) {
	case *ast.ExprStmt:
		return &ast.ExprStmt{X: s.X}
	case *ast.AssignStmt:
		newAssign := &ast.AssignStmt{
			Tok: s.Tok,
			Lhs: append([]ast.Expr{}, s.Lhs...),
			Rhs: append([]ast.Expr{}, s.Rhs...),
		}
		return newAssign
	case *ast.ReturnStmt:
		return &ast.ReturnStmt{
			Results: append([]ast.Expr{}, s.Results...),
		}
	case *ast.BlockStmt:
		newBlock := &ast.BlockStmt{}
		for _, stmt := range s.List {
			newBlock.List = append(newBlock.List, t.copyStmt(stmt))
		}
		return newBlock
	default:
		return stmt
	}
}

func (t *Transform) extractStatements(f *ast.File, startLine, endLine int) (extracted, remaining []ast.Stmt) {
	ast.Inspect(f, func(n ast.Node) bool {
		if fd, ok := n.(*ast.FuncDecl); ok {
			if fd.Body != nil {
				var newStmts []ast.Stmt
				var extractedBody []ast.Stmt

				for _, stmt := range fd.Body.List {
					stmtStart := t.fset.Position(stmt.Pos()).Line
					stmtEnd := t.fset.Position(stmt.End()).Line

					if stmtStart >= startLine && stmtEnd <= endLine {
						extractedBody = append(extractedBody, t.copyStmt(stmt))
					} else {
						newStmts = append(newStmts, stmt)
					}
				}

				if len(extractedBody) > 0 {
					fd.Body.List = newStmts
					extracted = append(extracted, extractedBody...)
					remaining = newStmts
					return false
				}
			}
		}
		return true
	})

	return
}

func (t *Transform) createExtractedFunc(name string, stmts []ast.Stmt) *ast.FuncDecl {
	return &ast.FuncDecl{
		Name: ast.NewIdent(name),
		Type: &ast.FuncType{
			Params:  &ast.FieldList{},
			Results: &ast.FieldList{},
		},
		Body: &ast.BlockStmt{
			List: stmts,
		},
	}
}

func (t *Transform) containsLine(fd *ast.FuncDecl, start, end int) bool {
	if fd == nil || fd.Body == nil {
		return false
	}
	for _, stmt := range fd.Body.List {
		stmtStart := t.fset.Position(stmt.Pos()).Line
		stmtEnd := t.fset.Position(stmt.End()).Line
		if stmtStart >= start && stmtEnd <= end {
			return true
		}
	}
	return false
}

func (t *Transform) replaceWithCall(fd *ast.FuncDecl, stmts []ast.Stmt, funcName string) []ast.Decl {
	if fd == nil || len(stmts) == 0 {
		return []ast.Decl{fd}
	}

	newFunc := &ast.FuncDecl{
		Recv: fd.Recv,
		Name: ast.NewIdent(fd.Name.Name),
		Type: fd.Type,
		Body: &ast.BlockStmt{},
	}

	callStmt := &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun:  ast.NewIdent(funcName),
			Args: []ast.Expr{},
		},
	}

	newFunc.Body.List = append([]ast.Stmt{callStmt}, stmts...)

	return []ast.Decl{newFunc}
}

func (t *Transform) containsStatementsInRange(gd *ast.GenDecl, start, end int) bool {
	for _, spec := range gd.Specs {
		if vs, ok := spec.(*ast.ValueSpec); ok {
			for _, name := range vs.Names {
				nameLine := t.fset.Position(name.Pos()).Line
				if nameLine >= start && nameLine <= end {
					return true
				}
			}
		}
	}
	return false
}

func (t *Transform) findDeclIndex(decls []ast.Decl, target ast.Decl) int {
	for i, decl := range decls {
		if decl == target {
			return i
		}
	}
	return -1
}

func (t *Transform) insertBefore(decls []ast.Decl, idx int, decl ast.Decl) []ast.Decl {
	if idx < 0 || idx > len(decls) {
		return append(decls, decl)
	}
	newDecls := make([]ast.Decl, 0, len(decls)+1)
	newDecls = append(newDecls, decls[:idx]...)
	newDecls = append(newDecls, decl)
	newDecls = append(newDecls, decls[idx:]...)
	return newDecls
}

type renamer struct {
	oldName, newName string
	scope            map[string]bool
}

func (r *renamer) Visit(node ast.Node) ast.Visitor {
	if r.scope == nil {
		r.scope = make(map[string]bool)
	}

	switch n := node.(type) {
	case *ast.File:
		r.visitDecls(n.Decls)
		return nil

	case *ast.Ident:
		if n.Name == r.oldName {
			if !r.isDeclaration(n) && !r.scope[n.Name] {
				n.Name = r.newName
			}
		}
		return nil

	case *ast.FuncDecl:
		if n.Name.Name == r.oldName {
			n.Name = ast.NewIdent(r.newName)
		}
		return r

	case *ast.GenDecl:
		return r

	case *ast.ValueSpec:
		for i, name := range n.Names {
			if name.Name == r.oldName {
				r.scope[name.Name] = true
				n.Names[i] = ast.NewIdent(r.newName)
			}
		}
		return r

	case *ast.TypeSpec:
		if n.Name.Name == r.oldName {
			r.scope[n.Name.Name] = true
			n.Name = ast.NewIdent(r.newName)
		}
		return r

	case *ast.SelectorExpr:
		if n.Sel.Name == r.oldName {
			if !r.isFieldSelector(n) {
				n.Sel = ast.NewIdent(r.newName)
			}
		}
		return nil

	case *ast.CompositeLit:
		for _, elt := range n.Elts {
			if kv, ok := elt.(*ast.KeyValueExpr); ok {
				if ident, ok := kv.Key.(*ast.Ident); ok && ident.Name == r.oldName {
					kv.Key = ast.NewIdent(r.newName)
				}
			}
		}
		return r

	default:
		return r
	}
}

func (r *renamer) visitDecls(decls []ast.Decl) {
	for _, decl := range decls {
		ast.Inspect(decl, func(node ast.Node) bool {
			if fd, ok := node.(*ast.FuncDecl); ok {
				r.visitFuncParams(fd)
				r.visitFuncResults(fd)
			}
			return true
		})
	}
}

func (r *renamer) visitFuncParams(fd *ast.FuncDecl) {
	if fd.Type.Params == nil {
		return
	}
	for _, field := range fd.Type.Params.List {
		for i, name := range field.Names {
			if name.Name == r.oldName {
				r.scope[name.Name] = true
				field.Names[i] = ast.NewIdent(r.newName)
			}
		}
	}
}

func (r *renamer) visitFuncResults(fd *ast.FuncDecl) {
	if fd.Type.Results == nil {
		return
	}
	for _, field := range fd.Type.Results.List {
		for i, name := range field.Names {
			if name.Name == r.oldName {
				r.scope[name.Name] = true
				field.Names[i] = ast.NewIdent(r.newName)
			}
		}
	}
}

func (r *renamer) isDeclaration(ident *ast.Ident) bool {
	return r.scope[ident.Name]
}

func (r *renamer) isFieldSelector(sel *ast.SelectorExpr) bool {
	if ident, ok := sel.X.(*ast.Ident); ok {
		return ident.Obj == nil || ident.Obj.Kind == ast.Var
	}
	return false
}

type inliner struct {
	varName    string
	varValue   ast.Expr
	removeStmt bool
}

func (i *inliner) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.AssignStmt:
		if i.isSimpleAssign(n) {
			if ident, ok := n.Lhs[0].(*ast.Ident); ok && ident.Name == i.varName {
				i.varValue = i.copyExpr(n.Rhs[0])
				i.removeStmt = true
				return nil
			}
		}
		return i

	case *ast.Ident:
		if i.removeStmt && n.Name == i.varName {
			return nil
		}
		if n.Name == i.varName && i.varValue != nil {
			n.Name = ""
			return nil
		}
		return nil

	default:
		return i
	}
}

func (i *inliner) isSimpleAssign(assign *ast.AssignStmt) bool {
	if len(assign.Lhs) != 1 || len(assign.Rhs) != 1 {
		return false
	}
	if assign.Tok != token.ASSIGN {
		return false
	}
	if _, ok := assign.Lhs[0].(*ast.Ident); !ok {
		return false
	}
	return i.isSimpleExpr(assign.Rhs[0])
}

func (i *inliner) copyExpr(expr ast.Expr) ast.Expr {
	switch e := expr.(type) {
	case *ast.Ident:
		return &ast.Ident{Name: e.Name}
	case *ast.BasicLit:
		return &ast.BasicLit{Kind: e.Kind, Value: e.Value}
	case *ast.UnaryExpr:
		return &ast.UnaryExpr{Op: e.Op, X: i.copyExpr(e.X)}
	default:
		return nil
	}
}

func (i *inliner) isSimpleExpr(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.Ident:
		return true
	case *ast.BasicLit:
		return true
	case *ast.UnaryExpr:
		return i.isSimpleExpr(e.X)
	default:
		return false
	}
}

func FindLineRange(f *ast.File, fset *token.FileSet, start, end int) []ast.Node {
	var nodes []ast.Node
	ast.Inspect(f, func(n ast.Node) bool {
		if n == nil {
			return false
		}
		pos := fset.Position(n.Pos())
		if pos.Line >= start && pos.Line <= end {
			nodes = append(nodes, n)
		}
		return true
	})
	return nodes
}

func FormatNode(fset *token.FileSet, node ast.Node) string {
	if node == nil {
		return ""
	}
	var sb strings.Builder
	pos := fset.Position(node.Pos())
	end := fset.Position(node.End())
	return fmt.Sprintf("%s:%d-%d", sb.String(), pos.Line, end.Line)
}
