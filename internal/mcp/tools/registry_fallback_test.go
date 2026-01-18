package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistryListHandler_FallbackToTasksMD(t *testing.T) {
	t.Parallel()

	// Use testdata directory
	testPath, err := filepath.Abs("testdata/fallback")
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	// Clean up any existing registry.db
	registryPath := filepath.Join(testPath, "openspec", "registry.db")
	os.Remove(registryPath)
	t.Cleanup(func() { os.Remove(registryPath) })

	// Ensure registry.db doesn't exist
	_, err = os.Stat(registryPath)
	if !os.IsNotExist(err) {
		t.Logf("Registry.db exists at %s, skipping fallback test", registryPath)
		t.Skip("Registry.db exists, skipping fallback test")
	}

	t.Run("fallback lists all tasks", func(t *testing.T) {
		input := RegistryListInput{Path: testPath}
		result, _, err := registryListHandler(ctx, req, input)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result.Content, 1)

		var data map[string]interface{}
		err = json.Unmarshal([]byte(result.Content[0].(*mcp.TextContent).Text), &data)
		require.NoError(t, err)

		total, ok := data["total"].(float64)
		require.True(t, ok)
		assert.Equal(t, float64(4), total, "Should have 4 tasks")

		tasks, ok := data["tasks"].([]interface{})
		require.True(t, ok)
		assert.Len(t, tasks, 4, "Should return all 4 tasks")

		note, ok := data["note"].(string)
		require.True(t, ok)
		assert.Contains(t, note, "tasks.md files (registry.db not found)", "Should indicate fallback mode")
	})

	t.Run("fallback with change_id filter", func(t *testing.T) {
		input := RegistryListInput{Path: testPath, ChangeID: "test-fallback"}
		result, _, err := registryListHandler(ctx, req, input)

		require.NoError(t, err)
		require.NotNil(t, result)

		var data map[string]interface{}
		err = json.Unmarshal([]byte(result.Content[0].(*mcp.TextContent).Text), &data)
		require.NoError(t, err)

		tasks, ok := data["tasks"].([]interface{})
		require.True(t, ok)
		assert.Len(t, tasks, 4, "Should return 4 tasks for test-fallback")
	})

	t.Run("fallback with status filter", func(t *testing.T) {
		input := RegistryListInput{Path: testPath, Status: "pending"}
		result, _, err := registryListHandler(ctx, req, input)

		require.NoError(t, err)
		require.NotNil(t, result)

		var data map[string]interface{}
		err = json.Unmarshal([]byte(result.Content[0].(*mcp.TextContent).Text), &data)
		require.NoError(t, err)

		tasks, ok := data["tasks"].([]interface{})
		require.True(t, ok)
		assert.Len(t, tasks, 4, "Should return 4 pending tasks")
	})

	t.Run("fallback with priority filter (no match in fallback mode)", func(t *testing.T) {
		// Note: In fallback mode, all tasks have default medium priority
		// Parsing priorities from comments is not implemented
		input := RegistryListInput{Path: testPath, Priority: "critical"}
		result, _, err := registryListHandler(ctx, req, input)

		require.NoError(t, err)
		require.NotNil(t, result)

		var data map[string]interface{}
		err = json.Unmarshal([]byte(result.Content[0].(*mcp.TextContent).Text), &data)
		require.NoError(t, err)

		tasks, ok := data["tasks"].([]interface{})
		require.True(t, ok)
		assert.Len(t, tasks, 0, "Should return 0 critical tasks (all are medium in fallback mode)")
	})

	t.Run("fallback with unblocked filter", func(t *testing.T) {
		input := RegistryListInput{Path: testPath, Unblocked: true}
		result, _, err := registryListHandler(ctx, req, input)

		require.NoError(t, err)
		require.NotNil(t, result)

		var data map[string]interface{}
		err = json.Unmarshal([]byte(result.Content[0].(*mcp.TextContent).Text), &data)
		require.NoError(t, err)

		tasks, ok := data["tasks"].([]interface{})
		require.True(t, ok)
		assert.Len(t, tasks, 1, "Should return 1 unblocked task (1.1)")
	})

	t.Run("fallback with limit", func(t *testing.T) {
		input := RegistryListInput{Path: testPath, Limit: 2}
		result, _, err := registryListHandler(ctx, req, input)

		require.NoError(t, err)
		require.NotNil(t, result)

		var data map[string]interface{}
		err = json.Unmarshal([]byte(result.Content[0].(*mcp.TextContent).Text), &data)
		require.NoError(t, err)

		tasks, ok := data["tasks"].([]interface{})
		require.True(t, ok)
		assert.Len(t, tasks, 2, "Should return only 2 tasks")
	})

	t.Run("fallback includes summary", func(t *testing.T) {
		input := RegistryListInput{Path: testPath}
		result, _, err := registryListHandler(ctx, req, input)

		require.NoError(t, err)
		require.NotNil(t, result)

		var data map[string]interface{}
		err = json.Unmarshal([]byte(result.Content[0].(*mcp.TextContent).Text), &data)
		require.NoError(t, err)

		summary, ok := data["summary"].(map[string]interface{})
		require.True(t, ok, "Should include summary")

		byStatus, ok := summary["by_status"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, float64(4), byStatus["pending"], "Should have 4 pending tasks")

		byPriority, ok := summary["by_priority"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, float64(4), byPriority["medium"], "Should have 4 medium priority tasks (default in fallback mode)")
	})
}

func TestRegistryNextHandler_FallbackToTasksMD(t *testing.T) {
	t.Parallel()

	testPath, err := filepath.Abs("testdata/fallback")
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	// Clean up any existing registry.db
	registryPath := filepath.Join(testPath, "openspec", "registry.db")
	os.Remove(registryPath)
	t.Cleanup(func() { os.Remove(registryPath) })

	// Ensure registry.db doesn't exist
	_, err = os.Stat(registryPath)
	if !os.IsNotExist(err) {
		t.Logf("Registry.db exists at %s, skipping fallback test", registryPath)
		t.Skip("Registry.db exists, skipping fallback test")
	}

	t.Run("fallback returns next task", func(t *testing.T) {
		input := RegistryNextInput{Path: testPath}
		result, _, err := registryNextHandler(ctx, req, input)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result.Content, 1)

		text := result.Content[0].(*mcp.TextContent).Text

		// Should contain JSON with next task recommendation
		assert.Contains(t, text, "\"recommended\"", "Should contain recommended task")
		assert.Contains(t, text, "test-fallback", "Should contain change ID")
		assert.Contains(t, text, "\"task_num\": \"1\"", "Should contain task number")

		// Should include note about fallback mode
		assert.Contains(t, text, "Note: Showing tasks from tasks.md files (registry.db not found)", "Should indicate fallback mode")
	})

	t.Run("fallback with multiple count", func(t *testing.T) {
		// Only 1 unblocked task exists, so no alternatives
		input := RegistryNextInput{Path: testPath, Count: 2}
		result, _, err := registryNextHandler(ctx, req, input)

		require.NoError(t, err)
		require.NotNil(t, result)

		text := result.Content[0].(*mcp.TextContent).Text

		// Should contain JSON with next task recommendation
		assert.Contains(t, text, "\"recommended\"", "Should contain recommended task")

		// Should include note about fallback mode
		assert.Contains(t, text, "Note: Showing tasks from tasks.md files (registry.db not found)", "Should indicate fallback mode")
	})

	t.Run("fallback includes note about missing registry", func(t *testing.T) {
		input := RegistryNextInput{Path: testPath}
		result, _, err := registryNextHandler(ctx, req, input)

		require.NoError(t, err)
		require.NotNil(t, result)

		text := result.Content[0].(*mcp.TextContent).Text
		assert.Contains(t, text, "tasks.md files (registry.db not found)", "Should indicate fallback mode")
	})
}

func TestRegistryUpdateHandler_ErrorOnMissingRegistry(t *testing.T) {
	t.Parallel()

	testPath, err := filepath.Abs("testdata/fallback")
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	// Clean up any existing registry.db
	registryPath := filepath.Join(testPath, "openspec", "registry.db")
	os.Remove(registryPath)
	t.Cleanup(func() { os.Remove(registryPath) })

	// Ensure registry.db doesn't exist
	_, err = os.Stat(registryPath)
	if !os.IsNotExist(err) {
		t.Skip("Registry.db exists, skipping fallback test")
	}

	input := RegistryUpdateInput{
		Path:   testPath,
		TaskID: "test-fallback/1.1",
		Status: "in_progress",
	}

	result, _, err := registryUpdateHandler(ctx, req, input)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Content, 1)

	text := result.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, text, "Registry not found", "Should indicate registry missing")
	assert.Contains(t, text, "registry_sync", "Should suggest running registry_sync")
}

func TestRegistryDepsHandler_ErrorOnMissingRegistry(t *testing.T) {
	t.Parallel()

	testPath, err := filepath.Abs("testdata/fallback")
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	// Clean up any existing registry.db
	registryPath := filepath.Join(testPath, "openspec", "registry.db")
	os.Remove(registryPath)
	t.Cleanup(func() { os.Remove(registryPath) })

	// Ensure registry.db doesn't exist
	_, err = os.Stat(registryPath)
	if !os.IsNotExist(err) {
		t.Skip("Registry.db exists, skipping fallback test")
	}

	input := RegistryDepsInput{
		Path:      testPath,
		TaskID:    "test-fallback/1.2",
		Operation: "show",
	}

	result, _, err := registryDepsHandler(ctx, req, input)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Content, 1)

	text := result.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, text, "Registry not found", "Should indicate registry missing")
	assert.Contains(t, text, "registry_sync", "Should suggest running registry_sync")
}

func TestRegistrySyncHandler_ErrorOnMissingRegistry(t *testing.T) {
	t.Parallel()

	testPath, err := filepath.Abs("testdata/fallback")
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	// Clean up any existing registry.db
	registryPath := filepath.Join(testPath, "openspec", "registry.db")
	os.Remove(registryPath)
	t.Cleanup(func() { os.Remove(registryPath) })

	// Ensure registry.db doesn't exist
	_, err = os.Stat(registryPath)
	if !os.IsNotExist(err) {
		t.Skip("Registry.db exists, skipping fallback test")
	}

	input := RegistrySyncInput{
		Path: testPath,
	}

	result, _, err := registrySyncHandler(ctx, req, input)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Content, 1)

	text := result.Content[0].(*mcp.TextContent).Text
	assert.Contains(t, text, "Registry not found", "Should indicate registry missing")
	assert.Contains(t, text, "registry_init", "Should suggest running registry_init")
}
