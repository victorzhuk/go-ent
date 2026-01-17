package tools

import (
	"context"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
	"github.com/victorzhuk/go-ent/internal/metrics"
)

func TestManualFilterDebug(t *testing.T) {
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

	count := store.Count()
	t.Logf("Store count: %d", count)

	all := store.GetAll()
	t.Logf("GetAll count: %d", len(all))
	for _, m := range all {
		t.Logf("  Tool: %s, Session: %s", m.ToolName, m.SessionID)
	}

	filtered := store.Filter(func(m metrics.Metric) bool {
		return m.ToolName == "tool_a"
	})
	t.Logf("Filtered count: %d", len(filtered))
	for _, m := range filtered {
		t.Logf("  Tool: %s, Session: %s", m.ToolName, m.SessionID)
	}

	input := MetricsShowInput{
		ToolName: "tool_a",
		Format:   "table",
	}

	result, _, err := metricsShowHandler(context.Background(), nil, input)
	require.NoError(t, err)
	require.NotNil(t, result)

	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")
	t.Logf("Result:\n%s", textContent.Text)
}
