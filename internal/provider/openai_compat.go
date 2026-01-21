package provider

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
)

type ProviderType string

const (
	ProviderMoonshot ProviderType = "moonshot"
	ProviderDeepSeek ProviderType = "deepseek"
	ProviderOpenAI   ProviderType = "openai"
)

type OpenAIModel string

const (
	ModelGLM4        OpenAIModel = "glm-4"
	ModelGLM4Plus    OpenAIModel = "glm-4-plus"
	ModelGLM3Turbo   OpenAIModel = "glm-3-turbo"
	ModelDeepSeekV3  OpenAIModel = "deepseek-chat"
	ModelDeepSeekV3R OpenAIModel = "deepseek-reasoner"
)

type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIClient struct {
	baseURL     string
	apiKey      string
	httpClient  *http.Client
	logger      *slog.Logger
	rateLimiter *RateLimiter
	retryConfig RetryConfig
}

type OpenAIStreamChunk struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index int `json:"index"`
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

type OpenAIRequest struct {
	Model     string          `json:"model"`
	Messages  []OpenAIMessage `json:"messages"`
	Stream    bool            `json:"stream"`
	MaxTokens int             `json:"max_tokens,omitempty"`
}

type OpenAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func NewOpenAICompatClient(provider ProviderType, logger *slog.Logger) (*OpenAIClient, error) {
	var baseURL, envVar string

	switch provider {
	case ProviderMoonshot:
		baseURL = "https://api.moonshot.cn/v1"
		envVar = "MOONSHOT_API_KEY"
	case ProviderDeepSeek:
		baseURL = "https://api.deepseek.com/v1"
		envVar = "DEEPSEEK_API_KEY"
	case ProviderOpenAI:
		baseURL = "https://api.openai.com/v1"
		envVar = "OPENAI_API_KEY"
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}

	apiKey := os.Getenv(envVar)
	if apiKey == "" {
		return nil, fmt.Errorf("%s environment variable not set", envVar)
	}

	return &OpenAIClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
		logger:      logger,
		rateLimiter: NewRateLimiter(50, logger),
		retryConfig: DefaultRetryConfig(),
	}, nil
}

func NewOpenAICompatClientWithConfig(baseURL, apiKeyEnv string, logger *slog.Logger) (*OpenAIClient, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("base URL cannot be empty")
	}

	if apiKeyEnv == "" {
		return nil, fmt.Errorf("API key environment variable name cannot be empty")
	}

	apiKey := os.Getenv(apiKeyEnv)
	if apiKey == "" {
		return nil, fmt.Errorf("%s environment variable not set", apiKeyEnv)
	}

	return &OpenAIClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
		logger:      logger,
		rateLimiter: NewRateLimiter(50, logger),
		retryConfig: DefaultRetryConfig(),
	}, nil
}

func (c *OpenAIClient) Complete(ctx context.Context, model OpenAIModel, prompt string) (string, error) {
	req := OpenAIRequest{
		Model:     string(model),
		MaxTokens: 4096,
		Messages: []OpenAIMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Stream: false,
	}

	return c.sendRequest(ctx, req)
}

func (c *OpenAIClient) CompleteWithHistory(ctx context.Context, model OpenAIModel, messages []OpenAIMessage) (string, error) {
	req := OpenAIRequest{
		Model:     string(model),
		MaxTokens: 4096,
		Messages:  messages,
		Stream:    false,
	}

	return c.sendRequest(ctx, req)
}

func (c *OpenAIClient) Stream(ctx context.Context, model OpenAIModel, prompt string, callback func(text string)) error {
	req := OpenAIRequest{
		Model:     string(model),
		MaxTokens: 4096,
		Messages: []OpenAIMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Stream: true,
	}

	return c.streamRequest(ctx, req, callback)
}

func (c *OpenAIClient) StreamWithHistory(ctx context.Context, model OpenAIModel, messages []OpenAIMessage, callback func(text string)) error {
	req := OpenAIRequest{
		Model:     string(model),
		MaxTokens: 4096,
		Messages:  messages,
		Stream:    true,
	}

	return c.streamRequest(ctx, req, callback)
}

func (c *OpenAIClient) sendRequest(ctx context.Context, req OpenAIRequest) (string, error) {
	var lastErr error
	var statusCode int

	for attempt := 1; attempt <= c.retryConfig.MaxAttempts; attempt++ {
		if attempt > 1 {
			backoff := calculateBackoff(attempt, c.retryConfig)
			c.logger.Info("retrying request", "attempt", attempt, "max_attempts", c.retryConfig.MaxAttempts, "delay", backoff)
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return "", ctx.Err()
			}
		}

		if err := c.rateLimiter.Wait(ctx); err != nil {
			return "", fmt.Errorf("rate limit wait: %w", err)
		}

		result, sc, err := c.doSendRequest(ctx, req)
		statusCode = sc

		if err == nil {
			return result, nil
		}

		lastErr = err

		if !isRetryableError(err, statusCode, c.retryConfig) {
			return "", err
		}

		c.logger.Warn("request failed, will retry", "attempt", attempt, "error", err, "status_code", statusCode)
	}

	return "", fmt.Errorf("max retry attempts reached: %w", lastErr)
}

func (c *OpenAIClient) doSendRequest(ctx context.Context, req OpenAIRequest) (string, int, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return "", 0, fmt.Errorf("marshal request: %w", err)
	}

	c.logger.Debug("openai compat request", "model", req.Model, "messages", len(req.Messages), "base_url", c.baseURL)

	url := c.baseURL + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", 0, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", 0, fmt.Errorf("send request: %w", err)
	}
	defer closeBody(resp)

	statusCode := resp.StatusCode
	if statusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", statusCode, fmt.Errorf("API error: %s: %s", resp.Status, string(body))
	}

	var result OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", statusCode, fmt.Errorf("decode response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", statusCode, nil
	}

	text := result.Choices[0].Message.Content
	c.logger.Debug("openai compat response", "id", result.ID, "tokens", result.Usage.CompletionTokens)

	return text, statusCode, nil
}

func (c *OpenAIClient) streamRequest(ctx context.Context, req OpenAIRequest, callback func(text string)) error {
	var lastErr error
	var statusCode int

	for attempt := 1; attempt <= c.retryConfig.MaxAttempts; attempt++ {
		if attempt > 1 {
			backoff := calculateBackoff(attempt, c.retryConfig)
			c.logger.Info("retrying stream request", "attempt", attempt, "max_attempts", c.retryConfig.MaxAttempts, "delay", backoff)
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		if err := c.rateLimiter.Wait(ctx); err != nil {
			return fmt.Errorf("rate limit wait: %w", err)
		}

		sc, err := c.doStreamRequest(ctx, req, callback)
		statusCode = sc

		if err == nil {
			return nil
		}

		lastErr = err

		if !isRetryableError(err, statusCode, c.retryConfig) {
			return err
		}

		c.logger.Warn("stream request failed, will retry", "attempt", attempt, "error", err, "status_code", statusCode)
	}

	return fmt.Errorf("max retry attempts reached: %w", lastErr)
}

func (c *OpenAIClient) doStreamRequest(ctx context.Context, req OpenAIRequest, callback func(text string)) (int, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return 0, fmt.Errorf("marshal request: %w", err)
	}

	c.logger.Debug("openai compat stream request", "model", req.Model, "messages", len(req.Messages), "base_url", c.baseURL)

	url := c.baseURL + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return 0, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return 0, fmt.Errorf("send request: %w", err)
	}
	defer closeBody(resp)

	statusCode := resp.StatusCode
	if statusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return statusCode, fmt.Errorf("API error: %s: %s", resp.Status, string(body))
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var chunk OpenAIStreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			c.logger.Warn("failed to parse stream chunk", "error", err, "data", data)
			continue
		}

		if len(chunk.Choices) > 0 {
			content := chunk.Choices[0].Delta.Content
			if content != "" {
				callback(content)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return statusCode, fmt.Errorf("read stream: %w", err)
	}

	return statusCode, nil
}

func (c *OpenAIClient) Validate(ctx context.Context) error {
	validateCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req := OpenAIRequest{
		Model:     string(ModelGLM3Turbo),
		MaxTokens: 10,
		Messages: []OpenAIMessage{
			{
				Role:    "user",
				Content: "Hi",
			},
		},
		Stream: false,
	}

	_, _, err := c.doSendRequest(validateCtx, req)
	if err != nil {
		return fmt.Errorf("openai compat validation failed: %w", err)
	}

	return nil
}
