package openspec

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
)

// Client wraps the openspec CLI for programmatic access.
type Client struct {
	cwd string // Working directory for openspec commands
}

// New creates a new OpenSpec CLI client.
func New(cwd string) *Client {
	return &Client{cwd: cwd}
}

// Archive archives a completed change and updates main specs.
func (c *Client) Archive(ctx context.Context, change string) error {
	_, err := c.run(ctx, "archive", change)
	return err
}

// Validate validates changes and specs and returns JSON output.
// Pass "all", "changes", or "specs" as scope, or empty string for default behavior.
func (c *Client) Validate(ctx context.Context, scope string) ([]byte, error) {
	args := []string{"validate", "--json"}
	if scope != "" {
		args = append(args, "--"+scope)
	}
	return c.run(ctx, args...)
}

// Instructions outputs enriched instructions for creating an artifact or applying tasks.
// change is the change name, artifact is optional (e.g., "proposal", "tasks").
func (c *Client) Instructions(ctx context.Context, change, artifact string) (string, error) {
	args := []string{"instructions"}
	if change != "" {
		args = append(args, change)
	}
	if artifact != "" {
		args = append(args, artifact)
	}
	out, err := c.run(ctx, args...)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// NewChange creates a new change proposal.
func (c *Client) NewChange(ctx context.Context, name string) error {
	_, err := c.run(ctx, "new", "change", name)
	return err
}

// run executes the openspec CLI command and returns stdout.
func (c *Client) run(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "openspec", args...)
	cmd.Dir = c.cwd

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("openspec %v: %w (stderr: %s)", args, err, stderr.String())
	}

	return stdout.Bytes(), nil
}

// ParseList parses the JSON output from List into a slice of ListItems.
func ParseList(data []byte) ([]ListItem, error) {
	var resp ListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parse list: %w", err)
	}
	// Return changes or specs depending on what was in the response
	if len(resp.Changes) > 0 {
		return resp.Changes, nil
	}
	return resp.Specs, nil
}

// ParseValidate parses the JSON output from Validate.
func ParseValidate(data []byte) (map[string]any, error) {
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parse validate: %w", err)
	}
	return result, nil
}
