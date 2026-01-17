package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/metrics"
)

var metricsStore *metrics.Store
var metricsEnabled bool = true

// InitMetricsStore initializes the metrics store for tools.
// This is called during MCP server initialization.
func InitMetricsStore(store *metrics.Store) {
	metricsStore = store
}

// SetMetricsEnabled sets whether metrics collection is enabled.
func SetMetricsEnabled(enabled bool) {
	metricsEnabled = enabled
}

// IsMetricsEnabled returns whether metrics collection is enabled.
func IsMetricsEnabled() bool {
	return metricsEnabled
}

type MetricsShowInput struct {
	SessionID string `json:"session_id,omitempty"`
	ToolName  string `json:"tool_name,omitempty"`
	StartTime string `json:"start_time,omitempty"`
	EndTime   string `json:"end_time,omitempty"`
	GroupBy   string `json:"group_by,omitempty"`
	Format    string `json:"format,omitempty"`
	Limit     int    `json:"limit,omitempty"`
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

type MetricsSummaryInput struct {
	SessionID string `json:"session_id,omitempty"`
	ToolName  string `json:"tool_name,omitempty"`
	StartTime string `json:"start_time,omitempty"`
	EndTime   string `json:"end_time,omitempty"`
	GroupBy   string `json:"group_by,omitempty"`
	Format    string `json:"format,omitempty"`
	Limit     int    `json:"limit,omitempty"`
}

type summaryGroup struct {
	Key         string  `json:"key,omitempty"`
	Count       int     `json:"count"`
	TokensIn    float64 `json:"tokens_in"`
	TokensOut   float64 `json:"tokens_out"`
	Duration    float64 `json:"duration_ms"`
	SuccessRate float64 `json:"success_rate"`
}

type summaryResult struct {
	Count     int            `json:"count"`
	TokensIn  float64        `json:"tokens_in"`
	TokensOut float64        `json:"tokens_out"`
	Duration  float64        `json:"duration_ms"`
	Success   float64        `json:"success_rate"`
	Groups    []summaryGroup `json:"groups,omitempty"`
}

func registerMetricsSummary(s *mcp.Server) {
	tool := &mcp.Tool{
		Name:        "metrics_summary",
		Description: "Get aggregated metrics summary with optional grouping",
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
					"description": "Start time (ISO 8601 format)",
				},
				"end_time": map[string]any{
					"type":        "string",
					"description": "End time (ISO 8601 format)",
				},
				"group_by": map[string]any{
					"type":        "string",
					"description": "Group by: none, tool, session, hour, day, week",
					"enum":        []string{"none", "tool", "session", "hour", "day", "week"},
				},
				"format": map[string]any{
					"type":        "string",
					"description": "Output format: table, json",
					"enum":        []string{"table", "json"},
				},
				"limit": map[string]any{
					"type":        "number",
					"description": "Maximum records to process (default: 1000, max: 10000)",
				},
			},
		},
	}

	mcp.AddTool(s, tool, metricsSummaryHandler)
}

func metricsSummaryHandler(_ context.Context, _ *mcp.CallToolRequest, input MetricsSummaryInput) (*mcp.CallToolResult, any, error) {
	if metricsStore == nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Metrics store not initialized"}},
		}, nil, nil
	}

	format := input.Format
	if format == "" {
		format = "table"
	}

	groupBy := input.GroupBy
	if groupBy == "" {
		groupBy = "none"
	}

	limit := input.Limit
	if limit == 0 {
		limit = 1000
	}
	if limit > 10000 {
		limit = 10000
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

	if format == "json" {
		return formatSummaryJSON(aggregator, filter, groupBy, limit)
	}
	return formatSummaryTable(aggregator, filter, groupBy, limit)
}

func formatSummaryJSON(aggregator *metrics.Aggregator, filter metrics.Filter, groupBy string, limit int) (*mcp.CallToolResult, any, error) {
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

	if len(metricsList) > limit {
		metricsList = metricsList[:limit]
	}

	result := summaryResult{
		Count: len(metricsList),
	}

	if groupBy == "none" {
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

		result.TokensIn = float64(totalTokensIn) / float64(len(metricsList))
		result.TokensOut = float64(totalTokensOut) / float64(len(metricsList))
		result.Duration = float64(totalDuration) / float64(len(metricsList)) / float64(time.Millisecond)
		result.Success = float64(successCount) / float64(len(metricsList)) * 100
	} else {
		groups := make(map[string][]metrics.Metric)
		for _, m := range metricsList {
			var key string
			switch groupBy {
			case "tool":
				key = m.ToolName
			case "session":
				key = m.SessionID
			case "hour":
				key = m.Timestamp.Format("2006-01-02T15")
			case "day":
				key = m.Timestamp.Format("2006-01-02")
			case "week":
				year, week := m.Timestamp.ISOWeek()
				key = fmt.Sprintf("%d-W%02d", year, week)
			}
			groups[key] = append(groups[key], m)
		}

		result.Groups = make([]summaryGroup, 0, len(groups))
		for key, groupMetrics := range groups {
			var totalTokensIn, totalTokensOut int
			var totalDuration time.Duration
			var successCount int

			for _, m := range groupMetrics {
				totalTokensIn += m.TokensIn
				totalTokensOut += m.TokensOut
				totalDuration += m.Duration
				if m.Success {
					successCount++
				}
			}

			count := len(groupMetrics)
			summaryGroupData := summaryGroup{
				Key:         key,
				Count:       count,
				TokensIn:    float64(totalTokensIn) / float64(count),
				TokensOut:   float64(totalTokensOut) / float64(count),
				Duration:    float64(totalDuration) / float64(count) / float64(time.Millisecond),
				SuccessRate: float64(successCount) / float64(count) * 100,
			}
			result.Groups = append(result.Groups, summaryGroupData)
		}
	}

	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(resultJSON)}},
	}, nil, nil
}

func formatSummaryTable(aggregator *metrics.Aggregator, filter metrics.Filter, groupBy string, limit int) (*mcp.CallToolResult, any, error) {
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

	if len(metricsList) > limit {
		metricsList = metricsList[:limit]
	}

	var builder strings.Builder
	builder.WriteString("## Metrics Summary\n\n")

	if groupBy == "none" {
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

		builder.WriteString(fmt.Sprintf("**Count:** %d\n\n", count))
		builder.WriteString(fmt.Sprintf("**Avg Tokens In:** %.0f\n", float64(totalTokensIn)/float64(count)))
		builder.WriteString(fmt.Sprintf("**Avg Tokens Out:** %.0f\n", float64(totalTokensOut)/float64(count)))
		builder.WriteString(fmt.Sprintf("**Avg Duration:** %v\n", totalDuration/time.Duration(count)))
		builder.WriteString(fmt.Sprintf("**Success Rate:** %.1f%%\n", float64(successCount)/float64(count)*100))
	} else {
		groups := make(map[string][]metrics.Metric)
		for _, m := range metricsList {
			var key string
			switch groupBy {
			case "tool":
				key = m.ToolName
			case "session":
				key = m.SessionID
			case "hour":
				key = m.Timestamp.Format("2006-01-02T15")
			case "day":
				key = m.Timestamp.Format("2006-01-02")
			case "week":
				year, week := m.Timestamp.ISOWeek()
				key = fmt.Sprintf("%d-W%02d", year, week)
			}
			groups[key] = append(groups[key], m)
		}

		builder.WriteString(fmt.Sprintf("**Grouped by:** %s\n\n", groupBy))
		builder.WriteString("| Key | Count | Avg Tokens In | Avg Tokens Out | Avg Duration | Success Rate |\n")
		builder.WriteString("|-----|-------|---------------|----------------|--------------|--------------|\n")

		for key, groupMetrics := range groups {
			var totalTokensIn, totalTokensOut int
			var totalDuration time.Duration
			var successCount int

			for _, m := range groupMetrics {
				totalTokensIn += m.TokensIn
				totalTokensOut += m.TokensOut
				totalDuration += m.Duration
				if m.Success {
					successCount++
				}
			}

			count := len(groupMetrics)
			builder.WriteString(fmt.Sprintf("| %s | %d | %.0f | %.0f | %v | %.1f%% |\n",
				key, count,
				float64(totalTokensIn)/float64(count),
				float64(totalTokensOut)/float64(count),
				totalDuration/time.Duration(count),
				float64(successCount)/float64(count)*100))
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: builder.String()}},
	}, nil, nil
}

type MetricsExportInput struct {
	Format    string `json:"format"`
	Filename  string `json:"filename,omitempty"`
	SessionID string `json:"session_id,omitempty"`
	ToolName  string `json:"tool_name,omitempty"`
	StartTime string `json:"start_time,omitempty"`
	EndTime   string `json:"end_time,omitempty"`
	Limit     int    `json:"limit,omitempty"`
}

func registerMetricsExport(s *mcp.Server) {
	tool := &mcp.Tool{
		Name:        "metrics_export",
		Description: "Export metrics to file (json, csv, prometheus)",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"format": map[string]any{
					"type":        "string",
					"description": "Export format: json, csv, or prometheus",
					"enum":        []string{"json", "csv", "prometheus"},
				},
				"filename": map[string]any{
					"type":        "string",
					"description": "Output filename (without extension). Auto-generated if empty",
				},
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
					"description": "Start time (ISO 8601)",
				},
				"end_time": map[string]any{
					"type":        "string",
					"description": "End time (ISO 8601)",
				},
				"limit": map[string]any{
					"type":        "number",
					"description": "Maximum records to export (default: 1000, max: 10000)",
				},
			},
			"required": []string{"format"},
		},
	}

	mcp.AddTool(s, tool, metricsExportHandler)
}

func metricsExportHandler(_ context.Context, _ *mcp.CallToolRequest, input MetricsExportInput) (*mcp.CallToolResult, any, error) {
	if metricsStore == nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "Metrics store not initialized"}},
		}, nil, nil
	}

	if input.Format != "json" && input.Format != "csv" && input.Format != "prometheus" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("invalid format: %s (must be json, csv, or prometheus)", input.Format)}},
		}, nil, fmt.Errorf("invalid format: %s", input.Format)
	}

	limit := input.Limit
	if limit == 0 {
		limit = 1000
	}
	if limit > 10000 {
		limit = 10000
	}

	aggregator := metrics.NewAggregator(metricsStore)

	var filter metrics.Filter
	if input.SessionID != "" {
		filter = aggregator.FilterBySession(input.SessionID)
	}
	if input.ToolName != "" {
		toolFilter := aggregator.FilterByTool(input.ToolName)
		if filter == nil {
			filter = toolFilter
		} else {
			baseFilter := filter
			filter = func(m metrics.Metric) bool {
				return baseFilter(m) && toolFilter(m)
			}
		}
	}

	if input.StartTime != "" || input.EndTime != "" {
		startTime, endTime, err := parseTimeRange(input.StartTime, input.EndTime)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("invalid time range: %v", err)}},
			}, nil, nil
		}
		timeFilter := func(m metrics.Metric) bool {
			if startTime != nil && m.Timestamp.Before(*startTime) {
				return false
			}
			if endTime != nil && m.Timestamp.After(*endTime) {
				return false
			}
			return true
		}
		if filter == nil {
			filter = timeFilter
		} else {
			baseFilter := filter
			filter = func(m metrics.Metric) bool {
				return baseFilter(m) && timeFilter(m)
			}
		}
	}

	var metricsList []metrics.Metric
	if filter == nil {
		metricsList = metricsStore.GetAll()
	} else {
		metricsList = metricsStore.Filter(filter)
	}

	if len(metricsList) > limit {
		metricsList = metricsList[:limit]
	}

	if len(metricsList) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "No metrics found matching filters"}},
		}, nil, nil
	}

	var ext string
	switch input.Format {
	case "json":
		ext = "json"
	case "csv":
		ext = "csv"
	case "prometheus":
		ext = "prom"
	}

	var filename string
	if input.Filename != "" {
		filename = input.Filename + "." + ext
	} else {
		filename = fmt.Sprintf("metrics_%s.%s", time.Now().Format("20060102_150405"), ext)
	}

	var data []byte
	var err error
	switch input.Format {
	case "json":
		data, err = json.MarshalIndent(metricsList, "", "  ")
	case "csv":
		var buf strings.Builder
		buf.WriteString("SessionID,ToolName,TokensIn,TokensOut,Duration,Success,ErrorMsg,Timestamp\n")
		for _, m := range metricsList {
			buf.WriteString(fmt.Sprintf("%s,%s,%d,%d,%v,%t,%s,%s\n",
				m.SessionID, m.ToolName, m.TokensIn, m.TokensOut, m.Duration, m.Success, m.ErrorMsg, m.Timestamp.Format(time.RFC3339)))
		}
		data = []byte(buf.String())
	case "prometheus":
		var buf strings.Builder
		buf.WriteString("# HELP tool_tokens_total Total tokens consumed per tool\n")
		buf.WriteString("# TYPE tool_tokens_total gauge\n")
		buf.WriteString("# HELP tool_duration_seconds Tool execution duration in seconds\n")
		buf.WriteString("# TYPE tool_duration_seconds gauge\n")
		buf.WriteString("# HELP tool_success_total Total tool execution success count\n")
		buf.WriteString("# TYPE tool_success_total counter\n")
		for _, m := range metricsList {
			success := "0"
			if m.Success {
				success = "1"
			}
			tokensTotal := m.TokensIn + m.TokensOut
			label := fmt.Sprintf(`{session="%s",tool="%s",success="%s"}`,
				strings.ReplaceAll(m.SessionID, `\`, `\\`),
				strings.ReplaceAll(m.ToolName, `\`, `\\`),
				success)
			buf.WriteString(fmt.Sprintf("tool_tokens_total%s %d\n", label, tokensTotal))
			buf.WriteString(fmt.Sprintf("tool_duration_seconds%s %.3f\n", label, m.Duration.Seconds()))
			buf.WriteString(fmt.Sprintf("tool_success_total%s %s\n", label, success))
		}
		data = []byte(buf.String())
	}

	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("export failed: %v", err)}},
		}, nil, nil
	}

	if filename != "" {
		idx := strings.LastIndex(filename, "/")
		if idx >= 0 {
			dir := filename[:idx]
			if err := os.MkdirAll(dir, 0700); err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("create dir: %v", err)}},
				}, nil, nil
			}
		}
		if err := os.WriteFile(filename, data, 0600); err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("write file: %v", err)}},
			}, nil, nil
		}
	}

	absPath, _ := os.Getwd()
	absPath = filepath.Join(absPath, filename)

	result := map[string]interface{}{
		"success":         true,
		"filename":        filename,
		"format":          input.Format,
		"records":         len(metricsList),
		"file_size_bytes": len(data),
		"path":            absPath,
	}

	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(resultJSON)}},
	}, nil, nil
}

func parseTimeRange(startStr, endStr string) (*time.Time, *time.Time, error) {
	var start, end *time.Time
	var err error

	if startStr != "" {
		t, parseErr := time.Parse(time.RFC3339, startStr)
		if parseErr != nil {
			return nil, nil, fmt.Errorf("parse start time: %w", parseErr)
		}
		start = &t
	}

	if endStr != "" {
		t, parseErr := time.Parse(time.RFC3339, endStr)
		if parseErr != nil {
			return nil, nil, fmt.Errorf("parse end time: %w", parseErr)
		}
		end = &t
	}

	return start, end, err
}
