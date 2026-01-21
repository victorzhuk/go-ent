package opencode

import (
	"context"
	"os/exec"
	"strings"
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewCLIClient("")
			got := client.buildArgs(tt.provider, tt.model, tt.prompt)
			assert.Equal(t, tt.want, got)
		})
	}
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseCLIOutput(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCLIResult_Success(t *testing.T) {
	result := &CLIResult{
		Output:   "Success output",
		ExitCode: 0,
		Error:    nil,
	}

	assert.Equal(t, "Success output", result.Output)
	assert.Equal(t, 0, result.ExitCode)
	assert.NoError(t, result.Error)
}

func TestCLIResult_Error(t *testing.T) {
	err := assert.AnError
	result := &CLIResult{
		Output:   "Error output",
		ExitCode: 1,
		Error:    err,
	}

	assert.Equal(t, "Error output", result.Output)
	assert.Equal(t, 1, result.ExitCode)
	assert.Equal(t, err, result.Error)
}

func TestCLIClient_RunWithTimeout(t *testing.T) {
	client := NewCLIClient("")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	result, err := client.Run(ctx, "moonshot", "glm-4", "test")

	assert.Error(t, err)
	assert.NotNil(t, result)
}

func TestCLIClient_RunWithTimeoutHelper(t *testing.T) {
	client := NewCLIClient("")

	result, err := client.RunWithTimeout("moonshot", "glm-4", "test", 1*time.Millisecond)

	assert.Error(t, err)
	assert.NotNil(t, result)
}

func TestCLIClient_RunNonBlocking(t *testing.T) {
	client := NewCLIClient("")

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	cmd, resultChan := client.RunNonBlocking(ctx, "moonshot", "glm-4", "test")

	assert.NotNil(t, cmd)
	assert.NotNil(t, resultChan)

	select {
	case result := <-resultChan:
		assert.NotNil(t, result)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timeout waiting for result")
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
}

func TestCLIClient_Validate(t *testing.T) {
	client := NewCLIClient("")

	ctx := context.Background()

	err := client.Validate(ctx)

	if err != nil {
		t.Skipf("opencode not available: %v", err)
	}
}

func TestCLIClient_Validate_Fails(t *testing.T) {
	client := NewCLIClient("")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	err := client.Validate(ctx)

	assert.Error(t, err)
}
