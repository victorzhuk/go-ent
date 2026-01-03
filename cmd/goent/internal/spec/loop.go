package spec

import (
	"fmt"
	"time"
)

type LoopStatus string

const (
	LoopStatusRunning   LoopStatus = "running"
	LoopStatusPaused    LoopStatus = "paused"
	LoopStatusCompleted LoopStatus = "completed"
	LoopStatusFailed    LoopStatus = "failed"
	LoopStatusCancelled LoopStatus = "cancelled"
)

type LoopState struct {
	Task        string     `yaml:"task" json:"task"`
	Iteration   int        `yaml:"iteration" json:"iteration"`
	MaxIter     int        `yaml:"max_iterations" json:"max_iterations"`
	LastError   string     `yaml:"last_error,omitempty" json:"last_error,omitempty"`
	Adjustments []string   `yaml:"adjustments,omitempty" json:"adjustments,omitempty"`
	Status      LoopStatus `yaml:"status" json:"status"`
	StartedAt   time.Time  `yaml:"started_at" json:"started_at"`
	UpdatedAt   time.Time  `yaml:"updated_at" json:"updated_at"`
}

func NewLoopState(task string, maxIter int) *LoopState {
	now := time.Now()
	return &LoopState{
		Task:        task,
		Iteration:   0,
		MaxIter:     maxIter,
		Status:      LoopStatusRunning,
		Adjustments: []string{},
		StartedAt:   now,
		UpdatedAt:   now,
	}
}

func (l *LoopState) NextIteration() {
	l.Iteration++
	l.UpdatedAt = time.Now()
}

func (l *LoopState) RecordError(err string) {
	l.LastError = err
	l.UpdatedAt = time.Now()
}

func (l *LoopState) RecordAdjustment(adjustment string) {
	l.Adjustments = append(l.Adjustments, adjustment)
	l.UpdatedAt = time.Now()
}

func (l *LoopState) MarkCompleted() {
	l.Status = LoopStatusCompleted
	l.UpdatedAt = time.Now()
}

func (l *LoopState) MarkFailed() {
	l.Status = LoopStatusFailed
	l.UpdatedAt = time.Now()
}

func (l *LoopState) Cancel() {
	l.Status = LoopStatusCancelled
	l.UpdatedAt = time.Now()
}

func (l *LoopState) ShouldContinue() bool {
	return l.Status == LoopStatusRunning && l.Iteration < l.MaxIter
}

func (s *Store) LoopPath() string {
	return fmt.Sprintf("%s/.loop-state.yaml", s.SpecPath())
}

func (s *Store) LoadLoop() (*LoopState, error) {
	return loadYAML[LoopState](s.LoopPath())
}

func (s *Store) SaveLoop(state *LoopState) error {
	return saveYAML(s.LoopPath(), state)
}

func (s *Store) LoopExists() bool {
	return fileExists(s.LoopPath())
}
