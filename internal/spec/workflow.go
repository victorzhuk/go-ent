package spec

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/victorzhuk/go-ent/internal/domain"
)

type WorkflowStatus string

const (
	WorkflowStatusActive    WorkflowStatus = "active"
	WorkflowStatusWaiting   WorkflowStatus = "waiting"
	WorkflowStatusCompleted WorkflowStatus = "completed"
	WorkflowStatusCancelled WorkflowStatus = "cancelled"
)

// ExecutionRecord tracks a single execution within the workflow.
type ExecutionRecord struct {
	TaskID     string           `yaml:"task_id" json:"task_id"`
	AgentRole  domain.AgentRole `yaml:"agent_role" json:"agent_role"`
	Model      string           `yaml:"model" json:"model"`
	Runtime    string           `yaml:"runtime" json:"runtime"`
	Strategy   string           `yaml:"strategy" json:"strategy"`
	Success    bool             `yaml:"success" json:"success"`
	TokensIn   int              `yaml:"tokens_in" json:"tokens_in"`
	TokensOut  int              `yaml:"tokens_out" json:"tokens_out"`
	Cost       float64          `yaml:"cost" json:"cost"`
	Duration   time.Duration    `yaml:"duration" json:"duration"`
	ExecutedAt time.Time        `yaml:"executed_at" json:"executed_at"`
}

// WorkflowState tracks the execution state of a change proposal workflow.
// It bridges the spec domain (what needs to be done) with the agent domain
// (who is doing it) by tracking the current agent role executing the workflow.
type WorkflowState struct {
	ID               string                 `yaml:"id" json:"id"`
	ChangeID         string                 `yaml:"change_id" json:"change_id"`
	Phase            string                 `yaml:"phase" json:"phase"`
	AgentRole        domain.AgentRole       `yaml:"agent_role,omitempty" json:"agent_role,omitempty"` // Current agent executing this workflow
	WaitPoint        string                 `yaml:"wait_point,omitempty" json:"wait_point,omitempty"`
	Status           WorkflowStatus         `yaml:"status" json:"status"`
	Context          map[string]interface{} `yaml:"context,omitempty" json:"context,omitempty"`
	ExecutionHistory []ExecutionRecord      `yaml:"execution_history,omitempty" json:"execution_history,omitempty"`
	CreatedAt        time.Time              `yaml:"created_at" json:"created_at"`
	UpdatedAt        time.Time              `yaml:"updated_at" json:"updated_at"`
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

func (w *WorkflowState) SetAgent(role domain.AgentRole) {
	w.AgentRole = role
	w.UpdatedAt = time.Now()
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

// RecordExecution adds an execution record to the history.
func (w *WorkflowState) RecordExecution(record ExecutionRecord) {
	if w.ExecutionHistory == nil {
		w.ExecutionHistory = []ExecutionRecord{}
	}
	w.ExecutionHistory = append(w.ExecutionHistory, record)
	w.UpdatedAt = time.Now()
}

// TotalCost returns the total cost of all executions.
func (w *WorkflowState) TotalCost() float64 {
	total := 0.0
	for _, record := range w.ExecutionHistory {
		total += record.Cost
	}
	return total
}

// TotalTokens returns the total tokens used across all executions.
func (w *WorkflowState) TotalTokens() int {
	total := 0
	for _, record := range w.ExecutionHistory {
		total += record.TokensIn + record.TokensOut
	}
	return total
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
