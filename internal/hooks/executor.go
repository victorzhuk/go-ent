package hooks

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// Executor runs hook commands and manages hook execution.
type Executor struct {
	logger *slog.Logger
}

// NewExecutor creates a new hook executor.
func NewExecutor(logger *slog.Logger) *Executor {
	if logger == nil {
		logger = slog.Default()
	}
	return &Executor{logger: logger}
}

// ExecuteCommand runs a shell command with environment variables and timeout.
func (e *Executor) ExecuteCommand(ctx context.Context, command string, env map[string]string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", "-c", command)
	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		e.logger.Warn("hook command failed",
			"command", command,
			"error", err,
			"output", string(output))
		return fmt.Errorf("hook command: %w", err)
	}

	if len(output) > 0 {
		e.logger.Debug("hook command output",
			"command", command,
			"output", string(output))
	}

	return nil
}

// MatchTool checks if a tool name matches the given pattern (regex).
func (e *Executor) MatchTool(pattern, toolName string) bool {
	if pattern == "" {
		return true // Empty pattern matches all
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		e.logger.Warn("invalid hook pattern", "pattern", pattern, "error", err)
		return false
	}

	return re.MatchString(toolName)
}

// RunPreToolHooks executes pre-tool hooks for a matched tool.
func (e *Executor) RunPreToolHooks(ctx context.Context, hooks []HookMatcher, toolName string, input any) error {
	for _, matcher := range hooks {
		if !e.MatchTool(matcher.Matcher, toolName) {
			continue
		}

		for _, hook := range matcher.Hooks {
			if err := e.executeHook(ctx, hook, toolName, input, nil, nil); err != nil {
				// Pre-hooks can block execution
				return err
			}
		}
	}
	return nil
}

// RunPostToolHooks executes post-tool hooks for a matched tool.
func (e *Executor) RunPostToolHooks(ctx context.Context, hooks []HookMatcher, toolName string, result any, err error) {
	for _, matcher := range hooks {
		if !e.MatchTool(matcher.Matcher, toolName) {
			continue
		}

		for _, hook := range matcher.Hooks {
			// Post-hooks don't block (errors are logged but not returned)
			if hookErr := e.executeHook(ctx, hook, toolName, nil, result, err); hookErr != nil {
				e.logger.Warn("post-tool hook failed",
					"tool", toolName,
					"error", hookErr)
			}
		}
	}
}

// RunOpenSpecHook executes an OpenSpec lifecycle hook.
func (e *Executor) RunOpenSpecHook(ctx context.Context, hook Hook, event string, data map[string]string) error {
	if hook.Type == "" {
		return nil // No hook configured
	}

	switch hook.Type {
	case HookTypeCommand:
		return e.ExecuteCommand(ctx, hook.Command, data)
	case HookTypeAgent:
		// Log agent suggestion instead of auto-invoking
		suggestion := fmt.Sprintf("💡 Suggestion: Run /ent:%s", hook.Agent)
		if hook.Prompt != "" {
			suggestion += fmt.Sprintf(" - %s", hook.Prompt)
		}
		e.logger.Info(suggestion, "event", event, "data", data)
		return nil
	default:
		return fmt.Errorf("unknown hook type: %s", hook.Type)
	}
}

// executeHook is the internal hook execution implementation.
func (e *Executor) executeHook(ctx context.Context, hook Hook, toolName string, input, result any, toolErr error) error {
	switch hook.Type {
	case HookTypeCommand:
		env := map[string]string{
			"TOOL_NAME": toolName,
		}

		// Provide tool input/result as JSON via stdin
		var inputJSON string
		if input != nil {
			if data, err := json.Marshal(map[string]any{
				"tool_name":  toolName,
				"tool_input": input,
			}); err == nil {
				inputJSON = string(data)
			}
		}
		if result != nil {
			if data, err := json.Marshal(map[string]any{
				"tool_name":   toolName,
				"tool_result": result,
				"tool_error":  toolErr,
			}); err == nil {
				inputJSON = string(data)
			}
		}

		// Run command with JSON input on stdin
		// #nosec G204 - hook.Command is validated by configuration
		cmd := exec.CommandContext(ctx, "bash", "-c", hook.Command)
		cmd.Env = os.Environ()
		for k, v := range env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
		if inputJSON != "" {
			cmd.Stdin = strings.NewReader(inputJSON)
		}

		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("hook command: %w: %s", err, output)
		}
		return nil

	case HookTypeAgent:
		// Log agent suggestion
		suggestion := fmt.Sprintf("💡 Suggestion: Run /ent:%s", hook.Agent)
		if hook.Prompt != "" {
			suggestion += fmt.Sprintf(" - %s", hook.Prompt)
		}
		e.logger.Info(suggestion, "tool", toolName)
		return nil

	default:
		return fmt.Errorf("unknown hook type: %s", hook.Type)
	}
}
