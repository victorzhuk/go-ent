package ast

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateInterfaceImplementation(t *testing.T) {
	tests := []struct {
		name         string
		ifaceSrc     string
		implTypeName string
		implMethods  map[string]*ast.FuncType
		wantErr      bool
		errContains  string
		checkOutput  func(t *testing.T, node ast.Node)
	}{
		{
			name: "simple interface",
			ifaceSrc: `package test
type Writer interface {
	Write(data []byte) (int, error)
}`,
			implTypeName: "FileWriter",
			implMethods:  nil,
			wantErr:      false,
			checkOutput: func(t *testing.T, node ast.Node) {
				f, ok := node.(*ast.File)
				require.True(t, ok)

				assert.Len(t, f.Decls, 2)

				structDecl, ok := f.Decls[0].(*ast.GenDecl)
				require.True(t, ok)
				ts, ok := structDecl.Specs[0].(*ast.TypeSpec)
				require.True(t, ok)
				assert.Equal(t, "FileWriter", ts.Name.Name)

				methodDecl, ok := f.Decls[1].(*ast.FuncDecl)
				require.True(t, ok)
				assert.Equal(t, "Write", methodDecl.Name.Name)
			},
		},
		{
			name: "multiple methods",
			ifaceSrc: `package test
type Closer interface {
	Close() error
	Flush() error
}`,
			implTypeName: "Buffer",
			implMethods:  nil,
			wantErr:      false,
			checkOutput: func(t *testing.T, node ast.Node) {
				f, ok := node.(*ast.File)
				require.True(t, ok)

				assert.Len(t, f.Decls, 3)

				_, ok = f.Decls[0].(*ast.GenDecl)
				assert.True(t, ok)

				method1, ok := f.Decls[1].(*ast.FuncDecl)
				require.True(t, ok)
				assert.Equal(t, "Close", method1.Name.Name)

				method2, ok := f.Decls[2].(*ast.FuncDecl)
				require.True(t, ok)
				assert.Equal(t, "Flush", method2.Name.Name)
			},
		},
		{
			name:         "nil interface",
			ifaceSrc:     "",
			implTypeName: "Test",
			wantErr:      true,
			errContains:  "nil interface",
		},
		{
			name: "empty type name",
			ifaceSrc: `package test
type Test interface {
	Do() error
}`,
			implTypeName: "",
			wantErr:      true,
			errContains:  "empty implementation type name",
		},
		{
			name: "interface with no methods",
			ifaceSrc: `package test

type Empty interface{}`,
			implTypeName: "Impl",
			wantErr:      true,
			errContains:  "no methods",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var iface *ast.InterfaceType
			if tt.ifaceSrc != "" {
				fset := token.NewFileSet()
				f, err := parser.ParseFile(fset, "", tt.ifaceSrc, parser.AllErrors)
				require.NoError(t, err)

				genDecl := f.Decls[0].(*ast.GenDecl)
				typeSpec := genDecl.Specs[0].(*ast.TypeSpec)
				var ok bool
				iface, ok = typeSpec.Type.(*ast.InterfaceType)
				require.True(t, ok)
			}

			node, err := GenerateInterfaceImplementation(iface, tt.implTypeName, tt.implMethods)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, node)
				if tt.checkOutput != nil {
					tt.checkOutput(t, node)
				}
			}
		})
	}
}

func TestGenerateInterfaceImplementation_PointerReceiver(t *testing.T) {
	ifaceSrc := `package test
type Reader interface {
	Read(p []byte) (n int, err error)
}`

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", ifaceSrc, parser.AllErrors)
	require.NoError(t, err)

	genDecl := f.Decls[0].(*ast.GenDecl)
	typeSpec := genDecl.Specs[0].(*ast.TypeSpec)
	iface := typeSpec.Type.(*ast.InterfaceType)

	node, err := GenerateInterfaceImplementation(iface, "ByteReader", nil)
	require.NoError(t, err)

	fileNode, ok := node.(*ast.File)
	require.True(t, ok)

	methodDecl, ok := fileNode.Decls[1].(*ast.FuncDecl)
	require.True(t, ok)

	assert.Equal(t, "Read", methodDecl.Name.Name)
	require.NotNil(t, methodDecl.Recv)
	assert.Len(t, methodDecl.Recv.List, 1)

	recvField := methodDecl.Recv.List[0]
	assert.Len(t, recvField.Names, 1)
	assert.Equal(t, "b", recvField.Names[0].Name)

	starType, ok := recvField.Type.(*ast.StarExpr)
	require.True(t, ok)
	ident, ok := starType.X.(*ast.Ident)
	require.True(t, ok)
	assert.Equal(t, "ByteReader", ident.Name)
}

func TestGenerateInterfaceImplementation_ReturnValues(t *testing.T) {
	ifaceSrc := `package test
type MultiReturn interface {
	Get() (int, string, error)
}`

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", ifaceSrc, parser.AllErrors)
	require.NoError(t, err)

	genDecl := f.Decls[0].(*ast.GenDecl)
	typeSpec := genDecl.Specs[0].(*ast.TypeSpec)
	iface := typeSpec.Type.(*ast.InterfaceType)

	node, err := GenerateInterfaceImplementation(iface, "Getter", nil)
	require.NoError(t, err)

	fileNode, ok := node.(*ast.File)
	require.True(t, ok)

	methodDecl, ok := fileNode.Decls[1].(*ast.FuncDecl)
	require.True(t, ok)

	require.NotNil(t, methodDecl.Body)
	assert.Len(t, methodDecl.Body.List, 1)

	retStmt, ok := methodDecl.Body.List[0].(*ast.ReturnStmt)
	require.True(t, ok)
	assert.Len(t, retStmt.Results, 3)
}

func TestGenerateInterfaceImplementation_NoReturn(t *testing.T) {
	ifaceSrc := `package test
type Closer interface {
	Close()
}`

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", ifaceSrc, parser.AllErrors)
	require.NoError(t, err)

	genDecl := f.Decls[0].(*ast.GenDecl)
	typeSpec := genDecl.Specs[0].(*ast.TypeSpec)
	iface := typeSpec.Type.(*ast.InterfaceType)

	node, err := GenerateInterfaceImplementation(iface, "CloserImpl", nil)
	require.NoError(t, err)

	fileNode, ok := node.(*ast.File)
	require.True(t, ok)

	methodDecl, ok := fileNode.Decls[1].(*ast.FuncDecl)
	require.True(t, ok)

	require.NotNil(t, methodDecl.Body)
	assert.Len(t, methodDecl.Body.List, 1)

	retStmt, ok := methodDecl.Body.List[0].(*ast.ReturnStmt)
	require.True(t, ok)
	assert.Nil(t, retStmt.Results)
}

func TestNewEngine(t *testing.T) {
	e := NewEngine(nil)
	assert.NotNil(t, e)
	assert.NotNil(t, e.fset)
	assert.NotNil(t, e.templates)
}

func TestNewEngine_WithFileSet(t *testing.T) {
	fset := token.NewFileSet()
	e := NewEngine(fset)
	assert.NotNil(t, e)
	assert.Equal(t, fset, e.fset)
}

func TestEngine_Register(t *testing.T) {
	tests := []struct {
		name        string
		tmpl        *Template
		wantErr     bool
		errContains string
	}{
		{
			name: "valid template",
			tmpl: &Template{
				Name:   "test",
				Source: `type TypeName struct {}`,
			},
			wantErr: false,
		},
		{
			name:        "nil template",
			tmpl:        nil,
			wantErr:     true,
			errContains: "nil template",
		},
		{
			name: "empty name",
			tmpl: &Template{
				Name:   "",
				Source: `type Test struct {}`,
			},
			wantErr:     true,
			errContains: "empty template name",
		},
		{
			name: "empty source",
			tmpl: &Template{
				Name:   "test",
				Source: "",
			},
			wantErr:     true,
			errContains: "empty template source",
		},
		{
			name: "invalid syntax",
			tmpl: &Template{
				Name:   "invalid",
				Source: `type {`,
			},
			wantErr:     true,
			errContains: "parse template",
		},
		{
			name: "no declarations",
			tmpl: &Template{
				Name:   "empty",
				Source: `// only a comment`,
			},
			wantErr:     true,
			errContains: "no declarations",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := NewEngine(nil)
			err := e.Register(tt.tmpl)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, tt.tmpl.AST)
				assert.NotNil(t, tt.tmpl.Params)
			}
		})
	}
}

func TestEngine_Get(t *testing.T) {
	e := NewEngine(nil)
	tmpl := &Template{
		Name:   "test",
		Source: `type Test struct {}`,
	}
	err := e.Register(tmpl)
	require.NoError(t, err)

	tests := []struct {
		name        string
		tmplName    string
		wantErr     bool
		errContains string
	}{
		{
			name:     "found",
			tmplName: "test",
			wantErr:  false,
		},
		{
			name:        "not found",
			tmplName:    "nonexistent",
			wantErr:     true,
			errContains: "template not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := e.Get(tt.tmplName)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tt.tmplName, got.Name)
			}
		})
	}
}

func TestEngine_Execute(t *testing.T) {
	e := NewEngine(nil)

	tmpl := &Template{
		Name: "simple",
		Source: `type TypeName struct {
			Field string
		}`,
	}
	err := e.Register(tmpl)
	require.NoError(t, err)

	tests := []struct {
		name    string
		tmpl    string
		data    map[string]string
		wantErr bool
		check   func(t *testing.T, node ast.Node)
	}{
		{
			name: "simple replacement",
			tmpl: "simple",
			data: map[string]string{
				"TypeName": "User",
			},
			wantErr: false,
			check: func(t *testing.T, node ast.Node) {
				genDecl, ok := node.(*ast.GenDecl)
				require.True(t, ok)
				ts, ok := genDecl.Specs[0].(*ast.TypeSpec)
				require.True(t, ok)
				assert.Equal(t, "User", ts.Name.Name)
			},
		},
		{
			name:    "no data",
			tmpl:    "simple",
			data:    map[string]string{},
			wantErr: false,
			check: func(t *testing.T, node ast.Node) {
				genDecl, ok := node.(*ast.GenDecl)
				require.True(t, ok)
				ts, ok := genDecl.Specs[0].(*ast.TypeSpec)
				require.True(t, ok)
				assert.Equal(t, "TypeName", ts.Name.Name)
			},
		},
		{
			name:    "template not found",
			tmpl:    "nonexistent",
			data:    map[string]string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			node, err := e.Execute(tt.tmpl, tt.data)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, node)
			} else {
				require.NoError(t, err)
				require.NotNil(t, node)
				if tt.check != nil {
					tt.check(t, node)
				}
			}
		})
	}
}

func TestEngine_ExecuteTemplate(t *testing.T) {
	e := NewEngine(nil)

	tmpl := &Template{
		Name: "func",
		Source: `func TypeName() {
			return
		}`,
	}
	err := e.Register(tmpl)
	require.NoError(t, err)

	tests := []struct {
		name    string
		tmpl    *Template
		data    map[string]string
		wantErr bool
		check   func(t *testing.T, node ast.Node)
	}{
		{
			name: "valid template",
			tmpl: tmpl,
			data: map[string]string{
				"TypeName": "hello",
			},
			wantErr: false,
			check: func(t *testing.T, node ast.Node) {
				fn, ok := node.(*ast.FuncDecl)
				require.True(t, ok)
				assert.Equal(t, "hello", fn.Name.Name)
			},
		},
		{
			name:    "nil template",
			tmpl:    nil,
			data:    map[string]string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			node, err := e.ExecuteTemplate(tt.tmpl, tt.data)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, node)
			} else {
				require.NoError(t, err)
				require.NotNil(t, node)
				if tt.check != nil {
					tt.check(t, node)
				}
			}
		})
	}
}

func TestEngine_RegisterBuiltIns(t *testing.T) {
	e := NewEngine(nil)
	err := e.RegisterBuiltIns()
	require.NoError(t, err)

	tests := []struct {
		name     string
		tmplName string
		wantErr  bool
	}{
		{
			name:     "struct template",
			tmplName: "struct",
			wantErr:  false,
		},
		{
			name:     "function template",
			tmplName: "function",
			wantErr:  false,
		},
		{
			name:     "method template",
			tmplName: "method",
			wantErr:  false,
		},
		{
			name:     "interface template",
			tmplName: "interface",
			wantErr:  false,
		},
		{
			name:     "nonexistent template",
			tmplName: "nonexistent",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpl, err := e.Get(tt.tmplName)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, tmpl)
				assert.NotNil(t, tmpl.AST)
			}
		})
	}
}

func TestPlaceholderReplacer_isPlaceholder(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]string
		testName string
		want     bool
	}{
		{
			name:     "is placeholder",
			data:     map[string]string{"TypeName": "test"},
			testName: "TypeName",
			want:     true,
		},
		{
			name:     "is not placeholder",
			data:     map[string]string{"TypeName": "test"},
			testName: "Name",
			want:     false,
		},
		{
			name:     "nil data",
			data:     nil,
			testName: "TypeName",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := &placeholderReplacer{data: tt.data}
			result := r.isPlaceholder(tt.testName)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestPlaceholderReplacer_getReplacement(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]string
		testName string
		want     string
	}{
		{
			name:     "get replacement",
			data:     map[string]string{"TypeName": "test"},
			testName: "TypeName",
			want:     "test",
		},
		{
			name:     "no replacement",
			data:     map[string]string{"TypeName": "test"},
			testName: "FieldType",
			want:     "",
		},
		{
			name:     "nil data",
			data:     nil,
			testName: "TypeName",
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := &placeholderReplacer{data: tt.data}
			result := r.getReplacement(tt.testName)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestExtractPlaceholders(t *testing.T) {
	tests := []struct {
		name         string
		src          string
		wantCount    int
		wantContains []string
	}{
		{
			name:      "no placeholders",
			src:       `type Test struct {}`,
			wantCount: 0,
		},
		{
			name:         "single placeholder",
			src:          `type TypeName struct {}`,
			wantCount:    1,
			wantContains: []string{"TypeName"},
		},
		{
			name: "multiple placeholders",
			src: `func FunctionName() TypeName {
				return DefaultValue
			}`,
			wantCount:    2,
			wantContains: []string{"FunctionName", "TypeName"},
		},
		{
			name: "with comments",
			src: `// This is a comment
			type TypeName struct {
				// Field comment
				FieldName string
			}`,
			wantCount:    2,
			wantContains: []string{"TypeName", "FieldName"},
		},
		{
			name: "duplicates removed",
			src: `type TypeName struct {
				TypeName string
			}`,
			wantCount:    1,
			wantContains: []string{"TypeName"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := extractParamNames(tt.src)
			assert.Len(t, result, tt.wantCount)

			for _, want := range tt.wantContains {
				assert.Contains(t, result, want)
			}
		})
	}
}

func TestGenerateTestScaffold(t *testing.T) {
	tests := []struct {
		name        string
		funcSrc     string
		wantErr     bool
		errContains string
		checkOutput func(t *testing.T, node ast.Node)
	}{
		{
			name: "simple function with error return",
			funcSrc: `package test
func CreateUser(email string) (*User, error) {
	return nil, nil
}`,
			wantErr: false,
			checkOutput: func(t *testing.T, node ast.Node) {
				fn, ok := node.(*ast.FuncDecl)
				require.True(t, ok)
				assert.Equal(t, "TestCreateUser", fn.Name.Name)
				assert.NotNil(t, fn.Body)
			},
		},
		{
			name: "function with multiple params",
			funcSrc: `package test
func Add(a, b int) int {
	return a + b
}`,
			wantErr: false,
			checkOutput: func(t *testing.T, node ast.Node) {
				fn, ok := node.(*ast.FuncDecl)
				require.True(t, ok)
				assert.Equal(t, "TestAdd", fn.Name.Name)
				assert.NotNil(t, fn.Body)
			},
		},
		{
			name: "function without returns",
			funcSrc: `package test
func Print(s string) {
	fmt.Println(s)
}`,
			wantErr: false,
			checkOutput: func(t *testing.T, node ast.Node) {
				fn, ok := node.(*ast.FuncDecl)
				require.True(t, ok)
				assert.Equal(t, "TestPrint", fn.Name.Name)
			},
		},
		{
			name:        "nil function",
			funcSrc:     "",
			wantErr:     true,
			errContains: "nil function",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var funcDecl *ast.FuncDecl
			if tt.funcSrc != "" {
				fset := token.NewFileSet()
				f, err := parser.ParseFile(fset, "", tt.funcSrc, parser.AllErrors)
				require.NoError(t, err)

				for _, decl := range f.Decls {
					if fn, ok := decl.(*ast.FuncDecl); ok {
						funcDecl = fn
						break
					}
				}
				require.NotNil(t, funcDecl)
			}

			node, err := GenerateTestScaffold(funcDecl)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, node)
				if tt.checkOutput != nil {
					tt.checkOutput(t, node)
				}
			}
		})
	}
}

func TestGenerateTestScaffold_TableDrivenStructure(t *testing.T) {
	funcSrc := `package test
func Greet(name string) string {
	return "Hello, " + name
}`

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", funcSrc, parser.AllErrors)
	require.NoError(t, err)

	var funcDecl *ast.FuncDecl
	for _, decl := range f.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			funcDecl = fn
			break
		}
	}
	require.NotNil(t, funcDecl)

	node, err := GenerateTestScaffold(funcDecl)
	require.NoError(t, err)

	fn, ok := node.(*ast.FuncDecl)
	require.True(t, ok)

	assert.Equal(t, "TestGreet", fn.Name.Name)
	assert.NotNil(t, fn.Body)

	testDecl, ok := fn.Body.List[0].(*ast.DeclStmt)
	require.True(t, ok)

	varDecl, ok := testDecl.Decl.(*ast.GenDecl)
	require.True(t, ok)
	require.Equal(t, token.VAR, varDecl.Tok)

	valueSpec, ok := varDecl.Specs[0].(*ast.ValueSpec)
	require.True(t, ok)
	assert.Equal(t, "tests", valueSpec.Names[0].Name)

	arrayType, ok := valueSpec.Type.(*ast.ArrayType)
	require.True(t, ok)

	structType, ok := arrayType.Elt.(*ast.StructType)
	require.True(t, ok)
	require.NotNil(t, structType.Fields)

	assert.Equal(t, "name", structType.Fields.List[0].Names[0].Name)
}
