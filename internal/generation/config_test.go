package generation

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name        string
		setupConfig func(t *testing.T) string
		wantErr     bool
		validate    func(t *testing.T, cfg *GenerationConfig)
	}{
		{
			name: "no config file returns defaults",
			setupConfig: func(t *testing.T) string {
				dir := t.TempDir()
				return dir
			},
			wantErr: false,
			validate: func(t *testing.T, cfg *GenerationConfig) {
				if cfg.Defaults.GoVersion != "1.25" {
					t.Errorf("want GoVersion=1.25, got %s", cfg.Defaults.GoVersion)
				}
				if cfg.Defaults.Archetype != "standard" {
					t.Errorf("want Archetype=standard, got %s", cfg.Defaults.Archetype)
				}
			},
		},
		{
			name: "valid config loaded",
			setupConfig: func(t *testing.T) string {
				dir := t.TempDir()
				openspecDir := filepath.Join(dir, "openspec")
				if err := os.MkdirAll(openspecDir, 0755); err != nil {
					t.Fatal(err)
				}

				configYAML := `defaults:
  go_version: "1.24"
  archetype: mcp
archetypes:
  custom:
    description: Custom archetype
    templates:
      - foo
      - bar
components:
  - name: test-component
    spec: specs/test.md
    archetype: custom
    output: internal/test/
`
				if err := os.WriteFile(filepath.Join(openspecDir, "generation.yaml"), []byte(configYAML), 0644); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			wantErr: false,
			validate: func(t *testing.T, cfg *GenerationConfig) {
				if cfg.Defaults.GoVersion != "1.24" {
					t.Errorf("want GoVersion=1.24, got %s", cfg.Defaults.GoVersion)
				}
				if cfg.Defaults.Archetype != "mcp" {
					t.Errorf("want Archetype=mcp, got %s", cfg.Defaults.Archetype)
				}
				if len(cfg.Archetypes) != 1 {
					t.Errorf("want 1 custom archetype, got %d", len(cfg.Archetypes))
				}
				if len(cfg.Components) != 1 {
					t.Errorf("want 1 component, got %d", len(cfg.Components))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectRoot := tt.setupConfig(t)
			cfg, err := LoadConfig(projectRoot)

			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && cfg != nil && tt.validate != nil {
				tt.validate(t, cfg)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := defaultConfig()

	if cfg.Defaults.GoVersion != "1.25" {
		t.Errorf("want GoVersion=1.25, got %s", cfg.Defaults.GoVersion)
	}
	if cfg.Defaults.Archetype != "standard" {
		t.Errorf("want Archetype=standard, got %s", cfg.Defaults.Archetype)
	}
}
