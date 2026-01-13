package ast

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

func TestParser_ParseString_Empty(t *testing.T) {
	p := NewParser()

	_, err := p.ParseString("")
	assert.Error(t, err)
}
