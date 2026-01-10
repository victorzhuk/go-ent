package execution

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"time"

	"github.com/victorzhuk/go-ent/internal/domain"
)

// CLIRunner executes tasks in standalone CLI mode.
type CLIRunner struct {
	logger *slog.Logger
}

// NewCLIRunner creates a new CLI runner.
func NewCLIRunner(logger *slog.Logger) *CLIRunner {
	if logger == nil {
		logger = slog.Default()
	}
	return &CLIRunner{logger: logger}
}

// Runtime returns the runtime this runner supports.
func (r *CLIRunner) Runtime() domain.Runtime {
	return domain.RuntimeCLI
}

// Available checks if CLI execution is available.
func (r *CLIRunner) Available(ctx context.Context) bool {
	// CLI runner is always available
	return true
}

// Execute runs a task in CLI mode.
func (r *CLIRunner) Execute(ctx context.Context, req *Request) (*Result, error) {
	start := time.Now()

	r.logger.Info("executing task in CLI mode",
		"agent", req.Agent,
		"model", req.Model,
		"task", truncate(req.Task, 100),
	)

	// Build prompt for CLI execution
	prompt := r.buildPrompt(req)

	// For now, CLI runner returns the prompt as output
	// In future, this would spawn an actual CLI process
	result := &Result{
		Success:  true,
		Output:   prompt,
		Duration: time.Since(start),
		Metadata: map[string]interface{}{
			"runtime": string(domain.RuntimeCLI),
			"agent":   string(req.Agent),
			"model":   req.Model,
		},
	}

	r.logger.Info("CLI execution completed",
		"duration", result.Duration,
		"success", result.Success,
	)

	return result, nil
}

// Interrupt attempts to stop a running execution.
func (r *CLIRunner) Interrupt(ctx context.Context) error {
	// CLI runner doesn't support interruption yet
	return fmt.Errorf("CLI runner does not support interruption")
}

// buildPrompt constructs the prompt for CLI execution.
func (r *CLIRunner) buildPrompt(req *Request) string {
	prompt := fmt.Sprintf("# Task\n\n%s\n\n", req.Task)

	if len(req.Skills) > 0 {
		prompt += "# Skills\n\n"
		for _, skill := range req.Skills {
			prompt += fmt.Sprintf("- %s\n", skill)
		}
		prompt += "\n"
	}

	if req.Context != nil && req.Context.HasFiles() {
		prompt += "# Context Files\n\n"
		for _, file := range req.Context.Files {
			prompt += fmt.Sprintf("- %s\n", file)
		}
		prompt += "\n"
	}

	prompt += fmt.Sprintf("# Agent: %s\n", req.Agent)
	prompt += fmt.Sprintf("# Model: %s\n", req.Model)

	return prompt
}

// ExecuteCLICommand executes a CLI command and returns the output.
func ExecuteCLICommand(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("execute %s: %w", name, err)
	}
	return string(output), nil
}

// truncate truncates a string to the given length.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
