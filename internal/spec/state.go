package spec

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var depCommentRegex = regexp.MustCompile(`<!--\s*depends:\s*(.+?)\s*-->`)

type StateStore struct {
	store *Store
	bolt  *BoltStore
}

func NewStateStore(store *Store, bolt *BoltStore) *StateStore {
	return &StateStore{
		store: store,
		bolt:  bolt,
	}
}

type ChangeState struct {
	ID             string
	Title          string
	Progress       ProgressInfo
	CurrentTask    *TaskInfo
	Blockers       []TaskInfo
	RecentActivity []ActivityInfo
	Updated        time.Time
}

type ProgressInfo struct {
	Completed int
	Total     int
	Percent   int
}

type TaskInfo struct {
	ID      TaskID
	Content string
	Section string
	Line    int
	Status  RegistryTaskStatus
}

type ActivityInfo struct {
	TaskID TaskID
	Action string
	Time   time.Time
}

type RootState struct {
	ActiveChanges    []ChangeSummary
	RecommendedTasks []RegistryTask
	Updated          time.Time
}

func (s *StateStore) GenerateChangeState(changeID string) (*ChangeState, error) {
	summary, err := s.bolt.GetChangeSummary(changeID)
	if err != nil {
		return nil, fmt.Errorf("get change summary: %w", err)
	}

	tasks, err := s.bolt.ListTasks(TaskFilter{ChangeID: changeID})
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}

	state := &ChangeState{
		ID:      changeID,
		Title:   summary.Title,
		Updated: time.Now(),
		Progress: ProgressInfo{
			Total:     summary.Total,
			Completed: summary.Completed,
			Percent:   0,
		},
	}

	if summary.Total > 0 {
		state.Progress.Percent = (summary.Completed * 100) / summary.Total
	}

	for i := range tasks {
		if tasks[i].Status == RegStatusInProgress ||
			(state.CurrentTask == nil && tasks[i].Status == RegStatusPending && len(tasks[i].BlockedBy) == 0) {
			state.CurrentTask = &TaskInfo{
				ID:      tasks[i].ID,
				Content: tasks[i].Content,
				Line:    tasks[i].SourceLine,
				Status:  tasks[i].Status,
			}
			break
		}
	}

	for i := range tasks {
		if tasks[i].Status == RegStatusPending && len(tasks[i].BlockedBy) > 0 {
			state.Blockers = append(state.Blockers, TaskInfo{
				ID:      tasks[i].ID,
				Content: tasks[i].Content,
				Status:  tasks[i].Status,
			})
		}
	}

	for i := len(tasks) - 1; i >= 0 && len(state.RecentActivity) < 5; i-- {
		if tasks[i].CompletedAt != nil {
			state.RecentActivity = append(state.RecentActivity, ActivityInfo{
				TaskID: tasks[i].ID,
				Action: "completed",
				Time:   *tasks[i].CompletedAt,
			})
		}
	}

	return state, nil
}

func (s *StateStore) GenerateRootState() (*RootState, error) {
	changes, err := s.bolt.ListChanges()
	if err != nil {
		return nil, fmt.Errorf("list changes: %w", err)
	}

	next, err := s.bolt.NextTasks(5)
	if err != nil {
		return nil, fmt.Errorf("get next tasks: %w", err)
	}

	return &RootState{
		ActiveChanges:    changes,
		RecommendedTasks: next,
		Updated:          time.Now(),
	}, nil
}

func (s *StateStore) WriteChangeStateMd(changeID string, outputPath string) error {
	state, err := s.GenerateChangeState(changeID)
	if err != nil {
		return err
	}

	f, err := os.Create(outputPath) // #nosec G304 -- controlled file path
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer func() { _ = f.Close() }()

	_, _ = fmt.Fprintf(f, "# State: %s\n\n", changeID)
	_, _ = fmt.Fprintf(f, "> Updated: %s\n\n", state.Updated.Format(time.RFC3339))

	_, _ = fmt.Fprintf(f, "## Progress\n")
	_, _ = fmt.Fprintf(f, "%d/%d complete (%d%%)\n\n",
		state.Progress.Completed, state.Progress.Total, state.Progress.Percent)

	if state.CurrentTask != nil {
		_, _ = fmt.Fprintf(f, "## Current Task\n")
		_, _ = fmt.Fprintf(f, "**T%s**: %s\n", state.CurrentTask.ID.TaskNum, state.CurrentTask.Content)
		_, _ = fmt.Fprintf(f, "- Line: tasks.md:%d\n", state.CurrentTask.Line)
		_, _ = fmt.Fprintf(f, "- Status: %s\n\n", state.CurrentTask.Status)
	} else {
		_, _ = fmt.Fprintf(f, "## Current Task\nNone (all tasks complete or blocked)\n\n")
	}

	_, _ = fmt.Fprintf(f, "## Blockers\n")
	if len(state.Blockers) == 0 {
		_, _ = fmt.Fprintf(f, "None\n\n")
	} else {
		for _, blocker := range state.Blockers {
			_, _ = fmt.Fprintf(f, "- **T%s**: %s\n", blocker.ID.TaskNum, blocker.Content)
		}
		_, _ = fmt.Fprintf(f, "\n")
	}

	_, _ = fmt.Fprintf(f, "## Recent Activity\n")
	if len(state.RecentActivity) == 0 {
		_, _ = fmt.Fprintf(f, "No recent activity\n\n")
	} else {
		_, _ = fmt.Fprintf(f, "| Task | Action | Time |\n")
		_, _ = fmt.Fprintf(f, "|------|--------|------|\n")
		for _, activity := range state.RecentActivity {
			_, _ = fmt.Fprintf(f, "| T%s | %s | %s |\n",
				activity.TaskID.TaskNum,
				activity.Action,
				activity.Time.Format("15:04"))
		}
		_, _ = fmt.Fprintf(f, "\n")
	}

	return nil
}

func (s *StateStore) WriteRootStateMd(outputPath string) error {
	state, err := s.GenerateRootState()
	if err != nil {
		return err
	}

	f, err := os.Create(outputPath) // #nosec G304 -- controlled file path
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer func() { _ = f.Close() }()

	_, _ = fmt.Fprintf(f, "# OpenSpec State\n\n")
	_, _ = fmt.Fprintf(f, "> Updated: %s\n\n", state.Updated.Format(time.RFC3339))

	_, _ = fmt.Fprintf(f, "## Active Changes\n\n")
	_, _ = fmt.Fprintf(f, "| Change | Progress | Blocked |\n")
	_, _ = fmt.Fprintf(f, "|--------|----------|---------||\n")
	for _, change := range state.ActiveChanges {
		percent := 0
		if change.Total > 0 {
			percent = (change.Completed * 100) / change.Total
		}
		_, _ = fmt.Fprintf(f, "| %s | %d%% (%d/%d) | %d |\n",
			change.ID, percent, change.Completed, change.Total, change.Blocked)
	}
	_, _ = fmt.Fprintf(f, "\n")

	_, _ = fmt.Fprintf(f, "## Recommended Next\n")
	if len(state.RecommendedTasks) == 0 {
		_, _ = fmt.Fprintf(f, "No unblocked tasks available\n\n")
	} else {
		for i, task := range state.RecommendedTasks {
			_, _ = fmt.Fprintf(f, "%d. **%s** - %s (%s priority)\n",
				i+1, task.ID.String(), task.Content, task.Priority)
		}
		_, _ = fmt.Fprintf(f, "\n")
	}

	return nil
}

func ParseDependencies(line string) []string {
	matches := depCommentRegex.FindStringSubmatch(line)
	if len(matches) < 2 {
		return nil
	}

	depsStr := strings.TrimSpace(matches[1])
	parts := strings.Split(depsStr, ",")
	deps := make([]string, 0, len(parts))
	for _, part := range parts {
		dep := strings.TrimSpace(part)
		if dep != "" {
			deps = append(deps, dep)
		}
	}
	return deps
}

func (s *StateStore) ParseTasksWithDependencies(changeID string, tasksPath string) ([]RegistryTask, error) {
	f, err := os.Open(tasksPath) // #nosec G304 -- controlled config/template file path
	if err != nil {
		return nil, fmt.Errorf("open tasks.md: %w", err)
	}
	defer func() { _ = f.Close() }()

	var tasks []RegistryTask
	scanner := bufio.NewScanner(f)
	lineNum := 0
	taskNum := 1

	taskPattern := regexp.MustCompile(`^[-*]\s+\[([ xX])\]\s+(.+)$`)

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		matches := taskPattern.FindStringSubmatch(line)
		if len(matches) < 3 {
			continue
		}

		checked := matches[1]
		content := strings.TrimSpace(matches[2])

		depStr := ""
		if idx := strings.Index(content, "<!--"); idx != -1 {
			depStr = content[idx:]
			content = strings.TrimSpace(content[:idx])
		}

		status := RegStatusPending
		if checked == "x" || checked == "X" {
			status = RegStatusCompleted
		}

		task := RegistryTask{
			ID: TaskID{
				ChangeID: changeID,
				TaskNum:  fmt.Sprintf("%d", taskNum),
			},
			Content:    content,
			Status:     status,
			Priority:   PriorityMedium,
			SourceLine: lineNum,
			SyncedAt:   time.Now(),
		}

		deps := ParseDependencies(depStr)
		for _, dep := range deps {
			depID := TaskID{
				ChangeID: changeID,
				TaskNum:  dep,
			}
			task.DependsOn = append(task.DependsOn, depID)
		}

		tasks = append(tasks, task)
		taskNum++
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan file: %w", err)
	}

	return tasks, nil
}

func (s *StateStore) SyncFromTasksMd() error {
	changesDir := filepath.Join(s.store.RootPath(), "openspec", "changes")
	entries, err := os.ReadDir(changesDir)
	if err != nil {
		return fmt.Errorf("read changes dir: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == "archive" {
			continue
		}

		changeID := entry.Name()
		tasksPath := filepath.Join(changesDir, changeID, "tasks.md")
		if _, err := os.Stat(tasksPath); os.IsNotExist(err) {
			continue
		}

		tasks, err := s.ParseTasksWithDependencies(changeID, tasksPath)
		if err != nil {
			return fmt.Errorf("parse tasks for %s: %w", changeID, err)
		}

		for i := range tasks {
			if err := s.bolt.UpdateTask(&tasks[i]); err != nil {
				return fmt.Errorf("update task %s: %w", tasks[i].ID, err)
			}

			for _, dep := range tasks[i].DependsOn {
				if err := s.bolt.AddDependency(tasks[i].ID, dep); err != nil {
					return fmt.Errorf("add dependency %s->%s: %w", tasks[i].ID, dep, err)
				}
			}
		}
	}

	if err := s.bolt.SetSyncedAt(time.Now()); err != nil {
		return fmt.Errorf("set synced_at: %w", err)
	}

	return nil
}

func (s *StateStore) UpdateTaskInFile(taskID TaskID, status RegistryTaskStatus, notes string) error {
	changeDir := filepath.Join(s.store.SpecPath(), "changes", taskID.ChangeID)
	tasksPath := filepath.Join(changeDir, "tasks.md")

	task, err := s.bolt.GetTask(taskID)
	if err != nil {
		return fmt.Errorf("get task from bolt: %w", err)
	}

	lines, err := s.readFileLines(tasksPath)
	if err != nil {
		return fmt.Errorf("read tasks file: %w", err)
	}

	if task.SourceLine < 1 || task.SourceLine > len(lines) {
		return fmt.Errorf("invalid task line: %d", task.SourceLine)
	}

	updatedLine, err := s.updateTaskLine(lines[task.SourceLine-1], status, notes)
	if err != nil {
		return fmt.Errorf("update task line: %w", err)
	}

	lines[task.SourceLine-1] = updatedLine

	if err := s.writeFileLines(tasksPath, lines); err != nil {
		return fmt.Errorf("write tasks file: %w", err)
	}

	return nil
}

func (s *StateStore) readFileLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}

func (s *StateStore) writeFileLines(path string, lines []string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	for _, line := range lines {
		if _, err := fmt.Fprintln(f, line); err != nil {
			return err
		}
	}

	return nil
}

func (s *StateStore) updateTaskLine(line string, status RegistryTaskStatus, notes string) (string, error) {
	taskPattern := regexp.MustCompile(`^([-*])\s+\[([ xX])\]\s+(.+)$`)
	matches := taskPattern.FindStringSubmatch(line)
	if len(matches) < 4 {
		return line, fmt.Errorf("line does not match task pattern")
	}

	bullet := matches[1]
	content := strings.TrimSpace(matches[3])

	depStr := ""
	if idx := strings.Index(content, "<!--"); idx != -1 {
		depStr = content[idx:]
		content = strings.TrimSpace(content[:idx])
	}

	newChecked := " "
	if status == RegStatusCompleted {
		newChecked = "x"
	}

	if notes != "" {
		trimmedContent := strings.TrimSpace(content)
		existingIdx := strings.Index(trimmedContent, "✓")
		if existingIdx != -1 {
			trimmedContent = strings.TrimSpace(trimmedContent[:existingIdx])
		}
		content = fmt.Sprintf("%s %s", trimmedContent, notes)
	}

	return fmt.Sprintf("%s [%s] %s%s", bullet, newChecked, content, depStr), nil
}
