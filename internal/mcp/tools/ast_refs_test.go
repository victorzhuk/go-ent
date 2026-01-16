package tools

//nolint:gosec // test file with necessary file operations

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestASTRefs_FunctionReferences(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	testCode := `package main

func greet(name string) string {
	return "Hello, " + name
}

func main() {
	msg := greet("Alice")
	_ = greet("Bob")
}
`

	err := os.WriteFile(testFile, []byte(testCode), 0600)
	require.NoError(t, err)

	input := ASTRefsInput{
		File:         testFile,
		Line:         3,
		Column:       6,
		IncludeTests: false,
	}

	result, _, err := astRefsHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Content, 1)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")
	assert.Contains(t, textContent.Text, "Symbol: greet")
	assert.Contains(t, textContent.Text, "function")
	assert.Contains(t, textContent.Text, "Definition:")
	assert.Contains(t, textContent.Text, "[definition]")
	assert.Contains(t, textContent.Text, "Found 3 reference(s)")
}

func TestASTRefs_VariableReferences(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	testCode := `package main

func main() {
	count := 0
	count = count + 1
	_ = count
}
`

	err := os.WriteFile(testFile, []byte(testCode), 0600)
	require.NoError(t, err)

	input := ASTRefsInput{
		File:         testFile,
		Line:         4,
		Column:       2,
		IncludeTests: false,
	}

	result, _, err := astRefsHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "Symbol: count")
	assert.Contains(t, textContent.Text, "variable")
}

func TestASTRefs_TypeReferences(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	testCode := `package main

type User struct {
	Name string
}

func process(u User) {
	var x User
}
`

	err := os.WriteFile(testFile, []byte(testCode), 0600)
	require.NoError(t, err)

	input := ASTRefsInput{
		File:         testFile,
		Line:         3,
		Column:       6,
		IncludeTests: false,
	}

	result, _, err := astRefsHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "Symbol: User")
	assert.Contains(t, textContent.Text, "type")
}

func TestASTRefs_FileNotFound(t *testing.T) {
	t.Parallel()

	input := ASTRefsInput{
		File:         "/nonexistent/file.go",
		Line:         1,
		Column:       1,
		IncludeTests: false,
	}

	result, _, err := astRefsHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "Error:")
	assert.Contains(t, textContent.Text, "not found")
}

func TestASTRefs_SymbolNotFound(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	testCode := `package main

func hello() {}
`

	err := os.WriteFile(testFile, []byte(testCode), 0600)
	require.NoError(t, err)

	input := ASTRefsInput{
		File:         testFile,
		Line:         1,
		Column:       1,
		IncludeTests: false,
	}

	result, _, err := astRefsHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "Error:")
	assert.Contains(t, textContent.Text, "not found")
}

func TestASTRefs_InvalidLine(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	testCode := `package main

func hello() {}
`

	err := os.WriteFile(testFile, []byte(testCode), 0600)
	require.NoError(t, err)

	input := ASTRefsInput{
		File:         testFile,
		Line:         0,
		Column:       1,
		IncludeTests: false,
	}

	result, _, err := astRefsHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "Error:")
	assert.Contains(t, textContent.Text, "line must be greater than 0")
}

func TestASTRefs_InvalidColumn(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	testCode := `package main

func hello() {}
`

	err := os.WriteFile(testFile, []byte(testCode), 0600)
	require.NoError(t, err)

	input := ASTRefsInput{
		File:         testFile,
		Line:         1,
		Column:       0,
		IncludeTests: false,
	}

	result, _, err := astRefsHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "Error:")
	assert.Contains(t, textContent.Text, "column must be greater than 0")
}

func TestASTRefs_EmptyFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	err := os.WriteFile(testFile, []byte(""), 0600)
	require.NoError(t, err)

	input := ASTRefsInput{
		File:         testFile,
		Line:         1,
		Column:       1,
		IncludeTests: false,
	}

	result, _, err := astRefsHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "Error:")
}

func TestASTRefs_ConstantReferences(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	testCode := `package main

const MaxSize = 100

func process() {
	x := MaxSize
	y := MaxSize * 2
}
`

	err := os.WriteFile(testFile, []byte(testCode), 0600)
	require.NoError(t, err)

	input := ASTRefsInput{
		File:         testFile,
		Line:         3,
		Column:       7,
		IncludeTests: false,
	}

	result, _, err := astRefsHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "Symbol: MaxSize")
	assert.Contains(t, textContent.Text, "constant")
	assert.Contains(t, textContent.Text, "Found 3 reference(s)")
}
