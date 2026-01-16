package metrics

//nolint:gosec // test file with necessary file operations

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAggregator_AverageTokensIn(t *testing.T) {
	t.Parallel()

	t.Run("calculates average tokens in", func(t *testing.T) {
		t.Parallel()
		store, _ := NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
		agg := NewAggregator(store)

		_ = store.Add(Metric{TokensIn: 100, Timestamp: time.Now()})
		_ = store.Add(Metric{TokensIn: 200, Timestamp: time.Now()})
		_ = store.Add(Metric{TokensIn: 300, Timestamp: time.Now()})

		avg := agg.AverageTokensIn(nil)
		assert.Equal(t, 200.0, avg)
	})

	t.Run("returns zero for empty results", func(t *testing.T) {
		t.Parallel()
		store, _ := NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
		agg := NewAggregator(store)

		avg := agg.AverageTokensIn(nil)
		assert.Equal(t, 0.0, avg)
	})

	t.Run("applies filter", func(t *testing.T) {
		t.Parallel()
		store, _ := NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
		agg := NewAggregator(store)

		_ = store.Add(Metric{TokensIn: 100, ToolName: "tool1", Timestamp: time.Now()})
		_ = store.Add(Metric{TokensIn: 200, ToolName: "tool2", Timestamp: time.Now()})
		_ = store.Add(Metric{TokensIn: 300, ToolName: "tool1", Timestamp: time.Now()})

		avg := agg.AverageTokensIn(agg.FilterByTool("tool1"))
		assert.Equal(t, 200.0, avg)
	})
}

func TestAggregator_AverageTokensOut(t *testing.T) {
	t.Parallel()

	t.Run("calculates average tokens out", func(t *testing.T) {
		t.Parallel()
		store, _ := NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
		agg := NewAggregator(store)

		_ = store.Add(Metric{TokensOut: 50, Timestamp: time.Now()})
		_ = store.Add(Metric{TokensOut: 100, Timestamp: time.Now()})
		_ = store.Add(Metric{TokensOut: 150, Timestamp: time.Now()})

		avg := agg.AverageTokensOut(nil)
		assert.Equal(t, 100.0, avg)
	})

	t.Run("returns zero for empty results", func(t *testing.T) {
		t.Parallel()
		store, _ := NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
		agg := NewAggregator(store)

		avg := agg.AverageTokensOut(nil)
		assert.Equal(t, 0.0, avg)
	})
}

func TestAggregator_AverageDuration(t *testing.T) {
	t.Parallel()

	t.Run("calculates average duration", func(t *testing.T) {
		t.Parallel()
		store, _ := NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
		agg := NewAggregator(store)

		_ = store.Add(Metric{Duration: 100 * time.Millisecond, Timestamp: time.Now()})
		_ = store.Add(Metric{Duration: 200 * time.Millisecond, Timestamp: time.Now()})
		_ = store.Add(Metric{Duration: 300 * time.Millisecond, Timestamp: time.Now()})

		avg := agg.AverageDuration(nil)
		assert.Equal(t, 200*time.Millisecond, avg)
	})

	t.Run("returns zero for empty results", func(t *testing.T) {
		t.Parallel()
		store, _ := NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
		agg := NewAggregator(store)

		avg := agg.AverageDuration(nil)
		assert.Equal(t, time.Duration(0), avg)
	})
}

func TestAggregator_Percentile(t *testing.T) {
	t.Parallel()

	t.Run("calculates p50 percentile for tokensIn", func(t *testing.T) {
		t.Parallel()
		store, _ := NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
		agg := NewAggregator(store)

		_ = store.Add(Metric{TokensIn: 100, Timestamp: time.Now()})
		_ = store.Add(Metric{TokensIn: 200, Timestamp: time.Now()})
		_ = store.Add(Metric{TokensIn: 300, Timestamp: time.Now()})
		_ = store.Add(Metric{TokensIn: 400, Timestamp: time.Now()})
		_ = store.Add(Metric{TokensIn: 500, Timestamp: time.Now()})

		p50, err := agg.Percentile("tokensIn", 0.5, nil)
		assert.NoError(t, err)
		assert.Equal(t, 300.0, p50)
	})

	t.Run("calculates p95 percentile for tokensOut", func(t *testing.T) {
		t.Parallel()
		store, _ := NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
		agg := NewAggregator(store)

		_ = store.Add(Metric{TokensOut: 50, Timestamp: time.Now()})
		_ = store.Add(Metric{TokensOut: 100, Timestamp: time.Now()})
		_ = store.Add(Metric{TokensOut: 150, Timestamp: time.Now()})
		_ = store.Add(Metric{TokensOut: 200, Timestamp: time.Now()})
		_ = store.Add(Metric{TokensOut: 250, Timestamp: time.Now()})

		p95, err := agg.Percentile("tokensOut", 0.95, nil)
		assert.NoError(t, err)
		assert.InDelta(t, 240.0, p95, 0.01)
	})

	t.Run("calculates p99 percentile for duration", func(t *testing.T) {
		t.Parallel()
		store, _ := NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
		agg := NewAggregator(store)

		_ = store.Add(Metric{Duration: 100 * time.Millisecond, Timestamp: time.Now()})
		_ = store.Add(Metric{Duration: 200 * time.Millisecond, Timestamp: time.Now()})
		_ = store.Add(Metric{Duration: 300 * time.Millisecond, Timestamp: time.Now()})
		_ = store.Add(Metric{Duration: 400 * time.Millisecond, Timestamp: time.Now()})
		_ = store.Add(Metric{Duration: 500 * time.Millisecond, Timestamp: time.Now()})

		p99, err := agg.Percentile("duration", 0.99, nil)
		assert.NoError(t, err)
		assert.InDelta(t, float64(496*time.Millisecond), p99, 0.01)
	})

	t.Run("returns zero for empty results", func(t *testing.T) {
		t.Parallel()
		store, _ := NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
		agg := NewAggregator(store)

		p50, err := agg.Percentile("tokensIn", 0.5, nil)
		assert.NoError(t, err)
		assert.Equal(t, 0.0, p50)
	})

	t.Run("handles single element", func(t *testing.T) {
		t.Parallel()
		store, _ := NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
		agg := NewAggregator(store)

		_ = store.Add(Metric{TokensIn: 150, Timestamp: time.Now()})

		p50, err := agg.Percentile("tokensIn", 0.5, nil)
		assert.NoError(t, err)
		assert.Equal(t, 150.0, p50)
	})

	t.Run("returns error for percentile > 1.0", func(t *testing.T) {
		t.Parallel()
		store, _ := NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
		agg := NewAggregator(store)

		_ = store.Add(Metric{TokensIn: 100, Timestamp: time.Now()})

		_, err := agg.Percentile("tokensIn", 1.5, nil)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidPercentile)
	})

	t.Run("returns error for percentile < 0", func(t *testing.T) {
		t.Parallel()
		store, _ := NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
		agg := NewAggregator(store)

		_ = store.Add(Metric{TokensIn: 100, Timestamp: time.Now()})

		_, err := agg.Percentile("tokensIn", -0.5, nil)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidPercentile)
	})

	t.Run("handles zero values in field", func(t *testing.T) {
		t.Parallel()
		store, _ := NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
		agg := NewAggregator(store)

		_ = store.Add(Metric{TokensIn: 100, Timestamp: time.Now()})
		_ = store.Add(Metric{TokensIn: 0, Timestamp: time.Now()})
		_ = store.Add(Metric{TokensIn: 200, Timestamp: time.Now()})

		p50, err := agg.Percentile("tokensIn", 0.5, nil)
		assert.NoError(t, err)
		assert.Equal(t, 100.0, p50)
	})

	t.Run("calculates percentile with two elements", func(t *testing.T) {
		t.Parallel()
		store, _ := NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
		agg := NewAggregator(store)

		_ = store.Add(Metric{TokensIn: 100, Timestamp: time.Now()})
		_ = store.Add(Metric{TokensIn: 200, Timestamp: time.Now()})

		p50, err := agg.Percentile("tokensIn", 0.5, nil)
		assert.NoError(t, err)
		assert.Equal(t, 150.0, p50)

		p25, err := agg.Percentile("tokensIn", 0.25, nil)
		assert.NoError(t, err)
		assert.Equal(t, 125.0, p25)
	})

	t.Run("calculates percentile with all zero values", func(t *testing.T) {
		t.Parallel()
		store, _ := NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
		agg := NewAggregator(store)

		_ = store.Add(Metric{TokensIn: 0, Timestamp: time.Now()})
		_ = store.Add(Metric{TokensIn: 0, Timestamp: time.Now()})
		_ = store.Add(Metric{TokensIn: 0, Timestamp: time.Now()})

		p50, err := agg.Percentile("tokensIn", 0.5, nil)
		assert.NoError(t, err)
		assert.Equal(t, 0.0, p50)
	})

	t.Run("calculates boundary percentile 0.0", func(t *testing.T) {
		t.Parallel()
		store, _ := NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
		agg := NewAggregator(store)

		_ = store.Add(Metric{TokensIn: 100, Timestamp: time.Now()})
		_ = store.Add(Metric{TokensIn: 200, Timestamp: time.Now()})
		_ = store.Add(Metric{TokensIn: 300, Timestamp: time.Now()})

		p0, err := agg.Percentile("tokensIn", 0.0, nil)
		assert.NoError(t, err)
		assert.Equal(t, 100.0, p0)
	})

	t.Run("calculates boundary percentile 1.0", func(t *testing.T) {
		t.Parallel()
		store, _ := NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
		agg := NewAggregator(store)

		_ = store.Add(Metric{TokensIn: 100, Timestamp: time.Now()})
		_ = store.Add(Metric{TokensIn: 200, Timestamp: time.Now()})
		_ = store.Add(Metric{TokensIn: 300, Timestamp: time.Now()})

		p100, err := agg.Percentile("tokensIn", 1.0, nil)
		assert.NoError(t, err)
		assert.Equal(t, 300.0, p100)
	})
}

func TestAggregator_GroupByTime(t *testing.T) {
	t.Parallel()

	t.Run("groups by hour", func(t *testing.T) {
		t.Parallel()
		store, _ := NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
		agg := NewAggregator(store)

		t1 := time.Date(2026, 1, 15, 14, 30, 0, 0, time.UTC)
		t2 := time.Date(2026, 1, 15, 14, 45, 0, 0, time.UTC)
		t3 := time.Date(2026, 1, 15, 15, 0, 0, 0, time.UTC)

		_ = store.Add(Metric{Timestamp: t1})
		_ = store.Add(Metric{Timestamp: t2})
		_ = store.Add(Metric{Timestamp: t3})

		groups := agg.GroupByTime(GroupByHour, nil)
		assert.Len(t, groups, 2)
		assert.Len(t, groups["2026-01-15T14"], 2)
		assert.Len(t, groups["2026-01-15T15"], 1)
	})

	t.Run("groups by day", func(t *testing.T) {
		t.Parallel()
		store, _ := NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
		agg := NewAggregator(store)

		t1 := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
		t2 := time.Date(2026, 1, 15, 14, 0, 0, 0, time.UTC)
		t3 := time.Date(2026, 1, 16, 10, 0, 0, 0, time.UTC)

		_ = store.Add(Metric{Timestamp: t1})
		_ = store.Add(Metric{Timestamp: t2})
		_ = store.Add(Metric{Timestamp: t3})

		groups := agg.GroupByTime(GroupByDay, nil)
		assert.Len(t, groups, 2)
		assert.Len(t, groups["2026-01-15"], 2)
		assert.Len(t, groups["2026-01-16"], 1)
	})

	t.Run("groups by week", func(t *testing.T) {
		t.Parallel()
		store, _ := NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
		agg := NewAggregator(store)

		t1 := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
		t2 := time.Date(2026, 1, 16, 10, 0, 0, 0, time.UTC)
		t3 := time.Date(2026, 1, 22, 10, 0, 0, 0, time.UTC)

		_ = store.Add(Metric{Timestamp: t1})
		_ = store.Add(Metric{Timestamp: t2})
		_ = store.Add(Metric{Timestamp: t3})

		groups := agg.GroupByTime(GroupByWeek, nil)
		assert.Len(t, groups, 2)
		assert.Len(t, groups["2026-W03"], 2)
		assert.Len(t, groups["2026-W04"], 1)
	})

	t.Run("applies filter before grouping", func(t *testing.T) {
		t.Parallel()
		store, _ := NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
		agg := NewAggregator(store)

		t1 := time.Date(2026, 1, 15, 14, 0, 0, 0, time.UTC)
		t2 := time.Date(2026, 1, 15, 14, 0, 0, 0, time.UTC)

		_ = store.Add(Metric{Timestamp: t1, ToolName: "tool1"})
		_ = store.Add(Metric{Timestamp: t2, ToolName: "tool2"})

		groups := agg.GroupByTime(GroupByHour, agg.FilterByTool("tool1"))
		assert.Len(t, groups, 1)
		assert.Len(t, groups["2026-01-15T14"], 1)
	})

	t.Run("groups metrics on hour boundaries", func(t *testing.T) {
		t.Parallel()
		store, _ := NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
		agg := NewAggregator(store)

		t1 := time.Date(2026, 1, 15, 13, 59, 59, 0, time.UTC)
		t2 := time.Date(2026, 1, 15, 14, 0, 0, 0, time.UTC)
		t3 := time.Date(2026, 1, 15, 14, 59, 59, 0, time.UTC)
		t4 := time.Date(2026, 1, 15, 15, 0, 0, 0, time.UTC)

		_ = store.Add(Metric{Timestamp: t1})
		_ = store.Add(Metric{Timestamp: t2})
		_ = store.Add(Metric{Timestamp: t3})
		_ = store.Add(Metric{Timestamp: t4})

		groups := agg.GroupByTime(GroupByHour, nil)
		assert.Len(t, groups, 3)
		assert.Len(t, groups["2026-01-15T13"], 1)
		assert.Len(t, groups["2026-01-15T14"], 2)
		assert.Len(t, groups["2026-01-15T15"], 1)
	})

	t.Run("groups metrics on day boundaries", func(t *testing.T) {
		t.Parallel()
		store, _ := NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
		agg := NewAggregator(store)

		t1 := time.Date(2026, 1, 15, 23, 59, 59, 0, time.UTC)
		t2 := time.Date(2026, 1, 16, 0, 0, 0, 0, time.UTC)
		t3 := time.Date(2026, 1, 16, 0, 0, 1, 0, time.UTC)

		_ = store.Add(Metric{Timestamp: t1})
		_ = store.Add(Metric{Timestamp: t2})
		_ = store.Add(Metric{Timestamp: t3})

		groups := agg.GroupByTime(GroupByDay, nil)
		assert.Len(t, groups, 2)
		assert.Len(t, groups["2026-01-15"], 1)
		assert.Len(t, groups["2026-01-16"], 2)
	})
}

func TestAggregator_FilterByTool(t *testing.T) {
	t.Parallel()

	t.Run("filters by tool name", func(t *testing.T) {
		t.Parallel()
		store, _ := NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
		agg := NewAggregator(store)

		_ = store.Add(Metric{ToolName: "tool1", Timestamp: time.Now()})
		_ = store.Add(Metric{ToolName: "tool2", Timestamp: time.Now()})
		_ = store.Add(Metric{ToolName: "tool1", Timestamp: time.Now()})

		count := len(store.Filter(agg.FilterByTool("tool1")))
		assert.Equal(t, 2, count)
	})

	t.Run("returns empty for non-existent tool", func(t *testing.T) {
		t.Parallel()
		store, _ := NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
		agg := NewAggregator(store)

		_ = store.Add(Metric{ToolName: "tool1", Timestamp: time.Now()})

		count := len(store.Filter(agg.FilterByTool("tool3")))
		assert.Equal(t, 0, count)
	})
}

func TestAggregator_FilterBySession(t *testing.T) {
	t.Parallel()

	t.Run("filters by session ID", func(t *testing.T) {
		t.Parallel()
		store, _ := NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
		agg := NewAggregator(store)

		_ = store.Add(Metric{SessionID: "session1", Timestamp: time.Now()})
		_ = store.Add(Metric{SessionID: "session2", Timestamp: time.Now()})
		_ = store.Add(Metric{SessionID: "session1", Timestamp: time.Now()})

		count := len(store.Filter(agg.FilterBySession("session1")))
		assert.Equal(t, 2, count)
	})

	t.Run("returns empty for non-existent session", func(t *testing.T) {
		t.Parallel()
		store, _ := NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
		agg := NewAggregator(store)

		_ = store.Add(Metric{SessionID: "session1", Timestamp: time.Now()})

		count := len(store.Filter(agg.FilterBySession("session3")))
		assert.Equal(t, 0, count)
	})
}

func TestAggregator_SuccessRate(t *testing.T) {
	t.Parallel()

	t.Run("calculates success rate", func(t *testing.T) {
		t.Parallel()
		store, _ := NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
		agg := NewAggregator(store)

		_ = store.Add(Metric{Success: true, Timestamp: time.Now()})
		_ = store.Add(Metric{Success: true, Timestamp: time.Now()})
		_ = store.Add(Metric{Success: false, Timestamp: time.Now()})

		rate := agg.SuccessRate(nil)
		assert.Equal(t, 66.66666666666666, rate)
	})

	t.Run("returns 0 for empty results", func(t *testing.T) {
		t.Parallel()
		store, _ := NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
		agg := NewAggregator(store)

		rate := agg.SuccessRate(nil)
		assert.Equal(t, 0.0, rate)
	})

	t.Run("calculates 100% success rate", func(t *testing.T) {
		t.Parallel()
		store, _ := NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
		agg := NewAggregator(store)

		_ = store.Add(Metric{Success: true, Timestamp: time.Now()})
		_ = store.Add(Metric{Success: true, Timestamp: time.Now()})

		rate := agg.SuccessRate(nil)
		assert.Equal(t, 100.0, rate)
	})

	t.Run("calculates 0% success rate", func(t *testing.T) {
		t.Parallel()
		store, _ := NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
		agg := NewAggregator(store)

		_ = store.Add(Metric{Success: false, Timestamp: time.Now()})
		_ = store.Add(Metric{Success: false, Timestamp: time.Now()})

		rate := agg.SuccessRate(nil)
		assert.Equal(t, 0.0, rate)
	})

	t.Run("applies filter", func(t *testing.T) {
		t.Parallel()
		store, _ := NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
		agg := NewAggregator(store)

		_ = store.Add(Metric{Success: true, ToolName: "tool1", Timestamp: time.Now()})
		_ = store.Add(Metric{Success: false, ToolName: "tool1", Timestamp: time.Now()})
		_ = store.Add(Metric{Success: true, ToolName: "tool2", Timestamp: time.Now()})

		rate := agg.SuccessRate(agg.FilterByTool("tool1"))
		assert.Equal(t, 50.0, rate)
	})

	t.Run("calculates success rate with single metric", func(t *testing.T) {
		t.Parallel()
		store, _ := NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
		agg := NewAggregator(store)

		_ = store.Add(Metric{Success: true, Timestamp: time.Now()})

		rate := agg.SuccessRate(nil)
		assert.Equal(t, 100.0, rate)
	})

	t.Run("calculates 0% success rate with single metric", func(t *testing.T) {
		t.Parallel()
		store, _ := NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
		agg := NewAggregator(store)

		_ = store.Add(Metric{Success: false, Timestamp: time.Now()})

		rate := agg.SuccessRate(nil)
		assert.Equal(t, 0.0, rate)
	})
}

func TestAggregator_Integration(t *testing.T) {
	t.Parallel()

	t.Run("comprehensive statistics calculation", func(t *testing.T) {
		t.Parallel()
		store, _ := NewStore(t.TempDir()+"/metrics.json", 24*time.Hour)
		agg := NewAggregator(store)

		_ = store.Add(Metric{
			SessionID: "session1",
			ToolName:  "tool1",
			TokensIn:  100,
			TokensOut: 50,
			Duration:  100 * time.Millisecond,
			Success:   true,
			Timestamp: time.Date(2026, 1, 15, 14, 30, 0, 0, time.UTC),
		})
		_ = store.Add(Metric{
			SessionID: "session1",
			ToolName:  "tool1",
			TokensIn:  200,
			TokensOut: 100,
			Duration:  200 * time.Millisecond,
			Success:   true,
			Timestamp: time.Date(2026, 1, 15, 14, 45, 0, 0, time.UTC),
		})
		_ = store.Add(Metric{
			SessionID: "session2",
			ToolName:  "tool2",
			TokensIn:  300,
			TokensOut: 150,
			Duration:  300 * time.Millisecond,
			Success:   false,
			Timestamp: time.Date(2026, 1, 15, 15, 0, 0, 0, time.UTC),
		})

		avgTokensIn := agg.AverageTokensIn(nil)
		assert.Equal(t, 200.0, avgTokensIn)

		avgTokensOut := agg.AverageTokensOut(nil)
		assert.Equal(t, 100.0, avgTokensOut)

		avgDuration := agg.AverageDuration(nil)
		assert.Equal(t, 200*time.Millisecond, avgDuration)

		p50, _ := agg.Percentile("tokensIn", 0.5, nil)
		assert.Equal(t, 200.0, p50)

		successRate := agg.SuccessRate(nil)
		assert.Equal(t, 66.66666666666666, successRate)

		tool1SuccessRate := agg.SuccessRate(agg.FilterByTool("tool1"))
		assert.Equal(t, 100.0, tool1SuccessRate)

		hourlyGroups := agg.GroupByTime(GroupByHour, nil)
		assert.Len(t, hourlyGroups["2026-01-15T14"], 2)
		assert.Len(t, hourlyGroups["2026-01-15T15"], 1)

		session1Metrics := len(store.Filter(agg.FilterBySession("session1")))
		assert.Equal(t, 2, session1Metrics)
	})
}

func TestCalculatePercentile(t *testing.T) {
	t.Parallel()

	t.Run("calculates correctly", func(t *testing.T) {
		t.Parallel()
		values := []float64{10, 20, 30, 40, 50}
		result := calculatePercentile(values, 0.5)
		assert.Equal(t, 30.0, result)
	})

	t.Run("handles empty array", func(t *testing.T) {
		t.Parallel()
		values := []float64{}
		result := calculatePercentile(values, 0.5)
		assert.Equal(t, 0.0, result)
	})

	t.Run("handles single element", func(t *testing.T) {
		t.Parallel()
		values := []float64{42}
		result := calculatePercentile(values, 0.5)
		assert.Equal(t, 42.0, result)
	})
}

func TestFormatTimeKey(t *testing.T) {
	t.Parallel()

	t.Run("formats hour key", func(t *testing.T) {
		t.Parallel()
		tm := time.Date(2026, 1, 15, 14, 30, 0, 0, time.UTC)
		key := formatTimeKey(tm, GroupByHour)
		assert.Equal(t, "2026-01-15T14", key)
	})

	t.Run("formats day key", func(t *testing.T) {
		t.Parallel()
		tm := time.Date(2026, 1, 15, 14, 30, 0, 0, time.UTC)
		key := formatTimeKey(tm, GroupByDay)
		assert.Equal(t, "2026-01-15", key)
	})

	t.Run("formats week key", func(t *testing.T) {
		t.Parallel()
		tm := time.Date(2026, 1, 15, 14, 30, 0, 0, time.UTC)
		key := formatTimeKey(tm, GroupByWeek)
		assert.Equal(t, "2026-W03", key)
	})
}

func TestExtractField(t *testing.T) {
	t.Parallel()

	t.Run("extracts tokensIn", func(t *testing.T) {
		t.Parallel()
		m := Metric{TokensIn: 150}
		val := extractField(m, "tokensIn")
		assert.Equal(t, 150.0, val)
	})

	t.Run("extracts tokensOut", func(t *testing.T) {
		t.Parallel()
		m := Metric{TokensOut: 250}
		val := extractField(m, "tokensOut")
		assert.Equal(t, 250.0, val)
	})

	t.Run("extracts duration", func(t *testing.T) {
		t.Parallel()
		m := Metric{Duration: 100 * time.Millisecond}
		val := extractField(m, "duration")
		assert.Equal(t, float64(100*time.Millisecond), val)
	})

	t.Run("returns zero for unknown field", func(t *testing.T) {
		t.Parallel()
		m := Metric{TokensIn: 100}
		val := extractField(m, "unknown")
		assert.Equal(t, 0.0, val)
	})
}
