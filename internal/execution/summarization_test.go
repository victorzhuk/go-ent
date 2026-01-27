package execution

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultSummarizationThreshold(t *testing.T) {
	t.Run("returns default values", func(t *testing.T) {
		threshold := DefaultSummarizationThreshold()

		assert.Equal(t, 50, threshold.FileCount)
		assert.Equal(t, 50000, threshold.ContextLength)
		assert.Equal(t, 10000, threshold.TokenCount)
	})
}

func TestTaskContext_Summarization(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		ctx := NewTaskContext("/test")

		assert.False(t, ctx.IsSummarized)
		assert.Equal(t, 0, ctx.SummarizationCount)
		assert.Nil(t, ctx.OriginalContext)
	})

	t.Run("mark as summarized", func(t *testing.T) {
		ctx := NewTaskContext("/test")
		ctx.IsSummarized = true
		ctx.SummarizationCount = 1

		assert.True(t, ctx.IsSummarized)
		assert.Equal(t, 1, ctx.SummarizationCount)
	})
}

func TestEngine_TriggerSummarization(t *testing.T) {
	t.Run("disabled summarization", func(t *testing.T) {
		cfg := Config{
			EnableSummarization: false,
		}
		engine := New(cfg, nil)

		task := &Task{
			Context: NewTaskContext("/test").WithFiles(make([]string, 100)),
		}

		result := engine.TriggerSummarization(nil, task, "test-id", "claude-3.5-sonnet", nil, false)

		assert.False(t, result)
		assert.False(t, task.Context.IsSummarized)
	})

	t.Run("force trigger", func(t *testing.T) {
		cfg := Config{
			EnableSummarization: true,
		}
		engine := New(cfg, nil)

		task := &Task{
			Context: NewTaskContext("/test").WithFiles(make([]string, 10)),
		}

		result := engine.TriggerSummarization(nil, task, "test-id", "claude-3.5-sonnet", nil, true)

		assert.True(t, result)
		assert.True(t, task.Context.IsSummarized)
		assert.Equal(t, 1, task.Context.SummarizationCount)
		assert.Equal(t, 1, len(task.Context.Files))
	})

	t.Run("automatic trigger on file count", func(t *testing.T) {
		cfg := Config{
			EnableSummarization: true,
			SummarizationThreshold: SummarizationThreshold{
				FileCount: 50,
			},
		}
		engine := New(cfg, nil)

		task := &Task{
			Context: NewTaskContext("/test").WithFiles(make([]string, 51)),
		}

		result := engine.TriggerSummarization(nil, task, "test-id", "claude-3.5-sonnet", nil, false)

		assert.True(t, result)
		assert.True(t, task.Context.IsSummarized)
	})

	t.Run("automatic trigger on context length", func(t *testing.T) {
		cfg := Config{
			EnableSummarization: true,
			SummarizationThreshold: SummarizationThreshold{
				ContextLength: 1000,
			},
		}
		engine := New(cfg, nil)

		files := make([]string, 10)
		for i := range files {
			files[i] = "x"
		}

		task := &Task{
			Context: NewTaskContext("/test").WithFiles(files),
		}

		result := engine.TriggerSummarization(nil, task, "test-id", "claude-3.5-sonnet", nil, false)

		assert.True(t, result)
		assert.True(t, task.Context.IsSummarized)
	})

	t.Run("no trigger when below thresholds", func(t *testing.T) {
		cfg := Config{
			EnableSummarization: true,
			SummarizationThreshold: SummarizationThreshold{
				FileCount:     50,
				ContextLength: 50000,
				TokenCount:    10000,
			},
		}
		engine := New(cfg, nil)

		task := &Task{
			Context: NewTaskContext("/test").WithFiles(make([]string, 10)),
		}

		result := engine.TriggerSummarization(nil, task, "test-id", "claude-3.5-sonnet", nil, false)

		assert.False(t, result)
		assert.False(t, task.Context.IsSummarized)
	})

	t.Run("nil context", func(t *testing.T) {
		cfg := Config{
			EnableSummarization: true,
		}
		engine := New(cfg, nil)

		task := &Task{
			Context: nil,
		}

		result := engine.TriggerSummarization(nil, task, "test-id", "claude-3.5-sonnet", nil, true)

		assert.False(t, result)
	})
}

func TestEngine_shouldSummarize(t *testing.T) {
	t.Run("file count exceeds threshold", func(t *testing.T) {
		threshold := SummarizationThreshold{FileCount: 50}
		ctx := NewTaskContext("/test").WithFiles(make([]string, 51))

		engine := &Engine{}
		result := engine.shouldSummarize(ctx, threshold)

		assert.True(t, result)
	})

	t.Run("context length exceeds threshold", func(t *testing.T) {
		threshold := SummarizationThreshold{ContextLength: 1000}
		ctx := NewTaskContext("/test")

		files := make([]string, 10)
		for i := range files {
			files[i] = string(make([]byte, 100))
		}

		ctx.WithFiles(files)

		engine := &Engine{}
		result := engine.shouldSummarize(ctx, threshold)

		assert.True(t, result)
	})

	t.Run("token count exceeds threshold", func(t *testing.T) {
		threshold := SummarizationThreshold{TokenCount: 500}
		ctx := NewTaskContext("/test")

		files := make([]string, 10)
		for i := range files {
			files[i] = string(make([]byte, 250))
		}

		ctx.WithFiles(files)

		engine := &Engine{}
		result := engine.shouldSummarize(ctx, threshold)

		assert.True(t, result)
	})

	t.Run("below all thresholds", func(t *testing.T) {
		threshold := SummarizationThreshold{
			FileCount:     50,
			ContextLength: 50000,
			TokenCount:    10000,
		}
		ctx := NewTaskContext("/test").WithFiles(make([]string, 10))

		engine := &Engine{}
		result := engine.shouldSummarize(ctx, threshold)

		assert.False(t, result)
	})
}

func TestEngine_Config(t *testing.T) {
	t.Run("returns configuration", func(t *testing.T) {
		cfg := Config{
			EnableSummarization: true,
			SummarizationThreshold: SummarizationThreshold{
				FileCount: 100,
			},
			SummarizationModel: "gpt-4",
		}
		engine := New(cfg, nil)

		config := engine.Config()

		assert.True(t, config.EnableSummarization)
		assert.Equal(t, 100, config.SummarizationThreshold.FileCount)
		assert.Equal(t, "gpt-4", config.SummarizationModel)
	})
}

func TestLoadSummarizationThreshold(t *testing.T) {
	t.Run("returns default when file not found", func(t *testing.T) {
		threshold, err := LoadSummarizationThreshold("/nonexistent/path")

		assert.NoError(t, err)
		assert.Equal(t, 50, threshold.FileCount)
		assert.Equal(t, 50000, threshold.ContextLength)
		assert.Equal(t, 10000, threshold.TokenCount)
	})

	t.Run("validates threshold values", func(t *testing.T) {
		threshold := DefaultSummarizationThreshold()

		assert.Greater(t, threshold.FileCount, 0)
		assert.Greater(t, threshold.ContextLength, 0)
		assert.Greater(t, threshold.TokenCount, 0)
	})
}

func TestLLMModelConstants(t *testing.T) {
	t.Run("model constants defined", func(t *testing.T) {
		assert.NotEmpty(t, LLMModelClaude35)
		assert.NotEmpty(t, LLMModelClaude4)
		assert.NotEmpty(t, LLMModelGPT4)
		assert.NotEmpty(t, LLMModelGPT4Turbo)
		assert.NotEmpty(t, LLMModelGPT35Turbo)
	})
}
