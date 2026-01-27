package execution

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

const (
	executionsDirName = "executions"
)

var (
	execDirPath string
)

type CorruptedStateError struct {
	ExecutionID string
	Filename    string
	Reason      string
	CanRecover  bool
}

func (e *CorruptedStateError) Error() string {
	msg := fmt.Sprintf("corrupted state for execution %s (%s): %s", e.ExecutionID, e.Filename, e.Reason)
	if e.CanRecover {
		msg += " (recovery possible)"
	} else {
		msg += " (cannot be recovered)"
	}
	return msg
}

func (e *CorruptedStateError) Unwrap() error {
	return errors.New(e.Reason)
}

type ValidationLevel string

const (
	ValidationLevelInfo    ValidationLevel = "info"
	ValidationLevelWarning ValidationLevel = "warning"
	ValidationLevelError   ValidationLevel = "error"
)

type ValidationIssue struct {
	Level   ValidationLevel
	Field   string
	Message string
}

type StateValidationResult struct {
	Valid             bool
	CanResume         bool
	Message           string
	Filename          string
	Size              int64
	Modified          time.Time
	ChecksumValid     bool
	VersionCompatible bool
	Issues            []ValidationIssue
}

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		slog.Warn("failed to get home directory, using current directory", "error", err)
		homeDir = "."
	}

	execDirPath = filepath.Join(homeDir, ".go-ent", executionsDirName)
}

func EnsureExecutionsDir() error {
	if execDirPath == "" {
		return fmt.Errorf("executions directory path not initialized")
	}

	err := os.MkdirAll(execDirPath, 0750)
	if err != nil {
		return fmt.Errorf("create executions directory: %w", err)
	}

	slog.Debug("ensured executions directory exists", "path", execDirPath)

	return nil
}

func SaveState(state *ExecutionState) error {
	if err := EnsureExecutionsDir(); err != nil {
		return err
	}

	if state == nil {
		return fmt.Errorf("cannot save nil state")
	}

	if state.ID == "" {
		return fmt.Errorf("state has no ID")
	}

	data, err := state.ToJSON()
	if err != nil {
		return fmt.Errorf("serialize state: %w", err)
	}

	filename := filepath.Join(execDirPath, state.ID+".json")
	tempFile := filename + ".tmp"

	if err := os.WriteFile(tempFile, data, 0600); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}

	if err := os.Rename(tempFile, filename); err != nil {
		_ = os.Remove(tempFile)
		return fmt.Errorf("rename temp file: %w", err)
	}

	slog.Debug("saved execution state", "id", state.ID, "path", filename)

	return nil
}

func LoadState(executionID string) (*ExecutionState, error) {
	if executionID == "" {
		return nil, fmt.Errorf("execution ID cannot be empty")
	}

	filename := filepath.Join(execDirPath, executionID+".json")

	data, err := os.ReadFile(filename) //nolint:gosec
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("execution state not found: %s", executionID)
		}
		return nil, fmt.Errorf("read state file: %w", err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("state file is empty: %s", filename)
	}

	state := &ExecutionState{}
	if err := state.FromJSON(data); err != nil {
		var syntaxErr *json.SyntaxError
		if errors.As(err, &syntaxErr) {
			return nil, &CorruptedStateError{
				ExecutionID: executionID,
				Filename:    filename,
				Reason:      fmt.Sprintf("invalid JSON at offset %d", syntaxErr.Offset),
				CanRecover:  false,
			}
		}
		return nil, &CorruptedStateError{
			ExecutionID: executionID,
			Filename:    filename,
			Reason:      fmt.Sprintf("deserialization failed: %v", err),
			CanRecover:  false,
		}
	}

	if !state.ValidateChecksum() {
		return nil, &CorruptedStateError{
			ExecutionID: executionID,
			Filename:    filename,
			Reason:      "checksum validation failed - data may be corrupted or tampered with",
			CanRecover:  false,
		}
	}

	if err := state.Validate(); err != nil {
		return nil, &CorruptedStateError{
			ExecutionID: executionID,
			Filename:    filename,
			Reason:      fmt.Sprintf("state validation failed: %v", err),
			CanRecover:  false,
		}
	}

	slog.Debug("loaded execution state", "id", executionID, "path", filename)

	return state, nil
}

func LoadStateWithValidation(executionID string) (*ExecutionState, []ValidationIssue, error) {
	state, err := LoadState(executionID)
	if err != nil {
		return nil, nil, err
	}

	var issues []ValidationIssue

	if !state.VersionCompatible() {
		issues = append(issues, ValidationIssue{
			Level:   ValidationLevelWarning,
			Field:   "version",
			Message: fmt.Sprintf("state version %s may not be compatible with current version %s", state.Version, StateVersion),
		})
	}

	if state.Status == ExecutionStatusRunning && !state.StartedAt.IsZero() {
		runningDuration := state.Duration()
		if runningDuration > 24*time.Hour {
			issues = append(issues, ValidationIssue{
				Level:   ValidationLevelWarning,
				Field:   "status",
				Message: fmt.Sprintf("execution has been running for %v, may be stale", runningDuration),
			})
		}
	}

	return state, issues, nil
}

func ValidateStateFile(executionID string) (*StateValidationResult, error) {
	filename := filepath.Join(execDirPath, executionID+".json")

	info, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return &StateValidationResult{
				Valid:   false,
				Message: "state file not found",
			}, nil
		}
		return nil, fmt.Errorf("stat state file: %w", err)
	}

	result := &StateValidationResult{
		Filename: filename,
		Size:     info.Size(),
		Modified: info.ModTime(),
	}

	data, err := os.ReadFile(filename) //nolint:gosec
	if err != nil {
		result.Valid = false
		result.Message = fmt.Sprintf("failed to read file: %v", err)
		return result, nil
	}

	result.Size = int64(len(data))

	if len(data) == 0 {
		result.Valid = false
		result.Message = "file is empty"
		return result, nil
	}

	state := &ExecutionState{}
	if err := state.FromJSON(data); err != nil {
		result.Valid = false
		result.Message = fmt.Sprintf("invalid JSON: %v", err)
		return result, nil
	}

	result.ChecksumValid = state.ChecksumValid()
	result.VersionCompatible = state.VersionCompatible()

	if !result.ChecksumValid {
		result.Valid = false
		result.Message = "checksum validation failed"
		return result, nil
	}

	if err := state.Validate(); err != nil {
		result.Valid = false
		result.Message = fmt.Sprintf("validation failed: %v", err)
		return result, nil
	}

	result.Valid = true
	result.Message = "state is valid and can be resumed"
	result.CanResume = state.CanResume()

	return result, nil
}

func DeleteState(executionID string) error {
	if executionID == "" {
		return fmt.Errorf("execution ID cannot be empty")
	}

	filename := filepath.Join(execDirPath, executionID+".json")

	if err := os.Remove(filename); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("execution state not found: %s", executionID)
		}
		return fmt.Errorf("delete state file: %w", err)
	}

	slog.Debug("deleted execution state", "id", executionID, "path", filename)

	return nil
}

func ListExecutions() ([]string, error) {
	if err := EnsureExecutionsDir(); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(execDirPath)
	if err != nil {
		return nil, fmt.Errorf("read executions directory: %w", err)
	}

	var executionIDs []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if filepath.Ext(name) == ".json" {
			executionID := name[:len(name)-5]
			executionIDs = append(executionIDs, executionID)
		}
	}

	slog.Debug("listed executions", "count", len(executionIDs))

	return executionIDs, nil
}
