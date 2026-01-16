package ast

//nolint:gosec // test file with necessary file operations

import (
	"go/ast"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindFunctions(t *testing.T) {
	src := `package main

func hello() string {
	return "hello"
}

func world() string {
	return "world"
}
`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	results := FindFunctions(f, "hello")
	assert.Len(t, results, 1)
	assert.Equal(t, "hello", results[0].Name)
	assert.Equal(t, "function", results[0].Type)
}

func TestFindFunctions_Wildcard(t *testing.T) {
	src := `package main

func hello() string {
	return "hello"
}

func world() string {
	return "world"
}
`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	results := FindFunctions(f, "*")
	assert.Len(t, results, 2)

	names := make(map[string]bool)
	for _, r := range results {
		names[r.Name] = true
	}
	assert.True(t, names["hello"])
	assert.True(t, names["world"])
}

func TestFindFunctions_Prefix(t *testing.T) {
	src := `package main

func hello() string {
	return "hello"
}

func helloWorld() string {
	return "hello world"
}

func world() string {
	return "world"
}
`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	results := FindFunctions(f, "hello*")
	assert.Len(t, results, 2)

	names := make(map[string]bool)
	for _, r := range results {
		names[r.Name] = true
	}
	assert.True(t, names["hello"])
	assert.True(t, names["helloWorld"])
}

func TestQuery_FindFunctions(t *testing.T) {
	src := `package main

func hello() string {
	return "hello"
}
`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	q := NewQuery(p.fset)
	files := map[string]*ast.File{"test.go": f}

	results := q.FindFunctions(files, "hello")
	assert.Len(t, results, 1)
	assert.Equal(t, "hello", results[0].Name)
	assert.Equal(t, "test.go", results[0].File)
	assert.Greater(t, results[0].Line, 0)
}

func TestFindImplementations(t *testing.T) {
	src := `package main

type Writer interface {
	Write(data string) error
}

type FileWriter struct{}

func (f *FileWriter) Write(data string) error {
	return nil
}
`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	files := map[string]*ast.File{"test.go": f}
	results := FindImplementations(files, "Writer")

	assert.Len(t, results, 1)
	assert.Equal(t, "FileWriter", results[0].Name)
	assert.Equal(t, "implementation", results[0].Type)
}

func TestQuery_FindImplementations(t *testing.T) {
	src := `package main

type Writer interface {
	Write(data string) error
}

type FileWriter struct{}

func (f *FileWriter) Write(data string) error {
	return nil
}
`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	q := NewQuery(p.fset)
	files := map[string]*ast.File{"test.go": f}
	results := q.FindImplementations(files, "Writer")

	assert.Len(t, results, 1)
	assert.Equal(t, "FileWriter", results[0].Name)
	assert.Equal(t, "test.go", results[0].File)
	assert.Greater(t, results[0].Line, 0)
}

func TestFindStructsByFieldType(t *testing.T) {
	src := `package main

type User struct {
	Name string
	Age  int
}

type Post struct {
	Title  string
	Author *User
}
`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	files := map[string]*ast.File{"test.go": f}
	results := FindStructsByFieldType(files, "*User")

	assert.Len(t, results, 1)
	assert.Equal(t, "Post", results[0].Name)
	assert.Equal(t, "struct_field", results[0].Type)
}

func TestQuery_FindStructsByFieldType(t *testing.T) {
	src := `package main

type User struct {
	Name string
	Age  int
}

type Post struct {
	Title  string
	Author *User
}
`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	q := NewQuery(p.fset)
	files := map[string]*ast.File{"test.go": f}
	results := q.FindStructsByFieldType(files, "*User")

	assert.Len(t, results, 1)
	assert.Equal(t, "Post", results[0].Name)
	assert.Equal(t, "test.go", results[0].File)
	assert.Greater(t, results[0].Line, 0)
}

func TestFindStructsByFieldType_Wildcard(t *testing.T) {
	src := `package main

type User struct {
	Name string
	Age  int
}
`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	files := map[string]*ast.File{"test.go": f}
	results := FindStructsByFieldType(files, "*")

	assert.Len(t, results, 1)
	assert.Equal(t, "User", results[0].Name)
}

func TestFindByImportDependency(t *testing.T) {
	src := `package main

import "fmt"

import "github.com/google/uuid"

func main() {
	fmt.Println("hello")
	_ = uuid.New()
}
`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	files := map[string]*ast.File{"test.go": f}
	results := FindByImportDependency(files, "fmt")

	assert.Len(t, results, 1)
	assert.Equal(t, "", results[0].Name)
	assert.Equal(t, "fmt", results[0].Signature)
	assert.Equal(t, "import_dependency", results[0].Type)
}

func TestFindByImportDependency_MultipleFiles(t *testing.T) {
	src1 := `package main

import "fmt"

func main() {
	fmt.Println("hello")
}
`

	src2 := `package main

import "fmt"
import "os"

func main() {
	fmt.Println(os.Args)
}
`

	p := NewParser()

	f1, err := p.ParseString(src1)
	require.NoError(t, err)

	f2, err := p.ParseString(src2)
	require.NoError(t, err)

	files := map[string]*ast.File{
		"file1.go": f1,
		"file2.go": f2,
	}

	results := FindByImportDependency(files, "fmt")

	assert.Len(t, results, 2)

	filenames := make(map[string]bool)
	for _, r := range results {
		filenames[r.File] = true
		assert.Equal(t, "fmt", r.Signature)
	}
	assert.True(t, filenames["file1.go"])
	assert.True(t, filenames["file2.go"])
}

func TestFindByImportDependency_FullPath(t *testing.T) {
	src := `package main

import "github.com/google/uuid"

func main() {
	_ = uuid.New()
}
`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	files := map[string]*ast.File{"test.go": f}
	results := FindByImportDependency(files, "github.com/google/uuid")

	assert.Len(t, results, 1)
	assert.Equal(t, "github.com/google/uuid", results[0].Signature)
}

func TestFindByImportDependency_Wildcard(t *testing.T) {
	src := `package main

import "github.com/google/uuid"
import "github.com/project/internal"

func main() {
	_ = uuid.New()
}
`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	files := map[string]*ast.File{"test.go": f}
	results := FindByImportDependency(files, "github.com/google/*")

	assert.Len(t, results, 1)
	assert.Equal(t, "github.com/google/uuid", results[0].Signature)
}

func TestFindByImportDependency_WildcardAll(t *testing.T) {
	src := `package main

import "fmt"
import "os"
import "github.com/google/uuid"
`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	files := map[string]*ast.File{"test.go": f}
	results := FindByImportDependency(files, "*")

	assert.Len(t, results, 3)

	importPaths := make(map[string]bool)
	for _, r := range results {
		importPaths[r.Signature] = true
	}
	assert.True(t, importPaths["fmt"])
	assert.True(t, importPaths["os"])
	assert.True(t, importPaths["github.com/google/uuid"])
}

func TestFindByImportDependency_NoMatch(t *testing.T) {
	src := `package main

import "fmt"

func main() {
	fmt.Println("hello")
}
`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	files := map[string]*ast.File{"test.go": f}
	results := FindByImportDependency(files, "github.com/google/uuid")

	assert.Len(t, results, 0)
}

func TestQuery_FindByImportDependency(t *testing.T) {
	src := `package main

import "fmt"

func main() {
	fmt.Println("hello")
}
`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	q := NewQuery(p.fset)
	files := map[string]*ast.File{"test.go": f}
	results := q.FindByImportDependency(files, "fmt")

	assert.Len(t, results, 1)
	assert.Equal(t, "test.go", results[0].File)
	assert.Equal(t, "fmt", results[0].Signature)
	assert.Greater(t, results[0].Line, 0)
}

func TestQuery_FindByImportDependency_MultipleImports(t *testing.T) {
	src := `package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println(os.Args)
}
`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	q := NewQuery(p.fset)
	files := map[string]*ast.File{"test.go": f}
	results := q.FindByImportDependency(files, "os")

	assert.Len(t, results, 1)
	assert.Equal(t, "test.go", results[0].File)
	assert.Equal(t, "os", results[0].Signature)
	assert.Greater(t, results[0].Line, 0)
}

func TestFindByImportDependency_NilFiles(t *testing.T) {
	results := FindByImportDependency(nil, "fmt")
	assert.Nil(t, results)
}

func TestFindByImportDependency_EmptyPattern(t *testing.T) {
	src := `package main

import "fmt"
`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	files := map[string]*ast.File{"test.go": f}
	results := FindByImportDependency(files, "")

	assert.Nil(t, results)
}

func TestFindByImportDependency_EmptyFile(t *testing.T) {
	src := `package main
`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	files := map[string]*ast.File{"test.go": f}
	results := FindByImportDependency(files, "fmt")

	assert.Len(t, results, 0)
}

func TestFindBySignature(t *testing.T) {
	src := `package main

func hello() string {
	return "hello"
}

func world() string {
	return "world"
}
`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	results := FindBySignature(f, "() string")
	assert.Len(t, results, 2)
}

func TestQuery_FindBySignature(t *testing.T) {
	src := `package main

func hello() string {
	return "hello"
}
`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	q := NewQuery(p.fset)
	files := map[string]*ast.File{"test.go": f}
	results := q.FindBySignature(files, "() string")

	assert.Len(t, results, 1)
	assert.Equal(t, "hello", results[0].Name)
	assert.Greater(t, results[0].Line, 0)
}

func TestMatchImportPath_Exact(t *testing.T) {
	tests := []struct {
		name       string
		pattern    string
		importPath string
		want       bool
	}{
		{"exact match", "fmt", "fmt", true},
		{"exact match with path", "github.com/google/uuid", "github.com/google/uuid", true},
		{"no match", "fmt", "os", false},
		{"no match different path", "github.com/google/uuid", "github.com/project/internal", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchImportPath(tt.pattern, tt.importPath)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestMatchImportPath_Wildcard(t *testing.T) {
	tests := []struct {
		name       string
		pattern    string
		importPath string
		want       bool
	}{
		{"wildcard all", "*", "fmt", true},
		{"wildcard prefix match", "github.com/google/*", "github.com/google/uuid", true},
		{"wildcard prefix match another", "github.com/google/*", "github.com/google/cloud", true},
		{"wildcard prefix no match", "github.com/google/*", "github.com/project/internal", false},
		{"wildcard prefix no match different root", "github.com/google/*", "golang.org/x/text", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchImportPath(tt.pattern, tt.importPath)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestQuery_Line(t *testing.T) {
	src := `package main

func hello() string {
	return "hello"
}

type User struct {
	Name string
}
`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	q := NewQuery(p.fset)

	line := q.line(f, "hello")
	assert.Greater(t, line, 0)

	line = q.line(f, "User")
	assert.Greater(t, line, 0)
}

func TestQuery_LineForImport(t *testing.T) {
	src := `package main

import "fmt"

func main() {
	fmt.Println("hello")
}
`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	q := NewQuery(p.fset)

	line := q.lineForImport(f, "fmt")
	assert.Greater(t, line, 0)
}

func TestQuery_NilFileSet(t *testing.T) {
	src := `package main

import "fmt"
`

	p := NewParser()
	f, err := p.ParseString(src)
	require.NoError(t, err)

	q := NewQuery(nil)
	files := map[string]*ast.File{"test.go": f}
	results := q.FindByImportDependency(files, "fmt")

	assert.Len(t, results, 1)
	assert.Equal(t, 0, results[0].Line)
}
