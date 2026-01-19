package skill

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateFrontmatter(t *testing.T) {
	tests := []struct {
		name     string
		meta     *SkillMeta
		content  string
		strict   bool
		wantErr  bool
		wantWarn bool
	}{
		{
			name: "valid frontmatter",
			meta: &SkillMeta{
				Name:             "test-skill",
				Description:      "A test skill",
				StructureVersion: "v1",
			},
			content: "---\nname: test-skill\ndescription: A test skill\n---",
			strict:  false,
		},
		{
			name: "missing name",
			meta: &SkillMeta{
				Description:      "A test skill",
				StructureVersion: "v1",
			},
			content: "---\ndescription: A test skill\n---",
			strict:  false,
			wantErr: true,
		},
		{
			name: "missing description",
			meta: &SkillMeta{
				Name:             "test-skill",
				StructureVersion: "v1",
			},
			content: "---\nname: test-skill\n---",
			strict:  false,
			wantErr: true,
		},
		{
			name: "v2 without version warning",
			meta: &SkillMeta{
				Name:             "test-skill",
				Description:      "A test skill",
				StructureVersion: "v2",
			},
			content:  "---\nname: test-skill\ndescription: A test skill\n---",
			strict:   false,
			wantWarn: true,
		},
		{
			name: "v2 without version strict error",
			meta: &SkillMeta{
				Name:             "test-skill",
				Description:      "A test skill",
				StructureVersion: "v2",
			},
			content: "---\nname: test-skill\ndescription: A test skill\n---",
			strict:  true,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := &ValidationContext{
				Content: tt.content,
				Lines:   splitLines(tt.content),
				Meta:    tt.meta,
				Strict:  tt.strict,
			}
			issues := validateFrontmatter(ctx)

			if tt.wantErr {
				assert.True(t, hasSeverity(issues, SeverityError), "expected error")
			} else {
				assert.False(t, hasSeverity(issues, SeverityError), "unexpected error")
			}

			if tt.wantWarn {
				assert.True(t, hasSeverity(issues, SeverityWarning), "expected warning")
			}
		})
	}
}

func TestValidateVersion(t *testing.T) {
	tests := []struct {
		name    string
		meta    *SkillMeta
		content string
		wantErr bool
	}{
		{
			name: "no version",
			meta: &SkillMeta{
				Name:    "test",
				Version: "",
			},
			content: "---\nname: test\n---",
		},
		{
			name: "valid semver with v",
			meta: &SkillMeta{
				Name:    "test",
				Version: "v1.0.0",
			},
			content: "---\nname: test\nversion: v1.0.0\n---",
		},
		{
			name: "valid semver without v",
			meta: &SkillMeta{
				Name:    "test",
				Version: "2.1.3",
			},
			content: "---\nname: test\nversion: 2.1.3\n---",
		},
		{
			name: "invalid format",
			meta: &SkillMeta{
				Name:    "test",
				Version: "1.0",
			},
			content: "---\nname: test\nversion: 1.0\n---",
			wantErr: true,
		},
		{
			name: "invalid text",
			meta: &SkillMeta{
				Name:    "test",
				Version: "latest",
			},
			content: "---\nname: test\nversion: latest\n---",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := &ValidationContext{
				Content: tt.content,
				Lines:   splitLines(tt.content),
				Meta:    tt.meta,
			}
			issues := validateVersion(ctx)

			if tt.wantErr {
				assert.True(t, hasSeverity(issues, SeverityError), "expected error")
			} else {
				assert.Empty(t, issues, "expected no issues")
			}
		})
	}
}

func TestValidateXMLTags(t *testing.T) {
	tests := []struct {
		name    string
		meta    *SkillMeta
		content string
		wantErr bool
	}{
		{
			name: "v1 skips validation",
			meta: &SkillMeta{
				Name:             "test",
				StructureVersion: "v1",
			},
			content: "---\nname: test\n---",
		},
		{
			name: "balanced tags",
			meta: &SkillMeta{
				Name:             "test",
				StructureVersion: "v2",
			},
			content: "---\nname: test\n---\n<role></role><instructions></instructions>",
		},
		{
			name: "unbalanced role tag",
			meta: &SkillMeta{
				Name:             "test",
				StructureVersion: "v2",
			},
			content: "---\nname: test\n---\n<role>content",
			wantErr: true,
		},
		{
			name: "duplicate tags",
			meta: &SkillMeta{
				Name:             "test",
				StructureVersion: "v2",
			},
			content: "---\nname: test\n---\n<role>a</role><role>b</role>",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := &ValidationContext{
				Content: tt.content,
				Lines:   splitLines(tt.content),
				Meta:    tt.meta,
			}
			issues := validateXMLTags(ctx)

			if tt.wantErr {
				assert.True(t, hasSeverity(issues, SeverityError), "expected error")
			} else {
				assert.Empty(t, issues, "expected no issues")
			}
		})
	}
}

func TestValidateRoleSection(t *testing.T) {
	tests := []struct {
		name     string
		meta     *SkillMeta
		content  string
		strict   bool
		wantErr  bool
		wantWarn bool
	}{
		{
			name: "v1 skips validation",
			meta: &SkillMeta{
				Name:             "test",
				StructureVersion: "v1",
			},
			content: "---\nname: test\n---",
		},
		{
			name: "valid role section",
			meta: &SkillMeta{
				Name:             "test",
				StructureVersion: "v2",
			},
			content: "<role>You are an expert Go developer focused on clean architecture and best practices.</role>",
		},
		{
			name: "missing role section warning",
			meta: &SkillMeta{
				Name:             "test",
				StructureVersion: "v2",
			},
			content:  "<instructions>do something</instructions>",
			wantWarn: true,
		},
		{
			name: "missing role section strict error",
			meta: &SkillMeta{
				Name:             "test",
				StructureVersion: "v2",
			},
			content: "<instructions>do something</instructions>",
			strict:  true,
			wantErr: true,
		},
		{
			name: "unclosed role tag",
			meta: &SkillMeta{
				Name:             "test",
				StructureVersion: "v2",
			},
			content: "<role>content",
			wantErr: true,
		},
		{
			name: "empty role section",
			meta: &SkillMeta{
				Name:             "test",
				StructureVersion: "v2",
			},
			content: "<role></role>",
			wantErr: true,
		},
		{
			name: "role too short warning",
			meta: &SkillMeta{
				Name:             "test",
				StructureVersion: "v2",
			},
			content:  "<role>test</role>",
			wantWarn: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := &ValidationContext{
				Content: tt.content,
				Lines:   splitLines(tt.content),
				Meta:    tt.meta,
				Strict:  tt.strict,
			}
			issues := validateRoleSection(ctx)

			if tt.wantErr {
				assert.True(t, hasSeverity(issues, SeverityError), "expected error")
			} else {
				assert.False(t, hasSeverity(issues, SeverityError), "unexpected error")
			}

			if tt.wantWarn {
				assert.True(t, hasSeverity(issues, SeverityWarning), "expected warning")
			}
		})
	}
}

func TestValidateInstructionsSection(t *testing.T) {
	tests := []struct {
		name     string
		meta     *SkillMeta
		content  string
		strict   bool
		wantErr  bool
		wantWarn bool
	}{
		{
			name: "v1 skips validation",
			meta: &SkillMeta{
				Name:             "test",
				StructureVersion: "v1",
			},
			content: "---\nname: test\n---",
		},
		{
			name: "valid instructions section",
			meta: &SkillMeta{
				Name:             "test",
				StructureVersion: "v2",
			},
			content: "<instructions>Do the following tasks...</instructions>",
		},
		{
			name: "missing instructions warning",
			meta: &SkillMeta{
				Name:             "test",
				StructureVersion: "v2",
			},
			content:  "<role>test</role>",
			wantWarn: true,
		},
		{
			name: "unclosed instructions tag",
			meta: &SkillMeta{
				Name:             "test",
				StructureVersion: "v2",
			},
			content: "<instructions>content",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := &ValidationContext{
				Content: tt.content,
				Lines:   splitLines(tt.content),
				Meta:    tt.meta,
				Strict:  tt.strict,
			}
			issues := validateInstructionsSection(ctx)

			if tt.wantErr {
				assert.True(t, hasSeverity(issues, SeverityError), "expected error")
			}

			if tt.wantWarn {
				assert.True(t, hasSeverity(issues, SeverityWarning), "expected warning")
			}
		})
	}
}

func TestValidateExamples(t *testing.T) {
	tests := []struct {
		name     string
		meta     *SkillMeta
		content  string
		wantErr  bool
		wantWarn bool
	}{
		{
			name: "v1 skips validation",
			meta: &SkillMeta{
				Name:             "test",
				StructureVersion: "v1",
			},
			content: "---\nname: test\n---",
		},
		{
			name: "no examples section",
			meta: &SkillMeta{
				Name:             "test",
				StructureVersion: "v2",
			},
			content: "<role>test</role>",
		},
		{
			name: "valid examples with input/output",
			meta: &SkillMeta{
				Name:             "test",
				StructureVersion: "v2",
			},
			content: `<examples>
<example>
<input>test input</input>
<output>test output</output>
</example>
</examples>`,
		},
		{
			name: "empty examples section",
			meta: &SkillMeta{
				Name:             "test",
				StructureVersion: "v2",
			},
			content:  `<examples></examples>`,
			wantWarn: true,
		},
		{
			name: "unclosed examples tag",
			meta: &SkillMeta{
				Name:             "test",
				StructureVersion: "v2",
			},
			content: `<examples><example><input>test</input><output>test</output></example>`,
			wantErr: true,
		},
		{
			name: "missing input/output tags",
			meta: &SkillMeta{
				Name:             "test",
				StructureVersion: "v2",
			},
			content: `<examples><example>content</example></examples>`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := &ValidationContext{
				Content: tt.content,
				Lines:   splitLines(tt.content),
				Meta:    tt.meta,
			}
			issues := validateExamples(ctx)

			if tt.wantErr {
				assert.True(t, hasSeverity(issues, SeverityError), "expected error")
			}

			if tt.wantWarn {
				assert.True(t, hasSeverity(issues, SeverityWarning), "expected warning")
			}
		})
	}
}

func TestValidateConstraints(t *testing.T) {
	tests := []struct {
		name     string
		meta     *SkillMeta
		content  string
		wantErr  bool
		wantWarn bool
	}{
		{
			name: "v1 skips validation",
			meta: &SkillMeta{
				Name:             "test",
				StructureVersion: "v1",
			},
			content: "---\nname: test\n---",
		},
		{
			name: "no constraints section",
			meta: &SkillMeta{
				Name:             "test",
				StructureVersion: "v2",
			},
			content: "<role>test</role>",
		},
		{
			name: "valid constraints with list items",
			meta: &SkillMeta{
				Name:             "test",
				StructureVersion: "v2",
			},
			content: `<constraints>- Follow clean code principles
- Use interfaces at consumer side
- Wrap errors with context</constraints>`,
		},
		{
			name: "empty constraints section",
			meta: &SkillMeta{
				Name:             "test",
				StructureVersion: "v2",
			},
			content:  `<constraints></constraints>`,
			wantWarn: true,
		},
		{
			name: "unclosed constraints tag",
			meta: &SkillMeta{
				Name:             "test",
				StructureVersion: "v2",
			},
			content: `<constraints>- test`,
			wantErr: true,
		},
		{
			name: "constraints not in list format",
			meta: &SkillMeta{
				Name:             "test",
				StructureVersion: "v2",
			},
			content: `<constraints>Test constraint
Another constraint</constraints>`,
			wantWarn: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := &ValidationContext{
				Content: tt.content,
				Lines:   splitLines(tt.content),
				Meta:    tt.meta,
			}
			issues := validateConstraints(ctx)

			if tt.wantErr {
				assert.True(t, hasSeverity(issues, SeverityError), "expected error")
			}

			if tt.wantWarn {
				assert.True(t, hasSeverity(issues, SeverityWarning), "expected warning")
			}
		})
	}
}

func TestValidateEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		meta     *SkillMeta
		content  string
		wantErr  bool
		wantWarn bool
	}{
		{
			name: "v1 skips validation",
			meta: &SkillMeta{
				Name:             "test",
				StructureVersion: "v1",
			},
			content: "---\nname: test\n---",
		},
		{
			name: "no edge cases section",
			meta: &SkillMeta{
				Name:             "test",
				StructureVersion: "v2",
			},
			content: "<role>test</role>",
		},
		{
			name: "valid edge cases with scenarios",
			meta: &SkillMeta{
				Name:             "test",
				StructureVersion: "v2",
			},
			content: `<edge_cases>If you encounter database errors, delegate to go-db
When performance issues arise, delegate to go-perf</edge_cases>`,
		},
		{
			name: "insufficient scenarios",
			meta: &SkillMeta{
				Name:             "test",
				StructureVersion: "v2",
			},
			content:  `<edge_cases>Test scenario</edge_cases>`,
			wantWarn: true,
		},
		{
			name: "unclosed edge_cases tag",
			meta: &SkillMeta{
				Name:             "test",
				StructureVersion: "v2",
			},
			content: `<edge_cases>If test then delegate`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := &ValidationContext{
				Content: tt.content,
				Lines:   splitLines(tt.content),
				Meta:    tt.meta,
			}
			issues := validateEdgeCases(ctx)

			if tt.wantErr {
				assert.True(t, hasSeverity(issues, SeverityError), "expected error")
			}

			if tt.wantWarn {
				assert.True(t, hasSeverity(issues, SeverityWarning), "expected warning")
			}
		})
	}
}

func TestValidateOutputFormat(t *testing.T) {
	tests := []struct {
		name     string
		meta     *SkillMeta
		content  string
		strict   bool
		wantErr  bool
		wantWarn bool
	}{
		{
			name: "v1 skips validation",
			meta: &SkillMeta{
				Name:             "test",
				StructureVersion: "v1",
			},
			content: "---\nname: test\n---",
		},
		{
			name: "valid output format",
			meta: &SkillMeta{
				Name:             "test",
				StructureVersion: "v2",
			},
			content: `<output_format>Return JSON with keys: code, message</output_format>`,
		},
		{
			name: "missing output format warning",
			meta: &SkillMeta{
				Name:             "test",
				StructureVersion: "v2",
			},
			content:  "<role>test</role>",
			wantWarn: true,
		},
		{
			name: "missing output format strict error",
			meta: &SkillMeta{
				Name:             "test",
				StructureVersion: "v2",
			},
			content: "<role>test</role>",
			strict:  true,
			wantErr: true,
		},
		{
			name: "unclosed output_format tag",
			meta: &SkillMeta{
				Name:             "test",
				StructureVersion: "v2",
			},
			content: `<output_format>Return JSON`,
			wantErr: true,
		},
		{
			name: "empty output format",
			meta: &SkillMeta{
				Name:             "test",
				StructureVersion: "v2",
			},
			content:  `<output_format></output_format>`,
			wantWarn: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := &ValidationContext{
				Content: tt.content,
				Lines:   splitLines(tt.content),
				Meta:    tt.meta,
				Strict:  tt.strict,
			}
			issues := validateOutputFormat(ctx)

			if tt.wantErr {
				assert.True(t, hasSeverity(issues, SeverityError), "expected error")
			} else {
				assert.False(t, hasSeverity(issues, SeverityError), "unexpected error")
			}

			if tt.wantWarn {
				assert.True(t, hasSeverity(issues, SeverityWarning), "expected warning")
			}
		})
	}
}

func TestValidator_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		meta    *SkillMeta
		content string
		wantErr bool
	}{
		{
			name: "valid v2 skill",
			meta: &SkillMeta{
				Name:             "test",
				Description:      "test skill",
				Version:          "1.0.0",
				StructureVersion: "v2",
			},
			content: `<role>You are a tester</role>
<instructions>Test things</instructions>
<examples>
<example>
<input>test</input>
<output>result</output>
</example>
</examples>
<output_format>JSON</output_format>`,
		},
		{
			name: "v1 skill with no errors",
			meta: &SkillMeta{
				Name:             "test",
				Description:      "test skill",
				StructureVersion: "v1",
			},
			content: "---\nname: test\ndescription: test skill\n---\ncontent",
		},
		{
			name: "multiple issues",
			meta: &SkillMeta{
				Name:             "",
				Description:      "",
				StructureVersion: "v2",
			},
			content: "invalid content",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := NewValidator()
			result := v.Validate(tt.meta, tt.content)

			if tt.wantErr {
				assert.True(t, result.ErrorCount() > 0, "expected errors")
				assert.False(t, result.Valid, "expected invalid result")
			} else {
				assert.False(t, hasSeverity(result.Issues, SeverityError), "unexpected errors")
				assert.True(t, result.Valid, "expected valid result")
			}
		})
	}
}

func TestValidator_ValidateStrict(t *testing.T) {
	t.Parallel()

	meta := &SkillMeta{
		Name:             "test",
		Description:      "test skill",
		StructureVersion: "v2",
	}

	content := `<role>test</role>`

	v := NewValidator()
	result := v.ValidateStrict(meta, content)

	assert.False(t, result.Valid, "expected invalid in strict mode")
	assert.True(t, len(result.Issues) > 0, "expected issues in strict mode")
}

func TestValidationResult_ErrorCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		issues []ValidationIssue
		want   int
	}{
		{
			name:   "no issues",
			issues: []ValidationIssue{},
			want:   0,
		},
		{
			name: "one error",
			issues: []ValidationIssue{
				{Severity: SeverityError},
			},
			want: 1,
		},
		{
			name: "errors and warnings",
			issues: []ValidationIssue{
				{Severity: SeverityError},
				{Severity: SeverityWarning},
				{Severity: SeverityError},
				{Severity: SeverityInfo},
			},
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := &ValidationResult{Issues: tt.issues}
			assert.Equal(t, tt.want, result.ErrorCount())
		})
	}
}

func TestValidationResult_WarningCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		issues []ValidationIssue
		want   int
	}{
		{
			name:   "no issues",
			issues: []ValidationIssue{},
			want:   0,
		},
		{
			name: "one warning",
			issues: []ValidationIssue{
				{Severity: SeverityWarning},
			},
			want: 1,
		},
		{
			name: "errors and warnings",
			issues: []ValidationIssue{
				{Severity: SeverityError},
				{Severity: SeverityWarning},
				{Severity: SeverityError},
				{Severity: SeverityInfo},
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := &ValidationResult{Issues: tt.issues}
			assert.Equal(t, tt.want, result.WarningCount())
		})
	}
}

func TestValidationIssue_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		issue ValidationIssue
		want  string
	}{
		{
			name: "with line number",
			issue: ValidationIssue{
				Rule:     "test-rule",
				Severity: SeverityError,
				Message:  "test message",
				Line:     10,
			},
			want: "[error] test-rule:10: test message",
		},
		{
			name: "without line number",
			issue: ValidationIssue{
				Rule:     "test-rule",
				Severity: SeverityError,
				Message:  "test message",
			},
			want: "[error] test-rule: test message",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.issue.String())
		})
	}
}

func TestFindLineNumber(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		lines   []string
		pattern string
		want    int
	}{
		{
			name:    "pattern found",
			lines:   []string{"line 1", "version: 1.0.0", "line 3"},
			pattern: `version:`,
			want:    2,
		},
		{
			name:    "pattern not found",
			lines:   []string{"line 1", "line 2", "line 3"},
			pattern: `version:`,
			want:    0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := findLineNumber(tt.lines, tt.pattern)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFindLineNumberForTag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		lines []string
		tag   string
		want  int
	}{
		{
			name:  "tag found",
			lines: []string{"line 1", "<role>content", "line 3"},
			tag:   "role",
			want:  2,
		},
		{
			name:  "tag not found",
			lines: []string{"line 1", "line 2", "line 3"},
			tag:   "role",
			want:  0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := findLineNumberForTag(tt.lines, tt.tag)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestHasErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		issues []ValidationIssue
		want   bool
	}{
		{
			name:   "no issues",
			issues: []ValidationIssue{},
			want:   false,
		},
		{
			name: "warnings only",
			issues: []ValidationIssue{
				{Severity: SeverityWarning},
				{Severity: SeverityInfo},
			},
			want: false,
		},
		{
			name: "has error",
			issues: []ValidationIssue{
				{Severity: SeverityWarning},
				{Severity: SeverityError},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := hasErrors(tt.issues)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestHasSeverity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		issues   []ValidationIssue
		severity Severity
		want     bool
	}{
		{
			name:     "has error",
			issues:   []ValidationIssue{{Severity: SeverityError}},
			severity: SeverityError,
			want:     true,
		},
		{
			name:     "no error",
			issues:   []ValidationIssue{{Severity: SeverityWarning}},
			severity: SeverityError,
			want:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := hasSeverity(tt.issues, tt.severity)
			assert.Equal(t, tt.want, got)
		})
	}
}

func splitLines(content string) []string {
	return strings.Split(content, "\n")
}

func hasSeverity(issues []ValidationIssue, severity Severity) bool {
	for _, issue := range issues {
		if issue.Severity == severity {
			return true
		}
	}
	return false
}

func TestValidateExplicitTriggers_SK012(t *testing.T) {
	tests := []struct {
		name     string
		meta     *SkillMeta
		content  string
		wantWarn bool
		verify   func(t *testing.T, issues []ValidationIssue)
	}{
		{
			name: "v1 skill skips SK012 validation",
			meta: &SkillMeta{
				Name:             "test-skill",
				Description:      "Auto-activates for: testing, TDD",
				StructureVersion: "v1",
				Triggers:         []string{"testing", "tdd"},
			},
			content:  "---\nname: test-skill\ndescription: Auto-activates for: testing, TDD\n---",
			wantWarn: false,
			verify: func(t *testing.T, issues []ValidationIssue) {
				for _, issue := range issues {
					assert.NotEqual(t, "SK012", issue.Rule)
				}
			},
		},
		{
			name: "v2 with explicit triggers does not trigger SK012",
			meta: &SkillMeta{
				Name:             "test-skill",
				Description:      "Go testing patterns",
				StructureVersion: "v2",
				ExplicitTriggers: []Trigger{
					{
						Patterns: []string{"write.*test"},
						Keywords: []string{"testing", "tdd"},
						Weight:   0.8,
					},
				},
			},
			content:  "---\nname: test-skill\ndescription: Go testing patterns\n---\n<role>test</role>",
			wantWarn: false,
			verify: func(t *testing.T, issues []ValidationIssue) {
				for _, issue := range issues {
					assert.NotEqual(t, "SK012", issue.Rule)
				}
			},
		},
		{
			name: "v2 with description-based triggers triggers SK012",
			meta: &SkillMeta{
				Name:             "test-skill",
				Description:      "Go testing patterns. Auto-activates for: testing, TDD",
				StructureVersion: "v2",
				Triggers:         []string{"testing", "tdd"},
				ExplicitTriggers: nil,
			},
			content:  "---\nname: test-skill\ndescription: Go testing patterns. Auto-activates for: testing, TDD\n---\n<role>test</role>",
			wantWarn: true,
			verify: func(t *testing.T, issues []ValidationIssue) {
				found := false
				for _, issue := range issues {
					if issue.Rule == "SK012" {
						found = true
						assert.Equal(t, SeverityInfo, issue.Severity)
						assert.Contains(t, issue.Message, "Consider using explicit triggers for better control")
						assert.Contains(t, issue.Message, "triggers:")
						assert.Contains(t, issue.Message, "pattern:")
						assert.Contains(t, issue.Message, "weight:")
						assert.Contains(t, issue.Message, "keywords:")
					}
				}
				assert.True(t, found, "expected to find SK012 issue")
			},
		},
		{
			name: "v2 with empty explicit triggers list triggers SK012",
			meta: &SkillMeta{
				Name:             "test-skill",
				Description:      "Go testing patterns",
				StructureVersion: "v2",
				ExplicitTriggers: []Trigger{},
			},
			content:  "---\nname: test-skill\ndescription: Go testing patterns\ntriggers: []\n---\n<role>test</role>",
			wantWarn: true,
			verify: func(t *testing.T, issues []ValidationIssue) {
				found := false
				for _, issue := range issues {
					if issue.Rule == "SK012" {
						found = true
						assert.Equal(t, SeverityInfo, issue.Severity)
					}
				}
				assert.True(t, found, "expected to find SK012 issue")
			},
		},
		{
			name: "v2 with multiple explicit triggers does not trigger SK012",
			meta: &SkillMeta{
				Name:             "test-skill",
				Description:      "Go testing patterns",
				StructureVersion: "v2",
				ExplicitTriggers: []Trigger{
					{
						Patterns: []string{"write.*test"},
						Weight:   0.9,
					},
					{
						Keywords: []string{"testing", "tdd"},
						Weight:   0.8,
					},
				},
			},
			content:  "---\nname: test-skill\ndescription: Go testing patterns\n---\n<role>test</role>",
			wantWarn: false,
			verify: func(t *testing.T, issues []ValidationIssue) {
				for _, issue := range issues {
					assert.NotEqual(t, "SK012", issue.Rule)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := &ValidationContext{
				Content: tt.content,
				Lines:   splitLines(tt.content),
				Meta:    tt.meta,
			}

			issues := validateExplicitTriggers(ctx)

			if tt.wantWarn {
				assert.True(t, hasSeverity(issues, SeverityInfo), "expected info warning")
			} else {
				assert.False(t, hasSeverity(issues, SeverityInfo), "unexpected info warning for SK012")
			}

			if tt.verify != nil {
				tt.verify(t, issues)
			}
		})
	}
}
