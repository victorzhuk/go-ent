package generation

//nolint:gosec // test file with necessary file operations

import (
	"testing"
)

func TestSelectArchetype(t *testing.T) {
	cfg := &GenerationConfig{
		Defaults: Defaults{
			GoVersion: "1.25",
			Archetype: "standard",
		},
	}

	tests := []struct {
		name              string
		analysis          *SpecAnalysis
		explicitArchetype string
		wantArchetype     string
		wantConfidence    float64
		checkConfidence   bool
	}{
		{
			name:              "explicit archetype takes precedence",
			analysis:          &SpecAnalysis{},
			explicitArchetype: "mcp",
			wantArchetype:     "mcp",
			wantConfidence:    1.0,
			checkConfidence:   true,
		},
		{
			name: "MCP pattern selects mcp archetype",
			analysis: &SpecAnalysis{
				Patterns: []PatternMatch{
					{Pattern: "mcp", Score: 0.8},
				},
			},
			explicitArchetype: "",
			wantArchetype:     "mcp",
			checkConfidence:   false,
		},
		{
			name: "CRUD pattern selects standard archetype",
			analysis: &SpecAnalysis{
				Patterns: []PatternMatch{
					{Pattern: "crud", Score: 0.7},
					{Pattern: "api", Score: 0.6},
				},
			},
			explicitArchetype: "",
			wantArchetype:     "standard",
			checkConfidence:   false,
		},
		{
			name: "async pattern selects worker archetype",
			analysis: &SpecAnalysis{
				Patterns: []PatternMatch{
					{Pattern: "async", Score: 0.9},
				},
			},
			explicitArchetype: "",
			wantArchetype:     "worker",
			checkConfidence:   false,
		},
		{
			name:              "no patterns uses default",
			analysis:          &SpecAnalysis{},
			explicitArchetype: "",
			wantArchetype:     "standard",
			wantConfidence:    0.0,
			checkConfidence:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arch, confidence, err := SelectArchetype(tt.analysis, cfg, tt.explicitArchetype)

			if err != nil {
				t.Errorf("SelectArchetype() error = %v", err)
				return
			}

			if arch != tt.wantArchetype {
				t.Errorf("SelectArchetype() archetype = %q, want %q", arch, tt.wantArchetype)
			}

			if tt.checkConfidence && confidence != tt.wantConfidence {
				t.Errorf("SelectArchetype() confidence = %v, want %v", confidence, tt.wantConfidence)
			}
		})
	}
}

func TestEnrichAnalysisWithArchetype(t *testing.T) {
	cfg := &GenerationConfig{
		Defaults: Defaults{
			Archetype: "standard",
		},
	}

	analysis := &SpecAnalysis{
		Patterns: []PatternMatch{
			{Pattern: "crud", Score: 0.8},
		},
	}

	err := EnrichAnalysisWithArchetype(analysis, cfg, "")
	if err != nil {
		t.Errorf("EnrichAnalysisWithArchetype() error = %v", err)
	}

	if analysis.Archetype == "" {
		t.Error("expected archetype to be set")
	}

	if analysis.Confidence < 0 || analysis.Confidence > 1 {
		t.Errorf("confidence out of range: %v", analysis.Confidence)
	}
}

func TestGetArchetypesForPattern(t *testing.T) {
	tests := []struct {
		pattern        string
		wantArchetypes []string
	}{
		{"crud", []string{"standard", "api"}},
		{"api", []string{"standard", "api"}},
		{"async", []string{"worker"}},
		{"mcp", []string{"mcp"}},
		{"grpc", []string{"grpc"}},
		{"unknown", []string{"standard"}},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			got := getArchetypesForPattern(tt.pattern)
			if len(got) != len(tt.wantArchetypes) {
				t.Errorf("getArchetypesForPattern(%q) = %v, want %v", tt.pattern, got, tt.wantArchetypes)
				return
			}

			for i, arch := range got {
				if arch != tt.wantArchetypes[i] {
					t.Errorf("archetype[%d] = %q, want %q", i, arch, tt.wantArchetypes[i])
				}
			}
		})
	}
}
