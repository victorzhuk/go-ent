package spec

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

type RegistryStore struct {
	store *Store
}

func NewRegistryStore(store *Store) *RegistryStore {
	return &RegistryStore{store: store}
}

func (r *RegistryStore) Load() (*Registry, error) {
	return r.store.LoadRegistry()
}

func (r *RegistryStore) Save(reg *Registry) error {
	reg.SyncedAt = time.Now()
	return r.store.SaveRegistry(reg)
}

func (r *RegistryStore) Exists() bool {
	return r.store.RegistryExists()
}

func (r *RegistryStore) Init() error {
	reg := &Registry{
		Version:  "1.0",
		SyncedAt: time.Now(),
		Changes:  make(map[string]ChangeSummary),
		Tasks:    []RegistryTask{},
	}
	return r.Save(reg)
}

func (r *RegistryStore) ListTasks(filter TaskFilter) ([]RegistryTask, error) {
	reg, err := r.Load()
	if err != nil {
		return nil, err
	}

	tasks := make([]RegistryTask, 0)
	for _, task := range reg.Tasks {
		if filter.ChangeID != "" && task.ID.ChangeID != filter.ChangeID {
			continue
		}
		if filter.Status != "" && task.Status != filter.Status {
			continue
		}
		if filter.Priority != "" && task.Priority != filter.Priority {
			continue
		}
		if filter.Assignee != "" && task.Assignee != filter.Assignee {
			continue
		}
		if filter.Unblocked && len(task.BlockedBy) > 0 {
			continue
		}

		tasks = append(tasks, task)
	}

	if filter.Limit > 0 && len(tasks) > filter.Limit {
		tasks = tasks[:filter.Limit]
	}

	return tasks, nil
}

func (r *RegistryStore) GetTask(id TaskID) (*RegistryTask, error) {
	reg, err := r.Load()
	if err != nil {
		return nil, err
	}

	for i := range reg.Tasks {
		if reg.Tasks[i].ID.ChangeID == id.ChangeID && reg.Tasks[i].ID.TaskNum == id.TaskNum {
			return &reg.Tasks[i], nil
		}
	}

	return nil, fmt.Errorf("task %s not found", id.String())
}

func (r *RegistryStore) UpdateTask(id TaskID, updates TaskUpdate) error {
	reg, err := r.Load()
	if err != nil {
		return err
	}

	var found bool
	for i := range reg.Tasks {
		if reg.Tasks[i].ID.ChangeID == id.ChangeID && reg.Tasks[i].ID.TaskNum == id.TaskNum {
			found = true
			now := time.Now()

			if updates.Status != nil {
				oldStatus := reg.Tasks[i].Status
				reg.Tasks[i].Status = *updates.Status

				if *updates.Status == RegStatusInProgress && oldStatus != RegStatusInProgress {
					reg.Tasks[i].StartedAt = &now
				}
				if *updates.Status == RegStatusCompleted && oldStatus != RegStatusCompleted {
					reg.Tasks[i].CompletedAt = &now
				}
			}
			if updates.Priority != nil {
				reg.Tasks[i].Priority = *updates.Priority
			}
			if updates.Assignee != nil {
				reg.Tasks[i].Assignee = *updates.Assignee
			}
			if updates.Notes != nil {
				reg.Tasks[i].Notes = *updates.Notes
			}
			break
		}
	}

	if !found {
		return fmt.Errorf("task %s not found", id.String())
	}

	r.recalculateBlockedBy(reg)
	r.updateChangeSummaries(reg)

	return r.Save(reg)
}

func (r *RegistryStore) NextTask(count int) (*NextTaskResult, error) {
	if count <= 0 {
		count = 1
	}

	reg, err := r.Load()
	if err != nil {
		return nil, err
	}

	r.recalculateBlockedBy(reg)

	candidates := make([]RegistryTask, 0)
	blockedCount := 0

	for _, task := range reg.Tasks {
		if task.Status == RegStatusPending && len(task.BlockedBy) == 0 {
			candidates = append(candidates, task)
		}
		if len(task.BlockedBy) > 0 {
			blockedCount++
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Priority != candidates[j].Priority {
			return priorityValue(candidates[i].Priority) < priorityValue(candidates[j].Priority)
		}

		progressI := r.changeProgress(reg, candidates[i].ID.ChangeID)
		progressJ := r.changeProgress(reg, candidates[j].ID.ChangeID)
		if progressI != progressJ {
			return progressI > progressJ
		}

		return candidates[i].ID.TaskNum < candidates[j].ID.TaskNum
	})

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
			if len(candidates) < count {
				limit = len(candidates)
			}
			result.Alternatives = candidates[1:limit]
		}
	}

	return result, nil
}

func (r *RegistryStore) AddDependency(task, dependsOn TaskID) error {
	reg, err := r.Load()
	if err != nil {
		return err
	}

	var taskFound, depFound bool
	var taskIdx int

	for i := range reg.Tasks {
		if reg.Tasks[i].ID.ChangeID == task.ChangeID && reg.Tasks[i].ID.TaskNum == task.TaskNum {
			taskFound = true
			taskIdx = i
		}
		if reg.Tasks[i].ID.ChangeID == dependsOn.ChangeID && reg.Tasks[i].ID.TaskNum == dependsOn.TaskNum {
			depFound = true
		}
	}

	if !taskFound {
		return fmt.Errorf("task %s not found", task.String())
	}
	if !depFound {
		return fmt.Errorf("dependency task %s not found", dependsOn.String())
	}

	for _, existing := range reg.Tasks[taskIdx].DependsOn {
		if existing.ChangeID == dependsOn.ChangeID && existing.TaskNum == dependsOn.TaskNum {
			return nil
		}
	}

	reg.Tasks[taskIdx].DependsOn = append(reg.Tasks[taskIdx].DependsOn, dependsOn)

	if r.hasCycle(reg, task) {
		return fmt.Errorf("adding dependency would create a cycle")
	}

	r.recalculateBlockedBy(reg)

	return r.Save(reg)
}

func (r *RegistryStore) RemoveDependency(task, dependsOn TaskID) error {
	reg, err := r.Load()
	if err != nil {
		return err
	}

	var found bool
	for i := range reg.Tasks {
		if reg.Tasks[i].ID.ChangeID == task.ChangeID && reg.Tasks[i].ID.TaskNum == task.TaskNum {
			found = true
			newDeps := make([]TaskID, 0)
			for _, dep := range reg.Tasks[i].DependsOn {
				if dep.ChangeID != dependsOn.ChangeID || dep.TaskNum != dependsOn.TaskNum {
					newDeps = append(newDeps, dep)
				}
			}
			reg.Tasks[i].DependsOn = newDeps
			break
		}
	}

	if !found {
		return fmt.Errorf("task %s not found", task.String())
	}

	r.recalculateBlockedBy(reg)

	return r.Save(reg)
}

func (r *RegistryStore) GetDependencyGraph(id TaskID) (*DependencyGraph, error) {
	reg, err := r.Load()
	if err != nil {
		return nil, err
	}

	task, err := r.GetTask(id)
	if err != nil {
		return nil, err
	}

	r.recalculateBlockedBy(reg)

	graph := &DependencyGraph{
		TaskID:        id,
		DependsOn:     make([]TaskDependencyInfo, 0),
		DependedBy:    make([]TaskDependencyInfo, 0),
		IsBlocked:     len(task.BlockedBy) > 0,
		BlockingTasks: make([]TaskID, 0),
	}

	for _, depID := range task.DependsOn {
		if dep, err := r.GetTask(depID); err == nil {
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

	for i := range reg.Tasks {
		for _, depID := range reg.Tasks[i].DependsOn {
			if depID.ChangeID == id.ChangeID && depID.TaskNum == id.TaskNum {
				graph.DependedBy = append(graph.DependedBy, TaskDependencyInfo{
					ID:      reg.Tasks[i].ID,
					Content: reg.Tasks[i].Content,
					Status:  reg.Tasks[i].Status,
				})
			}
		}
	}

	return graph, nil
}

func (r *RegistryStore) RebuildFromSource() (*SyncResult, error) {
	reg := &Registry{
		Version:  "1.0",
		SyncedAt: time.Now(),
		Changes:  make(map[string]ChangeSummary),
		Tasks:    make([]RegistryTask, 0),
	}

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
			return result, r.Save(reg)
		}
		return nil, fmt.Errorf("read changes dir: %w", err)
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

		tasks, err := r.parseTasksFile(tasksPath, changeID)
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

			reg.Tasks = append(reg.Tasks, task)
			result.Added = append(result.Added, task.ID)
		}

		reg.Changes[changeID] = summary
		result.SyncedChanges = append(result.SyncedChanges, changeID)
	}

	r.recalculateBlockedBy(reg)

	return result, r.Save(reg)
}

func (r *RegistryStore) Stats() (*RegistryStats, error) {
	reg, err := r.Load()
	if err != nil {
		return nil, err
	}

	stats := &RegistryStats{
		TotalTasks: len(reg.Tasks),
		ByStatus:   make(map[string]int),
		ByPriority: make(map[string]int),
		ByChange:   make(map[string]int),
	}

	for _, task := range reg.Tasks {
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

func (r *RegistryStore) parseTasksFile(path, changeID string) ([]RegistryTask, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read tasks file: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	tasks := make([]RegistryTask, 0)

	taskPattern := regexp.MustCompile(`^[-*]\s+\[([ xX])\]\s+(.+)$`)
	taskCounter := 0

	for lineNum, line := range lines {
		line = strings.TrimSpace(line)
		matches := taskPattern.FindStringSubmatch(line)
		if len(matches) != 3 {
			continue
		}

		taskCounter++
		checked := matches[1] == "x" || matches[1] == "X"
		content := matches[2]

		status := RegStatusPending
		if checked {
			status = RegStatusCompleted
		}

		task := RegistryTask{
			ID: TaskID{
				ChangeID: changeID,
				TaskNum:  fmt.Sprintf("%d", taskCounter),
			},
			Content:    content,
			Status:     status,
			Priority:   PriorityMedium,
			DependsOn:  []TaskID{},
			SourceLine: lineNum + 1,
			SyncedAt:   time.Now(),
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (r *RegistryStore) recalculateBlockedBy(reg *Registry) {
	for i := range reg.Tasks {
		reg.Tasks[i].BlockedBy = []TaskID{}
	}

	for i := range reg.Tasks {
		for _, depID := range reg.Tasks[i].DependsOn {
			var depCompleted bool
			for j := range reg.Tasks {
				if reg.Tasks[j].ID.ChangeID == depID.ChangeID && reg.Tasks[j].ID.TaskNum == depID.TaskNum {
					if reg.Tasks[j].Status == RegStatusCompleted {
						depCompleted = true
					}
					break
				}
			}

			if !depCompleted {
				reg.Tasks[i].BlockedBy = append(reg.Tasks[i].BlockedBy, depID)
			}
		}
	}
}

func (r *RegistryStore) updateChangeSummaries(reg *Registry) {
	for changeID := range reg.Changes {
		summary := reg.Changes[changeID]
		summary.Total = 0
		summary.Completed = 0
		summary.InProgress = 0
		summary.Blocked = 0

		for _, task := range reg.Tasks {
			if task.ID.ChangeID == changeID {
				summary.Total++
				if task.Status == RegStatusCompleted {
					summary.Completed++
				} else if task.Status == RegStatusInProgress {
					summary.InProgress++
				}
				if len(task.BlockedBy) > 0 {
					summary.Blocked++
				}
			}
		}

		reg.Changes[changeID] = summary
	}
}

func (r *RegistryStore) changeProgress(reg *Registry, changeID string) float64 {
	summary, ok := reg.Changes[changeID]
	if !ok || summary.Total == 0 {
		return 0
	}
	return float64(summary.Completed) / float64(summary.Total)
}

func (r *RegistryStore) hasCycle(reg *Registry, start TaskID) bool {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	return r.detectCycle(reg, start, visited, recStack)
}

func (r *RegistryStore) detectCycle(reg *Registry, current TaskID, visited, recStack map[string]bool) bool {
	key := current.String()
	visited[key] = true
	recStack[key] = true

	for i := range reg.Tasks {
		if reg.Tasks[i].ID.ChangeID == current.ChangeID && reg.Tasks[i].ID.TaskNum == current.TaskNum {
			for _, dep := range reg.Tasks[i].DependsOn {
				depKey := dep.String()
				if !visited[depKey] {
					if r.detectCycle(reg, dep, visited, recStack) {
						return true
					}
				} else if recStack[depKey] {
					return true
				}
			}
		}
	}

	recStack[key] = false
	return false
}

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
