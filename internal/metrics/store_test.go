package metrics

//nolint:gosec // test file with necessary file operations

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_NewStore(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		path      string
		retention time.Duration
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid store with default retention",
			path:      filepath.Join(t.TempDir(), "metrics.json"),
			retention: 0,
			wantErr:   false,
		},
		{
			name:      "valid store with custom retention",
			path:      filepath.Join(t.TempDir(), "metrics.json"),
			retention: 24 * time.Hour,
			wantErr:   false,
		},
		{
			name:      "valid store with nested path",
			path:      filepath.Join(t.TempDir(), "subdir", "metrics.json"),
			retention: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s, err := NewStore(tt.path, tt.retention)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, s)
				assert.Equal(t, 0, s.Count())
				require.NoError(t, s.Close())
			}
		})
	}
}

func TestStore_Add(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "metrics.json")
	s, err := NewStore(path, 24*time.Hour)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, s.Close())
	}()

	tests := []struct {
		name    string
		metric  Metric
		wantErr bool
	}{
		{
			name: "valid metric with timestamp",
			metric: Metric{
				SessionID: "session-1",
				ToolName:  "test-tool",
				TokensIn:  100,
				TokensOut: 50,
				Duration:  100 * time.Millisecond,
				Success:   true,
				Timestamp: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "valid metric without timestamp",
			metric: Metric{
				SessionID: "session-2",
				ToolName:  "test-tool",
				TokensIn:  200,
				TokensOut: 100,
				Duration:  200 * time.Millisecond,
				Success:   true,
			},
			wantErr: false,
		},
		{
			name: "metric with error",
			metric: Metric{
				SessionID: "session-3",
				ToolName:  "test-tool",
				TokensIn:  50,
				TokensOut: 0,
				Duration:  50 * time.Millisecond,
				Success:   false,
				ErrorMsg:  "timeout",
				Timestamp: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "metric with metadata",
			metric: Metric{
				SessionID: "session-4",
				ToolName:  "test-tool",
				TokensIn:  150,
				TokensOut: 75,
				Duration:  150 * time.Millisecond,
				Success:   true,
				Metadata: map[string]string{
					"model": "gpt-4",
					"env":   "production",
				},
				Timestamp: time.Now(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initialCount := s.Count()

			err := s.Add(tt.metric)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, initialCount+1, s.Count())
			}
		})
	}
}

func TestStore_Add_RingBufferOverflow(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "metrics.json")
	s, err := NewStore(path, 24*time.Hour)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, s.Close())
	}()

	now := time.Now()

	for i := 0; i < defaultRingBufferSize+10; i++ {
		err := s.Add(Metric{
			SessionID: "session",
			ToolName:  "tool",
			TokensIn:  i,
			Timestamp: now.Add(time.Duration(i) * time.Millisecond),
		})
		require.NoError(t, err)
	}

	assert.Equal(t, defaultRingBufferSize, s.Count())

	all := s.GetAll()
	assert.Equal(t, defaultRingBufferSize, len(all))

	assert.Equal(t, 10, all[0].TokensIn)
	assert.Equal(t, defaultRingBufferSize+9, all[len(all)-1].TokensIn)
}

func TestStore_GetAll(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "metrics.json")
	s, err := NewStore(path, 24*time.Hour)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, s.Close())
	}()

	metrics := []Metric{
		{SessionID: "s1", ToolName: "t1", TokensIn: 100, Timestamp: time.Now()},
		{SessionID: "s2", ToolName: "t2", TokensIn: 200, Timestamp: time.Now()},
		{SessionID: "s3", ToolName: "t3", TokensIn: 300, Timestamp: time.Now()},
	}

	for _, m := range metrics {
		err := s.Add(m)
		require.NoError(t, err)
	}

	all := s.GetAll()
	assert.Equal(t, len(metrics), len(all))

	for i, m := range all {
		assert.Equal(t, metrics[i].SessionID, m.SessionID)
		assert.Equal(t, metrics[i].ToolName, m.ToolName)
		assert.Equal(t, metrics[i].TokensIn, m.TokensIn)
	}
}

func TestStore_Filter(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "metrics.json")
	s, err := NewStore(path, 24*time.Hour)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, s.Close())
	}()

	now := time.Now()
	metrics := []Metric{
		{SessionID: "s1", ToolName: "t1", TokensIn: 100, Success: true, Timestamp: now.Add(-2 * time.Hour)},
		{SessionID: "s1", ToolName: "t2", TokensIn: 200, Success: false, Timestamp: now.Add(-1 * time.Hour)},
		{SessionID: "s2", ToolName: "t1", TokensIn: 300, Success: true, Timestamp: now},
	}

	for _, m := range metrics {
		err := s.Add(m)
		require.NoError(t, err)
	}

	t.Run("filter by success", func(t *testing.T) {
		success := s.Filter(func(m Metric) bool {
			return m.Success
		})
		assert.Equal(t, 2, len(success))
	})

	t.Run("filter by tokens", func(t *testing.T) {
		high := s.Filter(func(m Metric) bool {
			return m.TokensIn >= 200
		})
		assert.Equal(t, 2, len(high))
	})

	t.Run("filter no match", func(t *testing.T) {
		none := s.Filter(func(m Metric) bool {
			return m.TokensIn > 1000
		})
		assert.Equal(t, 0, len(none))
	})
}

func TestStore_FilterByTimeRange(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "metrics.json")
	s, err := NewStore(path, 24*time.Hour)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, s.Close())
	}()

	now := time.Now()
	metrics := []Metric{
		{SessionID: "s1", ToolName: "t1", Timestamp: now.Add(-3 * time.Hour)},
		{SessionID: "s2", ToolName: "t2", Timestamp: now.Add(-2 * time.Hour)},
		{SessionID: "s3", ToolName: "t3", Timestamp: now.Add(-1 * time.Hour)},
		{SessionID: "s4", ToolName: "t4", Timestamp: now},
	}

	for _, m := range metrics {
		err := s.Add(m)
		require.NoError(t, err)
	}

	t.Run("last 2 hours", func(t *testing.T) {
		recent := s.FilterByTimeRange(now.Add(-2*time.Hour), now)
		assert.Equal(t, 3, len(recent))
	})

	t.Run("all time", func(t *testing.T) {
		all := s.FilterByTimeRange(now.Add(-24*time.Hour), now.Add(24*time.Hour))
		assert.Equal(t, len(metrics), len(all))
	})

	t.Run("no overlap", func(t *testing.T) {
		none := s.FilterByTimeRange(now.Add(-24*time.Hour), now.Add(-4*time.Hour))
		assert.Equal(t, 0, len(none))
	})
}

func TestStore_FilterBySession(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "metrics.json")
	s, err := NewStore(path, 24*time.Hour)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, s.Close())
	}()

	now := time.Now()
	metrics := []Metric{
		{SessionID: "s1", ToolName: "t1", Timestamp: now},
		{SessionID: "s1", ToolName: "t2", Timestamp: now},
		{SessionID: "s2", ToolName: "t3", Timestamp: now},
	}

	for _, m := range metrics {
		err := s.Add(m)
		require.NoError(t, err)
	}

	s1 := s.FilterBySession("s1")
	assert.Equal(t, 2, len(s1))

	s2 := s.FilterBySession("s2")
	assert.Equal(t, 1, len(s2))

	none := s.FilterBySession("s3")
	assert.Equal(t, 0, len(none))
}

func TestStore_FilterByTool(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "metrics.json")
	s, err := NewStore(path, 24*time.Hour)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, s.Close())
	}()

	now := time.Now()
	metrics := []Metric{
		{SessionID: "s1", ToolName: "tool-1", Timestamp: now},
		{SessionID: "s2", ToolName: "tool-1", Timestamp: now},
		{SessionID: "s3", ToolName: "tool-2", Timestamp: now},
	}

	for _, m := range metrics {
		err := s.Add(m)
		require.NoError(t, err)
	}

	tool1 := s.FilterByTool("tool-1")
	assert.Equal(t, 2, len(tool1))

	tool2 := s.FilterByTool("tool-2")
	assert.Equal(t, 1, len(tool2))

	none := s.FilterByTool("tool-3")
	assert.Equal(t, 0, len(none))
}

func TestStore_Persistence(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "metrics.json")

	now := time.Now()
	metrics := []Metric{
		{SessionID: "s1", ToolName: "t1", TokensIn: 100, Timestamp: now},
		{SessionID: "s2", ToolName: "t2", TokensIn: 200, Timestamp: now},
	}

	t.Run("save and load", func(t *testing.T) {
		s, err := NewStore(path, 24*time.Hour)
		require.NoError(t, err)

		for _, m := range metrics {
			err := s.Add(m)
			require.NoError(t, err)
		}

		err = s.Close()
		require.NoError(t, err)

		s2, err := NewStore(path, 24*time.Hour)
		require.NoError(t, err)
		defer func() {
			require.NoError(t, s2.Close())
		}()

		all := s2.GetAll()
		assert.Equal(t, len(metrics), len(all))

		for i, m := range all {
			assert.Equal(t, metrics[i].SessionID, m.SessionID)
			assert.Equal(t, metrics[i].ToolName, m.ToolName)
			assert.Equal(t, metrics[i].TokensIn, m.TokensIn)
		}
	})

	t.Run("load non-existent file", func(t *testing.T) {
		newPath := filepath.Join(tmpDir, "new-metrics.json")
		s, err := NewStore(newPath, 24*time.Hour)
		require.NoError(t, err)
		defer func() {
			require.NoError(t, s.Close())
		}()

		assert.Equal(t, 0, s.Count())
	})
}

func TestStore_RetentionPolicy(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "metrics.json")
	retention := 2 * time.Hour

	s, err := NewStore(path, retention)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, s.Close())
	}()

	now := time.Now()

	oldMetric := Metric{
		SessionID: "old",
		ToolName:  "tool",
		Timestamp: now.Add(-3 * time.Hour),
	}

	recentMetric := Metric{
		SessionID: "recent",
		ToolName:  "tool",
		Timestamp: now.Add(-1 * time.Hour),
	}

	err = s.Add(oldMetric)
	require.NoError(t, err)

	err = s.Add(recentMetric)
	require.NoError(t, err)

	assert.Equal(t, 2, s.Count())

	err = s.applyRetention()
	require.NoError(t, err)

	all := s.GetAll()
	assert.Equal(t, 1, len(all))
	assert.Equal(t, "recent", all[0].SessionID)
}

func TestStore_Clear(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "metrics.json")
	s, err := NewStore(path, 24*time.Hour)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, s.Close())
	}()

	metric := Metric{
		SessionID: "s1",
		ToolName:  "t1",
		Timestamp: time.Now(),
	}

	for i := 0; i < 10; i++ {
		err := s.Add(metric)
		require.NoError(t, err)
	}

	assert.Equal(t, 10, s.Count())

	err = s.Clear()
	require.NoError(t, err)

	assert.Equal(t, 0, s.Count())
}

func TestStore_Close(t *testing.T) {
	t.Parallel()

	t.Run("close saves data", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "metrics.json")
		s, err := NewStore(path, 24*time.Hour)
		require.NoError(t, err)

		metric := Metric{
			SessionID: "s1",
			ToolName:  "t1",
			Timestamp: time.Now(),
		}

		err = s.Add(metric)
		require.NoError(t, err)

		err = s.Close()
		require.NoError(t, err)

		data, err := os.ReadFile(path) // #nosec G304 -- test file
		require.NoError(t, err)
		assert.NotEmpty(t, data)
	})

	t.Run("close multiple times", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "metrics.json")
		s, err := NewStore(path, 24*time.Hour)
		require.NoError(t, err)

		err = s.Close()
		require.NoError(t, err)

		err = s.Close()
		require.NoError(t, err)
	})

	t.Run("add after close", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "metrics.json")
		s, err := NewStore(path, 24*time.Hour)
		require.NoError(t, err)

		err = s.Close()
		require.NoError(t, err)

		metric := Metric{
			SessionID: "s1",
			ToolName:  "t1",
			Timestamp: time.Now(),
		}

		err = s.Add(metric)
		assert.Error(t, err)
		assert.Equal(t, ErrStoreClosed, err)
	})
}

func TestStore_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "metrics.json")
	s, err := NewStore(path, 24*time.Hour)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, s.Close())
	}()

	var wg sync.WaitGroup
	numGoroutines := 100
	metricsPerGoroutine := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < metricsPerGoroutine; j++ {
				metric := Metric{
					SessionID: "session",
					ToolName:  "tool",
					TokensIn:  id*100 + j,
					Timestamp: time.Now(),
				}

				err := s.Add(metric)
				assert.NoError(t, err)

				_ = s.Count()
				_ = s.GetAll()
			}
		}(i)
	}

	wg.Wait()

	expectedCount := numGoroutines * metricsPerGoroutine
	if expectedCount > defaultRingBufferSize {
		expectedCount = defaultRingBufferSize
	}

	assert.Equal(t, expectedCount, s.Count())
}

func TestStore_Count(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "metrics.json")
	s, err := NewStore(path, 24*time.Hour)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, s.Close())
	}()

	assert.Equal(t, 0, s.Count())

	metric := Metric{
		SessionID: "s1",
		ToolName:  "t1",
		Timestamp: time.Now(),
	}

	for i := 1; i <= 5; i++ {
		err := s.Add(metric)
		require.NoError(t, err)
		assert.Equal(t, i, s.Count())
	}
}

func TestStore_ConcurrentRetention(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "metrics.json")
	retention := 1 * time.Hour

	s, err := NewStore(path, retention)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, s.Close())
	}()

	now := time.Now()

	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			metric := Metric{
				SessionID: fmt.Sprintf("session-%d", idx),
				ToolName:  "tool",
				Timestamp: now.Add(-time.Duration(idx) * time.Minute),
			}

			err := s.Add(metric)
			require.NoError(t, err)

			err = s.applyRetention()
			assert.NoError(t, err)
		}(i)
	}

	wg.Wait()

	count := s.Count()
	assert.True(t, count >= 0 && count <= 10)
}

func TestStore_FileIOErrors(t *testing.T) {
	t.Parallel()

	t.Run("operations on closed store", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "metrics.json")
		s, err := NewStore(path, 24*time.Hour)
		require.NoError(t, err)

		err = s.Close()
		require.NoError(t, err)

		metric := Metric{
			SessionID: "s1",
			ToolName:  "t1",
			Timestamp: time.Now(),
		}

		err = s.Add(metric)
		assert.Error(t, err)
		assert.Equal(t, ErrStoreClosed, err)

		err = s.Clear()
		assert.Error(t, err)
		assert.Equal(t, ErrStoreClosed, err)
	})
}

func TestStore_BoundaryConditions(t *testing.T) {
	t.Parallel()

	t.Run("empty store", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "metrics.json")
		s, err := NewStore(path, 24*time.Hour)
		require.NoError(t, err)
		defer func() {
			require.NoError(t, s.Close())
		}()

		assert.Equal(t, 0, s.Count())
		assert.Empty(t, s.GetAll())
		assert.Empty(t, s.Filter(func(m Metric) bool {
			return true
		}))
	})

	t.Run("metrics with zero timestamp", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "metrics.json")
		s, err := NewStore(path, 24*time.Hour)
		require.NoError(t, err)
		defer func() {
			require.NoError(t, s.Close())
		}()

		metric := Metric{
			SessionID: "s1",
			ToolName:  "t1",
		}

		err = s.Add(metric)
		require.NoError(t, err)

		all := s.GetAll()
		assert.Equal(t, 1, len(all))
		assert.False(t, all[0].Timestamp.IsZero())
	})

	t.Run("filter with no matches", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "metrics.json")
		s, err := NewStore(path, 24*time.Hour)
		require.NoError(t, err)
		defer func() {
			require.NoError(t, s.Close())
		}()

		metric := Metric{
			SessionID: "s1",
			ToolName:  "t1",
			Timestamp: time.Now(),
		}

		err = s.Add(metric)
		require.NoError(t, err)

		result := s.FilterBySession("non-existent")
		assert.Empty(t, result)

		result = s.FilterByTool("non-existent")
		assert.Empty(t, result)

		now := time.Now()
		result = s.FilterByTimeRange(now.Add(-24*time.Hour), now.Add(-23*time.Hour))
		assert.Empty(t, result)
	})
}
