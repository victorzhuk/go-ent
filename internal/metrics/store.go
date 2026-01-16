package metrics

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	defaultRingBufferSize = 1000
	defaultRetentionDays  = 7
)

// Metric represents a single metric entry.
type Metric struct {
	SessionID string
	ToolName  string
	TokensIn  int
	TokensOut int
	Duration  time.Duration
	Success   bool
	ErrorMsg  string
	Timestamp time.Time
	Metadata  map[string]string
}

// Store manages metrics with in-memory ring buffer and file persistence.
type Store struct {
	mu        sync.RWMutex
	ring      []Metric
	head      int
	count     int
	path      string
	retention time.Duration
	closed    bool
}

// NewStore creates a new metrics store.
func NewStore(path string, retention time.Duration) (*Store, error) {
	if retention <= 0 {
		retention = defaultRetentionDays * 24 * time.Hour
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("create dir: %w", err)
	}

	s := &Store{
		ring:      make([]Metric, defaultRingBufferSize),
		head:      0,
		count:     0,
		path:      path,
		retention: retention,
		closed:    false,
	}

	if err := s.load(); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("load: %w", err)
		}
	}

	if err := s.applyRetention(); err != nil {
		return nil, fmt.Errorf("apply retention: %w", err)
	}

	return s, nil
}

// Add adds a metric to the store.
func (s *Store) Add(m Metric) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrStoreClosed
	}

	if m.Timestamp.IsZero() {
		m.Timestamp = time.Now()
	}

	s.ring[s.head] = m
	s.head = (s.head + 1) % defaultRingBufferSize

	if s.count < defaultRingBufferSize {
		s.count++
	}

	if err := s.saveLocked(); err != nil {
		return fmt.Errorf("save: %w", err)
	}

	return nil
}

// GetAll returns all metrics in the store.
func (s *Store) GetAll() []Metric {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.getAllLocked()
}

// Filter returns metrics that match the given filter function.
func (s *Store) Filter(filter func(Metric) bool) []Metric {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.filterLocked(filter)
}

// FilterByTimeRange returns metrics within the given time range.
func (s *Store) FilterByTimeRange(start, end time.Time) []Metric {
	return s.Filter(func(m Metric) bool {
		return (m.Timestamp.Equal(start) || m.Timestamp.After(start)) &&
			(m.Timestamp.Equal(end) || m.Timestamp.Before(end))
	})
}

// FilterBySession returns metrics for a specific session.
func (s *Store) FilterBySession(sessionID string) []Metric {
	return s.Filter(func(m Metric) bool {
		return m.SessionID == sessionID
	})
}

// FilterByTool returns metrics for a specific tool.
func (s *Store) FilterByTool(toolName string) []Metric {
	return s.Filter(func(m Metric) bool {
		return m.ToolName == toolName
	})
}

// Count returns the number of metrics in the store.
func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.count
}

// Clear removes all metrics from the store.
func (s *Store) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrStoreClosed
	}

	s.head = 0
	s.count = 0

	if err := s.saveLocked(); err != nil {
		return fmt.Errorf("save: %w", err)
	}

	return nil
}

// Close closes the store and ensures all data is persisted.
func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}

	if err := s.saveLocked(); err != nil {
		return fmt.Errorf("save: %w", err)
	}

	s.closed = true
	return nil
}

// applyRetention removes metrics older than the retention period.
func (s *Store) applyRetention() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().Add(-s.retention)
	filtered := s.filterLocked(func(m Metric) bool {
		return m.Timestamp.After(cutoff) || m.Timestamp.Equal(cutoff)
	})

	s.count = 0
	s.head = 0

	for _, m := range filtered {
		s.ring[s.head] = m
		s.head = (s.head + 1) % defaultRingBufferSize
		s.count++
	}

	if err := s.saveLocked(); err != nil {
		return fmt.Errorf("save: %w", err)
	}

	return nil
}

// load loads metrics from the persistent storage.
func (s *Store) load() error {
	data, err := os.ReadFile(s.path) // #nosec G304 -- controlled config/template file path
	if err != nil {
		if os.IsNotExist(err) {
			return err
		}
		return fmt.Errorf("read file: %w", err)
	}

	if len(data) == 0 {
		return nil
	}

	var metrics []Metric
	if err := json.Unmarshal(data, &metrics); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, m := range metrics {
		s.ring[s.head] = m
		s.head = (s.head + 1) % defaultRingBufferSize
		if s.count < defaultRingBufferSize {
			s.count++
		}
	}

	return nil
}

// saveLocked saves metrics to persistent storage (must hold lock).
func (s *Store) saveLocked() error {
	metrics := s.getAllLocked()

	data, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	if err := os.WriteFile(s.path, data, 0600); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

// getAllLocked returns all metrics (must hold read lock).
func (s *Store) getAllLocked() []Metric {
	if s.count == 0 {
		return []Metric{}
	}

	result := make([]Metric, s.count)
	start := (s.head - s.count + defaultRingBufferSize) % defaultRingBufferSize

	if start+s.count <= defaultRingBufferSize {
		copy(result, s.ring[start:start+s.count])
	} else {
		n := defaultRingBufferSize - start
		copy(result[:n], s.ring[start:])
		copy(result[n:], s.ring[:s.count-n])
	}

	return result
}

// filterLocked returns metrics matching the filter (must hold read lock).
func (s *Store) filterLocked(filter func(Metric) bool) []Metric {
	metrics := s.getAllLocked()
	result := make([]Metric, 0, len(metrics))

	for _, m := range metrics {
		if filter(m) {
			result = append(result, m)
		}
	}

	return result
}
