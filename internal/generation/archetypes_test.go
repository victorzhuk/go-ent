package generation

//nolint:gosec // test file with necessary file operations

import (
	"testing"
)

func TestGetArchetype(t *testing.T) {
	customArchetype := &Archetype{
		Description: "Custom test archetype",
		Templates:   []string{"custom1", "custom2"},
	}

	cfg := &GenerationConfig{
		Archetypes: map[string]*Archetype{
			"custom": customArchetype,
		},
	}

	tests := []struct {
		name     string
		archName string
		cfg      *GenerationConfig
		wantErr  bool
		validate func(t *testing.T, arch *Archetype)
	}{
		{
			name:     "built-in archetype standard",
			archName: "standard",
			cfg:      nil,
			wantErr:  false,
			validate: func(t *testing.T, arch *Archetype) {
				if arch.Description != "Web service with clean architecture" {
					t.Errorf("unexpected description: %s", arch.Description)
				}
				if len(arch.Templates) == 0 {
					t.Error("expected templates, got none")
				}
			},
		},
		{
			name:     "built-in archetype mcp",
			archName: "mcp",
			cfg:      nil,
			wantErr:  false,
			validate: func(t *testing.T, arch *Archetype) {
				if arch.Description != "MCP server plugin" {
					t.Errorf("unexpected description: %s", arch.Description)
				}
			},
		},
		{
			name:     "custom archetype",
			archName: "custom",
			cfg:      cfg,
			wantErr:  false,
			validate: func(t *testing.T, arch *Archetype) {
				if arch.Description != "Custom test archetype" {
					t.Errorf("unexpected description: %s", arch.Description)
				}
				if len(arch.Templates) != 2 {
					t.Errorf("want 2 templates, got %d", len(arch.Templates))
				}
			},
		},
		{
			name:     "non-existent archetype",
			archName: "nonexistent",
			cfg:      nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arch, err := GetArchetype(tt.archName, tt.cfg)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetArchetype() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && arch != nil && tt.validate != nil {
				tt.validate(t, arch)
			}
		})
	}
}

func TestListArchetypes(t *testing.T) {
	cfg := &GenerationConfig{
		Archetypes: map[string]*Archetype{
			"custom1": {
				Description: "Custom 1",
				Templates:   []string{"tmpl1"},
			},
			"custom2": {
				Description: "Custom 2",
				Templates:   []string{"tmpl2"},
			},
		},
	}

	archetypes := ListArchetypes(cfg)

	// Should have all built-in archetypes + custom ones
	builtInCount := len(builtInArchetypes)
	customCount := 2

	if len(archetypes) != builtInCount+customCount {
		t.Errorf("want %d archetypes, got %d", builtInCount+customCount, len(archetypes))
	}

	// Check that built-in archetypes are marked correctly
	foundBuiltIn := false
	foundCustom := false
	for _, arch := range archetypes {
		if arch.Name == "standard" && arch.BuiltIn {
			foundBuiltIn = true
		}
		if arch.Name == "custom1" && !arch.BuiltIn {
			foundCustom = true
		}
	}

	if !foundBuiltIn {
		t.Error("expected to find built-in archetype 'standard'")
	}
	if !foundCustom {
		t.Error("expected to find custom archetype 'custom1'")
	}
}

func TestResolveTemplateList(t *testing.T) {
	tests := []struct {
		name      string
		archetype *Archetype
		wantCount int
	}{
		{
			name: "no skip filter",
			archetype: &Archetype{
				Templates: []string{"a", "b", "c"},
			},
			wantCount: 3,
		},
		{
			name: "with skip filter",
			archetype: &Archetype{
				Templates: []string{"a", "b", "c", "d"},
				Skip:      []string{"b", "d"},
			},
			wantCount: 2,
		},
		{
			name: "skip all",
			archetype: &Archetype{
				Templates: []string{"a", "b"},
				Skip:      []string{"a", "b"},
			},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ResolveTemplateList(tt.archetype)
			if len(result) != tt.wantCount {
				t.Errorf("want %d templates, got %d", tt.wantCount, len(result))
			}
		})
	}
}
