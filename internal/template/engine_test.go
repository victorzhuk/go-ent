package template

//nolint:gosec // test file with necessary file operations

import (
	"embed"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/*.tmpl
var testFS embed.FS

func TestNewEngine(t *testing.T) {
	t.Parallel()

	engine := NewEngine(testFS)
	assert.NotNil(t, engine)
}

func TestProcess(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		templatePath string
		vars         TemplateVars
		wantErr      bool
		wantContent  string
	}{
		{
			name:         "basic substitution",
			templatePath: "testdata/basic.tmpl",
			vars: TemplateVars{
				ModulePath:  "github.com/user/project",
				ProjectName: "project",
				GoVersion:   "1.24",
			},
			wantErr: false,
			wantContent: `module github.com/user/project

go 1.24

Project: project
`,
		},
		{
			name:         "missing template file",
			templatePath: "testdata/nonexistent.tmpl",
			vars:         TemplateVars{},
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			outputPath := filepath.Join(tmpDir, "output.txt")

			engine := NewEngine(testFS)
			err := engine.Process(tt.templatePath, tt.vars, outputPath)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			got, err := os.ReadFile(outputPath) // #nosec G304 -- test file
			require.NoError(t, err)
			assert.Equal(t, tt.wantContent, string(got))
		})
	}
}

func TestProcessAll(t *testing.T) {
	t.Parallel()

	t.Run("process all templates in directory", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		engine := NewEngine(testFS)

		vars := TemplateVars{
			ModulePath:  "github.com/test/app",
			ProjectName: "app",
			GoVersion:   "1.24",
		}

		// List available templates first
		templates, err := engine.ListTemplates("testdata")
		require.NoError(t, err, "should be able to list templates")
		require.NotEmpty(t, templates, "should have at least one template")

		err = engine.ProcessAll("testdata", vars, tmpDir)
		require.NoError(t, err)

		// Verify basic file was created (tmpl extension is stripped)
		outputPath := filepath.Join(tmpDir, "basic")
		got, err := os.ReadFile(outputPath) // #nosec G304 -- test file
		require.NoError(t, err)

		// Verify substitution worked
		assert.Contains(t, string(got), "github.com/test/app")
		assert.Contains(t, string(got), "app")
		assert.Contains(t, string(got), "1.24")
	})

	t.Run("nonexistent directory", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		engine := NewEngine(testFS)
		err := engine.ProcessAll("nonexistent", TemplateVars{}, tmpDir)
		assert.Error(t, err)
	})
}

func TestListTemplates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		templateDir   string
		wantTemplates []string
		wantErr       bool
	}{
		{
			name:          "list templates in directory",
			templateDir:   "testdata",
			wantTemplates: []string{"testdata/basic.tmpl"},
			wantErr:       false,
		},
		{
			name:        "nonexistent directory",
			templateDir: "nonexistent",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			engine := NewEngine(testFS)
			templates, err := engine.ListTemplates(tt.templateDir)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.ElementsMatch(t, tt.wantTemplates, templates)
		})
	}
}
