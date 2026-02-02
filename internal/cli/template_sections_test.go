package cli

import (
	"strings"
	"testing"
)

func TestLoadSectionTemplate(t *testing.T) {
	tests := []struct {
		name    string
		section string
		wantErr bool
	}{
		{
			name:    "load tooling section",
			section: "_tooling",
			wantErr: false,
		},
		{
			name:    "load workflow section",
			section: "_workflow",
			wantErr: false,
		},
		{
			name:    "load principles section",
			section: "_principles",
			wantErr: false,
		},
		{
			name:    "load handoff section",
			section: "_handoff",
			wantErr: false,
		},
		{
			name:    "nonexistent section",
			section: "_nonexistent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tpl, err := loadSectionTemplate(tt.section)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadSectionTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tpl == nil {
				t.Error("loadSectionTemplate() returned nil template")
			}
		})
	}
}

func TestRenderSectionTemplate(t *testing.T) {
	// Load tooling template
	tpl, err := loadSectionTemplate("_tooling")
	if err != nil {
		t.Fatalf("loadSectionTemplate() error = %v", err)
	}

	tests := []struct {
		name    string
		data    *AgentTemplateData
		want    []string // Strings that should be in the output
		wantErr bool
	}{
		{
			name: "render with disallowed tools",
			data: &AgentTemplateData{
				Name:               "test-agent",
				Role:               "execution",
				DisallowedTools:    []string{"replace_symbol_body", "insert_after_symbol"},
				HasDisallowedTools: true,
			},
			want: []string{
				"Optimal Tooling",
				"rg \"pattern\"",
				"CRITICAL: Tool Usage",
				"replace_symbol_body",
				"insert_after_symbol",
			},
			wantErr: false,
		},
		{
			name: "render without disallowed tools",
			data: &AgentTemplateData{
				Name:               "test-agent",
				Role:               "execution",
				DisallowedTools:    []string{},
				HasDisallowedTools: false,
			},
			want: []string{
				"Optimal Tooling",
				"rg \"pattern\"",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := renderSectionTemplate(tpl, tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("renderSectionTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, want := range tt.want {
				if !strings.Contains(got, want) {
					t.Errorf("renderSectionTemplate() output missing expected string: %q\nGot:\n%s", want, got)
				}
			}

			// If no disallowed tools, should not have CRITICAL section
			if !tt.data.HasDisallowedTools && strings.Contains(got, "CRITICAL: Tool Usage") {
				t.Error("renderSectionTemplate() output contains CRITICAL section when it shouldn't")
			}
		})
	}
}

func TestRenderWorkflowSection(t *testing.T) {
	tpl, err := loadSectionTemplate("_workflow")
	if err != nil {
		t.Fatalf("loadSectionTemplate() error = %v", err)
	}

	tests := []struct {
		name string
		data *AgentTemplateData
		want []string
	}{
		{
			name: "execution role workflow",
			data: &AgentTemplateData{
				Name: "coder",
				Role: "execution",
			},
			want: []string{
				"Context Gathering",
				"openspec/changes",
				"serena_find_symbol",
			},
		},
		{
			name: "planning role workflow",
			data: &AgentTemplateData{
				Name: "planner",
				Role: "planning",
			},
			want: []string{
				"Context Gathering",
				"Clean Architecture",
				"Understand requirements",
			},
		},
		{
			name: "validation role workflow",
			data: &AgentTemplateData{
				Name: "reviewer",
				Role: "validation",
			},
			want: []string{
				"Context Gathering",
				"git diff",
				"Check architecture",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := renderSectionTemplate(tpl, tt.data)
			if err != nil {
				t.Errorf("renderSectionTemplate() error = %v", err)
				return
			}

			for _, want := range tt.want {
				if !strings.Contains(got, want) {
					t.Errorf("renderSectionTemplate() output missing expected string: %q\nGot:\n%s", want, got)
				}
			}
		})
	}
}

func TestRenderPrinciplesSection(t *testing.T) {
	tpl, err := loadSectionTemplate("_principles")
	if err != nil {
		t.Fatalf("loadSectionTemplate() error = %v", err)
	}

	tests := []struct {
		name string
		data *AgentTemplateData
		want []string
	}{
		{
			name: "execution role principles",
			data: &AgentTemplateData{
				Name:      "coder",
				Role:      "execution",
				RoleTitle: "Implementation",
			},
			want: []string{
				"Constitutional AI Principles",
				"Judgment for Implementation",
				"Implementation Judgment Examples",
				"Testing Decisions",
			},
		},
		{
			name: "planning role principles",
			data: &AgentTemplateData{
				Name:      "planner",
				Role:      "planning",
				RoleTitle: "Planning",
			},
			want: []string{
				"Constitutional AI Principles",
				"Judgment for Planning",
				"Planning Judgment Examples",
				"Task Granularity",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := renderSectionTemplate(tpl, tt.data)
			if err != nil {
				t.Errorf("renderSectionTemplate() error = %v", err)
				return
			}

			for _, want := range tt.want {
				if !strings.Contains(got, want) {
					t.Errorf("renderSectionTemplate() output missing expected string: %q\nGot:\n%s", want, got)
				}
			}
		})
	}
}

func TestAssembleAgentFromSections(t *testing.T) {
	sections, err := loadAllSectionTemplates()
	if err != nil {
		t.Fatalf("loadAllSectionTemplates() error = %v", err)
	}

	data := &AgentTemplateData{
		Name:               "test-coder",
		Role:               "execution",
		RoleTitle:          "Implementation",
		Dependencies:       []string{"tester", "reviewer"},
		DisallowedTools:    []string{"replace_symbol_body"},
		HasDisallowedTools: true,
		AgentContent:       "You are a test agent.",
	}

	got, err := assembleAgentFromSections(data, sections)
	if err != nil {
		t.Fatalf("assembleAgentFromSections() error = %v", err)
	}

	// Check that all sections are present
	expectedSections := []string{
		"You are a test agent",
		"Optimal Tooling",
		"Context Gathering",
		"Constitutional AI Principles",
		"@ent/tester",
		"@ent/reviewer",
	}

	for _, want := range expectedSections {
		if !strings.Contains(got, want) {
			t.Errorf("assembleAgentFromSections() output missing expected section: %q", want)
		}
	}
}
