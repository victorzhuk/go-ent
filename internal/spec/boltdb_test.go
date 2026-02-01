package spec

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBoltStore(t *testing.T) {
	t.Run("creates bolt store successfully", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		store, err := NewBoltStore(tmpDir)
		require.NoError(t, err)
		require.NotNil(t, store)
		defer store.Close()

		assert.Contains(t, store.path, ".cache")
	})

	t.Run("creates cache directory", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		store, err := NewBoltStore(tmpDir)
		require.NoError(t, err)
		defer store.Close()

		_, err = store.GetChange("nonexistent")
		require.NoError(t, err)
	})
}

func TestPutTask_GetTask(t *testing.T) {
	t.Run("store and retrieve task", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		store, err := NewBoltStore(tmpDir)
		require.NoError(t, err)
		defer store.Close()

		task := &Task{
			ID:         "change-1:1.1",
			ChangeID:   "change-1",
			TaskNum:    "1.1",
			Content:    "Test task content",
			Status:     TaskPending,
			Priority:   1,
			DependsOn:  []string{"1.0"},
			SourceLine: 10,
			SyncedAt:   time.Now(),
		}

		err = store.PutTask(task)
		require.NoError(t, err)

		retrieved, err := store.GetTask("change-1", "1.1")
		require.NoError(t, err)
		require.NotNil(t, retrieved)

		assert.Equal(t, task.ID, retrieved.ID)
		assert.Equal(t, task.ChangeID, retrieved.ChangeID)
		assert.Equal(t, task.TaskNum, retrieved.TaskNum)
		assert.Equal(t, task.Content, retrieved.Content)
		assert.Equal(t, task.Status, retrieved.Status)
		assert.Equal(t, task.Priority, retrieved.Priority)
		assert.Equal(t, task.DependsOn, retrieved.DependsOn)
		assert.Equal(t, task.SourceLine, retrieved.SourceLine)
	})

	t.Run("get non-existent task returns nil", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		store, err := NewBoltStore(tmpDir)
		require.NoError(t, err)
		defer store.Close()

		task, err := store.GetTask("nonexistent", "1.1")
		require.NoError(t, err)
		assert.Nil(t, task)
	})

	t.Run("put nil task returns error", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		store, err := NewBoltStore(tmpDir)
		require.NoError(t, err)
		defer store.Close()

		err = store.PutTask(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nil task")
	})
}

func TestDeleteTask(t *testing.T) {
	t.Run("delete existing task", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		store, err := NewBoltStore(tmpDir)
		require.NoError(t, err)
		defer store.Close()

		task := &Task{
			ID:       "change-1:1.1",
			ChangeID: "change-1",
			TaskNum:  "1.1",
			Content:  "Test task",
			Status:   TaskPending,
		}

		err = store.PutTask(task)
		require.NoError(t, err)

		err = store.DeleteTask("change-1", "1.1")
		require.NoError(t, err)

		retrieved, err := store.GetTask("change-1", "1.1")
		require.NoError(t, err)
		assert.Nil(t, retrieved)
	})

	t.Run("delete non-existent task succeeds", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		store, err := NewBoltStore(tmpDir)
		require.NoError(t, err)
		defer store.Close()

		err = store.DeleteTask("nonexistent", "1.1")
		require.NoError(t, err)
	})
}

func TestPutChange_GetChange(t *testing.T) {
	t.Run("store and retrieve change", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		store, err := NewBoltStore(tmpDir)
		require.NoError(t, err)
		defer store.Close()

		change := &ChangeMetadata{
			ID:         "change-1",
			Title:      "Test Change",
			Status:     StatusDraft,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
			TaskCount:  5,
			Completed:  2,
			InProgress: 1,
			Blocked:    0,
		}

		err = store.PutChange(change)
		require.NoError(t, err)

		retrieved, err := store.GetChange("change-1")
		require.NoError(t, err)
		require.NotNil(t, retrieved)

		assert.Equal(t, change.ID, retrieved.ID)
		assert.Equal(t, change.Title, retrieved.Title)
		assert.Equal(t, change.Status, retrieved.Status)
		assert.Equal(t, change.TaskCount, retrieved.TaskCount)
		assert.Equal(t, change.Completed, retrieved.Completed)
		assert.Equal(t, change.InProgress, retrieved.InProgress)
		assert.Equal(t, change.Blocked, retrieved.Blocked)
	})

	t.Run("get non-existent change returns nil", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		store, err := NewBoltStore(tmpDir)
		require.NoError(t, err)
		defer store.Close()

		change, err := store.GetChange("nonexistent")
		require.NoError(t, err)
		assert.Nil(t, change)
	})

	t.Run("put nil change returns error", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		store, err := NewBoltStore(tmpDir)
		require.NoError(t, err)
		defer store.Close()

		err = store.PutChange(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nil change")
	})
}

func TestPutDeps_GetDeps(t *testing.T) {
	t.Run("store and retrieve dependencies", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		store, err := NewBoltStore(tmpDir)
		require.NoError(t, err)
		defer store.Close()

		deps := &DependencyInfo{
			TaskID:        "change-1:1.2",
			DependsOn:     []string{"1.1"},
			DependedBy:    []string{"1.3"},
			IsBlocked:     true,
			BlockingTasks: []string{"change-1:1.1"},
		}

		err = store.PutDeps(deps)
		require.NoError(t, err)

		retrieved, err := store.GetDeps("change-1", "1.2")
		require.NoError(t, err)
		require.NotNil(t, retrieved)

		assert.Equal(t, deps.TaskID, retrieved.TaskID)
		assert.Equal(t, deps.DependsOn, retrieved.DependsOn)
		assert.Equal(t, deps.DependedBy, retrieved.DependedBy)
		assert.Equal(t, deps.IsBlocked, retrieved.IsBlocked)
		assert.Equal(t, deps.BlockingTasks, retrieved.BlockingTasks)
	})

	t.Run("get non-existent deps returns nil", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		store, err := NewBoltStore(tmpDir)
		require.NoError(t, err)
		defer store.Close()

		deps, err := store.GetDeps("nonexistent", "1.1")
		require.NoError(t, err)
		assert.Nil(t, deps)
	})

	t.Run("put nil deps returns error", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		store, err := NewBoltStore(tmpDir)
		require.NoError(t, err)
		defer store.Close()

		err = store.PutDeps(nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nil deps")
	})
}

func TestListAllChanges(t *testing.T) {
	t.Run("list multiple changes", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		store, err := NewBoltStore(tmpDir)
		require.NoError(t, err)
		defer store.Close()

		change1 := &ChangeMetadata{
			ID:     "change-1",
			Title:  "First Change",
			Status: StatusDraft,
		}

		change2 := &ChangeMetadata{
			ID:     "change-2",
			Title:  "Second Change",
			Status: StatusActive,
		}

		err = store.PutChange(change1)
		require.NoError(t, err)
		err = store.PutChange(change2)
		require.NoError(t, err)

		changes, err := store.ListAllChanges()
		require.NoError(t, err)
		assert.Len(t, changes, 2)
	})

	t.Run("list empty changes returns empty list", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		store, err := NewBoltStore(tmpDir)
		require.NoError(t, err)
		defer store.Close()

		changes, err := store.ListAllChanges()
		require.NoError(t, err)
		assert.Empty(t, changes)
	})
}

func TestListTasks(t *testing.T) {
	t.Run("list tasks by change ID", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		store, err := NewBoltStore(tmpDir)
		require.NoError(t, err)
		defer store.Close()

		task1 := &Task{
			ID:       "change-1:1.1",
			ChangeID: "change-1",
			TaskNum:  "1.1",
			Content:  "Task 1",
			Status:   TaskPending,
		}

		task2 := &Task{
			ID:       "change-1:1.2",
			ChangeID: "change-1",
			TaskNum:  "1.2",
			Content:  "Task 2",
			Status:   TaskCompleted,
		}

		task3 := &Task{
			ID:       "change-2:1.1",
			ChangeID: "change-2",
			TaskNum:  "1.1",
			Content:  "Other change task",
			Status:   TaskPending,
		}

		err = store.PutTask(task1)
		require.NoError(t, err)
		err = store.PutTask(task2)
		require.NoError(t, err)
		err = store.PutTask(task3)
		require.NoError(t, err)

		tasks, err := store.ListTasks("change-1", "")
		require.NoError(t, err)
		assert.Len(t, tasks, 2)
	})

	t.Run("list tasks by status", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		store, err := NewBoltStore(tmpDir)
		require.NoError(t, err)
		defer store.Close()

		task1 := &Task{
			ID:       "change-1:1.1",
			ChangeID: "change-1",
			TaskNum:  "1.1",
			Content:  "Task 1",
			Status:   TaskPending,
		}

		task2 := &Task{
			ID:       "change-1:1.2",
			ChangeID: "change-1",
			TaskNum:  "1.2",
			Content:  "Task 2",
			Status:   TaskCompleted,
		}

		task3 := &Task{
			ID:       "change-1:1.3",
			ChangeID: "change-1",
			TaskNum:  "1.3",
			Content:  "Task 3",
			Status:   TaskCompleted,
		}

		err = store.PutTask(task1)
		require.NoError(t, err)
		err = store.PutTask(task2)
		require.NoError(t, err)
		err = store.PutTask(task3)
		require.NoError(t, err)

		tasks, err := store.ListTasks("", string(TaskCompleted))
		require.NoError(t, err)
		assert.Len(t, tasks, 2)
	})
}
