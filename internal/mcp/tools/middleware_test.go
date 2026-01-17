package tools

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/metrics"
)

func TestWithMetrics(t *testing.T) {
	t.Run("calls handler when metrics store is nil", func(t *testing.T) {
		metricsStore = nil

		called := false
		baseHandler := func(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, any, error) {
			called = true
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "ok"}}}, nil, nil
		}

		wrapped := WithMetrics("test_tool", baseHandler)
		result, _, err := wrapped(context.Background(), &mcp.CallToolRequest{}, struct{}{})

		if !called {
			t.Error("handler not called")
		}
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("result is nil")
		}
	})

	t.Run("stores metric on success", func(t *testing.T) {
		store, err := metrics.NewStore(t.TempDir()+"/metrics.json", time.Hour)
		if err != nil {
			t.Fatalf("create store: %v", err)
		}
		defer store.Close()

		metricsStore = store

		baseHandler := func(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, any, error) {
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "ok"}}}, nil, nil
		}

		wrapped := WithMetrics("test_tool", baseHandler)
		wrapped(context.Background(), &mcp.CallToolRequest{}, struct{}{})

		metricsList := store.GetAll()
		if len(metricsList) != 1 {
			t.Errorf("expected 1 metric, got %d", len(metricsList))
		}
		if metricsList[0].ToolName != "test_tool" {
			t.Errorf("expected tool name 'test_tool', got '%s'", metricsList[0].ToolName)
		}
		if !metricsList[0].Success {
			t.Error("expected success to be true")
		}
		if metricsList[0].Duration == 0 {
			t.Error("expected duration > 0")
		}
	})

	t.Run("stores metric on error", func(t *testing.T) {
		store, err := metrics.NewStore(t.TempDir()+"/metrics.json", time.Hour)
		if err != nil {
			t.Fatalf("create store: %v", err)
		}
		defer store.Close()

		metricsStore = store

		testErr := errors.New("test error")
		baseHandler := func(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, any, error) {
			return nil, nil, testErr
		}

		wrapped := WithMetrics("test_tool", baseHandler)
		wrapped(context.Background(), &mcp.CallToolRequest{}, struct{}{})

		metricsList := store.GetAll()
		if len(metricsList) != 1 {
			t.Errorf("expected 1 metric, got %d", len(metricsList))
		}
		if metricsList[0].Success {
			t.Error("expected success to be false")
		}
		if metricsList[0].ErrorMsg != testErr.Error() {
			t.Errorf("expected error message '%s', got '%s'", testErr.Error(), metricsList[0].ErrorMsg)
		}
	})

	t.Run("propagates context with session ID", func(t *testing.T) {
		store, err := metrics.NewStore(t.TempDir()+"/metrics.json", time.Hour)
		if err != nil {
			t.Fatalf("create store: %v", err)
		}
		defer store.Close()

		metricsStore = store

		var capturedSessionID string
		baseHandler := func(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, any, error) {
			sid := getSessionID(ctx)
			capturedSessionID = sid
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "ok"}}}, nil, nil
		}

		ctx := context.Background()
		ctxWithSession := context.WithValue(ctx, sessionContextKey, "test-session-123")
		wrapped := WithMetrics("test_tool", baseHandler)
		wrapped(ctxWithSession, &mcp.CallToolRequest{}, struct{}{})

		if capturedSessionID != "test-session-123" {
			t.Errorf("expected session ID 'test-session-123', got '%s'", capturedSessionID)
		}
	})

	t.Run("handler succeeds even when store fails", func(t *testing.T) {
		store, err := metrics.NewStore(t.TempDir()+"/metrics.json", time.Hour)
		if err != nil {
			t.Fatalf("create store: %v", err)
		}

		metricsStore = store

		handlerCalled := false
		baseHandler := func(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, any, error) {
			handlerCalled = true
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "ok"}}}, "meta", nil
		}

		wrapped := WithMetrics("test_tool", baseHandler)

		store.Close()

		result, meta, err := wrapped(context.Background(), &mcp.CallToolRequest{}, struct{}{})

		if !handlerCalled {
			t.Error("handler not called when metrics store failed")
		}
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result == nil {
			t.Error("result is nil")
		}
		if meta != "meta" {
			t.Errorf("expected meta 'meta', got %v", meta)
		}
	})

	t.Run("handler error still returned when metrics fail", func(t *testing.T) {
		store, err := metrics.NewStore(t.TempDir()+"/metrics.json", time.Hour)
		if err != nil {
			t.Fatalf("create store: %v", err)
		}

		metricsStore = store

		testErr := errors.New("handler error")
		handlerCalled := false
		baseHandler := func(ctx context.Context, req *mcp.CallToolRequest, input struct{}) (*mcp.CallToolResult, any, error) {
			handlerCalled = true
			return nil, nil, testErr
		}

		wrapped := WithMetrics("test_tool", baseHandler)

		store.Close()

		result, _, err := wrapped(context.Background(), &mcp.CallToolRequest{}, struct{}{})

		if !handlerCalled {
			t.Error("handler not called")
		}
		if err != testErr {
			t.Errorf("expected error %v, got %v", testErr, err)
		}
		if result != nil {
			t.Error("expected nil result on handler error")
		}
	})
}
