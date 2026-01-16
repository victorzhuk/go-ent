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

func TestASTRename_SingleFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	testCode := `package main

func greet(name string) string {
	return "Hello, " + name
}

func main() {
	msg := greet("Alice")
	_ = msg
}
`

	err := os.WriteFile(testFile, []byte(testCode), 0600)
	require.NoError(t, err)

	input := ASTRenameInput{
		File:    testFile,
		Line:    3,
		Column:  6,
		NewName: "welcome",
		DryRun:  true,
	}

	result, _, err := astRenameHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Content, 1)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "Symbol: greet")
	assert.Contains(t, textContent.Text, "function")
	assert.Contains(t, textContent.Text, "welcome")
	assert.Contains(t, textContent.Text, "Dry run")
}

func TestASTRename_MultiFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	mainFile := filepath.Join(tmpDir, "main.go")
	mainCode := `package main

func greet(name string) string {
	return "Hello, " + name
}

func main() {
	msg := greet("Alice")
	_ = msg
}
`

	otherFile := filepath.Join(tmpDir, "other.go")
	otherCode := `package main

func test() {
	result := greet("Bob")
	_ = result
}
`

	err := os.WriteFile(mainFile, []byte(mainCode), 0600)
	require.NoError(t, err)
	err = os.WriteFile(otherFile, []byte(otherCode), 0600)
	require.NoError(t, err)

	input := ASTRenameInput{
		File:    mainFile,
		Line:    3,
		Column:  6,
		NewName: "welcome",
		DryRun:  true,
	}

	result, _, err := astRenameHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "Symbol: greet")
	assert.Contains(t, textContent.Text, "function")
}

func TestASTRename_DryRun(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	testCode := `package main

func greet(name string) string {
	return "Hello, " + name
}

func main() {
	msg := greet("Alice")
	_ = msg
}
`

	err := os.WriteFile(testFile, []byte(testCode), 0600)
	require.NoError(t, err)

	originalContent, err := os.ReadFile(testFile)
	require.NoError(t, err)

	input := ASTRenameInput{
		File:    testFile,
		Line:    3,
		Column:  6,
		NewName: "welcome",
		DryRun:  true,
	}

	_, _, err = astRenameHandler(context.Background(), nil, input)
	require.NoError(t, err)

	contentAfterDryRun, err := os.ReadFile(testFile)
	require.NoError(t, err)
	assert.Equal(t, originalContent, contentAfterDryRun)
}

func TestASTRename_Conflict(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	testCode := `package main

func greet(name string) string {
	return "Hello, " + name
}

func welcome(name string) string {
	return "Hi, " + name
}

func main() {
	msg := greet("Alice")
	_ = msg
}
`

	err := os.WriteFile(testFile, []byte(testCode), 0600)
	require.NoError(t, err)

	input := ASTRenameInput{
		File:    testFile,
		Line:    3,
		Column:  6,
		NewName: "welcome",
		DryRun:  true,
	}

	result, _, err := astRenameHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "Conflicts")
	assert.Contains(t, textContent.Text, "not applied due to conflicts")
}

func TestASTRename_NoUsages(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	testCode := `package main

func unused() string {
	return "hello"
}
`

	err := os.WriteFile(testFile, []byte(testCode), 0600)
	require.NoError(t, err)

	input := ASTRenameInput{
		File:    testFile,
		Line:    3,
		Column:  6,
		NewName: "newfunc",
		DryRun:  true,
	}

	result, _, err := astRenameHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "Symbol: unused")
	assert.Contains(t, textContent.Text, "newfunc")
	assert.Contains(t, textContent.Text, "1 change(s)")
}

func TestASTRename_FilesOnly(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	mainFile := filepath.Join(tmpDir, "main.go")
	mainCode := `package main

func greet(name string) string {
	return "Hello, " + name
}

func main() {
	msg := greet("Alice")
	_ = msg
}
`

	otherFile := filepath.Join(tmpDir, "other.go")
	otherCode := `package main

func test() {
	result := greet("Bob")
	_ = result
}
`

	err := os.WriteFile(mainFile, []byte(mainCode), 0600)
	require.NoError(t, err)
	err = os.WriteFile(otherFile, []byte(otherCode), 0600)
	require.NoError(t, err)

	input := ASTRenameInput{
		File:      mainFile,
		Line:      3,
		Column:    6,
		NewName:   "welcome",
		FilesOnly: true,
	}

	result, _, err := astRenameHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "Symbol: greet")
	assert.Contains(t, textContent.Text, mainFile)
	assert.Contains(t, textContent.Text, otherFile)

	originalMain, err := os.ReadFile(mainFile)
	require.NoError(t, err)
	assert.Contains(t, string(originalMain), "greet")

	originalOther, err := os.ReadFile(otherFile)
	require.NoError(t, err)
	assert.Contains(t, string(originalOther), "greet")
}

func TestASTRename_FileNotFound(t *testing.T) {
	t.Parallel()

	input := ASTRenameInput{
		File:    "/nonexistent/file.go",
		Line:    1,
		Column:  1,
		NewName: "newfunc",
		DryRun:  true,
	}

	result, _, err := astRenameHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "Error:")
	assert.Contains(t, textContent.Text, "not found")
}

func TestASTRename_InvalidLine(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	testCode := `package main

func hello() {}
`

	err := os.WriteFile(testFile, []byte(testCode), 0600)
	require.NoError(t, err)

	input := ASTRenameInput{
		File:    testFile,
		Line:    0,
		Column:  1,
		NewName: "newfunc",
		DryRun:  true,
	}

	result, _, err := astRenameHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "Error:")
	assert.Contains(t, textContent.Text, "line must be greater than 0")
}

func TestASTRename_InvalidColumn(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	testCode := `package main

func hello() {}
`

	err := os.WriteFile(testFile, []byte(testCode), 0600)
	require.NoError(t, err)

	input := ASTRenameInput{
		File:    testFile,
		Line:    1,
		Column:  0,
		NewName: "newfunc",
		DryRun:  true,
	}

	result, _, err := astRenameHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "Error:")
	assert.Contains(t, textContent.Text, "column must be greater than 0")
}

func TestASTRename_EmptyNewName(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	testCode := `package main

func hello() {}
`

	err := os.WriteFile(testFile, []byte(testCode), 0600)
	require.NoError(t, err)

	input := ASTRenameInput{
		File:    testFile,
		Line:    1,
		Column:  1,
		NewName: "",
		DryRun:  true,
	}

	result, _, err := astRenameHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "Error:")
	assert.Contains(t, textContent.Text, "new_name is required")
}

func TestASTRename_SameName(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	testCode := `package main

func greet(name string) string {
	return "Hello, " + name
}

func main() {
	msg := greet("Alice")
	_ = msg
}
`

	err := os.WriteFile(testFile, []byte(testCode), 0600)
	require.NoError(t, err)

	input := ASTRenameInput{
		File:    testFile,
		Line:    3,
		Column:  6,
		NewName: "greet",
		DryRun:  true,
	}

	result, _, err := astRenameHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "No changes needed")
}

func TestASTRename_SymbolNotFound(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	testCode := `package main

func hello() {}
`

	err := os.WriteFile(testFile, []byte(testCode), 0600)
	require.NoError(t, err)

	input := ASTRenameInput{
		File:    testFile,
		Line:    1,
		Column:  1,
		NewName: "newfunc",
		DryRun:  true,
	}

	result, _, err := astRenameHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "Error:")
	assert.Contains(t, textContent.Text, "symbol not found")
}

func TestASTRename_Variable(t *testing.T) {
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

	input := ASTRenameInput{
		File:    testFile,
		Line:    4,
		Column:  2,
		NewName: "total",
		DryRun:  true,
	}

	result, _, err := astRenameHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "Symbol: count")
	assert.Contains(t, textContent.Text, "variable")
	assert.Contains(t, textContent.Text, "total")
}

func TestASTRename_Type(t *testing.T) {
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

	input := ASTRenameInput{
		File:    testFile,
		Line:    3,
		Column:  6,
		NewName: "Person",
		DryRun:  true,
	}

	result, _, err := astRenameHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "Symbol: User")
	assert.Contains(t, textContent.Text, "type")
	assert.Contains(t, textContent.Text, "Person")
}
