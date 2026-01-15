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

func TestASTQuery(t *testing.T) {
	t.Parallel()

	t.Run("find functions by wildcard pattern", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()

		writeFile(t, tmpDir, "handler.go", `package main

func HandleRequest() error {
	return nil
}

func HandleResponse() string {
	return ""
}

func OtherFunc() int {
	return 0
}
`)

		input := ASTQueryInput{
			Package: tmpDir,
			Type:    "function",
			Pattern: "Handle*",
		}

		req := &mcp.CallToolRequest{}
		result, _, err := astQueryHandler(context.Background(), req, input)

		require.NoError(t, err)
		require.Len(t, result.Content, 1)

		textContent, ok := result.Content[0].(*mcp.TextContent)
		require.True(t, ok)

		assert.Contains(t, textContent.Text, "Found 2 matches")
		assert.Contains(t, textContent.Text, "HandleRequest")
		assert.Contains(t, textContent.Text, "HandleResponse")
		assert.NotContains(t, textContent.Text, "OtherFunc")
	})

	t.Run("find functions by exact name", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()

		writeFile(t, tmpDir, "service.go", `package main

func CreateUser() error {
	return nil
}

func UpdateUser() error {
	return nil
}
`)

		input := ASTQueryInput{
			Package: tmpDir,
			Type:    "function",
			Pattern: "CreateUser",
		}

		req := &mcp.CallToolRequest{}
		result, _, err := astQueryHandler(context.Background(), req, input)

		require.NoError(t, err)
		require.Len(t, result.Content, 1)

		textContent, ok := result.Content[0].(*mcp.TextContent)
		require.True(t, ok)

		assert.Contains(t, textContent.Text, "Found 1 match")
		assert.Contains(t, textContent.Text, "CreateUser")
		assert.NotContains(t, textContent.Text, "UpdateUser")
	})

	t.Run("find by signature", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()

		writeFile(t, tmpDir, "ops.go", `package main

import "context"

func WithContext(ctx context.Context) error {
	return nil
}

func WithContextAndStr(ctx context.Context, s string) error {
	return nil
}
`)

		input := ASTQueryInput{
			Package:   tmpDir,
			Type:      "function",
			Signature: "(context.Context) error",
		}

		req := &mcp.CallToolRequest{}
		result, _, err := astQueryHandler(context.Background(), req, input)

		require.NoError(t, err)
		require.Len(t, result.Content, 1)

		textContent, ok := result.Content[0].(*mcp.TextContent)
		require.True(t, ok)

		assert.Contains(t, textContent.Text, "Found 1 match")
		assert.Contains(t, textContent.Text, "WithContext")
		assert.NotContains(t, textContent.Text, "WithContextAndStr")
	})

	t.Run("find interface implementations", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()

		writeFile(t, tmpDir, "reader.go", `package main

type Reader interface {
	Read(p []byte) (int, error)
}

type MyReader struct{}

func (r *MyReader) Read(p []byte) (int, error) {
	return len(p), nil
}

type NotReader struct{}

func (n *NotReader) Write(p []byte) (int, error) {
	return len(p), nil
}
`)

		input := ASTQueryInput{
			Package:   tmpDir,
			Type:      "implements",
			Interface: "Reader",
		}

		req := &mcp.CallToolRequest{}
		result, _, err := astQueryHandler(context.Background(), req, input)

		require.NoError(t, err)
		require.Len(t, result.Content, 1)

		textContent, ok := result.Content[0].(*mcp.TextContent)
		require.True(t, ok)

		assert.Contains(t, textContent.Text, "Found 1 match")
		assert.Contains(t, textContent.Text, "MyReader")
		assert.Contains(t, textContent.Text, "Reader")
		assert.NotContains(t, textContent.Text, "NotReader")
	})

	t.Run("query single file", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()

		aPath := writeFile(t, tmpDir, "a.go", `package main
func AFunc() {}
`)
		writeFile(t, tmpDir, "b.go", `package main
func BFunc() {}
`)

		input := ASTQueryInput{
			File:    aPath,
			Type:    "function",
			Pattern: "*",
		}

		req := &mcp.CallToolRequest{}
		result, _, err := astQueryHandler(context.Background(), req, input)

		require.NoError(t, err)
		require.Len(t, result.Content, 1)

		textContent, ok := result.Content[0].(*mcp.TextContent)
		require.True(t, ok)

		assert.Contains(t, textContent.Text, "Found 1 match")
		assert.Contains(t, textContent.Text, "AFunc")
		assert.NotContains(t, textContent.Text, "BFunc")
	})

	t.Run("no matches found", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()

		writeFile(t, tmpDir, "empty.go", `package main
func OtherFunc() {}
`)

		input := ASTQueryInput{
			Package: tmpDir,
			Type:    "function",
			Pattern: "Handle*",
		}

		req := &mcp.CallToolRequest{}
		result, _, err := astQueryHandler(context.Background(), req, input)

		require.NoError(t, err)
		require.Len(t, result.Content, 1)

		textContent, ok := result.Content[0].(*mcp.TextContent)
		require.True(t, ok)

		assert.Contains(t, textContent.Text, "No matches found")
	})

	t.Run("missing file or package", func(t *testing.T) {
		t.Parallel()

		input := ASTQueryInput{}
		req := &mcp.CallToolRequest{}

		result, _, err := astQueryHandler(context.Background(), req, input)

		require.NoError(t, err)
		require.Len(t, result.Content, 1)

		textContent, ok := result.Content[0].(*mcp.TextContent)
		require.True(t, ok)
		assert.Contains(t, textContent.Text, "Error:")
		assert.Contains(t, textContent.Text, "either file or package must be specified")
	})

	t.Run("invalid query type", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()

		writeFile(t, tmpDir, "test.go", `package main
func TestFunc() {}
`)

		input := ASTQueryInput{
			Package: tmpDir,
			Type:    "invalid",
		}

		req := &mcp.CallToolRequest{}
		result, _, err := astQueryHandler(context.Background(), req, input)

		require.NoError(t, err)
		require.Len(t, result.Content, 1)

		textContent, ok := result.Content[0].(*mcp.TextContent)
		require.True(t, ok)
		assert.Contains(t, textContent.Text, "Error:")
		assert.Contains(t, textContent.Text, "invalid query type")
	})

	t.Run("file not found", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()

		input := ASTQueryInput{
			File: filepath.Join(tmpDir, "nonexistent.go"),
		}

		req := &mcp.CallToolRequest{}
		result, _, err := astQueryHandler(context.Background(), req, input)

		require.NoError(t, err)
		require.Len(t, result.Content, 1)

		textContent, ok := result.Content[0].(*mcp.TextContent)
		require.True(t, ok)

		assert.Contains(t, textContent.Text, "No matches found")
	})
}

func TestASTQuery_Scope(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	internalDir := filepath.Join(tmpDir, "internal", "service")
	require.NoError(t, os.MkdirAll(internalDir, 0755))

	writeFile(t, internalDir, "handler.go", `package service

func HandleInternal() error {
	return nil
}
`)

	cmdDir := filepath.Join(tmpDir, "cmd", "server")
	require.NoError(t, os.MkdirAll(cmdDir, 0755))

	writeFile(t, cmdDir, "main.go", `package main

func HandleCommand() error {
	return nil
}
`)

	t.Run("scope to internal directory", func(t *testing.T) {
		t.Parallel()

		input := ASTQueryInput{
			Package: filepath.Join(tmpDir, "internal", "..."),
			Type:    "function",
			Pattern: "Handle*",
		}

		req := &mcp.CallToolRequest{}
		result, _, err := astQueryHandler(context.Background(), req, input)

		require.NoError(t, err)
		require.Len(t, result.Content, 1)

		textContent, ok := result.Content[0].(*mcp.TextContent)
		require.True(t, ok)

		assert.Contains(t, textContent.Text, "Found 1 match")
		assert.Contains(t, textContent.Text, "HandleInternal")
		assert.NotContains(t, textContent.Text, "HandleCommand")
	})

	t.Run("entire project", func(t *testing.T) {
		t.Parallel()

		input := ASTQueryInput{
			Package: filepath.Join(tmpDir, "..."),
			Type:    "function",
			Pattern: "Handle*",
		}

		req := &mcp.CallToolRequest{}
		result, _, err := astQueryHandler(context.Background(), req, input)

		require.NoError(t, err)
		require.Len(t, result.Content, 1)

		textContent, ok := result.Content[0].(*mcp.TextContent)
		require.True(t, ok)

		assert.Contains(t, textContent.Text, "Found 2 matches")
		assert.Contains(t, textContent.Text, "HandleInternal")
		assert.Contains(t, textContent.Text, "HandleCommand")
	})
}

func writeFile(t *testing.T, dir, name, content string) string {
	path := filepath.Join(dir, name)
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)
	return path
}
