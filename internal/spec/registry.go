package spec

import "time"

type TaskPriority string

const (
	PriorityCritical TaskPriority = "critical"
	PriorityHigh     TaskPriority = "high"
	PriorityMedium   TaskPriority = "medium"
	PriorityLow      TaskPriority = "low"
	PriorityBacklog  TaskPriority = "backlog"
)

type RegistryTaskStatus string

const (
	RegStatusPending    RegistryTaskStatus = "pending"
	RegStatusInProgress RegistryTaskStatus = "in_progress"
	RegStatusCompleted  RegistryTaskStatus = "completed"
	RegStatusBlocked    RegistryTaskStatus = "blocked"
	RegStatusSkipped    RegistryTaskStatus = "skipped"
)

type TaskID struct {
	ChangeID string `yaml:"change_id" json:"change_id"`
	TaskNum  string `yaml:"task_num" json:"task_num"`
}

func (t TaskID) String() string {
	if t.ChangeID == "" || t.TaskNum == "" {
		return ""
	}
	return t.ChangeID + "/" + t.TaskNum
}

func (t TaskID) IsZero() bool {
	return t.ChangeID == "" && t.TaskNum == ""
}

type RegistryTask struct {
	ID          TaskID             `yaml:"id" json:"id"`
	Content     string             `yaml:"content" json:"content"`
	Status      RegistryTaskStatus `yaml:"status" json:"status"`
	Priority    TaskPriority       `yaml:"priority" json:"priority"`
	DependsOn   []TaskID           `yaml:"depends_on,omitempty" json:"depends_on,omitempty"`
	BlockedBy   []TaskID           `yaml:"blocked_by,omitempty" json:"blocked_by,omitempty"`
	Assignee    string             `yaml:"assignee,omitempty" json:"assignee,omitempty"`
	Session     string             `yaml:"session,omitempty" json:"session,omitempty"`
	StartedAt   *time.Time         `yaml:"started_at,omitempty" json:"started_at,omitempty"`
	CompletedAt *time.Time         `yaml:"completed_at,omitempty" json:"completed_at,omitempty"`
	Notes       string             `yaml:"notes,omitempty" json:"notes,omitempty"`
	SourceLine  int                `yaml:"source_line" json:"source_line"`
	SyncedAt    time.Time          `yaml:"synced_at" json:"synced_at"`
}

type ChangeSummary struct {
	ID         string       `yaml:"id" json:"id"`
	Title      string       `yaml:"title" json:"title"`
	Status     ChangeStatus `yaml:"status" json:"status"`
	TasksFile  string       `yaml:"tasks_file" json:"tasks_file"`
	Total      int          `yaml:"total" json:"total"`
	Completed  int          `yaml:"completed" json:"completed"`
	InProgress int          `yaml:"in_progress" json:"in_progress"`
	Blocked    int          `yaml:"blocked" json:"blocked"`
}

type Registry struct {
	Version  string                   `yaml:"version" json:"version"`
	SyncedAt time.Time                `yaml:"synced_at" json:"synced_at"`
	Changes  map[string]ChangeSummary `yaml:"changes" json:"changes"`
	Tasks    []RegistryTask           `yaml:"tasks" json:"tasks"`
}

type NextTaskResult struct {
	Recommended  *RegistryTask  `json:"recommended,omitempty"`
	Reason       string         `json:"reason,omitempty"`
	Alternatives []RegistryTask `json:"alternatives,omitempty"`
	BlockedCount int            `json:"blocked_count"`
}

type TaskFilter struct {
	ChangeID  string
	Status    RegistryTaskStatus
	Priority  TaskPriority
	Assignee  string
	Unblocked bool
}

type TaskUpdate struct {
	Status   *RegistryTaskStatus
	Priority *TaskPriority
	Assignee *string
	Notes    *string
}

type SyncResult struct {
	SyncedChanges []string       `json:"synced_changes"`
	Added         []TaskID       `json:"added"`
	Updated       []TaskID       `json:"updated"`
	Removed       []TaskID       `json:"removed"`
	Conflicts     []SyncConflict `json:"conflicts"`
}

type SyncConflict struct {
	TaskID      TaskID `json:"task_id"`
	Field       string `json:"field"`
	SourceValue string `json:"source_value"`
	RegistryVal string `json:"registry_value"`
	Resolution  string `json:"resolution"`
}

type DependencyGraph struct {
	TaskID        TaskID               `json:"task_id"`
	DependsOn     []TaskDependencyInfo `json:"depends_on"`
	DependedBy    []TaskDependencyInfo `json:"depended_by"`
	IsBlocked     bool                 `json:"is_blocked"`
	BlockingTasks []TaskID             `json:"blocking_tasks,omitempty"`
}

type TaskDependencyInfo struct {
	ID      TaskID             `json:"id"`
	Content string             `json:"content"`
	Status  RegistryTaskStatus `json:"status"`
}

type RegistryStats struct {
	TotalTasks   int            `json:"total_tasks"`
	ByStatus     map[string]int `json:"by_status"`
	ByPriority   map[string]int `json:"by_priority"`
	ByChange     map[string]int `json:"by_change"`
	BlockedCount int            `json:"blocked_count"`
	NextTask     *RegistryTask  `json:"next_task,omitempty"`
}
