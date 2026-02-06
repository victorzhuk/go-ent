package parser

import (
	"fmt"
	"strings"
	"time"
	"unicode"
)

type Task struct {
	ID         string
	ChangeID   string
	TaskNum    string
	Content    string
	Status     TaskStatus
	Priority   int
	DependsOn  []string
	SourceLine int
	SyncedAt   time.Time
}

type TaskStatus string

const (
	TaskPending    TaskStatus = "pending"
	TaskInProgress TaskStatus = "in_progress"
	TaskCompleted  TaskStatus = "completed"
)

func (s TaskStatus) Valid() bool {
	switch s {
	case TaskPending, TaskInProgress, TaskCompleted:
		return true
	default:
		return false
	}
}

func (t Task) Validate() error {
	if t.TaskNum == "" {
		return ErrMissingTaskNum
	}

	if !isValidTaskID(t.TaskNum) {
		return fmt.Errorf("%w: %s", ErrInvalidTaskID, t.TaskNum)
	}

	if !t.Status.Valid() {
		return fmt.Errorf("%w: %s", ErrInvalidStatus, t.Status)
	}

	for _, dep := range t.DependsOn {
		if !isValidTaskID(dep) {
			return fmt.Errorf("%w: %s", ErrInvalidDepends, dep)
		}
	}

	return nil
}

func isValidTaskID(id string) bool {
	if id == "" {
		return false
	}

	for _, r := range id {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '.' && r != '-' {
			return false
		}
	}

	return true
}

func NormalizeStatus(status string) (TaskStatus, error) {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "pending", "todo":
		return TaskPending, nil
	case "in_progress", "doing", "active":
		return TaskInProgress, nil
	case "completed", "done", "finished":
		return TaskCompleted, nil
	default:
		return "", ErrInvalidStatus
	}
}
