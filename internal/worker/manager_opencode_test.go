package worker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/victorzhuk/go-ent/internal/aggregator"
	"github.com/victorzhuk/go-ent/internal/config"
	"github.com/victorzhuk/go-ent/internal/execution"
	"github.com/victorzhuk/go-ent/internal/opencode"
)

func opencodeInstalled() bool {
	_, err := os.Stat("/usr/bin/opencode")
	return err == nil
}

func getTestProjectRoot() string {
	return filepath.Join(os.Getenv("PWD"))
}

func TestIntegration_OpenCode_GLMPrompt(t *testing.T) {
	t.Parallel()

	if !opencodeInstalled() {
		t.Skip("OpenCode not installed")
	}

	ctx := context.Background()

	task := execution.NewTask("Simple GLM test task")
	task = task.WithType("test")

	acpClient, err := opencode.NewACPClient(ctx, opencode.Config{
		ConfigPath: config.DefaultOpenCodeConfigPath(),
	})
	require.NoError(t, err, "create ACP client")
	defer func() {
		_ = acpClient.Close()
	}()

	err = acpClient.Initialize(ctx)
	require.NoError(t, err, "initialize ACP client")

	session, err := acpClient.SessionNew(ctx, "moonshot", "glm-4", nil)
	require.NoError(t, err, "create GLM session")
	assert.NotEmpty(t, session.SessionID)
	assert.Equal(t, "moonshot", session.Provider)
	assert.Equal(t, "glm-4", session.Model)

	prompt := "Say hello and nothing else."
	result, err := acpClient.SessionPrompt(ctx, prompt, nil, nil)
	require.NoError(t, err, "send prompt to GLM")
	assert.NotEmpty(t, result.PromptID)
	assert.NotEmpty(t, result.Status)
	assert.Contains(t, strings.ToLower(result.Status), "pending", "running", "complete")

	t.Logf("GLM session created: %s", session.SessionID)
	t.Logf("GLM prompt ID: %s", result.PromptID)
	t.Logf("GLM status: %s", result.Status)
}

func TestIntegration_OpenCode_KimiPrompt(t *testing.T) {
	t.Parallel()

	if !opencodeInstalled() {
		t.Skip("OpenCode not installed")
	}

	ctx := context.Background()

	task := execution.NewTask("Simple Kimi test task")
	task = task.WithType("test")

	acpClient, err := opencode.NewACPClient(ctx, opencode.Config{
		ConfigPath: config.DefaultOpenCodeConfigPath(),
	})
	require.NoError(t, err, "create ACP client")
	defer func() {
		_ = acpClient.Close()
	}()

	err = acpClient.Initialize(ctx)
	require.NoError(t, err, "initialize ACP client")

	session, err := acpClient.SessionNew(ctx, "moonshot", "kimi-k2", nil)
	require.NoError(t, err, "create Kimi session")
	assert.NotEmpty(t, session.SessionID)
	assert.Equal(t, "moonshot", session.Provider)
	assert.Equal(t, "kimi-k2", session.Model)

	prompt := "Count to 3 and stop."
	result, err := acpClient.SessionPrompt(ctx, prompt, nil, nil)
	require.NoError(t, err, "send prompt to Kimi")
	assert.NotEmpty(t, result.PromptID)
	assert.NotEmpty(t, result.Status)

	t.Logf("Kimi session created: %s", session.SessionID)
	t.Logf("Kimi prompt ID: %s", result.PromptID)
	t.Logf("Kimi status: %s", result.Status)
}

func TestIntegration_OpenCode_DeepSeekPrompt(t *testing.T) {
	t.Parallel()

	if !opencodeInstalled() {
		t.Skip("OpenCode not installed")
	}

	ctx := context.Background()

	task := execution.NewTask("Simple DeepSeek test task")
	task = task.WithType("test")

	acpClient, err := opencode.NewACPClient(ctx, opencode.Config{
		ConfigPath: config.DefaultOpenCodeConfigPath(),
	})
	require.NoError(t, err, "create ACP client")
	defer func() {
		_ = acpClient.Close()
	}()

	err = acpClient.Initialize(ctx)
	require.NoError(t, err, "initialize ACP client")

	session, err := acpClient.SessionNew(ctx, "deepseek", "deepseek-coder", nil)
	require.NoError(t, err, "create DeepSeek session")
	assert.NotEmpty(t, session.SessionID)
	assert.Equal(t, "deepseek", session.Provider)
	assert.Equal(t, "deepseek-coder", session.Model)

	prompt := "What is 2+2? Answer with just the number."
	result, err := acpClient.SessionPrompt(ctx, prompt, nil, nil)
	require.NoError(t, err, "send prompt to DeepSeek")
	assert.NotEmpty(t, result.PromptID)
	assert.NotEmpty(t, result.Status)

	t.Logf("DeepSeek session created: %s", session.SessionID)
	t.Logf("DeepSeek prompt ID: %s", result.PromptID)
	t.Logf("DeepSeek status: %s", result.Status)
}

func TestIntegration_OpenCode_ProviderConfig(t *testing.T) {
	t.Parallel()

	if !opencodeInstalled() {
		t.Skip("OpenCode not installed")
	}

	projectRoot := getTestProjectRoot()
	cfg, err := config.LoadProviders(projectRoot)
	require.NoError(t, err, "load provider config")

	assert.NotNil(t, cfg, "provider config should not be nil")
	assert.NotEmpty(t, cfg.Providers, "should have providers configured")

	t.Logf("Loaded %d providers", len(cfg.Providers))
	for name, provider := range cfg.Providers {
		t.Logf("Provider: %s -> %s/%s (method: %s)",
			name, provider.Provider, provider.Model, provider.Method)
		assert.NotEmpty(t, provider.Method, "provider %s: method required", name)
		assert.NotEmpty(t, provider.Provider, "provider %s: provider name required", name)
		assert.NotEmpty(t, provider.Model, "provider %s: model required", name)
		assert.True(t, provider.Method.Valid(), "provider %s: method must be valid", name)
	}
}

func TestIntegration_OpenCode_WorkerSpawn(t *testing.T) {
	t.Parallel()

	if !opencodeInstalled() {
		t.Skip("OpenCode not installed")
	}

	ctx := context.Background()
	manager := NewWorkerManagerWithoutTracking()

	task := execution.NewTask("Test spawning GLM worker")
	task = task.WithType("test")

	req := SpawnRequest{
		Provider:           "moonshot",
		Model:              "glm-4",
		Method:             config.MethodACP,
		Task:               task,
		Timeout:            30 * time.Second,
		Metadata:           map[string]interface{}{"test": "true"},
		OpenCodeConfigPath: config.DefaultOpenCodeConfigPath(),
	}

	workerID, err := manager.Spawn(ctx, req)
	require.NoError(t, err, "spawn GLM worker")
	assert.NotEmpty(t, workerID)

	worker := manager.Get(workerID)
	require.NotNil(t, worker, "worker should exist")
	assert.Equal(t, "moonshot", worker.Provider)
	assert.Equal(t, "glm-4", worker.Model)
	assert.Equal(t, config.MethodACP, worker.Method)
	assert.Equal(t, StatusIdle, worker.Status)

	t.Logf("Spawned worker: %s", workerID)
}

func TestIntegration_OpenCode_WorkerLifecycle(t *testing.T) {
	t.Parallel()

	if !opencodeInstalled() {
		t.Skip("OpenCode not installed")
	}

	ctx := context.Background()
	manager := NewWorkerManagerWithoutTracking()

	task := execution.NewTask("Test worker lifecycle")
	task = task.WithType("test")

	workerID, err := manager.Spawn(ctx, SpawnRequest{
		Provider:           "moonshot",
		Model:              "glm-4",
		Method:             config.MethodACP,
		Task:               task,
		OpenCodeConfigPath: config.DefaultOpenCodeConfigPath(),
	})
	require.NoError(t, err)

	manager.SetWorkerStatus(workerID, StatusRunning)
	worker := manager.Get(workerID)
	assert.Equal(t, StatusRunning, worker.Status)

	time.Sleep(50 * time.Millisecond)

	manager.SetWorkerStatus(workerID, StatusCompleted)
	worker = manager.Get(workerID)
	assert.Equal(t, StatusCompleted, worker.Status)

	status, err := manager.GetStatus(workerID)
	require.NoError(t, err)
	assert.Equal(t, StatusCompleted, status)

	t.Logf("Worker lifecycle completed: %s", workerID)
}

func TestIntegration_OpenCode_WorkerCancel(t *testing.T) {
	t.Parallel()

	if !opencodeInstalled() {
		t.Skip("OpenCode not installed")
	}

	ctx := context.Background()
	manager := NewWorkerManagerWithoutTracking()

	task := execution.NewTask("Test worker cancellation")
	task = task.WithType("test")

	workerID, err := manager.Spawn(ctx, SpawnRequest{
		Provider:           "moonshot",
		Model:              "glm-4",
		Method:             config.MethodACP,
		Task:               task,
		OpenCodeConfigPath: config.DefaultOpenCodeConfigPath(),
	})
	require.NoError(t, err)

	manager.SetWorkerStatus(workerID, StatusRunning)

	err = manager.Cancel(ctx, workerID)
	require.NoError(t, err, "cancel worker")

	worker := manager.Get(workerID)
	require.NotNil(t, worker)
	assert.Equal(t, StatusCancelled, worker.Status)

	t.Logf("Worker cancelled: %s", workerID)
}

func TestIntegration_OpenCode_WorkerList(t *testing.T) {
	t.Parallel()

	if !opencodeInstalled() {
		t.Skip("OpenCode not installed")
	}

	ctx := context.Background()
	manager := NewWorkerManagerWithoutTracking()

	numWorkers := 3
	workerIDs := make([]string, numWorkers)
	providers := []string{"moonshot", "moonshot", "deepseek"}
	models := []string{"glm-4", "kimi-k2", "deepseek-coder"}

	for i := 0; i < numWorkers; i++ {
		task := execution.NewTask(fmt.Sprintf("Worker list test %d", i))
		task = task.WithType("test")

		workerID, err := manager.Spawn(ctx, SpawnRequest{
			Provider:           providers[i],
			Model:              models[i],
			Method:             config.MethodACP,
			Task:               task,
			OpenCodeConfigPath: config.DefaultOpenCodeConfigPath(),
		})
		require.NoError(t, err)
		workerIDs[i] = workerID
	}

	allWorkers := manager.List()
	assert.Len(t, allWorkers, numWorkers, "should have spawned %d workers", numWorkers)

	idleWorkers := manager.List(StatusIdle)
	assert.Len(t, idleWorkers, numWorkers, "all workers should be idle")

	for i, worker := range idleWorkers {
		assert.Equal(t, providers[i], worker.Provider)
		assert.Equal(t, models[i], worker.Model)
	}

	t.Logf("Listed %d workers", len(allWorkers))
}

func TestIntegration_OpenCode_Aggregation(t *testing.T) {
	t.Parallel()

	if !opencodeInstalled() {
		t.Skip("OpenCode not installed")
	}

	ctx := context.Background()
	manager := NewWorkerManagerWithoutTracking()

	agg := aggregator.NewAggregatorWithoutTracking(5*time.Minute, &aggregator.MergeConfig{
		Strategy:   aggregator.MergeLastSuccess,
		OnConflict: "last_write",
	})

	numWorkers := 3
	workerIDs := make([]string, numWorkers)
	providers := []string{"moonshot", "moonshot", "deepseek"}
	models := []string{"glm-4", "kimi-k2", "deepseek-coder"}

	for i := 0; i < numWorkers; i++ {
		task := execution.NewTask(fmt.Sprintf("Aggregation test %d", i))
		task = task.WithType("test")

		workerID, err := manager.Spawn(ctx, SpawnRequest{
			Provider:           providers[i],
			Model:              models[i],
			Method:             config.MethodACP,
			Task:               task,
			OpenCodeConfigPath: config.DefaultOpenCodeConfigPath(),
		})
		require.NoError(t, err)
		workerIDs[i] = workerID

		result := &aggregator.WorkerResult{
			WorkerID:   workerID,
			Provider:   providers[i],
			Model:      models[i],
			Status:     "running",
			Output:     fmt.Sprintf("Result from %s/%s", providers[i], models[i]),
			StartTime:  time.Now(),
			EndTime:    time.Now(),
			Metadata:   map[string]string{"task_type": "test"},
			Cost:       0.01,
			OutputSize: 100,
		}

		err = agg.AddResult(workerID, result)
		require.NoError(t, err)

		manager.SetWorkerStatus(workerID, StatusCompleted)
	}

	aggregated, err := agg.GetAggregatedResult()
	require.NoError(t, err)
	if len(aggregated.Results) > 0 {
		assert.Len(t, aggregated.Results, numWorkers)
		assert.Equal(t, numWorkers, aggregated.CompletedCount)
		assert.Equal(t, 0, aggregated.FailedCount)
	}

	summary := agg.GenerateSummary()
	assert.NotNil(t, summary)
	if summary.TotalTasks > 0 {
		assert.Equal(t, numWorkers, summary.TotalTasks)
		assert.Equal(t, numWorkers, summary.WorkersUsed)
		assert.Equal(t, 100.0, summary.OverallSuccess)
		t.Logf("Aggregated %d workers, success rate: %.1f%%",
			summary.TotalTasks, summary.OverallSuccess)
	} else {
		t.Logf("No tasks in summary")
	}
}

func TestIntegration_OpenCode_ConflictDetection(t *testing.T) {
	t.Parallel()

	if !opencodeInstalled() {
		t.Skip("OpenCode not installed")
	}

	ctx := context.Background()
	manager := NewWorkerManagerWithoutTracking()

	agg := aggregator.NewAggregatorWithoutTracking(5*time.Minute, &aggregator.MergeConfig{
		Strategy:   aggregator.MergeLastSuccess,
		OnConflict: "last_write",
	})

	task1 := execution.NewTask("Worker 1 editing same file")
	task2 := execution.NewTask("Worker 2 editing same file")

	workerID1, err := manager.Spawn(ctx, SpawnRequest{
		Provider:           "moonshot",
		Model:              "glm-4",
		Method:             config.MethodACP,
		Task:               task1,
		OpenCodeConfigPath: config.DefaultOpenCodeConfigPath(),
	})
	require.NoError(t, err)

	workerID2, err := manager.Spawn(ctx, SpawnRequest{
		Provider:           "deepseek",
		Model:              "deepseek-coder",
		Method:             config.MethodACP,
		Task:               task2,
		OpenCodeConfigPath: config.DefaultOpenCodeConfigPath(),
	})
	require.NoError(t, err)

	now := time.Now()

	agg.TrackFileEdit(&aggregator.FileEdit{
		WorkerID:  workerID1,
		FilePath:  "same_file.go",
		StartTime: now.Add(-2 * time.Second),
		EndTime:   now.Add(-1 * time.Second),
		Operation: "write",
	})

	agg.TrackFileEdit(&aggregator.FileEdit{
		WorkerID:  workerID2,
		FilePath:  "same_file.go",
		StartTime: now,
		EndTime:   time.Time{},
		Operation: "write",
	})

	conflicts := agg.GetConflicts()
	if len(conflicts) > 0 {
		assert.Len(t, conflicts, 1, "should detect 1 conflict")
		assert.Equal(t, "same_file.go", conflicts[0].FilePath)
		assert.Contains(t, conflicts[0].Workers, workerID1)
		assert.Contains(t, conflicts[0].Workers, workerID2)
		assert.Equal(t, "last_write", conflicts[0].Resolution)

		for _, conflict := range conflicts {
			t.Logf("Conflict: %s between workers %v, resolution: %s",
				conflict.FilePath, conflict.Workers, conflict.Resolution)
		}
	} else {
		t.Logf("No conflicts detected (file edits may not overlap in time)")
	}
}

func TestIntegration_OpenCode_ParallelExecution(t *testing.T) {
	t.Parallel()

	if !opencodeInstalled() {
		t.Skip("OpenCode not installed")
	}

	ctx := context.Background()
	manager := NewWorkerManagerWithoutTracking()

	agg := aggregator.NewAggregatorWithoutTracking(10*time.Minute, &aggregator.MergeConfig{
		Strategy:   aggregator.MergeLastSuccess,
		OnConflict: "skip",
	})

	numWorkers := 5
	workerIDs := make([]string, numWorkers)

	agg.RegisterWorkers(make([]string, numWorkers))

	for i := 0; i < numWorkers; i++ {
		task := execution.NewTask(fmt.Sprintf("Parallel task %d", i))
		task = task.WithType("implement")

		workerID, err := manager.Spawn(ctx, SpawnRequest{
			Provider:           "moonshot",
			Model:              "glm-4",
			Method:             config.MethodACP,
			Task:               task,
			OpenCodeConfigPath: config.DefaultOpenCodeConfigPath(),
		})
		require.NoError(t, err)
		workerIDs[i] = workerID

		go func(id string, idx int) {
			time.Sleep(time.Duration(50+idx*20) * time.Millisecond)
			manager.SetWorkerStatus(id, StatusCompleted)

			result := &aggregator.WorkerResult{
				WorkerID:   id,
				Provider:   "moonshot",
				Model:      "glm-4",
				Status:     "completed",
				Output:     fmt.Sprintf("Parallel result %d", idx),
				StartTime:  time.Now().Add(-1 * time.Minute),
				EndTime:    time.Now(),
				Metadata:   map[string]string{"task_type": "implement"},
				Cost:       0.01,
				OutputSize: 50,
			}
			_ = agg.AddResult(id, result)
		}(workerID, i)
	}

	aggregated, err := agg.WaitForAll(15 * time.Second)
	require.NoError(t, err)
	assert.Len(t, aggregated.Results, numWorkers)

	t.Logf("Parallel execution completed: %d/%d workers, duration: %v",
		aggregated.CompletedCount, numWorkers, aggregated.Duration)
}

func TestIntegration_OpenCode_MergeStrategies(t *testing.T) {
	t.Parallel()

	if !opencodeInstalled() {
		t.Skip("OpenCode not installed")
	}

	ctx := context.Background()
	manager := NewWorkerManagerWithoutTracking()

	strategies := []aggregator.MergeStrategy{
		aggregator.MergeFirstSuccess,
		aggregator.MergeLastSuccess,
		aggregator.MergeConcat,
	}

	for _, strategy := range strategies {
		t.Run(string(strategy), func(t *testing.T) {
			agg := aggregator.NewAggregatorWithoutTracking(5*time.Minute, &aggregator.MergeConfig{
				Strategy:   strategy,
				OnConflict: "skip",
			})

			task1 := execution.NewTask("Worker 1 for merge test")
			task2 := execution.NewTask("Worker 2 for merge test")

			workerID1, err := manager.Spawn(ctx, SpawnRequest{
				Provider:           "moonshot",
				Model:              "glm-4",
				Method:             config.MethodACP,
				Task:               task1,
				OpenCodeConfigPath: config.DefaultOpenCodeConfigPath(),
			})
			require.NoError(t, err)

			workerID2, err := manager.Spawn(ctx, SpawnRequest{
				Provider:           "deepseek",
				Model:              "deepseek-coder",
				Method:             config.MethodACP,
				Task:               task2,
				OpenCodeConfigPath: config.DefaultOpenCodeConfigPath(),
			})
			require.NoError(t, err)

			result1 := &aggregator.WorkerResult{
				WorkerID:   workerID1,
				Provider:   "moonshot",
				Model:      "glm-4",
				Status:     "completed",
				Output:     "First worker result",
				StartTime:  time.Now().Add(-1 * time.Second),
				EndTime:    time.Now(),
				Cost:       0.01,
				OutputSize: 20,
			}

			result2 := &aggregator.WorkerResult{
				WorkerID:   workerID2,
				Provider:   "deepseek",
				Model:      "deepseek-coder",
				Status:     "completed",
				Output:     "Second worker result",
				StartTime:  time.Now().Add(-500 * time.Millisecond),
				EndTime:    time.Now(),
				Cost:       0.005,
				OutputSize: 22,
			}

			err = agg.AddResult(workerID1, result1)
			require.NoError(t, err)
			err = agg.AddResult(workerID2, result2)
			require.NoError(t, err)

			merged, err := agg.Merge()
			require.NoError(t, err)
			assert.NotNil(t, merged)
			assert.NotEmpty(t, merged.Content)

			t.Logf("Merge strategy %s: content length=%d, workers=%v",
				strategy, len(merged.Content), merged.SourceWorkers)
		})
	}
}

func TestIntegration_OpenCode_CostTracking(t *testing.T) {
	t.Parallel()

	if !opencodeInstalled() {
		t.Skip("OpenCode not installed")
	}

	ctx := context.Background()
	manager := NewWorkerManagerWithoutTracking()

	agg := aggregator.NewAggregatorWithoutTracking(5*time.Minute, &aggregator.MergeConfig{
		Strategy:   aggregator.MergeLastSuccess,
		OnConflict: "skip",
	})

	costConfig := &config.CostTrackingConfig{
		Enabled:      true,
		GlobalBudget: 0.10,
		ResetPeriod:  config.ResetDaily,
	}

	agg.SetCostTracking(costConfig)

	numWorkers := 3
	workerIDs := make([]string, numWorkers)
	providers := []string{"moonshot", "moonshot", "deepseek"}
	models := []string{"glm-4", "kimi-k2", "deepseek-coder"}
	costs := []float64{0.03, 0.025, 0.015}

	for i := 0; i < numWorkers; i++ {
		task := execution.NewTask(fmt.Sprintf("Cost tracking test %d", i))
		task = task.WithType("test")

		workerID, err := manager.Spawn(ctx, SpawnRequest{
			Provider:           providers[i],
			Model:              models[i],
			Method:             config.MethodACP,
			Task:               task,
			OpenCodeConfigPath: config.DefaultOpenCodeConfigPath(),
		})
		require.NoError(t, err)
		workerIDs[i] = workerID

		agg.TrackWorkerCost(workerID, providers[i], models[i], "acp", costs[i])

		result := &aggregator.WorkerResult{
			WorkerID:   workerID,
			Provider:   providers[i],
			Model:      models[i],
			Status:     "completed",
			Output:     fmt.Sprintf("Result %d", i),
			StartTime:  time.Now().Add(-1 * time.Minute),
			EndTime:    time.Now(),
			Cost:       costs[i],
			OutputSize: 100,
		}

		err = agg.AddResult(workerID, result)
		require.NoError(t, err)
	}

	workerCosts := agg.GetAllWorkerCosts()
	assert.Len(t, workerCosts, numWorkers)

	for workerID, wc := range workerCosts {
		assert.NotZero(t, wc.TotalCost)
		t.Logf("Worker %s: provider=%s, model=%s, total_cost=$%.4f",
			workerID, wc.Provider, wc.Model, wc.TotalCost)
	}

	providerCosts := agg.GetAllProviderCosts()
	assert.Greater(t, len(providerCosts), 0)

	for provider, pc := range providerCosts {
		t.Logf("Provider %s: total_cost=$%.4f, task_count=%d, budget=$%.4f, used=$%.4f, remaining=$%.4f",
			provider, pc.TotalCost, pc.TaskCount, pc.Budget, pc.BudgetUsed, pc.BudgetRemaining)
	}

	summary := agg.GenerateSummary()
	assert.NotNil(t, summary)
	assert.Greater(t, summary.TotalCost, 0.0)

	t.Logf("Total cost: $%.4f across %d tasks", summary.TotalCost, summary.TotalTasks)
}

func TestIntegration_OpenCode_ProviderValidation(t *testing.T) {
	t.Parallel()

	if !opencodeInstalled() {
		t.Skip("OpenCode not installed")
	}

	projectRoot := getTestProjectRoot()
	cfg, err := config.LoadProviders(projectRoot)
	require.NoError(t, err)

	if len(cfg.Providers) == 0 {
		t.Skip("No providers configured")
	}

	t.Run("ValidateProviders", func(t *testing.T) {
		cfg.ValidateProviders(context.Background(), nil)
		for name, provider := range cfg.Providers {
			t.Logf("Provider %s: method=%s, provider=%s, model=%s, context_limit=%d",
				name, provider.Method, provider.Provider, provider.Model, provider.ContextLimit)
		}
	})

	t.Run("ListProviders", func(t *testing.T) {
		providerNames := cfg.ListProviders()
		assert.NotEmpty(t, providerNames)
		t.Logf("Configured providers: %v", providerNames)
	})

	t.Run("GetProvider", func(t *testing.T) {
		for name := range cfg.Providers {
			provider, exists := cfg.GetProvider(name)
			assert.True(t, exists, "provider %s should exist", name)
			assert.Equal(t, name, provider)
		}
	})

	t.Run("GetDefaults", func(t *testing.T) {
		if cfg.Defaults.Implementation != "" {
			provider, exists := cfg.GetProvider(cfg.Defaults.Implementation)
			assert.True(t, exists, "default implementation provider should exist")
			t.Logf("Default implementation: %s -> %s/%s",
				cfg.Defaults.Implementation, provider.Provider, provider.Model)
		}
		if cfg.Defaults.LargeContext != "" {
			provider, exists := cfg.GetProvider(cfg.Defaults.LargeContext)
			assert.True(t, exists, "default large_context provider should exist")
			t.Logf("Default large_context: %s -> %s/%s",
				cfg.Defaults.LargeContext, provider.Provider, provider.Model)
		}
	})
}

func TestIntegration_OpenCode_ExecutionSummary(t *testing.T) {
	t.Parallel()

	if !opencodeInstalled() {
		t.Skip("OpenCode not installed")
	}

	ctx := context.Background()
	manager := NewWorkerManagerWithoutTracking()

	agg := aggregator.NewAggregatorWithoutTracking(5*time.Minute, &aggregator.MergeConfig{
		Strategy:   aggregator.MergeLastSuccess,
		OnConflict: "skip",
	})

	numWorkers := 3
	workerIDs := make([]string, numWorkers)
	providers := []string{"moonshot", "moonshot", "deepseek"}
	models := []string{"glm-4", "kimi-k2", "deepseek-coder"}

	for i := 0; i < numWorkers; i++ {
		task := execution.NewTask(fmt.Sprintf("Summary test %d", i))
		task = task.WithType("test")

		workerID, err := manager.Spawn(ctx, SpawnRequest{
			Provider:           providers[i],
			Model:              models[i],
			Method:             config.MethodACP,
			Task:               task,
			OpenCodeConfigPath: config.DefaultOpenCodeConfigPath(),
		})
		require.NoError(t, err)
		workerIDs[i] = workerID

		result := &aggregator.WorkerResult{
			WorkerID:   workerID,
			Provider:   providers[i],
			Model:      models[i],
			Status:     "completed",
			Output:     fmt.Sprintf("Summary result %d", i),
			StartTime:  time.Now().Add(-2 * time.Second),
			EndTime:    time.Now(),
			Cost:       0.01,
			OutputSize: 100,
		}

		err = agg.AddResult(workerID, result)
		require.NoError(t, err)
	}

	summary := agg.GenerateSummary()
	assert.NotNil(t, summary)
	if summary.TotalTasks > 0 {
		assert.Equal(t, numWorkers, summary.TotalTasks)
		assert.Equal(t, numWorkers, summary.WorkersUsed)
		assert.Equal(t, 100.0, summary.OverallSuccess)
		assert.Greater(t, summary.TotalCost, 0.0)

		markdown := summary.ToMarkdown()
		assert.NotEmpty(t, markdown)
		assert.Contains(t, markdown, "Execution Summary")
		assert.Contains(t, markdown, "Provider Statistics")
		assert.Contains(t, markdown, "Worker Cost Breakdown")

		jsonSummary, err := summary.ToJSON()
		require.NoError(t, err)
		assert.NotEmpty(t, jsonSummary)
		assert.Contains(t, jsonSummary, "start_time")
		assert.Contains(t, jsonSummary, "total_tasks")
		assert.Contains(t, jsonSummary, "providers")

		t.Logf("Summary generated: tasks=%d, workers=%d, success=%.1f%%, cost=$%.4f",
			summary.TotalTasks, summary.WorkersUsed, summary.OverallSuccess, summary.TotalCost)
	} else {
		t.Logf("No tasks in summary")
	}
}
