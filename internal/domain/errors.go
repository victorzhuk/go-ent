package domain

import (
	"errors"
	"fmt"
)

// Sentinel errors for common domain error conditions.
var (
	// ErrAgentNotFound indicates the requested agent role does not exist.
	ErrAgentNotFound = errors.New("agent not found")

	// ErrInvalidAgentConfig indicates the agent configuration is invalid.
	ErrInvalidAgentConfig = errors.New("invalid agent config")

	// ErrInvalidAction indicates an invalid or unsupported action was requested.
	ErrInvalidAction = errors.New("invalid action")

	// ErrInvalidStrategy indicates an invalid execution strategy was specified.
	ErrInvalidStrategy = errors.New("invalid strategy")

	// ErrSkillNotFound indicates the requested skill does not exist.
	ErrSkillNotFound = errors.New("skill not found")
)

// AgentError wraps agent-related errors with additional context.
type AgentError struct {
	Role AgentRole
	Err  error
}

func (e *AgentError) Error() string {
	return fmt.Sprintf("agent error [%s]: %v", e.Role, e.Err)
}

func (e *AgentError) Unwrap() error {
	return e.Err
}

// ActionError wraps action-related errors with additional context.
type ActionError struct {
	Action SpecAction
	Err    error
}

func (e *ActionError) Error() string {
	return fmt.Sprintf("action error [%s]: %v", e.Action, e.Err)
}

func (e *ActionError) Unwrap() error {
	return e.Err
}

// SkillError wraps skill-related errors with additional context.
type SkillError struct {
	Skill string
	Err   error
}

func (e *SkillError) Error() string {
	return fmt.Sprintf("skill error [%s]: %v", e.Skill, e.Err)
}

func (e *SkillError) Unwrap() error {
	return e.Err
}

// IsAgentError checks if an error is an agent-related error.
func IsAgentError(err error) bool {
	var agentErr *AgentError
	return errors.As(err, &agentErr)
}

// IsActionError checks if an error is an action-related error.
func IsActionError(err error) bool {
	var actionErr *ActionError
	return errors.As(err, &actionErr)
}

// IsSkillError checks if an error is a skill-related error.
func IsSkillError(err error) bool {
	var skillErr *SkillError
	return errors.As(err, &skillErr)
}
