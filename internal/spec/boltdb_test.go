package spec

//nolint:gosec // test file with necessary file operations

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupBoltStore(t *testing.T) *BoltStore {
	dbPath := t.TempDir() + "/test.db"

	store, err := NewBoltStore(dbPath)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = store.Close()
	})

	return store
}

func createTestTask(taskNum string, changeID string, status RegistryTaskStatus, priority TaskPriority) *RegistryTask {
	return &RegistryTask{
		ID: TaskID{
			ChangeID: changeID,
			TaskNum:  taskNum,
		},
		Content:    "Test task " + taskNum,
		Status:     status,
		Priority:   priority,
		SourceLine: 1,
		SyncedAt:   time.Now(),
	}
}

func TestBoltStore_GetTask(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func(*testing.T, *BoltStore) TaskID
		wantErr  bool
		errMsg   string
		checkNil bool
	}{
		{
			name: "found",
			setup: func(t *testing.T, s *BoltStore) TaskID {
				task := createTestTask("1", "change-1", RegStatusPending, PriorityMedium)
				err := s.UpdateTask(task)
				require.NoError(t, err)
				return task.ID
			},
			wantErr: false,
		},
		{
			name: "not found",
			setup: func(t *testing.T, s *BoltStore) TaskID {
				return TaskID{ChangeID: "change-1", TaskNum: "999"}
			},
			wantErr:  true,
			errMsg:   "task not found",
			checkNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			store := setupBoltStore(t)
			taskID := tt.setup(t, store)

			task, err := store.GetTask(taskID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				if tt.checkNil {
					assert.Nil(t, task)
				} else {
					assert.NotNil(t, task)
					assert.True(t, task.ID.IsZero())
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, task)
				assert.Equal(t, taskID, task.ID)
			}
		})
	}
}

func TestBoltStore_UpdateTask(t *testing.T) {
	t.Parallel()

	t.Run("creates new task", func(t *testing.T) {
		t.Parallel()

		store := setupBoltStore(t)
		task := createTestTask("1", "change-1", RegStatusPending, PriorityMedium)

		err := store.UpdateTask(task)
		assert.NoError(t, err)

		retrieved, err := store.GetTask(task.ID)
		assert.NoError(t, err)
		assert.Equal(t, task.ID, retrieved.ID)
		assert.Equal(t, task.Content, retrieved.Content)
		assert.Equal(t, task.Status, retrieved.Status)
		assert.Equal(t, task.Priority, retrieved.Priority)
	})

	t.Run("updates existing task", func(t *testing.T) {
		t.Parallel()

		store := setupBoltStore(t)
		task := createTestTask("1", "change-1", RegStatusPending, PriorityMedium)

		err := store.UpdateTask(task)
		require.NoError(t, err)

		task.Status = RegStatusCompleted
		task.Priority = PriorityHigh
		task.Notes = "Updated notes"

		err = store.UpdateTask(task)
		assert.NoError(t, err)

		retrieved, err := store.GetTask(task.ID)
		assert.NoError(t, err)
		assert.Equal(t, RegStatusCompleted, retrieved.Status)
		assert.Equal(t, PriorityHigh, retrieved.Priority)
		assert.Equal(t, "Updated notes", retrieved.Notes)
	})

	t.Run("updates change summary", func(t *testing.T) {
		t.Parallel()

		store := setupBoltStore(t)

		task1 := createTestTask("1", "change-1", RegStatusPending, PriorityMedium)
		task2 := createTestTask("2", "change-1", RegStatusPending, PriorityLow)

		err := store.UpdateTask(task1)
		require.NoError(t, err)

		summary, err := store.GetChangeSummary("change-1")
		require.NoError(t, err)
		assert.Equal(t, 1, summary.Total)

		err = store.UpdateTask(task2)
		require.NoError(t, err)

		summary, err = store.GetChangeSummary("change-1")
		assert.NoError(t, err)
		assert.Equal(t, 2, summary.Total)
	})
}

func TestBoltStore_ListTasks(t *testing.T) {
	t.Parallel()

	t.Run("empty store", func(t *testing.T) {
		t.Parallel()

		store := setupBoltStore(t)

		tasks, err := store.ListTasks(TaskFilter{})
		assert.NoError(t, err)
		assert.Empty(t, tasks)
	})

	t.Run("all tasks", func(t *testing.T) {
		t.Parallel()

		store := setupBoltStore(t)

		task1 := createTestTask("1", "change-1", RegStatusPending, PriorityMedium)
		task2 := createTestTask("2", "change-2", RegStatusCompleted, PriorityHigh)
		task3 := createTestTask("3", "change-1", RegStatusInProgress, PriorityLow)

		require.NoError(t, store.UpdateTask(task1))
		require.NoError(t, store.UpdateTask(task2))
		require.NoError(t, store.UpdateTask(task3))

		tasks, err := store.ListTasks(TaskFilter{})
		assert.NoError(t, err)
		assert.Len(t, tasks, 3)
	})

	t.Run("filter by changeID", func(t *testing.T) {
		t.Parallel()

		store := setupBoltStore(t)

		task1 := createTestTask("1", "change-1", RegStatusPending, PriorityMedium)
		task2 := createTestTask("2", "change-2", RegStatusCompleted, PriorityHigh)
		task3 := createTestTask("3", "change-1", RegStatusInProgress, PriorityLow)

		require.NoError(t, store.UpdateTask(task1))
		require.NoError(t, store.UpdateTask(task2))
		require.NoError(t, store.UpdateTask(task3))

		tasks, err := store.ListTasks(TaskFilter{ChangeID: "change-1"})
		assert.NoError(t, err)
		assert.Len(t, tasks, 2)

		changeIDs := make(map[string]bool)
		for _, task := range tasks {
			changeIDs[task.ID.ChangeID] = true
		}
		assert.True(t, changeIDs["change-1"])
		assert.False(t, changeIDs["change-2"])
	})

	t.Run("filter by status", func(t *testing.T) {
		t.Parallel()

		store := setupBoltStore(t)

		task1 := createTestTask("1", "change-1", RegStatusPending, PriorityMedium)
		task2 := createTestTask("2", "change-2", RegStatusCompleted, PriorityHigh)
		task3 := createTestTask("3", "change-1", RegStatusCompleted, PriorityLow)

		require.NoError(t, store.UpdateTask(task1))
		require.NoError(t, store.UpdateTask(task2))
		require.NoError(t, store.UpdateTask(task3))

		tasks, err := store.ListTasks(TaskFilter{Status: RegStatusCompleted})
		assert.NoError(t, err)
		assert.Len(t, tasks, 2)
		for _, task := range tasks {
			assert.Equal(t, RegStatusCompleted, task.Status)
		}
	})

	t.Run("filter by priority", func(t *testing.T) {
		t.Parallel()

		store := setupBoltStore(t)

		task1 := createTestTask("1", "change-1", RegStatusPending, PriorityMedium)
		task2 := createTestTask("2", "change-2", RegStatusPending, PriorityHigh)
		task3 := createTestTask("3", "change-1", RegStatusPending, PriorityHigh)

		require.NoError(t, store.UpdateTask(task1))
		require.NoError(t, store.UpdateTask(task2))
		require.NoError(t, store.UpdateTask(task3))

		tasks, err := store.ListTasks(TaskFilter{Priority: PriorityHigh})
		assert.NoError(t, err)
		assert.Len(t, tasks, 2)
		for _, task := range tasks {
			assert.Equal(t, PriorityHigh, task.Priority)
		}
	})

	t.Run("filter unblocked", func(t *testing.T) {
		t.Parallel()

		store := setupBoltStore(t)

		task1 := createTestTask("1", "change-1", RegStatusPending, PriorityMedium)
		task2 := createTestTask("2", "change-2", RegStatusPending, PriorityHigh)
		task2.BlockedBy = []TaskID{{ChangeID: "change-1", TaskNum: "1"}}

		require.NoError(t, store.UpdateTask(task1))
		require.NoError(t, store.UpdateTask(task2))

		tasks, err := store.ListTasks(TaskFilter{Unblocked: true})
		assert.NoError(t, err)
		assert.Len(t, tasks, 1)
		assert.Equal(t, TaskID{ChangeID: "change-1", TaskNum: "1"}, tasks[0].ID)
	})
}

func TestBoltStore_NextTasks(t *testing.T) {
	t.Parallel()

	t.Run("empty store", func(t *testing.T) {
		t.Parallel()

		store := setupBoltStore(t)

		tasks, err := store.NextTasks(5)
		assert.NoError(t, err)
		assert.Empty(t, tasks)
	})

	t.Run("unblocked tasks sorted by priority", func(t *testing.T) {
		t.Parallel()

		store := setupBoltStore(t)

		task1 := createTestTask("1", "change-1", RegStatusPending, PriorityMedium)
		task2 := createTestTask("2", "change-2", RegStatusPending, PriorityHigh)
		task3 := createTestTask("3", "change-1", RegStatusPending, PriorityLow)
		task4 := createTestTask("4", "change-2", RegStatusPending, PriorityCritical)

		require.NoError(t, store.UpdateTask(task1))
		require.NoError(t, store.UpdateTask(task2))
		require.NoError(t, store.UpdateTask(task3))
		require.NoError(t, store.UpdateTask(task4))

		tasks, err := store.NextTasks(10)
		assert.NoError(t, err)
		assert.Len(t, tasks, 4)

		assert.Equal(t, PriorityCritical, tasks[0].Priority)
		assert.Equal(t, PriorityHigh, tasks[1].Priority)
		assert.Equal(t, PriorityMedium, tasks[2].Priority)
		assert.Equal(t, PriorityLow, tasks[3].Priority)
	})

	t.Run("respects count limit", func(t *testing.T) {
		t.Parallel()

		store := setupBoltStore(t)

		for i := 1; i <= 5; i++ {
			task := createTestTask(string(rune('0'+i)), "change-1", RegStatusPending, PriorityMedium)
			require.NoError(t, store.UpdateTask(task))
		}

		tasks, err := store.NextTasks(3)
		assert.NoError(t, err)
		assert.Len(t, tasks, 3)
	})

	t.Run("excludes blocked tasks", func(t *testing.T) {
		t.Parallel()

		store := setupBoltStore(t)

		task1 := createTestTask("1", "change-1", RegStatusPending, PriorityHigh)
		task2 := createTestTask("2", "change-1", RegStatusPending, PriorityMedium)
		task2.BlockedBy = []TaskID{{ChangeID: "change-1", TaskNum: "1"}}

		require.NoError(t, store.UpdateTask(task1))
		require.NoError(t, store.UpdateTask(task2))

		tasks, err := store.NextTasks(10)
		assert.NoError(t, err)
		assert.Len(t, tasks, 1)
		assert.Equal(t, TaskID{ChangeID: "change-1", TaskNum: "1"}, tasks[0].ID)
	})

	t.Run("excludes completed tasks", func(t *testing.T) {
		t.Parallel()

		store := setupBoltStore(t)

		task1 := createTestTask("1", "change-1", RegStatusCompleted, PriorityHigh)
		task2 := createTestTask("2", "change-1", RegStatusPending, PriorityMedium)

		require.NoError(t, store.UpdateTask(task1))
		require.NoError(t, store.UpdateTask(task2))

		tasks, err := store.NextTasks(10)
		assert.NoError(t, err)
		assert.Len(t, tasks, 1)
		assert.Equal(t, TaskID{ChangeID: "change-1", TaskNum: "2"}, tasks[0].ID)
	})

	t.Run("zero count defaults to 1", func(t *testing.T) {
		t.Parallel()

		store := setupBoltStore(t)

		task1 := createTestTask("1", "change-1", RegStatusPending, PriorityHigh)
		task2 := createTestTask("2", "change-1", RegStatusPending, PriorityMedium)

		require.NoError(t, store.UpdateTask(task1))
		require.NoError(t, store.UpdateTask(task2))

		tasks, err := store.NextTasks(0)
		assert.NoError(t, err)
		assert.Len(t, tasks, 1)
		assert.Equal(t, PriorityHigh, tasks[0].Priority)
	})
}

func TestBoltStore_AddDependency(t *testing.T) {
	t.Parallel()

	t.Run("adds single dependency", func(t *testing.T) {
		t.Parallel()

		store := setupBoltStore(t)

		task1 := createTestTask("1", "change-1", RegStatusPending, PriorityMedium)
		task2 := createTestTask("2", "change-1", RegStatusPending, PriorityHigh)

		require.NoError(t, store.UpdateTask(task1))
		require.NoError(t, store.UpdateTask(task2))

		err := store.AddDependency(task2.ID, task1.ID)
		assert.NoError(t, err)

		blockers, err := store.GetBlockers(task2.ID)
		assert.NoError(t, err)
		assert.Len(t, blockers, 1)
		assert.Equal(t, task1.ID, blockers[0])
	})

	t.Run("adds multiple dependencies", func(t *testing.T) {
		t.Parallel()

		store := setupBoltStore(t)

		task1 := createTestTask("1", "change-1", RegStatusPending, PriorityMedium)
		task2 := createTestTask("2", "change-1", RegStatusPending, PriorityHigh)
		task3 := createTestTask("3", "change-1", RegStatusPending, PriorityLow)

		require.NoError(t, store.UpdateTask(task1))
		require.NoError(t, store.UpdateTask(task2))
		require.NoError(t, store.UpdateTask(task3))

		err := store.AddDependency(task3.ID, task1.ID)
		assert.NoError(t, err)

		err = store.AddDependency(task3.ID, task2.ID)
		assert.NoError(t, err)

		blockers, err := store.GetBlockers(task3.ID)
		assert.NoError(t, err)
		assert.Len(t, blockers, 2)
	})

	t.Run("duplicate dependency is ignored", func(t *testing.T) {
		t.Parallel()

		store := setupBoltStore(t)

		task1 := createTestTask("1", "change-1", RegStatusPending, PriorityMedium)
		task2 := createTestTask("2", "change-1", RegStatusPending, PriorityHigh)

		require.NoError(t, store.UpdateTask(task1))
		require.NoError(t, store.UpdateTask(task2))

		err := store.AddDependency(task2.ID, task1.ID)
		require.NoError(t, err)

		err = store.AddDependency(task2.ID, task1.ID)
		assert.NoError(t, err)

		blockers, err := store.GetBlockers(task2.ID)
		assert.NoError(t, err)
		assert.Len(t, blockers, 1)
	})
}

func TestBoltStore_RemoveDependency(t *testing.T) {
	t.Parallel()

	t.Run("removes existing dependency", func(t *testing.T) {
		t.Parallel()

		store := setupBoltStore(t)

		task1 := createTestTask("1", "change-1", RegStatusPending, PriorityMedium)
		task2 := createTestTask("2", "change-1", RegStatusPending, PriorityHigh)
		task3 := createTestTask("3", "change-1", RegStatusPending, PriorityLow)

		require.NoError(t, store.UpdateTask(task1))
		require.NoError(t, store.UpdateTask(task2))
		require.NoError(t, store.UpdateTask(task3))

		require.NoError(t, store.AddDependency(task3.ID, task1.ID))
		require.NoError(t, store.AddDependency(task3.ID, task2.ID))

		err := store.RemoveDependency(task3.ID, task1.ID)
		assert.NoError(t, err)

		blockers, err := store.GetBlockers(task3.ID)
		assert.NoError(t, err)
		assert.Len(t, blockers, 1)
		assert.Equal(t, task2.ID, blockers[0])
	})

	t.Run("remove non-existent is no-op", func(t *testing.T) {
		t.Parallel()

		store := setupBoltStore(t)

		task1 := createTestTask("1", "change-1", RegStatusPending, PriorityMedium)
		task2 := createTestTask("2", "change-1", RegStatusPending, PriorityHigh)

		require.NoError(t, store.UpdateTask(task1))
		require.NoError(t, store.UpdateTask(task2))

		err := store.RemoveDependency(task2.ID, task1.ID)
		assert.NoError(t, err)

		blockers, err := store.GetBlockers(task2.ID)
		assert.NoError(t, err)
		assert.Len(t, blockers, 0)
	})
}

func TestBoltStore_GetBlockers(t *testing.T) {
	t.Parallel()

	t.Run("no blockers", func(t *testing.T) {
		t.Parallel()

		store := setupBoltStore(t)

		task := createTestTask("1", "change-1", RegStatusPending, PriorityMedium)
		require.NoError(t, store.UpdateTask(task))

		blockers, err := store.GetBlockers(task.ID)
		assert.NoError(t, err)
		assert.Empty(t, blockers)
	})

	t.Run("completed dependencies are not blockers", func(t *testing.T) {
		t.Parallel()

		store := setupBoltStore(t)

		task1 := createTestTask("1", "change-1", RegStatusCompleted, PriorityHigh)
		task2 := createTestTask("2", "change-1", RegStatusPending, PriorityMedium)

		require.NoError(t, store.UpdateTask(task1))
		require.NoError(t, store.UpdateTask(task2))
		require.NoError(t, store.AddDependency(task2.ID, task1.ID))

		blockers, err := store.GetBlockers(task2.ID)
		assert.NoError(t, err)
		assert.Empty(t, blockers)
	})

	t.Run("incomplete dependencies are blockers", func(t *testing.T) {
		t.Parallel()

		store := setupBoltStore(t)

		task1 := createTestTask("1", "change-1", RegStatusPending, PriorityHigh)
		task2 := createTestTask("2", "change-1", RegStatusPending, PriorityMedium)

		require.NoError(t, store.UpdateTask(task1))
		require.NoError(t, store.UpdateTask(task2))
		require.NoError(t, store.AddDependency(task2.ID, task1.ID))

		blockers, err := store.GetBlockers(task2.ID)
		assert.NoError(t, err)
		assert.Len(t, blockers, 1)
		assert.Equal(t, task1.ID, blockers[0])
	})

	t.Run("mix of completed and incomplete", func(t *testing.T) {
		t.Parallel()

		store := setupBoltStore(t)

		task1 := createTestTask("1", "change-1", RegStatusCompleted, PriorityHigh)
		task2 := createTestTask("2", "change-1", RegStatusPending, PriorityMedium)
		task3 := createTestTask("3", "change-1", RegStatusPending, PriorityLow)

		require.NoError(t, store.UpdateTask(task1))
		require.NoError(t, store.UpdateTask(task2))
		require.NoError(t, store.UpdateTask(task3))
		require.NoError(t, store.AddDependency(task3.ID, task1.ID))
		require.NoError(t, store.AddDependency(task3.ID, task2.ID))

		blockers, err := store.GetBlockers(task3.ID)
		assert.NoError(t, err)
		assert.Len(t, blockers, 1)
		assert.Equal(t, task2.ID, blockers[0])
	})
}

func TestBoltStore_GetChangeSummary(t *testing.T) {
	t.Parallel()

	t.Run("existing change", func(t *testing.T) {
		t.Parallel()

		store := setupBoltStore(t)

		task1 := createTestTask("1", "change-1", RegStatusPending, PriorityMedium)
		task2 := createTestTask("2", "change-1", RegStatusCompleted, PriorityHigh)

		require.NoError(t, store.UpdateTask(task1))
		require.NoError(t, store.UpdateTask(task2))

		summary, err := store.GetChangeSummary("change-1")
		assert.NoError(t, err)
		assert.Equal(t, "change-1", summary.ID)
		assert.Equal(t, 2, summary.Total)
		assert.Equal(t, 1, summary.Completed)
		assert.Equal(t, 1, summary.Total-summary.Completed)
	})

	t.Run("not found error", func(t *testing.T) {
		t.Parallel()

		store := setupBoltStore(t)

		summary, err := store.GetChangeSummary("nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "change not found")
		assert.NotNil(t, summary)
		assert.True(t, summary.ID == "")
	})
}

func TestBoltStore_ListChanges(t *testing.T) {
	t.Parallel()

	t.Run("empty store", func(t *testing.T) {
		t.Parallel()

		store := setupBoltStore(t)

		changes, err := store.ListChanges()
		assert.NoError(t, err)
		assert.Empty(t, changes)
	})

	t.Run("multiple changes", func(t *testing.T) {
		t.Parallel()

		store := setupBoltStore(t)

		task1 := createTestTask("1", "change-1", RegStatusPending, PriorityMedium)
		task2 := createTestTask("2", "change-1", RegStatusCompleted, PriorityHigh)
		task3 := createTestTask("1", "change-2", RegStatusInProgress, PriorityLow)

		require.NoError(t, store.UpdateTask(task1))
		require.NoError(t, store.UpdateTask(task2))
		require.NoError(t, store.UpdateTask(task3))

		changes, err := store.ListChanges()
		assert.NoError(t, err)
		assert.Len(t, changes, 2)

		changeIDs := make(map[string]bool)
		for _, change := range changes {
			changeIDs[change.ID] = true
		}
		assert.True(t, changeIDs["change-1"])
		assert.True(t, changeIDs["change-2"])
	})
}

func TestBoltStore_UpdateChange(t *testing.T) {
	t.Parallel()

	t.Run("updates summary correctly", func(t *testing.T) {
		t.Parallel()

		store := setupBoltStore(t)

		task1 := createTestTask("1", "change-1", RegStatusPending, PriorityMedium)
		task2 := createTestTask("2", "change-1", RegStatusInProgress, PriorityHigh)
		task3 := createTestTask("3", "change-1", RegStatusCompleted, PriorityLow)

		require.NoError(t, store.UpdateTask(task1))
		require.NoError(t, store.UpdateTask(task2))
		require.NoError(t, store.UpdateTask(task3))

		task2.Status = RegStatusCompleted
		require.NoError(t, store.UpdateTask(task2))

		err := store.UpdateChange(ChangeSummary{ID: "change-1"})
		assert.NoError(t, err)

		summary, err := store.GetChangeSummary("change-1")
		assert.NoError(t, err)
		assert.Equal(t, 3, summary.Total)
		assert.Equal(t, 2, summary.Completed)
		assert.Equal(t, 0, summary.InProgress)
	})
}

func TestBoltStore_SetSyncedAt_GetSyncedAt(t *testing.T) {
	t.Parallel()

	t.Run("round trip", func(t *testing.T) {
		t.Parallel()

		store := setupBoltStore(t)

		expected := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

		err := store.SetSyncedAt(expected)
		assert.NoError(t, err)

		actual, err := store.GetSyncedAt()
		assert.NoError(t, err)
		assert.True(t, expected.Equal(actual))
	})

	t.Run("not found returns zero time", func(t *testing.T) {
		t.Parallel()

		store := setupBoltStore(t)

		actual, err := store.GetSyncedAt()
		assert.NoError(t, err)
		assert.True(t, actual.IsZero())
	})
}

func TestBoltStore_SetMeta_GetMeta(t *testing.T) {
	t.Parallel()

	t.Run("round trip", func(t *testing.T) {
		t.Parallel()

		store := setupBoltStore(t)

		err := store.SetMeta("test_key", "test_value")
		assert.NoError(t, err)

		value, err := store.GetMeta("test_key")
		assert.NoError(t, err)
		assert.Equal(t, "test_value", value)
	})

	t.Run("not found error", func(t *testing.T) {
		t.Parallel()

		store := setupBoltStore(t)

		value, err := store.GetMeta("nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "key not found")
		assert.Empty(t, value)
	})

	t.Run("multiple keys", func(t *testing.T) {
		t.Parallel()

		store := setupBoltStore(t)

		err := store.SetMeta("key1", "value1")
		require.NoError(t, err)

		err = store.SetMeta("key2", "value2")
		require.NoError(t, err)

		value1, err := store.GetMeta("key1")
		assert.NoError(t, err)
		assert.Equal(t, "value1", value1)

		value2, err := store.GetMeta("key2")
		assert.NoError(t, err)
		assert.Equal(t, "value2", value2)
	})
}

func TestBoltStore_ClearTasks(t *testing.T) {
	t.Parallel()

	t.Run("clears all task buckets", func(t *testing.T) {
		t.Parallel()

		store := setupBoltStore(t)

		task1 := createTestTask("1", "change-1", RegStatusPending, PriorityMedium)
		task2 := createTestTask("2", "change-1", RegStatusCompleted, PriorityHigh)
		task3 := createTestTask("3", "change-2", RegStatusInProgress, PriorityLow)

		require.NoError(t, store.UpdateTask(task1))
		require.NoError(t, store.UpdateTask(task2))
		require.NoError(t, store.UpdateTask(task3))

		require.NoError(t, store.AddDependency(task2.ID, task1.ID))

		tasks, err := store.ListTasks(TaskFilter{})
		require.NoError(t, err)
		assert.Len(t, tasks, 3)

		err = store.ClearTasks()
		assert.NoError(t, err)

		tasks, err = store.ListTasks(TaskFilter{})
		assert.NoError(t, err)
		assert.Empty(t, tasks)
	})

	t.Run("clears dependencies", func(t *testing.T) {
		t.Parallel()

		store := setupBoltStore(t)

		task1 := createTestTask("1", "change-1", RegStatusPending, PriorityMedium)
		task2 := createTestTask("2", "change-1", RegStatusCompleted, PriorityHigh)

		require.NoError(t, store.UpdateTask(task1))
		require.NoError(t, store.UpdateTask(task2))
		require.NoError(t, store.AddDependency(task2.ID, task1.ID))

		require.NoError(t, store.ClearTasks())

		err := store.UpdateTask(task1)
		assert.NoError(t, err)

		blockers, err := store.GetBlockers(task2.ID)
		assert.NoError(t, err)
		assert.Empty(t, blockers)
	})
}

func TestBoltStore_Close(t *testing.T) {
	t.Parallel()

	t.Run("closes db handle", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		dbPath := tmpDir + "/test.db"

		store, err := NewBoltStore(dbPath)
		require.NoError(t, err)

		err = store.Close()
		assert.NoError(t, err)

		err = store.Close()
		assert.NoError(t, err)
	})
}

func TestBoltStore_NewBoltStore(t *testing.T) {
	t.Parallel()

	t.Run("creates store with temp directory", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		dbPath := tmpDir + "/subdir/test.db"

		store, err := NewBoltStore(dbPath)
		assert.NoError(t, err)
		assert.NotNil(t, store)
		assert.Equal(t, dbPath, store.path)

		_, err = os.Stat(dbPath)
		assert.NoError(t, err)

		_ = store.Close()
	})

	t.Run("error on empty path", func(t *testing.T) {
		t.Parallel()

		store, err := NewBoltStore("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "path cannot be empty")
		assert.Nil(t, store)
	})

	t.Run("initializes all buckets", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		dbPath := tmpDir + "/test.db"

		store, err := NewBoltStore(dbPath)
		require.NoError(t, err)
		defer func() { _ = store.Close() }()

		task := createTestTask("1", "change-1", RegStatusPending, PriorityMedium)
		err = store.UpdateTask(task)
		assert.NoError(t, err)

		err = store.SetMeta("test", "value")
		assert.NoError(t, err)

		_, err = store.GetChangeSummary("change-1")
		assert.NoError(t, err)
	})
}

func TestBoltStore_integration(t *testing.T) {
	t.Parallel()

	t.Run("complete workflow", func(t *testing.T) {
		t.Parallel()

		store := setupBoltStore(t)

		task1 := createTestTask("1", "change-1", RegStatusPending, PriorityCritical)
		task2 := createTestTask("2", "change-1", RegStatusPending, PriorityHigh)
		task3 := createTestTask("3", "change-1", RegStatusPending, PriorityMedium)

		require.NoError(t, store.UpdateTask(task1))
		require.NoError(t, store.UpdateTask(task2))
		require.NoError(t, store.UpdateTask(task3))

		require.NoError(t, store.AddDependency(task2.ID, task1.ID))
		require.NoError(t, store.AddDependency(task3.ID, task1.ID))

		nextTasks, err := store.NextTasks(1)
		require.NoError(t, err)
		assert.Len(t, nextTasks, 1)
		assert.Equal(t, task1.ID, nextTasks[0].ID)

		blockers2, err := store.GetBlockers(task2.ID)
		require.NoError(t, err)
		assert.Len(t, blockers2, 1)
		assert.Equal(t, task1.ID, blockers2[0])

		task1.Status = RegStatusCompleted
		require.NoError(t, store.UpdateTask(task1))

		nextTasks, err = store.NextTasks(2)
		require.NoError(t, err)
		assert.Len(t, nextTasks, 2)

		blockers2, err = store.GetBlockers(task2.ID)
		require.NoError(t, err)
		assert.Empty(t, blockers2)

		summary, err := store.GetChangeSummary("change-1")
		require.NoError(t, err)
		assert.Equal(t, 3, summary.Total)
		assert.Equal(t, 1, summary.Completed)
	})
}
