package tools

import "sync"

// ToolRegistry tracks registered tools
type ToolRegistry struct {
	mu    sync.RWMutex
	tools []ToolSummary
}

// NewToolRegistry creates a new registry
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make([]ToolSummary, 0),
	}
}

// Register adds a tool to the registry
func (r *ToolRegistry) Register(name, description, category string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools = append(r.tools, ToolSummary{
		Name:        name,
		Description: description,
		Category:    category,
	})
}

// All returns all registered tools
func (r *ToolRegistry) All() []ToolSummary {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]ToolSummary, len(r.tools))
	copy(result, r.tools)
	return result
}
