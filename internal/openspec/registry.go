package openspec

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/victorzhuk/go-ent/internal/spec"
)

type registryUpdater interface {
	UpdateTask(id spec.TaskID, updates spec.TaskUpdate) error
}

type stateStoreUpdater interface {
	UpdateTaskInFile(id spec.TaskID, status spec.RegistryTaskStatus, notes string) error
}

type TaskTracker struct {
	registryStore registryUpdater
	changeID      string
	stateStore    stateStoreUpdater
}

func NewTaskTracker(registryStore *spec.RegistryStore, changeID string, stateStore *spec.StateStore) *TaskTracker {
	return &TaskTracker{
		registryStore: registryStore,
		changeID:      changeID,
		stateStore:    stateStore,
	}
}

func (t *TaskTracker) ExtractTaskID(description string) spec.TaskID {
	re := regexp.MustCompile(`(\d+\.\d+)`)
	matches := re.FindStringSubmatch(description)
	if len(matches) < 2 {
		return spec.TaskID{}
	}

	parts := strings.Split(matches[1], ".")
	if len(parts) != 2 {
		return spec.TaskID{}
	}

	return spec.TaskID{
		ChangeID: t.changeID,
		TaskNum:  parts[1],
	}
}

func (t *TaskTracker) MarkInProgress(taskID spec.TaskID) error {
	if taskID.IsZero() {
		return fmt.Errorf("task ID is empty")
	}

	update := spec.TaskUpdate{
		Status: ptrTo(spec.RegStatusInProgress),
	}

	if err := t.registryStore.UpdateTask(taskID, update); err != nil {
		return fmt.Errorf("mark task in progress: %w", err)
	}

	if t.stateStore != nil {
		if err := t.stateStore.UpdateTaskInFile(taskID, spec.RegStatusInProgress, ""); err != nil {
			return fmt.Errorf("update task in file: %w", err)
		}
	}

	return nil
}

func (t *TaskTracker) MarkCompleted(taskID spec.TaskID, notes ...string) error {
	if taskID.IsZero() {
		return fmt.Errorf("task ID is empty")
	}

	update := spec.TaskUpdate{
		Status: ptrTo(spec.RegStatusCompleted),
	}

	var fileNotes string
	if len(notes) > 0 && notes[0] != "" {
		update.Notes = &notes[0]
		fileNotes = notes[0]
	} else {
		now := time.Now().Format("2006-01-02")
		update.Notes = ptrTo(now)
		fileNotes = "✓ " + now
	}

	if err := t.registryStore.UpdateTask(taskID, update); err != nil {
		return fmt.Errorf("mark task completed: %w", err)
	}

	if t.stateStore != nil {
		if err := t.stateStore.UpdateTaskInFile(taskID, spec.RegStatusCompleted, fileNotes); err != nil {
			return fmt.Errorf("update task in file: %w", err)
		}
	}

	return nil
}

func (t *TaskTracker) MarkFailed(taskID spec.TaskID, errorMsg string) error {
	if taskID.IsZero() {
		return fmt.Errorf("task ID is empty")
	}

	update := spec.TaskUpdate{
		Notes: &errorMsg,
	}

	if err := t.registryStore.UpdateTask(taskID, update); err != nil {
		return fmt.Errorf("mark task failed: %w", err)
	}

	if t.stateStore != nil {
		if err := t.stateStore.UpdateTaskInFile(taskID, spec.RegStatusPending, "❌ "+errorMsg); err != nil {
			return fmt.Errorf("update task in file: %w", err)
		}
	}

	return nil
}

func ptrTo[T any](v T) *T {
	return &v
}
