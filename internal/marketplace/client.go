package marketplace

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	// DefaultBaseURL is the default marketplace API base URL.
	DefaultBaseURL = "https://marketplace.go-ent.dev/api/v1"
	// Timeout is the default HTTP client timeout.
	Timeout = 30 * time.Second
	// MaxDownloadSize is the maximum allowed download size (100MB).
	MaxDownloadSize = 100 * 1024 * 1024
	// MaxErrorBodySize is the maximum error response body to read (1MB).
	MaxErrorBodySize = 1024 * 1024
)

// Client handles marketplace API interactions.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new marketplace client with default base URL.
func NewClient() *Client {
	return &Client{
		baseURL: DefaultBaseURL,
		httpClient: &http.Client{
			Timeout: Timeout,
		},
	}
}

// NewClientWithURL creates a new marketplace client with custom base URL.
func NewClientWithURL(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: Timeout,
		},
	}
}

// Search searches for plugins in the marketplace.
func (c *Client) Search(ctx context.Context, query string, opts SearchOptions) ([]PluginInfo, error) {
	reqURL := fmt.Sprintf("%s/plugins/search?q=%s", c.baseURL, url.QueryEscape(query))

	if opts.Category != "" {
		reqURL += fmt.Sprintf("&category=%s", url.QueryEscape(opts.Category))
	}

	if opts.Author != "" {
		reqURL += fmt.Sprintf("&author=%s", url.QueryEscape(opts.Author))
	}

	if opts.SortBy != "" {
		reqURL += fmt.Sprintf("&sort_by=%s", url.QueryEscape(opts.SortBy))
	}

	if opts.Limit > 0 {
		reqURL += fmt.Sprintf("&limit=%d", opts.Limit)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, MaxErrorBodySize))
		return nil, fmt.Errorf("search failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	var response SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return response.Plugins, nil
}

// Download downloads a plugin from the marketplace.
func (c *Client) Download(ctx context.Context, name, version string) ([]byte, error) {
	url := fmt.Sprintf("%s/plugins/%s/versions/%s/download", c.baseURL, name, version)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, MaxErrorBodySize))
		return nil, fmt.Errorf("download failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, MaxDownloadSize))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if len(data) >= MaxDownloadSize {
		return nil, fmt.Errorf("download size exceeds maximum allowed size of %d bytes", MaxDownloadSize)
	}

	return data, nil
}

// GetPlugin retrieves plugin details from marketplace.
func (c *Client) GetPlugin(ctx context.Context, name string) (*PluginInfo, error) {
	url := fmt.Sprintf("%s/plugins/%s", c.baseURL, name)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, MaxErrorBodySize))
		return nil, fmt.Errorf("get plugin failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	var plugin PluginInfo
	if err := json.NewDecoder(resp.Body).Decode(&plugin); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &plugin, nil
}

// PluginInfo represents plugin metadata from marketplace.
type PluginInfo struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	Category    string   `json:"category"`
	Downloads   int      `json:"downloads"`
	Rating      float64  `json:"rating"`
	Tags        []string `json:"tags"`
	Skills      int      `json:"skills_count"`
	Agents      int      `json:"agents_count"`
	Rules       int      `json:"rules_count"`
}

// SearchOptions defines search parameters for marketplace queries.
type SearchOptions struct {
	Category string
	Author   string
	SortBy   string
	Limit    int
}

// SearchResponse represents marketplace search results.
type SearchResponse struct {
	Plugins []PluginInfo `json:"plugins"`
	Total   int          `json:"total"`
}
