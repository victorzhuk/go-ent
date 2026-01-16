package tool

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ComposedTool represents a dynamically created tool.
type ComposedTool struct {
	// ID is the unique tool identifier.
	ID string `json:"id"`

	// Name is the tool name.
	Name string `json:"name"`

	// Description explains what the tool does.
	Description string `json:"description"`

	// Code is the JavaScript implementation.
	Code string `json:"code"`

	// Scope defines where the tool is available ("project" or "global").
	Scope string `json:"scope"`

	// Created is when the tool was created.
	Created time.Time `json:"created"`

	// UsageCount tracks how many times the tool has been used.
	UsageCount int `json:"usage_count"`

	// LastUsed tracks the last usage time.
	LastUsed time.Time `json:"last_used"`

	// Metadata holds additional tool metadata.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Composer manages composed tool persistence and retrieval.
type Composer struct {
	mu       sync.RWMutex
	rootPath string
	cache    map[string]*ComposedTool
}

// NewComposer creates a new tool composer.
func NewComposer(rootPath string) *Composer {
	return &Composer{
		rootPath: rootPath,
		cache:    make(map[string]*ComposedTool),
	}
}

// Save persists a composed tool to disk.
func (c *Composer) Save(tool *ComposedTool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Ensure .go-ent/composed-tools directory exists
	dir := c.composedToolsDir()
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("create composed-tools dir: %w", err)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(tool, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal tool: %w", err)
	}

	// Write to file
	path := c.toolPath(tool.Name)
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write tool file: %w", err)
	}

	// Update cache
	c.cache[tool.Name] = tool

	return nil
}

// Load retrieves a composed tool by name.
func (c *Composer) Load(name string) (*ComposedTool, error) {
	c.mu.RLock()

	// Check cache first
	if tool, exists := c.cache[name]; exists {
		c.mu.RUnlock()
		return tool, nil
	}
	c.mu.RUnlock()

	// Load from disk
	c.mu.Lock()
	defer c.mu.Unlock()

	path := c.toolPath(name)
	data, err := os.ReadFile(path) // #nosec G304 -- controlled config/template file path
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("tool not found: %s", name)
		}
		return nil, fmt.Errorf("read tool file: %w", err)
	}

	var tool ComposedTool
	if err := json.Unmarshal(data, &tool); err != nil {
		return nil, fmt.Errorf("unmarshal tool: %w", err)
	}

	// Cache it
	c.cache[tool.Name] = &tool

	return &tool, nil
}

// List returns all composed tools.
func (c *Composer) List() ([]*ComposedTool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	dir := c.composedToolsDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*ComposedTool{}, nil
		}
		return nil, fmt.Errorf("read composed-tools dir: %w", err)
	}

	var tools []*ComposedTool
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		name := entry.Name()[:len(entry.Name())-5] // Remove .json
		tool, err := c.loadFromDisk(name)
		if err != nil {
			// Skip tools that fail to load
			continue
		}
		tools = append(tools, tool)
	}

	return tools, nil
}

// Delete removes a composed tool.
func (c *Composer) Delete(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	path := c.toolPath(name)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete tool file: %w", err)
	}

	// Remove from cache
	delete(c.cache, name)

	return nil
}

// IncrementUsage increments the usage count for a tool.
func (c *Composer) IncrementUsage(name string) error {
	tool, err := c.Load(name)
	if err != nil {
		return err
	}

	tool.UsageCount++
	tool.LastUsed = time.Now()

	return c.Save(tool)
}

// composedToolsDir returns the path to the composed-tools directory.
func (c *Composer) composedToolsDir() string {
	return filepath.Join(c.rootPath, ".go-ent", "composed-tools")
}

// toolPath returns the file path for a tool.
func (c *Composer) toolPath(name string) string {
	return filepath.Join(c.composedToolsDir(), name+".json")
}

// loadFromDisk loads a tool from disk without locking (internal use).
func (c *Composer) loadFromDisk(name string) (*ComposedTool, error) {
	path := c.toolPath(name)
	data, err := os.ReadFile(path) // #nosec G304 -- controlled config/template file path
	if err != nil {
		return nil, err
	}

	var tool ComposedTool
	if err := json.Unmarshal(data, &tool); err != nil {
		return nil, err
	}

	return &tool, nil
}

// Exists checks if a tool with the given name exists.
func (c *Composer) Exists(name string) bool {
	_, err := c.Load(name)
	return err == nil
}
