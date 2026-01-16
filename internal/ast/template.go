package ast

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

type Template struct {
	Name   string
	Source string
	AST    ast.Node
	Params map[string]bool
}

type Engine struct {
	fset      *token.FileSet
	templates map[string]*Template
}

func NewEngine(fset *token.FileSet) *Engine {
	if fset == nil {
		fset = token.NewFileSet()
	}
	return &Engine{
		fset:      fset,
		templates: make(map[string]*Template),
	}
}

func (e *Engine) Register(tmpl *Template) error {
	if tmpl == nil {
		return fmt.Errorf("nil template")
	}
	if tmpl.Name == "" {
		return fmt.Errorf("empty template name")
	}
	if tmpl.Source == "" {
		return fmt.Errorf("empty template source")
	}

	src := tmpl.Source
	if !hasPackageDecl(src) {
		src = "package template\n" + src
	}

	f, err := parser.ParseFile(e.fset, "", src, parser.AllErrors)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	if len(f.Decls) == 0 {
		return fmt.Errorf("no declarations")
	}

	tmpl.AST = f
	tmpl.Params = extractParamNames(tmpl.Source)

	e.templates[tmpl.Name] = tmpl
	return nil
}

func hasPackageDecl(src string) bool {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, parser.PackageClauseOnly)
	if err != nil {
		return false
	}
	return f.Name != nil && f.Name.Name != ""
}

func (e *Engine) Get(name string) (*Template, error) {
	tmpl, ok := e.templates[name]
	if !ok {
		return nil, fmt.Errorf("template not found: %s", name)
	}
	return tmpl, nil
}

func (e *Engine) Execute(name string, data map[string]string) (ast.Node, error) {
	tmpl, err := e.Get(name)
	if err != nil {
		return nil, err
	}
	return e.ExecuteTemplate(tmpl, data)
}

func (e *Engine) ExecuteTemplate(tmpl *Template, data map[string]string) (ast.Node, error) {
	if tmpl == nil {
		return nil, fmt.Errorf("nil template")
	}

	copier := &nodeCopier{}
	copied := copier.copy(tmpl.AST)
	if copied == nil {
		return nil, fmt.Errorf("copy template")
	}

	replacer := &placeholderReplacer{data: data}
	ast.Walk(replacer, copied)

	f, ok := copied.(*ast.File)
	if ok && len(f.Decls) == 1 {
		return f.Decls[0], nil
	}
	return copied, nil
}

func (e *Engine) RegisterBuiltIns() error {
	builtins := []*Template{
		{
			Name: "struct",
			Source: `type TypeName struct {
				FieldName FieldType
			}`,
		},
		{
			Name: "function",
			Source: `func FunctionName(param ParamType) ReturnType {
				return DefaultValue
			}`,
		},
		{
			Name: "method",
			Source: `func (r *TypeName) MethodName(param ParamType) ReturnType {
				return DefaultValue
			}`,
		},
		{
			Name: "interface",
			Source: `type TypeName interface {
				MethodName(param ParamType) ReturnType
			}`,
		},
	}

	for _, tmpl := range builtins {
		if err := e.Register(tmpl); err != nil {
			return fmt.Errorf("register builtin %s: %w", tmpl.Name, err)
		}
	}
	return nil
}

type placeholderReplacer struct {
	data map[string]string
}

func (r *placeholderReplacer) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}

	switch n := node.(type) {
	case *ast.Ident:
		if r.isPlaceholder(n.Name) {
			if replacement := r.getReplacement(n.Name); replacement != "" {
				n.Name = replacement
			}
		}
	case *ast.GenDecl:
		if n.Tok == token.IMPORT && len(n.Specs) > 0 {
			spec := n.Specs[0].(*ast.ImportSpec)
			if spec.Path != nil && r.isPlaceholder(spec.Path.Value) {
				if replacement := r.getReplacement(spec.Path.Value); replacement != "" {
					spec.Path.Value = replacement
				}
			}
		}
	}

	return r
}

func (r *placeholderReplacer) isPlaceholder(name string) bool {
	if r.data == nil {
		return false
	}
	_, ok := r.data[name]
	return ok
}

func (r *placeholderReplacer) getReplacement(name string) string {
	if r.data == nil {
		return ""
	}
	return r.data[name]
}

type nodeCopier struct{}

func (c *nodeCopier) copy(node ast.Node) ast.Node {
	if node == nil {
		return nil
	}

	switch n := node.(type) {
	case *ast.File:
		return c.copyFile(n)
	case *ast.GenDecl:
		return c.copyDecl(n)
	case *ast.FuncDecl:
		return c.copyFuncDecl(n)
	case *ast.DeclStmt:
		return &ast.DeclStmt{Decl: c.copyDecl(n.Decl.(*ast.GenDecl))}
	case *ast.BlockStmt:
		return c.copyBlockStmt(n)
	case *ast.ReturnStmt:
		return &ast.ReturnStmt{Results: c.copyExprs(n.Results)}
	case *ast.AssignStmt:
		return &ast.AssignStmt{
			Lhs: c.copyExprs(n.Lhs),
			Tok: n.Tok,
			Rhs: c.copyExprs(n.Rhs),
		}
	case *ast.IfStmt:
		return c.copyIfStmt(n)
	case *ast.ForStmt:
		return c.copyForStmt(n)
	case *ast.ExprStmt:
		return &ast.ExprStmt{X: c.copyExpr(n.X)}
	case *ast.CallExpr:
		return c.copyCallExpr(n)
	case *ast.SelectorExpr:
		return c.copySelectorExpr(n)
	case *ast.Ident:
		return c.copyIdent(n)
	case *ast.BasicLit:
		return c.copyBasicLit(n)
	case *ast.CompositeLit:
		return c.copyCompositeLit(n)
	case *ast.UnaryExpr:
		return &ast.UnaryExpr{Op: n.Op, X: c.copyExpr(n.X)}
	case *ast.BinaryExpr:
		return &ast.BinaryExpr{
			X:  c.copyExpr(n.X),
			Op: n.Op,
			Y:  c.copyExpr(n.Y),
		}
	case *ast.ValueSpec:
		return c.copyValueSpec(n)
	case *ast.TypeSpec:
		return c.copyTypeSpec(n)
	case *ast.StructType:
		return c.copyStructType(n)
	case *ast.InterfaceType:
		return c.copyInterfaceType(n)
	case *ast.FuncType:
		return c.copyFuncType(n)
	case *ast.FieldList:
		return c.copyFieldList(n)
	case *ast.Field:
		return c.copyField(n)
	case *ast.ArrayType:
		return &ast.ArrayType{
			Elt: c.copyExpr(n.Elt),
			Len: c.copyExpr(n.Len),
		}
	case *ast.MapType:
		return &ast.MapType{
			Key:   c.copyExpr(n.Key),
			Value: c.copyExpr(n.Value),
		}
	case *ast.StarExpr:
		return &ast.StarExpr{X: c.copyExpr(n.X)}
	case *ast.Ellipsis:
		return &ast.Ellipsis{Elt: c.copyExpr(n.Elt)}
	case *ast.KeyValueExpr:
		return &ast.KeyValueExpr{
			Key:   c.copyExpr(n.Key),
			Value: c.copyExpr(n.Value),
		}
	default:
		return node
	}
}

func (c *nodeCopier) copyFile(f *ast.File) *ast.File {
	if f == nil {
		return nil
	}

	decls := make([]ast.Decl, len(f.Decls))
	for i, decl := range f.Decls {
		decls[i] = c.copyDecl(decl)
	}

	imports := make([]*ast.ImportSpec, len(f.Imports))
	for i, imp := range f.Imports {
		imports[i] = &ast.ImportSpec{
			Path: &ast.BasicLit{Value: imp.Path.Value},
		}
	}

	return &ast.File{
		Name:  &ast.Ident{Name: f.Name.Name},
		Decls: decls,
	}
}

func (c *nodeCopier) copyDecl(d ast.Decl) ast.Decl {
	if d == nil {
		return nil
	}

	switch n := d.(type) {
	case *ast.GenDecl:
		specs := make([]ast.Spec, len(n.Specs))
		for i, spec := range n.Specs {
			specs[i] = c.copySpec(spec)
		}
		return &ast.GenDecl{
			TokPos: n.TokPos,
			Tok:    n.Tok,
			Specs:  specs,
		}
	case *ast.FuncDecl:
		return c.copyFuncDecl(n)
	default:
		return d
	}
}

func (c *nodeCopier) copySpec(s ast.Spec) ast.Spec {
	switch n := s.(type) {
	case *ast.TypeSpec:
		return c.copyTypeSpec(n)
	case *ast.ValueSpec:
		return c.copyValueSpec(n)
	case *ast.ImportSpec:
		return &ast.ImportSpec{
			Path: &ast.BasicLit{Value: n.Path.Value},
		}
	default:
		return s
	}
}

func (c *nodeCopier) copyFuncDecl(fn *ast.FuncDecl) *ast.FuncDecl {
	if fn == nil {
		return nil
	}

	return &ast.FuncDecl{
		Recv: c.copyFieldList(fn.Recv),
		Name: c.copyIdent(fn.Name),
		Type: c.copyFuncType(fn.Type),
		Body: c.copyBlockStmt(fn.Body),
	}
}

func (c *nodeCopier) copyBlockStmt(b *ast.BlockStmt) *ast.BlockStmt {
	if b == nil {
		return nil
	}

	stmts := make([]ast.Stmt, len(b.List))
	for i, stmt := range b.List {
		stmts[i] = c.copyStmt(stmt)
	}

	return &ast.BlockStmt{List: stmts}
}

func (c *nodeCopier) copyStmt(s ast.Stmt) ast.Stmt {
	if s == nil {
		return nil
	}

	switch n := s.(type) {
	case *ast.DeclStmt:
		return &ast.DeclStmt{Decl: c.copyDecl(n.Decl.(*ast.GenDecl))}
	case *ast.ReturnStmt:
		return &ast.ReturnStmt{Results: c.copyExprs(n.Results)}
	case *ast.AssignStmt:
		return &ast.AssignStmt{
			Lhs: c.copyExprs(n.Lhs),
			Tok: n.Tok,
			Rhs: c.copyExprs(n.Rhs),
		}
	case *ast.IfStmt:
		return c.copyIfStmt(n)
	case *ast.ForStmt:
		return c.copyForStmt(n)
	case *ast.ExprStmt:
		return &ast.ExprStmt{X: c.copyExpr(n.X)}
	case *ast.BlockStmt:
		return c.copyBlockStmt(n)
	default:
		return s
	}
}

func (c *nodeCopier) copyIfStmt(n *ast.IfStmt) *ast.IfStmt {
	return &ast.IfStmt{
		Init: c.copyStmt(n.Init),
		Cond: c.copyExpr(n.Cond),
		Body: c.copyBlockStmt(n.Body),
		Else: c.copyStmt(n.Else),
	}
}

func (c *nodeCopier) copyForStmt(n *ast.ForStmt) *ast.ForStmt {
	return &ast.ForStmt{
		Init: c.copyStmt(n.Init),
		Cond: c.copyExpr(n.Cond),
		Post: c.copyStmt(n.Post),
		Body: c.copyBlockStmt(n.Body),
	}
}

func (c *nodeCopier) copyExpr(e ast.Expr) ast.Expr {
	if e == nil {
		return nil
	}

	switch n := e.(type) {
	case *ast.Ident:
		return c.copyIdent(n)
	case *ast.BasicLit:
		return c.copyBasicLit(n)
	case *ast.CallExpr:
		return c.copyCallExpr(n)
	case *ast.SelectorExpr:
		return c.copySelectorExpr(n)
	case *ast.CompositeLit:
		return c.copyCompositeLit(n)
	case *ast.UnaryExpr:
		return &ast.UnaryExpr{Op: n.Op, X: c.copyExpr(n.X)}
	case *ast.BinaryExpr:
		return &ast.BinaryExpr{
			X:  c.copyExpr(n.X),
			Op: n.Op,
			Y:  c.copyExpr(n.Y),
		}
	case *ast.ParenExpr:
		return &ast.ParenExpr{X: c.copyExpr(n.X)}
	case *ast.StarExpr:
		return &ast.StarExpr{X: c.copyExpr(n.X)}
	case *ast.ArrayType:
		return &ast.ArrayType{Elt: c.copyExpr(n.Elt), Len: c.copyExpr(n.Len)}
	case *ast.MapType:
		return &ast.MapType{Key: c.copyExpr(n.Key), Value: c.copyExpr(n.Value)}
	case *ast.Ellipsis:
		return &ast.Ellipsis{Elt: c.copyExpr(n.Elt)}
	case *ast.KeyValueExpr:
		return &ast.KeyValueExpr{
			Key:   c.copyExpr(n.Key),
			Value: c.copyExpr(n.Value),
		}
	default:
		return e
	}
}

func (c *nodeCopier) copyExprs(exprs []ast.Expr) []ast.Expr {
	if exprs == nil {
		return nil
	}
	result := make([]ast.Expr, len(exprs))
	for i, e := range exprs {
		result[i] = c.copyExpr(e)
	}
	return result
}

func (c *nodeCopier) copyCallExpr(n *ast.CallExpr) *ast.CallExpr {
	return &ast.CallExpr{
		Fun:      c.copyExpr(n.Fun),
		Args:     c.copyExprs(n.Args),
		Ellipsis: n.Ellipsis,
	}
}

func (c *nodeCopier) copySelectorExpr(n *ast.SelectorExpr) *ast.SelectorExpr {
	return &ast.SelectorExpr{
		X:   c.copyExpr(n.X),
		Sel: c.copyIdent(n.Sel),
	}
}

func (c *nodeCopier) copyCompositeLit(n *ast.CompositeLit) *ast.CompositeLit {
	elts := make([]ast.Expr, len(n.Elts))
	for i, elt := range n.Elts {
		elts[i] = c.copyExpr(elt)
	}

	return &ast.CompositeLit{
		Type: c.copyExpr(n.Type),
		Elts: elts,
	}
}

func (c *nodeCopier) copyIdent(n *ast.Ident) *ast.Ident {
	if n == nil {
		return nil
	}
	return &ast.Ident{Name: n.Name}
}

func (c *nodeCopier) copyBasicLit(n *ast.BasicLit) *ast.BasicLit {
	if n == nil {
		return nil
	}
	return &ast.BasicLit{Value: n.Value, Kind: n.Kind}
}

func (c *nodeCopier) copyValueSpec(n *ast.ValueSpec) *ast.ValueSpec {
	if n == nil {
		return nil
	}

	names := make([]*ast.Ident, len(n.Names))
	for i, name := range n.Names {
		names[i] = c.copyIdent(name)
	}

	values := make([]ast.Expr, len(n.Values))
	for i, val := range n.Values {
		values[i] = c.copyExpr(val)
	}

	return &ast.ValueSpec{
		Names:  names,
		Type:   c.copyExpr(n.Type),
		Values: values,
	}
}

func (c *nodeCopier) copyTypeSpec(n *ast.TypeSpec) *ast.TypeSpec {
	if n == nil {
		return nil
	}

	return &ast.TypeSpec{
		Name: c.copyIdent(n.Name),
		Type: c.copyExpr(n.Type),
	}
}

func (c *nodeCopier) copyStructType(n *ast.StructType) *ast.StructType {
	if n == nil {
		return nil
	}

	return &ast.StructType{
		Fields: c.copyFieldList(n.Fields),
	}
}

func (c *nodeCopier) copyInterfaceType(n *ast.InterfaceType) *ast.InterfaceType {
	if n == nil {
		return nil
	}

	return &ast.InterfaceType{
		Methods: c.copyFieldList(n.Methods),
	}
}

func (c *nodeCopier) copyFuncType(n *ast.FuncType) *ast.FuncType {
	if n == nil {
		return nil
	}

	return &ast.FuncType{
		Params:  c.copyFieldList(n.Params),
		Results: c.copyFieldList(n.Results),
	}
}

func (c *nodeCopier) copyFieldList(n *ast.FieldList) *ast.FieldList {
	if n == nil {
		return nil
	}

	if len(n.List) == 0 {
		return nil
	}

	fields := make([]*ast.Field, len(n.List))
	for i, field := range n.List {
		fields[i] = c.copyField(field)
	}

	return &ast.FieldList{List: fields}
}

func (c *nodeCopier) copyField(n *ast.Field) *ast.Field {
	if n == nil {
		return nil
	}

	names := make([]*ast.Ident, len(n.Names))
	for i, name := range n.Names {
		names[i] = c.copyIdent(name)
	}

	return &ast.Field{
		Names: names,
		Type:  c.copyExpr(n.Type),
	}
}

func extractParamNames(src string) map[string]bool {
	fset := token.NewFileSet()

	parsedSrc := src
	var packageName string
	if !hasPackageDecl(src) {
		parsedSrc = "package template\n" + src
		packageName = "template"
	}

	f, err := parser.ParseFile(fset, "", parsedSrc, parser.AllErrors|parser.ParseComments)
	if err != nil {
		return make(map[string]bool)
	}

	params := make(map[string]bool)

	var visitor paramExtractor
	ast.Walk(&visitor, f)

	for _, name := range visitor.names {
		if name != packageName && isPlaceholderName(name) {
			params[name] = true
		}
	}

	return params
}

func isPlaceholderName(name string) bool {
	placeholderPatterns := []string{
		"TypeName",
		"FieldName",
		"FieldType",
		"FunctionName",
		"MethodName",
		"ReturnType",
		"ParamType",
		"Param",
		"VarName",
		"InterfaceName",
		"StructName",
		"ImplTypeName",
		"ReceiverName",
		"Element",
		"Key",
		"Value",
		"ErrorType",
	}

	for _, pattern := range placeholderPatterns {
		if name == pattern {
			return true
		}
	}
	return false
}

type paramExtractor struct {
	names []string
}

func (p *paramExtractor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}

	if ident, ok := node.(*ast.Ident); ok {
		if !isKeyword(ident.Name) && !isPredeclared(ident.Name) {
			if !contains(p.names, ident.Name) {
				p.names = append(p.names, ident.Name)
			}
		}
		return p
	}

	return p
}

func isKeyword(name string) bool {
	keywords := []string{
		"break", "case", "chan", "const", "continue", "default",
		"defer", "else", "fallthrough", "for", "func", "go",
		"goto", "if", "import", "interface", "map", "package",
		"range", "return", "select", "struct", "switch", "type",
		"var",
	}
	for _, kw := range keywords {
		if name == kw {
			return true
		}
	}
	return false
}

func isPredeclared(name string) bool {
	predeclared := []string{
		"bool", "byte", "complex64", "complex128", "error", "float32",
		"float64", "int", "int8", "int16", "int32", "int64",
		"rune", "string", "uint", "uint8", "uint16", "uint32",
		"uint64", "uintptr", "true", "false", "iota", "nil",
		"append", "cap", "close", "complex", "copy", "delete",
		"imag", "len", "make", "new", "panic", "print",
		"println", "real", "recover",
	}
	for _, pd := range predeclared {
		if name == pd {
			return true
		}
	}
	return false
}

func contains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func GenerateInterfaceImplementation(
	iface *ast.InterfaceType,
	implTypeName string,
	implMethods map[string]*ast.FuncType,
) (ast.Node, error) {
	if iface == nil {
		return nil, fmt.Errorf("nil interface")
	}
	if implTypeName == "" {
		return nil, fmt.Errorf("empty implementation type name")
	}
	if iface.Methods == nil || len(iface.Methods.List) == 0 {
		return nil, fmt.Errorf("interface has no methods")
	}

	var decls []ast.Decl

	structDecl := &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: &ast.Ident{Name: implTypeName},
				Type: &ast.StructType{
					Fields: &ast.FieldList{},
				},
			},
		},
	}
	decls = append(decls, structDecl)

	receiver := strings.ToLower(string(implTypeName[0]))

	for _, method := range iface.Methods.List {
		if len(method.Names) == 0 {
			continue
		}

		methodName := method.Names[0].Name
		methodType, ok := method.Type.(*ast.FuncType)
		if !ok {
			continue
		}

		body := generateMethodBody(methodType)

		methodDecl := &ast.FuncDecl{
			Recv: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{{Name: receiver}},
						Type:  &ast.StarExpr{X: &ast.Ident{Name: implTypeName}},
					},
				},
			},
			Name: &ast.Ident{Name: methodName},
			Type: methodType,
			Body: body,
		}

		decls = append(decls, methodDecl)
	}

	return &ast.File{
		Name:  &ast.Ident{Name: "main"},
		Decls: decls,
	}, nil
}

func generateMethodBody(ft *ast.FuncType) *ast.BlockStmt {
	stmts := []ast.Stmt{}

	if ft.Results != nil && len(ft.Results.List) > 0 {
		returnExprs := make([]ast.Expr, len(ft.Results.List))
		for i, result := range ft.Results.List {
			returnExprs[i] = generateZeroValue(result.Type)
		}

		stmts = append(stmts, &ast.ReturnStmt{
			Results: returnExprs,
		})
	} else {
		stmts = append(stmts, &ast.ReturnStmt{})
	}

	return &ast.BlockStmt{
		List: stmts,
	}
}

func generateZeroValue(expr ast.Expr) ast.Expr {
	switch t := expr.(type) {
	case *ast.Ident:
		switch t.Name {
		case "string":
			return &ast.BasicLit{Kind: token.STRING, Value: `""`}
		case "int", "int8", "int16", "int32", "int64",
			"uint", "uint8", "uint16", "uint32", "uint64", "uintptr",
			"float32", "float64", "complex64", "complex128", "rune", "byte":
			return &ast.BasicLit{Kind: token.INT, Value: "0"}
		case "bool":
			return &ast.Ident{Name: "false"}
		default:
			return &ast.Ident{Name: "nil"}
		}
	case *ast.ArrayType, *ast.MapType, *ast.ChanType, *ast.FuncType, *ast.InterfaceType:
		return &ast.Ident{Name: "nil"}
	case *ast.StarExpr:
		return &ast.Ident{Name: "nil"}
	case *ast.SelectorExpr:
		if ident, ok := t.X.(*ast.Ident); ok && ident.Name == "error" {
			return &ast.Ident{Name: "nil"}
		}
		return &ast.Ident{Name: "nil"}
	default:
		return &ast.Ident{Name: "nil"}
	}
}
