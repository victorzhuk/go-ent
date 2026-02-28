package spec

import (
	"time"
)

type ChangeMetadata struct {
	ID         string       `json:"id"`
	Title      string       `json:"title"`
	Status     ChangeStatus `json:"status"`
	CreatedAt  time.Time    `json:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at"`
	TaskCount  int          `json:"task_count"`
	Completed  int          `json:"completed"`
	InProgress int          `json:"in_progress"`
	Blocked    int          `json:"blocked"`
}

type Task struct {
	ID         string     `json:"id"`
	ChangeID   string     `json:"change_id"`
	TaskNum    string     `json:"task_num"`
	Content    string     `json:"content"`
	Status     TaskStatus `json:"status"`
	Priority   int        `json:"priority"`
	DependsOn  []string   `json:"depends_on"`
	SourceLine int        `json:"source_line"`
	SyncedAt   time.Time  `json:"synced_at"`
}

func (t *Task) StatusIcon() string {
	return StatusIconForStatus(t.Status)
}

func StatusIconForStatus(status TaskStatus) string {
	switch status {
	case TaskCompleted:
		return "✅"
	case TaskInProgress:
		return "🔄"
	default:
		return "⏳"
	}
}

type DependencyInfo struct {
	TaskID        string   `json:"task_id"`
	DependsOn     []string `json:"depends_on"`
	DependedBy    []string `json:"depended_by"`
	IsBlocked     bool     `json:"is_blocked"`
	BlockingTasks []string `json:"blocking_tasks"`
}

type RuntimeState struct {
	TaskID    string     `json:"task_id"`
	Status    TaskStatus `json:"status"`
	Assignee  string     `json:"assignee"`
	Session   string     `json:"session"`
	StartedAt time.Time  `json:"started_at"`
	Notes     []string   `json:"notes"`
}

type SyncMeta struct {
	Version      int              `json:"version"`
	LastFullSync time.Time        `json:"last_full_sync"`
	FileMTimes   map[string]int64 `json:"file_mtimes"`
}
