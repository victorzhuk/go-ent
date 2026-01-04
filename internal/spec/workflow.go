package spec

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type WorkflowStatus string

const (
	WorkflowStatusActive    WorkflowStatus = "active"
	WorkflowStatusWaiting   WorkflowStatus = "waiting"
	WorkflowStatusCompleted WorkflowStatus = "completed"
	WorkflowStatusCancelled WorkflowStatus = "cancelled"
)

type WorkflowState struct {
	ID        string                 `yaml:"id" json:"id"`
	ChangeID  string                 `yaml:"change_id" json:"change_id"`
	Phase     string                 `yaml:"phase" json:"phase"`
	WaitPoint string                 `yaml:"wait_point,omitempty" json:"wait_point,omitempty"`
	Status    WorkflowStatus         `yaml:"status" json:"status"`
	Context   map[string]interface{} `yaml:"context,omitempty" json:"context,omitempty"`
	CreatedAt time.Time              `yaml:"created_at" json:"created_at"`
	UpdatedAt time.Time              `yaml:"updated_at" json:"updated_at"`
}

func NewWorkflowState(changeID, phase string) *WorkflowState {
	now := time.Now()
	return &WorkflowState{
		ID:        uuid.New().String(),
		ChangeID:  changeID,
		Phase:     phase,
		Status:    WorkflowStatusActive,
		Context:   make(map[string]interface{}),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (w *WorkflowState) SetWaitPoint(waitPoint string) {
	w.WaitPoint = waitPoint
	w.Status = WorkflowStatusWaiting
	w.UpdatedAt = time.Now()
}

func (w *WorkflowState) Approve() {
	w.WaitPoint = ""
	w.Status = WorkflowStatusActive
	w.UpdatedAt = time.Now()
}

func (w *WorkflowState) Complete() {
	w.Status = WorkflowStatusCompleted
	w.UpdatedAt = time.Now()
}

func (w *WorkflowState) Cancel() {
	w.Status = WorkflowStatusCancelled
	w.UpdatedAt = time.Now()
}

func (s *Store) WorkflowPath() string {
	return fmt.Sprintf("%s/.workflow.yaml", s.SpecPath())
}

func (s *Store) LoadWorkflow() (*WorkflowState, error) {
	return loadYAML[WorkflowState](s.WorkflowPath())
}

func (s *Store) SaveWorkflow(state *WorkflowState) error {
	return saveYAML(s.WorkflowPath(), state)
}

func (s *Store) WorkflowExists() bool {
	return fileExists(s.WorkflowPath())
}
