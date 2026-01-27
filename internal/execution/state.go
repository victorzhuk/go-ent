package execution

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/victorzhuk/go-ent/internal/domain"
)

const (
	StateVersion = "1.0.0"

	ExecutionStatusPending     = "pending"
	ExecutionStatusRunning     = "running"
	ExecutionStatusInterrupted = "interrupted"
	ExecutionStatusCompleted   = "completed"
	ExecutionStatusFailed      = "failed"

	VersionPolicyBackwardCompatible = "backward-compatible"
	VersionPolicyIncompatible       = "incompatible"
	VersionPolicyMigrationRequired  = "migration-required"
)

type VersionCompatibility struct {
	Compatible         bool
	Policy             string
	Reason             string
	CanMigrate         bool
	MigrationAvailable bool
	CurrentVersion     string
	StateVersion       string
}

type ExecutionState struct {
	ID string `json:"id"`

	Task *Task `json:"task"`

	Context *TaskContext `json:"context"`

	Result *Result `json:"result,omitempty"`

	Status string `json:"status"`

	Agent    domain.AgentRole         `json:"agent"`
	Model    string                   `json:"model"`
	Runtime  domain.Runtime           `json:"runtime"`
	Strategy domain.ExecutionStrategy `json:"strategy"`

	Checksum string `json:"checksum"`

	Version string `json:"version"`

	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	StartedAt   time.Time `json:"started_at,omitempty"`
	CompletedAt time.Time `json:"completed_at,omitempty"`

	Metadata map[string]string `json:"metadata,omitempty"`
}

type ExecutionConfig struct {
	Agent    domain.AgentRole
	Model    string
	Runtime  domain.Runtime
	Strategy domain.ExecutionStrategy
	Budget   *BudgetLimit
	Skills   []string
}

func NewExecutionState(task *Task) *ExecutionState {
	id := uuid.New().String()
	now := time.Now()

	state := &ExecutionState{
		ID:        id,
		Task:      task,
		Context:   task.Context,
		Status:    ExecutionStatusPending,
		Version:   StateVersion,
		CreatedAt: now,
		UpdatedAt: now,
		Metadata:  make(map[string]string),
	}

	state.Checksum = state.computeChecksum()

	return state
}

func (s *ExecutionState) WithConfig(cfg ExecutionConfig) *ExecutionState {
	s.Agent = cfg.Agent
	s.Model = cfg.Model
	s.Runtime = cfg.Runtime
	s.Strategy = cfg.Strategy
	s.UpdatedAt = time.Now()
	s.Checksum = s.computeChecksum()
	return s
}

func (s *ExecutionState) Start() error {
	if s.Status != ExecutionStatusPending {
		return fmt.Errorf("cannot start execution with status %s", s.Status)
	}

	s.Status = ExecutionStatusRunning
	s.StartedAt = time.Now()
	s.UpdatedAt = time.Now()
	s.Checksum = s.computeChecksum()

	return nil
}

func (s *ExecutionState) Complete(result *Result) error {
	if s.Status != ExecutionStatusRunning {
		return fmt.Errorf("cannot complete execution with status %s", s.Status)
	}

	s.Result = result
	s.Status = ExecutionStatusCompleted
	s.CompletedAt = time.Now()
	s.UpdatedAt = time.Now()
	s.Checksum = s.computeChecksum()

	return nil
}

func (s *ExecutionState) Fail(err error) error {
	if s.Status != ExecutionStatusRunning {
		return fmt.Errorf("cannot fail execution with status %s", s.Status)
	}

	result := &Result{
		Success: false,
		Error:   err.Error(),
	}

	s.Result = result
	s.Status = ExecutionStatusFailed
	s.CompletedAt = time.Now()
	s.UpdatedAt = time.Now()
	s.Checksum = s.computeChecksum()

	return nil
}

func (s *ExecutionState) Interrupt() error {
	if s.Status != ExecutionStatusRunning {
		return fmt.Errorf("cannot interrupt execution with status %s", s.Status)
	}

	s.Status = ExecutionStatusInterrupted
	s.UpdatedAt = time.Now()
	s.Checksum = s.computeChecksum()

	return nil
}

func (s *ExecutionState) UpdateContext(ctx *TaskContext) error {
	s.Context = ctx
	s.UpdatedAt = time.Now()
	s.Checksum = s.computeChecksum()

	return nil
}

func (s *ExecutionState) SetMetadata(key, value string) {
	if s.Metadata == nil {
		s.Metadata = make(map[string]string)
	}
	s.Metadata[key] = value
	s.UpdatedAt = time.Now()
	s.Checksum = s.computeChecksum()
}

func (s *ExecutionState) GetMetadata(key string) (string, bool) {
	if s.Metadata == nil {
		return "", false
	}
	val, ok := s.Metadata[key]
	return val, ok
}

func (s *ExecutionState) IsRunning() bool {
	return s.Status == ExecutionStatusRunning
}

func (s *ExecutionState) IsCompleted() bool {
	return s.Status == ExecutionStatusCompleted
}

func (s *ExecutionState) IsFailed() bool {
	return s.Status == ExecutionStatusFailed
}

func (s *ExecutionState) IsInterrupted() bool {
	return s.Status == ExecutionStatusInterrupted
}

func (s *ExecutionState) Duration() time.Duration {
	if s.StartedAt.IsZero() {
		return 0
	}
	end := s.CompletedAt
	if end.IsZero() {
		end = time.Now()
	}
	return end.Sub(s.StartedAt)
}

func (s *ExecutionState) computeChecksum() string {
	checksumCopy := s.Checksum
	s.Checksum = ""
	defer func() { s.Checksum = checksumCopy }()

	data, err := json.Marshal(s)
	if err != nil {
		return ""
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func (s *ExecutionState) ValidateChecksum() bool {
	storedChecksum := s.Checksum
	if storedChecksum == "" {
		return false
	}

	clone := s.Clone()
	clone.Checksum = ""
	computedChecksum := clone.computeChecksum()

	return computedChecksum == storedChecksum
}

func (s *ExecutionState) ToJSON() ([]byte, error) {
	return json.MarshalIndent(s, "", "  ")
}

func (s *ExecutionState) FromJSON(data []byte) error {
	return json.Unmarshal(data, s)
}

func (s *ExecutionState) Clone() *ExecutionState {
	clone := *s

	if s.Task != nil {
		taskCopy := *s.Task
		clone.Task = &taskCopy
		if s.Task.Metadata != nil {
			clone.Task.Metadata = make(map[string]interface{})
			for k, v := range s.Task.Metadata {
				clone.Task.Metadata[k] = v
			}
		}
	}

	if s.Context != nil {
		ctxCopy := *s.Context
		clone.Context = &ctxCopy
		if s.Context.Files != nil {
			clone.Context.Files = make([]string, len(s.Context.Files))
			copy(clone.Context.Files, s.Context.Files)
		}
	}

	if s.Result != nil {
		resultCopy := *s.Result
		clone.Result = &resultCopy
		if s.Result.Metadata != nil {
			clone.Result.Metadata = make(map[string]interface{})
			for k, v := range s.Result.Metadata {
				clone.Result.Metadata[k] = v
			}
		}
		if s.Result.Adjustments != nil {
			clone.Result.Adjustments = make([]string, len(s.Result.Adjustments))
			copy(clone.Result.Adjustments, s.Result.Adjustments)
		}
	}

	if s.Metadata != nil {
		clone.Metadata = make(map[string]string)
		for k, v := range s.Metadata {
			clone.Metadata[k] = v
		}
	}

	return &clone
}

func (s *ExecutionState) String() string {
	return fmt.Sprintf("ExecutionState{id=%s, status=%s, task=%s}",
		s.ID,
		s.Status,
		truncate(s.Task.Description, 50),
	)
}

func (s *ExecutionState) Validate() error {
	if s.ID == "" {
		return fmt.Errorf("execution ID is required")
	}

	if s.Task == nil {
		return fmt.Errorf("task is required")
	}

	if s.Task.Description == "" {
		return fmt.Errorf("task description is required")
	}

	if s.Version == "" {
		return fmt.Errorf("state version is required")
	}

	if s.Status == "" {
		return fmt.Errorf("status is required")
	}

	validStatuses := map[string]bool{
		ExecutionStatusPending:     true,
		ExecutionStatusRunning:     true,
		ExecutionStatusInterrupted: true,
		ExecutionStatusCompleted:   true,
		ExecutionStatusFailed:      true,
	}

	if !validStatuses[s.Status] {
		return fmt.Errorf("invalid status: %s", s.Status)
	}

	if s.CreatedAt.IsZero() {
		return fmt.Errorf("created_at timestamp is required")
	}

	if s.UpdatedAt.IsZero() {
		return fmt.Errorf("updated_at timestamp is required")
	}

	if !s.ChecksumValid() {
		return fmt.Errorf("checksum validation failed")
	}

	if s.Status == ExecutionStatusRunning && s.StartedAt.IsZero() {
		return fmt.Errorf("running state must have started_at timestamp")
	}

	if (s.Status == ExecutionStatusCompleted || s.Status == ExecutionStatusFailed) && s.CompletedAt.IsZero() {
		return fmt.Errorf("completed or failed state must have completed_at timestamp")
	}

	return nil
}

func (s *ExecutionState) ChecksumValid() bool {
	return s.ValidateChecksum()
}

func (s *ExecutionState) VersionCompatible() bool {
	return s.Version == StateVersion
}

func (s *ExecutionState) CheckVersionCompatibility() VersionCompatibility {
	if s.Version == "" {
		return VersionCompatibility{
			Compatible:     false,
			Policy:         VersionPolicyIncompatible,
			Reason:         "state has no version information",
			CanMigrate:     false,
			CurrentVersion: StateVersion,
			StateVersion:   "unknown",
		}
	}

	if s.Version == StateVersion {
		return VersionCompatibility{
			Compatible:     true,
			Policy:         VersionPolicyBackwardCompatible,
			Reason:         "version matches current",
			CanMigrate:     false,
			CurrentVersion: StateVersion,
			StateVersion:   s.Version,
		}
	}

	major := StateVersion[:1]
	stateMajor := s.Version[:1]

	if major == stateMajor {
		return VersionCompatibility{
			Compatible:     true,
			Policy:         VersionPolicyBackwardCompatible,
			Reason:         "same major version - backward compatible",
			CanMigrate:     true,
			CurrentVersion: StateVersion,
			StateVersion:   s.Version,
		}
	}

	return VersionCompatibility{
		Compatible:     false,
		Policy:         VersionPolicyIncompatible,
		Reason:         "major version mismatch - cannot resume",
		CanMigrate:     false,
		CurrentVersion: StateVersion,
		StateVersion:   s.Version,
	}
}

func (s *ExecutionState) CanResume() bool {
	return s.IsInterrupted() || s.IsFailed() || s.Status == ExecutionStatusPending
}

func (s *ExecutionState) Resume() error {
	if s.Status != ExecutionStatusInterrupted && s.Status != ExecutionStatusFailed {
		return fmt.Errorf("cannot resume execution with status %s", s.Status)
	}

	s.Status = ExecutionStatusRunning
	s.UpdatedAt = time.Now()
	s.Checksum = s.computeChecksum()

	return nil
}

type ValidationError struct {
	Field   string
	Message string
	Level   ValidationLevel
}

func (s *ExecutionState) ValidateEnvironment() []ValidationError {
	var errors []ValidationError

	if s.Context == nil {
		errors = append(errors, ValidationError{
			Field:   "context",
			Message: "context is required for environment validation",
			Level:   ValidationLevelError,
		})
		return errors
	}

	if s.Context.ProjectPath == "" {
		errors = append(errors, ValidationError{
			Field:   "context.project_path",
			Message: "project path is required",
			Level:   ValidationLevelError,
		})
	}

	if len(s.Context.Files) > 0 {
		missingFiles := []string{}
		for _, file := range s.Context.Files {
			if file == "" {
				continue
			}
			if _, err := os.Stat(file); os.IsNotExist(err) {
				missingFiles = append(missingFiles, file)
			}
		}
		if len(missingFiles) > 0 {
			errors = append(errors, ValidationError{
				Field:   "context.files",
				Message: fmt.Sprintf("%d file(s) not found: %s", len(missingFiles), missingFiles[0]),
				Level:   ValidationLevelWarning,
			})
		}
	}

	return errors
}

func (s *ExecutionState) ValidateForResume() []ValidationError {
	var errors []ValidationError

	if s.Task == nil {
		errors = append(errors, ValidationError{
			Field:   "task",
			Message: "task is required for resume",
			Level:   ValidationLevelError,
		})
		return errors
	}

	if s.Task.Description == "" {
		errors = append(errors, ValidationError{
			Field:   "task.description",
			Message: "task description is required",
			Level:   ValidationLevelError,
		})
	}

	if !s.CanResume() {
		errors = append(errors, ValidationError{
			Field:   "status",
			Message: fmt.Sprintf("cannot resume from status %s", s.Status),
			Level:   ValidationLevelError,
		})
	}

	if !s.ChecksumValid() {
		errors = append(errors, ValidationError{
			Field:   "checksum",
			Message: "checksum validation failed - data may be corrupted",
			Level:   ValidationLevelError,
		})
	}

	compat := s.CheckVersionCompatibility()
	if !compat.Compatible {
		errors = append(errors, ValidationError{
			Field:   "version",
			Message: fmt.Sprintf("%s: %s (state: %s, current: %s)", compat.Policy, compat.Reason, compat.StateVersion, compat.CurrentVersion),
			Level:   ValidationLevelError,
		})
	} else if compat.Policy != VersionPolicyBackwardCompatible {
		errors = append(errors, ValidationError{
			Field:   "version",
			Message: fmt.Sprintf("%s: %s (state: %s, current: %s)", compat.Policy, compat.Reason, compat.StateVersion, compat.CurrentVersion),
			Level:   ValidationLevelWarning,
		})
	}

	if s.Runtime == "" {
		errors = append(errors, ValidationError{
			Field:   "runtime",
			Message: "runtime is required for resume",
			Level:   ValidationLevelError,
		})
	}

	if s.Agent == "" {
		errors = append(errors, ValidationError{
			Field:   "agent",
			Message: "agent role is required for resume",
			Level:   ValidationLevelError,
		})
	}

	return errors
}
