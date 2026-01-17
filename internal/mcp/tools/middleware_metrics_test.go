package tools

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/victorzhuk/go-ent/internal/metrics"
)

func TestWithMetrics_Enabled(t *testing.T) {
	oldStore := metricsStore
	oldEnabled := metricsEnabled
	defer func() {
		metricsStore = oldStore
		metricsEnabled = oldEnabled
	}()
	t.Run("metrics enabled and store initialized", func(t *testing.T) {
		store, err := metrics.NewStore("testdata/metrics.json", 24)
		if err != nil {
			t.Fatalf("create store: %v", err)
		}
		defer store.Close()

		InitMetricsStore(store)
		SetMetricsEnabled(true)

		called := false
		handler := func(ctx context.Context, req *mcp.CallToolRequest, input string) (*mcp.CallToolResult, string, error) {
			called = true
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "ok"}}}, "meta", nil
		}

		wrapped := WithMetrics("test_tool", handler)
		ctx := context.Background()
		req := &mcp.CallToolRequest{}
		result, meta, err := wrapped(ctx, req, "input")

		assert.True(t, called, "handler should be called")
		assert.NotNil(t, result, "result should not be nil")
		assert.Equal(t, "meta", meta)
		assert.NoError(t, err)

		metricsList := store.GetAll()
		assert.Greater(t, len(metricsList), 0, "should have collected metrics")
	})

	t.Run("metrics enabled but store not initialized", func(t *testing.T) {
		InitMetricsStore(nil)
		SetMetricsEnabled(true)

		called := false
		handler := func(ctx context.Context, req *mcp.CallToolRequest, input string) (*mcp.CallToolResult, string, error) {
			called = true
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "ok"}}}, "meta", nil
		}

		wrapped := WithMetrics("test_tool", handler)
		ctx := context.Background()
		req := &mcp.CallToolRequest{}
		result, meta, err := wrapped(ctx, req, "input")

		assert.True(t, called, "handler should be called")
		assert.NotNil(t, result, "result should not be nil")
		assert.Equal(t, "meta", meta)
		assert.NoError(t, err)
	})
}

func TestWithMetrics_Disabled(t *testing.T) {
	oldStore := metricsStore
	oldEnabled := metricsEnabled
	defer func() {
		metricsStore = oldStore
		metricsEnabled = oldEnabled
	}()
	t.Run("metrics disabled", func(t *testing.T) {
		store, err := metrics.NewStore("testdata/metrics_disabled.json", 24)
		if err != nil {
			t.Fatalf("create store: %v", err)
		}
		defer store.Close()

		InitMetricsStore(store)
		SetMetricsEnabled(false)

		called := false
		handler := func(ctx context.Context, req *mcp.CallToolRequest, input string) (*mcp.CallToolResult, string, error) {
			called = true
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "ok"}}}, "meta", nil
		}

		wrapped := WithMetrics("test_tool", handler)
		ctx := context.Background()
		req := &mcp.CallToolRequest{}
		result, meta, err := wrapped(ctx, req, "input")

		assert.True(t, called, "handler should still be called when metrics disabled")
		assert.NotNil(t, result, "result should not be nil")
		assert.Equal(t, "meta", meta)
		assert.NoError(t, err)

		metricsList := store.GetAll()
		assert.Equal(t, 0, len(metricsList), "should not have collected metrics when disabled")
	})

	t.Run("metrics disabled and store not initialized", func(t *testing.T) {
		InitMetricsStore(nil)
		SetMetricsEnabled(false)

		called := false
		handler := func(ctx context.Context, req *mcp.CallToolRequest, input string) (*mcp.CallToolResult, string, error) {
			called = true
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "ok"}}}, "meta", nil
		}

		wrapped := WithMetrics("test_tool", handler)
		ctx := context.Background()
		req := &mcp.CallToolRequest{}
		result, meta, err := wrapped(ctx, req, "input")

		assert.True(t, called, "handler should still be called")
		assert.NotNil(t, result, "result should not be nil")
		assert.Equal(t, "meta", meta)
		assert.NoError(t, err)
	})
}

func TestIsMetricsEnabled(t *testing.T) {
	oldEnabled := metricsEnabled
	defer func() {
		metricsEnabled = oldEnabled
	}()

	SetMetricsEnabled(true)
	assert.True(t, IsMetricsEnabled(), "should be enabled")

	SetMetricsEnabled(false)
	assert.False(t, IsMetricsEnabled(), "should be disabled")
}
