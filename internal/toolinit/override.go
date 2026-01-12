package toolinit

// ModelResolver resolves model names based on agent tags and override patterns
type ModelResolver struct {
	Overrides map[string]string // pattern -> model
}

// NewModelResolver creates a new ModelResolver with the given overrides
func NewModelResolver(overrides map[string]string) *ModelResolver {
	return &ModelResolver{
		Overrides: overrides,
	}
}

// Resolve returns the model for an agent based on tags and overrides
// Priority order (most specific wins):
// 1. role:complexity (e.g., planning:heavy)
// 2. complexity alone (e.g., heavy)
// 3. role alone (e.g., planning)
// 4. default from meta.Model
func (r *ModelResolver) Resolve(meta *AgentMeta) string {
	if r.Overrides == nil || len(r.Overrides) == 0 {
		return meta.Model
	}

	// Priority 1: Specific combination (role:complexity)
	if meta.Tags.Role != "" && meta.Tags.Complexity != "" {
		specificKey := meta.Tags.Role + ":" + meta.Tags.Complexity
		if model, ok := r.Overrides[specificKey]; ok {
			return model
		}
	}

	// Priority 2: Complexity alone
	if meta.Tags.Complexity != "" {
		if model, ok := r.Overrides[meta.Tags.Complexity]; ok {
			return model
		}
	}

	// Priority 3: Role alone
	if meta.Tags.Role != "" {
		if model, ok := r.Overrides[meta.Tags.Role]; ok {
			return model
		}
	}

	// Priority 4: Default from agent metadata
	return meta.Model
}
