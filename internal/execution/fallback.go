package execution

import "github.com/victorzhuk/go-ent/internal/domain"

// RuntimeFamily groups runtimes that can substitute for each other.
type RuntimeFamily string

const (
	// FamilyMCP represents MCP-compatible runtimes (claude-code, open-code).
	FamilyMCP RuntimeFamily = "mcp"

	// FamilyCLI represents standalone CLI runtime (isolated, no fallback).
	FamilyCLI RuntimeFamily = "cli"
)

// FallbackResolver determines fallback runtimes for failed executions.
type FallbackResolver struct{}

// NewFallbackResolver creates a new fallback resolver.
func NewFallbackResolver() *FallbackResolver {
	return &FallbackResolver{}
}

// GetFamily returns the runtime family for the given runtime.
func (f *FallbackResolver) GetFamily(rt domain.Runtime) RuntimeFamily {
	switch rt {
	case domain.RuntimeClaudeCode, domain.RuntimeOpenCode:
		return FamilyMCP
	case domain.RuntimeCLI:
		return FamilyCLI
	default:
		return FamilyCLI
	}
}

// GetFallbacks returns fallback runtimes for the given runtime.
// Same-family fallback: MCP runtimes can substitute each other.
// Cross-family fallback: CLI is isolated, no fallback.
func (f *FallbackResolver) GetFallbacks(rt domain.Runtime) []domain.Runtime {
	family := f.GetFamily(rt)

	switch family {
	case FamilyMCP:
		// MCP family: claude-code <-> open-code
		if rt == domain.RuntimeClaudeCode {
			return []domain.Runtime{domain.RuntimeOpenCode}
		}
		return []domain.Runtime{domain.RuntimeClaudeCode}

	case FamilyCLI:
		// CLI stays CLI - no cross-family fallback
		return []domain.Runtime{}
	}

	return nil
}

// CanFallback returns true if the runtime has fallback options.
func (f *FallbackResolver) CanFallback(rt domain.Runtime) bool {
	return len(f.GetFallbacks(rt)) > 0
}

// SameFamily returns true if both runtimes are in the same family.
func (f *FallbackResolver) SameFamily(rt1, rt2 domain.Runtime) bool {
	return f.GetFamily(rt1) == f.GetFamily(rt2)
}
