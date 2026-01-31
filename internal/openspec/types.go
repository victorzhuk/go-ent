package openspec

// ListItem represents a change or spec in list output.
type ListItem struct {
	Name           string `json:"name"`
	CompletedTasks int    `json:"completedTasks,omitempty"`
	TotalTasks     int    `json:"totalTasks,omitempty"`
	LastModified   string `json:"lastModified,omitempty"`
	Status         string `json:"status,omitempty"`
	Description    string `json:"description,omitempty"`
}

// ListResponse wraps the list output from openspec CLI.
type ListResponse struct {
	Changes []ListItem `json:"changes,omitempty"`
	Specs   []ListItem `json:"specs,omitempty"`
}

// ChangeStatus represents the status of a change.
type ChangeStatus string

const (
	StatusDraft    ChangeStatus = "draft"
	StatusActive   ChangeStatus = "active"
	StatusApproved ChangeStatus = "approved"
	StatusArchived ChangeStatus = "archived"
)

// TaskStatus represents the status of a task.
type TaskStatus string

const (
	TaskPending    TaskStatus = "pending"
	TaskInProgress TaskStatus = "in_progress"
	TaskCompleted  TaskStatus = "completed"
)

// DeltaOperation represents a delta spec operation.
type DeltaOperation string

const (
	OpAdded    DeltaOperation = "ADDED"
	OpModified DeltaOperation = "MODIFIED"
	OpRemoved  DeltaOperation = "REMOVED"
	OpRenamed  DeltaOperation = "RENAMED"
)
