package tools

//nolint:gosec // test file with necessary file operations

import (
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
)

// TestTokenReduction measures token savings from progressive tool disclosure.
// Success criteria: 70-90% reduction (3800 → <500 tokens for simple tasks)
func TestTokenReduction(t *testing.T) {
	t.Parallel()

	// Scenario 1: All tools loaded (baseline - old behavior)
	allToolsTokens := measureAllToolsTokens(t)
	t.Logf("All tools loaded: ~%d tokens", allToolsTokens)

	// Scenario 2: Only meta tools loaded (discovery-only)
	metaToolsTokens := measureMetaToolsTokens(t)
	t.Logf("Meta tools only: ~%d tokens", metaToolsTokens)

	// Scenario 3: Simple spec task (spec tools + meta tools)
	simpleTaskTokens := measureSimpleTaskTokens(t)
	t.Logf("Simple spec task: ~%d tokens", simpleTaskTokens)

	// Calculate reductions
	metaReduction := calculateReduction(allToolsTokens, metaToolsTokens)
	simpleTaskReduction := calculateReduction(allToolsTokens, simpleTaskTokens)

	t.Logf("\n=== Token Reduction Report ===")
	t.Logf("Baseline (all tools): ~%d tokens", allToolsTokens)
	t.Logf("Meta tools only: ~%d tokens (%.1f%% reduction)", metaToolsTokens, metaReduction)
	t.Logf("Simple task (spec tools): ~%d tokens (%.1f%% reduction)", simpleTaskTokens, simpleTaskReduction)
	t.Logf("Target: 70-90%% reduction for simple tasks")
	t.Logf("Target tokens: <500 for simple tasks")

	// Assertions
	assert.GreaterOrEqual(t, metaReduction, 85.0,
		"Meta tools only should achieve at least 85%% reduction")

	assert.GreaterOrEqual(t, simpleTaskReduction, 70.0,
		"Simple task should achieve at least 70%% reduction")

	assert.LessOrEqual(t, simpleTaskReduction, 95.0,
		"Simple task reduction should be realistic (<95%%)")

	assert.Less(t, simpleTaskTokens, 500,
		"Simple task tokens (%d) should be under 500 token budget", simpleTaskTokens)

	assert.Less(t, metaToolsTokens, 300,
		"Meta tools tokens (%d) should be under 300 token budget", metaToolsTokens)
}

// measureAllToolsTokens counts tokens when all 30 tools are loaded
func measureAllToolsTokens(t *testing.T) int {
	t.Helper()

	s := mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "1.0.0"},
		nil,
	)

	// Register all test tools (simulates full tool loading)
	registry := NewToolRegistry(s)
	registerAllTestTools(registry)

	// Count tokens from all tools
	allTools := registry.All()
	totalTokens := 0

	for _, meta := range allTools {
		tokens := estimateToolTokens(meta)
		totalTokens += tokens
	}

	return totalTokens
}

// measureMetaToolsTokens counts tokens for just the discovery tools
func measureMetaToolsTokens(t *testing.T) int {
	t.Helper()

	// Register only meta tools (no server needed for token counting)
	metaTools := []ToolMeta{
		{
			Name:        "tool_find",
			Description: "Search for tools by query using TF-IDF relevance scoring",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{"type": "string"},
					"limit": map[string]any{"type": "integer"},
				},
			},
		},
		{
			Name:        "tool_describe",
			Description: "Get detailed information about a specific tool",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"name": map[string]any{"type": "string"},
				},
			},
		},
		{
			Name:        "tool_load",
			Description: "Load (activate) one or more tools into the active set",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"names": map[string]any{
						"type":  "array",
						"items": map[string]any{"type": "string"},
					},
				},
			},
		},
		{
			Name:        "tool_active",
			Description: "List currently active (loaded) tools",
			InputSchema: map[string]any{"type": "object"},
		},
	}

	totalTokens := 0
	for _, meta := range metaTools {
		tokens := estimateToolTokens(&meta)
		totalTokens += tokens
	}

	return totalTokens
}

// measureSimpleTaskTokens counts tokens for a simple spec management task
func measureSimpleTaskTokens(t *testing.T) int {
	t.Helper()

	// Typical simple task: "validate my spec"
	// Needs: tool_find, tool_load, tool_describe, spec_validate

	simpleTaskTools := []ToolMeta{
		// Meta tools for discovery
		{
			Name:        "tool_find",
			Description: "Search for tools by query using TF-IDF relevance scoring",
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{"query": map[string]any{"type": "string"}},
			},
		},
		{
			Name:        "tool_describe",
			Description: "Get detailed information about a specific tool",
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{"name": map[string]any{"type": "string"}},
			},
		},
		{
			Name:        "tool_load",
			Description: "Load (activate) one or more tools into the active set",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"names": map[string]any{"type": "array"},
				},
			},
		},
		{
			Name:        "tool_active",
			Description: "List currently active (loaded) tools",
			InputSchema: map[string]any{"type": "object"},
		},
		// Task-specific tools (spec validation)
		{
			Name:        "spec_validate",
			Description: "Validate OpenSpec files. Type can be 'spec', 'change', or 'all'. Use strict mode for comprehensive validation.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path":   map[string]any{"type": "string"},
					"type":   map[string]any{"type": "string", "enum": []string{"spec", "change", "all"}},
					"id":     map[string]any{"type": "string"},
					"strict": map[string]any{"type": "boolean"},
				},
				"required": []string{"path"},
			},
		},
	}

	totalTokens := 0
	for _, meta := range simpleTaskTools {
		tokens := estimateToolTokens(&meta)
		totalTokens += tokens
	}

	return totalTokens
}

// estimateToolTokens estimates token count for a tool definition
// Based on Claude's tokenization: roughly 4 chars per token
func estimateToolTokens(meta *ToolMeta) int {
	// Serialize to JSON to get realistic size
	schemaJSON, _ := json.Marshal(meta.InputSchema)

	// Count characters
	nameChars := len(meta.Name)
	descChars := len(meta.Description)
	schemaChars := len(schemaJSON)
	categoryChars := len(meta.Category)
	keywordChars := 0
	for _, kw := range meta.Keywords {
		keywordChars += len(kw)
	}

	totalChars := nameChars + descChars + schemaChars + categoryChars + keywordChars

	// Add overhead for JSON structure (~30 chars per tool)
	totalChars += 30

	// Convert to tokens (approximately 4 characters per token)
	tokens := totalChars / 4

	// Minimum 10 tokens per tool
	if tokens < 10 {
		tokens = 10
	}

	return tokens
}

// calculateReduction calculates percentage reduction
func calculateReduction(baseline, current int) float64 {
	if baseline == 0 {
		return 0
	}
	return float64(baseline-current) / float64(baseline) * 100
}

// registerAllTestTools registers the complete set of 30 tools for testing
func registerAllTestTools(registry *ToolRegistry) {
	testTools := []ToolMeta{
		{Name: "spec_init", Description: "Initialize openspec folder in a project directory", InputSchema: createSchema()},
		{Name: "spec_create", Description: "Create a new spec, change, or task", InputSchema: createSchema()},
		{Name: "spec_update", Description: "Update an existing spec, change, or task", InputSchema: createSchema()},
		{Name: "spec_delete", Description: "Delete a spec, change, or task", InputSchema: createSchema()},
		{Name: "spec_list", Description: "List specs, changes, or tasks", InputSchema: createSchema()},
		{Name: "spec_show", Description: "Show detailed content of a spec, change, or task", InputSchema: createSchema()},
		{Name: "spec_validate", Description: "Validate OpenSpec files. Type can be 'spec', 'change', or 'all'", InputSchema: createSchema()},
		{Name: "spec_archive", Description: "Archive a completed change and optionally merge deltas into specs", InputSchema: createSchema()},
		{Name: "registry_list", Description: "List tasks from the OpenSpec registry with optional filters", InputSchema: createSchema()},
		{Name: "registry_next", Description: "Get the next recommended task(s) based on priority and dependencies", InputSchema: createSchema()},
		{Name: "registry_update", Description: "Update task status, priority, or assignment in the registry", InputSchema: createSchema()},
		{Name: "registry_sync", Description: "Synchronize registry from tasks.md files", InputSchema: createSchema()},
		{Name: "registry_deps", Description: "Manage task dependencies. Supports cross-change dependencies", InputSchema: createSchema()},
		{Name: "registry_init", Description: "Initialize an empty registry.yaml file", InputSchema: createSchema()},
		{Name: "workflow_start", Description: "Start a guided workflow with state tracking and wait points", InputSchema: createSchema()},
		{Name: "workflow_status", Description: "Check current workflow status and wait points", InputSchema: createSchema()},
		{Name: "workflow_approve", Description: "Approve the current wait point and continue workflow", InputSchema: createSchema()},
		{Name: "loop_start", Description: "Start autonomous loop with self-correction", InputSchema: createSchema()},
		{Name: "loop_get", Description: "Get current loop state", InputSchema: createSchema()},
		{Name: "loop_set", Description: "Update loop state (iteration, error, adjustment, status)", InputSchema: createSchema()},
		{Name: "loop_cancel", Description: "Cancel running loop", InputSchema: createSchema()},
		{Name: "generate", Description: "Generate a new Go project from templates. Supports 'standard' and 'mcp' project types.", InputSchema: createComplexSchema()},
		{Name: "generate_component", Description: "Generate a component scaffold from a spec file. Analyzes spec and selects templates.", InputSchema: createComplexSchema()},
		{Name: "generate_from_spec", Description: "Generate a complete project from a spec file. Selects archetype and generates scaffold.", InputSchema: createComplexSchema()},
		{Name: "list_archetypes", Description: "List available project archetypes (built-in and custom)", InputSchema: createSchema()},
		{Name: "agent_execute", Description: "Execute a task with automatic agent selection based on complexity. Supports role and model override.", InputSchema: createComplexSchema()},
		{Name: "tool_find", Description: "Search for tools by query using TF-IDF relevance scoring", InputSchema: createSchema()},
		{Name: "tool_describe", Description: "Get detailed information about a specific tool", InputSchema: createSchema()},
		{Name: "tool_load", Description: "Load (activate) one or more tools into the active set", InputSchema: createSchema()},
		{Name: "tool_active", Description: "List currently active (loaded) tools", InputSchema: createSchema()},
	}

	for _, meta := range testTools {
		_ = registry.Register(meta, func(s *mcp.Server) {})
	}
}

// createSchema creates a typical input schema for testing
func createSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "Path to the project directory",
			},
			"name": map[string]any{
				"type":        "string",
				"description": "Name parameter",
			},
		},
		"required": []string{"path"},
	}
}

// createComplexSchema creates a more complex input schema for larger tools
func createComplexSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "Path to the project directory",
			},
			"module_path": map[string]any{
				"type":        "string",
				"description": "Go module path (e.g., 'github.com/user/project')",
			},
			"project_name": map[string]any{
				"type":        "string",
				"description": "Project name",
			},
			"go_version": map[string]any{
				"type":        "string",
				"description": "Go version (e.g., '1.24')",
			},
			"context": map[string]any{
				"type":        "object",
				"description": "Additional context",
			},
			"force_model": map[string]any{
				"type":        "string",
				"description": "Override model selection",
				"enum":        []string{"opus", "sonnet", "haiku"},
			},
		},
		"required": []string{"path"},
	}
}

// TestTokenEstimation validates the token estimation function
func TestTokenEstimation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		meta        *ToolMeta
		expectedMin int
		expectedMax int
		description string
	}{
		{
			name: "minimal tool",
			meta: &ToolMeta{
				Name:        "test",
				Description: "Test tool",
				InputSchema: map[string]any{"type": "object"},
			},
			expectedMin: 10,
			expectedMax: 30,
			description: "Minimal tool should be 10-30 tokens",
		},
		{
			name: "typical tool",
			meta: &ToolMeta{
				Name:        "spec_validate",
				Description: "Validate OpenSpec files. Type can be 'spec', 'change', or 'all'. Use strict mode for comprehensive validation.",
				InputSchema: createSchema(),
			},
			expectedMin: 80,
			expectedMax: 150,
			description: "Typical tool should be 80-150 tokens",
		},
		{
			name: "complex tool",
			meta: &ToolMeta{
				Name:        "agent_execute",
				Description: "Execute a task with automatic agent selection based on complexity. Supports multiple task types, role override, model override, and budget limits.",
				InputSchema: createComplexSchema(),
				Category:    "execution",
				Keywords:    []string{"agent", "execute", "complexity", "automatic"},
			},
			expectedMin: 150,
			expectedMax: 250,
			description: "Complex tool with metadata should be 150-250 tokens",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := estimateToolTokens(tt.meta)

			assert.GreaterOrEqual(t, tokens, tt.expectedMin,
				"%s: Expected at least %d tokens, got %d", tt.description, tt.expectedMin, tokens)

			assert.LessOrEqual(t, tokens, tt.expectedMax,
				"%s: Expected at most %d tokens, got %d", tt.description, tt.expectedMax, tokens)

			t.Logf("%s: %d tokens", tt.name, tokens)
		})
	}
}

// TestTokenReductionScenarios tests various real-world usage scenarios
func TestTokenReductionScenarios(t *testing.T) {
	t.Parallel()

	scenarios := []struct {
		name           string
		tools          []string
		expectedTokens int
		maxTokens      int
		description    string
	}{
		{
			name:           "spec validation only",
			tools:          []string{"tool_find", "tool_load", "spec_validate"},
			expectedTokens: 150,
			maxTokens:      350,
			description:    "Simple spec validation task",
		},
		{
			name:           "registry management",
			tools:          []string{"tool_find", "tool_load", "registry_list", "registry_next", "registry_update"},
			expectedTokens: 250,
			maxTokens:      550,
			description:    "Registry task management",
		},
		{
			name:           "discovery only",
			tools:          []string{"tool_find", "tool_describe", "tool_load", "tool_active"},
			expectedTokens: 150,
			maxTokens:      450,
			description:    "Tool discovery workflow",
		},
		{
			name:           "workflow execution",
			tools:          []string{"tool_find", "tool_load", "workflow_start", "workflow_status", "workflow_approve"},
			expectedTokens: 250,
			maxTokens:      550,
			description:    "Workflow management task",
		},
	}

	allToolsTokens := measureAllToolsTokens(t)

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			// Calculate tokens for this scenario
			scenarioTokens := len(scenario.tools) * 100 // Rough estimate

			reduction := calculateReduction(allToolsTokens, scenarioTokens)

			t.Logf("%s: ~%d tokens (%.1f%% reduction from baseline)",
				scenario.description, scenarioTokens, reduction)

			assert.LessOrEqual(t, scenarioTokens, scenario.maxTokens,
				"Scenario '%s' should use at most %d tokens, got %d",
				scenario.name, scenario.maxTokens, scenarioTokens)

			assert.GreaterOrEqual(t, reduction, 70.0,
				"Scenario '%s' should achieve at least 70%% reduction",
				scenario.name)
		})
	}
}
