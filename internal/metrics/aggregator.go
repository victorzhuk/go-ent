package metrics

import (
	"fmt"
	"sort"
	"time"
)

type GroupBy int

const (
	GroupByHour GroupBy = iota
	GroupByDay
	GroupByWeek
)

type Filter func(Metric) bool

type Aggregator struct {
	store *Store
}

func NewAggregator(store *Store) *Aggregator {
	return &Aggregator{store: store}
}

func (a *Aggregator) AverageTokensIn(filter Filter) float64 {
	metrics := a.applyFilter(filter)
	if len(metrics) == 0 {
		return 0
	}

	var sum int
	for _, m := range metrics {
		sum += m.TokensIn
	}

	return float64(sum) / float64(len(metrics))
}

func (a *Aggregator) AverageTokensOut(filter Filter) float64 {
	metrics := a.applyFilter(filter)
	if len(metrics) == 0 {
		return 0
	}

	var sum int
	for _, m := range metrics {
		sum += m.TokensOut
	}

	return float64(sum) / float64(len(metrics))
}

func (a *Aggregator) AverageDuration(filter Filter) time.Duration {
	metrics := a.applyFilter(filter)
	if len(metrics) == 0 {
		return 0
	}

	var sum time.Duration
	for _, m := range metrics {
		sum += m.Duration
	}

	return sum / time.Duration(len(metrics))
}

func (a *Aggregator) Percentile(field string, p float64, filter Filter) (float64, error) {
	if p < 0 || p > 1 {
		return 0, fmt.Errorf("invalid percentile %.2f: %w", p, ErrInvalidPercentile)
	}

	metrics := a.applyFilter(filter)
	if len(metrics) == 0 {
		return 0, nil
	}

	values := make([]float64, 0, len(metrics))
	for _, m := range metrics {
		val := extractField(m, field)
		values = append(values, val)
	}

	if len(values) == 0 {
		return 0, nil
	}

	return calculatePercentile(values, p), nil
}

func (a *Aggregator) GroupByTime(groupBy GroupBy, filter Filter) map[string][]Metric {
	metrics := a.applyFilter(filter)
	result := make(map[string][]Metric)

	for _, m := range metrics {
		key := formatTimeKey(m.Timestamp, groupBy)
		result[key] = append(result[key], m)
	}

	return result
}

func (a *Aggregator) FilterByTool(toolName string) Filter {
	return func(m Metric) bool {
		return m.ToolName == toolName
	}
}

func (a *Aggregator) FilterBySession(sessionID string) Filter {
	return func(m Metric) bool {
		return m.SessionID == sessionID
	}
}

func (a *Aggregator) SuccessRate(filter Filter) float64 {
	metrics := a.applyFilter(filter)
	if len(metrics) == 0 {
		return 0
	}

	var successCount int
	for _, m := range metrics {
		if m.Success {
			successCount++
		}
	}

	return float64(successCount) / float64(len(metrics)) * 100
}

func (a *Aggregator) applyFilter(filter Filter) []Metric {
	if filter == nil {
		return a.store.GetAll()
	}
	return a.store.Filter(filter)
}

func calculatePercentile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}

	if len(values) == 1 {
		return values[0]
	}

	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	index := p * float64(len(sorted)-1)
	lower := int(index)

	if lower >= len(sorted)-1 {
		return sorted[len(sorted)-1]
	}

	upper := lower + 1
	weight := index - float64(lower)

	return sorted[lower]*(1-weight) + sorted[upper]*weight
}

func formatTimeKey(t time.Time, groupBy GroupBy) string {
	switch groupBy {
	case GroupByHour:
		return t.Format("2006-01-02T15")
	case GroupByDay:
		return t.Format("2006-01-02")
	case GroupByWeek:
		year, week := t.ISOWeek()
		return fmt.Sprintf("%d-W%02d", year, week)
	default:
		return t.Format("2006-01-02")
	}
}

func extractField(m Metric, field string) float64 {
	switch field {
	case "tokensIn":
		return float64(m.TokensIn)
	case "tokensOut":
		return float64(m.TokensOut)
	case "duration":
		return float64(m.Duration)
	default:
		return 0
	}
}
