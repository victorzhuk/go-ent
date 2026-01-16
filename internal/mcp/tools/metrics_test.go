package tools

//nolint:gosec // test file with necessary file operations

import (
	"context"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/victorzhuk/go-ent/internal/metrics"
)

func TestMetricsShow_SummaryFormat(t *testing.T) {
	t.Parallel()

	store, err := metrics.NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
	require.NoError(t, err)
	defer func() { _ = store.Close() }()

	InitMetricsStore(store)

	now := time.Now()

	_ = store.Add(metrics.Metric{
		SessionID: "session-1",
		ToolName:  "test_tool",
		TokensIn:  100,
		TokensOut: 50,
		Duration:  100 * time.Millisecond,
		Success:   true,
		Timestamp: now.Add(-1 * time.Hour),
	})

	_ = store.Add(metrics.Metric{
		SessionID: "session-1",
		ToolName:  "test_tool",
		TokensIn:  200,
		TokensOut: 100,
		Duration:  200 * time.Millisecond,
		Success:   true,
		Timestamp: now.Add(-30 * time.Minute),
	})

	_ = store.Add(metrics.Metric{
		SessionID: "session-1",
		ToolName:  "test_tool",
		TokensIn:  50,
		TokensOut: 25,
		Duration:  50 * time.Millisecond,
		Success:   false,
		ErrorMsg:  "test error",
		Timestamp: now,
	})

	input := MetricsShowInput{
		Format: "summary",
	}

	result, _, err := metricsShowHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Content, 1)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")
	assert.Contains(t, textContent.Text, "Metrics Summary")
	assert.Contains(t, textContent.Text, "Total Entries: 3")
	assert.Contains(t, textContent.Text, "Average Input Tokens")
	assert.Contains(t, textContent.Text, "Average Duration")
	assert.Contains(t, textContent.Text, "Success Rate")
}

func TestMetricsShow_RawFormat(t *testing.T) {
	t.Parallel()

	store, err := metrics.NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
	require.NoError(t, err)
	defer func() { _ = store.Close() }()

	InitMetricsStore(store)

	now := time.Now()

	_ = store.Add(metrics.Metric{
		SessionID: "session-1",
		ToolName:  "test_tool",
		TokensIn:  100,
		TokensOut: 50,
		Duration:  100 * time.Millisecond,
		Success:   true,
		Timestamp: now,
	})

	input := MetricsShowInput{
		Format: "raw",
	}

	result, _, err := metricsShowHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Content, 1)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")
	assert.Contains(t, textContent.Text, "Raw Metrics")
	assert.Contains(t, textContent.Text, "test_tool")
	assert.Contains(t, textContent.Text, "session-1")
	assert.Contains(t, textContent.Text, "Tokens: 100 in / 50 out")
}

func TestMetricsShow_FilterByTool(t *testing.T) {
	t.Parallel()

	store, err := metrics.NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
	require.NoError(t, err)
	defer func() { _ = store.Close() }()

	InitMetricsStore(store)

	now := time.Now()

	_ = store.Add(metrics.Metric{
		SessionID: "session-1",
		ToolName:  "tool_a",
		TokensIn:  100,
		TokensOut: 50,
		Duration:  100 * time.Millisecond,
		Success:   true,
		Timestamp: now,
	})

	_ = store.Add(metrics.Metric{
		SessionID: "session-1",
		ToolName:  "tool_b",
		TokensIn:  200,
		TokensOut: 100,
		Duration:  200 * time.Millisecond,
		Success:   true,
		Timestamp: now,
	})

	input := MetricsShowInput{
		ToolName: "tool_a",
		Format:   "summary",
	}

	result, _, err := metricsShowHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")
	assert.Contains(t, textContent.Text, "Total Entries: 1")
	assert.Contains(t, textContent.Text, "Average Input Tokens")
}

func TestMetricsShow_FilterBySession(t *testing.T) {
	t.Parallel()

	store, err := metrics.NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
	require.NoError(t, err)
	defer func() { _ = store.Close() }()

	InitMetricsStore(store)

	now := time.Now()

	_ = store.Add(metrics.Metric{
		SessionID: "session-1",
		ToolName:  "test_tool",
		TokensIn:  100,
		TokensOut: 50,
		Duration:  100 * time.Millisecond,
		Success:   true,
		Timestamp: now,
	})

	_ = store.Add(metrics.Metric{
		SessionID: "session-2",
		ToolName:  "test_tool",
		TokensIn:  200,
		TokensOut: 100,
		Duration:  200 * time.Millisecond,
		Success:   true,
		Timestamp: now,
	})

	input := MetricsShowInput{
		SessionID: "session-1",
		Format:    "summary",
	}

	result, _, err := metricsShowHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")
	assert.Contains(t, textContent.Text, "Total Entries: 1")
}

func TestMetricsShow_NoMetrics(t *testing.T) {
	t.Parallel()

	store, err := metrics.NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
	require.NoError(t, err)
	defer func() { _ = store.Close() }()

	InitMetricsStore(store)

	input := MetricsShowInput{
		Format: "summary",
	}

	result, _, err := metricsShowHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")
	assert.Contains(t, textContent.Text, "No metrics available")
}

func TestMetricsShow_NoStoreInitialized(t *testing.T) {
	t.Parallel()

	metricsStore = nil

	input := MetricsShowInput{
		Format: "summary",
	}

	result, _, err := metricsShowHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")
	assert.Contains(t, textContent.Text, "Metrics store not initialized")
}
