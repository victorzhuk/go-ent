package aggregator

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/victorzhuk/go-ent/internal/memory"
)

func TestNewAggregator(t *testing.T) {
	t.Run("creates aggregator with timeout", func(t *testing.T) {
		timeout := 30 * time.Second
		agg := NewAggregatorWithoutTracking(timeout, nil)

		require.NotNil(t, agg)
		assert.Equal(t, timeout, agg.timeout)
		assert.NotNil(t, agg.results)
		assert.NotNil(t, agg.completed)
		assert.NotNil(t, agg.failed)
		assert.False(t, agg.startTime.IsZero())
	})

	t.Run("creates aggregator with zero timeout", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(0, nil)

		require.NotNil(t, agg)
		assert.Equal(t, time.Duration(0), agg.timeout)
	})
}

func TestAddResult(t *testing.T) {
	t.Run("adds successful result", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)
		result := &WorkerResult{
			WorkerID: "worker-1",
			Provider: "glm",
			Model:    "glm-4",
			Status:   "completed",
			Output:   "success",
		}

		err := agg.AddResult("worker-1", result)

		require.NoError(t, err)
		assert.Len(t, agg.completed, 1)
		assert.Len(t, agg.failed, 0)
		assert.Equal(t, "worker-1", agg.completed[0])
	})

	t.Run("adds failed result", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)
		result := &WorkerResult{
			WorkerID: "worker-2",
			Provider: "glm",
			Model:    "glm-4",
			Status:   "failed",
			Error:    "something went wrong",
		}

		err := agg.AddResult("worker-2", result)

		require.NoError(t, err)
		assert.Len(t, agg.completed, 0)
		assert.Len(t, agg.failed, 1)
		assert.Equal(t, "worker-2", agg.failed[0])
	})

	t.Run("adds result with error status", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)
		result := &WorkerResult{
			WorkerID: "worker-3",
			Provider: "kimi",
			Model:    "kimi-k2",
			Status:   "running",
			Error:    "timeout",
		}

		err := agg.AddResult("worker-3", result)

		require.NoError(t, err)
		assert.Len(t, agg.completed, 0)
		assert.Len(t, agg.failed, 1)
	})

	t.Run("rejects nil result", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)

		err := agg.AddResult("worker-1", nil)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be nil")
	})

	t.Run("rejects result with mismatched worker ID", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)
		result := &WorkerResult{
			WorkerID: "worker-2",
			Status:   "completed",
		}

		err := agg.AddResult("worker-1", result)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "worker ID mismatch")
	})

	t.Run("updates existing result", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)

		_ = agg.AddResult("worker-1", &WorkerResult{
			WorkerID: "worker-1",
			Status:   "running",
			Output:   "initial",
		})

		err := agg.AddResult("worker-1", &WorkerResult{
			WorkerID: "worker-1",
			Status:   "completed",
			Output:   "final",
		})

		require.NoError(t, err)

		result, _ := agg.GetResult("worker-1")
		assert.Equal(t, "completed", result.Status)
		assert.Equal(t, "final", result.Output)
	})

	t.Run("sets end time", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)
		before := time.Now()
		result := &WorkerResult{
			WorkerID: "worker-1",
			Status:   "completed",
		}

		err := agg.AddResult("worker-1", result)
		after := time.Now()

		require.NoError(t, err)
		assert.False(t, result.EndTime.IsZero())
		assert.True(t, result.EndTime.After(before) || result.EndTime.Equal(before))
		assert.True(t, result.EndTime.Before(after) || result.EndTime.Equal(after))
	})
}

func TestGetResult(t *testing.T) {
	t.Run("gets existing result", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)
		result := &WorkerResult{
			WorkerID: "worker-1",
			Status:   "completed",
			Output:   "test output",
		}
		_ = agg.AddResult("worker-1", result)

		retrieved, err := agg.GetResult("worker-1")

		require.NoError(t, err)
		assert.Equal(t, "worker-1", retrieved.WorkerID)
		assert.Equal(t, "completed", retrieved.Status)
		assert.Equal(t, "test output", retrieved.Output)
	})

	t.Run("returns error for non-existent result", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)

		_, err := agg.GetResult("worker-999")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestAllCompleted(t *testing.T) {
	t.Run("returns false when no results", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)

		assert.False(t, agg.AllCompleted())
	})

	t.Run("returns false when some results pending", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)
		agg.RegisterWorkers([]string{"worker-1", "worker-2", "worker-3"})
		_ = agg.AddResult("worker-1", &WorkerResult{WorkerID: "worker-1", Status: "completed"})

		assert.False(t, agg.AllCompleted())
	})

	t.Run("returns true when all completed", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)
		agg.RegisterWorkers([]string{"worker-1", "worker-2"})
		_ = agg.AddResult("worker-1", &WorkerResult{WorkerID: "worker-1", Status: "completed"})
		_ = agg.AddResult("worker-2", &WorkerResult{WorkerID: "worker-2", Status: "completed"})

		assert.True(t, agg.AllCompleted())
	})

	t.Run("returns true when all completed or failed", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)
		agg.RegisterWorkers([]string{"worker-1", "worker-2", "worker-3"})
		_ = agg.AddResult("worker-1", &WorkerResult{WorkerID: "worker-1", Status: "completed"})
		_ = agg.AddResult("worker-2", &WorkerResult{WorkerID: "worker-2", Status: "failed", Error: "error"})
		_ = agg.AddResult("worker-3", &WorkerResult{WorkerID: "worker-3", Status: "completed"})

		assert.True(t, agg.AllCompleted())
	})
}

func TestFailedWorkers(t *testing.T) {
	t.Run("returns empty when no failures", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)
		_ = agg.AddResult("worker-1", &WorkerResult{WorkerID: "worker-1", Status: "completed"})

		failed := agg.FailedWorkers()

		assert.Empty(t, failed)
	})

	t.Run("returns list of failed workers", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)
		_ = agg.AddResult("worker-1", &WorkerResult{WorkerID: "worker-1", Status: "completed"})
		_ = agg.AddResult("worker-2", &WorkerResult{WorkerID: "worker-2", Status: "failed", Error: "error"})
		_ = agg.AddResult("worker-3", &WorkerResult{WorkerID: "worker-3", Status: "completed"})

		failed := agg.FailedWorkers()

		assert.Len(t, failed, 1)
		assert.Equal(t, "worker-2", failed[0])
	})
}

func TestCompletedWorkers(t *testing.T) {
	t.Run("returns empty when no completions", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)
		_ = agg.AddResult("worker-1", &WorkerResult{WorkerID: "worker-1", Status: "failed", Error: "error"})

		completed := agg.CompletedWorkers()

		assert.Empty(t, completed)
	})

	t.Run("returns list of completed workers", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)
		_ = agg.AddResult("worker-1", &WorkerResult{WorkerID: "worker-1", Status: "completed"})
		_ = agg.AddResult("worker-2", &WorkerResult{WorkerID: "worker-2", Status: "failed", Error: "error"})
		_ = agg.AddResult("worker-3", &WorkerResult{WorkerID: "worker-3", Status: "completed"})

		completed := agg.CompletedWorkers()

		assert.Len(t, completed, 2)
		assert.Contains(t, completed, "worker-1")
		assert.Contains(t, completed, "worker-3")
	})
}

func TestGetAggregatedResult(t *testing.T) {
	t.Run("calculates stats correctly", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)
		agg.RegisterWorkers([]string{"worker-1", "worker-2", "worker-3"})
		_ = agg.AddResult("worker-1", &WorkerResult{WorkerID: "worker-1", Status: "completed"})
		_ = agg.AddResult("worker-2", &WorkerResult{WorkerID: "worker-2", Status: "failed", Error: "error"})
		_ = agg.AddResult("worker-3", &WorkerResult{WorkerID: "worker-3", Status: "completed"})

		aggregated, err := agg.GetAggregatedResult()

		require.NoError(t, err)
		assert.Equal(t, 2, aggregated.CompletedCount)
		assert.Equal(t, 1, aggregated.FailedCount)
		assert.Equal(t, 66.66666666666666, aggregated.SuccessRate)
		assert.Len(t, aggregated.Results, 3)
		assert.False(t, aggregated.Duration <= 0)
	})

	t.Run("handles empty results", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)

		aggregated, err := agg.GetAggregatedResult()

		require.NoError(t, err)
		assert.Equal(t, 0, aggregated.CompletedCount)
		assert.Equal(t, 0, aggregated.FailedCount)
		assert.Equal(t, 0.0, aggregated.SuccessRate)
		assert.Len(t, aggregated.Results, 0)
	})

	t.Run("returns copy of results map", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)
		_ = agg.AddResult("worker-1", &WorkerResult{WorkerID: "worker-1", Status: "completed"})

		aggregated, _ := agg.GetAggregatedResult()

		delete(aggregated.Results, "worker-1")

		_, err := agg.GetResult("worker-1")
		assert.NoError(t, err)
	})
}

func TestWaitForAll(t *testing.T) {
	t.Run("waits for all workers to complete", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)
		agg.RegisterWorkers([]string{"worker-1", "worker-2"})

		go func() {
			time.Sleep(100 * time.Millisecond)
			_ = agg.AddResult("worker-1", &WorkerResult{WorkerID: "worker-1", Status: "completed"})
			time.Sleep(50 * time.Millisecond)
			_ = agg.AddResult("worker-2", &WorkerResult{WorkerID: "worker-2", Status: "completed"})
		}()

		result, err := agg.WaitForAll(2 * time.Second)

		require.NoError(t, err)
		assert.Equal(t, 2, result.CompletedCount)
		assert.Equal(t, 0, result.FailedCount)
	})

	t.Run("times out for slow workers", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)
		agg.RegisterWorkers([]string{"worker-1", "worker-2"})
		_ = agg.AddResult("worker-1", &WorkerResult{WorkerID: "worker-1", Status: "completed"})

		result, err := agg.WaitForAll(300 * time.Millisecond)

		require.NoError(t, err)
		assert.Equal(t, 1, result.CompletedCount)
		assert.Equal(t, 1, result.FailedCount)
		assert.Contains(t, agg.FailedWorkers(), "worker-2")
	})

	t.Run("uses aggregator timeout when timeout is zero", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(500*time.Millisecond, nil)
		agg.RegisterWorkers([]string{"worker-1"})

		result, err := agg.WaitForAll(0)

		require.NoError(t, err)
		assert.Equal(t, 0, result.CompletedCount)
		assert.Equal(t, 1, result.FailedCount)
	})

	t.Run("uses default 5 minute timeout when both are zero", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(0, nil)
		agg.RegisterWorkers([]string{"worker-1"})

		result, err := agg.WaitForAll(500 * time.Millisecond)

		require.NoError(t, err)
		assert.Equal(t, 0, result.CompletedCount)
		assert.Equal(t, 1, result.FailedCount)
	})
}

func TestRegisterWorkers(t *testing.T) {
	t.Run("registers multiple workers", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)
		workerIDs := []string{"worker-1", "worker-2", "worker-3"}

		agg.RegisterWorkers(workerIDs)

		assert.Equal(t, 3, agg.TotalWorkers())

		for _, id := range workerIDs {
			result, err := agg.GetResult(id)
			require.NoError(t, err)
			assert.Equal(t, "running", result.Status)
			assert.False(t, result.StartTime.IsZero())
		}
	})

	t.Run("does not duplicate existing workers", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)
		_ = agg.AddResult("worker-1", &WorkerResult{WorkerID: "worker-1", Status: "completed"})

		agg.RegisterWorkers([]string{"worker-1", "worker-2"})

		assert.Equal(t, 2, agg.TotalWorkers())
	})
}

func TestMarkFailed(t *testing.T) {
	t.Run("marks existing worker as failed", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)
		agg.RegisterWorkers([]string{"worker-1"})

		agg.MarkFailed("worker-1", "test error")

		failed := agg.FailedWorkers()
		assert.Len(t, failed, 1)
		assert.Equal(t, "worker-1", failed[0])

		result, _ := agg.GetResult("worker-1")
		assert.Equal(t, "failed", result.Status)
		assert.Equal(t, "test error", result.Error)
	})

	t.Run("creates worker entry if not exists", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)

		agg.MarkFailed("worker-new", "worker failed")

		assert.Equal(t, 1, agg.TotalWorkers())
		failed := agg.FailedWorkers()
		assert.Len(t, failed, 1)
	})

	t.Run("sets end time", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)
		agg.RegisterWorkers([]string{"worker-1"})

		before := time.Now()
		agg.MarkFailed("worker-1", "error")
		after := time.Now()

		result, _ := agg.GetResult("worker-1")
		assert.False(t, result.EndTime.IsZero())
		assert.True(t, result.EndTime.After(before) || result.EndTime.Equal(before))
		assert.True(t, result.EndTime.Before(after) || result.EndTime.Equal(after))
	})
}

func TestTotalWorkers(t *testing.T) {
	t.Run("returns zero for new aggregator", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)

		assert.Equal(t, 0, agg.TotalWorkers())
	})

	t.Run("counts registered workers", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)
		agg.RegisterWorkers([]string{"worker-1", "worker-2", "worker-3"})

		assert.Equal(t, 3, agg.TotalWorkers())
	})
}

func TestConcurrentAccess(t *testing.T) {
	t.Run("handles concurrent adds", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)
		numWorkers := 10

		workers := make([]string, numWorkers)
		for i := 0; i < numWorkers; i++ {
			workers[i] = fmt.Sprintf("worker-%d", i)
		}
		agg.RegisterWorkers(workers)

		done := make(chan bool, numWorkers)

		for i := 0; i < numWorkers; i++ {
			go func(id int) {
				workerID := fmt.Sprintf("worker-%d", id)
				result := &WorkerResult{
					WorkerID: workerID,
					Status:   "completed",
					Output:   fmt.Sprintf("output %d", id),
				}
				_ = agg.AddResult(workerID, result)
				done <- true
			}(i)
		}

		for i := 0; i < numWorkers; i++ {
			<-done
		}

		assert.Equal(t, numWorkers, agg.TotalWorkers())
		assert.Equal(t, numWorkers, len(agg.CompletedWorkers()))
	})

	t.Run("handles concurrent reads and writes", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)
		agg.RegisterWorkers([]string{"worker-1", "worker-2", "worker-3"})

		done := make(chan bool, 6)

		for i := 0; i < 3; i++ {
			go func(id int) {
				workerID := fmt.Sprintf("worker-%d", id+1)
				_ = agg.AddResult(workerID, &WorkerResult{WorkerID: workerID, Status: "completed"})
				done <- true
			}(i)

			go func(id int) {
				workerID := fmt.Sprintf("worker-%d", id+1)
				_, _ = agg.GetResult(workerID)
				done <- true
			}(i)
		}

		for i := 0; i < 6; i++ {
			<-done
		}

		assert.True(t, agg.AllCompleted())
	})
}

func TestTrackFileEdit(t *testing.T) {
	t.Run("tracks single file edit", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)

		edit := &FileEdit{
			WorkerID:  "worker-1",
			FilePath:  "/path/to/file.go",
			StartTime: time.Now().Add(-1 * time.Hour),
			EndTime:   time.Now(),
			Operation: "write",
		}

		agg.TrackFileEdit(edit)

		conflicts := agg.GetConflicts()
		assert.Len(t, conflicts, 0)
	})

	t.Run("detects conflict for overlapping edits", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)

		edit1 := &FileEdit{
			WorkerID:  "worker-1",
			FilePath:  "/path/to/file.go",
			StartTime: time.Now().Add(-1 * time.Hour),
			EndTime:   time.Now(),
			Operation: "write",
		}

		edit2 := &FileEdit{
			WorkerID:  "worker-2",
			FilePath:  "/path/to/file.go",
			StartTime: time.Now().Add(-30 * time.Minute),
			EndTime:   time.Now().Add(30 * time.Minute),
			Operation: "write",
		}

		agg.TrackFileEdit(edit1)
		agg.TrackFileEdit(edit2)

		conflicts := agg.GetConflicts()
		assert.Len(t, conflicts, 1)
		assert.Equal(t, "/path/to/file.go", conflicts[0].FilePath)
		assert.Contains(t, conflicts[0].Workers, "worker-1")
		assert.Contains(t, conflicts[0].Workers, "worker-2")
	})

	t.Run("no conflict for same worker", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)

		edit1 := &FileEdit{
			WorkerID:  "worker-1",
			FilePath:  "/path/to/file.go",
			StartTime: time.Now().Add(-1 * time.Hour),
			EndTime:   time.Now(),
			Operation: "write",
		}

		edit2 := &FileEdit{
			WorkerID:  "worker-1",
			FilePath:  "/path/to/file.go",
			StartTime: time.Now().Add(-30 * time.Minute),
			EndTime:   time.Now().Add(30 * time.Minute),
			Operation: "write",
		}

		agg.TrackFileEdit(edit1)
		agg.TrackFileEdit(edit2)

		conflicts := agg.GetConflicts()
		assert.Len(t, conflicts, 0)
	})

	t.Run("no conflict for different files", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)

		edit1 := &FileEdit{
			WorkerID:  "worker-1",
			FilePath:  "/path/to/file1.go",
			StartTime: time.Now().Add(-1 * time.Hour),
			EndTime:   time.Now(),
			Operation: "write",
		}

		edit2 := &FileEdit{
			WorkerID:  "worker-2",
			FilePath:  "/path/to/file2.go",
			StartTime: time.Now().Add(-30 * time.Minute),
			EndTime:   time.Now().Add(30 * time.Minute),
			Operation: "write",
		}

		agg.TrackFileEdit(edit1)
		agg.TrackFileEdit(edit2)

		conflicts := agg.GetConflicts()
		assert.Len(t, conflicts, 0)
	})

	t.Run("no conflict for non-overlapping time windows", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)

		edit1 := &FileEdit{
			WorkerID:  "worker-1",
			FilePath:  "/path/to/file.go",
			StartTime: time.Now().Add(-2 * time.Hour),
			EndTime:   time.Now().Add(-1 * time.Hour),
			Operation: "write",
		}

		edit2 := &FileEdit{
			WorkerID:  "worker-2",
			FilePath:  "/path/to/file.go",
			StartTime: time.Now(),
			EndTime:   time.Now().Add(1 * time.Hour),
			Operation: "write",
		}

		agg.TrackFileEdit(edit1)
		agg.TrackFileEdit(edit2)

		conflicts := agg.GetConflicts()
		assert.Len(t, conflicts, 0)
	})
}

func TestDetectConflict(t *testing.T) {
	t.Run("detects overlapping time windows", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)

		now := time.Now()

		edit1 := &FileEdit{
			WorkerID:  "worker-1",
			FilePath:  "/test/file.go",
			StartTime: now.Add(-10 * time.Minute),
			EndTime:   now.Add(-5 * time.Minute),
			Operation: "write",
		}

		edit2 := &FileEdit{
			WorkerID:  "worker-2",
			FilePath:  "/test/file.go",
			StartTime: now.Add(-8 * time.Minute),
			EndTime:   now.Add(-3 * time.Minute),
			Operation: "write",
		}

		agg.TrackFileEdit(edit1)
		agg.TrackFileEdit(edit2)

		conflicts := agg.GetConflicts()
		assert.Len(t, conflicts, 1)
	})

	t.Run("detects edge case - touching time windows", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)

		now := time.Now()

		edit1 := &FileEdit{
			WorkerID:  "worker-1",
			FilePath:  "/test/file.go",
			StartTime: now.Add(-10 * time.Minute),
			EndTime:   now,
			Operation: "write",
		}

		edit2 := &FileEdit{
			WorkerID:  "worker-2",
			FilePath:  "/test/file.go",
			StartTime: now,
			EndTime:   now.Add(10 * time.Minute),
			Operation: "write",
		}

		agg.TrackFileEdit(edit1)
		agg.TrackFileEdit(edit2)

		conflicts := agg.GetConflicts()
		assert.Len(t, conflicts, 1)
	})

	t.Run("handles zero end time", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)

		now := time.Now()

		edit1 := &FileEdit{
			WorkerID:  "worker-1",
			FilePath:  "/test/file.go",
			StartTime: now.Add(-10 * time.Minute),
			EndTime:   time.Time{},
			Operation: "write",
		}

		edit2 := &FileEdit{
			WorkerID:  "worker-2",
			FilePath:  "/test/file.go",
			StartTime: now.Add(-5 * time.Minute),
			EndTime:   now,
			Operation: "write",
		}

		agg.TrackFileEdit(edit1)
		agg.TrackFileEdit(edit2)

		conflicts := agg.GetConflicts()
		assert.Len(t, conflicts, 1)
	})
}

func TestSetResolutionStrategy(t *testing.T) {
	t.Run("sets first_write strategy", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)

		agg.SetResolutionStrategy("first_write")

		assert.Equal(t, "first_write", agg.resolution)
	})

	t.Run("sets last_write strategy", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)

		agg.SetResolutionStrategy("last_write")

		assert.Equal(t, "last_write", agg.resolution)
	})

	t.Run("sets merge_attempt strategy", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)

		agg.SetResolutionStrategy("merge_attempt")

		assert.Equal(t, "merge_attempt", agg.resolution)
	})

	t.Run("ignores invalid strategy", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)
		original := agg.resolution

		agg.SetResolutionStrategy("invalid_strategy")

		assert.Equal(t, original, agg.resolution)
	})
}

func TestGetConflicts(t *testing.T) {
	t.Run("returns copy of conflicts", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)

		edit := &FileEdit{
			WorkerID:  "worker-1",
			FilePath:  "/test/file.go",
			StartTime: time.Now().Add(-1 * time.Hour),
			EndTime:   time.Now(),
			Operation: "write",
		}

		agg.TrackFileEdit(edit)

		conflictEdit := &FileEdit{
			WorkerID:  "worker-2",
			FilePath:  "/test/file.go",
			StartTime: time.Now().Add(-30 * time.Minute),
			EndTime:   time.Now().Add(30 * time.Minute),
			Operation: "write",
		}

		agg.TrackFileEdit(conflictEdit)

		conflicts := agg.GetConflicts()

		assert.Len(t, conflicts, 1)

		conflicts[0].FilePath = "/modified/path"

		newConflicts := agg.GetConflicts()
		assert.Equal(t, "/test/file.go", newConflicts[0].FilePath)
	})
}

func TestAddResultWithFileEdits(t *testing.T) {
	t.Run("tracks file edits from worker result", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)

		now := time.Now()

		result := &WorkerResult{
			WorkerID: "worker-1",
			Status:   "completed",
			FileEdits: []FileEdit{
				{
					WorkerID:  "worker-1",
					FilePath:  "/test/file1.go",
					StartTime: now.Add(-10 * time.Minute),
					EndTime:   now,
					Operation: "write",
				},
				{
					WorkerID:  "worker-1",
					FilePath:  "/test/file2.go",
					StartTime: now.Add(-5 * time.Minute),
					EndTime:   now,
					Operation: "write",
				},
			},
		}

		err := agg.AddResult("worker-1", result)

		require.NoError(t, err)
		assert.Len(t, agg.GetConflicts(), 0)
	})

	t.Run("detects conflicts from worker results", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)

		now := time.Now()

		result1 := &WorkerResult{
			WorkerID: "worker-1",
			Status:   "completed",
			FileEdits: []FileEdit{
				{
					WorkerID:  "worker-1",
					FilePath:  "/test/file.go",
					StartTime: now.Add(-10 * time.Minute),
					EndTime:   now,
					Operation: "write",
				},
			},
		}

		result2 := &WorkerResult{
			WorkerID: "worker-2",
			Status:   "completed",
			FileEdits: []FileEdit{
				{
					WorkerID:  "worker-2",
					FilePath:  "/test/file.go",
					StartTime: now.Add(-5 * time.Minute),
					EndTime:   now.Add(5 * time.Minute),
					Operation: "write",
				},
			},
		}

		_ = agg.AddResult("worker-1", result1)
		_ = agg.AddResult("worker-2", result2)

		conflicts := agg.GetConflicts()
		assert.Len(t, conflicts, 1)
	})
}

func TestGetAggregatedResultWithConflicts(t *testing.T) {
	t.Run("includes conflicts in aggregated result", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)

		now := time.Now()

		result1 := &WorkerResult{
			WorkerID: "worker-1",
			Status:   "completed",
			FileEdits: []FileEdit{
				{
					WorkerID:  "worker-1",
					FilePath:  "/test/file.go",
					StartTime: now.Add(-10 * time.Minute),
					EndTime:   now,
					Operation: "write",
				},
			},
		}

		result2 := &WorkerResult{
			WorkerID: "worker-2",
			Status:   "completed",
			FileEdits: []FileEdit{
				{
					WorkerID:  "worker-2",
					FilePath:  "/test/file.go",
					StartTime: now.Add(-5 * time.Minute),
					EndTime:   now.Add(5 * time.Minute),
					Operation: "write",
				},
			},
		}

		_ = agg.AddResult("worker-1", result1)
		_ = agg.AddResult("worker-2", result2)

		aggregated, err := agg.GetAggregatedResult()

		require.NoError(t, err)
		assert.Len(t, aggregated.Conflicts, 1)
		assert.Equal(t, 1, aggregated.ConflictCount)
		assert.Equal(t, "/test/file.go", aggregated.Conflicts[0].FilePath)
	})

	t.Run("returns zero conflicts when none exist", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)

		aggregated, err := agg.GetAggregatedResult()

		require.NoError(t, err)
		assert.Empty(t, aggregated.Conflicts)
		assert.Equal(t, 0, aggregated.ConflictCount)
	})
}

func TestResolveConflicts(t *testing.T) {
	t.Run("applies last_write resolution", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)
		agg.SetResolutionStrategy("last_write")

		now := time.Now()

		result1 := &WorkerResult{
			WorkerID: "worker-1",
			Status:   "completed",
			FileEdits: []FileEdit{
				{
					WorkerID:  "worker-1",
					FilePath:  "/test/file.go",
					StartTime: now.Add(-10 * time.Minute),
					EndTime:   now.Add(-5 * time.Minute),
					Operation: "write",
				},
			},
		}

		result2 := &WorkerResult{
			WorkerID: "worker-2",
			Status:   "completed",
			FileEdits: []FileEdit{
				{
					WorkerID:  "worker-2",
					FilePath:  "/test/file.go",
					StartTime: now.Add(-5 * time.Minute),
					EndTime:   now,
					Operation: "write",
				},
			},
		}

		_ = agg.AddResult("worker-1", result1)
		_ = agg.AddResult("worker-2", result2)

		agg.ResolveConflicts()

		r1, _ := agg.GetResult("worker-1")
		r2, _ := agg.GetResult("worker-2")

		assert.True(t, r1.HasConflicts)
		assert.Equal(t, 1, r1.ConflictCount)
		assert.False(t, r2.HasConflicts)
		assert.Equal(t, 0, r2.ConflictCount)
	})

	t.Run("applies first_write resolution", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)
		agg.SetResolutionStrategy("first_write")

		now := time.Now()

		result1 := &WorkerResult{
			WorkerID: "worker-1",
			Status:   "completed",
			FileEdits: []FileEdit{
				{
					WorkerID:  "worker-1",
					FilePath:  "/test/file.go",
					StartTime: now.Add(-10 * time.Minute),
					EndTime:   now.Add(-5 * time.Minute),
					Operation: "write",
				},
			},
		}

		result2 := &WorkerResult{
			WorkerID: "worker-2",
			Status:   "completed",
			FileEdits: []FileEdit{
				{
					WorkerID:  "worker-2",
					FilePath:  "/test/file.go",
					StartTime: now.Add(-5 * time.Minute),
					EndTime:   now,
					Operation: "write",
				},
			},
		}

		_ = agg.AddResult("worker-1", result1)
		_ = agg.AddResult("worker-2", result2)

		agg.ResolveConflicts()

		r1, _ := agg.GetResult("worker-1")
		r2, _ := agg.GetResult("worker-2")

		assert.False(t, r1.HasConflicts)
		assert.Equal(t, 0, r1.ConflictCount)
		assert.True(t, r2.HasConflicts)
		assert.Equal(t, 1, r2.ConflictCount)
	})

	t.Run("applies merge_attempt resolution", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)
		agg.SetResolutionStrategy("merge_attempt")

		now := time.Now()

		result1 := &WorkerResult{
			WorkerID: "worker-1",
			Status:   "completed",
			FileEdits: []FileEdit{
				{
					WorkerID:  "worker-1",
					FilePath:  "/test/file.go",
					StartTime: now.Add(-10 * time.Minute),
					EndTime:   now.Add(-5 * time.Minute),
					Operation: "write",
				},
			},
		}

		result2 := &WorkerResult{
			WorkerID: "worker-2",
			Status:   "completed",
			FileEdits: []FileEdit{
				{
					WorkerID:  "worker-2",
					FilePath:  "/test/file.go",
					StartTime: now.Add(-5 * time.Minute),
					EndTime:   now,
					Operation: "write",
				},
			},
		}

		_ = agg.AddResult("worker-1", result1)
		_ = agg.AddResult("worker-2", result2)

		agg.ResolveConflicts()

		r1, _ := agg.GetResult("worker-1")
		r2, _ := agg.GetResult("worker-2")

		assert.True(t, r1.HasConflicts)
		assert.True(t, r2.HasConflicts)
	})
}

func TestMergeFirstSuccess(t *testing.T) {
	t.Run("merges first successful result", func(t *testing.T) {
		config := &MergeConfig{
			Strategy:   MergeFirstSuccess,
			OnConflict: "skip",
		}
		agg := NewAggregatorWithoutTracking(10*time.Second, config)

		_ = agg.AddResult("worker-1", &WorkerResult{
			WorkerID: "worker-1",
			Provider: "glm",
			Model:    "glm-4",
			Status:   "completed",
			Output:   "first output",
		})

		_ = agg.AddResult("worker-2", &WorkerResult{
			WorkerID: "worker-2",
			Provider: "kimi",
			Model:    "kimi-k2",
			Status:   "completed",
			Output:   "second output",
		})

		merged, err := agg.Merge()

		require.NoError(t, err)
		assert.Equal(t, "first output", merged.Content)
		assert.Len(t, merged.SourceWorkers, 1)
		assert.Equal(t, "worker-1", merged.SourceWorkers[0])
		assert.Equal(t, "glm", merged.Metadata["provider"])
		assert.Len(t, merged.Decisions, 1)
		assert.Equal(t, MergeFirstSuccess, merged.Decisions[0].Strategy)
	})

	t.Run("skips failed workers", func(t *testing.T) {
		config := &MergeConfig{
			Strategy:   MergeFirstSuccess,
			OnConflict: "skip",
		}
		agg := NewAggregatorWithoutTracking(10*time.Second, config)

		_ = agg.AddResult("worker-1", &WorkerResult{
			WorkerID: "worker-1",
			Status:   "failed",
			Error:    "error",
		})

		_ = agg.AddResult("worker-2", &WorkerResult{
			WorkerID: "worker-2",
			Provider: "kimi",
			Model:    "kimi-k2",
			Status:   "completed",
			Output:   "success output",
		})

		merged, err := agg.Merge()

		require.NoError(t, err)
		assert.Equal(t, "success output", merged.Content)
		assert.Equal(t, "worker-2", merged.SourceWorkers[0])
	})

	t.Run("returns error when no successful results", func(t *testing.T) {
		config := &MergeConfig{
			Strategy:   MergeFirstSuccess,
			OnConflict: "skip",
		}
		agg := NewAggregatorWithoutTracking(10*time.Second, config)

		_ = agg.AddResult("worker-1", &WorkerResult{
			WorkerID: "worker-1",
			Status:   "failed",
			Error:    "error",
		})

		_, err := agg.Merge()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no completed workers")
	})
}

func TestMergeLastSuccess(t *testing.T) {
	t.Run("merges last successful result", func(t *testing.T) {
		config := &MergeConfig{
			Strategy:   MergeLastSuccess,
			OnConflict: "skip",
		}
		agg := NewAggregatorWithoutTracking(10*time.Second, config)

		_ = agg.AddResult("worker-1", &WorkerResult{
			WorkerID: "worker-1",
			Provider: "glm",
			Model:    "glm-4",
			Status:   "completed",
			Output:   "first output",
		})

		_ = agg.AddResult("worker-2", &WorkerResult{
			WorkerID: "worker-2",
			Provider: "kimi",
			Model:    "kimi-k2",
			Status:   "completed",
			Output:   "last output",
		})

		merged, err := agg.Merge()

		require.NoError(t, err)
		assert.Equal(t, "last output", merged.Content)
		assert.Len(t, merged.SourceWorkers, 1)
		assert.Equal(t, "worker-2", merged.SourceWorkers[0])
		assert.Equal(t, MergeLastSuccess, merged.Decisions[0].Strategy)
	})

	t.Run("merges multiple workers correctly", func(t *testing.T) {
		config := &MergeConfig{
			Strategy:   MergeLastSuccess,
			OnConflict: "skip",
		}
		agg := NewAggregatorWithoutTracking(10*time.Second, config)

		_ = agg.AddResult("worker-1", &WorkerResult{
			WorkerID: "worker-1",
			Status:   "completed",
			Output:   "first",
		})

		_ = agg.AddResult("worker-2", &WorkerResult{
			WorkerID: "worker-2",
			Status:   "completed",
			Output:   "second",
		})

		_ = agg.AddResult("worker-3", &WorkerResult{
			WorkerID: "worker-3",
			Status:   "completed",
			Output:   "third",
		})

		merged, err := agg.Merge()

		require.NoError(t, err)
		assert.Equal(t, "third", merged.Content)
	})
}

func TestMergeConcat(t *testing.T) {
	t.Run("concatenates multiple worker outputs", func(t *testing.T) {
		config := &MergeConfig{
			Strategy:   MergeConcat,
			OnConflict: "skip",
		}
		agg := NewAggregatorWithoutTracking(10*time.Second, config)

		_ = agg.AddResult("worker-1", &WorkerResult{
			WorkerID: "worker-1",
			Status:   "completed",
			Output:   "first output",
		})

		_ = agg.AddResult("worker-2", &WorkerResult{
			WorkerID: "worker-2",
			Status:   "completed",
			Output:   "second output",
		})

		_ = agg.AddResult("worker-3", &WorkerResult{
			WorkerID: "worker-3",
			Status:   "completed",
			Output:   "third output",
		})

		merged, err := agg.Merge()

		require.NoError(t, err)
		assert.Contains(t, merged.Content, "first output")
		assert.Contains(t, merged.Content, "second output")
		assert.Contains(t, merged.Content, "third output")
		assert.Len(t, merged.SourceWorkers, 3)
		assert.Equal(t, 3, merged.Metadata["worker_count"])
		assert.Equal(t, MergeConcat, merged.Decisions[0].Strategy)
	})

	t.Run("handles empty output", func(t *testing.T) {
		config := &MergeConfig{
			Strategy:   MergeConcat,
			OnConflict: "skip",
		}
		agg := NewAggregatorWithoutTracking(10*time.Second, config)

		_ = agg.AddResult("worker-1", &WorkerResult{
			WorkerID: "worker-1",
			Status:   "completed",
			Output:   "",
		})

		merged, err := agg.Merge()

		require.NoError(t, err)
		assert.Len(t, merged.SourceWorkers, 1)
	})
}

func TestMergeJSON(t *testing.T) {
	t.Run("merges valid JSON outputs", func(t *testing.T) {
		config := &MergeConfig{
			Strategy:   MergeJSON,
			OnConflict: "skip",
		}
		agg := NewAggregatorWithoutTracking(10*time.Second, config)

		_ = agg.AddResult("worker-1", &WorkerResult{
			WorkerID: "worker-1",
			Status:   "completed",
			Output:   `{"key1": "value1", "key2": "value2"}`,
		})

		_ = agg.AddResult("worker-2", &WorkerResult{
			WorkerID: "worker-2",
			Status:   "completed",
			Output:   `{"key3": "value3", "key4": "value4"}`,
		})

		merged, err := agg.Merge()

		require.NoError(t, err)
		assert.Contains(t, merged.Content, "key1")
		assert.Contains(t, merged.Content, "value1")
		assert.Contains(t, merged.Content, "key3")
		assert.Contains(t, merged.Content, "value3")
		assert.Equal(t, 4, merged.Metadata["key_count"])
		assert.Equal(t, MergeJSON, merged.Decisions[0].Strategy)
	})

	t.Run("handles duplicate keys with conflict markers", func(t *testing.T) {
		config := &MergeConfig{
			Strategy:   MergeJSON,
			OnConflict: "markers",
		}
		agg := NewAggregatorWithoutTracking(10*time.Second, config)

		_ = agg.AddResult("worker-1", &WorkerResult{
			WorkerID: "worker-1",
			Status:   "completed",
			Output:   `{"key1": "value1"}`,
		})

		_ = agg.AddResult("worker-2", &WorkerResult{
			WorkerID: "worker-2",
			Status:   "completed",
			Output:   `{"key1": "value2"}`,
		})

		merged, err := agg.Merge()

		require.NoError(t, err)
		assert.Contains(t, merged.Content, "key1")
		assert.Contains(t, merged.Content, "key1_worker-2")
	})

	t.Run("handles invalid JSON gracefully", func(t *testing.T) {
		config := &MergeConfig{
			Strategy:   MergeJSON,
			OnConflict: "skip",
		}
		agg := NewAggregatorWithoutTracking(10*time.Second, config)

		_ = agg.AddResult("worker-1", &WorkerResult{
			WorkerID: "worker-1",
			Status:   "completed",
			Output:   "not json",
		})

		_ = agg.AddResult("worker-2", &WorkerResult{
			WorkerID: "worker-2",
			Status:   "completed",
			Output:   `{"key": "value"}`,
		})

		merged, err := agg.Merge()

		require.NoError(t, err)
		assert.Contains(t, merged.Content, "key")
		assert.Contains(t, merged.Content, "value")
	})

	t.Run("returns error when no valid JSON", func(t *testing.T) {
		config := &MergeConfig{
			Strategy:   MergeJSON,
			OnConflict: "skip",
		}
		agg := NewAggregatorWithoutTracking(10*time.Second, config)

		_ = agg.AddResult("worker-1", &WorkerResult{
			WorkerID: "worker-1",
			Status:   "completed",
			Output:   "not json",
		})

		_ = agg.AddResult("worker-2", &WorkerResult{
			WorkerID: "worker-2",
			Status:   "completed",
			Output:   "also not json",
		})

		merged, err := agg.Merge()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no valid JSON")
		assert.Len(t, merged.Errors, 1)
	})
}

func TestMergeByPriority(t *testing.T) {
	t.Run("selects worker by provider priority", func(t *testing.T) {
		config := &MergeConfig{
			Strategy:   MergeByPriority,
			Priority:   []string{"glm", "kimi", "deepseek"},
			OnConflict: "skip",
		}
		agg := NewAggregatorWithoutTracking(10*time.Second, config)

		_ = agg.AddResult("worker-1", &WorkerResult{
			WorkerID: "worker-1",
			Provider: "kimi",
			Model:    "kimi-k2",
			Status:   "completed",
			Output:   "kimi output",
		})

		_ = agg.AddResult("worker-2", &WorkerResult{
			WorkerID: "worker-2",
			Provider: "glm",
			Model:    "glm-4",
			Status:   "completed",
			Output:   "glm output",
		})

		_ = agg.AddResult("worker-3", &WorkerResult{
			WorkerID: "worker-3",
			Provider: "deepseek",
			Model:    "deepseek-coder",
			Status:   "completed",
			Output:   "deepseek output",
		})

		merged, err := agg.Merge()

		require.NoError(t, err)
		assert.Equal(t, "glm output", merged.Content)
		assert.Equal(t, "worker-2", merged.SourceWorkers[0])
		assert.Equal(t, 0, merged.Metadata["priority"])
		assert.Equal(t, MergeByPriority, merged.Decisions[0].Strategy)
	})

	t.Run("skips providers not in priority list", func(t *testing.T) {
		config := &MergeConfig{
			Strategy:   MergeByPriority,
			Priority:   []string{"glm"},
			OnConflict: "skip",
		}
		agg := NewAggregatorWithoutTracking(10*time.Second, config)

		_ = agg.AddResult("worker-1", &WorkerResult{
			WorkerID: "worker-1",
			Provider: "kimi",
			Status:   "completed",
			Output:   "kimi output",
		})

		_ = agg.AddResult("worker-2", &WorkerResult{
			WorkerID: "worker-2",
			Provider: "glm",
			Status:   "completed",
			Output:   "glm output",
		})

		merged, err := agg.Merge()

		require.NoError(t, err)
		assert.Equal(t, "glm output", merged.Content)
	})

	t.Run("returns error for empty priority list", func(t *testing.T) {
		config := &MergeConfig{
			Strategy:   MergeByPriority,
			Priority:   []string{},
			OnConflict: "skip",
		}
		agg := NewAggregatorWithoutTracking(10*time.Second, config)

		_ = agg.AddResult("worker-1", &WorkerResult{
			WorkerID: "worker-1",
			Status:   "completed",
			Output:   "output",
		})

		_, err := agg.Merge()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "priority list is empty")
	})

	t.Run("returns error when no provider matches", func(t *testing.T) {
		config := &MergeConfig{
			Strategy:   MergeByPriority,
			Priority:   []string{"anthropic"},
			OnConflict: "skip",
		}
		agg := NewAggregatorWithoutTracking(10*time.Second, config)

		_ = agg.AddResult("worker-1", &WorkerResult{
			WorkerID: "worker-1",
			Provider: "glm",
			Status:   "completed",
			Output:   "output",
		})

		_, err := agg.Merge()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no result from providers in priority list")
	})
}

func TestMergePreferredProvider(t *testing.T) {
	t.Run("selects preferred provider", func(t *testing.T) {
		config := &MergeConfig{
			Strategy:   MergePreferredProvider,
			Preferred:  "kimi",
			OnConflict: "skip",
		}
		agg := NewAggregatorWithoutTracking(10*time.Second, config)

		_ = agg.AddResult("worker-1", &WorkerResult{
			WorkerID: "worker-1",
			Provider: "glm",
			Status:   "completed",
			Output:   "glm output",
		})

		_ = agg.AddResult("worker-2", &WorkerResult{
			WorkerID: "worker-2",
			Provider: "kimi",
			Model:    "kimi-k2",
			Status:   "completed",
			Output:   "kimi output",
		})

		_ = agg.AddResult("worker-3", &WorkerResult{
			WorkerID: "worker-3",
			Provider: "deepseek",
			Status:   "completed",
			Output:   "deepseek output",
		})

		merged, err := agg.Merge()

		require.NoError(t, err)
		assert.Equal(t, "kimi output", merged.Content)
		assert.Equal(t, "worker-2", merged.SourceWorkers[0])
		assert.Equal(t, "kimi", merged.Metadata["provider"])
		assert.Equal(t, MergePreferredProvider, merged.Decisions[0].Strategy)
	})

	t.Run("returns error when preferred not specified", func(t *testing.T) {
		config := &MergeConfig{
			Strategy:   MergePreferredProvider,
			Preferred:  "",
			OnConflict: "skip",
		}
		agg := NewAggregatorWithoutTracking(10*time.Second, config)

		_ = agg.AddResult("worker-1", &WorkerResult{
			WorkerID: "worker-1",
			Status:   "completed",
			Output:   "output",
		})

		_, err := agg.Merge()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "preferred provider not specified")
	})

	t.Run("returns error when preferred provider not found", func(t *testing.T) {
		config := &MergeConfig{
			Strategy:   MergePreferredProvider,
			Preferred:  "anthropic",
			OnConflict: "skip",
		}
		agg := NewAggregatorWithoutTracking(10*time.Second, config)

		_ = agg.AddResult("worker-1", &WorkerResult{
			WorkerID: "worker-1",
			Provider: "glm",
			Status:   "completed",
			Output:   "output",
		})

		_, err := agg.Merge()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no result from preferred provider")
	})
}

func TestNewAggregatorWithMergeConfig(t *testing.T) {
	t.Run("creates aggregator with merge config", func(t *testing.T) {
		config := &MergeConfig{
			Strategy:   MergeByPriority,
			Priority:   []string{"glm", "kimi"},
			Preferred:  "glm",
			OnConflict: "markers",
		}
		agg := NewAggregatorWithoutTracking(10*time.Second, config)

		require.NotNil(t, agg)
		assert.Equal(t, MergeByPriority, agg.mergeConfig.Strategy)
		assert.Equal(t, []string{"glm", "kimi"}, agg.mergeConfig.Priority)
		assert.Equal(t, "glm", agg.mergeConfig.Preferred)
		assert.Equal(t, "markers", agg.mergeConfig.OnConflict)
	})

	t.Run("uses default config when nil", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)

		require.NotNil(t, agg)
		assert.Equal(t, MergeLastSuccess, agg.mergeConfig.Strategy)
		assert.Equal(t, "skip", agg.mergeConfig.OnConflict)
	})

	t.Run("creates aggregator with defaults", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(5*time.Minute, nil)

		require.NotNil(t, agg)
		assert.Equal(t, 5*time.Minute, agg.timeout)
		assert.Equal(t, MergeLastSuccess, agg.mergeConfig.Strategy)
	})
}

func TestGetMergeConfig(t *testing.T) {
	t.Run("returns copy of merge config", func(t *testing.T) {
		config := &MergeConfig{
			Strategy:   MergeByPriority,
			Priority:   []string{"glm", "kimi"},
			Preferred:  "glm",
			OnConflict: "skip",
		}
		agg := NewAggregatorWithoutTracking(10*time.Second, config)

		retrieved := agg.GetMergeConfig()

		require.NotNil(t, retrieved)
		assert.Equal(t, MergeByPriority, retrieved.Strategy)
		assert.Equal(t, []string{"glm", "kimi"}, retrieved.Priority)

		retrieved.Priority[0] = "modified"

		original := agg.GetMergeConfig()
		assert.Equal(t, "glm", original.Priority[0])
	})
}

func TestSetMergeConfig(t *testing.T) {
	t.Run("updates merge config", func(t *testing.T) {
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)

		newConfig := &MergeConfig{
			Strategy:   MergeConcat,
			Preferred:  "kimi",
			OnConflict: "markers",
		}

		agg.SetMergeConfig(newConfig)

		assert.Equal(t, MergeConcat, agg.mergeConfig.Strategy)
		assert.Equal(t, "kimi", agg.mergeConfig.Preferred)
	})
}

func TestGetMergedOutput(t *testing.T) {
	t.Run("returns nil before merge", func(t *testing.T) {
		config := &MergeConfig{
			Strategy:   MergeLastSuccess,
			OnConflict: "skip",
		}
		agg := NewAggregatorWithoutTracking(10*time.Second, config)

		_ = agg.AddResult("worker-1", &WorkerResult{
			WorkerID: "worker-1",
			Status:   "completed",
			Output:   "output",
		})

		merged := agg.GetMergedOutput()

		assert.Nil(t, merged)
	})

	t.Run("returns merged output after merge", func(t *testing.T) {
		config := &MergeConfig{
			Strategy:   MergeLastSuccess,
			OnConflict: "skip",
		}
		agg := NewAggregatorWithoutTracking(10*time.Second, config)

		_ = agg.AddResult("worker-1", &WorkerResult{
			WorkerID: "worker-1",
			Status:   "completed",
			Output:   "output",
		})

		_, _ = agg.Merge()
		merged := agg.GetMergedOutput()

		require.NotNil(t, merged)
		assert.Equal(t, "output", merged.Content)
	})
}

func TestGetMergeDecisions(t *testing.T) {
	t.Run("returns copy of merge decisions", func(t *testing.T) {
		config := &MergeConfig{
			Strategy:   MergeLastSuccess,
			OnConflict: "skip",
		}
		agg := NewAggregatorWithoutTracking(10*time.Second, config)

		_ = agg.AddResult("worker-1", &WorkerResult{
			WorkerID: "worker-1",
			Status:   "completed",
			Output:   "output",
		})

		_, _ = agg.Merge()

		decisions := agg.GetMergeDecisions()

		require.Len(t, decisions, 1)
		assert.Equal(t, "worker-1", decisions[0].SelectedWorker)

		decisions[0].SelectedWorker = "modified"

		original := agg.GetMergeDecisions()
		assert.Equal(t, "worker-1", original[0].SelectedWorker)
	})
}

func TestGetAggregatedResultWithMerge(t *testing.T) {
	t.Run("includes merge output in aggregated result", func(t *testing.T) {
		config := &MergeConfig{
			Strategy:   MergeLastSuccess,
			OnConflict: "skip",
		}
		agg := NewAggregatorWithoutTracking(10*time.Second, config)

		_ = agg.AddResult("worker-1", &WorkerResult{
			WorkerID: "worker-1",
			Status:   "completed",
			Output:   "merged output",
		})

		_, _ = agg.Merge()

		aggregated, err := agg.GetAggregatedResult()

		require.NoError(t, err)
		require.NotNil(t, aggregated.MergedOutput)
		assert.Equal(t, "merged output", aggregated.MergedOutput.Content)
		require.NotNil(t, aggregated.MergeConfig)
		assert.Equal(t, MergeLastSuccess, aggregated.MergeConfig.Strategy)
	})
}

func TestStorePattern(t *testing.T) {
	t.Run("stores pattern from worker result", func(t *testing.T) {
		mem := memory.NewMemoryStore()
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)
		agg.SetMemoryStore(mem)

		now := time.Now()
		result := &WorkerResult{
			WorkerID:   "worker-1",
			Provider:   "moonshot",
			Model:      "glm-4",
			Status:     "completed",
			Output:     "success",
			StartTime:  now.Add(-5 * time.Minute),
			EndTime:    now,
			Cost:       0.02,
			OutputSize: 1000,
			Metadata: map[string]string{
				"task_type":    "implement",
				"method":       "acp",
				"context_size": "30000",
			},
			FileEdits: []FileEdit{
				{
					WorkerID:  "worker-1",
					FilePath:  "/test/file1.go",
					StartTime: now.Add(-5 * time.Minute),
					EndTime:   now,
				},
			},
		}

		err := agg.StorePattern(result)
		require.NoError(t, err)

		stats, err := mem.GetProviderStats("moonshot", "glm-4", "acp")
		require.NoError(t, err)
		assert.Equal(t, 1, stats.TotalExecutions)
		assert.Equal(t, 1, stats.SuccessCount)
		assert.Equal(t, 0.02, stats.AverageCost)
	})

	t.Run("stores failed pattern", func(t *testing.T) {
		mem := memory.NewMemoryStore()
		agg := NewAggregatorWithoutTracking(10*time.Second, nil)
		agg.SetMemoryStore(mem)

		result := &WorkerResult{
			WorkerID: "worker-1",
			Provider: "moonshot",
			Model:    "glm-4",
			Status:   "failed",
			Error:    "timeout error",
			Cost:     0.01,
			Metadata: map[string]string{
				"task_type": "implement",
			},
		}

		err := agg.StorePattern(result)
		require.NoError(t, err)

		stats, err := mem.GetProviderStats("moonshot", "glm-4", "acp")
		require.NoError(t, err)
		assert.Equal(t, 1, stats.TotalExecutions)
		assert.Equal(t, 1, stats.FailureCount)
		assert.Equal(t, 0.0, stats.SuccessRate)
	})
}

func TestExtractPattern(t *testing.T) {
	t.Run("extracts pattern from completed result", func(t *testing.T) {
		now := time.Now()
		result := &WorkerResult{
			WorkerID:   "worker-1",
			Provider:   "moonshot",
			Model:      "glm-4",
			Status:     "completed",
			StartTime:  now.Add(-5 * time.Minute),
			EndTime:    now,
			Cost:       0.02,
			OutputSize: 1000,
			Metadata: map[string]string{
				"task_type":    "implement",
				"method":       "acp",
				"context_size": "30000",
			},
			FileEdits: []FileEdit{
				{WorkerID: "worker-1", FilePath: "file1.go"},
				{WorkerID: "worker-1", FilePath: "file2.go"},
				{WorkerID: "worker-1", FilePath: "file3.go"},
			},
		}

		pattern := ExtractPattern(result)

		assert.Equal(t, "moonshot", pattern.Provider)
		assert.Equal(t, "glm-4", pattern.Model)
		assert.Equal(t, "implement", pattern.TaskType)
		assert.Equal(t, "acp", pattern.Method)
		assert.Equal(t, 3, pattern.FileCount)
		assert.Equal(t, 30000, pattern.ContextSize)
		assert.True(t, pattern.Success)
		assert.Equal(t, 0.02, pattern.Cost)
		assert.Equal(t, 1000, pattern.OutputSize)
		assert.Greater(t, pattern.Duration, time.Duration(0))
	})

	t.Run("extracts pattern from failed result", func(t *testing.T) {
		result := &WorkerResult{
			WorkerID: "worker-1",
			Provider: "moonshot",
			Model:    "glm-4",
			Status:   "failed",
			Error:    "timeout: operation timed out after 30s",
			Cost:     0.01,
		}

		pattern := ExtractPattern(result)

		assert.Equal(t, "moonshot", pattern.Provider)
		assert.Equal(t, "glm-4", pattern.Model)
		assert.False(t, pattern.Success)
		assert.Equal(t, "timeout", pattern.ErrorPattern)
	})

	t.Run("classifies error patterns", func(t *testing.T) {
		tests := []struct {
			errorMsg string
			expected string
		}{
			{"operation timed out", "timeout"},
			{"request timeout", "timeout"},
			{"rate limit exceeded", "rate_limit"},
			{"too many requests", "rate_limit"},
			{"unauthorized access", "auth_error"},
			{"authentication failed", "auth_error"},
			{"context limit exceeded", "context_limit"},
			{"memory exceeded", "memory_exceeded"},
			{"network connection failed", "network_error"},
			{"unknown error occurred", "unknown_error"},
		}

		for _, tt := range tests {
			pattern := ExtractPattern(&WorkerResult{
				Error: tt.errorMsg,
			})
			assert.Equal(t, tt.expected, pattern.ErrorPattern, tt.errorMsg)
		}
	})
}
