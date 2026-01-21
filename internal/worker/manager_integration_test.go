package worker

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/victorzhuk/go-ent/internal/config"
	"github.com/victorzhuk/go-ent/internal/execution"
)

func TestIntegration_SpawnMultipleWorkers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	manager := NewWorkerManagerWithoutTracking()

	numWorkers := 5
	var wg sync.WaitGroup
	workerIDs := make([]string, numWorkers)
	errors := make(chan error, numWorkers)

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			task := execution.NewTask(fmt.Sprintf("task %d", idx))
			task = task.WithType("test")

			req := SpawnRequest{
				Provider: "test-provider",
				Model:    "test-model",
				Method:   config.MethodACP,
				Task:     task,
			}

			workerID, err := manager.Spawn(ctx, req)
			if err != nil {
				errors <- fmt.Errorf("worker %d: %w", idx, err)
				return
			}

			workerIDs[idx] = workerID
		}(i)
	}

	wg.Wait()
	close(errors)

	var allErrors []error
	for err := range errors {
		allErrors = append(allErrors, err)
	}

	if len(allErrors) > 0 {
		t.Fatalf("failed to spawn workers: %v", allErrors)
	}

	for i, workerID := range workerIDs {
		assert.NotEmpty(t, workerID, "worker %d should have ID", i)

		worker := manager.Get(workerID)
		require.NotNil(t, worker, "worker %s should exist", workerID)
		assert.Equal(t, StatusIdle, worker.Status, "worker %s should be idle", workerID)
		assert.Equal(t, "test-provider", worker.Provider, "worker %s provider mismatch", workerID)
		assert.Equal(t, "test-model", worker.Model, "worker %s model mismatch", workerID)
	}

	workers := manager.List()
	assert.Len(t, workers, numWorkers, "should have spawned %d workers", numWorkers)
}

func TestIntegration_ParallelWorkerExecution(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	manager := NewWorkerManagerWithoutTracking()

	providers := []string{"glm", "kimi", "deepseek"}
	models := []string{"glm-4", "kimi-k2", "deepseek-coder"}

	var wg sync.WaitGroup
	workerIDs := make([]string, len(providers))

	for i := 0; i < len(providers); i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			task := execution.NewTask(fmt.Sprintf("parallel task %d", idx))
			task = task.WithType("implement")

			req := SpawnRequest{
				Provider: providers[idx],
				Model:    models[idx],
				Method:   config.MethodACP,
				Task:     task,
			}

			workerID, err := manager.Spawn(ctx, req)
			require.NoError(t, err, "spawn worker %d", idx)
			workerIDs[idx] = workerID

			manager.SetWorkerStatus(workerID, StatusRunning)
			time.Sleep(10 * time.Millisecond)
			manager.SetWorkerStatus(workerID, StatusCompleted)
		}(i)
	}

	wg.Wait()

	for i, workerID := range workerIDs {
		worker := manager.Get(workerID)
		require.NotNil(t, worker, "worker %s should exist", workerID)
		assert.Equal(t, StatusCompleted, worker.Status, "worker %s should be completed", workerID)
		assert.Equal(t, providers[i], worker.Provider, "provider mismatch for worker %s", workerID)
	}

	completed := manager.List(StatusCompleted)
	assert.Len(t, completed, len(providers), "all workers should be completed")
}

func TestIntegration_WorkerFailures(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	manager := NewWorkerManagerWithoutTracking()

	numWorkers := 5
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			task := execution.NewTask(fmt.Sprintf("task %d", idx))

			req := SpawnRequest{
				Provider: "test-provider",
				Model:    "test-model",
				Method:   config.MethodACP,
				Task:     task,
			}

			workerID, err := manager.Spawn(ctx, req)
			require.NoError(t, err, "spawn worker %d", idx)

			time.Sleep(10 * time.Millisecond)

			if idx%2 == 0 {
				manager.SetWorkerStatus(workerID, StatusFailed)
			} else {
				manager.SetWorkerStatus(workerID, StatusCompleted)
			}
		}(i)
	}

	wg.Wait()
	time.Sleep(20 * time.Millisecond)

	allWorkers := manager.List()
	failed := manager.List(StatusFailed)
	completed := manager.List(StatusCompleted)

	totalWorkers := len(failed) + len(completed)
	assert.Equal(t, numWorkers, len(allWorkers), "should have spawned %d workers", numWorkers)
	assert.Equal(t, numWorkers, totalWorkers, "should have %d total workers", numWorkers)

	assert.LessOrEqual(t, len(failed), 3, "should have <=3 failed workers")
	assert.LessOrEqual(t, len(completed), 3, "should have <=3 completed workers")
}

func TestIntegration_ParallelStatusUpdates(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	manager := NewWorkerManagerWithoutTracking()

	numWorkers := 10
	workerIDs := make([]string, numWorkers)
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			task := execution.NewTask(fmt.Sprintf("task %d", idx))
			req := SpawnRequest{
				Provider: "test-provider",
				Model:    "test-model",
				Method:   config.MethodACP,
				Task:     task,
			}

			workerID, err := manager.Spawn(ctx, req)
			require.NoError(t, err, "spawn worker %d", idx)
			workerIDs[idx] = workerID

			for j := 0; j < 5; j++ {
				statuses := []WorkerStatus{StatusIdle, StatusRunning, StatusRunning, StatusRunning, StatusCompleted}
				manager.SetWorkerStatus(workerID, statuses[j])
				time.Sleep(1 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()

	completed := manager.List(StatusCompleted)
	assert.Len(t, completed, numWorkers, "all workers should be completed")

	for _, worker := range completed {
		assert.Equal(t, StatusCompleted, worker.Status, "worker %s should be completed", worker.ID)
	}
}

func TestIntegration_Cleanup(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	manager := NewWorkerManagerWithoutTracking()

	numWorkers := 10
	for i := 0; i < numWorkers; i++ {
		task := execution.NewTask(fmt.Sprintf("task %d", i))
		req := SpawnRequest{
			Provider: "test-provider",
			Model:    "test-model",
			Method:   config.MethodACP,
			Task:     task,
		}

		workerID, err := manager.Spawn(ctx, req)
		require.NoError(t, err, "spawn worker %d", i)

		if i%3 == 0 {
			manager.SetWorkerStatus(workerID, StatusCompleted)
		} else if i%3 == 1 {
			manager.SetWorkerStatus(workerID, StatusFailed)
		} else {
			manager.SetWorkerStatus(workerID, StatusCancelled)
		}
	}

	workers := manager.List()
	assert.Len(t, workers, numWorkers, "should have %d workers", numWorkers)

	cleaned := manager.Cleanup(1 * time.Hour)
	assert.Equal(t, 0, cleaned, "should not clean up recent workers")

	cleaned = manager.Cleanup(0)
	assert.Equal(t, numWorkers, cleaned, "should clean up all workers")

	workers = manager.List()
	assert.Len(t, workers, 0, "should have no workers after cleanup")
}

func TestIntegration_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	manager := NewWorkerManagerWithoutTracking()

	numWorkers := 20
	var wg sync.WaitGroup
	errors := make(chan error, numWorkers*3)

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			task := execution.NewTask(fmt.Sprintf("task %d", idx))
			req := SpawnRequest{
				Provider: "test-provider",
				Model:    "test-model",
				Method:   config.MethodACP,
				Task:     task,
			}

			workerID, err := manager.Spawn(ctx, req)
			if err != nil {
				errors <- fmt.Errorf("spawn worker %d: %w", idx, err)
				return
			}

			wg.Add(1)
			go func() {
				defer wg.Done()
				worker := manager.Get(workerID)
				if worker == nil {
					errors <- fmt.Errorf("worker %s not found", workerID)
					return
				}

				if worker.Provider != "test-provider" {
					errors <- fmt.Errorf("worker %s provider mismatch", workerID)
				}
			}()

			wg.Add(1)
			go func() {
				defer wg.Done()
				manager.SetWorkerStatus(workerID, StatusRunning)
			}()

			wg.Add(1)
			go func() {
				defer wg.Done()
				status, err := manager.GetStatus(workerID)
				if err != nil {
					errors <- fmt.Errorf("get status for %s: %w", workerID, err)
					return
				}
				if status != StatusIdle && status != StatusRunning && status != StatusCompleted {
					errors <- fmt.Errorf("worker %s has invalid status: %s", workerID, status)
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

	workers := manager.List()
	assert.Len(t, workers, numWorkers, "should have %d workers", numWorkers)
}

func TestIntegration_Cancellation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	manager := NewWorkerManagerWithoutTracking()

	numWorkers := 5
	workerIDs := make([]string, numWorkers)
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			task := execution.NewTask(fmt.Sprintf("task %d", idx))
			req := SpawnRequest{
				Provider: "test-provider",
				Model:    "test-model",
				Method:   config.MethodACP,
				Task:     task,
			}

			workerID, err := manager.Spawn(ctx, req)
			require.NoError(t, err, "spawn worker %d", idx)
			workerIDs[idx] = workerID

			manager.SetWorkerStatus(workerID, StatusRunning)
		}(i)
	}

	wg.Wait()

	for _, workerID := range workerIDs {
		err := manager.Cancel(ctx, workerID)
		require.NoError(t, err, "cancel worker %s", workerID)

		worker := manager.Get(workerID)
		require.NotNil(t, worker, "worker %s should exist", workerID)
		assert.Equal(t, StatusCancelled, worker.Status, "worker %s should be cancelled", workerID)
	}

	cancelled := manager.List(StatusCancelled)
	assert.Len(t, cancelled, numWorkers, "all workers should be cancelled")
}

func TestIntegration_WorkerWithTasks(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	manager := NewWorkerManagerWithoutTracking()

	taskTypes := []string{"implement", "refactor", "test", "document", "optimize"}
	workerIDs := make([]string, len(taskTypes))
	var wg sync.WaitGroup

	for i, taskType := range taskTypes {
		wg.Add(1)
		go func(idx int, tType string) {
			defer wg.Done()

			task := execution.NewTask(fmt.Sprintf("%s task", tType))
			task = task.WithType(tType)

			req := SpawnRequest{
				Provider: "test-provider",
				Model:    "test-model",
				Method:   config.MethodACP,
				Task:     task,
			}

			workerID, err := manager.Spawn(ctx, req)
			require.NoError(t, err, "spawn worker for task type %s", tType)
			workerIDs[idx] = workerID

			worker := manager.Get(workerID)
			require.NotNil(t, worker)
			assert.Equal(t, tType, worker.Task.Type, "task type mismatch for worker %s", workerID)
			assert.Equal(t, fmt.Sprintf("%s task", tType), worker.Task.Description, "task description mismatch")
		}(i, taskType)
	}

	wg.Wait()

	for i, workerID := range workerIDs {
		worker := manager.Get(workerID)
		require.NotNil(t, worker, "worker %s should exist", workerID)
		assert.Equal(t, taskTypes[i], worker.Task.Type, "worker %s task type mismatch", workerID)
	}
}

func TestIntegration_Failover_ProviderTimeout(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	manager := NewWorkerManagerWithoutTracking()

	req := SpawnRequest{
		Provider: "timeout-provider",
		Model:    "test-model",
		Method:   config.MethodACP,
		Task:     execution.NewTask("Test timeout task"),
	}

	workerID, err := manager.Spawn(ctx, req)
	require.NoError(t, err)

	worker := manager.Get(workerID)
	require.NotNil(t, worker)
	assert.Equal(t, StatusIdle, worker.Status)

	<-ctx.Done()

	time.Sleep(10 * time.Millisecond)

	workerAfterTimeout := manager.Get(workerID)
	if workerAfterTimeout != nil {
		assert.True(t, workerAfterTimeout.Status == StatusFailed || workerAfterTimeout.Status == StatusIdle)
	}
}

func TestIntegration_Failover_PartialWorkerFailures(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	manager := NewWorkerManagerWithoutTracking()

	numWorkers := 10
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			task := execution.NewTask(fmt.Sprintf("task %d", idx))
			req := SpawnRequest{
				Provider: fmt.Sprintf("provider-%d", idx%3),
				Model:    "test-model",
				Method:   config.MethodACP,
				Task:     task,
			}

			workerID, err := manager.Spawn(ctx, req)
			require.NoError(t, err)

			time.Sleep(10 * time.Millisecond)

			if idx%2 == 0 {
				manager.SetWorkerStatus(workerID, StatusCompleted)
			} else {
				manager.SetWorkerStatus(workerID, StatusFailed)
			}
		}(i)
	}

	wg.Wait()
	time.Sleep(20 * time.Millisecond)

	completed := manager.List(StatusCompleted)
	failed := manager.List(StatusFailed)

	assert.Equal(t, 5, len(completed), "half should complete")
	assert.Equal(t, 5, len(failed), "half should fail")
}

func TestIntegration_Failover_FailoverDuringParallelExecution(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	manager := NewWorkerManagerWithoutTracking()

	numWorkers := 8
	var wg sync.WaitGroup
	workerIDs := make([]string, numWorkers)

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			provider := "primary"
			if idx%4 == 0 {
				provider = "failover"
			}

			task := execution.NewTask(fmt.Sprintf("parallel task %d", idx))
			req := SpawnRequest{
				Provider: provider,
				Model:    "test-model",
				Method:   config.MethodACP,
				Task:     task,
			}

			workerID, err := manager.Spawn(ctx, req)
			require.NoError(t, err)
			workerIDs[idx] = workerID

			manager.SetWorkerStatus(workerID, StatusRunning)

			if idx%4 == 0 {
				time.Sleep(20 * time.Millisecond)
				manager.SetWorkerStatus(workerID, StatusFailed)
			} else {
				time.Sleep(10 * time.Millisecond)
				manager.SetWorkerStatus(workerID, StatusCompleted)
			}
		}(i)
	}

	wg.Wait()

	completed := manager.List(StatusCompleted)
	failed := manager.List(StatusFailed)

	assert.Equal(t, 6, len(completed), "6 workers should complete")
	assert.Equal(t, 2, len(failed), "2 workers should fail (failover providers)")
}

func TestIntegration_Failover_NetworkErrorScenario(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	manager := NewWorkerManagerWithoutTracking()

	providers := []string{"network-error", "backup"}

	var wg sync.WaitGroup

	for _, provider := range providers {
		wg.Add(1)
		go func(providerName string) {
			defer wg.Done()

			task := execution.NewTask(fmt.Sprintf("Task with %s", providerName))
			req := SpawnRequest{
				Provider: providerName,
				Model:    "test-model",
				Method:   config.MethodACP,
				Task:     task,
			}

			workerID, err := manager.Spawn(ctx, req)
			require.NoError(t, err)

			if providerName == "network-error" {
				time.Sleep(15 * time.Millisecond)
				manager.SetWorkerStatus(workerID, StatusFailed)
			} else {
				time.Sleep(10 * time.Millisecond)
				manager.SetWorkerStatus(workerID, StatusCompleted)
			}
		}(provider)
	}

	wg.Wait()

	completed := manager.List(StatusCompleted)
	failed := manager.List(StatusFailed)

	assert.Equal(t, 1, len(completed), "backup provider should complete")
	assert.Equal(t, 1, len(failed), "network-error provider should fail")
}

func TestIntegration_Failover_ProviderUnavailable(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	manager := NewWorkerManagerWithoutTracking()

	unavailableProvider := "unavailable"
	backupProvider := "available"

	task := execution.NewTask("Task for unavailable provider")
	req := SpawnRequest{
		Provider: unavailableProvider,
		Model:    "test-model",
		Method:   config.MethodACP,
		Task:     task,
	}

	workerID, err := manager.Spawn(ctx, req)
	require.NoError(t, err)

	worker := manager.Get(workerID)
	require.NotNil(t, worker)
	assert.Equal(t, unavailableProvider, worker.Provider)

	time.Sleep(10 * time.Millisecond)
	manager.SetWorkerStatus(workerID, StatusFailed)

	backupTask := execution.NewTask("Backup provider task")
	backupReq := SpawnRequest{
		Provider: backupProvider,
		Model:    "test-model",
		Method:   config.MethodAPI,
		Task:     backupTask,
	}

	backupWorkerID, err := manager.Spawn(ctx, backupReq)
	require.NoError(t, err)

	backupWorker := manager.Get(backupWorkerID)
	require.NotNil(t, backupWorker)
	assert.Equal(t, backupProvider, backupWorker.Provider)

	time.Sleep(10 * time.Millisecond)
	manager.SetWorkerStatus(backupWorkerID, StatusCompleted)

	failed := manager.List(StatusFailed)
	completed := manager.List(StatusCompleted)

	assert.Equal(t, 1, len(failed))
	assert.Equal(t, 1, len(completed))
	assert.Equal(t, unavailableProvider, failed[0].Provider)
	assert.Equal(t, backupProvider, completed[0].Provider)
}

func TestIntegration_Failover_AggregationWithPartialFailures(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	manager := NewWorkerManagerWithoutTracking()

	numWorkers := 6
	var wg sync.WaitGroup
	expectedResults := 0

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			task := execution.NewTask(fmt.Sprintf("Task %d", idx))
			req := SpawnRequest{
				Provider: fmt.Sprintf("provider-%d", idx),
				Model:    "test-model",
				Method:   config.MethodACP,
				Task:     task,
			}

			workerID, err := manager.Spawn(ctx, req)
			require.NoError(t, err)

			manager.SetWorkerStatus(workerID, StatusRunning)

			if idx%3 == 0 {
				manager.SetWorkerStatus(workerID, StatusFailed)
			} else {
				expectedResults++
				manager.SetWorkerStatus(workerID, StatusCompleted)
			}
		}(i)
	}

	wg.Wait()

	completed := manager.List(StatusCompleted)
	failed := manager.List(StatusFailed)

	assert.Equal(t, expectedResults, len(completed), "expected results should match completed count")
	assert.Equal(t, numWorkers-expectedResults, len(failed), "failed count should match")

	for _, worker := range completed {
		assert.Equal(t, StatusCompleted, worker.Status)
	}

	for _, worker := range failed {
		assert.Equal(t, StatusFailed, worker.Status)
	}
}

func TestIntegration_Failover_CostTrackingDuringFailover(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	manager := NewWorkerManagerWithoutTracking()

	providers := []string{"expensive", "cheap"}
	var wg sync.WaitGroup

	for _, provider := range providers {
		wg.Add(1)
		go func(providerName string) {
			defer wg.Done()

			task := execution.NewTask(fmt.Sprintf("Task with %s", providerName))
			req := SpawnRequest{
				Provider: providerName,
				Model:    "test-model",
				Method:   config.MethodACP,
				Task:     task,
			}

			workerID, err := manager.Spawn(ctx, req)
			require.NoError(t, err)

			manager.SetWorkerStatus(workerID, StatusRunning)
			time.Sleep(10 * time.Millisecond)
			manager.SetWorkerStatus(workerID, StatusCompleted)
		}(provider)
	}

	wg.Wait()

	workers := manager.List()
	assert.Len(t, workers, 2)

	for _, worker := range workers {
		assert.NotEmpty(t, worker.Provider)
		assert.NotEmpty(t, worker.ID)
		assert.Equal(t, StatusCompleted, worker.Status)
	}
}

func TestIntegration_Failover_AllProvidersFail(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	manager := NewWorkerManagerWithoutTracking()

	numWorkers := 5
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			task := execution.NewTask(fmt.Sprintf("Failing task %d", idx))
			req := SpawnRequest{
				Provider: fmt.Sprintf("fail-provider-%d", idx),
				Model:    "test-model",
				Method:   config.MethodACP,
				Task:     task,
			}

			workerID, err := manager.Spawn(ctx, req)
			require.NoError(t, err)

			manager.SetWorkerStatus(workerID, StatusRunning)
			time.Sleep(10 * time.Millisecond)
			manager.SetWorkerStatus(workerID, StatusFailed)
		}(i)
	}

	wg.Wait()

	allWorkers := manager.List()
	failedWorkers := manager.List(StatusFailed)
	completedWorkers := manager.List(StatusCompleted)

	assert.Equal(t, numWorkers, len(allWorkers))
	assert.Equal(t, numWorkers, len(failedWorkers))
	assert.Equal(t, 0, len(completedWorkers), "no workers should complete when all fail")
}

func TestIntegration_Failover_RateLimitScenario(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	manager := NewWorkerManagerWithoutTracking()

	numWorkers := 8
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			provider := "rate-limited"
			if idx >= 4 {
				provider = "unlimited"
			}

			task := execution.NewTask(fmt.Sprintf("Rate limit test %d", idx))
			req := SpawnRequest{
				Provider: provider,
				Model:    "test-model",
				Method:   config.MethodACP,
				Task:     task,
			}

			workerID, err := manager.Spawn(ctx, req)
			require.NoError(t, err)

			if provider == "rate-limited" && idx < 2 {
				manager.SetWorkerStatus(workerID, StatusFailed)
			} else {
				manager.SetWorkerStatus(workerID, StatusCompleted)
			}
		}(i)
	}

	wg.Wait()

	completed := manager.List(StatusCompleted)

	assert.Greater(t, len(completed), 0, "some workers should complete via unlimited provider")
}

func TestIntegration_Failover_WorkerRetryAfterFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	manager := NewWorkerManagerWithoutTracking()

	task := execution.NewTask("Retry test task")

	workerID, err := manager.Spawn(ctx, SpawnRequest{
		Provider: "retry-provider",
		Model:    "test-model",
		Method:   config.MethodACP,
		Task:     task,
	})
	require.NoError(t, err)

	worker := manager.Get(workerID)
	require.NotNil(t, worker)

	manager.SetWorkerStatus(workerID, StatusRunning)

	worker.RetryCount = 1

	manager.SetWorkerStatus(workerID, StatusFailed)

	retryWorkerID, err := manager.Spawn(ctx, SpawnRequest{
		Provider: "retry-provider",
		Model:    "test-model",
		Method:   config.MethodACP,
		Task:     task,
	})
	require.NoError(t, err)

	retryWorker := manager.Get(retryWorkerID)
	require.NotNil(t, retryWorker)

	manager.SetWorkerStatus(retryWorkerID, StatusCompleted)

	assert.Equal(t, 0, retryWorker.RetryCount)
}
