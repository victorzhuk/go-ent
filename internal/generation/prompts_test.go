package generation

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadPromptTemplate(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T) string
		templateType string
		wantErr      bool
		wantBuiltIn  bool
	}{
		{
			name: "load built-in template when file not found",
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			templateType: "usecase",
			wantErr:      false,
			wantBuiltIn:  true,
		},
		{
			name: "load custom template from file",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				promptsDir := filepath.Join(dir, "prompts")
				if err := os.MkdirAll(promptsDir, 0755); err != nil {
					t.Fatal(err)
				}

				customPrompt := `# Custom Prompt
{{.SpecContent}}
`
				if err := os.WriteFile(filepath.Join(promptsDir, "usecase.md"), []byte(customPrompt), 0644); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			templateType: "usecase",
			wantErr:      false,
			wantBuiltIn:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectRoot := tt.setup(t)
			tmpl, err := LoadPromptTemplate(projectRoot, tt.templateType)

			if (err != nil) != tt.wantErr {
				t.Errorf("LoadPromptTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tmpl != nil {
				if tmpl.Name != tt.templateType {
					t.Errorf("template name = %q, want %q", tmpl.Name, tt.templateType)
				}

				if tt.wantBuiltIn && !strings.Contains(tmpl.Content, "Context") {
					t.Error("expected built-in template to contain 'Context'")
				}
			}
		})
	}
}

func TestPromptTemplateExecute(t *testing.T) {
	tests := []struct {
		name     string
		template string
		ctx      PromptContext
		wantErr  bool
		validate func(t *testing.T, result string)
	}{
		{
			name:     "basic substitution",
			template: "Project: {{.ProjectName}}\nSpec: {{.SpecContent}}",
			ctx: PromptContext{
				ProjectName: "test-project",
				SpecContent: "# Test Spec",
			},
			wantErr: false,
			validate: func(t *testing.T, result string) {
				if !strings.Contains(result, "test-project") {
					t.Error("expected result to contain project name")
				}
				if !strings.Contains(result, "# Test Spec") {
					t.Error("expected result to contain spec content")
				}
			},
		},
		{
			name:     "requirements loop",
			template: "{{range .Requirements}}- {{.Name}}: {{.Description}}\n{{end}}",
			ctx: PromptContext{
				Requirements: []Requirement{
					{Name: "R1", Description: "First requirement"},
					{Name: "R2", Description: "Second requirement"},
				},
			},
			wantErr: false,
			validate: func(t *testing.T, result string) {
				if !strings.Contains(result, "R1") || !strings.Contains(result, "R2") {
					t.Error("expected result to contain both requirements")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl := &PromptTemplate{
				Name:    "test",
				Content: tt.template,
			}

			result, err := tmpl.Execute(tt.ctx)

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestGetBuiltInPrompt(t *testing.T) {
	types := []string{"usecase", "handler", "repository", "unknown"}

	for _, templateType := range types {
		t.Run(templateType, func(t *testing.T) {
			tmpl := getBuiltInPrompt(templateType)

			if tmpl == nil {
				t.Error("expected non-nil template")
				return
			}

			if tmpl.Name != templateType {
				t.Errorf("template name = %q, want %q", tmpl.Name, templateType)
			}

			if len(tmpl.Content) == 0 {
				t.Error("expected non-empty template content")
			}
		})
	}
}
