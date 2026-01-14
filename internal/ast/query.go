package ast

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

type Query struct {
	fset *token.FileSet
}

func NewQuery(fset *token.FileSet) *Query {
	return &Query{fset: fset}
}

type Result struct {
	File      string
	Line      int
	Name      string
	Signature string
	Type      string
}

func FindFunctions(f *ast.File, pattern string) []Result {
	if f == nil || pattern == "" {
		return nil
	}

	var results []Result
	v := &funcVisitor{
		pattern: pattern,
		collect: func(file string, line int, name, sig string) {
			results = append(results, Result{
				File:      file,
				Line:      line,
				Name:      name,
				Signature: sig,
				Type:      "function",
			})
		},
	}

	ast.Walk(v, f)
	return results
}

func (q *Query) FindFunctions(files map[string]*ast.File, pattern string) []Result {
	if files == nil || pattern == "" {
		return nil
	}

	var results []Result
	for filename, f := range files {
		fileResults := FindFunctions(f, pattern)
		for i := range fileResults {
			fileResults[i].File = filename
			fileResults[i].Line = q.line(f, fileResults[i].Name)
		}
		results = append(results, fileResults...)
	}
	return results
}

func FindImplementations(files map[string]*ast.File, interfaceName string) []Result {
	if files == nil || interfaceName == "" {
		return nil
	}

	iface := findInterface(files, interfaceName)
	if iface == nil {
		return nil
	}

	var results []Result
	for filename, f := range files {
		impls := findImplementationsInFile(f, iface)
		for i := range impls {
			impls[i].File = filename
		}
		results = append(results, impls...)
	}
	return results
}

func (q *Query) FindImplementations(files map[string]*ast.File, interfaceName string) []Result {
	if files == nil || interfaceName == "" {
		return nil
	}

	results := FindImplementations(files, interfaceName)
	for i := range results {
		if f, ok := files[results[i].File]; ok {
			results[i].Line = q.line(f, results[i].Name)
		}
	}
	return results
}

func FindBySignature(f *ast.File, sig string) []Result {
	if f == nil || sig == "" {
		return nil
	}

	normalizedSig := normalizeSignature(sig)
	var results []Result

	v := &signatureVisitor{
		signature: normalizedSig,
		collect: func(file string, line int, name, sig string) {
			results = append(results, Result{
				File:      file,
				Line:      line,
				Name:      name,
				Signature: sig,
				Type:      "function",
			})
		},
	}

	ast.Walk(v, f)
	return results
}

func (q *Query) FindBySignature(files map[string]*ast.File, sig string) []Result {
	if files == nil || sig == "" {
		return nil
	}

	var results []Result
	for filename, f := range files {
		fileResults := FindBySignature(f, sig)
		for i := range fileResults {
			fileResults[i].File = filename
			fileResults[i].Line = q.line(f, fileResults[i].Name)
		}
		results = append(results, fileResults...)
	}
	return results
}

func (q *Query) line(f *ast.File, name string) int {
	var line int
	ast.Inspect(f, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok && fn.Name.Name == name {
			line = q.fset.Position(fn.Pos()).Line
			return false
		}
		if ts, ok := n.(*ast.TypeSpec); ok && ts.Name.Name == name {
			line = q.fset.Position(ts.Pos()).Line
			return false
		}
		return true
	})
	return line
}

type funcVisitor struct {
	pattern string
	collect func(file string, line int, name, sig string)
}

func (v *funcVisitor) Visit(node ast.Node) ast.Visitor {
	if fn, ok := node.(*ast.FuncDecl); ok {
		if v.match(fn.Name.Name) {
			sig := formatSignature(fn)
			v.collect("", 0, fn.Name.Name, sig)
		}
		return nil
	}
	return v
}

func (v *funcVisitor) match(name string) bool {
	if v.pattern == "*" || v.pattern == "" {
		return true
	}
	if strings.HasSuffix(v.pattern, "*") {
		prefix := strings.TrimSuffix(v.pattern, "*")
		return strings.HasPrefix(name, prefix)
	}
	return strings.EqualFold(name, v.pattern)
}

type signatureVisitor struct {
	signature string
	collect   func(file string, line int, name, sig string)
}

func (v *signatureVisitor) Visit(node ast.Node) ast.Visitor {
	if fn, ok := node.(*ast.FuncDecl); ok {
		sig := formatSignature(fn)
		normalizedSig := normalizeSignature(sig)
		if normalizedSig == v.signature {
			v.collect("", 0, fn.Name.Name, sig)
		}
		return nil
	}
	return v
}

func findInterface(files map[string]*ast.File, name string) *ast.InterfaceType {
	var iface *ast.InterfaceType

	for _, f := range files {
		ast.Inspect(f, func(n ast.Node) bool {
			if ts, ok := n.(*ast.TypeSpec); ok && ts.Name.Name == name {
				if it, ok := ts.Type.(*ast.InterfaceType); ok {
					iface = it
					return false
				}
			}
			return true
		})

		if iface != nil {
			break
		}
	}

	return iface
}

func findImplementationsInFile(f *ast.File, iface *ast.InterfaceType) []Result {
	var results []Result

	methods := ifaceMethods(iface)
	if len(methods) == 0 {
		return results
	}

	ast.Inspect(f, func(n ast.Node) bool {
		ts, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		implMethods := findTypeMethods(f, ts.Name.Name)
		if implementsAllMethods(implMethods, methods) {
			results = append(results, Result{
				File:      "",
				Line:      0,
				Name:      ts.Name.Name,
				Signature: typeSignature(ts),
				Type:      "implementation",
			})
		}

		return false
	})

	return results
}

func ifaceMethods(iface *ast.InterfaceType) map[string]*ast.FuncType {
	methods := make(map[string]*ast.FuncType)

	for _, m := range iface.Methods.List {
		if ft, ok := m.Type.(*ast.FuncType); ok {
			for _, name := range m.Names {
				methods[name.Name] = ft
			}
		}
	}

	return methods
}

func signatureMatch(a, b *ast.FuncType) bool {
	if a == nil || b == nil {
		return a == b
	}

	if len(a.Params.List) != len(b.Params.List) {
		return false
	}

	if len(a.Results.List) != len(b.Results.List) {
		return false
	}

	for i := range a.Params.List {
		if !exprMatch(a.Params.List[i].Type, b.Params.List[i].Type) {
			return false
		}
	}

	return true
}

func findTypeMethods(f *ast.File, typeName string) map[string]*ast.FuncType {
	methods := make(map[string]*ast.FuncType)

	ast.Inspect(f, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok && fn.Recv != nil && len(fn.Recv.List) > 0 {
			recvType := receiverTypeName(fn.Recv.List[0].Type)
			if recvType == typeName || recvType == "*"+typeName {
				methods[fn.Name.Name] = fn.Type
			}
		}
		return true
	})

	return methods
}

func receiverTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + receiverTypeName(t.X)
	case *ast.SelectorExpr:
		return receiverTypeName(t.X) + "." + t.Sel.Name
	case *ast.IndexExpr:
		return receiverTypeName(t.X)
	case *ast.IndexListExpr:
		return receiverTypeName(t.X)
	default:
		return ""
	}
}

func implementsAllMethods(implMethods, requiredMethods map[string]*ast.FuncType) bool {
	if len(implMethods) < len(requiredMethods) {
		return false
	}

	for name, sig := range requiredMethods {
		if !signatureMatch(implMethods[name], sig) {
			return false
		}
	}

	return true
}

func exprMatch(a, b ast.Expr) bool {
	return formatExpr(a) == formatExpr(b)
}

func formatSignature(fn *ast.FuncDecl) string {
	return formatFuncType(fn.Type)
}

func formatFuncType(ft *ast.FuncType) string {
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

func formatExpr(e ast.Expr) string {
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
	default:
		return "any"
	}
}

func formatChanType(ct *ast.ChanType) string {
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

func typeSignature(ts *ast.TypeSpec) string {
	return formatExpr(ts.Type)
}

func normalizeSignature(sig string) string {
	sig = strings.TrimSpace(sig)
	sig = strings.ReplaceAll(sig, " ", "")
	sig = strings.ReplaceAll(sig, "\t", "")
	sig = strings.ReplaceAll(sig, "\n", "")
	return sig
}
