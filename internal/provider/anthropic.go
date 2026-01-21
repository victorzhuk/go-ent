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

const (
	anthropicAPIURL  = "https://api.anthropic.com/v1/messages"
	anthropicVersion = "2023-06-01"
)

func closeBody(resp *http.Response) {
	if resp != nil {
		_ = resp.Body.Close()
	}
}

type Model string

const (
	ModelHaiku  Model = "claude-3-haiku-20240307"
	ModelSonnet Model = "claude-3-5-sonnet-20241022"
	ModelOpus   Model = "claude-3-opus-20240229"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Client struct {
	apiKey      string
	httpClient  *http.Client
	logger      *slog.Logger
	rateLimiter *RateLimiter
	retryConfig RetryConfig
}

type StreamEvent struct {
	Type  string `json:"type"`
	Delta *struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"delta,omitempty"`
	Message *struct {
		ID      string `json:"id"`
		Type    string `json:"type"`
		Role    string `json:"role"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Model      string `json:"model"`
		StopReason string `json:"stop_reason"`
	} `json:"message,omitempty"`
}

type Request struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []Message `json:"messages"`
	Stream    bool      `json:"stream"`
}

func NewAnthropicClient(logger *slog.Logger) (*Client, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY environment variable not set")
	}

	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
		logger:      logger,
		rateLimiter: NewRateLimiter(50, logger),
		retryConfig: DefaultRetryConfig(),
	}, nil
}

func (c *Client) Complete(ctx context.Context, model Model, prompt string) (string, error) {
	req := Request{
		Model:     string(model),
		MaxTokens: 4096,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Stream: false,
	}

	return c.sendRequest(ctx, req)
}

func (c *Client) CompleteWithHistory(ctx context.Context, model Model, messages []Message) (string, error) {
	req := Request{
		Model:     string(model),
		MaxTokens: 4096,
		Messages:  messages,
		Stream:    false,
	}

	return c.sendRequest(ctx, req)
}

func (c *Client) Stream(ctx context.Context, model Model, prompt string, callback func(text string)) error {
	req := Request{
		Model:     string(model),
		MaxTokens: 4096,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Stream: true,
	}

	return c.streamRequest(ctx, req, callback)
}

func (c *Client) StreamWithHistory(ctx context.Context, model Model, messages []Message, callback func(text string)) error {
	req := Request{
		Model:     string(model),
		MaxTokens: 4096,
		Messages:  messages,
		Stream:    true,
	}

	return c.streamRequest(ctx, req, callback)
}

func (c *Client) sendRequest(ctx context.Context, req Request) (string, error) {
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

func (c *Client) doSendRequest(ctx context.Context, req Request) (string, int, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return "", 0, fmt.Errorf("marshal request: %w", err)
	}

	c.logger.Debug("anthropic request", "model", req.Model, "messages", len(req.Messages))

	httpReq, err := http.NewRequestWithContext(ctx, "POST", anthropicAPIURL, bytes.NewReader(body))
	if err != nil {
		return "", 0, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", anthropicVersion)

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

	var result struct {
		ID      string `json:"id"`
		Type    string `json:"type"`
		Role    string `json:"role"`
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Model        string `json:"model"`
		StopReason   string `json:"stop_reason"`
		StopSequence string `json:"stop_sequence"`
		Usage        struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", statusCode, fmt.Errorf("decode response: %w", err)
	}

	if len(result.Content) == 0 {
		return "", statusCode, nil
	}

	text := result.Content[0].Text
	c.logger.Debug("anthropic response", "id", result.ID, "tokens", result.Usage.OutputTokens)

	return text, statusCode, nil
}

func (c *Client) streamRequest(ctx context.Context, req Request, callback func(text string)) error {
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

func (c *Client) doStreamRequest(ctx context.Context, req Request, callback func(text string)) (int, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return 0, fmt.Errorf("marshal request: %w", err)
	}

	c.logger.Debug("anthropic stream request", "model", req.Model, "messages", len(req.Messages))

	httpReq, err := http.NewRequestWithContext(ctx, "POST", anthropicAPIURL, bytes.NewReader(body))
	if err != nil {
		return 0, fmt.Errorf("create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", anthropicVersion)

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

		var event StreamEvent
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			c.logger.Warn("failed to parse stream event", "error", err, "data", data)
			continue
		}

		if event.Type == "content_block_delta" && event.Delta != nil && event.Delta.Text != "" {
			callback(event.Delta.Text)
		}
	}

	if err := scanner.Err(); err != nil {
		return statusCode, fmt.Errorf("read stream: %w", err)
	}

	return statusCode, nil
}

func (c *Client) Validate(ctx context.Context) error {
	validateCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req := Request{
		Model:     string(ModelHaiku),
		MaxTokens: 10,
		Messages: []Message{
			{
				Role:    "user",
				Content: "Hi",
			},
		},
		Stream: false,
	}

	_, _, err := c.doSendRequest(validateCtx, req)
	if err != nil {
		return fmt.Errorf("anthropic validation failed: %w", err)
	}

	return nil
}
