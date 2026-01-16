package metrics

//nolint:gosec // test file with necessary file operations

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/victorzhuk/go-ent/internal/execution"
)

func TestCollector_StartEndExecution(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "metrics.json")
	store, err := NewStore(path, 24*time.Hour)
	require.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, store.Close())
	})

	c := NewCollector(store)
	t.Cleanup(func() {
		assert.NoError(t, c.Shutdown(context.Background()))
	})

	ctx := context.Background()
	sessionID := c.StartExecution(ctx, "test_tool")

	ctx = ContextWithSession(ctx, sessionID)

	result := &execution.Result{
		Success:   true,
		TokensIn:  100,
		TokensOut: 50,
		Duration:  100 * time.Millisecond,
	}

	err = c.EndExecution(ctx, result)
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	metrics := store.GetAll()
	assert.Len(t, metrics, 1)
	assert.Equal(t, sessionID, metrics[0].SessionID)
	assert.Equal(t, "test_tool", metrics[0].ToolName)
	assert.True(t, metrics[0].Success)
	assert.Equal(t, 100, metrics[0].TokensIn)
	assert.Equal(t, 50, metrics[0].TokensOut)
}

func TestCollector_SessionIDGeneration(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "metrics.json")
	store, err := NewStore(path, 24*time.Hour)
	require.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, store.Close())
	})

	c := NewCollector(store)
	t.Cleanup(func() {
		assert.NoError(t, c.Shutdown(context.Background()))
	})

	ctx := context.Background()
	sessionID := c.StartExecution(ctx, "tool1")
	sessionID2 := c.StartExecution(ctx, "tool2")

	assert.NotEmpty(t, sessionID)
	assert.NotEmpty(t, sessionID2)
	assert.NotEqual(t, sessionID, sessionID2)

	_, err = uuid.Parse(sessionID)
	assert.NoError(t, err)

	_, err = uuid.Parse(sessionID2)
	assert.NoError(t, err)
}

func TestCollector_SessionFromContext(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "metrics.json")
	store, err := NewStore(path, 24*time.Hour)
	require.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, store.Close())
	})

	c := NewCollector(store)
	t.Cleanup(func() {
		assert.NoError(t, c.Shutdown(context.Background()))
	})

	ctx := context.Background()
	sessionID := "test-session-123"

	ctx = ContextWithSession(ctx, sessionID)

	retrieved, ok := SessionFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, sessionID, retrieved)

	_, ok = SessionFromContext(context.Background())
	assert.False(t, ok)
}

func TestCollector_AsyncWrite(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "metrics.json")
	store, err := NewStore(path, 24*time.Hour)
	require.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, store.Close())
	})

	c := NewCollector(store)
	t.Cleanup(func() {
		assert.NoError(t, c.Shutdown(context.Background()))
	})

	ctx := context.Background()

	for i := 0; i < 50; i++ {
		sessionID := c.StartExecution(ctx, "async_tool")
		ctx = ContextWithSession(ctx, sessionID)

		result := &execution.Result{
			Success:   true,
			TokensIn:  i,
			TokensOut: i * 2,
			Duration:  time.Duration(i) * time.Millisecond,
		}

		err = c.EndExecution(ctx, result)
		assert.NoError(t, err)
	}

	time.Sleep(200 * time.Millisecond)

	metrics := store.GetAll()
	assert.Len(t, metrics, 50)
}

func TestCollector_ConcurrentExecutions(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "metrics.json")
	store, err := NewStore(path, 24*time.Hour)
	require.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, store.Close())
	})

	c := NewCollector(store)
	t.Cleanup(func() {
		assert.NoError(t, c.Shutdown(context.Background()))
	})

	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(idx int) {
			defer func() { done <- true }()

			for j := 0; j < 5; j++ {
				ctx := context.Background()
				sessionID := c.StartExecution(ctx, "concurrent_tool")
				ctx = ContextWithSession(ctx, sessionID)

				result := &execution.Result{
					Success:   true,
					TokensIn:  idx*10 + j,
					TokensOut: (idx*10 + j) * 2,
					Duration:  time.Duration(j) * time.Millisecond,
				}

				err := c.EndExecution(ctx, result)
				assert.NoError(t, err)
			}
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	time.Sleep(200 * time.Millisecond)

	metrics := store.GetAll()
	assert.Len(t, metrics, 50)
}

func TestCollector_ShutdownWaitsForMetrics(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "metrics.json")
	store, err := NewStore(path, 24*time.Hour)
	require.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, store.Close())
	})

	c := NewCollector(store)

	ctx := context.Background()

	for i := 0; i < 20; i++ {
		sessionID := c.StartExecution(ctx, "shutdown_tool")
		ctx = ContextWithSession(ctx, sessionID)

		result := &execution.Result{
			Success:   true,
			TokensIn:  i,
			TokensOut: i * 2,
			Duration:  time.Duration(i) * time.Millisecond,
		}

		err = c.EndExecution(ctx, result)
		assert.NoError(t, err)
	}

	err = c.Shutdown(context.Background())
	require.NoError(t, err)

	metrics := store.GetAll()
	assert.Len(t, metrics, 20)
}

func TestCollector_ChannelOverflow(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "metrics.json")
	store, err := NewStore(path, 24*time.Hour)
	require.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, store.Close())
	})

	c := NewCollector(store)
	t.Cleanup(func() {
		assert.NoError(t, c.Shutdown(context.Background()))
	})

	ctx := context.Background()

	for i := 0; i < 150; i++ {
		sessionID := c.StartExecution(ctx, "overflow_tool")
		ctx = ContextWithSession(ctx, sessionID)

		result := &execution.Result{
			Success:   true,
			TokensIn:  i,
			TokensOut: i * 2,
			Duration:  time.Duration(i) * time.Millisecond,
		}

		_ = c.EndExecution(ctx, result)
	}

	time.Sleep(500 * time.Millisecond)

	metrics := store.GetAll()
	assert.LessOrEqual(t, len(metrics), 150)
}

func TestCollector_EndExecution_NoSession(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "metrics.json")
	store, err := NewStore(path, 24*time.Hour)
	require.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, store.Close())
	})

	c := NewCollector(store)
	t.Cleanup(func() {
		assert.NoError(t, c.Shutdown(context.Background()))
	})

	ctx := context.Background()
	result := &execution.Result{
		Success:  true,
		TokensIn: 100,
	}

	err = c.EndExecution(ctx, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no session")
}

func TestCollector_EndExecution_SessionNotStarted(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "metrics.json")
	store, err := NewStore(path, 24*time.Hour)
	require.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, store.Close())
	})

	c := NewCollector(store)
	t.Cleanup(func() {
		assert.NoError(t, c.Shutdown(context.Background()))
	})

	ctx := ContextWithSession(context.Background(), "non-existent-session")
	result := &execution.Result{
		Success:  true,
		TokensIn: 100,
	}

	err = c.EndExecution(ctx, result)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not started")
}

func TestCollector_FailedExecution(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "metrics.json")
	store, err := NewStore(path, 24*time.Hour)
	require.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, store.Close())
	})

	c := NewCollector(store)
	t.Cleanup(func() {
		assert.NoError(t, c.Shutdown(context.Background()))
	})

	ctx := context.Background()
	sessionID := c.StartExecution(ctx, "failing_tool")
	ctx = ContextWithSession(ctx, sessionID)

	result := &execution.Result{
		Success:   false,
		Error:     "something went wrong",
		TokensIn:  100,
		TokensOut: 0,
		Duration:  50 * time.Millisecond,
	}

	err = c.EndExecution(ctx, result)
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	metrics := store.GetAll()
	assert.Len(t, metrics, 1)
	assert.False(t, metrics[0].Success)
	assert.Equal(t, "something went wrong", metrics[0].ErrorMsg)
}

func BenchmarkCollector_StartEnd(b *testing.B) {
	tmpDir := b.TempDir()
	path := filepath.Join(tmpDir, "metrics.json")
	store, err := NewStore(path, 24*time.Hour)
	if err != nil {
		b.Fatal(err)
	}
	defer func() { _ = store.Close() }()

	c := NewCollector(store)
	defer func() { _ = c.Shutdown(context.Background()) }()

	ctx := context.Background()
	result := &execution.Result{
		Success:   true,
		TokensIn:  100,
		TokensOut: 50,
		Duration:  100 * time.Millisecond,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sessionID := c.StartExecution(ctx, "benchmark_tool")
		ctx = ContextWithSession(ctx, sessionID)

		if err := c.EndExecution(ctx, result); err != nil {
			b.Fatal(err)
		}
	}
}
