package aggregator

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/victorzhuk/go-ent/internal/memory"
)

func TestIntegration_AggregateParallelResults(t *testing.T) {
	t.Parallel()

	agg := NewAggregatorWithoutTracking(10*time.Second, nil)

	numWorkers := 5
	workerIDs := make([]string, numWorkers)
	providers := []string{"glm", "kimi", "deepseek", "anthropic", "openai"}
	models := []string{"glm-4", "kimi-k2", "deepseek-coder", "claude-3-opus", "gpt-4"}

	var wg sync.WaitGroup
	startTime := time.Now()

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			workerID := fmt.Sprintf("worker-%d", idx)
			workerIDs[idx] = workerID

			time.Sleep(time.Duration(idx*10) * time.Millisecond)

			result := &WorkerResult{
				WorkerID:  workerID,
				Provider:  providers[idx],
				Model:     models[idx],
				Status:    "completed",
				Output:    fmt.Sprintf("output from %s", workerID),
				StartTime: startTime,
				EndTime:   startTime.Add(time.Duration(100+idx*50) * time.Millisecond),
				Cost:      float64(0.01 + float64(idx)*0.005),
				Metadata: map[string]string{
					"method": "acp",
				},
			}

			err := agg.AddResult(workerID, result)
			require.NoError(t, err, "add result for worker %s", workerID)
		}(i)
	}

	wg.Wait()

	aggregated, err := agg.GetAggregatedResult()
	require.NoError(t, err)

	assert.Equal(t, numWorkers, aggregated.CompletedCount, "should have %d completed workers", numWorkers)
	assert.Equal(t, 0, aggregated.FailedCount, "should have no failed workers")
	assert.Equal(t, 100.0, aggregated.SuccessRate, "success rate should be 100%")
	assert.Len(t, aggregated.Results, numWorkers, "should have %d results", numWorkers)
	assert.Greater(t, aggregated.Duration, time.Duration(0), "duration should be positive")

	for i, workerID := range workerIDs {
		result, exists := aggregated.Results[workerID]
		require.True(t, exists, "result for worker %s should exist", workerID)
		assert.Equal(t, providers[i], result.Provider, "provider mismatch for worker %s", workerID)
		assert.Equal(t, models[i], result.Model, "model mismatch for worker %s", workerID)
		assert.NotEmpty(t, result.Output, "output should not be empty for worker %s", workerID)
	}
}

func TestIntegration_ParallelConflicts(t *testing.T) {
	t.Parallel()

	agg := NewAggregatorWithoutTracking(10*time.Second, nil)

	now := time.Now()
	numWorkers := 4
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			workerID := fmt.Sprintf("worker-%d", idx)

			time.Sleep(time.Duration(idx*5) * time.Millisecond)

			result := &WorkerResult{
				WorkerID: workerID,
				Provider: "test-provider",
				Model:    "test-model",
				Status:   "completed",
				Output:   fmt.Sprintf("output %s", workerID),
				FileEdits: []FileEdit{
					{
						WorkerID:  workerID,
						FilePath:  "/test/file.go",
						StartTime: now.Add(-1 * time.Hour),
						EndTime:   now.Add(-30 * time.Minute),
						Operation: "write",
					},
				},
			}

			_ = agg.AddResult(workerID, result)
		}(i)
	}

	wg.Wait()

	conflicts := agg.GetConflicts()
	assert.Greater(t, len(conflicts), 0, "should detect conflicts")
	assert.Equal(t, "/test/file.go", conflicts[0].FilePath, "conflict should be for expected file")

	aggregated, _ := agg.GetAggregatedResult()
	assert.Greater(t, aggregated.ConflictCount, 0, "aggregated result should show conflict count")
}

func TestIntegration_ParallelWaitForAll(t *testing.T) {
	t.Parallel()

	agg := NewAggregatorWithoutTracking(10*time.Second, nil)

	numWorkers := 5
	workerIDs := make([]string, numWorkers)
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			workerID := fmt.Sprintf("worker-%d", idx)
			workerIDs[idx] = workerID

			time.Sleep(time.Duration(50+idx*30) * time.Millisecond)

			result := &WorkerResult{
				WorkerID: workerID,
				Status:   "completed",
				Output:   fmt.Sprintf("output %s", workerID),
			}

			_ = agg.AddResult(workerID, result)
		}(i)
	}

	wg.Wait()

	aggregated, err := agg.WaitForAll(5 * time.Second)
	require.NoError(t, err)

	assert.Equal(t, numWorkers, aggregated.CompletedCount, "all workers should complete")
	assert.Equal(t, 0, aggregated.FailedCount, "no workers should fail")
}

func TestIntegration_MixedSuccessAndFailure(t *testing.T) {
	t.Parallel()

	agg := NewAggregatorWithoutTracking(10*time.Second, nil)

	numWorkers := 6
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			workerID := fmt.Sprintf("worker-%d", idx)

			time.Sleep(time.Duration(idx*10) * time.Millisecond)

			result := &WorkerResult{
				WorkerID: workerID,
				Provider: "test-provider",
				Model:    "test-model",
			}

			if idx%3 == 0 {
				result.Status = "failed"
				result.Error = fmt.Sprintf("worker %s failed", workerID)
			} else {
				result.Status = "completed"
				result.Output = fmt.Sprintf("output %s", workerID)
			}

			_ = agg.AddResult(workerID, result)
		}(i)
	}

	wg.Wait()

	aggregated, err := agg.GetAggregatedResult()
	require.NoError(t, err)

	assert.Equal(t, 4, aggregated.CompletedCount, "should have 4 completed workers")
	assert.Equal(t, 2, aggregated.FailedCount, "should have 2 failed workers")
	assert.Equal(t, 66.66666666666666, aggregated.SuccessRate, "success rate should be ~66.67%")

	completed := agg.CompletedWorkers()
	assert.Len(t, completed, 4, "should have 4 completed worker IDs")

	failed := agg.FailedWorkers()
	assert.Len(t, failed, 2, "should have 2 failed worker IDs")
}

func TestIntegration_ParallelFileEditsNoConflicts(t *testing.T) {
	t.Parallel()

	agg := NewAggregatorWithoutTracking(10*time.Second, nil)

	numWorkers := 3
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			workerID := fmt.Sprintf("worker-%d", idx)

			result := &WorkerResult{
				WorkerID: workerID,
				Status:   "completed",
				Output:   fmt.Sprintf("output %s", workerID),
				FileEdits: []FileEdit{
					{
						WorkerID:  workerID,
						FilePath:  fmt.Sprintf("/test/file%d.go", idx),
						StartTime: time.Now().Add(-1 * time.Hour),
						EndTime:   time.Now(),
						Operation: "write",
					},
				},
			}

			_ = agg.AddResult(workerID, result)
		}(i)
	}

	wg.Wait()

	conflicts := agg.GetConflicts()
	assert.Len(t, conflicts, 0, "should have no conflicts with different files")

	aggregated, _ := agg.GetAggregatedResult()
	assert.Equal(t, 0, aggregated.ConflictCount, "aggregated result should show 0 conflicts")
}

func TestIntegration_ParallelCostTracking(t *testing.T) {
	t.Parallel()

	agg := NewAggregatorWithoutTracking(10*time.Second, nil)

	numWorkers := 5
	providers := []string{"glm", "kimi", "deepseek", "glm", "kimi"}
	costs := []float64{0.02, 0.03, 0.025, 0.015, 0.035}

	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			workerID := fmt.Sprintf("worker-%d", idx)

			result := &WorkerResult{
				WorkerID:   workerID,
				Provider:   providers[idx],
				Model:      "test-model",
				Status:     "completed",
				Output:     fmt.Sprintf("output %s", workerID),
				Cost:       costs[idx],
				OutputSize: 1000 + idx*100,
			}

			_ = agg.AddResult(workerID, result)
		}(i)
	}

	wg.Wait()

	workerCosts := agg.GetAllWorkerCosts()
	assert.Len(t, workerCosts, numWorkers, "should have costs for all workers")

	for i, workerID := range []string{"worker-0", "worker-1", "worker-2", "worker-3", "worker-4"} {
		wc, exists := workerCosts[workerID]
		require.True(t, exists, "worker cost should exist for %s", workerID)
		assert.Equal(t, providers[i], wc.Provider, "provider mismatch for worker %s", workerID)
		assert.Equal(t, costs[i], wc.TotalCost, "cost mismatch for worker %s", workerID)
		assert.Equal(t, 1, wc.TaskCount, "task count should be 1 for worker %s", workerID)
	}

	providerCosts := agg.GetAllProviderCosts()
	assert.Greater(t, len(providerCosts), 0, "should have provider costs")

	glmCost, exists := providerCosts["glm"]
	require.True(t, exists, "glm provider cost should exist")
	assert.Equal(t, 0.035, glmCost.TotalCost, "glm total cost should be 0.035")

	kimiCost, exists := providerCosts["kimi"]
	require.True(t, exists, "kimi provider cost should exist")
	assert.Equal(t, 0.065, kimiCost.TotalCost, "kimi total cost should be 0.065")
}

func TestIntegration_ParallelMerge(t *testing.T) {
	t.Parallel()

	mergeConfig := &MergeConfig{
		Strategy:   MergeConcat,
		OnConflict: "skip",
	}
	agg := NewAggregatorWithoutTracking(10*time.Second, mergeConfig)

	numWorkers := 4
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			workerID := fmt.Sprintf("worker-%d", idx)

			result := &WorkerResult{
				WorkerID: workerID,
				Status:   "completed",
				Output:   fmt.Sprintf("Worker %d output\nLine 2 from worker %d", idx, idx),
			}

			_ = agg.AddResult(workerID, result)
		}(i)
	}

	wg.Wait()

	merged, err := agg.Merge()
	require.NoError(t, err, "merge should succeed")

	assert.NotEmpty(t, merged.Content, "merged content should not be empty")
	assert.Len(t, merged.SourceWorkers, numWorkers, "should have %d source workers", numWorkers)
	assert.Equal(t, numWorkers, merged.Metadata["worker_count"], "metadata should have worker count")

	for i := 0; i < numWorkers; i++ {
		assert.Contains(t, merged.Content, fmt.Sprintf("worker %d", i), "content should contain worker %d", i)
	}

	aggregated, _ := agg.GetAggregatedResult()
	assert.NotNil(t, aggregated.MergedOutput, "aggregated result should have merged output")
	assert.Equal(t, merged.Content, aggregated.MergedOutput.Content, "merged content should match")
}

func TestIntegration_ParallelMergeByPriority(t *testing.T) {
	t.Parallel()

	mergeConfig := &MergeConfig{
		Strategy:   MergeByPriority,
		Priority:   []string{"glm", "kimi", "deepseek"},
		OnConflict: "skip",
	}
	agg := NewAggregatorWithoutTracking(10*time.Second, mergeConfig)

	providers := []string{"kimi", "deepseek", "glm", "kimi"}
	numWorkers := len(providers)
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			workerID := fmt.Sprintf("worker-%d", idx)

			result := &WorkerResult{
				WorkerID: workerID,
				Provider: providers[idx],
				Model:    "test-model",
				Status:   "completed",
				Output:   fmt.Sprintf("output from %s", providers[idx]),
			}

			_ = agg.AddResult(workerID, result)
		}(i)
	}

	wg.Wait()

	merged, err := agg.Merge()
	require.NoError(t, err)

	assert.Equal(t, "output from glm", merged.Content, "should select glm output (highest priority)")
	assert.Len(t, merged.SourceWorkers, 1, "should have 1 source worker")
	assert.Equal(t, "worker-2", merged.SourceWorkers[0], "should select worker-2 (glm)")
}

func TestIntegration_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	agg := NewAggregatorWithoutTracking(10*time.Second, nil)

	numWorkers := 20
	var wg sync.WaitGroup
	errors := make(chan error, numWorkers*3)

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			workerID := fmt.Sprintf("worker-%d", idx)

			wg.Add(1)
			go func() {
				defer wg.Done()
				result := &WorkerResult{
					WorkerID: workerID,
					Status:   "completed",
					Output:   fmt.Sprintf("output %s", workerID),
					Cost:     0.01,
				}
				if err := agg.AddResult(workerID, result); err != nil {
					errors <- fmt.Errorf("add result for %s: %w", workerID, err)
				}
			}()

			wg.Add(1)
			go func() {
				defer wg.Done()
				time.Sleep(10 * time.Millisecond)
				if _, err := agg.GetResult(workerID); err != nil {
					errors <- fmt.Errorf("get result for %s: %w", workerID, err)
				}
			}()

			wg.Add(1)
			go func() {
				defer wg.Done()
				time.Sleep(20 * time.Millisecond)
				completed := agg.CompletedWorkers()
				if len(completed) > numWorkers {
					errors <- fmt.Errorf("too many completed workers: %d", len(completed))
				}
			}()
		}(i)
	}

	wg.Wait()
	close(errors)

	var allErrors []error
	for err := range errors {
		allErrors = append(allErrors, err)
	}

	if len(allErrors) > 0 {
		t.Fatalf("concurrent access errors: %v", allErrors)
	}

	aggregated, err := agg.GetAggregatedResult()
	require.NoError(t, err)
	assert.Equal(t, numWorkers, aggregated.CompletedCount, "should have %d completed workers", numWorkers)
}

func TestIntegration_PatternStorage(t *testing.T) {
	t.Parallel()

	mem := memory.NewMemoryStore()
	agg := NewAggregatorWithoutTracking(10*time.Second, nil)
	agg.SetMemoryStore(mem)

	numWorkers := 5
	taskTypes := []string{"implement", "refactor", "test", "document", "optimize"}
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			workerID := fmt.Sprintf("worker-%d", idx)

			result := &WorkerResult{
				WorkerID:   workerID,
				Provider:   "test-provider",
				Model:      "test-model",
				Status:     "completed",
				Output:     fmt.Sprintf("output %s", workerID),
				Cost:       0.01 + float64(idx)*0.005,
				OutputSize: 1000 + idx*100,
				Metadata: map[string]string{
					"task_type": taskTypes[idx],
					"method":    "acp",
				},
				FileEdits: []FileEdit{
					{WorkerID: workerID, FilePath: fmt.Sprintf("file%d.go", idx)},
				},
			}

			_ = agg.AddResult(workerID, result)
		}(i)
	}

	wg.Wait()

	totalPatterns := mem.GetTotalPatterns()
	assert.GreaterOrEqual(t, totalPatterns, numWorkers, "should have stored at least %d patterns", numWorkers)

	stats, err := mem.GetProviderStats("test-provider", "test-model", "acp")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, stats.TotalExecutions, numWorkers, "should have stats for test-provider")
	assert.Equal(t, numWorkers, stats.SuccessCount, "all patterns should be successful")
}
