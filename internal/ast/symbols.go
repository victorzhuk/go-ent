package ast

import (
	"fmt"
	"go/ast"
	"go/token"
)

type SymbolKind int

const (
	SymbolPackage SymbolKind = iota
	SymbolFunction
	SymbolType
	SymbolVariable
	SymbolConstant
	SymbolField
	SymbolMethod
)

type ReferenceKind int

const (
	RefDefinition ReferenceKind = iota
	RefRead
	RefWrite
)

func (k SymbolKind) String() string {
	switch k {
	case SymbolPackage:
		return "package"
	case SymbolFunction:
		return "function"
	case SymbolType:
		return "type"
	case SymbolVariable:
		return "variable"
	case SymbolConstant:
		return "constant"
	case SymbolField:
		return "field"
	case SymbolMethod:
		return "method"
	default:
		return "unknown"
	}
}

func (k ReferenceKind) String() string {
	switch k {
	case RefDefinition:
		return "definition"
	case RefRead:
		return "read"
	case RefWrite:
		return "write"
	default:
		return "unknown"
	}
}

type Symbol struct {
	Name  string
	Kind  SymbolKind
	Pos   token.Pos
	Scope *Scope
	Type  ast.Expr
}

type Scope struct {
	Parent   *Scope
	Children []*Scope
	Symbols  map[string]*Symbol
}

func newScope(parent *Scope) *Scope {
	s := &Scope{
		Parent:  parent,
		Symbols: make(map[string]*Symbol),
	}
	if parent != nil {
		parent.Children = append(parent.Children, s)
	}
	return s
}

func (s *Scope) Lookup(name string) *Symbol {
	if sym, ok := s.Symbols[name]; ok {
		return sym
	}
	if s.Parent != nil {
		return s.Parent.Lookup(name)
	}
	return nil
}

func (s *Scope) Insert(sym *Symbol) {
	if sym != nil {
		s.Symbols[sym.Name] = sym
	}
}

type Builder struct {
	fset    *token.FileSet
	root    *Scope
	current *Scope
	file    *ast.File
}

func NewBuilder(fset *token.FileSet) *Builder {
	return &Builder{
		fset: fset,
	}
}

func (b *Builder) BuildFile(f *ast.File) (*Scope, error) {
	if f == nil {
		return nil, fmt.Errorf("nil file")
	}

	b.file = f
	b.root = newScope(nil)
	b.current = b.root

	pkgSym := &Symbol{
		Name: f.Name.Name,
		Kind: SymbolPackage,
		Pos:  f.Name.Pos(),
		Type: nil,
	}
	b.current.Insert(pkgSym)

	for _, decl := range f.Decls {
		b.visitDecl(decl)
	}

	return b.root, nil
}

func (b *Builder) visitDecl(decl ast.Decl) {
	switch d := decl.(type) {
	case *ast.GenDecl:
		b.visitGenDecl(d)
	case *ast.FuncDecl:
		b.visitFuncDecl(d)
	}
}

func (b *Builder) visitGenDecl(decl *ast.GenDecl) {
	switch decl.Tok {
	case token.VAR, token.CONST:
		for _, spec := range decl.Specs {
			if vs, ok := spec.(*ast.ValueSpec); ok {
				kind := SymbolVariable
				if decl.Tok == token.CONST {
					kind = SymbolConstant
				}
				b.visitValueSpec(vs, kind)
			}
		}
	case token.TYPE:
		for _, spec := range decl.Specs {
			if ts, ok := spec.(*ast.TypeSpec); ok {
				b.visitTypeSpec(ts)
			}
		}
	}
}

func (b *Builder) visitValueSpec(spec *ast.ValueSpec, kind SymbolKind) {
	for i, name := range spec.Names {
		var typ ast.Expr
		if i < len(spec.Values) {
			typ = spec.Values[i]
		} else if spec.Type != nil {
			typ = spec.Type
		}

		sym := &Symbol{
			Name:  name.Name,
			Kind:  kind,
			Pos:   name.Pos(),
			Scope: b.current,
			Type:  typ,
		}
		b.current.Insert(sym)
	}
}

func (b *Builder) visitTypeSpec(spec *ast.TypeSpec) {
	sym := &Symbol{
		Name:  spec.Name.Name,
		Kind:  SymbolType,
		Pos:   spec.Name.Pos(),
		Scope: b.current,
		Type:  spec.Type,
	}
	b.current.Insert(sym)

	if st, ok := spec.Type.(*ast.StructType); ok {
		b.visitStructType(st)
	}

	if it, ok := spec.Type.(*ast.InterfaceType); ok {
		b.visitInterfaceType(it)
	}
}

func (b *Builder) visitStructType(st *ast.StructType) {
	if st.Fields == nil {
		return
	}

	for _, field := range st.Fields.List {
		kind := SymbolField
		for _, name := range field.Names {
			sym := &Symbol{
				Name:  name.Name,
				Kind:  kind,
				Pos:   name.Pos(),
				Scope: b.current,
				Type:  field.Type,
			}
			b.current.Insert(sym)
		}
	}
}

func (b *Builder) visitInterfaceType(it *ast.InterfaceType) {
	if it.Methods == nil {
		return
	}

	for _, method := range it.Methods.List {
		if ft, ok := method.Type.(*ast.FuncType); ok {
			for _, name := range method.Names {
				sym := &Symbol{
					Name:  name.Name,
					Kind:  SymbolMethod,
					Pos:   name.Pos(),
					Scope: b.current,
					Type:  ft,
				}
				b.current.Insert(sym)
			}
		}
	}
}

func (b *Builder) visitFuncDecl(decl *ast.FuncDecl) {
	kind := SymbolFunction
	if decl.Recv != nil {
		kind = SymbolMethod
	}

	sym := &Symbol{
		Name:  decl.Name.Name,
		Kind:  kind,
		Pos:   decl.Name.Pos(),
		Scope: b.current,
		Type:  decl.Type,
	}
	b.current.Insert(sym)

	oldScope := b.current
	funcScope := newScope(b.current)
	b.current = funcScope

	if decl.Type.Params != nil {
		b.visitFieldList(decl.Type.Params, SymbolVariable)
	}

	if decl.Type.Results != nil {
		b.visitFieldList(decl.Type.Results, SymbolVariable)
	}

	if decl.Body != nil {
		b.visitBlockStmt(decl.Body)
	}

	b.current = oldScope
}

func (b *Builder) visitFieldList(fl *ast.FieldList, kind SymbolKind) {
	if fl == nil {
		return
	}

	for _, field := range fl.List {
		for _, name := range field.Names {
			sym := &Symbol{
				Name:  name.Name,
				Kind:  kind,
				Pos:   name.Pos(),
				Scope: b.current,
				Type:  field.Type,
			}
			b.current.Insert(sym)
		}
	}
}

func (b *Builder) visitBlockStmt(block *ast.BlockStmt) {
	if block == nil {
		return
	}

	for _, stmt := range block.List {
		b.visitStmt(stmt)
	}
}

func (b *Builder) visitStmt(stmt ast.Stmt) {
	switch s := stmt.(type) {
	case *ast.DeclStmt:
		b.visitDecl(s.Decl)
	case *ast.AssignStmt:
		b.visitAssignStmt(s)
	case *ast.BlockStmt:
		oldScope := b.current
		blockScope := newScope(b.current)
		b.current = blockScope
		b.visitBlockStmt(s)
		b.current = oldScope
	case *ast.IfStmt:
		b.visitIfStmt(s)
	case *ast.ForStmt:
		b.visitForStmt(s)
	case *ast.RangeStmt:
		b.visitRangeStmt(s)
	}
}

func (b *Builder) visitAssignStmt(assign *ast.AssignStmt) {
	if assign.Tok == token.DEFINE {
		for _, lhs := range assign.Lhs {
			if ident, ok := lhs.(*ast.Ident); ok {
				sym := &Symbol{
					Name:  ident.Name,
					Kind:  SymbolVariable,
					Pos:   ident.Pos(),
					Scope: b.current,
					Type:  nil,
				}
				b.current.Insert(sym)
			}
		}
	}
}

func (b *Builder) visitIfStmt(ifStmt *ast.IfStmt) {
	oldScope := b.current
	ifScope := newScope(b.current)
	b.current = ifScope

	if ifStmt.Init != nil {
		b.visitStmt(ifStmt.Init)
	}

	if ifStmt.Body != nil {
		b.visitBlockStmt(ifStmt.Body)
	}

	b.current = oldScope

	if ifStmt.Else != nil {
		b.visitStmt(ifStmt.Else)
	}
}

func (b *Builder) visitForStmt(forStmt *ast.ForStmt) {
	oldScope := b.current
	forScope := newScope(b.current)
	b.current = forScope

	if forStmt.Init != nil {
		b.visitStmt(forStmt.Init)
	}

	if forStmt.Body != nil {
		b.visitBlockStmt(forStmt.Body)
	}

	b.current = oldScope
}

func (b *Builder) visitRangeStmt(rangeStmt *ast.RangeStmt) {
	if ident, ok := rangeStmt.Key.(*ast.Ident); ok && rangeStmt.Tok == token.DEFINE {
		sym := &Symbol{
			Name:  ident.Name,
			Kind:  SymbolVariable,
			Pos:   ident.Pos(),
			Scope: b.current,
			Type:  nil,
		}
		b.current.Insert(sym)
	}

	if value, ok := rangeStmt.Value.(*ast.Ident); ok && rangeStmt.Tok == token.DEFINE {
		sym := &Symbol{
			Name:  value.Name,
			Kind:  SymbolVariable,
			Pos:   value.Pos(),
			Scope: b.current,
			Type:  nil,
		}
		b.current.Insert(sym)
	}

	if rangeStmt.Body != nil {
		oldScope := b.current
		bodyScope := newScope(b.current)
		b.current = bodyScope
		b.visitBlockStmt(rangeStmt.Body)
		b.current = oldScope
	}
}

type Reference struct {
	Symbol *Symbol
	Pos    token.Pos
	Kind   ReferenceKind
}

type DefinitionResult struct {
	Name     string
	Kind     SymbolKind
	File     string
	Line     int
	Column   int
	Type     string
	Exported bool
}

func (b *Builder) FindDefinition(name string, pos token.Pos) *DefinitionResult {
	if b.fset == nil || b.root == nil {
		return nil
	}

	ident := b.findIdentifierAtPos(name, pos)
	if ident == nil {
		return nil
	}

	sym := b.FindSymbol(ident.Name, pos)
	if sym == nil {
		return nil
	}

	posData := b.fset.Position(sym.Pos)

	exported := false
	if sym.Kind == SymbolType || sym.Kind == SymbolFunction || sym.Kind == SymbolMethod || sym.Kind == SymbolConstant || sym.Kind == SymbolField {
		exported = isExported(sym.Name)
	}

	var typeInfo string
	if sym.Type != nil {
		typeInfo = formatExpr(sym.Type)
	}

	return &DefinitionResult{
		Name:     sym.Name,
		Kind:     sym.Kind,
		File:     posData.Filename,
		Line:     posData.Line,
		Column:   posData.Column,
		Type:     typeInfo,
		Exported: exported,
	}
}

func (b *Builder) findIdentifierAtPos(name string, pos token.Pos) *ast.Ident {
	if b.fset == nil || b.file == nil {
		return nil
	}

	var target *ast.Ident

	ast.Inspect(b.file, func(n ast.Node) bool {
		if target != nil {
			return false
		}

		if ident, ok := n.(*ast.Ident); ok {
			if ident.Name == name {
				if ident.Pos() <= pos && pos < ident.End() {
					target = ident
				}
			}
		}

		return target == nil
	})

	return target
}

func isExported(name string) bool {
	if len(name) == 0 {
		return false
	}
	return name[0] >= 'A' && name[0] <= 'Z'
}

func (b *Builder) FindSymbol(name string, pos token.Pos) *Symbol {
	if b.root == nil {
		return nil
	}

	var target *Symbol
	visited := make(map[*Scope]bool)

	var walkScope func(scope *Scope)
	walkScope = func(scope *Scope) {
		if scope == nil || visited[scope] {
			return
		}
		visited[scope] = true

		if sym := scope.Lookup(name); sym != nil {
			if sym.Pos <= pos && (target == nil || target.Pos < sym.Pos) {
				target = sym
			}
		}

		for _, child := range scope.Children {
			walkScope(child)
		}
	}

	walkScope(b.root)
	return target
}

func (b *Builder) FindReferences(name string, pos token.Pos, includeTests bool) []Reference {
	if b.root == nil || b.file == nil {
		return nil
	}

	target := b.FindSymbol(name, pos)
	if target == nil {
		return nil
	}

	var refs []Reference

	// First add the definition itself
	refs = append(refs, Reference{
		Symbol: target,
		Pos:    target.Pos,
		Kind:   RefDefinition,
	})

	// Walk the AST to find all usages of this symbol
	ast.Inspect(b.file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.Ident:
			if node.Name == target.Name && b.isReferenceToSymbol(node, target) {
				// Skip the definition itself since we already added it
				if node.Pos() != target.Pos {
					kind := b.categorizeReferenceByUsage(node)
					refs = append(refs, Reference{
						Symbol: target,
						Pos:    node.Pos(),
						Kind:   kind,
					})
				}
			}
		case *ast.SelectorExpr:
			if node.Sel != nil && node.Sel.Name == target.Name && b.isReferenceToSymbol(node.Sel, target) {
				kind := b.categorizeReferenceByUsage(node.Sel)
				refs = append(refs, Reference{
					Symbol: target,
					Pos:    node.Sel.Pos(),
					Kind:   kind,
				})
			}
		}
		return true
	})

	return refs
}

func (b *Builder) isReferenceToSymbol(ident *ast.Ident, target *Symbol) bool {
	if ident == nil || target == nil {
		return false
	}

	// Check if this identifier refers to our target symbol
	// We need to resolve the identifier to its definition and see if it matches our target
	sym := b.FindSymbol(ident.Name, ident.Pos())
	return sym == target
}

func (b *Builder) categorizeReferenceByUsage(ident *ast.Ident) ReferenceKind {
	if ident == nil {
		return RefRead
	}

	// For now, just return RefRead for all references
	// We can enhance this later to detect write vs read contexts
	return RefRead
}
