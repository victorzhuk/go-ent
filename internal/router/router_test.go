package router

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
	"github.com/victorzhuk/go-ent/internal/config"
	"github.com/victorzhuk/go-ent/internal/execution"
	"github.com/victorzhuk/go-ent/internal/memory"
	"github.com/victorzhuk/go-ent/internal/worker"
)

func TestNewRouter(t *testing.T) {
	t.Parallel()

	cfg := worker.DefaultConfig()
	cfg.Providers = map[string]worker.ProviderDefinition{
		"glm": {
			Method:   config.MethodACP,
			Provider: "moonshot",
			Model:    "glm-4",
		},
		"haiku": {
			Method:   config.MethodAPI,
			Provider: "anthropic",
			Model:    "claude-3-haiku-3-5",
		},
	}

	router, err := NewRouter(cfg, nil)
	require.NoError(t, err)
	assert.NotNil(t, router)
	assert.NotNil(t, router.providers)
}

func TestNewRouter_InvalidConfig(t *testing.T) {
	t.Parallel()

	router, err := NewRouter(nil, nil)
	assert.Error(t, err)
	assert.Nil(t, router)
}

func TestRouter_SimpleTask(t *testing.T) {
	t.Parallel()

	router := createTestRouter()

	task := execution.NewTask("Fix a simple typo")
	task = task.WithType("bugfix").WithContext(
		execution.NewTaskContext("/test").WithFiles([]string{"main.go"}),
	)

	decision, err := router.Route(context.Background(), task)
	require.NoError(t, err)
	assert.Equal(t, config.MethodAPI, decision.Method)
	assert.NotEmpty(t, decision.Reason)
}

func TestRouter_ComplexTask(t *testing.T) {
	t.Parallel()

	router := createTestRouter()

	task := execution.NewTask("Refactor entire module architecture")
	task = task.WithType("refactor").WithContext(
		execution.NewTaskContext("/test").WithFiles([]string{"file1.go", "file2.go", "file3.go", "file4.go", "file5.go", "file6.go"}),
	)

	decision, err := router.Route(context.Background(), task)
	require.NoError(t, err)
	assert.Equal(t, config.MethodACP, decision.Method)
	assert.Contains(t, strings.ToLower(decision.Reason), "complex")
}

func TestRouter_LargeContext(t *testing.T) {
	t.Parallel()

	router := createTestRouter()

	task := execution.NewTask("Analyze large config file with many files")

	files := make([]string, 30)
	for i := 0; i < 30; i++ {
		files[i] = fmt.Sprintf("file%d.go", i)
	}

	task = task.WithType("analysis").WithContext(
		execution.NewTaskContext("/test").WithFiles(files),
	)

	decision, err := router.Route(context.Background(), task)
	require.NoError(t, err)
	assert.Equal(t, config.MethodACP, decision.Method)
	assert.Contains(t, strings.ToLower(decision.Reason), "context")
}

func TestRouter_BulkImplementation(t *testing.T) {
	t.Parallel()

	router := createTestRouter()

	task := execution.NewTask("Implement new feature across multiple files")
	task = task.WithType("feature").WithContext(
		execution.NewTaskContext("/test").WithFiles([]string{
			"file1.go", "file2.go", "file3.go", "file4.go",
		}),
	)

	decision, err := router.Route(context.Background(), task)
	require.NoError(t, err)
	assert.Equal(t, config.MethodACP, decision.Method)
	assert.Contains(t, strings.ToLower(decision.Reason), "bulk")
}

func TestRouter_RulePriority(t *testing.T) {
	t.Parallel()

	router := createTestRouter()

	rules := []RoutingRule{
		{
			ID:       "high-priority-rule",
			Priority: 100,
			Match: MatchConditions{
				Type: []string{"feature"},
			},
			Action: RouteAction{
				Method:   "api",
				Provider: "haiku",
				Model:    "claude-3-haiku-3-5",
			},
		},
		{
			ID:       "low-priority-rule",
			Priority: 1,
			Match: MatchConditions{
				Type: []string{"feature"},
			},
			Action: RouteAction{
				Method:   "cli",
				Provider: "glm",
			},
		},
	}

	router.SetRules(rules)

	task := execution.NewTask("Implement feature")
	task = task.WithType("feature")

	decision, err := router.Route(context.Background(), task)
	require.NoError(t, err)
	assert.Equal(t, config.MethodAPI, decision.Method)
	assert.Equal(t, "haiku", decision.Provider)
}

func TestRouter_FileCountMatching(t *testing.T) {
	t.Parallel()

	router := createTestRouter()

	rules := []RoutingRule{
		{
			ID:       "single-file-rule",
			Priority: 10,
			Match: MatchConditions{
				FileCount: intPtr(1),
			},
			Action: RouteAction{
				Method:   "cli",
				Provider: "haiku",
			},
		},
		{
			ID:       "multi-file-rule",
			Priority: 20,
			Match: MatchConditions{
				FileCount: intPtr(2),
			},
			Action: RouteAction{
				Method:   "acp",
				Provider: "glm",
			},
		},
	}

	router.SetRules(rules)

	tests := []struct {
		name           string
		files          []string
		expectedMethod config.CommunicationMethod
	}{
		{"single file", []string{"main.go"}, config.MethodCLI},
		{"multiple files", []string{"main.go", "util.go"}, config.MethodACP},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			task := execution.NewTask("Test task")
			task = task.WithContext(execution.NewTaskContext("/test").WithFiles(tt.files))

			decision, err := router.Route(context.Background(), task)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedMethod, decision.Method)
		})
	}
}

func TestRouter_ComplexityMatching(t *testing.T) {
	t.Parallel()

	router := createTestRouter()

	rules := []RoutingRule{
		{
			ID:       "simple-rule",
			Priority: 10,
			Match: MatchConditions{
				Complexity: "simple",
			},
			Action: RouteAction{
				Method:   "cli",
				Provider: "haiku",
			},
		},
		{
			ID:       "complex-rule",
			Priority: 10,
			Match: MatchConditions{
				Complexity: "complex",
			},
			Action: RouteAction{
				Method:   "acp",
				Provider: "glm",
			},
		},
	}

	router.SetRules(rules)

	tests := []struct {
		name           string
		description    string
		taskType       string
		expectedMethod config.CommunicationMethod
	}{
		{"simple bugfix", "Fix typo", "bugfix", config.MethodCLI},
		{"complex refactor", "Refactor entire architecture", "refactor", config.MethodACP},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			task := execution.NewTask(tt.description)
			task = task.WithType(tt.taskType)

			decision, err := router.Route(context.Background(), task)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedMethod, decision.Method)
		})
	}
}

func TestRouter_NilTask(t *testing.T) {
	t.Parallel()

	router := createTestRouter()

	decision, err := router.Route(context.Background(), nil)
	assert.Error(t, err)
	assert.Nil(t, decision)
	assert.ErrorIs(t, err, ErrInvalidTask)
}

func TestRouter_LoadRoutingConfig(t *testing.T) {
	t.Parallel()

	cfg := worker.DefaultConfig()
	cfg.Providers = map[string]worker.ProviderDefinition{
		"glm": {
			Method:   config.MethodACP,
			Provider: "moonshot",
			Model:    "glm-4",
		},
		"haiku": {
			Method:   config.MethodAPI,
			Provider: "anthropic",
			Model:    "claude-3-haiku-3-5",
		},
	}

	router, err := NewRouter(cfg, nil)
	require.NoError(t, err)

	router.SetRules(DefaultRoutingRules())

	assert.Equal(t, 7, len(router.rules))
}

func TestRouter_WithRoutingConfig(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "routing.yaml")

	configContent := `
rules:
  - id: "simple-cli"
    priority: 1000
    match:
      complexity: "trivial"
    action:
      method: "cli"
      provider: "anthropic"
      model: "claude-3-haiku-3-5"
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	cfg := worker.DefaultConfig()
	cfg.Providers = map[string]worker.ProviderDefinition{
		"glm": {
			Method:   config.MethodACP,
			Provider: "moonshot",
			Model:    "glm-4",
		},
		"haiku": {
			Method:   config.MethodAPI,
			Provider: "anthropic",
			Model:    "claude-3-haiku-3-5",
		},
	}

	router, err := NewRouter(cfg, nil, WithRoutingConfig(configPath))
	require.NoError(t, err)
	assert.Equal(t, 1, len(router.rules))
	assert.Equal(t, "simple-cli", router.rules[0].ID)
}

func TestRouter_WithRoutingConfig_MissingFile(t *testing.T) {
	t.Parallel()

	cfg := worker.DefaultConfig()
	cfg.Providers = map[string]worker.ProviderDefinition{
		"glm": {
			Method:   config.MethodACP,
			Provider: "moonshot",
			Model:    "glm-4",
		},
	}

	missingPath := filepath.Join(t.TempDir(), "nonexistent.yaml")

	router, err := NewRouter(cfg, nil, WithRoutingConfig(missingPath))
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(router.rules), 1)
}

func TestRouter_WithRoutingConfig_InvalidYAML(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "routing.yaml")

	invalidContent := `
rules:
  - id: "test"
    priority: "not a number"
    match:
      complexity: "trivial"
`
	err := os.WriteFile(configPath, []byte(invalidContent), 0644)
	require.NoError(t, err)

	cfg := worker.DefaultConfig()
	cfg.Providers = map[string]worker.ProviderDefinition{
		"glm": {
			Method:   config.MethodACP,
			Provider: "moonshot",
			Model:    "glm-4",
		},
	}

	router, err := NewRouter(cfg, nil, WithRoutingConfig(configPath))
	require.NoError(t, err)
	assert.Empty(t, router.rules)
}

func TestRouter_RoutingConfigApplied(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "routing.yaml")

	configContent := `
rules:
  - id: "bulk-implementation"
    priority: 1000
    match:
      type:
        - "implement"
        - "feature"
      file_count: 3
    action:
      method: "acp"
      provider: "moonshot"
      model: "glm-4"
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	cfg := worker.DefaultConfig()
	cfg.Providers = map[string]worker.ProviderDefinition{
		"glm": {
			Method:   config.MethodACP,
			Provider: "moonshot",
			Model:    "glm-4",
		},
		"haiku": {
			Method:   config.MethodAPI,
			Provider: "anthropic",
			Model:    "claude-3-haiku-3-5",
		},
	}

	router, err := NewRouter(cfg, nil, WithRoutingConfig(configPath))
	require.NoError(t, err)

	task := execution.NewTask("Implement new feature")
	task = task.WithType("feature").WithContext(
		execution.NewTaskContext("/test").WithFiles([]string{"file1.go", "file2.go", "file3.go"}),
	)

	decision, err := router.Route(context.Background(), task)
	require.NoError(t, err)
	assert.Equal(t, config.MethodACP, decision.Method)
	assert.Equal(t, "moonshot", decision.Provider)
	assert.Equal(t, "glm-4", decision.Model)
	assert.Equal(t, "bulk-implementation", decision.RuleName)
}

func TestRouter_CustomConfigPath(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "custom-rules.yaml")

	configContent := `
rules:
  - id: "custom-rule"
    priority: 2000
    match:
      keywords:
        - "urgent"
    action:
      method: "api"
      provider: "anthropic"
      model: "claude-3-5-sonnet"
 `
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	cfg := worker.DefaultConfig()
	cfg.Providers = map[string]worker.ProviderDefinition{
		"glm": {
			Method:   config.MethodACP,
			Provider: "moonshot",
			Model:    "glm-4",
		},
		"haiku": {
			Method:   config.MethodAPI,
			Provider: "anthropic",
			Model:    "claude-3-5-sonnet",
		},
	}

	router, err := NewRouter(cfg, nil, WithRoutingConfig(configPath))
	require.NoError(t, err)

	task := execution.NewTask("URGENT: Fix critical bug")

	decision, err := router.Route(context.Background(), task)
	require.NoError(t, err)
	assert.Equal(t, config.MethodAPI, decision.Method)
	assert.Equal(t, "custom-rule", decision.RuleName)
}

func TestRouter_OverrideProvider(t *testing.T) {
	t.Parallel()

	router := createTestRouter()

	task := execution.NewTask("Fix a simple typo")
	task = task.WithType("bugfix").WithProvider("glm")

	decision, err := router.Route(context.Background(), task)
	require.NoError(t, err)
	assert.Equal(t, "glm", decision.Provider)
	assert.Equal(t, config.MethodACP, decision.Method)
	assert.Contains(t, strings.ToLower(decision.Reason), "override")
	assert.Contains(t, strings.ToLower(decision.Reason), "provider")
	assert.Equal(t, "override", decision.RuleName)
}

func TestRouter_OverrideModel(t *testing.T) {
	t.Parallel()

	router := createTestRouter()

	task := execution.NewTask("Implement feature")
	task = task.WithType("feature").WithModel("glm-4")

	decision, err := router.Route(context.Background(), task)
	require.NoError(t, err)
	assert.Equal(t, "glm-4", decision.Model)
	assert.Contains(t, strings.ToLower(decision.Reason), "override")
	assert.Contains(t, strings.ToLower(decision.Reason), "model")
	assert.Equal(t, "override", decision.RuleName)
}

func TestRouter_OverrideMethod(t *testing.T) {
	t.Parallel()

	router := createTestRouter()

	task := execution.NewTask("Fix bug")
	task = task.WithType("bugfix").WithMethod("api")

	decision, err := router.Route(context.Background(), task)
	require.NoError(t, err)
	assert.Equal(t, config.MethodAPI, decision.Method)
	assert.Contains(t, strings.ToLower(decision.Reason), "override")
	assert.Contains(t, strings.ToLower(decision.Reason), "method")
	assert.Equal(t, "override", decision.RuleName)
}

func TestRouter_OverrideAgent(t *testing.T) {
	t.Parallel()

	router := createTestRouter()

	task := execution.NewTask("Test agent override")
	task = task.WithAgent("ent-coder")

	decision, err := router.Route(context.Background(), task)
	require.NoError(t, err)
	assert.Contains(t, strings.ToLower(decision.Reason), "override")
	assert.Contains(t, strings.ToLower(decision.Reason), "agent")
	assert.Equal(t, "override", decision.RuleName)
}

func TestRouter_OverrideProviderAndModel(t *testing.T) {
	t.Parallel()

	router := createTestRouter()

	task := execution.NewTask("Complex refactor")
	task = task.WithType("refactor").WithProvider("glm").WithModel("glm-4")

	decision, err := router.Route(context.Background(), task)
	require.NoError(t, err)
	assert.Equal(t, "glm", decision.Provider)
	assert.Equal(t, "glm-4", decision.Model)
	assert.Contains(t, strings.ToLower(decision.Reason), "override")
	assert.Equal(t, "override", decision.RuleName)
}

func TestRouter_OverrideAll(t *testing.T) {
	t.Parallel()

	router := createTestRouter()

	task := execution.NewTask("Full override test")
	task = task.WithProvider("glm").WithModel("glm-4").WithMethod("acp").WithAgent("ent-coder")

	decision, err := router.Route(context.Background(), task)
	require.NoError(t, err)
	assert.Equal(t, "glm", decision.Provider)
	assert.Equal(t, "glm-4", decision.Model)
	assert.Equal(t, config.MethodACP, decision.Method)
	assert.Contains(t, strings.ToLower(decision.Reason), "provider")
	assert.Contains(t, strings.ToLower(decision.Reason), "model")
	assert.Contains(t, strings.ToLower(decision.Reason), "method")
	assert.Contains(t, strings.ToLower(decision.Reason), "agent")
	assert.Equal(t, "override", decision.RuleName)
}

func TestRouter_OverrideInvalidProvider(t *testing.T) {
	t.Parallel()

	router := createTestRouter()

	task := execution.NewTask("Test")
	task = task.WithProvider("nonexistent")

	_, err := router.Route(context.Background(), task)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrProviderNotFound)
}

func TestRouter_OverrideInvalidModel(t *testing.T) {
	t.Parallel()

	router := createTestRouter()

	task := execution.NewTask("Test")
	task = task.WithProvider("glm").WithModel("nonexistent-model")

	_, err := router.Route(context.Background(), task)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidModel)
}

func TestRouter_OverrideInvalidMethod(t *testing.T) {
	t.Parallel()

	router := createTestRouter()

	task := execution.NewTask("Test")
	task = task.WithMethod("invalid-method")

	_, err := router.Route(context.Background(), task)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidMethod)
}

func TestRouter_OverrideTakesPrecedence(t *testing.T) {
	t.Parallel()

	router := createTestRouter()

	rules := []RoutingRule{
		{
			ID:       "high-priority-rule",
			Priority: 1000,
			Match: MatchConditions{
				Type: []string{"feature"},
			},
			Action: RouteAction{
				Method:   "cli",
				Provider: "haiku",
			},
		},
	}

	router.SetRules(rules)

	task := execution.NewTask("Implement feature")
	task = task.WithType("feature").WithProvider("glm").WithMethod("acp")

	decision, err := router.Route(context.Background(), task)
	require.NoError(t, err)
	assert.Equal(t, "glm", decision.Provider)
	assert.Equal(t, config.MethodACP, decision.Method)
	assert.Equal(t, "override", decision.RuleName)
}

func createTestRouter() *Router {
	cfg := worker.DefaultConfig()
	cfg.Providers = map[string]worker.ProviderDefinition{
		"glm": {
			Method:   config.MethodACP,
			Provider: "moonshot",
			Model:    "glm-4",
		},
		"kimi": {
			Method:   config.MethodACP,
			Provider: "moonshot",
			Model:    "kimi-k2",
		},
		"haiku": {
			Method:   config.MethodAPI,
			Provider: "anthropic",
			Model:    "claude-3-haiku-3-5",
		},
	}

	router, err := NewRouter(cfg, nil)
	if err != nil {
		panic(err)
	}

	router.SetDefaults(DefaultRoutes{
		SimpleTasks:    "haiku",
		Implementation: "glm",
		LargeContext:   "kimi",
		ComplexTasks:   "glm",
	})

	return router
}

func TestRouter_Learning(t *testing.T) {
	t.Parallel()

	cfg := worker.DefaultConfig()
	cfg.Providers = map[string]worker.ProviderDefinition{
		"glm": {
			Method:   config.MethodACP,
			Provider: "moonshot",
			Model:    "glm-4",
		},
		"haiku": {
			Method:   config.MethodAPI,
			Provider: "anthropic",
			Model:    "claude-3-haiku-3-5",
		},
	}

	mem := memory.NewMemoryStore()
	router, err := NewRouter(cfg, mem)
	require.NoError(t, err)
	assert.NotNil(t, router)

	assert.True(t, router.IsLearningEnabled())

	router.EnableLearning(false)
	assert.False(t, router.IsLearningEnabled())

	router.EnableLearning(true)
	assert.True(t, router.IsLearningEnabled())
}

func TestRouter_RouteWithLearning(t *testing.T) {
	t.Parallel()

	cfg := worker.DefaultConfig()
	cfg.Providers = map[string]worker.ProviderDefinition{
		"glm": {
			Method:   config.MethodACP,
			Provider: "moonshot",
			Model:    "glm-4",
		},
	}

	mem := memory.NewMemoryStore()

	for i := 0; i < 10; i++ {
		err := mem.Store(&memory.Pattern{
			TaskType:    "implement",
			Provider:    "glm",
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

	router, err := NewRouter(cfg, mem)
	require.NoError(t, err)

	task := &execution.Task{
		Type:        "implement",
		Description: "Add new feature",
		Context: &execution.TaskContext{
			Files: []string{
				"file1.go",
				"file2.go",
			},
		},
	}

	decision, err := router.Route(context.Background(), task)
	require.NoError(t, err)
	assert.NotNil(t, decision)
	assert.Equal(t, "glm", decision.Provider)
	assert.Equal(t, "glm-4", decision.Model)
	assert.Contains(t, decision.Reason, "learned pattern")
}

func TestRouter_Failover_BudgetExceeded(t *testing.T) {
	t.Parallel()

	cfg := worker.DefaultConfig()
	cfg.Providers = map[string]worker.ProviderDefinition{
		"deepseek": {
			Method:   config.MethodACP,
			Provider: "deepseek",
			Model:    "deepseek-coder",
		},
		"glm": {
			Method:   config.MethodACP,
			Provider: "moonshot",
			Model:    "glm-4",
		},
	}

	router, err := NewRouter(cfg, nil)
	require.NoError(t, err)

	router.SetCostBudget(0.01)

	task := execution.NewTask("Complex task requiring many tokens")
	task = task.WithType("refactor").WithContext(
		execution.NewTaskContext("/test").WithFiles(make([]string, 50)),
	)

	decision, err := router.Route(context.Background(), task)
	require.NoError(t, err)
	assert.Contains(t, decision.Reason, "budget fallback")
	assert.Equal(t, "deepseek", decision.Provider)
	assert.Less(t, decision.EstimatedCost, 0.02)
}

func TestRouter_Failover_BudgetExceededAllProviders(t *testing.T) {
	t.Parallel()

	cfg := worker.DefaultConfig()
	cfg.Providers = map[string]worker.ProviderDefinition{
		"glm": {
			Method:   config.MethodACP,
			Provider: "moonshot",
			Model:    "glm-4",
		},
	}

	router, err := NewRouter(cfg, nil)
	require.NoError(t, err)

	router.SetCostBudget(0.001)

	task := execution.NewTask("Large complex task")

	_, err = router.Route(context.Background(), task)
	assert.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "insufficient budget")
	assert.Contains(t, strings.ToLower(err.Error()), "no cheaper provider")
}

func TestRouter_Failover_CostTrackingDuringFailover(t *testing.T) {
	t.Parallel()

	cfg := worker.DefaultConfig()
	cfg.Providers = map[string]worker.ProviderDefinition{
		"haiku": {
			Method:   config.MethodAPI,
			Provider: "anthropic",
			Model:    "claude-3-haiku-3-5",
		},
		"glm": {
			Method:   config.MethodACP,
			Provider: "moonshot",
			Model:    "glm-4",
		},
	}

	router, err := NewRouter(cfg, nil)
	require.NoError(t, err)

	router.SetCostBudget(0.10)

	initialBudget := router.GetRemainingBudget()
	assert.Equal(t, 0.10, initialBudget)

	router.RecordCost("glm", 0.02)

	afterFirst := router.GetRemainingBudget()
	assert.Equal(t, 0.08, afterFirst)

	costsByProvider := router.GetCostsByProvider()
	assert.Equal(t, 0.02, costsByProvider["glm"])
}

func TestRouter_Failover_BudgetReset(t *testing.T) {
	t.Parallel()

	cfg := worker.DefaultConfig()
	cfg.Providers = map[string]worker.ProviderDefinition{
		"haiku": {
			Method:   config.MethodAPI,
			Provider: "anthropic",
			Model:    "claude-3-haiku-3-5",
		},
	}

	router, err := NewRouter(cfg, nil)
	require.NoError(t, err)

	router.SetCostBudget(0.10)

	router.RecordCost("haiku", 0.05)

	assert.Equal(t, 0.05, router.GetRemainingBudget())

	router.ResetBudget()

	assert.Equal(t, 0.10, router.GetRemainingBudget())
	assert.Equal(t, 0.0, router.GetCostsByProvider()["haiku"])
}

func TestRouter_Failover_MultipleTasksDepleteBudget(t *testing.T) {
	t.Parallel()

	cfg := worker.DefaultConfig()
	cfg.Providers = map[string]worker.ProviderDefinition{
		"glm": {
			Method:   config.MethodACP,
			Provider: "moonshot",
			Model:    "glm-4",
		},
		"deepseek": {
			Method:   config.MethodACP,
			Provider: "deepseek",
			Model:    "deepseek-coder",
		},
	}

	router, err := NewRouter(cfg, nil)
	require.NoError(t, err)

	router.SetCostBudget(0.04)

	task := execution.NewTask("Implementation task")

	for i := 0; i < 3; i++ {
		decision, err := router.Route(context.Background(), task)
		if i < 2 {
			require.NoError(t, err)
			cost := 0.02
			router.RecordCost(decision.Provider, cost)
		} else {
			assert.Error(t, err)
			assert.Contains(t, strings.ToLower(err.Error()), "insufficient budget")
		}
	}
}

func TestRouter_Failover_ProviderPriorityDuringBudgetFailover(t *testing.T) {
	t.Parallel()

	cfg := worker.DefaultConfig()
	cfg.Providers = map[string]worker.ProviderDefinition{
		"deepseek": {
			Method:   config.MethodACP,
			Provider: "deepseek",
			Model:    "deepseek-coder",
		},
		"glm": {
			Method:   config.MethodACP,
			Provider: "moonshot",
			Model:    "glm-4",
		},
		"haiku": {
			Method:   config.MethodAPI,
			Provider: "anthropic",
			Model:    "claude-3-haiku-3-5",
		},
	}

	router, err := NewRouter(cfg, nil)
	require.NoError(t, err)

	router.SetCostBudget(0.01)

	task := execution.NewTask("Complex implementation task")
	task = task.WithType("refactor").WithContext(
		execution.NewTaskContext("/test").WithFiles(make([]string, 50)),
	)

	decision, err := router.Route(context.Background(), task)
	require.NoError(t, err)

	assert.NotEqual(t, "haiku", decision.Provider)
	assert.Contains(t, decision.Reason, "budget fallback")
}

func TestRouter_Failover_ZeroBudgetDisablesChecks(t *testing.T) {
	t.Parallel()

	cfg := worker.DefaultConfig()
	cfg.Providers = map[string]worker.ProviderDefinition{
		"haiku": {
			Method:   config.MethodAPI,
			Provider: "anthropic",
			Model:    "claude-3-haiku-3-5",
		},
		"glm": {
			Method:   config.MethodACP,
			Provider: "moonshot",
			Model:    "glm-4",
		},
	}

	router, err := NewRouter(cfg, nil)
	require.NoError(t, err)

	router.SetCostBudget(0)

	task := execution.NewTask("Very large expensive task")
	task = task.WithType("feature").WithContext(
		execution.NewTaskContext("/test").WithFiles(make([]string, 100)),
	)

	decision, err := router.Route(context.Background(), task)
	require.NoError(t, err)
	assert.NotNil(t, decision)
	assert.NotEmpty(t, decision.Provider)
}
