package execution

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/victorzhuk/go-ent/internal/domain"
)

// ClaudeCodeRunner executes tasks via Claude Code using MCP protocol.
type ClaudeCodeRunner struct {
	logger *slog.Logger
}

// NewClaudeCodeRunner creates a new Claude Code MCP runner.
func NewClaudeCodeRunner(logger *slog.Logger) *ClaudeCodeRunner {
	if logger == nil {
		logger = slog.Default()
	}
	return &ClaudeCodeRunner{logger: logger}
}

// Runtime returns the runtime this runner supports.
func (r *ClaudeCodeRunner) Runtime() domain.Runtime {
	return domain.RuntimeClaudeCode
}

// Available checks if Claude Code MCP execution is available.
func (r *ClaudeCodeRunner) Available(ctx context.Context) bool {
	// Claude Code is available when running within MCP context
	// For now, assume available (can add MCP context detection later)
	return true
}

// Execute runs a task via Claude Code MCP.
func (r *ClaudeCodeRunner) Execute(ctx context.Context, req *Request) (*Result, error) {
	start := time.Now()

	r.logger.Info("executing task via Claude Code MCP",
		"agent", req.Agent,
		"model", req.Model,
		"skills", req.Skills,
		"task", truncate(req.Task, 100),
	)

	// Build MCP prompt with agent role and skills
	prompt := r.buildMCPPrompt(req)

	// Return prompt as instruction for Claude Code to execute
	result := &Result{
		Success:  true,
		Output:   prompt,
		Duration: time.Since(start),
		Metadata: map[string]interface{}{
			"runtime": string(domain.RuntimeClaudeCode),
			"agent":   string(req.Agent),
			"model":   req.Model,
			"skills":  req.Skills,
		},
	}

	r.logger.Info("Claude Code MCP prompt generated",
		"duration", result.Duration,
		"prompt_length", len(prompt),
	)

	return result, nil
}

// Interrupt attempts to stop a running execution.
func (r *ClaudeCodeRunner) Interrupt(ctx context.Context) error {
	return fmt.Errorf("claude Code runner interruption not implemented")
}

// buildMCPPrompt constructs the MCP prompt for Claude Code execution.
func (r *ClaudeCodeRunner) buildMCPPrompt(req *Request) string {
	var prompt strings.Builder

	// Agent role context
	prompt.WriteString(fmt.Sprintf("# Agent Role: %s\n\n", req.Agent))
	prompt.WriteString(r.getAgentContext(req.Agent))
	prompt.WriteString("\n\n")

	// Model selection
	prompt.WriteString(fmt.Sprintf("# Model: %s\n\n", req.Model))

	// Skills activation
	if len(req.Skills) > 0 {
		prompt.WriteString("# Active Skills\n\n")
		for _, skill := range req.Skills {
			prompt.WriteString(fmt.Sprintf("- %s\n", skill))
		}
		prompt.WriteString("\n")
	}

	// Task context
	if req.Context != nil {
		if req.Context.ChangeID != "" {
			prompt.WriteString(fmt.Sprintf("# Change: %s\n", req.Context.ChangeID))
		}
		if req.Context.TaskID != "" {
			prompt.WriteString(fmt.Sprintf("# Task ID: %s\n", req.Context.TaskID))
		}
		if req.Context.HasFiles() {
			prompt.WriteString("\n## Context Files\n\n")
			for _, file := range req.Context.Files {
				prompt.WriteString(fmt.Sprintf("- `%s`\n", file))
			}
		}
		prompt.WriteString("\n")
	}

	// Main task
	prompt.WriteString("# Task\n\n")
	prompt.WriteString(req.Task)
	prompt.WriteString("\n")

	return prompt.String()
}

// getAgentContext returns contextual information for the agent role.
func (r *ClaudeCodeRunner) getAgentContext(agent domain.AgentRole) string {
	switch agent {
	case domain.AgentRoleArchitect:
		return "You are an architect agent responsible for system design, " +
			"architecture decisions, and technical planning."

	case domain.AgentRoleSenior:
		return "You are a senior developer agent responsible for complex implementation, " +
			"code review, and technical leadership."

	case domain.AgentRoleDeveloper:
		return "You are a developer agent responsible for implementing features, " +
			"writing tests, and following established patterns."

	case domain.AgentRoleReviewer:
		return "You are a reviewer agent responsible for code review, " +
			"quality assurance, and identifying issues."

	case domain.AgentRoleOps:
		return "You are an ops agent responsible for deployment, infrastructure, " +
			"and operational concerns."

	default:
		return fmt.Sprintf("You are a %s agent.", agent)
	}
}
