package tools

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/metrics"
)

func WithMetrics[In, Out any](toolName string, handler func(context.Context, *mcp.CallToolRequest, In) (*mcp.CallToolResult, Out, error)) func(context.Context, *mcp.CallToolRequest, In) (*mcp.CallToolResult, Out, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input In) (*mcp.CallToolResult, Out, error) {
		if !IsMetricsEnabled() {
			return handler(ctx, req, input)
		}

		if metricsStore == nil {
			return handler(ctx, req, input)
		}

		sessionID := getSessionID(ctx)
		if sessionID == "" {
			sessionID = uuid.Must(uuid.NewV7()).String()
			ctx = context.WithValue(ctx, sessionContextKey, sessionID)
		}

		slog.Debug("collecting metrics",
			"tool", toolName,
			"session_id", sessionID,
		)

		start := time.Now()

		result, meta, err := handler(ctx, req, input)

		duration := time.Since(start)

		metric := metrics.Metric{
			SessionID: sessionID,
			ToolName:  toolName,
			TokensIn:  estimateTokens(req),
			TokensOut: estimateOutputTokens(result),
			Duration:  duration,
			Success:   err == nil,
			ErrorMsg:  getErrorMessage(err),
			Timestamp: time.Now(),
		}

		if err := metricsStore.Add(metric); err != nil {
			slog.Warn("failed to add metric",
				"tool", toolName,
				"session_id", sessionID,
				"error", err,
			)
		}

		return result, meta, err
	}
}

type contextKey string

const sessionContextKey contextKey = "session"

func getSessionID(ctx context.Context) string {
	if sid, ok := ctx.Value(sessionContextKey).(string); ok && sid != "" {
		return sid
	}
	return ""
}

func estimateTokens(req *mcp.CallToolRequest) int {
	return 0
}

func estimateOutputTokens(result *mcp.CallToolResult) int {
	if result == nil || len(result.Content) == 0 {
		return 0
	}

	count := 0
	for _, c := range result.Content {
		if text, ok := c.(*mcp.TextContent); ok {
			count += len(text.Text) / 4
		}
	}

	return count
}

func getErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
