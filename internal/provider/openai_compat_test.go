package provider

import (
	"context"
	"os"
	"testing"

	"log/slog"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOpenAICompatClient(t *testing.T) {
	tests := []struct {
		name     string
		provider ProviderType
		apiKey   string
		envVar   string
		wantErr  string
		wantURL  string
	}{
		{
			name:     "creates moonshot client with valid API key",
			provider: ProviderMoonshot,
			apiKey:   "moonshot-key-123",
			wantURL:  "https://api.moonshot.cn/v1",
		},
		{
			name:     "creates deepseek client with valid API key",
			provider: ProviderDeepSeek,
			apiKey:   "deepseek-key-456",
			wantURL:  "https://api.deepseek.com/v1",
		},
		{
			name:     "creates openai client with valid API key",
			provider: ProviderOpenAI,
			apiKey:   "openai-key-789",
			wantURL:  "https://api.openai.com/v1",
		},
		{
			name:     "fails moonshot without API key",
			provider: ProviderMoonshot,
			wantErr:  "MOONSHOT_API_KEY environment variable not set",
		},
		{
			name:     "fails deepseek without API key",
			provider: ProviderDeepSeek,
			wantErr:  "DEEPSEEK_API_KEY environment variable not set",
		},
		{
			name:     "fails openai without API key",
			provider: ProviderOpenAI,
			wantErr:  "OPENAI_API_KEY environment variable not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.provider == ProviderMoonshot {
				if tt.apiKey != "" {
					t.Setenv("MOONSHOT_API_KEY", tt.apiKey)
				} else {
					t.Setenv("MOONSHOT_API_KEY", "")
				}
			} else if tt.provider == ProviderDeepSeek {
				if tt.apiKey != "" {
					t.Setenv("DEEPSEEK_API_KEY", tt.apiKey)
				} else {
					t.Setenv("DEEPSEEK_API_KEY", "")
				}
			} else if tt.provider == ProviderOpenAI {
				if tt.apiKey != "" {
					t.Setenv("OPENAI_API_KEY", tt.apiKey)
				} else {
					t.Setenv("OPENAI_API_KEY", "")
				}
			}

			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			client, err := NewOpenAICompatClient(tt.provider, logger)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				assert.Nil(t, client)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, client)
				assert.Equal(t, tt.apiKey, client.apiKey)
				assert.Equal(t, tt.wantURL, client.baseURL)
				assert.NotNil(t, client.httpClient)
				assert.NotNil(t, client.logger)
			}
		})
	}
}

func TestNewOpenAICompatClientWithConfig(t *testing.T) {
	tests := []struct {
		name      string
		baseURL   string
		apiKeyEnv string
		apiKey    string
		wantErr   string
	}{
		{
			name:      "creates client with custom config",
			baseURL:   "https://custom.api.com/v1",
			apiKeyEnv: "CUSTOM_API_KEY",
			apiKey:    "custom-key",
		},
		{
			name:      "fails with empty base URL",
			baseURL:   "",
			apiKeyEnv: "CUSTOM_API_KEY",
			apiKey:    "custom-key",
			wantErr:   "base URL cannot be empty",
		},
		{
			name:      "fails with empty API key env var",
			baseURL:   "https://custom.api.com/v1",
			apiKeyEnv: "",
			apiKey:    "custom-key",
			wantErr:   "API key environment variable name cannot be empty",
		},
		{
			name:      "fails when env var not set",
			baseURL:   "https://custom.api.com/v1",
			apiKeyEnv: "UNSET_API_KEY",
			wantErr:   "UNSET_API_KEY environment variable not set",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.apiKey != "" && tt.apiKeyEnv != "" {
				t.Setenv(tt.apiKeyEnv, tt.apiKey)
			}

			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			client, err := NewOpenAICompatClientWithConfig(tt.baseURL, tt.apiKeyEnv, logger)

			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				assert.Nil(t, client)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, client)
				assert.Equal(t, tt.apiKey, client.apiKey)
				assert.Equal(t, tt.baseURL, client.baseURL)
				assert.NotNil(t, client.httpClient)
				assert.NotNil(t, client.logger)
			}
		})
	}
}

func TestOpenAIProviderTypes(t *testing.T) {
	tests := []struct {
		provider ProviderType
		wantVal  string
	}{
		{ProviderMoonshot, "moonshot"},
		{ProviderDeepSeek, "deepseek"},
		{ProviderOpenAI, "openai"},
	}

	for _, tt := range tests {
		t.Run(tt.wantVal, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.wantVal, string(tt.provider))
		})
	}
}

func TestOpenAIModels(t *testing.T) {
	tests := []struct {
		model   OpenAIModel
		wantVal string
	}{
		{ModelGLM4, "glm-4"},
		{ModelGLM4Plus, "glm-4-plus"},
		{ModelGLM3Turbo, "glm-3-turbo"},
		{ModelDeepSeekV3, "deepseek-chat"},
		{ModelDeepSeekV3R, "deepseek-reasoner"},
	}

	for _, tt := range tests {
		t.Run(tt.wantVal, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.wantVal, string(tt.model))
		})
	}
}

func TestOpenAIClient_Complete(t *testing.T) {
	if os.Getenv("MOONSHOT_API_KEY") == "" && os.Getenv("DEEPSEEK_API_KEY") == "" {
		t.Skip("No OpenAI-compatible API keys set - skipping integration test")
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	tests := []struct {
		name     string
		provider ProviderType
		model    OpenAIModel
		prompt   string
	}{
		{
			name:     "completes with Moonshot GLM-4",
			provider: ProviderMoonshot,
			model:    ModelGLM4,
			prompt:   "Say 'Hello, World!'",
		},
		{
			name:     "completes with DeepSeek V3",
			provider: ProviderDeepSeek,
			model:    ModelDeepSeekV3,
			prompt:   "Count to 3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client, err := NewOpenAICompatClient(tt.provider, logger)
			if err != nil {
				t.Skipf("Failed to create client: %v", err)
			}

			ctx := context.Background()
			response, err := client.Complete(ctx, tt.model, tt.prompt)

			require.NoError(t, err)
			assert.NotEmpty(t, response)
		})
	}
}

func TestOpenAIClient_CompleteWithHistory(t *testing.T) {
	if os.Getenv("MOONSHOT_API_KEY") == "" && os.Getenv("DEEPSEEK_API_KEY") == "" {
		t.Skip("No OpenAI-compatible API keys set - skipping integration test")
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	client, err := NewOpenAICompatClient(ProviderMoonshot, logger)
	if err != nil {
		client, err = NewOpenAICompatClient(ProviderDeepSeek, logger)
		if err != nil {
			t.Skip("No available OpenAI-compatible provider")
		}
	}

	ctx := context.Background()

	messages := []OpenAIMessage{
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

	response, err := client.CompleteWithHistory(ctx, ModelGLM4, messages)
	require.NoError(t, err)
	assert.NotEmpty(t, response)
}

func TestOpenAIClient_Stream(t *testing.T) {
	if os.Getenv("MOONSHOT_API_KEY") == "" && os.Getenv("DEEPSEEK_API_KEY") == "" {
		t.Skip("No OpenAI-compatible API keys set - skipping integration test")
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	tests := []struct {
		name     string
		provider ProviderType
		model    OpenAIModel
		prompt   string
	}{
		{
			name:     "streams with Moonshot GLM-4",
			provider: ProviderMoonshot,
			model:    ModelGLM4,
			prompt:   "Count from 1 to 5 slowly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client, err := NewOpenAICompatClient(tt.provider, logger)
			if err != nil {
				t.Skipf("Failed to create client: %v", err)
			}

			ctx := context.Background()

			var chunks []string
			err = client.Stream(ctx, tt.model, tt.prompt, func(text string) {
				chunks = append(chunks, text)
			})

			require.NoError(t, err)
			assert.NotEmpty(t, chunks)
			assert.NotEmpty(t, chunks[0])
		})
	}
}

func TestOpenAIClient_StreamWithHistory(t *testing.T) {
	if os.Getenv("MOONSHOT_API_KEY") == "" && os.Getenv("DEEPSEEK_API_KEY") == "" {
		t.Skip("No OpenAI-compatible API keys set - skipping integration test")
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	client, err := NewOpenAICompatClient(ProviderMoonshot, logger)
	if err != nil {
		client, err = NewOpenAICompatClient(ProviderDeepSeek, logger)
		if err != nil {
			t.Skip("No available OpenAI-compatible provider")
		}
	}

	ctx := context.Background()

	messages := []OpenAIMessage{
		{
			Role:    "user",
			Content: "Tell me a short joke",
		},
	}

	var chunks []string
	err = client.StreamWithHistory(ctx, ModelGLM4, messages, func(text string) {
		chunks = append(chunks, text)
	})

	require.NoError(t, err)
	assert.NotEmpty(t, chunks)
}

func TestOpenAIClient_Validate(t *testing.T) {
	if os.Getenv("MOONSHOT_API_KEY") == "" && os.Getenv("DEEPSEEK_API_KEY") == "" {
		t.Skip("No OpenAI-compatible API keys set - skipping integration test")
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	client, err := NewOpenAICompatClient(ProviderMoonshot, logger)
	if err != nil {
		client, err = NewOpenAICompatClient(ProviderDeepSeek, logger)
		if err != nil {
			t.Skip("No available OpenAI-compatible provider")
		}
	}

	ctx := context.Background()

	err = client.Validate(ctx)

	assert.NoError(t, err)
}

func TestOpenAIClient_Validate_Fails(t *testing.T) {
	t.Setenv("MOONSHOT_API_KEY", "invalid-key")

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	_, err := NewOpenAICompatClient(ProviderMoonshot, logger)

	assert.NoError(t, err)

	client, err := NewOpenAICompatClient(ProviderMoonshot, logger)
	require.NoError(t, err)

	ctx := context.Background()

	err = client.Validate(ctx)

	assert.Error(t, err)
}
