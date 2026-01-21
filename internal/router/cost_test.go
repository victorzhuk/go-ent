package router

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/victorzhuk/go-ent/internal/config"
	"github.com/victorzhuk/go-ent/internal/execution"
	"github.com/victorzhuk/go-ent/internal/worker"
)

func TestRouter_SetCostBudget(t *testing.T) {
	t.Run("sets budget and remaining", func(t *testing.T) {
		config := &worker.Config{
			Providers: map[string]worker.ProviderDefinition{
				"test": {
					Provider: "anthropic",
					Model:    "claude-3-haiku",
					Method:   config.MethodAPI,
				},
			},
		}
		r, err := NewRouter(config, nil)
		require.NoError(t, err)

		r.SetCostBudget(10.0)
		assert.Equal(t, 10.0, r.costBudget)
		assert.Equal(t, 10.0, r.remainingBudget)
	})
}

func TestRouter_RecordCost(t *testing.T) {
	t.Run("deducts from remaining budget", func(t *testing.T) {
		config := &worker.Config{
			Providers: map[string]worker.ProviderDefinition{
				"test": {
					Provider: "anthropic",
					Model:    "claude-3-haiku",
					Method:   config.MethodAPI,
				},
			},
		}
		r, err := NewRouter(config, nil)
		require.NoError(t, err)

		r.SetCostBudget(10.0)

		r.RecordCost("provider1", 2.5)
		assert.Equal(t, 7.5, r.remainingBudget)
		assert.Equal(t, 2.5, r.costsByProvider["provider1"])

		r.RecordCost("provider1", 1.0)
		assert.Equal(t, 6.5, r.remainingBudget)
		assert.Equal(t, 3.5, r.costsByProvider["provider1"])
	})

	t.Run("tracks costs per provider", func(t *testing.T) {
		config := &worker.Config{
			Providers: map[string]worker.ProviderDefinition{
				"test": {
					Provider: "anthropic",
					Model:    "claude-3-haiku",
					Method:   config.MethodAPI,
				},
			},
		}
		r, err := NewRouter(config, nil)
		require.NoError(t, err)

		r.SetCostBudget(20.0)

		r.RecordCost("provider1", 2.5)
		r.RecordCost("provider2", 3.0)
		r.RecordCost("provider1", 1.0)

		costs := r.GetCostsByProvider()
		assert.Equal(t, 3.5, costs["provider1"])
		assert.Equal(t, 3.0, costs["provider2"])
		assert.Equal(t, 13.5, r.remainingBudget)
	})
}

func TestRouter_GetRemainingBudget(t *testing.T) {
	t.Run("returns correct remaining budget", func(t *testing.T) {
		config := &worker.Config{
			Providers: map[string]worker.ProviderDefinition{
				"test": {
					Provider: "anthropic",
					Model:    "claude-3-haiku",
					Method:   config.MethodAPI,
				},
			},
		}
		r, err := NewRouter(config, nil)
		require.NoError(t, err)

		r.SetCostBudget(10.0)
		assert.Equal(t, 10.0, r.GetRemainingBudget())

		r.RecordCost("provider1", 3.0)
		assert.Equal(t, 7.0, r.GetRemainingBudget())
	})
}

func TestRouter_ResetBudget(t *testing.T) {
	t.Run("resets to initial budget", func(t *testing.T) {
		config := &worker.Config{
			Providers: map[string]worker.ProviderDefinition{
				"test": {
					Provider: "anthropic",
					Model:    "claude-3-haiku",
					Method:   config.MethodAPI,
				},
			},
		}
		r, err := NewRouter(config, nil)
		require.NoError(t, err)

		r.SetCostBudget(10.0)
		r.RecordCost("provider1", 3.0)
		r.RecordCost("provider2", 2.0)

		r.ResetBudget()

		assert.Equal(t, 10.0, r.remainingBudget)
		assert.Equal(t, 0.0, r.costsByProvider["provider1"])
		assert.Equal(t, 0.0, r.costsByProvider["provider2"])
		assert.Equal(t, 0, len(r.costsByProvider))
	})
}

func TestRouter_Route_BudgetConstraints(t *testing.T) {
	config := &worker.Config{
		Providers: map[string]worker.ProviderDefinition{
			"expensive": {
				Provider: "anthropic",
				Model:    "gpt-4",
				Method:   config.MethodAPI,
			},
			"cheap": {
				Provider: "moonshot",
				Model:    "gpt-3.5",
				Method:   config.MethodCLI,
			},
		},
	}

	t.Run("allows routing within budget", func(t *testing.T) {
		r, err := NewRouter(config, nil)
		require.NoError(t, err)

		r.SetDefaults(DefaultRoutes{SimpleTasks: "expensive"})
		r.SetCostBudget(10.0)

		task := &execution.Task{
			Type:        "test",
			Description: "simple task",
			Context:     &execution.TaskContext{},
		}

		decision, err := r.Route(context.Background(), task)
		require.NoError(t, err)
		assert.Equal(t, "expensive", decision.Provider)
		assert.Equal(t, "default", decision.RuleName)
	})

	t.Run("falls back to cheaper provider when budget exceeded", func(t *testing.T) {
		r, err := NewRouter(config, nil)
		require.NoError(t, err)

		r.SetDefaults(DefaultRoutes{SimpleTasks: "expensive"})
		r.SetCostBudget(0.005)

		task := &execution.Task{
			Type:        "test",
			Description: "simple task",
			Context:     &execution.TaskContext{},
		}

		decision, err := r.Route(context.Background(), task)
		require.NoError(t, err)
		assert.Equal(t, "cheap", decision.Provider)
		assert.Contains(t, decision.Reason, "budget fallback")
	})

	t.Run("skips budget check when budget is zero", func(t *testing.T) {
		r, err := NewRouter(config, nil)
		require.NoError(t, err)

		r.SetDefaults(DefaultRoutes{SimpleTasks: "expensive"})
		r.SetCostBudget(0)

		task := &execution.Task{
			Type:        "test",
			Description: "simple task",
			Context:     &execution.TaskContext{},
		}

		decision, err := r.Route(context.Background(), task)
		require.NoError(t, err)
		assert.Equal(t, "expensive", decision.Provider)
	})

	t.Run("errors when no cheaper provider available", func(t *testing.T) {
		r, err := NewRouter(config, nil)
		require.NoError(t, err)

		r.SetDefaults(DefaultRoutes{SimpleTasks: "cheap"})
		r.SetCostBudget(0.001)

		task := &execution.Task{
			Type:        "test",
			Description: "simple task",
			Context:     &execution.TaskContext{},
		}

		_, err = r.Route(context.Background(), task)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient budget")
	})
}

func TestRouter_findCheaperProvider(t *testing.T) {
	config := &worker.Config{
		Providers: map[string]worker.ProviderDefinition{
			"expensive": {
				Provider: "anthropic",
				Model:    "gpt-4",
				Method:   config.MethodAPI,
			},
			"medium": {
				Provider: "moonshot",
				Model:    "gpt-3.5",
				Method:   config.MethodAPI,
			},
			"cheap": {
				Provider: "deepseek",
				Model:    "claude-haiku",
				Method:   config.MethodCLI,
			},
		},
	}

	t.Run("finds cheapest provider within budget", func(t *testing.T) {
		r, err := NewRouter(config, nil)
		require.NoError(t, err)

		r.SetCostBudget(1.0)

		decision := &RoutingDecision{
			Provider:      "expensive",
			EstimatedCost: 0.10,
		}

		fallback, err := r.findCheaperProvider(decision)
		require.NoError(t, err)
		assert.Equal(t, "cheap", fallback.Provider)
		assert.Equal(t, "expensive", fallback.OriginalProvider)
		assert.Less(t, fallback.EstimatedCost, decision.EstimatedCost)
	})

	t.Run("errors when no cheaper provider available", func(t *testing.T) {
		r, err := NewRouter(config, nil)
		require.NoError(t, err)

		decision := &RoutingDecision{
			Provider:      "cheap",
			EstimatedCost: 0.01,
		}

		_, err = r.findCheaperProvider(decision)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no cheaper provider available")
	})
}
