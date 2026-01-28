package config

import "strings"

type ModelResolver struct {
	cfg     *ModelConfig
	runtime string
}

func NewModelResolver(cfg *ModelConfig, runtime string) *ModelResolver {
	if cfg == nil {
		cfg = DefaultModelConfig()
	}
	return &ModelResolver{
		cfg:     cfg,
		runtime: runtime,
	}
}

// Resolve returns the actual model ID for a category
func (r *ModelResolver) Resolve(cat ModelCategory) string {
	mapping, ok := r.cfg.Runtimes[r.runtime]
	if !ok {
		return string(cat)
	}
	return mapping.Get(cat)
}

// ResolveAgent resolves model for an agent, handling legacy names
func (r *ModelResolver) ResolveAgent(agentModel string) string {
	if agentModel == "" {
		return r.Resolve(ModelMain)
	}

	agentModel = strings.ToLower(agentModel)

	// Check aliases first
	if alias, ok := r.cfg.Aliases[agentModel]; ok {
		agentModel = alias
	}

	// If it's a valid category, resolve it
	if IsValidModelCategory(agentModel) {
		return r.Resolve(ModelCategory(agentModel))
	}

	// Already a full model ID (contains /) or unknown, return as-is
	return agentModel
}
