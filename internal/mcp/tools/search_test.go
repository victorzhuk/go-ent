package tools

//nolint:gosec // test file with necessary file operations

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchAccuracy(t *testing.T) {
	t.Parallel()

	// Test cases map queries to expected top-3 results
	// Each test case should have at least one of the expected tools in top-3
	tests := []struct {
		name           string
		query          string
		expectedTop3   []string // At least one of these should be in top 3
		description    string
		strictTopMatch string // If set, this MUST be the #1 result
	}{
		{
			name:           "spec initialization",
			query:          "initialize project",
			expectedTop3:   []string{"spec_init", "registry_init"},
			strictTopMatch: "spec_init",
			description:    "Init query should find spec_init as top result",
		},
		{
			name:           "validation",
			query:          "validate spec",
			expectedTop3:   []string{"spec_validate"},
			strictTopMatch: "spec_validate",
			description:    "Validation query should find spec_validate",
		},
		{
			name:         "list tasks",
			query:        "list all tasks",
			expectedTop3: []string{"registry_list", "spec_list"},
			description:  "List tasks should find registry tools",
		},
		{
			name:           "next task",
			query:          "get next task",
			expectedTop3:   []string{"registry_next"},
			strictTopMatch: "registry_next",
			description:    "Next task query should find registry_next",
		},
		{
			name:         "update task status",
			query:        "update task",
			expectedTop3: []string{"registry_update", "spec_update"},
			description:  "Update query should find update tools",
		},
		{
			name:           "archive change",
			query:          "archive completed change",
			expectedTop3:   []string{"spec_archive"},
			strictTopMatch: "spec_archive",
			description:    "Archive query should find spec_archive",
		},
		{
			name:           "create spec",
			query:          "create new specification",
			expectedTop3:   []string{"spec_create"},
			strictTopMatch: "spec_create",
			description:    "Create spec query should find spec_create",
		},
		{
			name:           "show details",
			query:          "show spec details",
			expectedTop3:   []string{"spec_show"},
			strictTopMatch: "spec_show",
			description:    "Show query should find spec_show",
		},
		{
			name:         "workflow management",
			query:        "start workflow",
			expectedTop3: []string{"workflow_start", "loop_start"},
			description:  "Workflow query should find workflow tools",
		},
		{
			name:           "approve workflow",
			query:          "approve current workflow",
			expectedTop3:   []string{"workflow_approve"},
			strictTopMatch: "workflow_approve",
			description:    "Approve query should find workflow_approve",
		},
		{
			name:         "workflow status",
			query:        "check workflow status",
			expectedTop3: []string{"workflow_status", "loop_get"},
			description:  "Status query should find status tools",
		},
		{
			name:         "autonomous loop",
			query:        "start autonomous loop",
			expectedTop3: []string{"loop_start", "workflow_start"},
			description:  "Autonomous loop query should find loop tools",
		},
		{
			name:           "cancel loop",
			query:          "cancel running loop",
			expectedTop3:   []string{"loop_cancel"},
			strictTopMatch: "loop_cancel",
			description:    "Cancel loop query should find loop_cancel",
		},
		{
			name:           "sync registry",
			query:          "synchronize registry",
			expectedTop3:   []string{"registry_sync"},
			strictTopMatch: "registry_sync",
			description:    "Sync query should find registry_sync",
		},
		{
			name:           "task dependencies",
			query:          "manage task dependencies",
			expectedTop3:   []string{"registry_deps"},
			strictTopMatch: "registry_deps",
			description:    "Dependencies query should find registry_deps",
		},
		{
			name:         "code generation",
			query:        "generate code",
			expectedTop3: []string{"generate", "generate_component", "generate_from_spec"},
			description:  "Generate query should find generation tools",
		},
		{
			name:         "generate from spec",
			query:        "generate project from spec",
			expectedTop3: []string{"generate_from_spec", "generate"},
			description:  "Spec generation query should find generate_from_spec",
		},
		{
			name:         "component scaffold",
			query:        "scaffold component",
			expectedTop3: []string{"generate_component"},
			description:  "Component query should find generate_component",
		},
		{
			name:           "list archetypes",
			query:          "list available archetypes",
			expectedTop3:   []string{"list_archetypes"},
			strictTopMatch: "list_archetypes",
			description:    "Archetype query should find list_archetypes",
		},
		{
			name:           "execute with agent",
			query:          "execute task with agent",
			expectedTop3:   []string{"agent_execute"},
			strictTopMatch: "agent_execute",
			description:    "Agent execution query should find agent_execute",
		},
		{
			name:           "find tools",
			query:          "search for tools",
			expectedTop3:   []string{"tool_find"},
			strictTopMatch: "tool_find",
			description:    "Tool search query should find tool_find",
		},
		{
			name:           "tool details",
			query:          "describe tool",
			expectedTop3:   []string{"tool_describe"},
			strictTopMatch: "tool_describe",
			description:    "Describe query should find tool_describe",
		},
		{
			name:           "load tools",
			query:          "load activate tools",
			expectedTop3:   []string{"tool_load"},
			strictTopMatch: "tool_load",
			description:    "Load query should find tool_load",
		},
		{
			name:           "active tools",
			query:          "list active tools",
			expectedTop3:   []string{"tool_active"},
			strictTopMatch: "tool_active",
			description:    "Active tools query should find tool_active",
		},
		{
			name:           "delete spec",
			query:          "delete specification",
			expectedTop3:   []string{"spec_delete"},
			strictTopMatch: "spec_delete",
			description:    "Delete query should find spec_delete",
		},
	}

	// Build search index with all test tools
	index := NewSearchIndex()
	docs := buildTestDocuments()
	err := index.Index(docs)
	require.NoError(t, err)

	successCount := 0
	strictSuccessCount := 0
	strictTotal := 0

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := index.Search(tt.query, 3)
			require.Greater(t, len(results), 0, "Query '%s' returned no results. %s", tt.query, tt.description)

			// Extract top 3 tool names
			top3Names := make([]string, 0, len(results))
			for _, r := range results {
				top3Names = append(top3Names, r.ToolName)
			}

			// Check if at least one expected tool is in top 3
			foundExpected := false
			for _, expected := range tt.expectedTop3 {
				for _, actual := range top3Names {
					if expected == actual {
						foundExpected = true
						break
					}
				}
				if foundExpected {
					break
				}
			}

			if foundExpected {
				successCount++
			}

			assert.True(t, foundExpected,
				"Expected at least one of %v in top 3 results for query '%s'. Got: %v. %s",
				tt.expectedTop3, tt.query, top3Names, tt.description)

			// Check strict top match if specified
			if tt.strictTopMatch != "" {
				strictTotal++
				isTopMatch := results[0].ToolName == tt.strictTopMatch
				if isTopMatch {
					strictSuccessCount++
				}
				assert.Equal(t, tt.strictTopMatch, results[0].ToolName,
					"Expected '%s' as top result for query '%s', got '%s'. %s",
					tt.strictTopMatch, tt.query, results[0].ToolName, tt.description)
			}
		})
	}

	// Calculate accuracy
	totalTests := len(tests)
	accuracy := float64(successCount) / float64(totalTests) * 100
	strictAccuracy := float64(strictSuccessCount) / float64(strictTotal) * 100

	t.Logf("\n=== Search Accuracy Report ===")
	t.Logf("Top-3 Accuracy: %.1f%% (%d/%d tests passed)", accuracy, successCount, totalTests)
	t.Logf("Top-1 Strict Accuracy: %.1f%% (%d/%d tests passed)", strictAccuracy, strictSuccessCount, strictTotal)
	t.Logf("Threshold: 80%%")

	// Assert overall accuracy meets threshold
	assert.GreaterOrEqual(t, accuracy, 80.0,
		"Search accuracy (%.1f%%) is below 80%% threshold. %d/%d tests passed.",
		accuracy, successCount, totalTests)
}

func TestSearchEdgeCases(t *testing.T) {
	t.Parallel()

	index := NewSearchIndex()
	docs := buildTestDocuments()
	err := index.Index(docs)
	require.NoError(t, err)

	tests := []struct {
		name        string
		query       string
		expectEmpty bool
		description string
	}{
		{
			name:        "empty query",
			query:       "",
			expectEmpty: true,
			description: "Empty query should return no results",
		},
		{
			name:        "stopwords only",
			query:       "the a an and or",
			expectEmpty: true,
			description: "Stopwords only should return no results",
		},
		{
			name:        "nonsense query",
			query:       "xyzabc123impossible",
			expectEmpty: true,
			description: "Nonsense query should return no results",
		},
		{
			name:        "single character",
			query:       "a",
			expectEmpty: true,
			description: "Single character query should return no results",
		},
		{
			name:        "special characters only",
			query:       "!@#$%^&*()",
			expectEmpty: true,
			description: "Special characters should return no results",
		},
		{
			name:        "mixed case",
			query:       "VALIDATE Spec",
			expectEmpty: false,
			description: "Mixed case should work (case insensitive)",
		},
		{
			name:        "hyphenated",
			query:       "self-correction",
			expectEmpty: false,
			description: "Hyphenated words should be tokenized",
		},
		{
			name:        "short word filtered",
			query:       "va",
			expectEmpty: true,
			description: "Two-letter words get filtered out",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := index.Search(tt.query, 10)

			if tt.expectEmpty {
				assert.Equal(t, 0, len(results),
					"Query '%s' should return no results. Got %d results. %s",
					tt.query, len(results), tt.description)
			} else {
				assert.Greater(t, len(results), 0,
					"Query '%s' should return results. %s",
					tt.query, tt.description)
			}
		})
	}
}

func TestSearchRelevanceScoring(t *testing.T) {
	t.Parallel()

	index := NewSearchIndex()
	docs := buildTestDocuments()
	err := index.Index(docs)
	require.NoError(t, err)

	// Test that more specific queries return more relevant results
	tests := []struct {
		name            string
		query           string
		expectedTopTool string
		description     string
	}{
		{
			name:            "exact tool name match",
			query:           "spec_validate",
			expectedTopTool: "spec_validate",
			description:     "Exact name match should be top result",
		},
		{
			name:            "exact match with spaces",
			query:           "spec validate",
			expectedTopTool: "spec_validate",
			description:     "Name match with spaces should be top result",
		},
		{
			name:            "registry list exact",
			query:           "registry_list",
			expectedTopTool: "registry_list",
			description:     "Exact registry_list match",
		},
		{
			name:            "tool find exact",
			query:           "tool_find",
			expectedTopTool: "tool_find",
			description:     "Exact tool_find match",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := index.Search(tt.query, 5)
			require.Greater(t, len(results), 0, "Query '%s' returned no results", tt.query)

			topResult := results[0].ToolName
			assert.Equal(t, tt.expectedTopTool, topResult,
				"Expected '%s' as top result for query '%s', got '%s'. %s",
				tt.expectedTopTool, tt.query, topResult, tt.description)

			// Verify score is positive
			assert.Greater(t, results[0].Score, 0.0,
				"Top result should have positive score")

			// Verify results are sorted by score
			for i := 1; i < len(results); i++ {
				assert.GreaterOrEqual(t, results[i-1].Score, results[i].Score,
					"Results should be sorted by score descending")
			}
		})
	}
}

func TestExtractTerms(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple text",
			input:    "hello world",
			expected: []string{"hello", "world"},
		},
		{
			name:     "with stopwords",
			input:    "the quick brown fox",
			expected: []string{"quick", "brown", "fox"},
		},
		{
			name:     "mixed case",
			input:    "Hello WORLD",
			expected: []string{"hello", "world"},
		},
		{
			name:     "with punctuation",
			input:    "hello, world!",
			expected: []string{"hello", "world"},
		},
		{
			name:     "hyphenated",
			input:    "self-correction autonomous-system",
			expected: []string{"self", "correction", "autonomous", "system"},
		},
		{
			name:     "with numbers",
			input:    "task123 test456",
			expected: []string{"task123", "test456"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "stopwords only",
			input:    "the a an and",
			expected: nil,
		},
		{
			name:     "single letter words filtered",
			input:    "a b c test",
			expected: []string{"test"},
		},
		{
			name:     "underscore separated",
			input:    "spec_init registry_list",
			expected: []string{"spec", "init", "registry", "list"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractTerms(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsStopword(t *testing.T) {
	t.Parallel()

	stopwords := []string{"a", "an", "and", "are", "as", "at", "be", "by", "for", "from",
		"has", "he", "in", "is", "it", "its", "of", "on", "that", "the",
		"to", "was", "will", "with"}

	notStopwords := []string{"hello", "world", "test", "spec", "registry", "validate",
		"loop", "workflow", "agent", "execute"}

	for _, word := range stopwords {
		assert.True(t, isStopword(word), "Word '%s' should be a stopword", word)
	}

	for _, word := range notStopwords {
		assert.False(t, isStopword(word), "Word '%s' should not be a stopword", word)
	}
}

func TestBuildDocument(t *testing.T) {
	t.Parallel()

	doc := BuildDocument(1, "spec_validate", "Validate OpenSpec files")

	assert.Equal(t, 1, doc.ID)
	assert.Equal(t, "spec_validate", doc.ToolName)
	assert.Contains(t, doc.Terms, "spec")
	assert.Contains(t, doc.Terms, "validate")
	assert.Contains(t, doc.Terms, "openspec")
	assert.Contains(t, doc.Terms, "files")
	assert.NotEmpty(t, doc.TF)

	// Verify TF values sum to reasonable range
	totalTF := 0.0
	for _, tf := range doc.TF {
		totalTF += tf
		assert.Greater(t, tf, 0.0, "TF should be positive")
		assert.LessOrEqual(t, tf, 1.0, "TF should not exceed 1.0")
	}
}

// buildTestDocuments creates the full set of test tool documents
func buildTestDocuments() []Document {
	testTools := []struct {
		name        string
		description string
	}{
		{"spec_init", "Initialize openspec folder in a project directory"},
		{"spec_create", "Create a new spec, change, or task"},
		{"spec_update", "Update an existing spec, change, or task"},
		{"spec_delete", "Delete a spec, change, or task"},
		{"spec_list", "List specs, changes, or tasks"},
		{"spec_show", "Show detailed content of a spec, change, or task"},
		{"spec_validate", "Validate OpenSpec files. Type can be 'spec', 'change', or 'all'"},
		{"spec_archive", "Archive a completed change and optionally merge deltas into specs"},
		{"registry_list", "List tasks from the OpenSpec registry with optional filters"},
		{"registry_next", "Get the next recommended task(s) based on priority and dependencies"},
		{"registry_update", "Update task status, priority, or assignment in the registry"},
		{"registry_sync", "Synchronize registry from tasks.md files"},
		{"registry_deps", "Manage task dependencies. Supports cross-change dependencies"},
		{"registry_init", "Initialize an empty registry.yaml file"},
		{"workflow_start", "Start a guided workflow with state tracking and wait points"},
		{"workflow_status", "Check current workflow status and wait points"},
		{"workflow_approve", "Approve the current wait point and continue workflow"},
		{"loop_start", "Start autonomous loop with self-correction"},
		{"loop_get", "Get current loop state"},
		{"loop_set", "Update loop state (iteration, error, adjustment, status)"},
		{"loop_cancel", "Cancel running loop"},
		{"generate", "Generate a new Go project from templates"},
		{"generate_component", "Generate a component scaffold from a spec file"},
		{"generate_from_spec", "Generate a complete project from a spec file"},
		{"list_archetypes", "List available project archetypes"},
		{"agent_execute", "Execute a task with automatic agent selection based on complexity"},
		{"tool_find", "Search for tools by query using TF-IDF relevance scoring"},
		{"tool_describe", "Get detailed information about a specific tool"},
		{"tool_load", "Load (activate) one or more tools into the active set"},
		{"tool_active", "List currently active (loaded) tools"},
	}

	docs := make([]Document, len(testTools))
	for i, tool := range testTools {
		docs[i] = BuildDocument(i, tool.name, tool.description)
	}
	return docs
}
