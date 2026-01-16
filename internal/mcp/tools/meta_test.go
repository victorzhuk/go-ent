package tools

//nolint:gosec // test file with necessary file operations

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToolFind(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		query          string
		limit          int
		expectedMinLen int
		expectedTools  []string
		description    string
	}{
		{
			name:           "spec management",
			query:          "manage specs",
			limit:          5,
			expectedMinLen: 2,
			expectedTools:  []string{"spec_create", "spec_update", "spec_list"},
			description:    "Should find spec CRUD tools",
		},
		{
			name:           "validation",
			query:          "validate",
			limit:          3,
			expectedMinLen: 1,
			expectedTools:  []string{"spec_validate"},
			description:    "Should find validation tools",
		},
		{
			name:           "task registry",
			query:          "registry tasks",
			limit:          10,
			expectedMinLen: 3,
			expectedTools:  []string{"registry_list", "registry_next", "registry_update"},
			description:    "Should find registry tools",
		},
		{
			name:           "initialization",
			query:          "initialize project",
			limit:          5,
			expectedMinLen: 1,
			expectedTools:  []string{"spec_init"},
			description:    "Should find init tools",
		},
		{
			name:           "code generation",
			query:          "generate code",
			limit:          5,
			expectedMinLen: 1,
			expectedTools:  []string{"generate", "generate_component"},
			description:    "Should find generation tools",
		},
		{
			name:           "workflow",
			query:          "workflow approve",
			limit:          5,
			expectedMinLen: 1,
			expectedTools:  []string{"workflow_approve", "workflow_start"},
			description:    "Should find workflow tools",
		},
		{
			name:           "autonomous loop",
			query:          "loop self-correction",
			limit:          5,
			expectedMinLen: 1,
			expectedTools:  []string{"loop_start", "loop_cancel"},
			description:    "Should find loop tools",
		},
		{
			name:           "archive",
			query:          "archive completed change",
			limit:          3,
			expectedMinLen: 1,
			expectedTools:  []string{"spec_archive"},
			description:    "Should find archive tool",
		},
		{
			name:           "tool discovery",
			query:          "find tools",
			limit:          5,
			expectedMinLen: 2,
			expectedTools:  []string{"tool_find", "tool_describe"},
			description:    "Should find meta tools",
		},
		{
			name:           "dependencies",
			query:          "task dependencies",
			limit:          3,
			expectedMinLen: 1,
			expectedTools:  []string{"registry_deps"},
			description:    "Should find dependency management",
		},
		{
			name:           "synchronization",
			query:          "sync registry",
			limit:          3,
			expectedMinLen: 1,
			expectedTools:  []string{"registry_sync"},
			description:    "Should find sync tools",
		},
		{
			name:           "agent execution",
			query:          "execute task agent",
			limit:          3,
			expectedMinLen: 1,
			expectedTools:  []string{"agent_execute"},
			description:    "Should find agent tools",
		},
		{
			name:           "list operations",
			query:          "list",
			limit:          5,
			expectedMinLen: 2,
			expectedTools:  []string{"spec_list", "registry_list"},
			description:    "Should find list tools",
		},
		{
			name:           "empty query",
			query:          "",
			limit:          5,
			expectedMinLen: 0,
			expectedTools:  nil,
			description:    "Empty query should error",
		},
		{
			name:           "no matches",
			query:          "xyzabc123impossible",
			limit:          5,
			expectedMinLen: 0,
			expectedTools:  nil,
			description:    "Nonsense query should return empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := mcp.NewServer(
				&mcp.Implementation{Name: "test", Version: "1.0.0"},
				nil,
			)
			registry := NewToolRegistry(s)

			registerTestTools(s, registry)

			_ = registry.BuildIndex()

			if tt.query == "" {
				results := registry.Find(tt.query, tt.limit)
				assert.Equal(t, 0, len(results), "Empty query should return no results")
				return
			}

			results := registry.Find(tt.query, tt.limit)

			assert.GreaterOrEqual(t, len(results), tt.expectedMinLen,
				"Expected at least %d results for query '%s', got %d. %s",
				tt.expectedMinLen, tt.query, len(results), tt.description)

			if len(tt.expectedTools) > 0 {
				foundNames := make(map[string]bool)
				for _, r := range results {
					foundNames[r.Name] = true
				}

				foundAny := false
				for _, expected := range tt.expectedTools {
					if foundNames[expected] {
						foundAny = true
						break
					}
				}

				assert.True(t, foundAny,
					"Expected at least one of %v in results for query '%s'. Got: %v. %s",
					tt.expectedTools, tt.query, getToolNames(results), tt.description)
			}

			assert.LessOrEqual(t, len(results), tt.limit,
				"Results should respect limit of %d", tt.limit)
		})
	}
}

func TestToolFindRelevanceScoring(t *testing.T) {
	t.Parallel()

	s := mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "1.0.0"},
		nil,
	)
	registry := NewToolRegistry(s)

	registerTestTools(s, registry)
	_ = registry.BuildIndex()

	results := registry.Find("spec validate", 10)
	require.Greater(t, len(results), 0, "Should find at least one result")

	assert.Equal(t, "spec_validate", results[0].Name,
		"spec_validate should be the top result for 'spec validate' query")
}

func TestToolFindLimitBehavior(t *testing.T) {
	t.Parallel()

	s := mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "1.0.0"},
		nil,
	)
	registry := NewToolRegistry(s)

	registerTestTools(s, registry)
	_ = registry.BuildIndex()

	tests := []struct {
		name  string
		limit int
	}{
		{"limit 1", 1},
		{"limit 3", 3},
		{"limit 10", 10},
		{"limit 100", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := registry.Find("tool", tt.limit)
			assert.LessOrEqual(t, len(results), tt.limit,
				"Results should not exceed limit")
		})
	}
}

func TestToolFindHandler(t *testing.T) {
	t.Parallel()

	s := mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "1.0.0"},
		nil,
	)
	registry := NewToolRegistry(s)

	registerTestTools(s, registry)
	_ = registry.BuildIndex()

	handler := makeToolFindHandler(registry)
	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	tests := []struct {
		name        string
		input       ToolFindInput
		expectError bool
		checkOutput func(t *testing.T, result *mcp.CallToolResult, data any)
	}{
		{
			name:        "valid query",
			input:       ToolFindInput{Query: "spec", Limit: 5},
			expectError: false,
			checkOutput: func(t *testing.T, result *mcp.CallToolResult, data any) {
				require.NotNil(t, result)
				require.Len(t, result.Content, 1)

				textContent, ok := result.Content[0].(*mcp.TextContent)
				require.True(t, ok, "Expected TextContent")

				assert.Contains(t, textContent.Text, "Found")
				assert.Contains(t, textContent.Text, "tools matching 'spec'")
			},
		},
		{
			name:        "empty query",
			input:       ToolFindInput{Query: "", Limit: 5},
			expectError: true,
		},
		{
			name:        "default limit",
			input:       ToolFindInput{Query: "tool"},
			expectError: false,
			checkOutput: func(t *testing.T, result *mcp.CallToolResult, data any) {
				results, ok := data.([]*ToolMeta)
				require.True(t, ok, "Expected []*ToolMeta")
				assert.LessOrEqual(t, len(results), 10, "Default limit should be 10")
			},
		},
		{
			name:        "no matches",
			input:       ToolFindInput{Query: "xyzabc123impossible", Limit: 5},
			expectError: false,
			checkOutput: func(t *testing.T, result *mcp.CallToolResult, data any) {
				textContent := result.Content[0].(*mcp.TextContent)
				assert.Contains(t, textContent.Text, "No tools found")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, data, err := handler(ctx, req, tt.input)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tt.checkOutput != nil {
				tt.checkOutput(t, result, data)
			}
		})
	}
}

func TestToolDescribe(t *testing.T) {
	t.Parallel()

	s := mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "1.0.0"},
		nil,
	)
	registry := NewToolRegistry(s)

	registerTestTools(s, registry)
	_ = registry.BuildIndex()

	tests := []struct {
		name        string
		toolName    string
		expectError bool
		checkMeta   func(t *testing.T, meta *ToolMeta)
	}{
		{
			name:        "spec_init",
			toolName:    "spec_init",
			expectError: false,
			checkMeta: func(t *testing.T, meta *ToolMeta) {
				assert.Equal(t, "spec_init", meta.Name)
				assert.Contains(t, meta.Description, "Initialize openspec folder")
				assert.NotNil(t, meta.InputSchema)
			},
		},
		{
			name:        "spec_validate",
			toolName:    "spec_validate",
			expectError: false,
			checkMeta: func(t *testing.T, meta *ToolMeta) {
				assert.Equal(t, "spec_validate", meta.Name)
				assert.Contains(t, meta.Description, "Validate")
			},
		},
		{
			name:        "registry_list",
			toolName:    "registry_list",
			expectError: false,
			checkMeta: func(t *testing.T, meta *ToolMeta) {
				assert.Equal(t, "registry_list", meta.Name)
				assert.Contains(t, meta.Description, "List tasks")
			},
		},
		{
			name:        "tool_find",
			toolName:    "tool_find",
			expectError: false,
			checkMeta: func(t *testing.T, meta *ToolMeta) {
				assert.Equal(t, "tool_find", meta.Name)
				assert.Contains(t, meta.Description, "Search")
			},
		},
		{
			name:        "nonexistent tool",
			toolName:    "nonexistent_tool",
			expectError: true,
		},
		{
			name:        "empty name",
			toolName:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			meta, err := registry.Describe(tt.toolName)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, meta)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, meta)

			if tt.checkMeta != nil {
				tt.checkMeta(t, meta)
			}
		})
	}
}

func TestToolDescribeAllTools(t *testing.T) {
	t.Parallel()

	s := mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "1.0.0"},
		nil,
	)
	registry := NewToolRegistry(s)

	registerTestTools(s, registry)
	_ = registry.BuildIndex()

	allTools := []string{
		"spec_init", "spec_create", "spec_update", "spec_delete",
		"spec_list", "spec_show", "spec_validate", "spec_archive",
		"registry_list", "registry_next", "registry_update", "registry_sync",
		"registry_deps", "registry_init",
		"workflow_start", "workflow_status", "workflow_approve",
		"loop_start", "loop_get", "loop_set", "loop_cancel",
		"generate", "generate_component", "generate_from_spec",
		"list_archetypes", "agent_execute",
		"tool_find", "tool_describe", "tool_load", "tool_active",
	}

	for _, toolName := range allTools {
		t.Run(toolName, func(t *testing.T) {
			meta, err := registry.Describe(toolName)

			require.NoError(t, err, "Failed to describe tool: %s", toolName)
			require.NotNil(t, meta, "Nil metadata for tool: %s", toolName)

			assert.Equal(t, toolName, meta.Name, "Tool name mismatch")
			assert.NotEmpty(t, meta.Description, "Tool %s has empty description", toolName)
			assert.NotNil(t, meta.InputSchema, "Tool %s has nil input schema", toolName)
		})
	}
}

func TestToolDescribeHandler(t *testing.T) {
	t.Parallel()

	s := mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "1.0.0"},
		nil,
	)
	registry := NewToolRegistry(s)

	registerTestTools(s, registry)
	_ = registry.BuildIndex()

	handler := makeToolDescribeHandler(registry)
	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	tests := []struct {
		name        string
		input       ToolDescribeInput
		expectError bool
		checkOutput func(t *testing.T, result *mcp.CallToolResult, data any)
	}{
		{
			name:        "valid tool",
			input:       ToolDescribeInput{Name: "spec_validate"},
			expectError: false,
			checkOutput: func(t *testing.T, result *mcp.CallToolResult, data any) {
				require.NotNil(t, result)
				require.Len(t, result.Content, 1)

				textContent, ok := result.Content[0].(*mcp.TextContent)
				require.True(t, ok, "Expected TextContent")

				assert.Contains(t, textContent.Text, "spec_validate")
				assert.Contains(t, textContent.Text, "Description:")
				assert.Contains(t, textContent.Text, "Input Schema:")
				assert.Contains(t, textContent.Text, "Active:")

				meta, ok := data.(*ToolMeta)
				require.True(t, ok, "Expected *ToolMeta")
				assert.Equal(t, "spec_validate", meta.Name)
			},
		},
		{
			name:        "empty name",
			input:       ToolDescribeInput{Name: ""},
			expectError: true,
		},
		{
			name:        "nonexistent tool",
			input:       ToolDescribeInput{Name: "nonexistent"},
			expectError: false,
			checkOutput: func(t *testing.T, result *mcp.CallToolResult, data any) {
				textContent := result.Content[0].(*mcp.TextContent)
				assert.Contains(t, textContent.Text, "Error:")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, data, err := handler(ctx, req, tt.input)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tt.checkOutput != nil {
				tt.checkOutput(t, result, data)
			}
		})
	}
}

func TestToolDescribeWithMetadata(t *testing.T) {
	t.Parallel()

	s := mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "1.0.0"},
		nil,
	)
	registry := NewToolRegistry(s)

	toolWithMetadata := ToolMeta{
		Name:        "test_tool",
		Description: "Test tool with metadata",
		InputSchema: map[string]any{"type": "object"},
		Category:    "testing",
		Keywords:    []string{"test", "example", "demo"},
	}

	err := registry.Register(toolWithMetadata, func(s *mcp.Server) {})
	require.NoError(t, err)

	meta, err := registry.Describe("test_tool")
	require.NoError(t, err)
	require.NotNil(t, meta)

	assert.Equal(t, "test_tool", meta.Name)
	assert.Equal(t, "Test tool with metadata", meta.Description)
	assert.Equal(t, "testing", meta.Category)
	assert.Equal(t, []string{"test", "example", "demo"}, meta.Keywords)
}

func TestToolLoad(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		toolsToLoad   []string
		expectError   bool
		checkRegistry func(t *testing.T, registry *ToolRegistry)
	}{
		{
			name:        "load single tool",
			toolsToLoad: []string{"spec_init"},
			expectError: false,
			checkRegistry: func(t *testing.T, registry *ToolRegistry) {
				assert.True(t, registry.IsActive("spec_init"))
				active := registry.Active()
				assert.Contains(t, active, "spec_init")
			},
		},
		{
			name:        "load multiple tools",
			toolsToLoad: []string{"spec_init", "spec_validate", "registry_list"},
			expectError: false,
			checkRegistry: func(t *testing.T, registry *ToolRegistry) {
				assert.True(t, registry.IsActive("spec_init"))
				assert.True(t, registry.IsActive("spec_validate"))
				assert.True(t, registry.IsActive("registry_list"))
				active := registry.Active()
				assert.GreaterOrEqual(t, len(active), 3)
			},
		},
		{
			name:        "load already loaded tool",
			toolsToLoad: []string{"spec_init", "spec_init"},
			expectError: false,
			checkRegistry: func(t *testing.T, registry *ToolRegistry) {
				assert.True(t, registry.IsActive("spec_init"))
			},
		},
		{
			name:        "load nonexistent tool",
			toolsToLoad: []string{"nonexistent_tool"},
			expectError: true,
		},
		{
			name:        "load empty list",
			toolsToLoad: []string{},
			expectError: false,
			checkRegistry: func(t *testing.T, registry *ToolRegistry) {
				active := registry.Active()
				assert.Equal(t, 0, len(active))
			},
		},
		{
			name:        "load all discovery tools",
			toolsToLoad: []string{"tool_find", "tool_describe", "tool_load", "tool_active"},
			expectError: false,
			checkRegistry: func(t *testing.T, registry *ToolRegistry) {
				assert.True(t, registry.IsActive("tool_find"))
				assert.True(t, registry.IsActive("tool_describe"))
				assert.True(t, registry.IsActive("tool_load"))
				assert.True(t, registry.IsActive("tool_active"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := mcp.NewServer(
				&mcp.Implementation{Name: "test", Version: "1.0.0"},
				nil,
			)
			registry := NewToolRegistry(s)

			registerTestTools(s, registry)
			_ = registry.BuildIndex()

			err := registry.Load(tt.toolsToLoad)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tt.checkRegistry != nil {
				tt.checkRegistry(t, registry)
			}
		})
	}
}

func TestToolLoadIdempotent(t *testing.T) {
	t.Parallel()

	s := mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "1.0.0"},
		nil,
	)
	registry := NewToolRegistry(s)

	registerTestTools(s, registry)
	_ = registry.BuildIndex()

	err := registry.Load([]string{"spec_init"})
	require.NoError(t, err)
	assert.True(t, registry.IsActive("spec_init"))

	activeBefore := len(registry.Active())

	err = registry.Load([]string{"spec_init"})
	require.NoError(t, err)

	activeAfter := len(registry.Active())
	assert.Equal(t, activeBefore, activeAfter, "Loading same tool twice should be idempotent")
}

func TestToolLoadHandler(t *testing.T) {
	t.Parallel()

	s := mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "1.0.0"},
		nil,
	)
	registry := NewToolRegistry(s)

	registerTestTools(s, registry)
	_ = registry.BuildIndex()

	handler := makeToolLoadHandler(registry)
	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	tests := []struct {
		name        string
		input       ToolLoadInput
		expectError bool
		checkOutput func(t *testing.T, result *mcp.CallToolResult, data any)
	}{
		{
			name:        "valid load",
			input:       ToolLoadInput{Names: []string{"spec_init", "spec_validate"}},
			expectError: false,
			checkOutput: func(t *testing.T, result *mcp.CallToolResult, data any) {
				require.NotNil(t, result)
				require.Len(t, result.Content, 1)

				textContent, ok := result.Content[0].(*mcp.TextContent)
				require.True(t, ok, "Expected TextContent")

				assert.Contains(t, textContent.Text, "Loaded 2 tool(s)")
				assert.Contains(t, textContent.Text, "spec_init")
				assert.Contains(t, textContent.Text, "spec_validate")
				assert.Contains(t, textContent.Text, "Total active tools:")

				active, ok := data.([]string)
				require.True(t, ok, "Expected []string")
				assert.GreaterOrEqual(t, len(active), 2)
			},
		},
		{
			name:        "empty names",
			input:       ToolLoadInput{Names: []string{}},
			expectError: true,
		},
		{
			name:        "nonexistent tool",
			input:       ToolLoadInput{Names: []string{"nonexistent"}},
			expectError: false,
			checkOutput: func(t *testing.T, result *mcp.CallToolResult, data any) {
				textContent := result.Content[0].(*mcp.TextContent)
				assert.Contains(t, textContent.Text, "Error loading tools:")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, data, err := handler(ctx, req, tt.input)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tt.checkOutput != nil {
				tt.checkOutput(t, result, data)
			}
		})
	}
}

func TestToolLoadConcurrent(t *testing.T) {
	t.Parallel()

	s := mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "1.0.0"},
		nil,
	)
	registry := NewToolRegistry(s)

	registerTestTools(s, registry)
	_ = registry.BuildIndex()

	tools := []string{"spec_init", "spec_validate", "registry_list", "workflow_start"}

	done := make(chan bool, len(tools))

	for _, tool := range tools {
		go func() {
			err := registry.Load([]string{tool})
			assert.NoError(t, err)
			done <- true
		}()
	}

	for range tools {
		<-done
	}

	for _, tool := range tools {
		assert.True(t, registry.IsActive(tool), "Tool %s should be active", tool)
	}

	active := registry.Active()
	assert.GreaterOrEqual(t, len(active), len(tools))
}

func TestToolActive(t *testing.T) {
	t.Parallel()

	s := mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "1.0.0"},
		nil,
	)
	registry := NewToolRegistry(s)

	registerTestTools(s, registry)
	_ = registry.BuildIndex()

	active := registry.Active()
	assert.Equal(t, 0, len(active), "Initially no tools should be active")

	err := registry.Load([]string{"spec_init"})
	require.NoError(t, err)

	active = registry.Active()
	assert.Equal(t, 1, len(active))
	assert.Contains(t, active, "spec_init")

	err = registry.Load([]string{"spec_validate", "registry_list"})
	require.NoError(t, err)

	active = registry.Active()
	assert.Equal(t, 3, len(active))
	assert.Contains(t, active, "spec_init")
	assert.Contains(t, active, "spec_validate")
	assert.Contains(t, active, "registry_list")
}

func TestToolActiveHandler(t *testing.T) {
	t.Parallel()

	s := mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "1.0.0"},
		nil,
	)
	registry := NewToolRegistry(s)

	registerTestTools(s, registry)
	_ = registry.BuildIndex()

	handler := makeToolActiveHandler(registry)
	ctx := context.Background()
	req := &mcp.CallToolRequest{}

	t.Run("no active tools", func(t *testing.T) {
		result, data, err := handler(ctx, req, ToolActiveInput{})
		require.NoError(t, err)
		require.NotNil(t, result)

		textContent := result.Content[0].(*mcp.TextContent)
		assert.Contains(t, textContent.Text, "Currently active tools (0)")
		assert.Contains(t, textContent.Text, "No tools are currently active")

		active, ok := data.([]string)
		require.True(t, ok)
		assert.Equal(t, 0, len(active))
	})

	t.Run("with active tools", func(t *testing.T) {
		err := registry.Load([]string{"spec_init", "spec_validate", "registry_list"})
		require.NoError(t, err)

		result, data, err := handler(ctx, req, ToolActiveInput{})
		require.NoError(t, err)
		require.NotNil(t, result)

		textContent := result.Content[0].(*mcp.TextContent)
		assert.Contains(t, textContent.Text, "Currently active tools (3)")
		assert.Contains(t, textContent.Text, "spec_init")
		assert.Contains(t, textContent.Text, "spec_validate")
		assert.Contains(t, textContent.Text, "registry_list")

		active, ok := data.([]string)
		require.True(t, ok)
		assert.Equal(t, 3, len(active))
		assert.Contains(t, active, "spec_init")
		assert.Contains(t, active, "spec_validate")
		assert.Contains(t, active, "registry_list")
	})
}

func TestToolActiveWithDescriptions(t *testing.T) {
	t.Parallel()

	s := mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "1.0.0"},
		nil,
	)
	registry := NewToolRegistry(s)

	registerTestTools(s, registry)
	_ = registry.BuildIndex()

	err := registry.Load([]string{"spec_init", "tool_find"})
	require.NoError(t, err)

	handler := makeToolActiveHandler(registry)
	result, _, err := handler(context.Background(), &mcp.CallToolRequest{}, ToolActiveInput{})
	require.NoError(t, err)

	textContent := result.Content[0].(*mcp.TextContent)

	assert.Contains(t, textContent.Text, "spec_init")
	assert.Contains(t, textContent.Text, "Initialize openspec folder")

	assert.Contains(t, textContent.Text, "tool_find")
	assert.Contains(t, textContent.Text, "Search for tools")
}

func TestToolActiveEmpty(t *testing.T) {
	t.Parallel()

	s := mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "1.0.0"},
		nil,
	)
	registry := NewToolRegistry(s)

	registerTestTools(s, registry)

	active := registry.Active()
	assert.NotNil(t, active, "Active() should never return nil")
	assert.Equal(t, 0, len(active), "Should return empty slice, not nil")
}

func TestIsActive(t *testing.T) {
	t.Parallel()

	s := mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "1.0.0"},
		nil,
	)
	registry := NewToolRegistry(s)

	registerTestTools(s, registry)
	_ = registry.BuildIndex()

	assert.False(t, registry.IsActive("spec_init"))
	assert.False(t, registry.IsActive("nonexistent"))

	err := registry.Load([]string{"spec_init"})
	require.NoError(t, err)

	assert.True(t, registry.IsActive("spec_init"))
	assert.False(t, registry.IsActive("spec_validate"))
	assert.False(t, registry.IsActive("nonexistent"))
}

func TestToolActiveOrder(t *testing.T) {
	t.Parallel()

	s := mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "1.0.0"},
		nil,
	)
	registry := NewToolRegistry(s)

	registerTestTools(s, registry)
	_ = registry.BuildIndex()

	tools := []string{"spec_init", "spec_validate", "registry_list", "workflow_start"}
	err := registry.Load(tools)
	require.NoError(t, err)

	active := registry.Active()
	assert.Equal(t, len(tools), len(active))

	for _, tool := range tools {
		assert.Contains(t, active, tool)
	}
}

func registerTestTools(s *mcp.Server, registry *ToolRegistry) {
	testTools := []ToolMeta{
		{
			Name:        "spec_init",
			Description: "Initialize openspec folder in a project directory",
			InputSchema: map[string]any{"type": "object"},
		},
		{
			Name:        "spec_create",
			Description: "Create a new spec, change, or task",
			InputSchema: map[string]any{"type": "object"},
		},
		{
			Name:        "spec_update",
			Description: "Update an existing spec, change, or task",
			InputSchema: map[string]any{"type": "object"},
		},
		{
			Name:        "spec_delete",
			Description: "Delete a spec, change, or task",
			InputSchema: map[string]any{"type": "object"},
		},
		{
			Name:        "spec_list",
			Description: "List specs, changes, or tasks",
			InputSchema: map[string]any{"type": "object"},
		},
		{
			Name:        "spec_show",
			Description: "Show detailed content of a spec, change, or task",
			InputSchema: map[string]any{"type": "object"},
		},
		{
			Name:        "spec_validate",
			Description: "Validate OpenSpec files. Type can be 'spec', 'change', or 'all'",
			InputSchema: map[string]any{"type": "object"},
		},
		{
			Name:        "spec_archive",
			Description: "Archive a completed change and optionally merge deltas into specs",
			InputSchema: map[string]any{"type": "object"},
		},
		{
			Name:        "registry_list",
			Description: "List tasks from the OpenSpec registry with optional filters",
			InputSchema: map[string]any{"type": "object"},
		},
		{
			Name:        "registry_next",
			Description: "Get the next recommended task(s) based on priority and dependencies",
			InputSchema: map[string]any{"type": "object"},
		},
		{
			Name:        "registry_update",
			Description: "Update task status, priority, or assignment in the registry",
			InputSchema: map[string]any{"type": "object"},
		},
		{
			Name:        "registry_sync",
			Description: "Synchronize registry from tasks.md files",
			InputSchema: map[string]any{"type": "object"},
		},
		{
			Name:        "registry_deps",
			Description: "Manage task dependencies. Supports cross-change dependencies",
			InputSchema: map[string]any{"type": "object"},
		},
		{
			Name:        "registry_init",
			Description: "Initialize an empty registry.yaml file",
			InputSchema: map[string]any{"type": "object"},
		},
		{
			Name:        "workflow_start",
			Description: "Start a guided workflow with state tracking and wait points",
			InputSchema: map[string]any{"type": "object"},
		},
		{
			Name:        "workflow_status",
			Description: "Check current workflow status and wait points",
			InputSchema: map[string]any{"type": "object"},
		},
		{
			Name:        "workflow_approve",
			Description: "Approve the current wait point and continue workflow",
			InputSchema: map[string]any{"type": "object"},
		},
		{
			Name:        "loop_start",
			Description: "Start autonomous loop with self-correction",
			InputSchema: map[string]any{"type": "object"},
		},
		{
			Name:        "loop_get",
			Description: "Get current loop state",
			InputSchema: map[string]any{"type": "object"},
		},
		{
			Name:        "loop_set",
			Description: "Update loop state (iteration, error, adjustment, status)",
			InputSchema: map[string]any{"type": "object"},
		},
		{
			Name:        "loop_cancel",
			Description: "Cancel running loop",
			InputSchema: map[string]any{"type": "object"},
		},
		{
			Name:        "generate",
			Description: "Generate a new Go project from templates",
			InputSchema: map[string]any{"type": "object"},
		},
		{
			Name:        "generate_component",
			Description: "Generate a component scaffold from a spec file",
			InputSchema: map[string]any{"type": "object"},
		},
		{
			Name:        "generate_from_spec",
			Description: "Generate a complete project from a spec file",
			InputSchema: map[string]any{"type": "object"},
		},
		{
			Name:        "list_archetypes",
			Description: "List available project archetypes",
			InputSchema: map[string]any{"type": "object"},
		},
		{
			Name:        "agent_execute",
			Description: "Execute a task with automatic agent selection based on complexity",
			InputSchema: map[string]any{"type": "object"},
		},
		{
			Name:        "tool_find",
			Description: "Search for tools by query using TF-IDF relevance scoring",
			InputSchema: map[string]any{"type": "object"},
		},
		{
			Name:        "tool_describe",
			Description: "Get detailed information about a specific tool",
			InputSchema: map[string]any{"type": "object"},
		},
		{
			Name:        "tool_load",
			Description: "Load (activate) one or more tools into the active set",
			InputSchema: map[string]any{"type": "object"},
		},
		{
			Name:        "tool_active",
			Description: "List currently active (loaded) tools",
			InputSchema: map[string]any{"type": "object"},
		},
	}

	for _, meta := range testTools {
		_ = registry.Register(meta, func(s *mcp.Server) {
			// No-op registration for tests
		})
	}
}

func getToolNames(metas []*ToolMeta) []string {
	names := make([]string, len(metas))
	for i, m := range metas {
		names[i] = m.Name
	}
	return names
}
