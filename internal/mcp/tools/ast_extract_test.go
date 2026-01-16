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

func TestASTExtract_SimpleExtraction(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	testCode := `package main

func main() {
	x := 1
	y := 2
	z := x + y
	println(z)
}
`

	err := os.WriteFile(testFile, []byte(testCode), 0600)
	require.NoError(t, err)

	input := ASTExtractInput{
		File:    testFile,
		Line:    4,
		EndLine: 5,
		Name:    "calculate",
		DryRun:  false,
	}

	result, _, err := astExtractHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Content, 1)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")
	assert.Contains(t, textContent.Text, "Extracted function: calculate")
	assert.Contains(t, textContent.Text, "Line range: 4-5")
	assert.Contains(t, textContent.Text, "Extraction applied successfully")

	content, err := os.ReadFile(testFile) // #nosec G304 -- test file
	require.NoError(t, err)
	assert.Contains(t, string(content), "func calculate()")
	assert.Contains(t, string(content), "calculate()")
}

func TestASTExtract_DryRun(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	testCode := `package main

func main() {
	x := 1
	y := 2
	z := x + y
	println(z)
}
`

	err := os.WriteFile(testFile, []byte(testCode), 0600)
	require.NoError(t, err)

	input := ASTExtractInput{
		File:    testFile,
		Line:    4,
		EndLine: 5,
		Name:    "calculate",
		DryRun:  true,
	}

	result, _, err := astExtractHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "Extracted function: calculate")
	assert.Contains(t, textContent.Text, "Line range: 4-5")
	assert.Contains(t, textContent.Text, "Dry run: changes not applied")
}

func TestASTExtract_NoStatementsInRange(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	testCode := `package main

func main() {
	x := 1
}
`

	err := os.WriteFile(testFile, []byte(testCode), 0600)
	require.NoError(t, err)

	input := ASTExtractInput{
		File:    testFile,
		Line:    10,
		EndLine: 15,
		Name:    "extracted",
	}

	result, _, err := astExtractHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "Error:")
	assert.Contains(t, textContent.Text, "no statements in range")
}

func TestASTExtract_EmptyFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	err := os.WriteFile(testFile, []byte(""), 0600)
	require.NoError(t, err)

	input := ASTExtractInput{
		File:    testFile,
		Line:    1,
		EndLine: 5,
		Name:    "extracted",
	}

	result, _, err := astExtractHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "Error:")
}

func TestASTExtract_ExtractFromNestedBlock(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	testCode := `package main

func main() {
	if true {
		a := 1
		b := 2
		c := a + b
	}
}
`

	err := os.WriteFile(testFile, []byte(testCode), 0600)
	require.NoError(t, err)

	input := ASTExtractInput{
		File:    testFile,
		Line:    5,
		EndLine: 6,
		Name:    "sum",
	}

	result, _, err := astExtractHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.Contains(t, textContent.Text, "Error:")
}
