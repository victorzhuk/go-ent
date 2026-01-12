package spec

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type RegistryStore struct {
	store      *Store
	bolt       *BoltStore
	stateStore *StateStore
}

func NewRegistryStore(store *Store) (*RegistryStore, error) {
	boltPath := filepath.Join(store.RootPath(), "openspec", "registry.db")
	bolt, err := NewBoltStore(boltPath)
	if err != nil {
		return nil, fmt.Errorf("create bolt store: %w", err)
	}

	stateStore := NewStateStore(store, bolt)

	return &RegistryStore{
		store:      store,
		bolt:       bolt,
		stateStore: stateStore,
	}, nil
}

func (r *RegistryStore) Close() error {
	if r.bolt != nil {
		return r.bolt.Close()
	}
	return nil
}

func (r *RegistryStore) Load() (*Registry, error) {
	// Load from BoltDB
	tasks, err := r.bolt.ListTasks(TaskFilter{})
	if err != nil {
		return nil, fmt.Errorf("list tasks from bolt: %w", err)
	}

	changes, err := r.bolt.ListChanges()
	if err != nil {
		return nil, fmt.Errorf("list changes from bolt: %w", err)
	}

	reg := &Registry{
		Version:  "1.0",
		SyncedAt: time.Now(),
		Changes:  make(map[string]ChangeSummary),
		Tasks:    tasks,
	}

	for _, change := range changes {
		reg.Changes[change.ID] = change
	}

	return reg, nil
}

func (r *RegistryStore) Save(reg *Registry) error {
	// Save is deprecated - use UpdateTask instead
	// BoltDB operations are transactional and immediate
	reg.SyncedAt = time.Now()
	return r.bolt.SetMeta("synced_at", reg.SyncedAt.Format(time.RFC3339))
}

func (r *RegistryStore) Exists() bool {
	boltPath := filepath.Join(r.store.RootPath(), "openspec", "registry.db")
	_, err := os.Stat(boltPath)
	return err == nil
}

func (r *RegistryStore) Init() error {
	// Initialize BoltDB buckets (already done in NewBoltStore)
	return r.bolt.SetMeta("version", "1.0")
}

func (r *RegistryStore) ListTasks(filter TaskFilter) ([]RegistryTask, error) {
	// Use BoltDB O(1) lookup directly
	return r.bolt.ListTasks(filter)
}

func (r *RegistryStore) GetTask(id TaskID) (*RegistryTask, error) {
	// Use BoltDB O(1) lookup directly
	return r.bolt.GetTask(id)
}

func (r *RegistryStore) UpdateTask(id TaskID, updates TaskUpdate) error {
	// Get current task to check old status
	task, err := r.bolt.GetTask(id)
	if err != nil {
		return fmt.Errorf("get task: %w", err)
	}

	now := time.Now()

	// Handle status transitions
	if updates.Status != nil {
		oldStatus := task.Status
		task.Status = *updates.Status

		if *updates.Status == RegStatusInProgress && oldStatus != RegStatusInProgress {
			task.StartedAt = &now
		}
		if *updates.Status == RegStatusCompleted && oldStatus != RegStatusCompleted {
			task.CompletedAt = &now
		}
	}

	// Apply other updates
	if updates.Priority != nil {
		task.Priority = *updates.Priority
	}
	if updates.Assignee != nil {
		task.Assignee = *updates.Assignee
	}
	if updates.Notes != nil {
		task.Notes = *updates.Notes
	}

	// Update in BoltDB (transactional)
	if err := r.bolt.UpdateTask(task); err != nil {
		return fmt.Errorf("update task in bolt: %w", err)
	}

	// Recalculate blocked tasks using BoltDB reverse index
	if err := r.recalculateBlockedByBolt(task.ID); err != nil {
		return fmt.Errorf("recalculate blockers: %w", err)
	}

	return nil
}

func (r *RegistryStore) NextTask(count int) (*NextTaskResult, error) {
	if count <= 0 {
		count = 1
	}

	// Use BoltDB NextTasks directly
	candidates, err := r.bolt.NextTasks(count + 10) // Get extras for alternatives
	if err != nil {
		return nil, fmt.Errorf("get next tasks from bolt: %w", err)
	}

	// Count blocked tasks
	allTasks, err := r.bolt.ListTasks(TaskFilter{})
	if err != nil {
		return nil, fmt.Errorf("list all tasks: %w", err)
	}

	blockedCount := 0
	for _, task := range allTasks {
		if len(task.BlockedBy) > 0 {
			blockedCount++
		}
	}

	result := &NextTaskResult{
		BlockedCount: blockedCount,
	}

	if len(candidates) > 0 {
		result.Recommended = &candidates[0]
		result.Reason = fmt.Sprintf("Highest priority (%s) unblocked task", candidates[0].Priority)
		if len(candidates[0].DependsOn) > 0 {
			result.Reason += fmt.Sprintf(". Dependencies completed: %d", len(candidates[0].DependsOn))
		}

		if len(candidates) > 1 {
			limit := count
			if len(candidates)-1 < count {
				limit = len(candidates) - 1
			}
			result.Alternatives = candidates[1 : limit+1]
		}
	}

	return result, nil
}

func (r *RegistryStore) AddDependency(task, dependsOn TaskID) error {
	// Use BoltDB AddDependency directly (includes cycle detection)
	return r.bolt.AddDependency(task, dependsOn)
}

func (r *RegistryStore) RemoveDependency(task, dependsOn TaskID) error {
	// Use BoltDB RemoveDependency directly
	return r.bolt.RemoveDependency(task, dependsOn)
}

func (r *RegistryStore) GetDependencyGraph(id TaskID) (*DependencyGraph, error) {
	task, err := r.bolt.GetTask(id)
	if err != nil {
		return nil, fmt.Errorf("get task: %w", err)
	}

	graph := &DependencyGraph{
		TaskID:        id,
		DependsOn:     make([]TaskDependencyInfo, 0),
		DependedBy:    make([]TaskDependencyInfo, 0),
		IsBlocked:     len(task.BlockedBy) > 0,
		BlockingTasks: make([]TaskID, 0),
	}

	// Get tasks this depends on
	for _, depID := range task.DependsOn {
		if dep, err := r.bolt.GetTask(depID); err == nil {
			graph.DependsOn = append(graph.DependsOn, TaskDependencyInfo{
				ID:      dep.ID,
				Content: dep.Content,
				Status:  dep.Status,
			})
			if dep.Status != RegStatusCompleted {
				graph.BlockingTasks = append(graph.BlockingTasks, dep.ID)
			}
		}
	}

	// Get tasks that depend on this (using BoltDB reverse index)
	blockedTasks, err := r.bolt.GetBlockers(id)
	if err != nil {
		return nil, fmt.Errorf("get blockers: %w", err)
	}

	for _, blockedID := range blockedTasks {
		if blocked, err := r.bolt.GetTask(blockedID); err == nil {
			graph.DependedBy = append(graph.DependedBy, TaskDependencyInfo{
				ID:      blocked.ID,
				Content: blocked.Content,
				Status:  blocked.Status,
			})
		}
	}

	return graph, nil
}

func (r *RegistryStore) RebuildFromSource() (*SyncResult, error) {
	result := &SyncResult{
		SyncedChanges: make([]string, 0),
		Added:         make([]TaskID, 0),
		Updated:       make([]TaskID, 0),
		Removed:       make([]TaskID, 0),
		Conflicts:     make([]SyncConflict, 0),
	}

	changesPath := filepath.Join(r.store.SpecPath(), "changes")
	entries, err := os.ReadDir(changesPath)
	if err != nil {
		if os.IsNotExist(err) {
			return result, nil
		}
		return nil, fmt.Errorf("read changes dir: %w", err)
	}

	// Clear existing tasks from BoltDB (full rebuild)
	if err := r.bolt.ClearTasks(); err != nil {
		return nil, fmt.Errorf("clear bolt tasks: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == "archive" {
			continue
		}

		changeID := entry.Name()
		tasksPath := filepath.Join(changesPath, changeID, "tasks.md")

		if _, err := os.Stat(tasksPath); os.IsNotExist(err) {
			continue
		}

		// Use StateStore to parse tasks WITH dependencies from HTML comments
		tasks, err := r.stateStore.ParseTasksWithDependencies(changeID, tasksPath)
		if err != nil {
			continue
		}

		summary := ChangeSummary{
			ID:        changeID,
			TasksFile: tasksPath,
			Total:     len(tasks),
		}

		for _, task := range tasks {
			if task.Status == RegStatusCompleted {
				summary.Completed++
			} else if task.Status == RegStatusInProgress {
				summary.InProgress++
			}

			// Store task in BoltDB
			if err := r.bolt.UpdateTask(&task); err != nil {
				return nil, fmt.Errorf("store task %s: %w", task.ID.String(), err)
			}

			// Store dependencies
			for _, depID := range task.DependsOn {
				if err := r.bolt.AddDependency(task.ID, depID); err != nil {
					// Ignore cycle errors during initial load - they'll be caught later
					continue
				}
			}

			result.Added = append(result.Added, task.ID)
		}

		// Store change summary
		if err := r.bolt.UpdateChange(summary); err != nil {
			return nil, fmt.Errorf("store change summary: %w", err)
		}

		result.SyncedChanges = append(result.SyncedChanges, changeID)
	}

	// Set sync metadata
	if err := r.bolt.SetMeta("synced_at", time.Now().Format(time.RFC3339)); err != nil {
		return nil, fmt.Errorf("set sync metadata: %w", err)
	}

	return result, nil
}

func (r *RegistryStore) Stats() (*RegistryStats, error) {
	tasks, err := r.bolt.ListTasks(TaskFilter{})
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}

	stats := &RegistryStats{
		TotalTasks: len(tasks),
		ByStatus:   make(map[string]int),
		ByPriority: make(map[string]int),
		ByChange:   make(map[string]int),
	}

	for _, task := range tasks {
		stats.ByStatus[string(task.Status)]++
		stats.ByPriority[string(task.Priority)]++
		stats.ByChange[task.ID.ChangeID]++
		if len(task.BlockedBy) > 0 {
			stats.BlockedCount++
		}
	}

	next, err := r.NextTask(1)
	if err == nil && next.Recommended != nil {
		stats.NextTask = next.Recommended
	}

	return stats, nil
}

// parseTasksFile is deprecated - use StateStore.ParseTasksWithDependencies instead
// Kept for backward compatibility during migration
func (r *RegistryStore) parseTasksFile(path, changeID string) ([]RegistryTask, error) {
	return r.stateStore.ParseTasksWithDependencies(changeID, path)
}

// recalculateBlockedBy is deprecated - BoltDB maintains reverse index automatically
// Kept for backward compatibility during migration
func (r *RegistryStore) recalculateBlockedBy(reg *Registry) {
	// BoltDB handles this automatically via the blocking bucket
	// This method is now a no-op
}

// recalculateBlockedByBolt updates BlockedBy for a specific task using BoltDB
func (r *RegistryStore) recalculateBlockedByBolt(taskID TaskID) error {
	task, err := r.bolt.GetTask(taskID)
	if err != nil {
		return err
	}

	// Recalculate which dependencies are still blocking this task
	blockedBy := make([]TaskID, 0)
	for _, depID := range task.DependsOn {
		dep, err := r.bolt.GetTask(depID)
		if err != nil {
			continue
		}
		if dep.Status != RegStatusCompleted {
			blockedBy = append(blockedBy, depID)
		}
	}

	task.BlockedBy = blockedBy
	return r.bolt.UpdateTask(task)
}

// updateChangeSummaries is deprecated - BoltDB stores changes separately
// Kept for backward compatibility during migration
func (r *RegistryStore) updateChangeSummaries(reg *Registry) {
	// BoltDB handles change summaries automatically
	// This method is now a no-op
}

// changeProgress calculates completion percentage for a change
func (r *RegistryStore) changeProgress(changeID string) (float64, error) {
	changes, err := r.bolt.ListChanges()
	if err != nil {
		return 0, err
	}

	for _, change := range changes {
		if change.ID == changeID {
			if change.Total == 0 {
				return 0, nil
			}
			return float64(change.Completed) / float64(change.Total), nil
		}
	}

	return 0, fmt.Errorf("change %s not found", changeID)
}

// hasCycle, detectCycle - deprecated, BoltDB handles cycle detection in AddDependency
// Kept for backward compatibility during migration but are no-ops

func priorityValue(p TaskPriority) int {
	switch p {
	case PriorityCritical:
		return 1
	case PriorityHigh:
		return 2
	case PriorityMedium:
		return 3
	case PriorityLow:
		return 4
	case PriorityBacklog:
		return 5
	default:
		return 6
	}
}
