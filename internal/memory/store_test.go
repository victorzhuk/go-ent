package memory

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMemoryStore(t *testing.T) {
	t.Run("creates empty memory store", func(t *testing.T) {
		mem := NewMemoryStore()
		assert.NotNil(t, mem)
		assert.Equal(t, 0, mem.GetTotalPatterns())
	})
}

func TestStore(t *testing.T) {
	t.Run("stores pattern successfully", func(t *testing.T) {
		mem := NewMemoryStore()

		pattern := &Pattern{
			TaskType:    "implement",
			Provider:    "moonshot",
			Model:       "glm-4",
			Method:      "acp",
			FileCount:   5,
			ContextSize: 30000,
			Success:     true,
			Cost:        0.02,
			Duration:    5 * time.Second,
			OutputSize:  1000,
		}

		err := mem.Store(pattern)
		require.NoError(t, err)
		assert.Equal(t, 1, mem.GetTotalPatterns())
	})

	t.Run("updates stats on subsequent stores", func(t *testing.T) {
		mem := NewMemoryStore()

		for i := 0; i < 10; i++ {
			pattern := &Pattern{
				TaskType:   "implement",
				Provider:   "moonshot",
				Model:      "glm-4",
				Method:     "acp",
				Success:    i%3 != 0,
				Cost:       0.02,
				Duration:   5 * time.Second,
				OutputSize: 1000,
			}

			err := mem.Store(pattern)
			require.NoError(t, err)
		}

		stats, err := mem.GetProviderStats("moonshot", "glm-4", "acp")
		require.NoError(t, err)
		assert.Equal(t, 10, stats.TotalExecutions)
		assert.Equal(t, 4, stats.SuccessCount)
		assert.Equal(t, 6, stats.FailureCount)
		assert.InDelta(t, 40.0, stats.SuccessRate, 0.1)
		assert.InDelta(t, 0.02, stats.AverageCost, 0.01)
	})
}

func TestQuery(t *testing.T) {
	t.Run("returns nil for empty store", func(t *testing.T) {
		mem := NewMemoryStore()

		recommendation := mem.Query("implement", 5, 30000)
		assert.Nil(t, recommendation)
	})

	t.Run("returns nil for insufficient patterns", func(t *testing.T) {
		mem := NewMemoryStore()

		err := mem.Store(&Pattern{
			TaskType: "implement",
			Provider: "moonshot",
			Model:    "glm-4",
			Success:  true,
			Cost:     0.02,
		})
		require.NoError(t, err)

		recommendation := mem.Query("implement", 5, 30000)
		assert.Nil(t, recommendation)
	})

	t.Run("returns recommendation for matching pattern", func(t *testing.T) {
		mem := NewMemoryStore()

		for i := 0; i < 10; i++ {
			err := mem.Store(&Pattern{
				TaskType:    "implement",
				Provider:    "moonshot",
				Model:       "glm-4",
				Method:      "acp",
				FileCount:   5,
				ContextSize: 30000,
				Success:     true,
				Cost:        0.02,
				Duration:    5 * time.Second,
				OutputSize:  1000,
			})
			require.NoError(t, err)
		}

		recommendation := mem.Query("implement", 5, 30000)
		assert.NotNil(t, recommendation)
		assert.Equal(t, "moonshot", recommendation.Provider)
		assert.Equal(t, "glm-4", recommendation.Model)
		assert.Equal(t, "acp", recommendation.Method)
		assert.Contains(t, recommendation.Reason, "learned")
		assert.Greater(t, recommendation.Confidence, 0.0)
		assert.GreaterOrEqual(t, recommendation.Confidence, 1.0)
	})

	t.Run("ignores patterns with low success rate", func(t *testing.T) {
		mem := NewMemoryStore()

		for i := 0; i < 10; i++ {
			err := mem.Store(&Pattern{
				TaskType: "implement",
				Provider: "moonshot",
				Model:    "glm-4",
				Success:  i < 5,
				Cost:     0.02,
			})
			require.NoError(t, err)
		}

		recommendation := mem.Query("implement", 5, 30000)
		assert.Nil(t, recommendation)
	})
}

func TestGetProviderStats(t *testing.T) {
	t.Run("returns stats for existing provider", func(t *testing.T) {
		mem := NewMemoryStore()

		for i := 0; i < 5; i++ {
			err := mem.Store(&Pattern{
				Provider: "moonshot",
				Model:    "glm-4",
				Method:   "acp",
				Success:  i%2 == 0,
				Cost:     0.02,
			})
			require.NoError(t, err)
		}

		stats, err := mem.GetProviderStats("moonshot", "glm-4", "acp")
		require.NoError(t, err)
		assert.Equal(t, 5, stats.TotalExecutions)
		assert.Equal(t, 3, stats.SuccessCount)
		assert.Equal(t, 2, stats.FailureCount)
		assert.InDelta(t, 60.0, stats.SuccessRate, 0.1)
		assert.InDelta(t, 0.02, stats.AverageCost, 0.01)
	})

	t.Run("returns error for non-existent provider", func(t *testing.T) {
		mem := NewMemoryStore()

		stats, err := mem.GetProviderStats("nonexistent", "model", "method")
		assert.Error(t, err)
		assert.Nil(t, stats)
	})
}

func TestGetBestProviderForTask(t *testing.T) {
	t.Run("returns best provider for task type", func(t *testing.T) {
		mem := NewMemoryStore()

		err := mem.Store(&Pattern{
			TaskType: "implement",
			Provider: "moonshot",
			Model:    "glm-4",
			Success:  true,
			Cost:     0.01,
		})
		require.NoError(t, err)

		for i := 0; i < 10; i++ {
			err := mem.Store(&Pattern{
				TaskType: "implement",
				Provider: "moonshot",
				Model:    "glm-4",
				Success:  true,
				Cost:     0.02,
			})
			require.NoError(t, err)
		}

		recommendation, err := mem.GetBestProviderForTask("implement", 0.1)
		require.NoError(t, err)
		assert.Equal(t, "moonshot", recommendation.Provider)
		assert.Equal(t, "glm-4", recommendation.Model)
		assert.InDelta(t, 0.015, recommendation.EstimatedCost, 0.01)
	})

	t.Run("respects cost constraint", func(t *testing.T) {
		mem := NewMemoryStore()

		for i := 0; i < 10; i++ {
			err := mem.Store(&Pattern{
				TaskType: "implement",
				Provider: "moonshot",
				Model:    "glm-4",
				Success:  true,
				Cost:     0.05,
			})
			require.NoError(t, err)
		}

		_, err := mem.GetBestProviderForTask("implement", 0.01)
		assert.Error(t, err)
	})

	t.Run("returns error for no patterns", func(t *testing.T) {
		mem := NewMemoryStore()

		_, err := mem.GetBestProviderForTask("implement", 0.1)
		assert.Error(t, err)
	})
}

func TestGetAllProviders(t *testing.T) {
	t.Run("returns list of unique providers", func(t *testing.T) {
		mem := NewMemoryStore()

		providers := []string{"moonshot", "deepseek", "anthropic"}
		models := []string{"glm-4", "deepseek-coder", "haiku"}
		methods := []string{"acp", "acp", "api"}

		for i := 0; i < len(providers); i++ {
			err := mem.Store(&Pattern{
				Provider: providers[i],
				Model:    models[i],
				Method:   methods[i],
				Success:  true,
				Cost:     0.02,
			})
			require.NoError(t, err)
		}

		result := mem.GetAllProviders()
		assert.ElementsMatch(t, providers, result)
	})

	t.Run("returns empty list for no patterns", func(t *testing.T) {
		mem := NewMemoryStore()

		result := mem.GetAllProviders()
		assert.Empty(t, result)
	})
}

func TestGetTotalPatterns(t *testing.T) {
	t.Run("counts stored patterns", func(t *testing.T) {
		mem := NewMemoryStore()

		for i := 0; i < 10; i++ {
			err := mem.Store(&Pattern{
				Provider: "moonshot",
				Success:  true,
				Cost:     0.02,
			})
			require.NoError(t, err)
		}

		assert.Equal(t, 10, mem.GetTotalPatterns())
	})

	t.Run("returns zero for empty store", func(t *testing.T) {
		mem := NewMemoryStore()
		assert.Equal(t, 0, mem.GetTotalPatterns())
	})
}
