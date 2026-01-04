package generation

import (
	"fmt"
)

// ArchetypeMetadata contains metadata about an archetype.
type ArchetypeMetadata struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Templates   []string `json:"templates"`
	BuiltIn     bool     `json:"built_in"`
}

// builtInArchetypes defines the built-in project archetypes.
var builtInArchetypes = map[string]*Archetype{
	"standard": {
		Description: "Web service with clean architecture",
		Templates: []string{
			"go.mod",
			"Makefile",
			"CLAUDE.md",
			"build/Dockerfile",
			"deploy/docker-compose.yml",
			"cmd/server/main.go",
			"internal/app/app.go",
			"internal/config/config.go",
		},
	},
	"mcp": {
		Description: "MCP server plugin",
		Templates: []string{
			"mcp/go.mod",
			"mcp/Makefile",
			"mcp/build/Dockerfile",
			"mcp/cmd/server/main.go",
			"mcp/internal/server/server.go",
		},
	},
	"api": {
		Description: "API-only service (no web UI)",
		Templates: []string{
			"go.mod",
			"Makefile",
			"CLAUDE.md",
			"build/Dockerfile",
			"cmd/server/main.go",
			"internal/app/app.go",
			"internal/config/config.go",
		},
	},
	"grpc": {
		Description: "gRPC service",
		Templates: []string{
			"go.mod",
			"Makefile",
			"CLAUDE.md",
			"build/Dockerfile",
			"cmd/server/main.go",
			"internal/app/app.go",
			"internal/config/config.go",
		},
	},
	"worker": {
		Description: "Background worker service",
		Templates: []string{
			"go.mod",
			"Makefile",
			"CLAUDE.md",
			"build/Dockerfile",
			"cmd/server/main.go",
			"internal/app/app.go",
			"internal/config/config.go",
		},
	},
}

// GetArchetype retrieves an archetype by name, checking custom archetypes first,
// then built-in archetypes.
func GetArchetype(name string, cfg *GenerationConfig) (*Archetype, error) {
	// Check custom archetypes first
	if cfg != nil && cfg.Archetypes != nil {
		if arch, ok := cfg.Archetypes[name]; ok {
			return arch, nil
		}
	}

	// Check built-in archetypes
	if arch, ok := builtInArchetypes[name]; ok {
		return arch, nil
	}

	return nil, fmt.Errorf("archetype not found: %s", name)
}

// ListArchetypes returns all available archetypes (built-in + custom).
func ListArchetypes(cfg *GenerationConfig) []ArchetypeMetadata {
	var result []ArchetypeMetadata

	// Add built-in archetypes
	for name, arch := range builtInArchetypes {
		result = append(result, ArchetypeMetadata{
			Name:        name,
			Description: arch.Description,
			Templates:   arch.Templates,
			BuiltIn:     true,
		})
	}

	// Add custom archetypes
	if cfg != nil && cfg.Archetypes != nil {
		for name, arch := range cfg.Archetypes {
			// Skip if it overrides a built-in (already added)
			if _, isBuiltIn := builtInArchetypes[name]; isBuiltIn {
				continue
			}
			result = append(result, ArchetypeMetadata{
				Name:        name,
				Description: arch.Description,
				Templates:   arch.Templates,
				BuiltIn:     false,
			})
		}
	}

	return result
}

// ResolveTemplateList returns the final template list for an archetype,
// applying skip filters if specified.
func ResolveTemplateList(archetype *Archetype) []string {
	if len(archetype.Skip) == 0 {
		return archetype.Templates
	}

	// Build skip map for O(1) lookup
	skip := make(map[string]bool)
	for _, s := range archetype.Skip {
		skip[s] = true
	}

	// Filter templates
	var result []string
	for _, tmpl := range archetype.Templates {
		if !skip[tmpl] {
			result = append(result, tmpl)
		}
	}

	return result
}
