package background

//nolint:gosec // test file with necessary file operations

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewRegistry(t *testing.T) {
	t.Run("creates registry with manager", func(t *testing.T) {
		mgr := NewManager(nil, DefaultConfig())
		reg := NewRegistry(mgr)

		assert.NotNil(t, reg)
		assert.Same(t, mgr, reg.manager)
	})

	t.Run("creates registry with nil manager", func(t *testing.T) {
		reg := NewRegistry(nil)

		assert.NotNil(t, reg)
		assert.Nil(t, reg.manager)
	})
}

func TestRegistry_Get(t *testing.T) {
	t.Run("returns existing agent", func(t *testing.T) {
		mgr := NewManager(nil, DefaultConfig())
		agent, _ := NewAgent("test-id", "dev", "haiku", "test task")

		mgr.mu.Lock()
		mgr.agents["test-id"] = agent
		mgr.mu.Unlock()

		reg := NewRegistry(mgr)

		result, err := reg.Get("test-id")

		assert.NoError(t, err)
		assert.Equal(t, "test-id", result.ID)
	})

	t.Run("returns error for non-existent agent", func(t *testing.T) {
		mgr := NewManager(nil, DefaultConfig())
		reg := NewRegistry(mgr)

		_, err := reg.Get("non-existent")

		assert.Error(t, err)
		assert.Equal(t, ErrAgentNotFound, err)
	})

	t.Run("returns error when manager is nil", func(t *testing.T) {
		reg := NewRegistry(nil)

		_, err := reg.Get("test-id")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "manager not initialized")
	})
}

func TestRegistry_List(t *testing.T) {
	t.Run("lists all agents with empty status", func(t *testing.T) {
		mgr := NewManager(nil, DefaultConfig())
		agent1, _ := NewAgent("id1", "dev", "haiku", "task1")
		agent2, _ := NewAgent("id2", "dev", "haiku", "task2")

		mgr.mu.Lock()
		mgr.agents["id1"] = agent1
		mgr.agents["id2"] = agent2
		mgr.mu.Unlock()

		reg := NewRegistry(mgr)

		result := reg.List("")

		assert.Len(t, result, 2)
	})

	t.Run("filters agents by status", func(t *testing.T) {
		mgr := NewManager(nil, DefaultConfig())
		agent1, _ := NewAgent("id1", "dev", "haiku", "task1")
		agent2, _ := NewAgent("id2", "dev", "haiku", "task2")

		agent1.Start()
		agent2.Complete("done")

		mgr.mu.Lock()
		mgr.agents["id1"] = agent1
		mgr.agents["id2"] = agent2
		mgr.mu.Unlock()

		reg := NewRegistry(mgr)

		running := reg.List(StatusRunning)
		completed := reg.List(StatusCompleted)

		assert.Len(t, running, 1)
		assert.Len(t, completed, 1)
		assert.Equal(t, "id1", running[0].ID)
		assert.Equal(t, "id2", completed[0].ID)
	})

	t.Run("returns nil when manager is nil", func(t *testing.T) {
		reg := NewRegistry(nil)

		result := reg.List(StatusRunning)

		assert.Nil(t, result)
	})

	t.Run("returns empty slice for no agents", func(t *testing.T) {
		mgr := NewManager(nil, DefaultConfig())
		reg := NewRegistry(mgr)

		result := reg.List(StatusRunning)

		assert.Len(t, result, 0)
	})
}

func TestRegistry_ListAll(t *testing.T) {
	t.Run("lists all agents regardless of status", func(t *testing.T) {
		mgr := NewManager(nil, DefaultConfig())
		agent1, _ := NewAgent("id1", "dev", "haiku", "task1")
		agent2, _ := NewAgent("id2", "dev", "haiku", "task2")
		agent3, _ := NewAgent("id3", "dev", "haiku", "task3")

		agent1.Start()
		agent2.Complete("done")
		agent3.Fail(assert.AnError)

		mgr.mu.Lock()
		mgr.agents["id1"] = agent1
		mgr.agents["id2"] = agent2
		mgr.agents["id3"] = agent3
		mgr.mu.Unlock()

		reg := NewRegistry(mgr)

		result := reg.ListAll()

		assert.Len(t, result, 3)
	})

	t.Run("returns nil when manager is nil", func(t *testing.T) {
		reg := NewRegistry(nil)

		result := reg.ListAll()

		assert.Nil(t, result)
	})
}

func TestRegistry_Count(t *testing.T) {
	t.Run("returns total count", func(t *testing.T) {
		mgr := NewManager(nil, DefaultConfig())
		agent1, _ := NewAgent("id1", "dev", "haiku", "task1")
		agent2, _ := NewAgent("id2", "dev", "haiku", "task2")

		mgr.mu.Lock()
		mgr.agents["id1"] = agent1
		mgr.agents["id2"] = agent2
		mgr.mu.Unlock()

		reg := NewRegistry(mgr)

		result := reg.Count()

		assert.Equal(t, 2, result)
	})

	t.Run("returns zero when manager is nil", func(t *testing.T) {
		reg := NewRegistry(nil)

		result := reg.Count()

		assert.Equal(t, 0, result)
	})

	t.Run("returns zero for no agents", func(t *testing.T) {
		mgr := NewManager(nil, DefaultConfig())
		reg := NewRegistry(mgr)

		result := reg.Count()

		assert.Equal(t, 0, result)
	})
}

func TestRegistry_CountByStatus(t *testing.T) {
	t.Run("counts agents by status", func(t *testing.T) {
		mgr := NewManager(nil, DefaultConfig())
		agent1, _ := NewAgent("id1", "dev", "haiku", "task1")
		agent2, _ := NewAgent("id2", "dev", "haiku", "task2")
		agent3, _ := NewAgent("id3", "dev", "haiku", "task3")

		agent1.Start()
		agent2.Start()
		agent3.Complete("done")

		mgr.mu.Lock()
		mgr.agents["id1"] = agent1
		mgr.agents["id2"] = agent2
		mgr.agents["id3"] = agent3
		mgr.mu.Unlock()

		reg := NewRegistry(mgr)

		running := reg.CountByStatus(StatusRunning)
		completed := reg.CountByStatus(StatusCompleted)

		assert.Equal(t, 2, running)
		assert.Equal(t, 1, completed)
	})

	t.Run("returns zero when manager is nil", func(t *testing.T) {
		reg := NewRegistry(nil)

		result := reg.CountByStatus(StatusRunning)

		assert.Equal(t, 0, result)
	})

	t.Run("returns zero for no matching status", func(t *testing.T) {
		mgr := NewManager(nil, DefaultConfig())
		agent1, _ := NewAgent("id1", "dev", "haiku", "task1")

		mgr.mu.Lock()
		mgr.agents["id1"] = agent1
		mgr.mu.Unlock()

		reg := NewRegistry(mgr)

		result := reg.CountByStatus(StatusCompleted)

		assert.Equal(t, 0, result)
	})
}

func TestRegistry_GetStats(t *testing.T) {
	t.Run("returns statistics for all statuses", func(t *testing.T) {
		mgr := NewManager(nil, DefaultConfig())
		agent1, _ := NewAgent("id1", "dev", "haiku", "task1")
		agent2, _ := NewAgent("id2", "dev", "haiku", "task2")
		agent3, _ := NewAgent("id3", "dev", "haiku", "task3")
		agent4, _ := NewAgent("id4", "dev", "haiku", "task4")
		agent5, _ := NewAgent("id5", "dev", "haiku", "task5")

		agent1.Start()
		agent2.Start()
		agent3.Complete("done")
		agent4.Fail(assert.AnError)
		agent5.Kill()

		mgr.mu.Lock()
		mgr.agents["id1"] = agent1
		mgr.agents["id2"] = agent2
		mgr.agents["id3"] = agent3
		mgr.agents["id4"] = agent4
		mgr.agents["id5"] = agent5
		mgr.mu.Unlock()

		reg := NewRegistry(mgr)

		stats := reg.GetStats()

		assert.Equal(t, 0, stats.Pending)
		assert.Equal(t, 2, stats.Running)
		assert.Equal(t, 1, stats.Completed)
		assert.Equal(t, 1, stats.Failed)
		assert.Equal(t, 1, stats.Killed)
		assert.Equal(t, 5, stats.Total)
	})

	t.Run("returns empty stats when manager is nil", func(t *testing.T) {
		reg := NewRegistry(nil)

		stats := reg.GetStats()

		assert.Equal(t, RegistryStats{}, stats)
	})

	t.Run("returns zeros for empty manager", func(t *testing.T) {
		mgr := NewManager(nil, DefaultConfig())
		reg := NewRegistry(mgr)

		stats := reg.GetStats()

		assert.Equal(t, 0, stats.Pending)
		assert.Equal(t, 0, stats.Running)
		assert.Equal(t, 0, stats.Completed)
		assert.Equal(t, 0, stats.Failed)
		assert.Equal(t, 0, stats.Killed)
		assert.Equal(t, 0, stats.Total)
	})
}

func TestRegistryStats_Struct(t *testing.T) {
	t.Run("creates valid stats struct", func(t *testing.T) {
		stats := RegistryStats{
			Pending:   1,
			Running:   2,
			Completed: 3,
			Failed:    4,
			Killed:    5,
			Total:     15,
		}

		assert.Equal(t, 1, stats.Pending)
		assert.Equal(t, 2, stats.Running)
		assert.Equal(t, 3, stats.Completed)
		assert.Equal(t, 4, stats.Failed)
		assert.Equal(t, 5, stats.Killed)
		assert.Equal(t, 15, stats.Total)
	})
}

func TestRegistry_ThreadSafety(t *testing.T) {
	t.Run("concurrent reads are safe", func(t *testing.T) {
		mgr := NewManager(nil, DefaultConfig())

		agent1, _ := NewAgent("id1", "dev", "haiku", "task1")
		agent2, _ := NewAgent("id2", "dev", "haiku", "task2")
		agent1.Start()
		agent2.Complete("done")

		mgr.mu.Lock()
		mgr.agents["id1"] = agent1
		mgr.agents["id2"] = agent2
		mgr.mu.Unlock()

		reg := NewRegistry(mgr)

		done := make(chan bool)
		iterations := 100

		for i := 0; i < iterations; i++ {
			go func() {
				_, _ = reg.Get("id1")
				_ = reg.ListAll()
				_ = reg.Count()
				_ = reg.CountByStatus(StatusRunning)
				_ = reg.GetStats()
				done <- true
			}()
		}

		for i := 0; i < iterations; i++ {
			<-done
		}

		assert.Equal(t, 2, reg.Count())
	})
}

func TestRegistry_IntegrationWithManager(t *testing.T) {
	t.Run("delegates correctly to manager methods", func(t *testing.T) {
		mgr := NewManager(nil, DefaultConfig())
		reg := NewRegistry(mgr)

		agent1, _ := mgr.Spawn(context.TODO(), "task1", SpawnOpts{})
		agent2, _ := mgr.Spawn(context.TODO(), "task2", SpawnOpts{})

		agent1.Start()
		time.Sleep(10 * time.Millisecond)
		agent1.Complete("done")
		agent2.Start()

		assert.Equal(t, 2, reg.Count())
		assert.Equal(t, 2, len(reg.ListAll()))
		assert.Equal(t, 1, reg.CountByStatus(StatusCompleted))
		assert.Equal(t, 1, reg.CountByStatus(StatusRunning))

		retrieved, err := reg.Get(agent1.ID)
		assert.NoError(t, err)
		assert.Equal(t, agent1.ID, retrieved.ID)

		stats := reg.GetStats()
		assert.Equal(t, 2, stats.Total)
	})
}
