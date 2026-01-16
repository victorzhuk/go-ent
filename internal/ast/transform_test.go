package ast

//nolint:gosec // test file with necessary file operations

import (
	"go/ast"
	"go/token"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testFile struct {
	path string
	src  string
	file *ast.File
	fset *token.FileSet
	p    *Parser
}

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

func TestRenameSymbol_StructEmbedding(t *testing.T) {
	t.Parallel()

	src := `package main

type Base struct {
	Name string
}

type Derived struct {
	Base
	Extra string
}

func (d *Derived) GetInfo() string {
	return d.Name + d.Extra
}

func main() {
	derived := Derived{
		Base: Base{Name: "Test"},
		Extra: "Extra",
	}
	_ = derived.GetInfo()
	_ = derived.Name
}
`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	transform := NewTransform(p.fset)

	var baseTypePos token.Pos
	ast.Inspect(f, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok && ident.Name == "Base" {
			if ident.Obj != nil && ident.Obj.Kind == ast.Typ {
				baseTypePos = ident.Pos()
				return false
			}
		}
		return true
	})

	require.NotZero(t, baseTypePos)

	newFile, err := transform.RenameSymbolAtPos(f, baseTypePos, "BaseType")
	require.NoError(t, err)
	require.NotNil(t, newFile)

	printer := NewPrinter(p.fset)
	result, err := printer.PrintFile(newFile)
	require.NoError(t, err)

	t.Logf("Result:\n%s\n", result)

	assert.Contains(t, result, "type BaseType struct")
	assert.Contains(t, result, "type Derived struct {")
	assert.Contains(t, result, "BaseType")
	assert.Contains(t, result, "BaseType: BaseType{Name: \"Test\"}")
	assert.NotContains(t, result, "type Base struct")
}

func TestRenameSymbol_GenericTypeParam(t *testing.T) {
	t.Parallel()

	src := `package main

func Process[T any](items []T) []T {
	return items
}

func main() {
	nums := []int{1, 2, 3}
	_ = Process(nums)
}
`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	transform := NewTransform(p.fset)

	var typeParamPos token.Pos
	ast.Inspect(f, func(n ast.Node) bool {
		if fd, ok := n.(*ast.FuncDecl); ok {
			if fd.Name.Name == "Process" && fd.Type.TypeParams != nil {
				for _, param := range fd.Type.TypeParams.List {
					for _, name := range param.Names {
						if name.Name == "T" {
							typeParamPos = name.Pos()
							return false
						}
					}
				}
			}
		}
		return true
	})

	require.NotZero(t, typeParamPos)

	newFile, err := transform.RenameSymbolAtPos(f, typeParamPos, "Item")
	require.NoError(t, err)
	require.NotNil(t, newFile)

	printer := NewPrinter(p.fset)
	result, err := printer.PrintFile(newFile)
	require.NoError(t, err)

	t.Logf("Result:\n%s\n", result)

	assert.Contains(t, result, "func Process[Item any]")
	assert.Contains(t, result, "items []Item")
	assert.Contains(t, result, "[]Item")
	assert.NotContains(t, result, "[T any]")
}

func TestRenameSymbol_GenericStruct(t *testing.T) {
	t.Parallel()

	src := `package main

type Container[T any] struct {
	Value T
}

func NewContainer[T any](v T) Container[T] {
	return Container[T]{Value: v}
}

func main() {
	c := NewContainer(42)
	_ = c.Value
}
`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	transform := NewTransform(p.fset)

	var typeParamPos token.Pos
	ast.Inspect(f, func(n ast.Node) bool {
		if ts, ok := n.(*ast.TypeSpec); ok {
			if ts.Name.Name == "Container" && ts.TypeParams != nil {
				for _, param := range ts.TypeParams.List {
					for _, name := range param.Names {
						if name.Name == "T" {
							typeParamPos = name.Pos()
							return false
						}
					}
				}
			}
		}
		return true
	})

	require.NotZero(t, typeParamPos)

	newFile, err := transform.RenameSymbolAtPos(f, typeParamPos, "E")
	require.NoError(t, err)
	require.NotNil(t, newFile)

	printer := NewPrinter(p.fset)
	result, err := printer.PrintFile(newFile)
	require.NoError(t, err)

	t.Logf("Result:\n%s\n", result)

	assert.Contains(t, result, "type Container[E any]")
	assert.Contains(t, result, "Value E")
	assert.Contains(t, result, "func NewContainer[T any](v T) Container[T]")
	assert.NotContains(t, result, "struct {\n\tValue T\n}")
}

func TestRenameSymbol_ComplexShadowing(t *testing.T) {
	t.Parallel()

	src := `package main

func main() {
	x := 1
	_ = x

	if true {
		y := x + 1
		_ = y

		if true {
			x := 3
			_ = x
		}

		_ = y
	}

	_ = x
}
`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	transform := NewTransform(p.fset)

	var outerXPos token.Pos
	count := 0
	ast.Inspect(f, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok && ident.Name == "x" {
			count++
			if count == 2 {
				outerXPos = ident.Pos()
				return false
			}
		}
		return true
	})

	require.NotZero(t, outerXPos)

	newFile, err := transform.RenameSymbolAtPos(f, outerXPos, "value")
	require.NoError(t, err)
	require.NotNil(t, newFile)

	printer := NewPrinter(p.fset)
	result, err := printer.PrintFile(newFile)
	require.NoError(t, err)

	t.Logf("Result:\n%s\n", result)

	assert.Contains(t, result, "value := 1")
	assert.Contains(t, result, "_ = value")
	assert.Contains(t, result, "y := value + 1")

	lines := strings.Split(result, "\n")

	innerIfIndex := -1
	for i, line := range lines {
		if strings.Contains(line, "x := 3") {
			innerIfIndex = i
			break
		}
	}
	require.Greater(t, innerIfIndex, 0)

	innerBlock := strings.Join(lines[innerIfIndex:], "\n")
	assert.Contains(t, innerBlock, "x := 3")
	assert.Contains(t, innerBlock, "_ = x")

	afterInnerIf := strings.Join(lines[:innerIfIndex], "\n")
	assert.NotContains(t, afterInnerIf, "x := 3")
}

func TestRenameSymbol_NestedShadowing(t *testing.T) {
	t.Parallel()

	src := `package main

func main() {
	result := 10

	for i := 0; i < 3; i++ {
		result += i
		_ = result
	}

	for _, val := range []int{1, 2, 3} {
		result += val
		_ = result
	}

	_ = result
}
`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	transform := NewTransform(p.fset)

	var resultPos token.Pos
	ast.Inspect(f, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok && ident.Name == "result" {
			resultPos = ident.Pos()
			return false
		}
		return true
	})

	require.NotZero(t, resultPos)

	newFile, err := transform.RenameSymbolAtPos(f, resultPos, "total")
	require.NoError(t, err)
	require.NotNil(t, newFile)

	printer := NewPrinter(p.fset)
	result, err := printer.PrintFile(newFile)
	require.NoError(t, err)

	t.Logf("Result:\n%s\n", result)

	assert.Contains(t, result, "total := 10")
	assert.Contains(t, result, "total += i")
	assert.Contains(t, result, "total += val")
	assert.Contains(t, result, "_ = total")
	assert.NotContains(t, result, "result :=")
}

func TestRenameSymbol_FunctionParamShadowing(t *testing.T) {
	src := `package main

func process(x int) int {
	if x > 0 {
		x := x * 2
		_ = x
	}
	return x
}

func main() {
	_ = process(5)
}
`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	transform := NewTransform(p.fset)

	var paramPos token.Pos
	ast.Inspect(f, func(n ast.Node) bool {
		if fd, ok := n.(*ast.FuncDecl); ok && fd.Name.Name == "process" {
			if fd.Type.Params != nil {
				for _, field := range fd.Type.Params.List {
					for _, name := range field.Names {
						if name.Name == "x" {
							paramPos = name.Pos()
							return false
						}
					}
				}
			}
			return false
		}
		return true
	})

	require.NotZero(t, paramPos)

	newFile, err := transform.RenameSymbolAtPos(f, paramPos, "input")
	require.NoError(t, err)
	require.NotNil(t, newFile)

	printer := NewPrinter(p.fset)
	result, err := printer.PrintFile(newFile)
	require.NoError(t, err)

	t.Logf("Result:\n%s\n", result)

	assert.Contains(t, result, "func process(input int) int")
	assert.Contains(t, result, "if input > 0")
	assert.Contains(t, result, "return input")

	assert.Contains(t, result, "x := input * 2")
	assert.Contains(t, result, "_ = x")
}

func TestRenameSymbol_EmbeddedFieldAccess(t *testing.T) {
	t.Parallel()

	src := `package main

type Logger struct {
	Level string
}

type Service struct {
	Logger
	Name string
}

func (s *Service) Log(msg string) {
	_ = s.Name
}

func main() {
	svc := &Service{
		Logger: Logger{Level: "info"},
		Name: "test",
	}
	_ = svc.Log()
	_ = svc.Logger.Level
	_ = svc.Name
}
`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	transform := NewTransform(p.fset)

	var loggerPos token.Pos
	ast.Inspect(f, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok && ident.Name == "Logger" {
			if ident.Obj != nil && ident.Obj.Kind == ast.Typ {
				loggerPos = ident.Pos()
				return false
			}
		}
		return true
	})

	require.NotZero(t, loggerPos)

	newFile, err := transform.RenameSymbolAtPos(f, loggerPos, "BaseLogger")
	require.NoError(t, err)
	require.NotNil(t, newFile)

	printer := NewPrinter(p.fset)
	result, err := printer.PrintFile(newFile)
	require.NoError(t, err)

	t.Logf("Result:\n%s\n", result)

	assert.Contains(t, result, "type BaseLogger struct")
	assert.Contains(t, result, "type Service struct {")
	assert.Contains(t, result, "BaseLogger")
	assert.Contains(t, result, "Logger: BaseLogger{Level: \"info\"}")
	assert.Contains(t, result, "svc.Logger.Level")
	assert.NotContains(t, result, "type Logger struct")
}

func TestRenameSymbol_MultipleFiles_FunctionAcrossPackage(t *testing.T) {
	t.Parallel()

	files := map[string]string{
		"types.go": `package testpkg

type Config struct {
	Debug bool
}

func LoadConfig() *Config {
	return &Config{Debug: true}
}
`,
		"handler.go": `package testpkg

func Handle(cfg *Config) string {
	if cfg.Debug {
		return "debug mode"
	}
	return "normal mode"
}
`,
		"main.go": `package main

import "testpkg"

func main() {
	cfg := testpkg.LoadConfig()
	_ = testpkg.Handle(cfg)
}
`,
	}

	parsedFiles := make(map[string]*testFile)

	for path, src := range files {
		p := NewParser()
		f, err := p.ParseString(src)
		require.NoError(t, err)

		tf := &testFile{
			path: path,
			src:  src,
			file: f,
			fset: p.fset,
			p:    p,
		}
		parsedFiles[path] = tf
	}

	tf := parsedFiles["types.go"]

	var funcPos token.Pos
	ast.Inspect(tf.file, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok && ident.Name == "LoadConfig" {
			if ident.Obj != nil && ident.Obj.Kind == ast.Fun {
				funcPos = ident.Pos()
				return false
			}
		}
		return true
	})

	require.NotZero(t, funcPos)

	transform := NewTransform(tf.fset)
	newFile, err := transform.RenameSymbolAtPos(tf.file, funcPos, "GetConfig")
	require.NoError(t, err)

	printer := NewPrinter(tf.fset)
	result, err := printer.PrintFile(newFile)
	require.NoError(t, err)

	t.Logf("Result for types.go:\n%s\n", result)

	assert.Contains(t, result, "func GetConfig()")
	assert.NotContains(t, result, "func LoadConfig()")
}

func TestRenameSymbol_MultipleFiles_TypeAcrossPackage(t *testing.T) {
	t.Parallel()

	files := map[string]string{
		"user.go": `package testpkg

type User struct {
	ID   int
	Name string
}

func (u *User) Validate() bool {
	return u.Name != ""
}
`,
		"service.go": `package testpkg

func GetUser(id int) *User {
	return &User{ID: id, Name: "Test"}
}

func SaveUser(u *User) error {
	return nil
}
`,
		"main.go": `package main

import "testpkg"

func main() {
	u := testpkg.GetUser(1)
	_ = u.Validate()
	_ = testpkg.SaveUser(u)
}
`,
	}

	parsedFiles := make(map[string]*testFile)

	for path, src := range files {
		p := NewParser()
		f, err := p.ParseString(src)
		require.NoError(t, err)

		parsedFiles[path] = &testFile{
			path: path,
			src:  src,
			file: f,
			fset: p.fset,
			p:    p,
		}
	}

	tf := parsedFiles["user.go"]

	var typePos token.Pos
	ast.Inspect(tf.file, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok && ident.Name == "User" {
			if ident.Obj != nil && ident.Obj.Kind == ast.Typ {
				typePos = ident.Pos()
				return false
			}
		}
		return true
	})

	require.NotZero(t, typePos)

	transform := NewTransform(tf.fset)
	newFile, err := transform.RenameSymbolAtPos(tf.file, typePos, "Person")
	require.NoError(t, err)

	printer := NewPrinter(tf.fset)
	result, err := printer.PrintFile(newFile)
	require.NoError(t, err)

	t.Logf("Result for user.go:\n%s\n", result)

	assert.Contains(t, result, "type Person struct")
	assert.Contains(t, result, "func (u *Person) Validate()")
	assert.Contains(t, result, "return u.Name")
	assert.NotContains(t, result, "type User struct")
}

func TestRenameSymbol_MultipleFiles_MethodAcrossPackage(t *testing.T) {
	t.Parallel()

	files := map[string]string{
		"model.go": `package testpkg

type Item struct {
	Name  string
	Price int
}

func (i *Item) GetName() string {
	return i.Name
}

func (i *Item) GetPrice() int {
	return i.Price
}
`,
		"repository.go": `package testpkg

func FindItem(id int) *Item {
	return &Item{Name: "Test", Price: 100}
}

func UpdateItem(i *Item) error {
	i.Name = "Updated"
	return nil
}
`,
		"main.go": `package main

import "testpkg"

func main() {
	item := testpkg.FindItem(1)
	_ = item.GetName()
	_ = testpkg.UpdateItem(item)
}
`,
	}

	parsedFiles := make(map[string]*testFile)

	for path, src := range files {
		p := NewParser()
		f, err := p.ParseString(src)
		require.NoError(t, err)

		parsedFiles[path] = &testFile{
			path: path,
			src:  src,
			file: f,
			fset: p.fset,
			p:    p,
		}
	}

	tf := parsedFiles["model.go"]

	var methodPos token.Pos
	ast.Inspect(tf.file, func(n ast.Node) bool {
		if fd, ok := n.(*ast.FuncDecl); ok && fd.Name.Name == "GetName" {
			methodPos = fd.Name.Pos()
			return false
		}
		return true
	})

	require.NotZero(t, methodPos)

	transform := NewTransform(tf.fset)
	newFile, err := transform.RenameSymbolAtPos(tf.file, methodPos, "Name")
	require.NoError(t, err)

	printer := NewPrinter(tf.fset)
	result, err := printer.PrintFile(newFile)
	require.NoError(t, err)

	t.Logf("Result for model.go:\n%s\n", result)

	assert.Contains(t, result, "func (i *Item) Name()")
	assert.NotContains(t, result, "func (i *Item) GetName()")
}

func TestRenameSymbol_MultipleFiles_ConstantAcrossPackage(t *testing.T) {
	t.Parallel()

	files := map[string]string{
		"constants.go": `package testpkg

const (
	MaxItems = 100
	MinItems = 1
)

func GetMax() int {
	return MaxItems
}
`,
		"validator.go": `package testpkg

func ValidateCount(count int) bool {
	return count >= MinItems && count <= MaxItems
}
`,
		"main.go": `package main

import "testpkg"

func main() {
	_ = testpkg.ValidateCount(50)
	_ = testpkg.GetMax()
}
`,
	}

	parsedFiles := make(map[string]*testFile)

	for path, src := range files {
		p := NewParser()
		f, err := p.ParseString(src)
		require.NoError(t, err)

		parsedFiles[path] = &testFile{
			path: path,
			src:  src,
			file: f,
			fset: p.fset,
			p:    p,
		}
	}

	tf := parsedFiles["constants.go"]

	var constPos token.Pos
	ast.Inspect(tf.file, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok && ident.Name == "MaxItems" {
			if ident.Obj != nil && ident.Obj.Kind == ast.Con {
				constPos = ident.Pos()
				return false
			}
		}
		return true
	})

	require.NotZero(t, constPos)

	transform := NewTransform(tf.fset)
	newFile, err := transform.RenameSymbolAtPos(tf.file, constPos, "Limit")
	require.NoError(t, err)

	printer := NewPrinter(tf.fset)
	result, err := printer.PrintFile(newFile)
	require.NoError(t, err)

	t.Logf("Result for constants.go:\n%s\n", result)

	assert.Contains(t, result, "const (")
	assert.Contains(t, result, "Limit")
	assert.Contains(t, result, "return Limit")
	assert.NotContains(t, result, "MaxItems")
}
