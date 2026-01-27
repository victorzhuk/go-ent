package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// WithMetrics is a no-op wrapper that passes through to the handler
// Metrics collection has been removed in the simplification
func WithMetrics[In, Out any](toolName string, handler func(context.Context, *mcp.CallToolRequest, In) (*mcp.CallToolResult, Out, error)) func(context.Context, *mcp.CallToolRequest, In) (*mcp.CallToolResult, Out, error) {
	return handler
}
