package ast

//nolint:gosec // test file with necessary file operations

import (
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewParser(t *testing.T) {
	p := NewParser()
	assert.NotNil(t, p)
	assert.NotNil(t, p.fset)
}

func TestParser_ParseFile(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		content string
		wantErr bool
		check   func(t *testing.T, f *ast.File)
	}{
		{
			name: "simple package",
			content: `package main
`,
			wantErr: false,
			check: func(t *testing.T, f *ast.File) {
				assert.Equal(t, "main", f.Name.Name)
				assert.Empty(t, f.Decls)
			},
		},
		{
			name: "package with imports",
			content: `package main

import "fmt"
`,
			wantErr: false,
			check: func(t *testing.T, f *ast.File) {
				assert.Equal(t, "main", f.Name.Name)
				assert.Len(t, f.Imports, 1)
				assert.Equal(t, `"fmt"`, f.Imports[0].Path.Value)
			},
		},
		{
			name: "package with function",
			content: `package main

func hello() string {
	return "hello"
}
`,
			wantErr: false,
			check: func(t *testing.T, f *ast.File) {
				assert.Equal(t, "main", f.Name.Name)
				assert.Len(t, f.Decls, 1)
				fnDecl, ok := f.Decls[0].(*ast.FuncDecl)
				require.True(t, ok)
				assert.Equal(t, "hello", fnDecl.Name.Name)
			},
		},
		{
			name: "package with struct",
			content: `package main

type User struct {
	Name string
	Age  int
}
`,
			wantErr: false,
			check: func(t *testing.T, f *ast.File) {
				assert.Equal(t, "main", f.Name.Name)
				assert.Len(t, f.Decls, 1)
				genDecl, ok := f.Decls[0].(*ast.GenDecl)
				require.True(t, ok)
				assert.Equal(t, token.TYPE, genDecl.Tok)
			},
		},
		{
			name: "complex file",
			content: `package main

import (
	"fmt"
)

type User struct {
	Name string
}

func (u *User) Greet() string {
	return "Hello, " + u.Name
}

func main() {
	u := &User{Name: "Alice"}
	fmt.Println(u.Greet())
}
`,
			wantErr: false,
			check: func(t *testing.T, f *ast.File) {
				assert.Equal(t, "main", f.Name.Name)
				assert.Len(t, f.Decls, 4)
			},
		},
		{
			name: "syntax error",
			content: `package main

func broken( {
}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := NewParser()

			testFile := filepath.Join(tmpDir, tt.name+".go")
			err := os.WriteFile(testFile, []byte(tt.content), 0600)
			require.NoError(t, err)

			f, err := p.ParseFile(testFile)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, f)
			} else {
				require.NoError(t, err)
				require.NotNil(t, f)
				if tt.check != nil {
					tt.check(t, f)
				}
			}
		})
	}
}

func TestParser_ParseFile_NotExist(t *testing.T) {
	p := NewParser()
	_, err := p.ParseFile("/nonexistent/file.go")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "open")
}

func TestParser_ParseFile_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "empty.go")

	err := os.WriteFile(testFile, []byte(""), 0600)
	require.NoError(t, err)

	p := NewParser()
	_, err = p.ParseFile(testFile)
	assert.Error(t, err)
}

func TestParser_ParseFile_NonGoFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	err := os.WriteFile(testFile, []byte("not go code"), 0600)
	require.NoError(t, err)

	p := NewParser()
	_, err = p.ParseFile(testFile)
	assert.Error(t, err)
}

func TestParser_ParseString(t *testing.T) {
	p := NewParser()

	src := `package main

import "fmt"

func hello() {
	fmt.Println("hello")
}
`

	f, err := p.ParseString(src)
	require.NoError(t, err)
	assert.NotNil(t, f)
	assert.Equal(t, "main", f.Name.Name)
	assert.Len(t, f.Decls, 2)
}

func TestParser_ParseString_InvalidSyntax(t *testing.T) {
	p := NewParser()

	src := `package main

func broken( {
}
`

	_, err := p.ParseString(src)
	assert.Error(t, err)
}

func TestFindDefinition_Function(t *testing.T) {
	src := `package main

func hello() string {
	return "hello"
}

func main() {
	fmt.Println(hello())
}`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	builder := NewBuilder(p.fset)
	root, err := builder.BuildFile(f)
	require.NoError(t, err)
	require.NotNil(t, root)

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
	def := builder.FindDefinition("hello", callPos)
	require.NotNil(t, def)
	assert.Equal(t, "hello", def.Name)
	assert.Equal(t, "function", def.Kind.String())
	assert.True(t, !def.Exported)
	assert.Equal(t, "() string", def.Type)
}

func TestFindDefinition_ExportedFunction(t *testing.T) {
	src := `package main

func Hello() string {
	return "hello"
}

func main() {
	fmt.Println(Hello())
}`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	builder := NewBuilder(p.fset)
	root, err := builder.BuildFile(f)
	require.NoError(t, err)
	require.NotNil(t, root)

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
	def := builder.FindDefinition("Hello", callPos)
	require.NotNil(t, def)
	assert.Equal(t, "Hello", def.Name)
	assert.Equal(t, "function", def.Kind.String())
	assert.True(t, def.Exported)
}

func TestFindDefinition_Variable(t *testing.T) {
	src := `package main

func main() {
	name := "Alice"
	fmt.Println(name)
}`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	builder := NewBuilder(p.fset)
	root, err := builder.BuildFile(f)
	require.NoError(t, err)
	require.NotNil(t, root)

	var varPos token.Pos
	count := 0
	ast.Inspect(f, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok && ident.Name == "name" {
			count++
			if count == 2 {
				varPos = ident.Pos()
				return false
			}
		}
		return true
	})

	require.NotZero(t, varPos)
	def := builder.FindDefinition("name", varPos)
	require.NotNil(t, def)
	assert.Equal(t, "name", def.Name)
	assert.Equal(t, "variable", def.Kind.String())
}

func TestFindDefinition_StructField(t *testing.T) {
	src := `package main

type User struct {
	Name string
	Age  int
}

func main() {
	u := User{Name: "Alice"}
	fmt.Println(u.Name)
}`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	builder := NewBuilder(p.fset)
	root, err := builder.BuildFile(f)
	require.NoError(t, err)
	require.NotNil(t, root)

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
	def := builder.FindDefinition("Name", fieldPos)
	require.NotNil(t, def)
	assert.Equal(t, "Name", def.Name)
	assert.Equal(t, "field", def.Kind.String())
}

func TestFindDefinition_Type(t *testing.T) {
	src := `package main

type User struct {
	Name string
}

func createUser() User {
	return User{Name: "Alice"}
}

func main() {
	u := createUser()
}`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	builder := NewBuilder(p.fset)
	root, err := builder.BuildFile(f)
	require.NoError(t, err)
	require.NotNil(t, root)

	var typePos token.Pos
	ast.Inspect(f, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok && ident.Name == "User" {
			typePos = ident.Pos()
			return false
		}
		return true
	})

	require.NotZero(t, typePos)
	def := builder.FindDefinition("User", typePos)
	require.NotNil(t, def)
	assert.Equal(t, "User", def.Name)
	assert.Equal(t, "type", def.Kind.String())
	assert.True(t, def.Exported)
}

func TestFindDefinition_NilBuilder(t *testing.T) {
	builder := &Builder{}
	def := builder.FindDefinition("test", token.NoPos)
	assert.Nil(t, def)
}

func TestFindDefinition_NilFileSet(t *testing.T) {
	builder := NewBuilder(nil)
	def := builder.FindDefinition("test", token.NoPos)
	assert.Nil(t, def)
}

func TestFindDefinition_NotFound(t *testing.T) {
	src := `package main

func main() {
	fmt.Println("hello")
}`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	builder := NewBuilder(p.fset)
	_, err = builder.BuildFile(f)
	require.NoError(t, err)

	def := builder.FindDefinition("nonexistent", token.NoPos)
	assert.Nil(t, def)
}

func TestFindDefinition_Method(t *testing.T) {
	src := `package main

type User struct {
	Name string
}

func (u *User) Greet() string {
	return "Hello, " + u.Name
}

func main() {
	u := &User{Name: "Alice"}
	fmt.Println(u.Greet())
}`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	builder := NewBuilder(p.fset)
	root, err := builder.BuildFile(f)
	require.NoError(t, err)
	require.NotNil(t, root)

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
	def := builder.FindDefinition("Greet", methodPos)
	require.NotNil(t, def)
	assert.Equal(t, "Greet", def.Name)
	assert.Equal(t, "method", def.Kind.String())
	assert.True(t, def.Exported)
}

func TestFindDefinition_Constant(t *testing.T) {
	src := `package main

const MaxSize = 100

func main() {
	fmt.Println(MaxSize)
}`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	builder := NewBuilder(p.fset)
	root, err := builder.BuildFile(f)
	require.NoError(t, err)
	require.NotNil(t, root)

	var constPos token.Pos
	ast.Inspect(f, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok && ident.Name == "MaxSize" {
			constPos = ident.Pos()
			return false
		}
		return true
	})

	require.NotZero(t, constPos)
	def := builder.FindDefinition("MaxSize", constPos)
	require.NotNil(t, def)
	assert.Equal(t, "MaxSize", def.Name)
	assert.Equal(t, "constant", def.Kind.String())
	assert.True(t, def.Exported)
}

func TestParser_ParseString_Empty(t *testing.T) {
	p := NewParser()

	_, err := p.ParseString("")
	assert.Error(t, err)
}
