package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExecutionStrategy_String(t *testing.T) {
	tests := []struct {
		name     string
		strategy ExecutionStrategy
		want     string
	}{
		{"single", ExecutionStrategySingle, "single"},
		{"multi", ExecutionStrategyMulti, "multi"},
		{"parallel", ExecutionStrategyParallel, "parallel"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.strategy.String())
		})
	}
}

func TestExecutionStrategy_Valid(t *testing.T) {
	tests := []struct {
		name     string
		strategy ExecutionStrategy
		want     bool
	}{
		{"valid single", ExecutionStrategySingle, true},
		{"valid multi", ExecutionStrategyMulti, true},
		{"valid parallel", ExecutionStrategyParallel, true},
		{"invalid empty", ExecutionStrategy(""), false},
		{"invalid unknown", ExecutionStrategy("unknown"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.strategy.Valid())
		})
	}
}

func TestExecutionContext(t *testing.T) {
	ctx := ExecutionContext{
		Runtime:  RuntimeClaudeCode,
		Agent:    AgentRoleDeveloper,
		Strategy: ExecutionStrategySingle,
		ChangeID: "add-feature",
		TaskID:   "task-001",
		Budget:   100000,
		Metadata: map[string]string{
			"priority": "high",
			"author":   "system",
		},
	}

	assert.Equal(t, RuntimeClaudeCode, ctx.Runtime)
	assert.Equal(t, AgentRoleDeveloper, ctx.Agent)
	assert.Equal(t, ExecutionStrategySingle, ctx.Strategy)
	assert.Equal(t, "add-feature", ctx.ChangeID)
	assert.Equal(t, "task-001", ctx.TaskID)
	assert.Equal(t, 100000, ctx.Budget)
	assert.Equal(t, "high", ctx.Metadata["priority"])
	assert.Equal(t, "system", ctx.Metadata["author"])
}

func TestExecutionContext_Empty(t *testing.T) {
	ctx := ExecutionContext{}

	assert.Equal(t, Runtime(""), ctx.Runtime)
	assert.Equal(t, AgentRole(""), ctx.Agent)
	assert.Equal(t, ExecutionStrategy(""), ctx.Strategy)
	assert.Empty(t, ctx.ChangeID)
	assert.Empty(t, ctx.TaskID)
	assert.Equal(t, 0, ctx.Budget)
	assert.Nil(t, ctx.Metadata)
}

func TestExecutionResult_Success(t *testing.T) {
	result := ExecutionResult{
		Success:  true,
		Output:   "Task completed successfully",
		Error:    "",
		Tokens:   5000,
		Cost:     0.25,
		Duration: 30 * time.Second,
		Metadata: map[string]interface{}{
			"files_modified": 3,
			"tests_passed":   true,
		},
	}

	assert.True(t, result.Success)
	assert.Equal(t, "Task completed successfully", result.Output)
	assert.Empty(t, result.Error)
	assert.Equal(t, 5000, result.Tokens)
	assert.Equal(t, 0.25, result.Cost)
	assert.Equal(t, 30*time.Second, result.Duration)
	assert.Equal(t, 3, result.Metadata["files_modified"])
	assert.Equal(t, true, result.Metadata["tests_passed"])
}

func TestExecutionResult_Failure(t *testing.T) {
	result := ExecutionResult{
		Success:  false,
		Output:   "",
		Error:    "compilation failed",
		Tokens:   2000,
		Cost:     0.10,
		Duration: 10 * time.Second,
		Metadata: map[string]interface{}{
			"error_type": "compilation",
		},
	}

	assert.False(t, result.Success)
	assert.Empty(t, result.Output)
	assert.Equal(t, "compilation failed", result.Error)
	assert.Equal(t, 2000, result.Tokens)
	assert.Equal(t, 0.10, result.Cost)
	assert.Equal(t, 10*time.Second, result.Duration)
	assert.Equal(t, "compilation", result.Metadata["error_type"])
}

func TestExecutionResult_Empty(t *testing.T) {
	result := ExecutionResult{}

	assert.False(t, result.Success)
	assert.Empty(t, result.Output)
	assert.Empty(t, result.Error)
	assert.Equal(t, 0, result.Tokens)
	assert.Equal(t, 0.0, result.Cost)
	assert.Equal(t, time.Duration(0), result.Duration)
	assert.Nil(t, result.Metadata)
}

func TestExecutionContext_WithMetadata(t *testing.T) {
	metadata := map[string]string{
		"source":  "cli",
		"version": "1.0.0",
	}

	ctx := ExecutionContext{
		Runtime:  RuntimeCLI,
		Agent:    AgentRoleOps,
		Strategy: ExecutionStrategySingle,
		Metadata: metadata,
	}

	assert.Equal(t, "cli", ctx.Metadata["source"])
	assert.Equal(t, "1.0.0", ctx.Metadata["version"])
	assert.Len(t, ctx.Metadata, 2)
}

func TestExecutionResult_WithComplexMetadata(t *testing.T) {
	metadata := map[string]interface{}{
		"files": []string{"file1.go", "file2.go"},
		"stats": map[string]int{
			"lines_added":   100,
			"lines_removed": 50,
		},
		"timestamp": time.Now(),
	}

	result := ExecutionResult{
		Success:  true,
		Metadata: metadata,
	}

	files := result.Metadata["files"].([]string)
	assert.Len(t, files, 2)
	assert.Contains(t, files, "file1.go")

	stats := result.Metadata["stats"].(map[string]int)
	assert.Equal(t, 100, stats["lines_added"])
	assert.Equal(t, 50, stats["lines_removed"])

	_, ok := result.Metadata["timestamp"].(time.Time)
	assert.True(t, ok)
}
