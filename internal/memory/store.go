package memory

import (
	"fmt"
	"sync"
	"time"
)

type Pattern struct {
	ID           string
	TaskType     string
	Provider     string
	Model        string
	Method       string
	FileCount    int
	ContextSize  int
	Success      bool
	Cost         float64
	Duration     time.Duration
	OutputSize   int
	Timestamp    time.Time
	ErrorPattern string
}

type PatternStats struct {
	TotalExecutions   int
	SuccessCount      int
	FailureCount      int
	SuccessRate       float64
	AverageCost       float64
	AverageDuration   time.Duration
	AverageOutputSize int
	FirstSeen         time.Time
	LastSeen          time.Time
}

type ProviderPerformance struct {
	Provider        string
	Model           string
	Method          string
	TaskType        string
	SuccessRate     float64
	AverageCost     float64
	AverageDuration time.Duration
	TotalExecutions int
	RecommendedFor  []string
}

type RoutingRecommendation struct {
	Provider      string
	Model         string
	Method        string
	Reason        string
	Confidence    float64
	EstimatedCost float64
}

type MemoryStore struct {
	patterns      map[string]*Pattern
	taskPatterns  map[string][]*Pattern
	providerStats map[string]*PatternStats
	mu            sync.RWMutex
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		patterns:      make(map[string]*Pattern),
		taskPatterns:  make(map[string][]*Pattern),
		providerStats: make(map[string]*PatternStats),
	}
}

func (m *MemoryStore) Store(pattern *Pattern) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if pattern.ID == "" {
		pattern.ID = fmt.Sprintf("pattern_%d_%d", time.Now().Unix(), len(m.patterns))
	}

	pattern.Timestamp = time.Now()

	m.patterns[pattern.ID] = pattern

	key := fmt.Sprintf("%s:%s:%s:%s", pattern.TaskType, pattern.Provider, pattern.Model, pattern.Method)
	m.taskPatterns[key] = append(m.taskPatterns[key], pattern)

	providerKey := fmt.Sprintf("%s:%s:%s", pattern.Provider, pattern.Model, pattern.Method)
	stats, exists := m.providerStats[providerKey]
	if !exists {
		stats = &PatternStats{
			FirstSeen:   pattern.Timestamp,
			LastSeen:    pattern.Timestamp,
			AverageCost: pattern.Cost,
		}
		m.providerStats[providerKey] = stats
	}

	stats.TotalExecutions++
	if pattern.Success {
		stats.SuccessCount++
	} else {
		stats.FailureCount++
	}
	stats.SuccessRate = float64(stats.SuccessCount) / float64(stats.TotalExecutions) * 100

	stats.AverageCost = (stats.AverageCost*float64(stats.TotalExecutions-1) + pattern.Cost) / float64(stats.TotalExecutions)

	stats.AverageDuration = (stats.AverageDuration*time.Duration(stats.TotalExecutions-1) + pattern.Duration) / time.Duration(stats.TotalExecutions)

	stats.AverageOutputSize = (stats.AverageOutputSize*(stats.TotalExecutions-1) + pattern.OutputSize) / stats.TotalExecutions

	if pattern.Timestamp.After(stats.LastSeen) {
		stats.LastSeen = pattern.Timestamp
	}

	return nil
}

func (m *MemoryStore) Query(taskType string, fileCount int, contextSize int) *RoutingRecommendation {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var bestProvider *RoutingRecommendation
	var bestScore float64

	for key, patterns := range m.taskPatterns {
		parts := splitKey(key)
		if len(parts) < 4 {
			continue
		}

		statsKey := fmt.Sprintf("%s:%s:%s", parts[1], parts[2], parts[3])
		stats := m.providerStats[statsKey]
		if stats == nil {
			continue
		}

		recommendation := m.evaluatePattern(patterns, stats, taskType, fileCount, contextSize)
		if recommendation != nil && recommendation.Confidence > bestScore {
			bestProvider = recommendation
			bestScore = recommendation.Confidence
		}
	}

	return bestProvider
}

func (m *MemoryStore) evaluatePattern(patterns []*Pattern, stats *PatternStats, taskType string, fileCount int, contextSize int) *RoutingRecommendation {
	if len(patterns) < 3 {
		return nil
	}

	if stats.SuccessRate < 50 {
		return nil
	}

	var taskTypeMatch bool
	var fileCountMatch bool
	var contextSizeMatch bool

	for _, p := range patterns {
		if p.TaskType == taskType {
			taskTypeMatch = true
		}
		if fileCount > 0 && p.FileCount == fileCount {
			fileCountMatch = true
		}
		if contextSize > 0 && p.ContextSize == contextSize {
			contextSizeMatch = true
		}
	}

	if !taskTypeMatch && len(patterns) > 0 {
		taskType = patterns[0].TaskType
	}

	confidence := stats.SuccessRate / 100.0

	if fileCountMatch {
		confidence *= 1.2
	}
	if contextSizeMatch {
		confidence *= 1.1
	}

	if confidence > 1.0 {
		confidence = 1.0
	}

	if confidence < 0.6 {
		return nil
	}

	recommendedFor := []string{}
	if taskType != "" {
		recommendedFor = append(recommendedFor, taskType)
	}

	if len(patterns) > 0 {
		return &RoutingRecommendation{
			Provider:      patterns[0].Provider,
			Model:         patterns[0].Model,
			Method:        patterns[0].Method,
			Reason:        fmt.Sprintf("learned from %d executions (%.1f%% success, avg cost $%.4f)", len(patterns), stats.SuccessRate, stats.AverageCost),
			Confidence:    confidence,
			EstimatedCost: stats.AverageCost,
		}
	}

	return nil
}

func (m *MemoryStore) GetProviderStats(provider, model, method string) (*PatternStats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := fmt.Sprintf("%s:%s:%s", provider, model, method)
	stats, exists := m.providerStats[key]
	if !exists {
		return nil, fmt.Errorf("provider stats not found: %s", key)
	}

	return stats, nil
}

func (m *MemoryStore) GetBestProviderForTask(taskType string, maxCost float64) (*RoutingRecommendation, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var best *RoutingRecommendation
	var bestScore float64

	for key, patterns := range m.taskPatterns {
		stats := m.providerStats[key]
		if stats == nil {
			continue
		}

		for _, p := range patterns {
			if p.TaskType == taskType {
				recommendation := m.evaluatePattern(patterns, stats, taskType, 0, 0)
				if recommendation != nil {
					if maxCost > 0 && recommendation.EstimatedCost > maxCost {
						continue
					}
					if recommendation.Confidence > bestScore {
						best = recommendation
						bestScore = recommendation.Confidence
					}
				}
			}
		}
	}

	if best == nil {
		return nil, fmt.Errorf("no pattern data for task type: %s", taskType)
	}

	return best, nil
}

func (m *MemoryStore) GetAllProviders() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	providers := make(map[string]bool)
	for key := range m.providerStats {
		parts := splitKey(key)
		if len(parts) > 0 {
			providers[parts[0]] = true
		}
	}

	result := make([]string, 0, len(providers))
	for p := range providers {
		result = append(result, p)
	}
	return result
}

func (m *MemoryStore) GetTotalPatterns() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.patterns)
}

func splitKey(key string) []string {
	parts := []string{}
	current := ""
	for _, c := range key {
		if c == ':' {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}
