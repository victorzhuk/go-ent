package model

import "strings"

type Resolver struct {
	cfg     *Config
	runtime string
}

func NewResolver(cfg *Config, runtime string) *Resolver {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Resolver{
		cfg:     cfg,
		runtime: runtime,
	}
}

// Resolve returns the actual model ID for a category
func (r *Resolver) Resolve(cat Category) string {
	mapping, ok := r.cfg.Runtimes[r.runtime]
	if !ok {
		return string(cat)
	}
	return mapping.Get(cat)
}

// ResolveAgent resolves model for an agent, handling legacy names
func (r *Resolver) ResolveAgent(agentModel string) string {
	if agentModel == "" {
		return r.Resolve(Main)
	}

	agentModel = strings.ToLower(agentModel)

	// Check aliases first
	if alias, ok := r.cfg.Aliases[agentModel]; ok {
		agentModel = alias
	}

	// If it's a valid category, resolve it
	if IsValid(agentModel) {
		return r.Resolve(Category(agentModel))
	}

	// Already a full model ID (contains /) or unknown, return as-is
	return agentModel
}
