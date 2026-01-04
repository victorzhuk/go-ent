package generation

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAnalyzeSpec(t *testing.T) {
	tests := []struct {
		name        string
		specContent string
		wantErr     bool
		validate    func(t *testing.T, analysis *SpecAnalysis)
	}{
		{
			name: "CRUD pattern detected",
			specContent: `# User Management Spec

## Requirements

The system SHALL allow creating new users.
The system SHALL retrieve user by ID.
The system SHALL update user information.
The system SHALL delete users.
`,
			wantErr: false,
			validate: func(t *testing.T, analysis *SpecAnalysis) {
				foundCRUD := false
				for _, p := range analysis.Patterns {
					if p.Pattern == "crud" && p.Score > 0 {
						foundCRUD = true
					}
				}
				if !foundCRUD {
					t.Error("expected to find CRUD pattern")
				}
			},
		},
		{
			name: "API pattern detected",
			specContent: `# API Spec

The system SHALL provide HTTP endpoint for user creation.
The system SHALL accept JSON request body.
The system SHALL return JSON response.
`,
			wantErr: false,
			validate: func(t *testing.T, analysis *SpecAnalysis) {
				foundAPI := false
				for _, p := range analysis.Patterns {
					if p.Pattern == "api" && p.Score > 0 {
						foundAPI = true
					}
				}
				if !foundAPI {
					t.Error("expected to find API pattern")
				}
			},
		},
		{
			name: "MCP pattern detected",
			specContent: `# MCP Server Spec

The system SHALL provide MCP tools for code generation.
The system SHALL implement Model Context Protocol.
The system SHALL expose resources.
`,
			wantErr: false,
			validate: func(t *testing.T, analysis *SpecAnalysis) {
				foundMCP := false
				for _, p := range analysis.Patterns {
					if p.Pattern == "mcp" && p.Score > 0 {
						foundMCP = true
					}
				}
				if !foundMCP {
					t.Error("expected to find MCP pattern")
				}
			},
		},
		{
			name: "components extracted",
			specContent: `# Component Spec

The system SHALL provide API endpoint for user creation.
The system SHALL validate user input.
The system SHALL store data in repository.
`,
			wantErr: false,
			validate: func(t *testing.T, analysis *SpecAnalysis) {
				if len(analysis.Components) == 0 {
					t.Error("expected components to be extracted")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp spec file
			dir := t.TempDir()
			specPath := filepath.Join(dir, "spec.md")
			if err := os.WriteFile(specPath, []byte(tt.specContent), 0644); err != nil {
				t.Fatal(err)
			}

			analysis, err := AnalyzeSpec(specPath)

			if (err != nil) != tt.wantErr {
				t.Errorf("AnalyzeSpec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && analysis != nil && tt.validate != nil {
				tt.validate(t, analysis)
			}
		})
	}
}

func TestInferComponentType(t *testing.T) {
	tests := []struct {
		desc     string
		wantType string
	}{
		{"provide API endpoint for users", "handler"},
		{"store data in database", "repository"},
		{"process business logic for orders", "usecase"},
		{"run background worker for emails", "worker"},
		{"validate user input", "usecase"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got := inferComponentType(tt.desc)
			if got != tt.wantType {
				t.Errorf("inferComponentType(%q) = %q, want %q", tt.desc, got, tt.wantType)
			}
		})
	}
}

func TestGenerateComponentName(t *testing.T) {
	tests := []struct {
		desc     string
		wantName string
	}{
		{"create new user", "create_new_user"},
		{"Process Order Items", "process_order_items"},
		{"validate-email-address", "validate_email_address"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got := generateComponentName(tt.desc)
			if got != tt.wantName {
				t.Errorf("generateComponentName(%q) = %q, want %q", tt.desc, got, tt.wantName)
			}
		})
	}
}
