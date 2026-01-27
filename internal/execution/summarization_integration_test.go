package execution

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/victorzhuk/go-ent/internal/agent"
	"github.com/victorzhuk/go-ent/internal/domain"
)

type fileContextTracker struct {
	initialSize int
	summaries   []string
	contextSize int
}

func (t *fileContextTracker) calculateContextSize(files []string) int {
	total := 0
	for _, file := range files {
		total += len(file)
	}
	return total
}

func (t *fileContextTracker) track(files []string) {
	t.contextSize = t.calculateContextSize(files)
}

func (t *fileContextTracker) countTokens() int {
	return t.contextSize / 4
}

func TestIntegration_LongExecutionWithSummarization(t *testing.T) {
	t.Run("triggers summarization at 80% threshold", func(t *testing.T) {
		testDir := filepath.Join(os.TempDir(), "go-ent-test-long-exec-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		err := os.MkdirAll(testDir, 0755)
		require.NoError(t, err)

		selector := agent.NewSelector(agent.Config{}, nil)
		engine := New(Config{
			IsMCPMode:           true,
			EnableSummarization: true,
			SummarizationThreshold: SummarizationThreshold{
				FileCount:     40,
				ContextLength: 20000,
				TokenCount:    5000,
			},
			SummarizationModel: LLMModelClaude35,
		}, selector)

		tracker := &fileContextTracker{}

		taskContext := NewTaskContext("/test/project").
			WithChange("change-123").
			WithTask("task-456")

		addedFiles := 0
		for addedFiles < 45 {
			taskContext.AddFile(fmt.Sprintf("/test/project/src/file%d.go", addedFiles))
			addedFiles++
		}

		tracker.track(taskContext.Files)
		tracker.initialSize = tracker.contextSize

		assert.True(t, len(taskContext.Files) > 40, "should exceed file count threshold")

		task := NewTask("Long execution task that grows context").
			WithType("test").
			WithAgent(domain.AgentRoleDeveloper).
			WithModel("haiku").
			WithRuntime(domain.RuntimeCLI).
			WithStrategy(domain.ExecutionStrategySingle)
		task.Context = taskContext

		ctx := context.Background()

		result := engine.TriggerSummarization(ctx, task, "exec-1", LLMModelClaude35, nil, false)
		assert.True(t, result, "summarization should trigger when approaching limit")
		assert.True(t, task.Context.IsSummarized, "context should be marked as summarized")
		assert.Equal(t, 1, task.Context.SummarizationCount, "should have 1 summarization")

		afterSummSize := tracker.calculateContextSize(task.Context.Files)

		assert.Less(t, afterSummSize, tracker.initialSize, "context size should reduce after summarization")
		assert.Equal(t, 1, len(task.Context.Files), "should have single summary file")

		summary := task.Context.Files[0]
		tracker.summaries = append(tracker.summaries, summary)
		assert.Contains(t, summary, "Summary", "file should contain summary")
		assert.Contains(t, summary, "change-123", "summary should mention change ID")
		assert.Contains(t, summary, "task-456", "summary should mention task ID")
	})
}

func TestIntegration_ContextSizeStaysWithinLimits(t *testing.T) {
	t.Run("monitors context size throughout execution", func(t *testing.T) {
		testDir := filepath.Join(os.TempDir(), "go-ent-test-size-monitor-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		err := os.MkdirAll(testDir, 0755)
		require.NoError(t, err)

		selector := agent.NewSelector(agent.Config{}, nil)
		engine := New(Config{
			IsMCPMode:           true,
			EnableSummarization: true,
			SummarizationThreshold: SummarizationThreshold{
				FileCount:     20,
				ContextLength: 5000,
				TokenCount:    1000,
			},
			SummarizationModel: LLMModelGPT4,
		}, selector)

		tracker := &fileContextTracker{}

		taskContext := NewTaskContext("/test/monitored-project").
			WithChange("change-size").
			WithTask("task-size")

		ctx := context.Background()
		task := NewTask("Task monitoring context size").
			WithType("test").
			WithAgent(domain.AgentRoleDeveloper).
			WithModel("haiku").
			WithRuntime(domain.RuntimeCLI).
			WithStrategy(domain.ExecutionStrategySingle)
		task.Context = taskContext

		limit := engine.Config().SummarizationThreshold

		for i := 0; i < 25; i++ {
			content := fmt.Sprintf("File %d content with some data to simulate real file", i)
			taskContext.AddFile(content)
		}
		tracker.track(taskContext.Files)

		assert.Greater(t, len(taskContext.Files), limit.FileCount, "should exceed file count threshold")

		beforeSummFiles := len(taskContext.Files)

		result := engine.TriggerSummarization(ctx, task, "exec-size", LLMModelGPT4, nil, false)
		assert.True(t, result, "summarization should trigger")
		assert.True(t, task.Context.IsSummarized)

		assert.Less(t, len(taskContext.Files), beforeSummFiles, "file count should reduce after summarization")
		assert.LessOrEqual(t, len(taskContext.Files), limit.FileCount, "should be below file count limit")
	})

	t.Run("multiple summarization cycles keep context bounded", func(t *testing.T) {
		testDir := filepath.Join(os.TempDir(), "go-ent-test-multiple-cycles-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		err := os.MkdirAll(testDir, 0755)
		require.NoError(t, err)

		selector := agent.NewSelector(agent.Config{}, nil)
		engine := New(Config{
			IsMCPMode:           true,
			EnableSummarization: true,
			SummarizationThreshold: SummarizationThreshold{
				FileCount:     20,
				ContextLength: 5000,
				TokenCount:    1000,
			},
			SummarizationModel: LLMModelGPT4Turbo,
		}, selector)

		tracker := &fileContextTracker{}

		taskContext := NewTaskContext("/test/cycle-project").
			WithChange("change-cycle").
			WithTask("task-cycle")

		ctx := context.Background()
		task := NewTask("Task testing multiple summarization cycles").
			WithType("test").
			WithAgent(domain.AgentRoleDeveloper).
			WithModel("haiku").
			WithRuntime(domain.RuntimeCLI).
			WithStrategy(domain.ExecutionStrategySingle)
		task.Context = taskContext

		cycleCount := 3
		limit := engine.Config().SummarizationThreshold

		for cycle := 0; cycle < cycleCount; cycle++ {
			t.Run(fmt.Sprintf("cycle_%d", cycle), func(t *testing.T) {
				for i := 0; i < 25; i++ {
					content := fmt.Sprintf("Cycle %d file %d with substantial content to trigger summarization", cycle, i)
					taskContext.AddFile(content)
				}

				tracker.track(taskContext.Files)
				beforeSize := tracker.contextSize

				result := engine.TriggerSummarization(ctx, task, fmt.Sprintf("exec-cycle-%d", cycle), LLMModelGPT4Turbo, nil, false)
				assert.True(t, result, "summarization should trigger in cycle %d", cycle)
				assert.Equal(t, cycle+1, task.Context.SummarizationCount, "should track summarization count")

				tracker.track(taskContext.Files)
				afterSize := tracker.contextSize

				assert.Less(t, afterSize, beforeSize, "size should reduce in cycle %d", cycle)
				assert.LessOrEqual(t, len(taskContext.Files), limit.FileCount, "should stay below file count in cycle %d", cycle)
				assert.LessOrEqual(t, afterSize/4, limit.TokenCount, "should stay below token limit in cycle %d", cycle)
			})
		}

		assert.Equal(t, cycleCount, task.Context.SummarizationCount, "should have exactly %d summarizations", cycleCount)
	})
}

func TestIntegration_CriticalInfoPreserved(t *testing.T) {
	t.Run("critical files preserved in summary", func(t *testing.T) {
		testDir := filepath.Join(os.TempDir(), "go-ent-test-critical-preserve-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		err := os.MkdirAll(testDir, 0755)
		require.NoError(t, err)

		selector := agent.NewSelector(agent.Config{}, nil)
		engine := New(Config{
			IsMCPMode:           true,
			EnableSummarization: true,
			SummarizationThreshold: SummarizationThreshold{
				FileCount:     40,
				ContextLength: 20000,
				TokenCount:    4000,
			},
			SummarizationModel: LLMModelClaude35,
		}, selector)

		taskContext := NewTaskContext("/test/critical-project").
			WithChange("change-critical").
			WithTask("task-critical")

		for i := 0; i < 50; i++ {
			taskContext.AddFile(fmt.Sprintf("/test/critical-project/regular/file%d.go", i))
		}

		ctx := context.Background()
		task := NewTask("Task preserving critical information").
			WithType("test").
			WithAgent(domain.AgentRoleDeveloper).
			WithModel("haiku").
			WithRuntime(domain.RuntimeCLI).
			WithStrategy(domain.ExecutionStrategySingle)
		task.Context = taskContext

		beforeCount := len(taskContext.Files)

		result := engine.TriggerSummarization(ctx, task, "exec-critical", LLMModelClaude35, nil, false)
		assert.True(t, result, "summarization should trigger")
		assert.Equal(t, 1, len(taskContext.Files), "should have single summary file")

		summary := taskContext.Files[0]

		assert.Contains(t, summary, "change-critical", "summary should preserve change ID")
		assert.Contains(t, summary, "task-critical", "summary should preserve task ID")

		assert.Less(t, len(taskContext.Files), beforeCount, "file count should reduce")
		assert.True(t, task.Context.IsSummarized, "context should be marked as summarized")
	})

	t.Run("various critical file types handled", func(t *testing.T) {
		testDir := filepath.Join(os.TempDir(), "go-ent-test-file-types-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		err := os.MkdirAll(testDir, 0755)
		require.NoError(t, err)

		selector := agent.NewSelector(agent.Config{}, nil)
		engine := New(Config{
			IsMCPMode:           true,
			EnableSummarization: true,
			SummarizationThreshold: SummarizationThreshold{
				FileCount:     35,
				ContextLength: 15000,
				TokenCount:    3000,
			},
			SummarizationModel: LLMModelClaude35,
		}, selector)

		taskContext := NewTaskContext("/test/mixed-project").
			WithChange("change-mixed").
			WithTask("task-mixed")

		testCases := []struct {
			fileType string
			content  string
		}{
			{"go", "package main\nfunc main() {}"},
			{"js", "console.log('hello');"},
			{"py", "print('hello')"},
			{"md", "# Documentation\nImportant info"},
			{"yaml", "config:\n  key: value"},
			{"json", "{\"key\": \"value\"}"},
			{"txt", "Plain text file"},
		}

		for i := 0; i < 40; i++ {
			taskContext.AddFile(fmt.Sprintf("/test/mixed-project/regular/file%d.txt", i))
		}

		for _, tc := range testCases {
			taskContext.AddFile(tc.content)
		}

		for i := 0; i < 40; i++ {
			taskContext.AddFile(fmt.Sprintf("/test/mixed-project/regular/file%d.txt", i))
		}

		for _, tc := range testCases {
			taskContext.AddFile(tc.content)
		}

		ctx := context.Background()
		task := NewTask("Task with various file types").
			WithType("test").
			WithAgent(domain.AgentRoleDeveloper).
			WithModel("haiku").
			WithRuntime(domain.RuntimeCLI).
			WithStrategy(domain.ExecutionStrategySingle)
		task.Context = taskContext

		result := engine.TriggerSummarization(ctx, task, "exec-types", LLMModelClaude35, nil, false)
		assert.True(t, result, "summarization should trigger")

		summary := taskContext.Files[0]

		assert.Contains(t, summary, "change-mixed", "summary should preserve change ID")
		assert.Contains(t, summary, "task-mixed", "summary should preserve task ID")
		assert.Contains(t, summary, "Summary", "should be a summary")
		assert.True(t, task.Context.IsSummarized, "context should be marked as summarized")
	})
}

func TestIntegration_MultiLevelSummarization(t *testing.T) {
	t.Run("maintains context chain across multiple summaries", func(t *testing.T) {
		testDir := filepath.Join(os.TempDir(), "go-ent-test-multi-level-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		err := os.MkdirAll(testDir, 0755)
		require.NoError(t, err)

		selector := agent.NewSelector(agent.Config{}, nil)
		engine := New(Config{
			IsMCPMode:           true,
			EnableSummarization: true,
			SummarizationThreshold: SummarizationThreshold{
				FileCount:     25,
				ContextLength: 8000,
				TokenCount:    1600,
			},
			SummarizationModel: LLMModelClaude35,
		}, selector)

		taskContext := NewTaskContext("/test/multi-level-project").
			WithChange("change-multi").
			WithTask("task-multi")

		ctx := context.Background()
		task := NewTask("Task testing multi-level summarization").
			WithType("test").
			WithAgent(domain.AgentRoleDeveloper).
			WithModel("haiku").
			WithRuntime(domain.RuntimeCLI).
			WithStrategy(domain.ExecutionStrategySingle)
		task.Context = taskContext

		summaries := []string{}
		summCount := 3
		tracker := &fileContextTracker{}

		for level := 0; level < summCount; level++ {
			t.Run(fmt.Sprintf("level_%d", level), func(t *testing.T) {
				for i := 0; i < 30; i++ {
					content := fmt.Sprintf("Level %d file %d with content", level, i)
					taskContext.AddFile(content)
				}

				beforeSize := tracker.calculateContextSize(taskContext.Files)

				result := engine.TriggerSummarization(ctx, task, fmt.Sprintf("exec-level-%d", level), LLMModelClaude35, nil, false)
				assert.True(t, result, "summarization should trigger at level %d", level)
				assert.Equal(t, level+1, task.Context.SummarizationCount, "should track summarization level %d", level)

				afterSize := tracker.calculateContextSize(taskContext.Files)
				assert.Less(t, afterSize, beforeSize, "size should reduce at level %d", level)

				if len(taskContext.Files) > 0 {
					summary := taskContext.Files[0]
					summaries = append(summaries, summary)

					expectedFiles := 30
					if level > 0 {
						expectedFiles = 31
					}
					assert.Contains(t, summary, fmt.Sprintf("Summary of %d files", expectedFiles), "should summarize %d files", expectedFiles)
					assert.Contains(t, summary, "change-multi", "should preserve change ID at level %d", level)
					assert.Contains(t, summary, "task-multi", "should preserve task ID at level %d", level)
				}
			})
		}

		assert.Equal(t, summCount, task.Context.SummarizationCount, "should have %d summarizations", summCount)
		assert.Equal(t, summCount, len(summaries), "should have captured %d summaries", summCount)

		for i, summary := range summaries {
			assert.NotEmpty(t, summary, "summary %d should not be empty", i)
			assert.Contains(t, summary, "Summary", "summary %d should be a summary", i)
			assert.Contains(t, summary, "change-multi", "summary %d should preserve change ID", i)
			assert.Contains(t, summary, "task-multi", "summary %d should preserve task ID", i)
		}
	})

	t.Run("summary-to-summary summarization", func(t *testing.T) {
		testDir := filepath.Join(os.TempDir(), "go-ent-test-summary-to-summary-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		err := os.MkdirAll(testDir, 0755)
		require.NoError(t, err)

		selector := agent.NewSelector(agent.Config{}, nil)
		engine := New(Config{
			IsMCPMode:           true,
			EnableSummarization: true,
			SummarizationThreshold: SummarizationThreshold{
				FileCount:     15,
				ContextLength: 4000,
				TokenCount:    800,
			},
			SummarizationModel: LLMModelGPT4,
		}, selector)

		taskContext := NewTaskContext("/test/summary-chain-project").
			WithChange("change-chain").
			WithTask("task-chain")

		ctx := context.Background()
		task := NewTask("Task testing summary-to-summary summarization").
			WithType("test").
			WithAgent(domain.AgentRoleDeveloper).
			WithModel("haiku").
			WithRuntime(domain.RuntimeCLI).
			WithStrategy(domain.ExecutionStrategySingle)
		task.Context = taskContext

		tracker := &fileContextTracker{}

		for i := 0; i < 20; i++ {
			content := fmt.Sprintf("File %d content for summary chain", i)
			taskContext.AddFile(content)
		}

		beforeFirstSize := tracker.calculateContextSize(taskContext.Files)

		result := engine.TriggerSummarization(ctx, task, "exec-first", LLMModelGPT4, nil, false)
		assert.True(t, result, "first summarization should trigger")
		assert.Equal(t, 1, task.Context.SummarizationCount)

		afterFirstSize := tracker.calculateContextSize(taskContext.Files)
		assert.Less(t, afterFirstSize, beforeFirstSize, "size should reduce after first summary")
		assert.True(t, task.Context.IsSummarized, "context should be marked as summarized")

		firstSummary := ""
		if len(taskContext.Files) > 0 {
			firstSummary = taskContext.Files[0]
		}

		for i := 20; i < 40; i++ {
			content := fmt.Sprintf("File %d content for second summary", i)
			taskContext.AddFile(content)
		}

		beforeSecondSize := tracker.calculateContextSize(taskContext.Files)

		result = engine.TriggerSummarization(ctx, task, "exec-second", LLMModelGPT4, nil, false)
		assert.True(t, result, "second summarization should trigger")
		assert.Equal(t, 2, task.Context.SummarizationCount)

		afterSecondSize := tracker.calculateContextSize(taskContext.Files)
		assert.Less(t, afterSecondSize, beforeSecondSize, "size should reduce after second summary")

		secondSummary := ""
		if len(taskContext.Files) > 0 {
			secondSummary = taskContext.Files[0]
		}

		assert.NotEmpty(t, firstSummary, "first summary should not be empty")
		assert.NotEmpty(t, secondSummary, "second summary should not be empty")

		assert.Contains(t, firstSummary, "change-chain", "first summary should preserve change ID")
		assert.Contains(t, firstSummary, "task-chain", "first summary should preserve task ID")

		assert.Contains(t, secondSummary, "change-chain", "second summary should preserve change ID")
		assert.Contains(t, secondSummary, "task-chain", "second summary should preserve task ID")

		assert.Contains(t, firstSummary, "Summary of 20 files", "first summary should mention file count")
		assert.Contains(t, secondSummary, "Summary of", "second summary should be a summary")
	})
}

func TestIntegration_SummarizationEdgeCases(t *testing.T) {
	t.Run("empty context handling", func(t *testing.T) {
		testDir := filepath.Join(os.TempDir(), "go-ent-test-empty-context-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		err := os.MkdirAll(testDir, 0755)
		require.NoError(t, err)

		selector := agent.NewSelector(agent.Config{}, nil)
		engine := New(Config{
			IsMCPMode:              true,
			EnableSummarization:    true,
			SummarizationThreshold: DefaultSummarizationThreshold(),
		}, selector)

		taskContext := NewTaskContext("/test/empty-project").
			WithChange("change-empty").
			WithTask("task-empty")

		ctx := context.Background()
		task := NewTask("Task with empty context").
			WithType("test").
			WithAgent(domain.AgentRoleDeveloper).
			WithModel("haiku").
			WithRuntime(domain.RuntimeCLI).
			WithStrategy(domain.ExecutionStrategySingle)
		task.Context = taskContext

		result := engine.TriggerSummarization(ctx, task, "exec-empty", LLMModelClaude35, nil, false)
		assert.False(t, result, "summarization should not trigger on empty context")
		assert.False(t, task.Context.IsSummarized, "context should not be marked as summarized")
		assert.Equal(t, 0, task.Context.SummarizationCount, "should have zero summarizations")
	})

	t.Run("single file context", func(t *testing.T) {
		testDir := filepath.Join(os.TempDir(), "go-ent-test-single-file-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		err := os.MkdirAll(testDir, 0755)
		require.NoError(t, err)

		selector := agent.NewSelector(agent.Config{}, nil)
		engine := New(Config{
			IsMCPMode:           true,
			EnableSummarization: true,
			SummarizationThreshold: SummarizationThreshold{
				FileCount:     50,
				ContextLength: 10000,
				TokenCount:    2000,
			},
		}, selector)

		taskContext := NewTaskContext("/test/single-file-project").
			WithChange("change-single").
			WithTask("task-single")
		taskContext.AddFile("/test/single-file-project/main.go")

		ctx := context.Background()
		task := NewTask("Task with single file context").
			WithType("test").
			WithAgent(domain.AgentRoleDeveloper).
			WithModel("haiku").
			WithRuntime(domain.RuntimeCLI).
			WithStrategy(domain.ExecutionStrategySingle)
		task.Context = taskContext

		result := engine.TriggerSummarization(ctx, task, "exec-single", LLMModelClaude35, nil, false)
		assert.False(t, result, "summarization should not trigger on single file below threshold")
		assert.False(t, task.Context.IsSummarized, "context should not be marked as summarized")
		assert.Equal(t, 1, len(taskContext.Files), "should still have 1 file")
	})

	t.Run("very large context with multiple thresholds", func(t *testing.T) {
		testDir := filepath.Join(os.TempDir(), "go-ent-test-large-context-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		err := os.MkdirAll(testDir, 0755)
		require.NoError(t, err)

		selector := agent.NewSelector(agent.Config{}, nil)
		engine := New(Config{
			IsMCPMode:           true,
			EnableSummarization: true,
			SummarizationThreshold: SummarizationThreshold{
				FileCount:     60,
				ContextLength: 60000,
				TokenCount:    12000,
			},
			SummarizationModel: LLMModelClaude4,
		}, selector)

		taskContext := NewTaskContext("/test/very-large-project").
			WithChange("change-large").
			WithTask("task-large")

		for i := 0; i < 100; i++ {
			content := strings.Repeat("x", 700)
			taskContext.AddFile(content)
		}

		ctx := context.Background()
		task := NewTask("Task with very large context").
			WithType("test").
			WithAgent(domain.AgentRoleDeveloper).
			WithModel("haiku").
			WithRuntime(domain.RuntimeCLI).
			WithStrategy(domain.ExecutionStrategySingle)
		task.Context = taskContext

		tracker := &fileContextTracker{}
		tracker.track(taskContext.Files)

		assert.Greater(t, len(taskContext.Files), engine.Config().SummarizationThreshold.FileCount, "should exceed file count")
		assert.Greater(t, tracker.contextSize, engine.Config().SummarizationThreshold.ContextLength, "should exceed context length")
		assert.Greater(t, tracker.countTokens(), engine.Config().SummarizationThreshold.TokenCount, "should exceed token count")

		result := engine.TriggerSummarization(ctx, task, "exec-large", LLMModelClaude4, nil, false)
		assert.True(t, result, "summarization should trigger on very large context")
		assert.True(t, task.Context.IsSummarized)
		assert.Equal(t, 1, task.Context.SummarizationCount)

		tracker.track(taskContext.Files)

		assert.Less(t, len(taskContext.Files), engine.Config().SummarizationThreshold.FileCount, "should be below file count after summarization")
		assert.Less(t, tracker.contextSize, engine.Config().SummarizationThreshold.ContextLength, "should be below context length after summarization")
	})
}

func TestIntegration_SummarizationWithDifferentModels(t *testing.T) {
	t.Run("test with various LLM models", func(t *testing.T) {
		testDir := filepath.Join(os.TempDir(), "go-ent-test-models-"+strings.ReplaceAll(t.Name(), "/", "-"))
		defer os.RemoveAll(testDir)

		err := os.MkdirAll(testDir, 0755)
		require.NoError(t, err)

		models := []string{
			LLMModelClaude35,
			LLMModelClaude4,
			LLMModelGPT4,
			LLMModelGPT4Turbo,
			LLMModelGPT35Turbo,
		}

		for _, model := range models {
			t.Run(model, func(t *testing.T) {
				selector := agent.NewSelector(agent.Config{}, nil)
				engine := New(Config{
					IsMCPMode:           true,
					EnableSummarization: true,
					SummarizationThreshold: SummarizationThreshold{
						FileCount:     30,
						ContextLength: 8000,
						TokenCount:    1600,
					},
					SummarizationModel: model,
				}, selector)

				taskContext := NewTaskContext("/test/model-project").
					WithChange(fmt.Sprintf("change-%s", model)).
					WithTask(fmt.Sprintf("task-%s", model))

				for i := 0; i < 35; i++ {
					filename := fmt.Sprintf("/test/model-project/%s/file%d.go", model, i)
					taskContext.AddFile(filename)
				}

				ctx := context.Background()
				task := NewTask(fmt.Sprintf("Task with %s model", model)).
					WithType("test").
					WithAgent(domain.AgentRoleDeveloper).
					WithModel("haiku").
					WithRuntime(domain.RuntimeCLI).
					WithStrategy(domain.ExecutionStrategySingle)
				task.Context = taskContext

				result := engine.TriggerSummarization(ctx, task, fmt.Sprintf("exec-%s", model), model, nil, false)
				assert.True(t, result, "summarization should trigger with %s", model)
				assert.True(t, task.Context.IsSummarized)
				assert.Equal(t, 1, task.Context.SummarizationCount)

				if len(taskContext.Files) > 0 {
					summary := taskContext.Files[0]
					assert.Contains(t, summary, fmt.Sprintf("change-%s", model), "summary should preserve change ID")
					assert.Contains(t, summary, fmt.Sprintf("task-%s", model), "summary should preserve task ID")
				}
			})
		}
	})
}
