package background

import (
	"fmt"
	"sync"
	"time"
)

// Status represents the current state of a background agent.
type Status string

const (
	// StatusPending indicates the agent is waiting to start.
	StatusPending Status = "pending"

	// StatusRunning indicates the agent is actively executing.
	StatusRunning Status = "running"

	// StatusCompleted indicates the agent finished successfully.
	StatusCompleted Status = "completed"

	// StatusFailed indicates the agent encountered an error.
	StatusFailed Status = "failed"

	// StatusKilled indicates the agent was terminated externally.
	StatusKilled Status = "killed"
)

// Valid returns true if the status is valid.
func (s Status) Valid() bool {
	switch s {
	case StatusPending, StatusRunning, StatusCompleted, StatusFailed, StatusKilled:
		return true
	default:
		return false
	}
}

// Agent represents a background agent execution.
type Agent struct {
	mu sync.RWMutex

	// ID uniquely identifies this agent instance.
	ID string

	// Role is the agent role assigned to this instance.
	Role string

	// Model is the model used for this agent.
	Model string

	// Task is the task description given to the agent.
	Task string

	// Status is the current execution status.
	Status Status

	// CreatedAt is when the agent was created.
	CreatedAt time.Time

	// StartedAt is when the agent started execution.
	StartedAt time.Time

	// CompletedAt is when the agent finished execution.
	CompletedAt time.Time

	// Output contains the agent's execution output.
	Output string

	// Error contains any error that occurred during execution.
	Error error
}

// NewAgent creates a new background agent instance.
func NewAgent(id, role, model, task string) (*Agent, error) {
	if id == "" {
		return nil, fmt.Errorf("agent id required")
	}
	if role == "" {
		return nil, fmt.Errorf("agent role required")
	}
	if model == "" {
		return nil, fmt.Errorf("model required")
	}
	if task == "" {
		return nil, fmt.Errorf("task required")
	}

	return &Agent{
		ID:        id,
		Role:      role,
		Model:     model,
		Task:      task,
		Status:    StatusPending,
		CreatedAt: time.Now(),
	}, nil
}

// Start marks the agent as started.
func (a *Agent) Start() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Status = StatusRunning
	a.StartedAt = time.Now()
}

// Complete marks the agent as completed with the given output.
func (a *Agent) Complete(output string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Status = StatusCompleted
	a.CompletedAt = time.Now()
	a.Output = output
}

// Fail marks the agent as failed with the given error.
func (a *Agent) Fail(err error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Status = StatusFailed
	a.CompletedAt = time.Now()
	a.Error = err
}

// Kill marks the agent as killed.
func (a *Agent) Kill() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Status = StatusKilled
	a.CompletedAt = time.Now()
}

// Duration returns the execution duration.
// Returns 0 if the agent hasn't started.
func (a *Agent) Duration() time.Duration {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if a.StartedAt.IsZero() {
		return 0
	}
	if !a.CompletedAt.IsZero() {
		return a.CompletedAt.Sub(a.StartedAt)
	}
	return time.Since(a.StartedAt)
}

// Snapshot represents a thread-safe snapshot of an agent's state.
type Snapshot struct {
	ID          string
	Role        string
	Model       string
	Task        string
	Status      Status
	CreatedAt   time.Time
	StartedAt   time.Time
	CompletedAt time.Time
	Output      string
	Error       error
}

// GetSnapshot returns a thread-safe snapshot of the agent's current state.
func (a *Agent) GetSnapshot() Snapshot {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return Snapshot{
		ID:          a.ID,
		Role:        a.Role,
		Model:       a.Model,
		Task:        a.Task,
		Status:      a.Status,
		CreatedAt:   a.CreatedAt,
		StartedAt:   a.StartedAt,
		CompletedAt: a.CompletedAt,
		Output:      a.Output,
		Error:       a.Error,
	}
}

// ToAgent converts a snapshot back to an Agent instance.
func (s Snapshot) ToAgent() *Agent {
	return &Agent{
		ID:          s.ID,
		Role:        s.Role,
		Model:       s.Model,
		Task:        s.Task,
		Status:      s.Status,
		CreatedAt:   s.CreatedAt,
		StartedAt:   s.StartedAt,
		CompletedAt: s.CompletedAt,
		Output:      s.Output,
		Error:       s.Error,
	}
}
