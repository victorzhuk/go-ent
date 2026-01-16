package tools

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/metrics"
)

var metricsStore *metrics.Store

// InitMetricsStore initializes the metrics store for tools.
// This is called during MCP server initialization.
func InitMetricsStore(store *metrics.Store) {
	metricsStore = store
}

type MetricsShowInput struct {
	SessionID string `json:"session_id,omitempty"`
	ToolName  string `json:"tool_name,omitempty"`
	StartTime string `json:"start_time,omitempty"`
	EndTime   string `json:"end_time,omitempty"`
	GroupBy   string `json:"group_by,omitempty"`
	Format    string `json:"format,omitempty"`
}

func registerMetricsShow(s *mcp.Server) {
	tool := &mcp.Tool{
		Name:        "metrics_show",
		Description: "Query and display metrics for tools, sessions, or time periods",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"session_id": map[string]any{
					"type":        "string",
					"description": "Filter by session ID",
				},
				"tool_name": map[string]any{
					"type":        "string",
					"description": "Filter by tool name",
				},
				"start_time": map[string]any{
					"type":        "string",
					"description": "Start time (ISO 8601 format, e.g., 2026-01-15T10:00:00Z)",
				},
				"end_time": map[string]any{
					"type":        "string",
					"description": "End time (ISO 8601 format, e.g., 2026-01-15T18:00:00Z)",
				},
				"group_by": map[string]any{
					"type":        "string",
					"description": "Group results by: none, tool, session, hour, day, week",
					"enum":        []string{"none", "tool", "session", "hour", "day", "week"},
				},
				"format": map[string]any{
					"type":        "string",
					"description": "Output format: summary, raw, table",
					"enum":        []string{"summary", "raw", "table"},
				},
			},
		},
	}

	mcp.AddTool(s, tool, metricsShowHandler)
}

func metricsShowHandler(_ context.Context, _ *mcp.CallToolRequest, input MetricsShowInput) (*mcp.CallToolResult, any, error) {
	if metricsStore == nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Metrics store not initialized"}},
		}, nil, nil
	}

	format := input.Format
	if format == "" {
		format = "summary"
	}

	groupBy := input.GroupBy
	if groupBy == "" {
		groupBy = "none"
	}

	var filter metrics.Filter

	if input.SessionID != "" {
		sessionFilter := func(m metrics.Metric) bool {
			return m.SessionID == input.SessionID
		}
		if filter == nil {
			filter = sessionFilter
		} else {
			filter = combineFilters(filter, sessionFilter)
		}
	}
	if input.ToolName != "" {
		toolFilter := func(m metrics.Metric) bool {
			return m.ToolName == input.ToolName
		}
		if filter == nil {
			filter = toolFilter
		} else {
			filter = combineFilters(filter, toolFilter)
		}
	}

	if input.StartTime != "" || input.EndTime != "" {
		timeFilter := createTimeFilter(input.StartTime, input.EndTime)
		if filter == nil {
			filter = timeFilter
		} else {
			filter = combineFilters(filter, timeFilter)
		}
	}

	aggregator := metrics.NewAggregator(metricsStore)

	switch format {
	case "raw":
		return formatRawMetrics(filter)
	case "table":
		return formatTableMetrics(aggregator, filter, groupBy)
	default:
		return formatSummaryMetrics(aggregator, filter, groupBy)
	}
}

func combineFilters(f1, f2 metrics.Filter) metrics.Filter {
	return func(m metrics.Metric) bool {
		return f1(m) && f2(m)
	}
}

func createTimeFilter(startTime, endTime string) metrics.Filter {
	return func(m metrics.Metric) bool {
		if startTime != "" {
			t, err := time.Parse(time.RFC3339, startTime)
			if err == nil && m.Timestamp.Before(t) {
				return false
			}
		}
		if endTime != "" {
			t, err := time.Parse(time.RFC3339, endTime)
			if err == nil && m.Timestamp.After(t) {
				return false
			}
		}
		return true
	}
}

func formatRawMetrics(filter metrics.Filter) (*mcp.CallToolResult, any, error) {
	var metricsList []metrics.Metric
	if filter == nil {
		metricsList = metricsStore.GetAll()
	} else {
		metricsList = metricsStore.Filter(filter)
	}

	if len(metricsList) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "No metrics found matching filters"}},
		}, nil, nil
	}

	var builder strings.Builder
	builder.WriteString("## Raw Metrics\n\n")
	for _, m := range metricsList {
		builder.WriteString(fmt.Sprintf("- **%s** (Session: %s)\n", m.ToolName, m.SessionID))
		builder.WriteString(fmt.Sprintf("  Tokens: %d in / %d out\n", m.TokensIn, m.TokensOut))
		builder.WriteString(fmt.Sprintf("  Duration: %v\n", m.Duration))
		builder.WriteString(fmt.Sprintf("  Success: %v\n", m.Success))
		if m.ErrorMsg != "" {
			builder.WriteString(fmt.Sprintf("  Error: %s\n", m.ErrorMsg))
		}
		builder.WriteString(fmt.Sprintf("  Timestamp: %s\n\n", m.Timestamp.Format(time.RFC3339)))
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: builder.String()}},
	}, nil, nil
}

func formatSummaryMetrics(aggregator *metrics.Aggregator, filter metrics.Filter, groupBy string) (*mcp.CallToolResult, any, error) {
	if groupBy != "none" {
		return formatGroupedSummary(aggregator, filter, groupBy)
	}

	count := metricsStore.Count()
	if count == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "No metrics available"}},
		}, nil, nil
	}

	avgTokensIn := aggregator.AverageTokensIn(filter)
	avgTokensOut := aggregator.AverageTokensOut(filter)
	avgDuration := aggregator.AverageDuration(filter)
	successRate := aggregator.SuccessRate(filter)

	p50Duration, _ := aggregator.Percentile("duration", 0.50, filter)
	p95Duration, _ := aggregator.Percentile("duration", 0.95, filter)
	p99Duration, _ := aggregator.Percentile("duration", 0.99, filter)

	var builder strings.Builder
	builder.WriteString("## Metrics Summary\n\n")
	builder.WriteString(fmt.Sprintf("**Total Entries:** %d\n\n", count))
	builder.WriteString("### Token Usage\n")
	builder.WriteString(fmt.Sprintf("- Average Input Tokens: %.0f\n", avgTokensIn))
	builder.WriteString(fmt.Sprintf("- Average Output Tokens: %.0f\n\n", avgTokensOut))
	builder.WriteString("### Performance\n")
	builder.WriteString(fmt.Sprintf("- Average Duration: %v\n", avgDuration))
	builder.WriteString(fmt.Sprintf("- P50 Duration: %v\n", p50Duration))
	builder.WriteString(fmt.Sprintf("- P95 Duration: %v\n", p95Duration))
	builder.WriteString(fmt.Sprintf("- P99 Duration: %v\n\n", p99Duration))
	builder.WriteString(fmt.Sprintf("- Success Rate: %.1f%%\n", successRate))

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: builder.String()}},
	}, nil, nil
}

func formatGroupedSummary(aggregator *metrics.Aggregator, filter metrics.Filter, groupBy string) (*mcp.CallToolResult, any, error) {
	var groupByType metrics.GroupBy
	switch groupBy {
	case "hour":
		groupByType = metrics.GroupByHour
	case "day":
		groupByType = metrics.GroupByDay
	case "week":
		groupByType = metrics.GroupByWeek
	default:
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Invalid group_by: %s", groupBy)}},
		}, nil, nil
	}

	groups := aggregator.GroupByTime(groupByType, filter)

	if len(groups) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "No metrics found for grouping"}},
		}, nil, nil
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("## Metrics Summary (grouped by %s)\n\n", groupBy))

	for timeKey, metricsList := range groups {
		builder.WriteString(fmt.Sprintf("### %s\n", timeKey))
		builder.WriteString(fmt.Sprintf("**Count:** %d\n", len(metricsList)))

		count := len(metricsList)
		var totalTokensIn, totalTokensOut int
		var totalDuration time.Duration
		var successCount int

		for _, m := range metricsList {
			totalTokensIn += m.TokensIn
			totalTokensOut += m.TokensOut
			totalDuration += m.Duration
			if m.Success {
				successCount++
			}
		}

		avgDuration := totalDuration / time.Duration(count)
		successRate := float64(successCount) / float64(count) * 100

		builder.WriteString(fmt.Sprintf("- Avg Tokens In: %.0f\n", float64(totalTokensIn)/float64(count)))
		builder.WriteString(fmt.Sprintf("- Avg Tokens Out: %.0f\n", float64(totalTokensOut)/float64(count)))
		builder.WriteString(fmt.Sprintf("- Avg Duration: %v\n", avgDuration))
		builder.WriteString(fmt.Sprintf("- Success Rate: %.1f%%\n\n", successRate))
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: builder.String()}},
	}, nil, nil
}

func formatTableMetrics(aggregator *metrics.Aggregator, filter metrics.Filter, groupBy string) (*mcp.CallToolResult, any, error) {
	var metricsList []metrics.Metric
	if filter == nil {
		metricsList = metricsStore.GetAll()
	} else {
		metricsList = metricsStore.Filter(filter)
	}

	if len(metricsList) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "No metrics found"}},
		}, nil, nil
	}

	var builder strings.Builder
	builder.WriteString("## Metrics Table\n\n")
	builder.WriteString("| Tool | Session | Tokens In | Tokens Out | Duration | Success |\n")
	builder.WriteString("|------|---------|-----------|------------|----------|--------|\n")

	for _, m := range metricsList {
		success := "✓"
		if !m.Success {
			success = "✗"
		}
		builder.WriteString(fmt.Sprintf("| %s | %s | %d | %d | %v | %s |\n",
			m.ToolName, m.SessionID, m.TokensIn, m.TokensOut, m.Duration, success))
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: builder.String()}},
	}, nil, nil
}
