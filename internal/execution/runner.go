package execution

import (
	"context"
	"time"

	"github.com/victorzhuk/go-ent/internal/domain"
)

// Runner executes tasks in a specific runtime environment.
type Runner interface {
	// Runtime returns the runtime this runner supports.
	Runtime() domain.Runtime

	// Available checks if this runner can execute in current environment.
	Available(ctx context.Context) bool

	// Execute runs a task and returns the result.
	Execute(ctx context.Context, req *Request) (*Result, error)

	// Interrupt attempts to stop a running execution.
	Interrupt(ctx context.Context) error
}

// Request holds everything needed to execute a task.
type Request struct {
	// Task is the task description.
	Task string

	// Agent is the selected agent role.
	Agent domain.AgentRole

	// Model is the model ID (opus, sonnet, haiku).
	Model string

	// Skills to activate during execution.
	Skills []string

	// Strategy defines the execution approach.
	Strategy domain.ExecutionStrategy

	// Budget specifies token/cost limits.
	Budget *BudgetLimit

	// Context provides project and file context.
	Context *TaskContext

	// Metadata holds additional request data.
	Metadata map[string]interface{}
}

// Result captures execution outcome.
type Result struct {
	// Success indicates if execution completed successfully.
	Success bool

	// Output contains the execution output.
	Output string

	// Error contains the error message if failed.
	Error string

	// TokensIn is the number of input tokens.
	TokensIn int

	// TokensOut is the number of output tokens.
	TokensOut int

	// Cost is the estimated execution cost.
	Cost float64

	// Duration is how long execution took.
	Duration time.Duration

	// Adjustments contains any self-corrections made during execution.
	Adjustments []string

	// Metadata holds additional result data.
	Metadata map[string]interface{}
}

// TaskContext provides project and file context for execution.
type TaskContext struct {
	// ProjectPath is the root path of the project.
	ProjectPath string

	// ChangeID identifies the change being worked on.
	ChangeID string

	// TaskID identifies the specific task.
	TaskID string

	// Files are the relevant file paths for this task.
	Files []string

	// WorkflowID identifies the workflow if part of one.
	WorkflowID string

	// IsSummarized indicates if context has been summarized.
	IsSummarized bool

	// SummarizationCount tracks how many times context was summarized.
	SummarizationCount int

	// OriginalContext holds the pre-summarization context if summarized.
	OriginalContext *TaskContext
}

// BudgetLimit defines spending constraints.
type BudgetLimit struct {
	// MaxTokens is the maximum tokens allowed (0 = unlimited).
	MaxTokens int

	// MaxCost is the maximum cost allowed (0 = unlimited).
	MaxCost float64

	// AutoProceed determines behavior when limit exceeded.
	// In MCP mode: warn and proceed.
	// In CLI mode: prompt user.
	AutoProceed bool
}

// TotalTokens returns the sum of input and output tokens.
func (r *Result) TotalTokens() int {
	return r.TokensIn + r.TokensOut
}

// Failed returns true if execution failed.
func (r *Result) Failed() bool {
	return !r.Success
}

// LLM model constants for summarization.
const (
	LLMModelClaude35   = "claude-3.5-sonnet"
	LLMModelClaude4    = "claude-4-5-20251101"
	LLMModelGPT4       = "gpt-4"
	LLMModelGPT4Turbo  = "gpt-4-turbo"
	LLMModelGPT35Turbo = "gpt-3.5-turbo"
)

// SummarizationThreshold defines when to trigger automatic context summarization.
type SummarizationThreshold struct {
	// FileCount triggers summarization when context exceeds this many files.
	FileCount int

	// ContextLength triggers summarization when total character count exceeds this.
	ContextLength int

	// TokenCount triggers summarization when estimated tokens exceed this.
	TokenCount int
}

// DefaultSummarizationThreshold returns default thresholds.
func DefaultSummarizationThreshold() SummarizationThreshold {
	return SummarizationThreshold{
		FileCount:     50,
		ContextLength: 50000,
		TokenCount:    10000,
	}
}
