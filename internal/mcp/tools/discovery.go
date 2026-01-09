package tools

import (
	"fmt"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ToolMeta holds metadata about a tool without its implementation.
type ToolMeta struct {
	Name        string
	Description string
	InputSchema map[string]any
	Category    string
	Keywords    []string
}

// RegistrationFunc is a function that registers a tool with the MCP server.
type RegistrationFunc func(*mcp.Server)

// ToolRegistry manages tool discovery and lazy loading.
type ToolRegistry struct {
	mu           sync.RWMutex
	metadata     map[string]*ToolMeta
	registrators map[string]RegistrationFunc
	active       map[string]bool
	index        *SearchIndex
	server       *mcp.Server
}

// NewToolRegistry creates a new tool registry.
func NewToolRegistry(server *mcp.Server) *ToolRegistry {
	return &ToolRegistry{
		metadata:     make(map[string]*ToolMeta),
		registrators: make(map[string]RegistrationFunc),
		active:       make(map[string]bool),
		index:        NewSearchIndex(),
		server:       server,
	}
}

// Register adds a tool's metadata and registration function to the registry.
// The tool is not activated until Load() is called.
func (r *ToolRegistry) Register(meta ToolMeta, registrator RegistrationFunc) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.metadata[meta.Name]; exists {
		return fmt.Errorf("tool already registered: %s", meta.Name)
	}

	r.metadata[meta.Name] = &meta
	r.registrators[meta.Name] = registrator

	return nil
}

// Find searches for tools matching the query using TF-IDF.
func (r *ToolRegistry) Find(query string, limit int) []*ToolMeta {
	r.mu.RLock()
	defer r.mu.RUnlock()

	results := r.index.Search(query, limit)
	tools := make([]*ToolMeta, 0, len(results))

	for _, result := range results {
		if meta, ok := r.metadata[result.ToolName]; ok {
			tools = append(tools, meta)
		}
	}

	return tools
}

// Describe returns metadata for a specific tool.
func (r *ToolRegistry) Describe(name string) (*ToolMeta, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	meta, ok := r.metadata[name]
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", name)
	}

	return meta, nil
}

// Load activates tools by registering them with the MCP server.
func (r *ToolRegistry) Load(names []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, name := range names {
		if r.active[name] {
			continue // Already loaded
		}

		_, ok := r.metadata[name]
		if !ok {
			return fmt.Errorf("tool not found: %s", name)
		}

		registrator, ok := r.registrators[name]
		if !ok {
			return fmt.Errorf("registrator not found: %s", name)
		}

		// Call the registration function
		registrator(r.server)

		// Mark as active
		r.active[name] = true
	}

	return nil
}

// Active returns a list of currently active tool names.
func (r *ToolRegistry) Active() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	active := make([]string, 0, len(r.active))
	for name, isActive := range r.active {
		if isActive {
			active = append(active, name)
		}
	}

	return active
}

// All returns metadata for all registered tools.
func (r *ToolRegistry) All() []*ToolMeta {
	r.mu.RLock()
	defer r.mu.RUnlock()

	all := make([]*ToolMeta, 0, len(r.metadata))
	for _, meta := range r.metadata {
		all = append(all, meta)
	}

	return all
}

// BuildIndex builds the TF-IDF search index from registered tools.
func (r *ToolRegistry) BuildIndex() error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	docs := make([]Document, 0, len(r.metadata))
	id := 0

	for name, meta := range r.metadata {
		doc := BuildDocument(id, name, meta.Description)
		docs = append(docs, doc)
		id++
	}

	return r.index.Index(docs)
}

// IsActive checks if a tool is currently active.
func (r *ToolRegistry) IsActive(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.active[name]
}

// Count returns the number of registered tools.
func (r *ToolRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.metadata)
}
