package execution

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"time"

	"github.com/victorzhuk/go-ent/internal/domain"
)

// OpenCodeRunner executes tasks via OpenCode CLI subprocess.
type OpenCodeRunner struct {
	binaryPath string
	logger     *slog.Logger
}

// NewOpenCodeRunner creates a new OpenCode subprocess runner.
func NewOpenCodeRunner(logger *slog.Logger) *OpenCodeRunner {
	if logger == nil {
		logger = slog.Default()
	}
	return &OpenCodeRunner{
		binaryPath: "opencode",
		logger:     logger,
	}
}

// Runtime returns the runtime this runner supports.
func (r *OpenCodeRunner) Runtime() domain.Runtime {
	return domain.RuntimeOpenCode
}

// Available checks if OpenCode CLI is available.
func (r *OpenCodeRunner) Available(ctx context.Context) bool {
	// Check if opencode binary is in PATH
	_, err := exec.LookPath(r.binaryPath)
	return err == nil
}

// Execute runs a task via OpenCode subprocess.
func (r *OpenCodeRunner) Execute(ctx context.Context, req *Request) (*Result, error) {
	start := time.Now()

	// Verify binary is available
	if !r.Available(ctx) {
		return nil, fmt.Errorf("opencode binary not found in PATH")
	}

	r.logger.Info("executing task via OpenCode subprocess",
		"agent", req.Agent,
		"model", req.Model,
		"task", truncate(req.Task, 100),
	)

	// Build prompt for OpenCode
	prompt := r.buildPrompt(req)

	// Execute OpenCode subprocess
	output, err := r.executeSubprocess(ctx, prompt, req)
	if err != nil {
		return &Result{
			Success:  false,
			Error:    err.Error(),
			Duration: time.Since(start),
			Metadata: map[string]interface{}{
				"runtime": string(domain.RuntimeOpenCode),
			},
		}, err
	}

	result := &Result{
		Success:  true,
		Output:   output,
		Duration: time.Since(start),
		Metadata: map[string]interface{}{
			"runtime": string(domain.RuntimeOpenCode),
			"agent":   string(req.Agent),
			"model":   req.Model,
		},
	}

	r.logger.Info("OpenCode execution completed",
		"duration", result.Duration,
		"success", result.Success,
	)

	return result, nil
}

// Interrupt attempts to stop a running execution.
func (r *OpenCodeRunner) Interrupt(ctx context.Context) error {
	// Would need to track running subprocess and send interrupt signal
	return fmt.Errorf("OpenCode runner interruption not implemented")
}

// buildPrompt constructs the prompt for OpenCode execution.
func (r *OpenCodeRunner) buildPrompt(req *Request) string {
	var prompt strings.Builder

	// Agent context
	prompt.WriteString(fmt.Sprintf("# Agent: %s\n\n", req.Agent))

	// Skills
	if len(req.Skills) > 0 {
		prompt.WriteString("## Skills\n\n")
		for _, skill := range req.Skills {
			prompt.WriteString(fmt.Sprintf("- %s\n", skill))
		}
		prompt.WriteString("\n")
	}

	// Context
	if req.Context != nil && req.Context.HasFiles() {
		prompt.WriteString("## Context Files\n\n")
		for _, file := range req.Context.Files {
			prompt.WriteString(fmt.Sprintf("- %s\n", file))
		}
		prompt.WriteString("\n")
	}

	// Task
	prompt.WriteString("## Task\n\n")
	prompt.WriteString(req.Task)

	return prompt.String()
}

// executeSubprocess spawns OpenCode process and captures output.
func (r *OpenCodeRunner) executeSubprocess(ctx context.Context, prompt string, req *Request) (string, error) {
	// Build OpenCode command
	// Example: opencode --model sonnet --prompt "task description"
	args := []string{}

	// Add model selection if specified
	if req.Model != "" {
		args = append(args, "--model", req.Model)
	}

	// Add prompt
	args = append(args, "--prompt", prompt)

	// If context has project path, use it as working directory
	cmd := exec.CommandContext(ctx, r.binaryPath, args...)
	if req.Context != nil && req.Context.ProjectPath != "" {
		cmd.Dir = req.Context.ProjectPath
	}

	// Execute and capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("opencode subprocess: %w", err)
	}

	return string(output), nil
}
