package opencode

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type CLIResult struct {
	Output   string
	ExitCode int
	Error    error
}

type CLIClient struct {
	configPath string
}

func NewCLIClient(configPath string) *CLIClient {
	return &CLIClient{
		configPath: configPath,
	}
}

func (c *CLIClient) Run(ctx context.Context, provider, model, prompt string) (*CLIResult, error) {
	args := c.buildArgs(provider, model, prompt)

	cmd := exec.CommandContext(ctx, "opencode", args...) // #nosec G204 -- controlled binary path

	c.setEnvironment(cmd)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := &CLIResult{
		Output: stdout.String(),
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = -1
		}

		if stderr.Len() > 0 {
			result.Output += fmt.Sprintf("\n[stderr]\n%s", stderr.String())
		}

		result.Error = fmt.Errorf("opencode run failed (exit code %d): %w", result.ExitCode, err)
		return result, result.Error
	}

	result.ExitCode = 0
	return result, nil
}

func (c *CLIClient) RunWithTimeout(provider, model, prompt string, timeout time.Duration) (*CLIResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return c.Run(ctx, provider, model, prompt)
}

func (c *CLIClient) RunNonBlocking(ctx context.Context, provider, model, prompt string) (*exec.Cmd, <-chan *CLIResult) {
	args := c.buildArgs(provider, model, prompt)

	cmd := exec.CommandContext(ctx, "opencode", args...) // #nosec G204 -- controlled binary path

	c.setEnvironment(cmd)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	resultChan := make(chan *CLIResult, 1)

	go func() {
		err := cmd.Run()

		result := &CLIResult{
			Output: stdout.String(),
		}

		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				result.ExitCode = exitErr.ExitCode()
			} else {
				result.ExitCode = -1
			}

			if stderr.Len() > 0 {
				result.Output += fmt.Sprintf("\n[stderr]\n%s", stderr.String())
			}

			result.Error = fmt.Errorf("opencode run failed (exit code %d): %w", result.ExitCode, err)
		} else {
			result.ExitCode = 0
		}

		resultChan <- result
	}()

	return cmd, resultChan
}

func (c *CLIClient) buildArgs(provider, model, prompt string) []string {
	args := []string{"run"}

	if provider != "" && model != "" {
		fullModel := fmt.Sprintf("%s/%s", provider, model)
		args = append(args, "--model", fullModel)
	} else if model != "" {
		args = append(args, "--model", model)
	}

	args = append(args, "--prompt", prompt)

	return args
}

func (c *CLIClient) setEnvironment(cmd *exec.Cmd) {
	if c.configPath != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("OPENCODE_CONFIG=%s", c.configPath))
	}
}

func ParseCLIOutput(output string) string {
	lines := strings.Split(output, "\n")

	var cleaned []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			cleaned = append(cleaned, trimmed)
		}
	}

	return strings.Join(cleaned, "\n")
}

func (c *CLIClient) Validate(ctx context.Context) error {
	validateCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	args := []string{"run", "--help"}
	cmd := exec.CommandContext(validateCtx, "opencode", args...)

	c.setEnvironment(cmd)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("opencode validation failed: %w", err)
	}

	return nil
}
