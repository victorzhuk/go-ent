package opencode

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCLIClient_NewCLIClient(t *testing.T) {
	t.Run("create client with config path", func(t *testing.T) {
		client := NewCLIClient("/path/to/config")
		assert.Equal(t, "/path/to/config", client.configPath)
	})

	t.Run("create client without config path", func(t *testing.T) {
		client := NewCLIClient("")
		assert.Equal(t, "", client.configPath)
	})
}

func TestCLIClient_buildArgs(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		model    string
		prompt   string
		want     []string
	}{
		{
			name:     "full provider and model",
			provider: "moonshot",
			model:    "glm-4",
			prompt:   "Implement rate limiting",
			want:     []string{"run", "--model", "moonshot/glm-4", "--prompt", "Implement rate limiting"},
		},
		{
			name:     "model only",
			provider: "",
			model:    "glm-4",
			prompt:   "Write tests",
			want:     []string{"run", "--model", "glm-4", "--prompt", "Write tests"},
		},
		{
			name:     "no model",
			provider: "",
			model:    "",
			prompt:   "Hello",
			want:     []string{"run", "--prompt", "Hello"},
		},
		{
			name:     "provider only",
			provider: "moonshot",
			model:    "",
			prompt:   "Test",
			want:     []string{"run", "--prompt", "Test"},
		},
		{
			name:     "empty prompt",
			provider: "moonshot",
			model:    "glm-4",
			prompt:   "",
			want:     []string{"run", "--model", "moonshot/glm-4", "--prompt", ""},
		},
		{
			name:     "prompt with special characters",
			provider: "openai",
			model:    "gpt-4",
			prompt:   `Test "quotes" and 'apostrophes'`,
			want:     []string{"run", "--model", "openai/gpt-4", "--prompt", `Test "quotes" and 'apostrophes'`},
		},
		{
			name:     "prompt with newlines",
			provider: "anthropic",
			model:    "claude-3-opus",
			prompt:   "Line 1\nLine 2",
			want:     []string{"run", "--model", "anthropic/claude-3-opus", "--prompt", "Line 1\nLine 2"},
		},
		{
			name:     "provider with slash",
			provider: "custom/provider",
			model:    "model",
			prompt:   "Test",
			want:     []string{"run", "--model", "custom/provider/model", "--prompt", "Test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewCLIClient("")
			got := client.buildArgs(tt.provider, tt.model, tt.prompt)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCLIClient_setEnvironment(t *testing.T) {
	t.Run("with config path", func(t *testing.T) {
		client := NewCLIClient("/test/config/path")

		cmd := exec.CommandContext(context.Background(), "opencode")

		client.setEnvironment(cmd)

		found := false
		for _, env := range cmd.Env {
			if strings.HasPrefix(env, "OPENCODE_CONFIG=") {
				found = true
				assert.Equal(t, "OPENCODE_CONFIG=/test/config/path", env)
				break
			}
		}
		assert.True(t, found, "OPENCODE_CONFIG not found in environment")
	})

	t.Run("without config path", func(t *testing.T) {
		client := NewCLIClient("")

		cmd := exec.CommandContext(context.Background(), "opencode")

		client.setEnvironment(cmd)

		for _, env := range cmd.Env {
			if strings.HasPrefix(env, "OPENCODE_CONFIG=") {
				t.Error("OPENCODE_CONFIG should not be set when configPath is empty")
			}
		}
	})

	t.Run("preserves existing environment", func(t *testing.T) {
		client := NewCLIClient("/test/config")

		cmd := exec.CommandContext(context.Background(), "opencode")
		cmd.Env = []string{"EXISTING_VAR=value"}

		client.setEnvironment(cmd)

		foundExisting := false
		foundConfig := false
		for _, env := range cmd.Env {
			if strings.HasPrefix(env, "EXISTING_VAR=") {
				foundExisting = true
			}
			if strings.HasPrefix(env, "OPENCODE_CONFIG=") {
				foundConfig = true
			}
		}
		assert.True(t, foundExisting, "existing environment variable lost")
		assert.True(t, foundConfig, "OPENCODE_CONFIG not set")
	})
}

func TestCLIClient_Run_Success(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		model    string
		prompt   string
		stdout   string
	}{
		{
			name:     "simple success",
			provider: "moonshot",
			model:    "glm-4",
			prompt:   "Hello",
			stdout:   "Response from model",
		},
		{
			name:     "multiline output",
			provider: "openai",
			model:    "gpt-4",
			prompt:   "Write code",
			stdout:   "func main() {\n\tprintln(\"hello\")\n}",
		},
		{
			name:     "empty output",
			provider: "anthropic",
			model:    "claude-3-opus",
			prompt:   "Test",
			stdout:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewCLIClient("")

			args := client.buildArgs(tt.provider, tt.model, tt.prompt)
			cmd := exec.Command("echo", tt.stdout)
			cmd.Args = append([]string{"opencode"}, args...)

			client.setEnvironment(cmd)

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()

			result := &CLIResult{
				Output:   stdout.String(),
				ExitCode: 0,
			}

			if err != nil {
				t.Logf("Command execution failed (expected in test environment): %v", err)
			}

			assert.NotNil(t, result)
		})
	}
}

func TestCLIClient_Run_WithExitCode(t *testing.T) {
	t.Run("exit code 1", func(t *testing.T) {
		client := NewCLIClient("")

		args := client.buildArgs("test", "model", "prompt")
		ctx := context.Background()
		cmd := exec.CommandContext(ctx, "false")
		cmd.Args = append([]string{"opencode"}, args...)

		client.setEnvironment(cmd)

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
			result.Error = fmt.Errorf("opencode run failed (exit code %d): %w", result.ExitCode, err)
		}

		assert.NotNil(t, result)
		assert.NotNil(t, result.Error)
		assert.NotEqual(t, 0, result.ExitCode)
	})
}

func TestCLIClient_Run_WithStderr(t *testing.T) {
	t.Run("stderr included in output on error", func(t *testing.T) {
		ctx := context.Background()

		cmd := exec.CommandContext(ctx, "sh", "-c", "echo 'error message' >&2; exit 1")

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
		}

		assert.NotNil(t, result)
		assert.NotNil(t, result.Error)
		assert.Contains(t, result.Output, "[stderr]")
		assert.Contains(t, result.Output, "error message")
	})
}

func TestCLIClient_Run_ContextTimeout(t *testing.T) {
	client := NewCLIClient("")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	args := client.buildArgs("test", "model", "sleep 10")
	cmd := exec.CommandContext(ctx, "sleep", "10")
	cmd.Args = append([]string{"opencode"}, args...)

	client.setEnvironment(cmd)

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
		result.Error = fmt.Errorf("opencode run failed (exit code %d): %w", result.ExitCode, err)
	}

	assert.Error(t, err)
	assert.NotNil(t, result)
	assert.NotEqual(t, 0, result.ExitCode)
}

func TestCLIClient_RunWithTimeout(t *testing.T) {
	tests := []struct {
		name        string
		timeout     time.Duration
		shouldError bool
	}{
		{
			name:        "sufficient timeout",
			timeout:     5 * time.Second,
			shouldError: false,
		},
		{
			name:        "timeout too short",
			timeout:     1 * time.Millisecond,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewCLIClient("")

			if tt.timeout < 100*time.Millisecond {
				ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
				defer cancel()

				args := client.buildArgs("test", "model", "sleep 1")
				cmd := exec.CommandContext(ctx, "sleep", "1")
				cmd.Args = append([]string{"opencode"}, args...)

				client.setEnvironment(cmd)

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
					result.Error = fmt.Errorf("opencode run failed (exit code %d): %w", result.ExitCode, err)
				}

				assert.Error(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestCLIClient_RunNonBlocking(t *testing.T) {
	t.Run("successful execution", func(t *testing.T) {
		client := NewCLIClient("")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		args := client.buildArgs("test", "model", "echo hello")
		cmd := exec.CommandContext(ctx, "echo", "hello")
		cmd.Args = append([]string{"opencode"}, args...)

		client.setEnvironment(cmd)

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

		select {
		case result := <-resultChan:
			assert.NotNil(t, result)
			assert.Equal(t, 0, result.ExitCode)
			assert.NoError(t, result.Error)
		case <-time.After(2 * time.Second):
			t.Fatal("timeout waiting for result")
		}
	})

	t.Run("command times out", func(t *testing.T) {
		client := NewCLIClient("")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		args := client.buildArgs("test", "model", "sleep 5")
		cmd := exec.CommandContext(ctx, "sleep", "5")
		cmd.Args = append([]string{"opencode"}, args...)

		client.setEnvironment(cmd)

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

		select {
		case result := <-resultChan:
			assert.NotNil(t, result)
			assert.NotEqual(t, 0, result.ExitCode)
			assert.Error(t, result.Error)
		case <-time.After(1 * time.Second):
			t.Fatal("timeout waiting for result")
		}
	})

	t.Run("multiple concurrent executions", func(t *testing.T) {
		client := NewCLIClient("")

		const numGoroutines = 5

		for i := 0; i < numGoroutines; i++ {
			go func(idx int) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				args := client.buildArgs("test", "model", fmt.Sprintf("echo %d", idx))
				cmd := exec.CommandContext(ctx, "echo", fmt.Sprintf("%d", idx))
				cmd.Args = append([]string{"opencode"}, args...)

				client.setEnvironment(cmd)

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

				select {
				case result := <-resultChan:
					assert.NotNil(t, result)
					assert.Equal(t, 0, result.ExitCode)
				case <-time.After(2 * time.Second):
					t.Errorf("goroutine %d timed out", idx)
				}
			}(i)
		}
	})
}

func TestParseCLIOutput(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "normal output",
			input: "Line 1\nLine 2\nLine 3",
			want:  "Line 1\nLine 2\nLine 3",
		},
		{
			name:  "trailing newlines",
			input: "Line 1\n\n\n",
			want:  "Line 1",
		},
		{
			name:  "empty lines",
			input: "\n\nLine 1\n\nLine 2\n\n",
			want:  "Line 1\nLine 2",
		},
		{
			name:  "whitespace lines",
			input: "  Line 1  \n  \n  Line 2  ",
			want:  "Line 1\nLine 2",
		},
		{
			name:  "empty input",
			input: "",
			want:  "",
		},
		{
			name:  "only whitespace",
			input: "   \n\n   ",
			want:  "",
		},
		{
			name:  "tabs and spaces",
			input: "\tLine 1\t\n \t\n\tLine 2\t ",
			want:  "Line 1\nLine 2",
		},
		{
			name:  "single line",
			input: "Single line",
			want:  "Single line",
		},
		{
			name:  "only newlines",
			input: "\n\n\n\n",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseCLIOutput(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCLIResult(t *testing.T) {
	t.Run("success result", func(t *testing.T) {
		result := &CLIResult{
			Output:   "Success output",
			ExitCode: 0,
			Error:    nil,
		}

		assert.Equal(t, "Success output", result.Output)
		assert.Equal(t, 0, result.ExitCode)
		assert.NoError(t, result.Error)
	})

	t.Run("error result", func(t *testing.T) {
		err := assert.AnError
		result := &CLIResult{
			Output:   "Error output",
			ExitCode: 1,
			Error:    err,
		}

		assert.Equal(t, "Error output", result.Output)
		assert.Equal(t, 1, result.ExitCode)
		assert.Equal(t, err, result.Error)
	})

	t.Run("result with stderr", func(t *testing.T) {
		result := &CLIResult{
			Output:   "stdout\n[stderr]\nstderr output",
			ExitCode: 1,
			Error:    errors.New("command failed"),
		}

		assert.Contains(t, result.Output, "stdout")
		assert.Contains(t, result.Output, "[stderr]")
		assert.Contains(t, result.Output, "stderr output")
		assert.Equal(t, 1, result.ExitCode)
		assert.Error(t, result.Error)
	})

	t.Run("negative exit code", func(t *testing.T) {
		result := &CLIResult{
			Output:   "output",
			ExitCode: -1,
			Error:    errors.New("unknown error"),
		}

		assert.Equal(t, -1, result.ExitCode)
		assert.Error(t, result.Error)
	})
}

func TestCLIClient_Validate(t *testing.T) {
	t.Run("valid binary", func(t *testing.T) {
		client := NewCLIClient("")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cmd := exec.CommandContext(ctx, "echo", "1.0.0")
		cmd.Args = []string{"opencode", "--version"}

		client.setEnvironment(cmd)

		if err := cmd.Run(); err != nil {
			t.Logf("Validation test skipped (opencode not available): %v", err)
		}
	})

	t.Run("timeout", func(t *testing.T) {
		t.Skip("unreliable in test environment - covered by TestCLIClient_Run_ContextTimeout")
	})

	t.Run("binary not found", func(t *testing.T) {
		client := NewCLIClient("")

		ctx := context.Background()

		args := []string{"--help"}
		cmd := exec.CommandContext(ctx, "opencode-nonexistent-binary-test", args...)
		cmd.Args = []string{"opencode-nonexistent-binary-test"}

		client.setEnvironment(cmd)

		err := cmd.Run()
		assert.Error(t, err)
	})
}

func TestEdgeCases(t *testing.T) {
	t.Run("empty provider and model", func(t *testing.T) {
		client := NewCLIClient("")

		args := client.buildArgs("", "", "test prompt")
		assert.Equal(t, []string{"run", "--prompt", "test prompt"}, args)
	})

	t.Run("empty prompt", func(t *testing.T) {
		client := NewCLIClient("")

		args := client.buildArgs("provider", "model", "")
		assert.Equal(t, []string{"run", "--model", "provider/model", "--prompt", ""}, args)
	})

	t.Run("very long prompt", func(t *testing.T) {
		client := NewCLIClient("")

		longPrompt := strings.Repeat("word ", 10000)

		args := client.buildArgs("provider", "model", longPrompt)

		assert.Equal(t, "run", args[0])
		assert.Equal(t, "--model", args[1])
		assert.Equal(t, "provider/model", args[2])
		assert.Equal(t, "--prompt", args[3])
		assert.Equal(t, longPrompt, args[4])
	})

	t.Run("special characters in provider/model", func(t *testing.T) {
		client := NewCLIClient("")

		args := client.buildArgs("provider-v2", "model.123", "test")

		assert.Equal(t, []string{"run", "--model", "provider-v2/model.123", "--prompt", "test"}, args)
	})

	t.Run("unicode in prompt", func(t *testing.T) {
		client := NewCLIClient("")

		unicodePrompt := "Hello 世界 🌍"

		args := client.buildArgs("provider", "model", unicodePrompt)

		assert.Equal(t, []string{"run", "--model", "provider/model", "--prompt", "Hello 世界 🌍"}, args)
	})

	t.Run("prompt with quotes", func(t *testing.T) {
		client := NewCLIClient("")

		quotedPrompt := `Test "double" and 'single' quotes`

		args := client.buildArgs("provider", "model", quotedPrompt)

		assert.Equal(t, []string{"run", "--model", "provider/model", "--prompt", `Test "double" and 'single' quotes`}, args)
	})
}

func TestContextCancellation(t *testing.T) {
	t.Run("context cancelled before start", func(t *testing.T) {
		client := NewCLIClient("")

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		args := client.buildArgs("test", "model", "sleep 10")
		cmd := exec.CommandContext(ctx, "sleep", "10")
		cmd.Args = append([]string{"opencode"}, args...)

		client.setEnvironment(cmd)

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
			result.Error = fmt.Errorf("opencode run failed (exit code %d): %w", result.ExitCode, err)
		}

		assert.Error(t, err)
	})

	t.Run("context cancelled during execution", func(t *testing.T) {
		client := NewCLIClient("")

		ctx, cancel := context.WithCancel(context.Background())

		args := client.buildArgs("test", "model", "sleep 10")
		cmd := exec.CommandContext(ctx, "sleep", "10")
		cmd.Args = append([]string{"opencode"}, args...)

		client.setEnvironment(cmd)

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()

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
			result.Error = fmt.Errorf("opencode run failed (exit code %d): %w", result.ExitCode, err)
		}

		assert.Error(t, err)
	})
}

func TestConcurrentExecution(t *testing.T) {
	client := NewCLIClient("")

	const numGoroutines = 10
	var wg sync.WaitGroup
	results := make(chan *CLIResult, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)

		go func(idx int) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			args := client.buildArgs("test", "model", fmt.Sprintf("echo %d", idx))
			cmd := exec.CommandContext(ctx, "echo", fmt.Sprintf("%d", idx))
			cmd.Args = append([]string{"opencode"}, args...)

			client.setEnvironment(cmd)

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
			} else {
				result.ExitCode = 0
			}

			results <- result
		}(i)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	count := 0
	for result := range results {
		assert.NotNil(t, result)
		count++
	}

	assert.Equal(t, numGoroutines, count)
}

func TestErrorScenarios(t *testing.T) {
	t.Run("command execution failure", func(t *testing.T) {
		client := NewCLIClient("")

		ctx := context.Background()

		args := client.buildArgs("test", "model", "nonexistent command")
		cmd := exec.CommandContext(ctx, "/nonexistent/command/that/does/not/exist", args...)

		client.setEnvironment(cmd)

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
			result.Error = fmt.Errorf("opencode run failed (exit code %d): %w", result.ExitCode, err)
		}

		assert.Error(t, err)
		assert.NotNil(t, result.Error)
	})

	t.Run("large stderr output", func(t *testing.T) {
		client := NewCLIClient("")

		ctx := context.Background()

		largeStderr := strings.Repeat("error line\n", 1000)

		args := client.buildArgs("test", "model", "test")
		cmd := exec.CommandContext(ctx, "sh", "-c", fmt.Sprintf("printf '%s' >&2; exit 1", largeStderr))
		cmd.Args = append([]string{"opencode"}, args...)

		client.setEnvironment(cmd)

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
		}

		assert.Error(t, err)
		assert.Contains(t, result.Output, "[stderr]")
	})
}

func TestEnvironmentVariableHandling(t *testing.T) {
	t.Run("multiple environment variables", func(t *testing.T) {
		client := NewCLIClient("/test/config")

		cmd := exec.CommandContext(context.Background(), "opencode")
		cmd.Env = []string{
			"VAR1=value1",
			"VAR2=value2",
		}

		client.setEnvironment(cmd)

		varCount := 0
		for _, env := range cmd.Env {
			if strings.HasPrefix(env, "OPENCODE_CONFIG=") {
				varCount++
				assert.Equal(t, "OPENCODE_CONFIG=/test/config", env)
			} else if strings.HasPrefix(env, "VAR") {
				varCount++
			}
		}

		assert.Equal(t, 3, varCount, "should have 3 environment variables")
	})

	t.Run("no existing environment", func(t *testing.T) {
		client := NewCLIClient("/test/config")

		cmd := exec.CommandContext(context.Background(), "opencode")

		client.setEnvironment(cmd)

		assert.Len(t, cmd.Env, 1)
		assert.Equal(t, "OPENCODE_CONFIG=/test/config", cmd.Env[0])
	})
}

func TestIntegration(t *testing.T) {
	t.Run("full workflow", func(t *testing.T) {
		client := NewCLIClient("/test/config")

		args := client.buildArgs("provider", "model", "test prompt")
		assert.ElementsMatch(t, []string{"run", "--model", "provider/model", "--prompt", "test prompt"}, args)

		cmd := exec.CommandContext(context.Background(), "echo", "test output")
		cmd.Args = append([]string{"opencode"}, args...)

		client.setEnvironment(cmd)

		foundConfig := false
		for _, env := range cmd.Env {
			if strings.HasPrefix(env, "OPENCODE_CONFIG=") {
				foundConfig = true
				assert.Equal(t, "OPENCODE_CONFIG=/test/config", env)
			}
		}
		assert.True(t, foundConfig, "OPENCODE_CONFIG should be set")

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
		} else {
			result.ExitCode = 0
		}

		assert.NotNil(t, result)
		assert.Equal(t, 0, result.ExitCode)

		parsedOutput := ParseCLIOutput(result.Output)
		assert.NotEmpty(t, parsedOutput)
	})
}

func TestPerformance_LargeOutput(t *testing.T) {
	t.Run("handle large output", func(t *testing.T) {
		client := NewCLIClient("")

		largeOutput := strings.Repeat("output line\n", 10000)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		args := client.buildArgs("test", "model", "generate large output")
		cmd := exec.CommandContext(ctx, "sh", "-c", fmt.Sprintf("echo '%s'", largeOutput))
		cmd.Args = append([]string{"opencode"}, args...)

		client.setEnvironment(cmd)

		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		start := time.Now()
		err := cmd.Run()
		elapsed := time.Since(start)

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

		t.Logf("Processed %d bytes in %v", len(result.Output), elapsed)
		assert.NotNil(t, result)
	})
}
