package provider

import (
	"context"
	"os"
	"testing"

	"log/slog"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAnthropicClient(t *testing.T) {
	tests := []struct {
		name    string
		apiKey  string
		wantErr string
	}{
		{
			name:   "creates client with valid API key",
			apiKey: "test-key-123",
		},
		{
			name:    "fails without API key",
			wantErr: "ANTHROPIC_API_KEY environment variable not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.apiKey != "" {
				t.Setenv("ANTHROPIC_API_KEY", tt.apiKey)
			} else {
				t.Setenv("ANTHROPIC_API_KEY", "")
			}

			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			client, err := NewAnthropicClient(logger)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				assert.Nil(t, client)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, client)
				assert.Equal(t, tt.apiKey, client.apiKey)
				assert.NotNil(t, client.httpClient)
				assert.NotNil(t, client.logger)
			}
		})
	}
}

func TestClient_Complete(t *testing.T) {
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("ANTHROPIC_API_KEY not set - skipping integration test")
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	client, err := NewAnthropicClient(logger)
	require.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name    string
		model   Model
		prompt  string
		wantErr string
	}{
		{
			name:   "completes with Haiku",
			model:  ModelHaiku,
			prompt: "Say 'Hello, World!'",
		},
		{
			name:   "completes with Sonnet",
			model:  ModelSonnet,
			prompt: "Count to 3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			response, err := client.Complete(ctx, tt.model, tt.prompt)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, response)
			}
		})
	}
}

func TestClient_CompleteWithHistory(t *testing.T) {
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("ANTHROPIC_API_KEY not set - skipping integration test")
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	client, err := NewAnthropicClient(logger)
	require.NoError(t, err)

	ctx := context.Background()

	messages := []Message{
		{
			Role:    "user",
			Content: "What is 2+2?",
		},
		{
			Role:    "assistant",
			Content: "2+2 equals 4.",
		},
		{
			Role:    "user",
			Content: "What about 3+3?",
		},
	}

	response, err := client.CompleteWithHistory(ctx, ModelHaiku, messages)
	require.NoError(t, err)
	assert.NotEmpty(t, response)
}

func TestClient_Stream(t *testing.T) {
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("ANTHROPIC_API_KEY not set - skipping integration test")
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	client, err := NewAnthropicClient(logger)
	require.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name   string
		model  Model
		prompt string
	}{
		{
			name:   "streams with Haiku",
			model:  ModelHaiku,
			prompt: "Count from 1 to 5 slowly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var chunks []string
			err := client.Stream(ctx, tt.model, tt.prompt, func(text string) {
				chunks = append(chunks, text)
			})

			require.NoError(t, err)
			assert.NotEmpty(t, chunks)
			assert.NotEmpty(t, chunks[0])
		})
	}
}

func TestClient_StreamWithHistory(t *testing.T) {
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("ANTHROPIC_API_KEY not set - skipping integration test")
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	client, err := NewAnthropicClient(logger)
	require.NoError(t, err)

	ctx := context.Background()

	messages := []Message{
		{
			Role:    "user",
			Content: "Tell me a short joke",
		},
	}

	var chunks []string
	err = client.StreamWithHistory(ctx, ModelHaiku, messages, func(text string) {
		chunks = append(chunks, text)
	})

	require.NoError(t, err)
	assert.NotEmpty(t, chunks)
}

func TestModels(t *testing.T) {
	tests := []struct {
		model   Model
		wantVal string
	}{
		{ModelHaiku, "claude-3-haiku-20240307"},
		{ModelSonnet, "claude-3-5-sonnet-20241022"},
		{ModelOpus, "claude-3-opus-20240229"},
	}

	for _, tt := range tests {
		t.Run(tt.wantVal, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.wantVal, string(tt.model))
		})
	}
}

func TestClient_Validate(t *testing.T) {
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Skip("ANTHROPIC_API_KEY not set - skipping integration test")
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	client, err := NewAnthropicClient(logger)
	require.NoError(t, err)

	ctx := context.Background()

	err = client.Validate(ctx)

	assert.NoError(t, err)
}

func TestClient_Validate_Fails(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "invalid-key")

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	_, err := NewAnthropicClient(logger)

	assert.NoError(t, err)

	client, err := NewAnthropicClient(logger)
	require.NoError(t, err)

	ctx := context.Background()

	err = client.Validate(ctx)

	assert.Error(t, err)
}
