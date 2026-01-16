package tools

//nolint:gosec // test file with necessary file operations

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConcurrentRegistration tests concurrent tool registration
func TestConcurrentRegistration(t *testing.T) {
	t.Parallel()

	s := mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "1.0.0"},
		nil,
	)
	registry := NewToolRegistry(s)

	const numGoroutines = 50
	const toolsPerGoroutine = 10

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*toolsPerGoroutine)

	// Concurrently register tools
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < toolsPerGoroutine; j++ {
				meta := ToolMeta{
					Name:        fmt.Sprintf("tool_%d_%d", id, j),
					Description: fmt.Sprintf("Test tool %d/%d", id, j),
					InputSchema: map[string]any{"type": "object"},
				}
				if err := registry.Register(meta, func(s *mcp.Server) {}); err != nil {
					errors <- err
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Registration error: %v", err)
	}

	// Verify all tools were registered
	allTools := registry.All()
	expectedCount := numGoroutines * toolsPerGoroutine
	assert.Equal(t, expectedCount, len(allTools),
		"Expected %d tools to be registered", expectedCount)
}

// TestConcurrentFind tests concurrent search operations
func TestConcurrentFind(t *testing.T) {
	t.Parallel()

	s := mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "1.0.0"},
		nil,
	)
	registry := NewToolRegistry(s)

	registerTestTools(s, registry)
	require.NoError(t, registry.BuildIndex())

	const numGoroutines = 100
	const queriesPerGoroutine = 50

	queries := []string{
		"spec validate",
		"registry list",
		"workflow start",
		"loop cancel",
		"generate code",
		"tool find",
	}

	var wg sync.WaitGroup
	var totalSearches atomic.Int64

	// Concurrently search for tools
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < queriesPerGoroutine; j++ {
				query := queries[j%len(queries)]
				results := registry.Find(query, 10)
				totalSearches.Add(1)

				// Verify results are not corrupted
				for _, result := range results {
					assert.NotNil(t, result)
					assert.NotEmpty(t, result.Name)
					assert.NotEmpty(t, result.Description)
				}
			}
		}(i)
	}

	wg.Wait()

	expectedSearches := int64(numGoroutines * queriesPerGoroutine)
	assert.Equal(t, expectedSearches, totalSearches.Load(),
		"All searches should complete")
}

// TestConcurrentLoad tests concurrent tool loading
func TestConcurrentLoad(t *testing.T) {
	t.Parallel()

	s := mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "1.0.0"},
		nil,
	)
	registry := NewToolRegistry(s)

	registerTestTools(s, registry)
	require.NoError(t, registry.BuildIndex())

	const numGoroutines = 50

	tools := [][]string{
		{"spec_init", "spec_validate"},
		{"registry_list", "registry_next"},
		{"workflow_start", "workflow_status"},
		{"loop_start", "loop_cancel"},
		{"tool_find", "tool_describe"},
	}

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	// Concurrently load different tool sets
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			toolSet := tools[id%len(tools)]
			if err := registry.Load(toolSet); err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Load error: %v", err)
	}

	// Verify tools are loaded
	active := registry.Active()
	assert.GreaterOrEqual(t, len(active), 10,
		"At least 10 tools should be active after concurrent loading")
}

// TestConcurrentDescribe tests concurrent describe operations
func TestConcurrentDescribe(t *testing.T) {
	t.Parallel()

	s := mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "1.0.0"},
		nil,
	)
	registry := NewToolRegistry(s)

	registerTestTools(s, registry)

	const numGoroutines = 100
	const describesPerGoroutine = 50

	toolNames := []string{
		"spec_init", "spec_validate", "registry_list",
		"workflow_start", "loop_start", "tool_find",
	}

	var wg sync.WaitGroup
	var successCount atomic.Int64

	// Concurrently describe tools
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < describesPerGoroutine; j++ {
				toolName := toolNames[j%len(toolNames)]
				meta, err := registry.Describe(toolName)
				if err == nil && meta != nil {
					successCount.Add(1)
					// Verify metadata integrity
					assert.Equal(t, toolName, meta.Name)
					assert.NotEmpty(t, meta.Description)
				}
			}
		}()
	}

	wg.Wait()

	expectedSuccesses := int64(numGoroutines * describesPerGoroutine)
	assert.Equal(t, expectedSuccesses, successCount.Load(),
		"All describe operations should succeed")
}

// TestConcurrentActive tests concurrent active tool queries
func TestConcurrentActive(t *testing.T) {
	t.Parallel()

	s := mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "1.0.0"},
		nil,
	)
	registry := NewToolRegistry(s)

	registerTestTools(s, registry)
	require.NoError(t, registry.BuildIndex())

	// Pre-load some tools
	require.NoError(t, registry.Load([]string{"spec_init", "spec_validate", "registry_list"}))

	const numGoroutines = 100
	const queriesPerGoroutine = 50

	var wg sync.WaitGroup

	// Concurrently query active tools
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < queriesPerGoroutine; j++ {
				active := registry.Active()
				// Verify consistency
				assert.GreaterOrEqual(t, len(active), 3,
					"At least 3 tools should be active")
				assert.NotNil(t, active)
			}
		}()
	}

	wg.Wait()
}

// TestConcurrentIsActive tests concurrent IsActive checks
func TestConcurrentIsActive(t *testing.T) {
	t.Parallel()

	s := mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "1.0.0"},
		nil,
	)
	registry := NewToolRegistry(s)

	registerTestTools(s, registry)

	// Pre-load some tools
	require.NoError(t, registry.Load([]string{"spec_init", "spec_validate"}))

	const numGoroutines = 100
	const checksPerGoroutine = 100

	var wg sync.WaitGroup

	// Concurrently check tool activation status
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < checksPerGoroutine; j++ {
				// Check loaded tools
				assert.True(t, registry.IsActive("spec_init"))
				assert.True(t, registry.IsActive("spec_validate"))

				// Check unloaded tools
				assert.False(t, registry.IsActive("workflow_start"))
			}
		}()
	}

	wg.Wait()
}

// TestConcurrentMixedOperations tests mixed concurrent operations
func TestConcurrentMixedOperations(t *testing.T) {
	t.Parallel()

	s := mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "1.0.0"},
		nil,
	)
	registry := NewToolRegistry(s)

	registerTestTools(s, registry)
	require.NoError(t, registry.BuildIndex())

	const numGoroutines = 20
	const opsPerGoroutine = 50

	var wg sync.WaitGroup

	// Mix of all operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				switch j % 5 {
				case 0: // Find
					results := registry.Find("spec", 5)
					assert.NotNil(t, results)

				case 1: // Describe
					meta, err := registry.Describe("spec_validate")
					assert.NoError(t, err)
					assert.NotNil(t, meta)

				case 2: // Load
					err := registry.Load([]string{"spec_init"})
					assert.NoError(t, err)

				case 3: // Active
					active := registry.Active()
					assert.NotNil(t, active)

				case 4: // IsActive
					_ = registry.IsActive("spec_init")
				}
			}
		}(i)
	}

	wg.Wait()
}

// TestBuildIndexOnce tests that BuildIndex should only be called once during initialization
// It is NOT thread-safe to call BuildIndex concurrently - this is by design.
// BuildIndex is meant to be called once during server initialization.
func TestBuildIndexOnce(t *testing.T) {
	t.Parallel()

	s := mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "1.0.0"},
		nil,
	)
	registry := NewToolRegistry(s)

	registerTestTools(s, registry)

	// Build index once (normal usage)
	err := registry.BuildIndex()
	require.NoError(t, err)

	// Verify index is functional
	results := registry.Find("spec validate", 5)
	assert.Greater(t, len(results), 0, "Index should be functional after build")

	// Building again should work (idempotent)
	err = registry.BuildIndex()
	require.NoError(t, err)

	// Verify index still works
	results = registry.Find("spec validate", 5)
	assert.Greater(t, len(results), 0, "Index should still work after rebuild")
}

// TestRaceDetection runs with -race flag to detect data races
func TestRaceDetection(t *testing.T) {
	t.Parallel()

	s := mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "1.0.0"},
		nil,
	)
	registry := NewToolRegistry(s)

	// Register tools
	for i := 0; i < 30; i++ {
		meta := ToolMeta{
			Name:        fmt.Sprintf("tool_%d", i),
			Description: fmt.Sprintf("Test tool %d", i),
			InputSchema: map[string]any{"type": "object"},
		}
		require.NoError(t, registry.Register(meta, func(s *mcp.Server) {}))
	}

	require.NoError(t, registry.BuildIndex())

	const numGoroutines = 50
	var wg sync.WaitGroup

	// Stress test with all operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Register
			meta := ToolMeta{
				Name:        fmt.Sprintf("race_tool_%d", id),
				Description: "Race detection test tool",
				InputSchema: map[string]any{"type": "object"},
			}
			_ = registry.Register(meta, func(s *mcp.Server) {})

			// Find
			_ = registry.Find(fmt.Sprintf("tool_%d", id%10), 5)

			// Describe
			_, _ = registry.Describe(fmt.Sprintf("tool_%d", id%10))

			// Load
			_ = registry.Load([]string{fmt.Sprintf("tool_%d", id%10)})

			// Active
			_ = registry.Active()

			// IsActive
			_ = registry.IsActive(fmt.Sprintf("tool_%d", id%10))

			// All
			_ = registry.All()

			// Count
			_ = registry.Count()
		}(i)
	}

	wg.Wait()
}

// TestConcurrentAll tests concurrent All() operations
func TestConcurrentAll(t *testing.T) {
	t.Parallel()

	s := mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "1.0.0"},
		nil,
	)
	registry := NewToolRegistry(s)

	registerTestTools(s, registry)

	const numGoroutines = 100
	const queriesPerGoroutine = 50

	var wg sync.WaitGroup

	// Concurrently get all tools
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < queriesPerGoroutine; j++ {
				allTools := registry.All()
				assert.NotNil(t, allTools)
				assert.Equal(t, 30, len(allTools),
					"Should always return 30 test tools")
			}
		}()
	}

	wg.Wait()
}

// TestConcurrentCount tests concurrent Count() operations
func TestConcurrentCount(t *testing.T) {
	t.Parallel()

	s := mcp.NewServer(
		&mcp.Implementation{Name: "test", Version: "1.0.0"},
		nil,
	)
	registry := NewToolRegistry(s)

	registerTestTools(s, registry)

	const numGoroutines = 100
	const queriesPerGoroutine = 50

	var wg sync.WaitGroup

	// Concurrently get tool count
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < queriesPerGoroutine; j++ {
				count := registry.Count()
				assert.Equal(t, 30, count,
					"Should always return 30 test tools")
			}
		}()
	}

	wg.Wait()
}
