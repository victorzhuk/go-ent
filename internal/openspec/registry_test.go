package openspec

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/victorzhuk/go-ent/internal/spec"
)

func TestTaskTracker_ExtractTaskID(t *testing.T) {
	t.Run("extracts task ID from description", func(t *testing.T) {
		tracker := &TaskTracker{
			changeID: "test-change",
		}

		taskID := tracker.ExtractTaskID("Complete task 9.1")
		assert.Equal(t, "test-change", taskID.ChangeID)
		assert.Equal(t, "1", taskID.TaskNum)
		assert.Equal(t, "test-change/1", taskID.String())
	})

	t.Run("returns empty for description without task ID", func(t *testing.T) {
		tracker := &TaskTracker{
			changeID: "test-change",
		}

		taskID := tracker.ExtractTaskID("Just implement the feature")
		assert.True(t, taskID.IsZero())
	})

	t.Run("handles various formats", func(t *testing.T) {
		tracker := &TaskTracker{
			changeID: "my-change",
		}

		tests := []struct {
			desc     string
			expected spec.TaskID
		}{
			{"Task 5.3 implementation", spec.TaskID{ChangeID: "my-change", TaskNum: "3"}},
			{"9.1.2", spec.TaskID{ChangeID: "my-change", TaskNum: "1"}},
			{"Complete 10.15 now", spec.TaskID{ChangeID: "my-change", TaskNum: "15"}},
		}

		for _, tt := range tests {
			taskID := tracker.ExtractTaskID(tt.desc)
			assert.Equal(t, tt.expected.ChangeID, taskID.ChangeID)
			assert.Equal(t, tt.expected.TaskNum, taskID.TaskNum)
		}
	})
}

func TestTaskTracker_MarkInProgress(t *testing.T) {
	t.Run("marks task as in progress", func(t *testing.T) {
		store := &mockRegistryStore{}
		tracker := &TaskTracker{
			registryStore: store,
			changeID:      "test-change",
		}

		taskID := spec.TaskID{ChangeID: "test-change", TaskNum: "1"}

		err := tracker.MarkInProgress(taskID)
		assert.NoError(t, err)
		assert.True(t, store.updateCalled)
		assert.Equal(t, spec.RegStatusInProgress, *store.lastStatus)
	})

	t.Run("returns error for empty task ID", func(t *testing.T) {
		store := &mockRegistryStore{}
		tracker := &TaskTracker{
			registryStore: store,
			changeID:      "test-change",
		}

		err := tracker.MarkInProgress(spec.TaskID{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "task ID is empty")
	})
}

func TestTaskTracker_MarkCompleted(t *testing.T) {
	t.Run("marks task as completed", func(t *testing.T) {
		store := &mockRegistryStore{}
		tracker := &TaskTracker{
			registryStore: store,
			changeID:      "test-change",
		}

		taskID := spec.TaskID{ChangeID: "test-change", TaskNum: "1"}

		err := tracker.MarkCompleted(taskID)
		assert.NoError(t, err)
		assert.True(t, store.updateCalled)
		assert.Equal(t, spec.RegStatusCompleted, *store.lastStatus)
	})

	t.Run("adds notes when marking completed", func(t *testing.T) {
		store := &mockRegistryStore{}
		tracker := &TaskTracker{
			registryStore: store,
			changeID:      "test-change",
		}

		taskID := spec.TaskID{ChangeID: "test-change", TaskNum: "1"}
		notes := "Completed successfully by worker abc123"

		err := tracker.MarkCompleted(taskID, notes)
		assert.NoError(t, err)
		assert.Equal(t, &notes, store.lastNotes)
	})

	t.Run("returns error for empty task ID", func(t *testing.T) {
		store := &mockRegistryStore{}
		tracker := &TaskTracker{
			registryStore: store,
			changeID:      "test-change",
		}

		err := tracker.MarkCompleted(spec.TaskID{})
		assert.Error(t, err)
	})
}

func TestTaskTracker_MarkFailed(t *testing.T) {
	t.Run("marks task as failed", func(t *testing.T) {
		store := &mockRegistryStore{}
		tracker := &TaskTracker{
			registryStore: store,
			changeID:      "test-change",
		}

		taskID := spec.TaskID{ChangeID: "test-change", TaskNum: "1"}
		errorMsg := "Worker timed out after 5 minutes"

		err := tracker.MarkFailed(taskID, errorMsg)
		assert.NoError(t, err)
		assert.True(t, store.updateCalled)
		assert.Equal(t, &errorMsg, store.lastNotes)
	})

	t.Run("returns error for empty task ID", func(t *testing.T) {
		store := &mockRegistryStore{}
		tracker := &TaskTracker{
			registryStore: store,
			changeID:      "test-change",
		}

		err := tracker.MarkFailed(spec.TaskID{}, "error")
		assert.Error(t, err)
	})
}

type mockRegistryStore struct {
	updateCalled bool
	lastStatus   *spec.RegistryTaskStatus
	lastNotes    *string
}

func (m *mockRegistryStore) UpdateTask(id spec.TaskID, updates spec.TaskUpdate) error {
	m.updateCalled = true
	m.lastStatus = updates.Status
	m.lastNotes = updates.Notes
	return nil
}
