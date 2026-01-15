package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestASTParse_SingleFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	testCode := `package main

import "fmt"

type User struct {
	Name string
}

func greet(name string) string {
	return "Hello, " + name
}

func main() {
	u := User{Name: "Alice"}
	fmt.Println(greet(u.Name))
}
`

	err := os.WriteFile(testFile, []byte(testCode), 0644)
	require.NoError(t, err)

	input := ASTParseInput{
		File:             testFile,
		IncludePositions: true,
	}

	result, _, err := astParseHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Content, 1)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")
	assert.Contains(t, textContent.Text, "Package: main")
	assert.Contains(t, textContent.Text, "User")
	assert.Contains(t, textContent.Text, "greet")
	assert.Contains(t, textContent.Text, "main")
	assert.Contains(t, textContent.Text, "fmt")
}

func TestASTParse_WithoutPositions(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	testCode := `package main

func hello() {}
`

	err := os.WriteFile(testFile, []byte(testCode), 0644)
	require.NoError(t, err)

	input := ASTParseInput{
		File:             testFile,
		IncludePositions: false,
	}

	result, _, err := astParseHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "hello")
	assert.NotContains(t, textContent.Text, "line:")
	assert.NotContains(t, textContent.Text, "column:")
}

func TestASTParse_SyntaxError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	testCode := `package main

func broken( {
	// missing closing paren
}
`

	err := os.WriteFile(testFile, []byte(testCode), 0644)
	require.NoError(t, err)

	input := ASTParseInput{
		File:             testFile,
		IncludePositions: true,
	}

	result, _, err := astParseHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "error")
}

func TestASTParse_FileNotFound(t *testing.T) {
	t.Parallel()

	input := ASTParseInput{
		File: "/nonexistent/file.go",
	}

	result, _, err := astParseHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "Errors:")
	assert.Contains(t, textContent.Text, "not found")
}

func TestASTParse_EmptyFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	err := os.WriteFile(testFile, []byte(""), 0644)
	require.NoError(t, err)

	input := ASTParseInput{
		File:             testFile,
		IncludePositions: false,
	}

	result, _, err := astParseHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "error")
}

func TestASTParse_InterfaceType(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	testCode := `package main

type Reader interface {
	Read() ([]byte, error)
	Close() error
}

type FileWriter struct{}

func (f *FileWriter) Read() ([]byte, error) {
	return []byte("hello"), nil
}

func (f *FileWriter) Close() error {
	return nil
}
`

	err := os.WriteFile(testFile, []byte(testCode), 0644)
	require.NoError(t, err)

	input := ASTParseInput{
		File:             testFile,
		IncludePositions: false,
	}

	result, _, err := astParseHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "Reader")
	assert.Contains(t, textContent.Text, "FileWriter")
	assert.Contains(t, textContent.Text, "Read")
	assert.Contains(t, textContent.Text, "Close")
}

func TestASTParse_PackageLevel(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	testFiles := map[string]string{
		"user.go": `package service

type User struct {
	ID   string
	Name string
}

func (u *User) Save() error {
	return nil
}
`,
		"auth.go": `package service

type AuthService struct{}

func (a *AuthService) Login(username, password string) error {
	return nil
}
`,
	}

	for filename, content := range testFiles {
		err := os.WriteFile(filepath.Join(tmpDir, filename), []byte(content), 0644)
		require.NoError(t, err)
	}

	input := ASTParseInput{
		Package:          tmpDir,
		IncludePositions: false,
	}

	result, _, err := astParseHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "User")
	assert.Contains(t, textContent.Text, "AuthService")
	assert.Contains(t, textContent.Text, "Save")
	assert.Contains(t, textContent.Text, "Login")
}

func TestASTParse_ConstantsAndVariables(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	testCode := `package main

const MaxRetries = 3
const (
	EnvDev  = "dev"
	EnvProd = "prod"
)

var count int
var (
	name string
	age  int
)
`

	err := os.WriteFile(testFile, []byte(testCode), 0644)
	require.NoError(t, err)

	input := ASTParseInput{
		File:             testFile,
		IncludePositions: false,
	}

	result, _, err := astParseHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "MaxRetries")
	assert.Contains(t, textContent.Text, "EnvDev")
	assert.Contains(t, textContent.Text, "EnvProd")
	assert.Contains(t, textContent.Text, "count")
	assert.Contains(t, textContent.Text, "name")
	assert.Contains(t, textContent.Text, "age")
}

func TestASTParse_Channels(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	testCode := `package main

func producer() chan<- int {
	return make(chan<- int)
}

func consumer() <-chan int {
	return make(<-chan int)
}

func bidirectional() chan int {
	return make(chan int)
}
`

	err := os.WriteFile(testFile, []byte(testCode), 0644)
	require.NoError(t, err)

	input := ASTParseInput{
		File:             testFile,
		IncludePositions: false,
	}

	result, _, err := astParseHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "producer")
	assert.Contains(t, textContent.Text, "consumer")
	assert.Contains(t, textContent.Text, "bidirectional")
}
