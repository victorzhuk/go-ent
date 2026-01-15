package ast

import (
	"go/ast"
	"go/token"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenameSymbolAtPos_SimpleFunction(t *testing.T) {
	t.Parallel()

	src := `package main

func hello() string {
	return "hello"
}

func main() {
	msg := hello()
	_ = msg
}
`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	transform := NewTransform(p.fset)

	var callPos token.Pos
	ast.Inspect(f, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if ident, ok := call.Fun.(*ast.Ident); ok && ident.Name == "hello" {
				callPos = ident.Pos()
				return false
			}
		}
		return true
	})

	require.NotZero(t, callPos)

	newFile, err := transform.RenameSymbolAtPos(f, callPos, "greet")
	require.NoError(t, err)
	require.NotNil(t, newFile)

	printer := NewPrinter(p.fset)
	result, err := printer.PrintFile(newFile)
	require.NoError(t, err)

	assert.Contains(t, result, "func greet()")
	assert.Contains(t, result, "msg := greet()")
	assert.NotContains(t, result, "hello()")
}

func TestRenameSymbolAtPos_TypeAware(t *testing.T) {
	t.Parallel()

	src := `package main

func hello() string {
	return "hello"
}

func main() {
	hello := "world"
	_ = hello
}
`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	transform := NewTransform(p.fset)

	var varPos token.Pos
	ast.Inspect(f, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok && ident.Name == "hello" {
			if ident.Obj != nil && ident.Obj.Kind == ast.Var {
				varPos = ident.Pos()
				return false
			}
		}
		return true
	})

	require.NotZero(t, varPos)

	newFile, err := transform.RenameSymbolAtPos(f, varPos, "greeting")
	require.NoError(t, err)
	require.NotNil(t, newFile)

	printer := NewPrinter(p.fset)
	result, err := printer.PrintFile(newFile)
	require.NoError(t, err)

	assert.Contains(t, result, "func hello()")
	assert.Contains(t, result, `greeting := "world"`)
	assert.NotContains(t, result, `hello := "world"`)
}

func TestRenameSymbolAtPos_NoShadowing(t *testing.T) {
	t.Parallel()

	src := `package main

func main() {
	x := 1
	_ = x

	if true {
		x := 2
		_ = x
	}
}
`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	transform := NewTransform(p.fset)

	var outerVarPos token.Pos
	count := 0
	ast.Inspect(f, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok && ident.Name == "x" {
			count++
			if count == 2 {
				outerVarPos = ident.Pos()
				return false
			}
		}
		return true
	})

	require.NotZero(t, outerVarPos)

	newFile, err := transform.RenameSymbolAtPos(f, outerVarPos, "value")
	require.NoError(t, err)
	require.NotNil(t, newFile)

	printer := NewPrinter(p.fset)
	result, err := printer.PrintFile(newFile)
	require.NoError(t, err)

	t.Logf("Result:\n%s\n", result)

	lines := strings.Split(result, "\n")

	ifIndex := -1
	for i, line := range lines {
		if strings.Contains(line, "if true {") {
			ifIndex = i
			break
		}
	}

	beforeIf := strings.Join(lines[:ifIndex], "\n")
	afterIf := strings.Join(lines[ifIndex:], "\n")

	assert.Contains(t, result, "value := 1")
	assert.Contains(t, beforeIf, "_ = value")
	assert.NotContains(t, beforeIf, "_ = x")
	assert.Contains(t, result, "if true {")
	assert.Contains(t, afterIf, "x := 2")
	assert.Contains(t, afterIf, "_ = x")
}

func TestRenameSymbolAtPos_StructField(t *testing.T) {
	t.Parallel()

	src := `package main

type User struct {
	Name string
}

func (u *User) GetName() string {
	return u.Name
}

func main() {
	u := User{Name: "Alice"}
	_ = u.Name
}
`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	transform := NewTransform(p.fset)

	var fieldPos token.Pos
	ast.Inspect(f, func(n ast.Node) bool {
		if sel, ok := n.(*ast.SelectorExpr); ok {
			if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "u" && sel.Sel.Name == "Name" {
				fieldPos = sel.Sel.Pos()
				return false
			}
		}
		return true
	})

	require.NotZero(t, fieldPos)

	newFile, err := transform.RenameSymbolAtPos(f, fieldPos, "Username")
	require.NoError(t, err)
	require.NotNil(t, newFile)

	printer := NewPrinter(p.fset)
	result, err := printer.PrintFile(newFile)
	require.NoError(t, err)

	t.Logf("Result:\n%s\n", result)

	assert.Contains(t, result, `Username string`)
	assert.Contains(t, result, "return u.Username")
	assert.Contains(t, result, `User{Username: "Alice"}`)
	assert.Contains(t, result, "_ = u.Username")
	assert.NotContains(t, result, "Name string")
	assert.NotContains(t, result, `User{Name: "Alice"}`)
}

func TestRenameSymbolAtPos_ExportedFunction(t *testing.T) {
	t.Parallel()

	src := `package main

func Hello() string {
	return "hello"
}

func main() {
	_ = Hello()
}
`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	transform := NewTransform(p.fset)

	var callPos token.Pos
	ast.Inspect(f, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if ident, ok := call.Fun.(*ast.Ident); ok && ident.Name == "Hello" {
				callPos = ident.Pos()
				return false
			}
		}
		return true
	})

	require.NotZero(t, callPos)

	newFile, err := transform.RenameSymbolAtPos(f, callPos, "Greet")
	require.NoError(t, err)
	require.NotNil(t, newFile)

	printer := NewPrinter(p.fset)
	result, err := printer.PrintFile(newFile)
	require.NoError(t, err)

	assert.Contains(t, result, "func Greet()")
	assert.Contains(t, result, "_ = Greet()")
	assert.NotContains(t, result, "Hello()")
}

func TestRenameSymbolAtPos_EmptyName(t *testing.T) {
	t.Parallel()

	src := `package main

func hello() string {
	return "hello"
}
`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	transform := NewTransform(p.fset)

	newFile, err := transform.RenameSymbolAtPos(f, token.NoPos, "")
	assert.Error(t, err)
	assert.Nil(t, newFile)
}

func TestRenameSymbolAtPos_SameName(t *testing.T) {
	t.Parallel()

	src := `package main

func hello() string {
	return "hello"
}
`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	transform := NewTransform(p.fset)

	var callPos token.Pos
	ast.Inspect(f, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok && ident.Name == "hello" {
			callPos = ident.Pos()
			return false
		}
		return true
	})

	newFile, err := transform.RenameSymbolAtPos(f, callPos, "hello")
	require.NoError(t, err)
	assert.Equal(t, f, newFile)
}

func TestRenameSymbolAtPos_NilFile(t *testing.T) {
	t.Parallel()

	transform := NewTransform(token.NewFileSet())
	_, err := transform.RenameSymbolAtPos(nil, token.NoPos, "newname")
	assert.Error(t, err)
}

func TestRenameSymbolAtPos_SymbolNotFound(t *testing.T) {
	t.Parallel()

	src := `package main

func main() {
}
`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	transform := NewTransform(p.fset)

	newFile, err := transform.RenameSymbolAtPos(f, token.NoPos, "newname")
	assert.Error(t, err)
	assert.Nil(t, newFile)
}

func TestRenameSymbolAtPos_Constant(t *testing.T) {
	t.Parallel()

	src := `package main

const MaxSize = 100

func main() {
	_ = MaxSize
}
`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	transform := NewTransform(p.fset)

	var constPos token.Pos
	ast.Inspect(f, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok && ident.Name == "MaxSize" {
			constPos = ident.Pos()
			return false
		}
		return true
	})

	require.NotZero(t, constPos)

	newFile, err := transform.RenameSymbolAtPos(f, constPos, "Limit")
	require.NoError(t, err)
	require.NotNil(t, newFile)

	printer := NewPrinter(p.fset)
	result, err := printer.PrintFile(newFile)
	require.NoError(t, err)

	assert.Contains(t, result, "const Limit = 100")
	assert.Contains(t, result, "_ = Limit")
	assert.NotContains(t, result, "MaxSize")
}

func TestRenameSymbolAtPos_Type(t *testing.T) {
	t.Parallel()

	src := `package main

type User struct {
	Name string
}

func createUser() User {
	return User{Name: "Alice"}
}

func main() {
	u := createUser()
}
`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	transform := NewTransform(p.fset)

	var typePos token.Pos
	ast.Inspect(f, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok && ident.Name == "User" {
			if ident.Obj != nil && ident.Obj.Kind == ast.Typ {
				typePos = ident.Pos()
				return false
			}
		}
		return true
	})

	require.NotZero(t, typePos)

	newFile, err := transform.RenameSymbolAtPos(f, typePos, "Person")
	require.NoError(t, err)
	require.NotNil(t, newFile)

	printer := NewPrinter(p.fset)
	result, err := printer.PrintFile(newFile)
	require.NoError(t, err)

	assert.Contains(t, result, "type Person struct")
	assert.Contains(t, result, "func createUser() Person")
	assert.Contains(t, result, "return Person{Name: \"Alice\"}")
	assert.NotContains(t, result, "type User struct")
}

func TestRenameSymbolAtPos_Method(t *testing.T) {
	t.Parallel()

	src := `package main

type User struct {
	Name string
}

func (u *User) Greet() string {
	return "Hello, " + u.Name
}

func main() {
	u := &User{Name: "Alice"}
	_ = u.Greet()
}
`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	transform := NewTransform(p.fset)

	var methodPos token.Pos
	ast.Inspect(f, func(n ast.Node) bool {
		if sel, ok := n.(*ast.SelectorExpr); ok {
			if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "u" && sel.Sel.Name == "Greet" {
				methodPos = sel.Sel.Pos()
				return false
			}
		}
		return true
	})

	require.NotZero(t, methodPos)

	newFile, err := transform.RenameSymbolAtPos(f, methodPos, "SayHello")
	require.NoError(t, err)
	require.NotNil(t, newFile)

	printer := NewPrinter(p.fset)
	result, err := printer.PrintFile(newFile)
	require.NoError(t, err)

	assert.Contains(t, result, "func (u *User) SayHello() string")
	assert.Contains(t, result, "_ = u.SayHello()")
	assert.NotContains(t, result, "func (u *User) Greet()")
}

func TestRenameSymbolAtPos_Variable(t *testing.T) {
	t.Parallel()

	src := `package main

func main() {
	name := "Alice"
	age := 30
	_ = name
	_ = age
}
`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	transform := NewTransform(p.fset)

	var varPos token.Pos
	ast.Inspect(f, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok && ident.Name == "name" {
			varPos = ident.Pos()
			return false
		}
		return true
	})

	require.NotZero(t, varPos)

	newFile, err := transform.RenameSymbolAtPos(f, varPos, "username")
	require.NoError(t, err)
	require.NotNil(t, newFile)

	printer := NewPrinter(p.fset)
	result, err := printer.PrintFile(newFile)
	require.NoError(t, err)

	t.Logf("Result:\n%s\n", result)

	assert.Contains(t, result, `username := "Alice"`)
	assert.Contains(t, result, "_ = username")
}
