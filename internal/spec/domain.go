package spec

type ChangeStatus string

const (
	StatusDraft    ChangeStatus = "draft"
	StatusActive   ChangeStatus = "active"
	StatusApproved ChangeStatus = "approved"
	StatusArchived ChangeStatus = "archived"
)

type TaskStatus string

const (
	TaskPending    TaskStatus = "pending"
	TaskInProgress TaskStatus = "in_progress"
	TaskCompleted  TaskStatus = "completed"
)

type DeltaOperation string

const (
	OpAdded    DeltaOperation = "ADDED"
	OpModified DeltaOperation = "MODIFIED"
	OpRemoved  DeltaOperation = "REMOVED"
	OpRenamed  DeltaOperation = "RENAMED"
)

type Project struct {
	Name        string            `yaml:"name"`
	Module      string            `yaml:"module"`
	Description string            `yaml:"description"`
	Conventions map[string]string `yaml:"conventions"`
}

type ListItem struct {
	ID          string
	Name        string
	Type        string
	Status      string
	Path        string
	Description string
}
