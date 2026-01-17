package tools

//nolint:gosec // test file with necessary file operations

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/victorzhuk/go-ent/internal/metrics"
)

func TestMetricsShow_TableFormat(t *testing.T) {
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
		Format: "table",
	}

	result, _, err := metricsShowHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Content, 1)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")
	assert.Contains(t, textContent.Text, "Metrics Table")
	assert.Contains(t, textContent.Text, "Tool | Status | Duration | Tokens | Cost")
	assert.Contains(t, textContent.Text, "test_tool")
	assert.Contains(t, textContent.Text, "✓")
	assert.Contains(t, textContent.Text, "✗")
	assert.Contains(t, textContent.Text, "150")
}

func TestMetricsShow_JSONFormat(t *testing.T) {
	// t.Parallel()

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
		Format: "json",
	}

	result, _, err := metricsShowHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Content, 1)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")

	var metricsList []metrics.Metric
	err = json.Unmarshal([]byte(textContent.Text), &metricsList)
	require.NoError(t, err)
	assert.Len(t, metricsList, 1)
	assert.Equal(t, "test_tool", metricsList[0].ToolName)
	assert.Equal(t, 100, metricsList[0].TokensIn)
}

func TestMetricsShow_CSVFormat(t *testing.T) {
	// t.Parallel()

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
		Format: "csv",
	}

	result, _, err := metricsShowHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Content, 1)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")
	assert.Contains(t, textContent.Text, "SessionID")
	assert.Contains(t, textContent.Text, "ToolName")
	assert.Contains(t, textContent.Text, "TokensIn")
	assert.Contains(t, textContent.Text, "TokensOut")
	assert.Contains(t, textContent.Text, "session-1")
	assert.Contains(t, textContent.Text, "test_tool")
}

func TestMetricsShow_FilterByTool(t *testing.T) {
	// t.Parallel()

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
		Format:   "table",
	}

	result, _, err := metricsShowHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")
	assert.Contains(t, textContent.Text, "tool_a")
	assert.NotContains(t, textContent.Text, "tool_b")
}

func TestMetricsShow_FilterBySession(t *testing.T) {
	// t.Parallel()

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
		SessionID: "session-2",
		ToolName:  "tool_b",
		TokensIn:  200,
		TokensOut: 100,
		Duration:  200 * time.Millisecond,
		Success:   true,
		Timestamp: now,
	})

	input := MetricsShowInput{
		SessionID: "session-1",
		Format:    "table",
	}

	result, _, err := metricsShowHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")
	assert.Contains(t, textContent.Text, "tool_a")
	assert.NotContains(t, textContent.Text, "tool_b")
}

func TestMetricsShow_FilterByTimeRange(t *testing.T) {
	// t.Parallel()

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
		Timestamp: now.Add(-2 * time.Hour),
	})

	_ = store.Add(metrics.Metric{
		SessionID: "session-1",
		ToolName:  "test_tool",
		TokensIn:  200,
		TokensOut: 100,
		Duration:  200 * time.Millisecond,
		Success:   true,
		Timestamp: now,
	})

	input := MetricsShowInput{
		StartTime: now.Add(-1 * time.Hour).Format(time.RFC3339),
		EndTime:   now.Add(1 * time.Hour).Format(time.RFC3339),
		Format:    "table",
	}

	result, _, err := metricsShowHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")
	assert.Contains(t, textContent.Text, "200")
}

func TestMetricsShow_Limit(t *testing.T) {
	// t.Parallel()

	store, err := metrics.NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
	require.NoError(t, err)
	defer func() { _ = store.Close() }()

	InitMetricsStore(store)

	now := time.Now()

	for i := 0; i < 10; i++ {
		_ = store.Add(metrics.Metric{
			SessionID: "session-1",
			ToolName:  "test_tool",
			TokensIn:  100 + i,
			TokensOut: 50 + i,
			Duration:  time.Duration(100+i) * time.Millisecond,
			Success:   true,
			Timestamp: now,
		})
	}

	input := MetricsShowInput{
		Limit:  5,
		Format: "table",
	}

	result, _, err := metricsShowHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")
	assert.Contains(t, textContent.Text, "first 5 results")

	lines := strings.Split(textContent.Text, "\n")
	tableRows := 0
	for _, line := range lines {
		if strings.Contains(line, "test_tool") && strings.Contains(line, "|") {
			tableRows++
		}
	}
	assert.Equal(t, 5, tableRows)
}

func TestMetricsShow_LimitMax(t *testing.T) {
	// t.Parallel()

	store, err := metrics.NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
	require.NoError(t, err)
	defer func() { _ = store.Close() }()

	InitMetricsStore(store)

	now := time.Now()

	for i := 0; i < 1500; i++ {
		_ = store.Add(metrics.Metric{
			SessionID: "session-1",
			ToolName:  "test_tool",
			TokensIn:  100,
			TokensOut: 50,
			Duration:  100 * time.Millisecond,
			Success:   true,
			Timestamp: now,
		})
	}

	input := MetricsShowInput{
		Limit:  2000,
		Format: "table",
	}

	result, _, err := metricsShowHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")
	assert.Contains(t, textContent.Text, "first 1000 results")
}

func TestMetricsShow_LimitDefault(t *testing.T) {
	// t.Parallel()

	store, err := metrics.NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
	require.NoError(t, err)
	defer func() { _ = store.Close() }()

	InitMetricsStore(store)

	now := time.Now()

	for i := 0; i < 200; i++ {
		_ = store.Add(metrics.Metric{
			SessionID: "session-1",
			ToolName:  "test_tool",
			TokensIn:  100,
			TokensOut: 50,
			Duration:  100 * time.Millisecond,
			Success:   true,
			Timestamp: now,
		})
	}

	input := MetricsShowInput{
		Format: "table",
	}

	result, _, err := metricsShowHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")
	assert.Contains(t, textContent.Text, "first 100 results")
}

func TestMetricsShow_NoMetrics(t *testing.T) {
	// t.Parallel()

	store, err := metrics.NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
	require.NoError(t, err)
	defer func() { _ = store.Close() }()

	InitMetricsStore(store)

	input := MetricsShowInput{
		Format: "table",
	}

	result, _, err := metricsShowHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")
	assert.Contains(t, textContent.Text, "No metrics found")
}

func TestMetricsShow_NoStoreInitialized(t *testing.T) {
	// t.Parallel()

	metricsStore = nil

	input := MetricsShowInput{
		Format: "table",
	}

	result, _, err := metricsShowHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")
	assert.Contains(t, textContent.Text, "Metrics store not initialized")
}

func TestMetricsShow_CombinedFilters(t *testing.T) {
	// t.Parallel()

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
		Timestamp: now.Add(-2 * time.Hour),
	})

	_ = store.Add(metrics.Metric{
		SessionID: "session-2",
		ToolName:  "tool_a",
		TokensIn:  200,
		TokensOut: 100,
		Duration:  200 * time.Millisecond,
		Success:   true,
		Timestamp: now,
	})

	_ = store.Add(metrics.Metric{
		SessionID: "session-1",
		ToolName:  "tool_b",
		TokensIn:  150,
		TokensOut: 75,
		Duration:  150 * time.Millisecond,
		Success:   true,
		Timestamp: now,
	})

	_ = store.Add(metrics.Metric{
		SessionID: "session-1",
		ToolName:  "tool_a",
		TokensIn:  300,
		TokensOut: 150,
		Duration:  300 * time.Millisecond,
		Success:   true,
		Timestamp: now.Add(-3 * time.Hour),
	})

	input := MetricsShowInput{
		SessionID: "session-1",
		ToolName:  "tool_a",
		StartTime: now.Add(-4 * time.Hour).Format(time.RFC3339),
		EndTime:   now.Add(1 * time.Hour).Format(time.RFC3339),
		Format:    "table",
	}

	result, _, err := metricsShowHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")
	assert.Contains(t, textContent.Text, "tool_a")
	assert.NotContains(t, textContent.Text, "tool_b")
}

func TestMetricsShow_FormatDefault(t *testing.T) {
	// t.Parallel()

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

	input := MetricsShowInput{}

	result, _, err := metricsShowHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")
	assert.Contains(t, textContent.Text, "Metrics Table")
}

func TestMetricsSummary_GroupByNone(t *testing.T) {
	// t.Parallel()

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
		ToolName:  "tool_a",
		TokensIn:  50,
		TokensOut: 25,
		Duration:  50 * time.Millisecond,
		Success:   false,
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

	input := MetricsSummaryInput{
		GroupBy: "none",
		Format:  "table",
	}

	result, _, err := metricsSummaryHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")
	assert.Contains(t, textContent.Text, "Metrics Summary")
	assert.Contains(t, textContent.Text, "Count | Success")
	assert.Contains(t, textContent.Text, "|     3 |")
	assert.Contains(t, textContent.Text, "Percentiles")
	assert.Contains(t, textContent.Text, "P50")
	assert.Contains(t, textContent.Text, "P95")
}

func TestMetricsSummary_GroupByTool(t *testing.T) {
	// t.Parallel()

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
		ToolName:  "tool_a",
		TokensIn:  50,
		TokensOut: 25,
		Duration:  50 * time.Millisecond,
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

	input := MetricsSummaryInput{
		GroupBy: "tool",
		Format:  "table",
	}

	result, _, err := metricsSummaryHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")
	assert.Contains(t, textContent.Text, "Metrics Summary")
	assert.Contains(t, textContent.Text, "Tool | Count")
	assert.Contains(t, textContent.Text, "tool_a")
	assert.Contains(t, textContent.Text, "tool_b")
	assert.Contains(t, textContent.Text, "|     2 |")
	assert.Contains(t, textContent.Text, "|     1 |")
}

func TestMetricsSummary_GroupByHour(t *testing.T) {
	// t.Parallel()

	store, err := metrics.NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
	require.NoError(t, err)
	defer func() { _ = store.Close() }()

	InitMetricsStore(store)

	now := time.Now()

	hour := now.Truncate(time.Hour)

	_ = store.Add(metrics.Metric{
		SessionID: "session-1",
		ToolName:  "test_tool",
		TokensIn:  100,
		TokensOut: 50,
		Duration:  100 * time.Millisecond,
		Success:   true,
		Timestamp: hour,
	})

	_ = store.Add(metrics.Metric{
		SessionID: "session-1",
		ToolName:  "test_tool",
		TokensIn:  50,
		TokensOut: 25,
		Duration:  50 * time.Millisecond,
		Success:   false,
		Timestamp: hour.Add(time.Hour),
	})

	input := MetricsSummaryInput{
		GroupBy: "hour",
		Format:  "table",
	}

	result, _, err := metricsSummaryHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")
	assert.Contains(t, textContent.Text, "Metrics Summary")
	assert.Contains(t, textContent.Text, "Time | Count")
}

func TestMetricsSummary_GroupByDay(t *testing.T) {
	// t.Parallel()

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
		SessionID: "session-1",
		ToolName:  "test_tool",
		TokensIn:  50,
		TokensOut: 25,
		Duration:  50 * time.Millisecond,
		Success:   true,
		Timestamp: now.Add(24 * time.Hour),
	})

	input := MetricsSummaryInput{
		GroupBy: "day",
		Format:  "table",
	}

	result, _, err := metricsSummaryHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")
	assert.Contains(t, textContent.Text, "Metrics Summary")
	assert.Contains(t, textContent.Text, "Time | Count")
}

func TestMetricsSummary_GroupByWeek(t *testing.T) {
	// t.Parallel()

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

	input := MetricsSummaryInput{
		GroupBy: "week",
		Format:  "table",
	}

	result, _, err := metricsSummaryHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")
	assert.Contains(t, textContent.Text, "Metrics Summary")
	assert.Contains(t, textContent.Text, "Time | Count")
}

func TestMetricsSummary_JSONFormat(t *testing.T) {
	// t.Parallel()

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

	input := MetricsSummaryInput{
		GroupBy: "tool",
		Format:  "json",
	}

	result, _, err := metricsSummaryHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")

	var summary summaryResult
	err = json.Unmarshal([]byte(textContent.Text), &summary)
	require.NoError(t, err)
	assert.Len(t, summary.Groups, 1)
	assert.Equal(t, 1, summary.Count)
}

func TestMetricsSummary_FilterByTool(t *testing.T) {
	// t.Parallel()

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

	input := MetricsSummaryInput{
		GroupBy:  "none",
		ToolName: "tool_a",
		Format:   "table",
	}

	result, _, err := metricsSummaryHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")
	assert.Contains(t, textContent.Text, "|     1 |")
}

func TestMetricsSummary_FilterBySession(t *testing.T) {
	// t.Parallel()

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
		SessionID: "session-2",
		ToolName:  "tool_a",
		TokensIn:  200,
		TokensOut: 100,
		Duration:  200 * time.Millisecond,
		Success:   true,
		Timestamp: now,
	})

	input := MetricsSummaryInput{
		GroupBy:   "none",
		SessionID: "session-1",
		Format:    "table",
	}

	result, _, err := metricsSummaryHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")
	assert.Contains(t, textContent.Text, "|     1 |")
}

func TestMetricsSummary_FilterByTimeRange(t *testing.T) {
	// t.Parallel()

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
		Timestamp: now.Add(-2 * time.Hour),
	})

	_ = store.Add(metrics.Metric{
		SessionID: "session-1",
		ToolName:  "test_tool",
		TokensIn:  200,
		TokensOut: 100,
		Duration:  200 * time.Millisecond,
		Success:   true,
		Timestamp: now,
	})

	input := MetricsSummaryInput{
		GroupBy:   "none",
		StartTime: now.Add(-1 * time.Hour).Format(time.RFC3339),
		EndTime:   now.Add(1 * time.Hour).Format(time.RFC3339),
		Format:    "table",
	}

	result, _, err := metricsSummaryHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")
	assert.Contains(t, textContent.Text, "|     1 |")
}

func TestMetricsSummary_NoMetrics(t *testing.T) {
	// t.Parallel()

	store, err := metrics.NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
	require.NoError(t, err)
	defer func() { _ = store.Close() }()

	InitMetricsStore(store)

	input := MetricsSummaryInput{
		GroupBy: "none",
		Format:  "table",
	}

	result, _, err := metricsSummaryHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")
	assert.Contains(t, textContent.Text, "No metrics found")
}

func TestMetricsSummary_NoStoreInitialized(t *testing.T) {
	// t.Parallel()

	metricsStore = nil

	input := MetricsSummaryInput{
		GroupBy: "none",
		Format:  "table",
	}

	result, _, err := metricsSummaryHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")
	assert.Contains(t, textContent.Text, "Metrics store not initialized")
}

func TestMetricsSummary_DefaultValues(t *testing.T) {
	// t.Parallel()

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

	input := MetricsSummaryInput{}

	result, _, err := metricsSummaryHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")
	assert.Contains(t, textContent.Text, "Metrics Summary")
}

func TestMetricsExport_JSONWithFilename(t *testing.T) {
	// t.Parallel()

	tmpDir := t.TempDir()

	store, err := metrics.NewStore(tmpDir+"/metrics.json", 24*time.Hour)
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

	input := MetricsExportInput{
		Format:   "json",
		Filename: tmpDir + "/test_export",
	}

	result, _, err := metricsExportHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Content, 1)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")

	var exportResult map[string]any
	err = json.Unmarshal([]byte(textContent.Text), &exportResult)
	require.NoError(t, err)
	assert.Equal(t, true, exportResult["success"])
	assert.Contains(t, exportResult["filename"].(string), "test_export.json")
	assert.Equal(t, "json", exportResult["format"])
	assert.Equal(t, float64(1), exportResult["records"])
	assert.Contains(t, exportResult["path"].(string), "test_export.json")

	content, err := os.ReadFile(tmpDir + "/test_export.json")
	require.NoError(t, err)
	assert.Contains(t, string(content), "test_tool")
}

func TestMetricsExport_CSVWithTimestamp(t *testing.T) {
	// t.Parallel()

	tmpDir := t.TempDir()

	store, err := metrics.NewStore(tmpDir+"/metrics.json", 24*time.Hour)
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

	input := MetricsExportInput{
		Format: "csv",
	}

	result, _, err := metricsExportHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")

	var exportResult map[string]any
	err = json.Unmarshal([]byte(textContent.Text), &exportResult)
	require.NoError(t, err)
	assert.Equal(t, true, exportResult["success"])
	assert.Equal(t, "csv", exportResult["format"])
	assert.Equal(t, float64(1), exportResult["records"])
	assert.Contains(t, exportResult["filename"].(string), "metrics_")
	assert.Contains(t, exportResult["filename"].(string), ".csv")

	filename := exportResult["filename"].(string)
	content, err := os.ReadFile(filename)
	require.NoError(t, err)
	assert.Contains(t, string(content), "SessionID")
	assert.Contains(t, string(content), "test_tool")

	os.Remove(filename)
}

func TestMetricsExport_Prometheus(t *testing.T) {
	// t.Parallel()

	tmpDir := t.TempDir()

	store, err := metrics.NewStore(tmpDir+"/metrics.json", 24*time.Hour)
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

	input := MetricsExportInput{
		Format:   "prometheus",
		Filename: tmpDir + "/metrics",
	}

	result, _, err := metricsExportHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")

	var exportResult map[string]any
	err = json.Unmarshal([]byte(textContent.Text), &exportResult)
	require.NoError(t, err)
	assert.Equal(t, true, exportResult["success"])
	assert.Contains(t, exportResult["filename"].(string), "metrics.prom")
	assert.Equal(t, "prometheus", exportResult["format"])
	assert.Equal(t, float64(1), exportResult["records"])

	content, err := os.ReadFile(tmpDir + "/metrics.prom")
	require.NoError(t, err)
	assert.Contains(t, string(content), "tool_tokens_total")
	assert.Contains(t, string(content), "tool_duration_seconds")
	assert.Contains(t, string(content), "tool_success_total")
}

func TestMetricsExport_FilterByToolName(t *testing.T) {
	// t.Parallel()

	tmpDir := t.TempDir()

	store, err := metrics.NewStore(tmpDir+"/metrics.json", 24*time.Hour)
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

	input := MetricsExportInput{
		Format:   "json",
		Filename: tmpDir + "/filtered",
		ToolName: "tool_a",
	}

	result, _, err := metricsExportHandler(context.Background(), nil, input)
	require.NoError(t, err)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")

	var exportResult map[string]any
	err = json.Unmarshal([]byte(textContent.Text), &exportResult)
	require.NoError(t, err)
	assert.Equal(t, float64(1), exportResult["records"])

	content, err := os.ReadFile(tmpDir + "/filtered.json")
	require.NoError(t, err)
	assert.Contains(t, string(content), "tool_a")
	assert.NotContains(t, string(content), "tool_b")
}

func TestMetricsExport_FilterBySessionID(t *testing.T) {
	// t.Parallel()

	tmpDir := t.TempDir()

	store, err := metrics.NewStore(tmpDir+"/metrics.json", 24*time.Hour)
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

	input := MetricsExportInput{
		Format:    "json",
		Filename:  tmpDir + "/session_filtered",
		SessionID: "session-1",
	}

	result, _, err := metricsExportHandler(context.Background(), nil, input)
	require.NoError(t, err)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")

	var exportResult map[string]any
	err = json.Unmarshal([]byte(textContent.Text), &exportResult)
	require.NoError(t, err)
	assert.Equal(t, float64(1), exportResult["records"])

	content, err := os.ReadFile(tmpDir + "/session_filtered.json")
	require.NoError(t, err)
	var metricsList []metrics.Metric
	err = json.Unmarshal(content, &metricsList)
	require.NoError(t, err)
	assert.Equal(t, "session-1", metricsList[0].SessionID)
}

func TestMetricsExport_FilterByTimeRange(t *testing.T) {
	// t.Parallel()

	tmpDir := t.TempDir()

	store, err := metrics.NewStore(tmpDir+"/metrics.json", 24*time.Hour)
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
		Timestamp: now.Add(-2 * time.Hour),
	})

	_ = store.Add(metrics.Metric{
		SessionID: "session-1",
		ToolName:  "test_tool",
		TokensIn:  200,
		TokensOut: 100,
		Duration:  200 * time.Millisecond,
		Success:   true,
		Timestamp: now,
	})

	input := MetricsExportInput{
		Format:    "json",
		Filename:  tmpDir + "/time_filtered",
		StartTime: now.Add(-1 * time.Hour).Format(time.RFC3339),
		EndTime:   now.Add(1 * time.Hour).Format(time.RFC3339),
	}

	result, _, err := metricsExportHandler(context.Background(), nil, input)
	require.NoError(t, err)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")

	var exportResult map[string]any
	err = json.Unmarshal([]byte(textContent.Text), &exportResult)
	require.NoError(t, err)
	assert.Equal(t, float64(1), exportResult["records"])

	content, err := os.ReadFile(tmpDir + "/time_filtered.json")
	require.NoError(t, err)
	var metricsList []metrics.Metric
	err = json.Unmarshal(content, &metricsList)
	require.NoError(t, err)
	assert.Equal(t, 200, metricsList[0].TokensIn)
}

func TestMetricsExport_Limit(t *testing.T) {
	// t.Parallel()

	tmpDir := t.TempDir()

	store, err := metrics.NewStore(tmpDir+"/metrics.json", 24*time.Hour)
	require.NoError(t, err)
	defer func() { _ = store.Close() }()

	InitMetricsStore(store)

	now := time.Now()

	for i := 0; i < 10; i++ {
		_ = store.Add(metrics.Metric{
			SessionID: "session-1",
			ToolName:  "test_tool",
			TokensIn:  100 + i,
			TokensOut: 50 + i,
			Duration:  time.Duration(100+i) * time.Millisecond,
			Success:   true,
			Timestamp: now,
		})
	}

	input := MetricsExportInput{
		Format:   "json",
		Filename: tmpDir + "/limited",
		Limit:    5,
	}

	result, _, err := metricsExportHandler(context.Background(), nil, input)
	require.NoError(t, err)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")

	var exportResult map[string]any
	err = json.Unmarshal([]byte(textContent.Text), &exportResult)
	require.NoError(t, err)
	assert.Equal(t, float64(5), exportResult["records"])

	content, err := os.ReadFile(tmpDir + "/limited.json")
	require.NoError(t, err)
	var metricsList []metrics.Metric
	err = json.Unmarshal(content, &metricsList)
	require.NoError(t, err)
	assert.Len(t, metricsList, 5)
}

func TestMetricsExport_LimitMax(t *testing.T) {
	// t.Parallel()

	tmpDir := t.TempDir()

	store, err := metrics.NewStore(tmpDir+"/metrics.json", 24*time.Hour)
	require.NoError(t, err)
	defer func() { _ = store.Close() }()

	InitMetricsStore(store)

	now := time.Now()

	for i := 0; i < 15000; i++ {
		_ = store.Add(metrics.Metric{
			SessionID: "session-1",
			ToolName:  "test_tool",
			TokensIn:  100,
			TokensOut: 50,
			Duration:  100 * time.Millisecond,
			Success:   true,
			Timestamp: now,
		})
	}

	input := MetricsExportInput{
		Format:   "json",
		Filename: tmpDir + "/max_limit",
		Limit:    20000,
	}

	result, _, err := metricsExportHandler(context.Background(), nil, input)
	require.NoError(t, err)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")

	var exportResult map[string]any
	err = json.Unmarshal([]byte(textContent.Text), &exportResult)
	require.NoError(t, err)
	assert.Equal(t, float64(10000), exportResult["records"])
}

func TestMetricsExport_LimitDefault(t *testing.T) {
	// t.Parallel()

	store, err := metrics.NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
	require.NoError(t, err)
	defer func() { _ = store.Close() }()

	InitMetricsStore(store)

	now := time.Now()

	for i := 0; i < 2000; i++ {
		_ = store.Add(metrics.Metric{
			SessionID: "session-1",
			ToolName:  "test_tool",
			TokensIn:  100,
			TokensOut: 50,
			Duration:  100 * time.Millisecond,
			Success:   true,
			Timestamp: now,
		})
	}

	input := MetricsExportInput{
		Format:   "json",
		Filename: "default_limit",
	}

	result, _, err := metricsExportHandler(context.Background(), nil, input)
	require.NoError(t, err)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")

	var exportResult map[string]any
	err = json.Unmarshal([]byte(textContent.Text), &exportResult)
	require.NoError(t, err)
	assert.Equal(t, float64(1000), exportResult["records"])
}

func TestMetricsExport_ExportToSubdirectory(t *testing.T) {
	// t.Parallel()

	tmpDir := t.TempDir()

	store, err := metrics.NewStore(tmpDir+"/metrics.json", 24*time.Hour)
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

	input := MetricsExportInput{
		Format:   "json",
		Filename: tmpDir + "/exports/my_metrics",
	}

	result, _, err := metricsExportHandler(context.Background(), nil, input)
	require.NoError(t, err)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")

	var exportResult map[string]any
	err = json.Unmarshal([]byte(textContent.Text), &exportResult)
	require.NoError(t, err)
	assert.Equal(t, true, exportResult["success"])
	assert.Contains(t, exportResult["filename"].(string), "exports/my_metrics.json")

	content, err := os.ReadFile(tmpDir + "/exports/my_metrics.json")
	require.NoError(t, err)
	assert.Contains(t, string(content), "test_tool")
}

func TestMetricsExport_InvalidFormat(t *testing.T) {
	// t.Parallel()

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

	input := MetricsExportInput{
		Format: "xml",
	}

	result, _, err := metricsExportHandler(context.Background(), nil, input)
	require.Error(t, err)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")
	assert.Contains(t, textContent.Text, "invalid format")
}

func TestMetricsExport_NoMetricsFound(t *testing.T) {
	// t.Parallel()

	store, err := metrics.NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
	require.NoError(t, err)
	defer func() { _ = store.Close() }()

	InitMetricsStore(store)

	input := MetricsExportInput{
		Format:   "json",
		Filename: "empty",
	}

	result, _, err := metricsExportHandler(context.Background(), nil, input)
	require.NoError(t, err)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")
	assert.Contains(t, textContent.Text, "No metrics found")
}

func TestMetricsExport_NoStoreInitialized(t *testing.T) {
	// t.Parallel()

	metricsStore = nil

	input := MetricsExportInput{
		Format:   "json",
		Filename: "test",
	}

	result, _, err := metricsExportHandler(context.Background(), nil, input)
	require.NoError(t, err)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")
	assert.Contains(t, textContent.Text, "Metrics store not initialized")
}

func TestMetricsExport_CombinedFilters(t *testing.T) {
	// t.Parallel()

	tmpDir := t.TempDir()

	store, err := metrics.NewStore(tmpDir+"/metrics.json", 24*time.Hour)
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
		Timestamp: now.Add(-2 * time.Hour),
	})

	_ = store.Add(metrics.Metric{
		SessionID: "session-2",
		ToolName:  "tool_a",
		TokensIn:  200,
		TokensOut: 100,
		Duration:  200 * time.Millisecond,
		Success:   true,
		Timestamp: now,
	})

	_ = store.Add(metrics.Metric{
		SessionID: "session-1",
		ToolName:  "tool_b",
		TokensIn:  150,
		TokensOut: 75,
		Duration:  150 * time.Millisecond,
		Success:   true,
		Timestamp: now,
	})

	_ = store.Add(metrics.Metric{
		SessionID: "session-1",
		ToolName:  "tool_a",
		TokensIn:  300,
		TokensOut: 150,
		Duration:  300 * time.Millisecond,
		Success:   true,
		Timestamp: now.Add(-3 * time.Hour),
	})

	input := MetricsExportInput{
		Format:    "json",
		Filename:  tmpDir + "/combined_filtered",
		SessionID: "session-1",
		ToolName:  "tool_a",
		StartTime: now.Add(-4 * time.Hour).Format(time.RFC3339),
		EndTime:   now.Add(1 * time.Hour).Format(time.RFC3339),
	}

	result, _, err := metricsExportHandler(context.Background(), nil, input)
	require.NoError(t, err)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")

	var exportResult map[string]any
	err = json.Unmarshal([]byte(textContent.Text), &exportResult)
	require.NoError(t, err)
	assert.Equal(t, float64(1), exportResult["records"])

	content, err := os.ReadFile(tmpDir + "/combined_filtered.json")
	require.NoError(t, err)
	var metricsList []metrics.Metric
	err = json.Unmarshal(content, &metricsList)
	require.NoError(t, err)
	assert.Len(t, metricsList, 1)
	assert.Equal(t, "session-1", metricsList[0].SessionID)
	assert.Equal(t, "tool_a", metricsList[0].ToolName)
	assert.Equal(t, 100, metricsList[0].TokensIn)
}

func TestMetricsReset_Success(t *testing.T) {
	// t.Parallel()

	store, err := metrics.NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
	require.NoError(t, err)
	defer func() { _ = store.Close() }()

	InitMetricsStore(store)

	for i := 0; i < 5; i++ {
		_ = store.Add(metrics.Metric{
			SessionID: "test-session",
			ToolName:  "test_tool",
			TokensIn:  100 + i,
			TokensOut: 200 + i,
			Duration:  time.Duration(i+1) * time.Second,
			Success:   true,
			Timestamp: time.Now(),
		})
	}

	assert.Equal(t, 5, store.Count())

	result, _, err := metricsResetHandler(context.Background(), nil, struct{}{})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Content, 1)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")

	var resetResult map[string]any
	err = json.Unmarshal([]byte(textContent.Text), &resetResult)
	require.NoError(t, err)
	assert.Equal(t, true, resetResult["success"])
	assert.Equal(t, "All metrics cleared from store", resetResult["message"])
	assert.Equal(t, float64(5), resetResult["previous_count"])
	assert.Equal(t, float64(0), resetResult["current_count"])

	assert.Equal(t, 0, store.Count())
}

func TestMetricsReset_EmptyStore(t *testing.T) {
	// t.Parallel()

	store, err := metrics.NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
	require.NoError(t, err)
	defer func() { _ = store.Close() }()

	InitMetricsStore(store)

	assert.Equal(t, 0, store.Count())

	result, _, err := metricsResetHandler(context.Background(), nil, struct{}{})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Content, 1)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")

	var resetResult map[string]any
	err = json.Unmarshal([]byte(textContent.Text), &resetResult)
	require.NoError(t, err)
	assert.Equal(t, true, resetResult["success"])
	assert.Equal(t, float64(0), resetResult["previous_count"])
	assert.Equal(t, float64(0), resetResult["current_count"])

	assert.Equal(t, 0, store.Count())
}

func TestMetricsReset_StoreNil(t *testing.T) {
	// t.Parallel()

	metricsStore = nil

	result, _, err := metricsResetHandler(context.Background(), nil, struct{}{})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Content, 1)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")
	assert.Contains(t, textContent.Text, "Metrics store not initialized")
}
