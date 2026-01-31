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

// StateStore provides parsing and updating of tasks.md files
type StateStore struct {
	store *Store
}

func NewStateStore(store *Store) *StateStore {
	return &StateStore{
		store: store,
	}
}

// Removed: ChangeState, ProgressInfo, TaskInfo, ActivityInfo, RootState
// These were used for BoltDB-backed state generation
// If needed, recreate as simple structs for tasks.md parsing

// Removed: GenerateChangeState, GenerateRootState, WriteChangeStateMd, WriteRootStateMd, SyncFromTasksMd
// These methods relied on BoltDB for state management
// For state generation, parse tasks.md directly using ParseTasksWithDependencies

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

// ParseDependencies extracts dependency task numbers from a comment string
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

// UpdateTaskInFile updates a task's status in tasks.md by line number
// Note: This is a simplified version without BoltDB dependency
func (s *StateStore) UpdateTaskInFile(changeID string, lineNum int, status RegistryTaskStatus, notes string) error {
	changeDir := filepath.Join(s.store.SpecPath(), "changes", changeID)
	tasksPath := filepath.Join(changeDir, "tasks.md")

	lines, err := s.readFileLines(tasksPath)
	if err != nil {
		return fmt.Errorf("read tasks file: %w", err)
	}

	if lineNum < 1 || lineNum > len(lines) {
		return fmt.Errorf("invalid line number: %d", lineNum)
	}

	updatedLine, err := s.updateTaskLine(lines[lineNum-1], status, notes)
	if err != nil {
		return fmt.Errorf("update task line: %w", err)
	}

	lines[lineNum-1] = updatedLine

	if err := s.writeFileLines(tasksPath, lines); err != nil {
		return fmt.Errorf("write tasks file: %w", err)
	}

	return nil
}

func (s *StateStore) readFileLines(path string) ([]string, error) {
	path = filepath.Clean(path)
	f, err := os.Open(path) //nolint:gosec
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
	path = filepath.Clean(path)
	f, err := os.Create(path) //nolint:gosec
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
