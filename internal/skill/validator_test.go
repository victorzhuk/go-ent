package skill

import (
	"os"
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

func TestCheckExampleDiversity_SK010(t *testing.T) {
	tests := []struct {
		name     string
		meta     *SkillMeta
		content  string
		wantWarn bool
		verify   func(t *testing.T, issues []ValidationIssue)
	}{
		{
			name:     "v1 skill skips SK010 validation",
			meta:     &SkillMeta{Name: "test-skill", Description: "Test skill", StructureVersion: "v1"},
			content:  `<examples><example><input>test</input><output>test</output></example></examples>`,
			wantWarn: false,
			verify: func(t *testing.T, issues []ValidationIssue) {
				for _, issue := range issues {
					assert.NotEqual(t, "SK010", issue.Rule)
				}
			},
		},
		{
			name:     "high diversity examples pass",
			meta:     &SkillMeta{Name: "test-skill", Description: "Test skill", StructureVersion: "v2"},
			content:  `<examples><example><input>valid string</input><output>success</output></example><example><input>invalid null</input><output>error</output></example><example><input>zero</input><output>edge handled</output></example></examples>`,
			wantWarn: false,
			verify: func(t *testing.T, issues []ValidationIssue) {
				for _, issue := range issues {
					assert.NotEqual(t, "SK010", issue.Rule)
				}
			},
		},
		{
			name:     "low diversity examples trigger SK010 warning",
			meta:     &SkillMeta{Name: "test-skill", Description: "Test skill", StructureVersion: "v2"},
			content:  `<examples><example><input>test string</input><output>test string</output></example><example><input>another test</input><output>another test</output></example><example><input>more test</input><output>more test</output></example></examples>`,
			wantWarn: true,
			verify: func(t *testing.T, issues []ValidationIssue) {
				found := false
				for _, issue := range issues {
					if issue.Rule == "SK010" {
						found = true
						assert.Equal(t, SeverityWarning, issue.Severity)
						assert.Contains(t, issue.Message, "Low example diversity")
						assert.Contains(t, issue.Message, "SK010")
					}
				}
				assert.True(t, found, "expected to find SK010 issue")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := &ValidationContext{Content: tt.content, Lines: splitLines(tt.content), Meta: tt.meta}
			issues := checkExampleDiversity(ctx)
			if tt.wantWarn {
				assert.True(t, hasSeverity(issues, SeverityWarning), "expected warning")
			} else {
				assert.False(t, hasSeverity(issues, SeverityWarning), "unexpected warning for SK010")
			}
			if tt.verify != nil {
				tt.verify(t, issues)
			}
		})
	}
}

func TestValidationRules_EdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		meta   *SkillMeta
		action func(t *testing.T, ctx *ValidationContext)
	}{
		{
			name: "SK010 handles empty strings without panic",
			meta: &SkillMeta{Name: "test-skill", StructureVersion: "v2"},
			action: func(t *testing.T, ctx *ValidationContext) {
				assert.NotPanics(t, func() { checkExampleDiversity(ctx) })
			},
		},
		{
			name: "SK011 handles empty strings without panic",
			meta: &SkillMeta{Name: "test-skill", StructureVersion: "v2"},
			action: func(t *testing.T, ctx *ValidationContext) {
				assert.NotPanics(t, func() { checkInstructionConcise(ctx) })
			},
		},
		{
			name: "SK012 handles empty triggers without panic",
			meta: &SkillMeta{Name: "test-skill", StructureVersion: "v2", Triggers: []string{}},
			action: func(t *testing.T, ctx *ValidationContext) {
				assert.NotPanics(t, func() { checkTriggerExplicit(ctx) })
			},
		},
		{
			name: "SK013 handles nil registry without panic",
			meta: &SkillMeta{Name: "test-skill", StructureVersion: "v2"},
			action: func(t *testing.T, ctx *ValidationContext) {
				assert.NotPanics(t, func() { checkRedundancy(ctx, nil) })
			},
		},
		{
			name: "v1 skill does not trigger SK rules",
			meta: &SkillMeta{Name: "test-skill", Description: "Test", Triggers: []string{"test"}, StructureVersion: "v1"},
			action: func(t *testing.T, ctx *ValidationContext) {
				ctx.Content = `<instructions>test</instructions><examples><example><input>test</input><output>test</output></example></examples>`
				assert.Empty(t, checkExampleDiversity(ctx), "v1 should not trigger SK010")
				assert.Empty(t, checkInstructionConcise(ctx), "v1 should not trigger SK011")
				assert.Empty(t, checkTriggerExplicit(ctx), "v1 should not trigger SK012")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := &ValidationContext{Content: "", Lines: []string{}, Meta: tt.meta}
			tt.action(t, ctx)
		})
	}
}

// Integration tests for SK010-SK013 validation rules

// TestIntegration_FullValidationWithAllNewRules validates a complete skill file with all new rules (6.2.1)
func TestIntegration_FullValidationWithAllNewRules(t *testing.T) {
	t.Parallel()

	skillContent := `---
name: test-skill
description: Test skill with all sections
version: 1.0.0
triggers:
  - keywords: ["test", "testing"]
    weight: 0.8
---

<role>Expert tester focused on TDD and test patterns</role>

<instructions>
Write comprehensive tests using table-driven patterns.
Ensure proper error handling and edge case coverage.
</instructions>

<examples>
<example>
<input>test string</input>
<output>test string</output>
</example>
<example>
<input>another test</input>
<output>another test</output>
</example>
<example>
<input>more test</input>
<output>more test</output>
</example>
</examples>

<output_format>JSON with test results</output_format>`

	meta := &SkillMeta{
		Name:             "test-skill",
		Description:      "Test skill with all sections",
		Version:          "1.0.0",
		StructureVersion: "v2",
		Triggers:         []string{"test", "testing"},
		ExplicitTriggers: []Trigger{
			{Keywords: []string{"test", "testing"}, Weight: 0.8},
		},
	}

	v := NewValidator()
	result := v.Validate(meta, skillContent)

	assert.True(t, result.Valid, "skill should be valid (warnings don't block validation)")

	rulesFound := make(map[string]bool)
	for _, issue := range result.Issues {
		rulesFound[issue.Rule] = true
	}

	assert.Contains(t, rulesFound, "SK010", "SK010 rule should run")
	assert.True(t, result.WarningCount() > 0, "should have at least one warning")

	for _, issue := range result.Issues {
		if issue.Rule == "SK010" {
			assert.Equal(t, SeverityWarning, issue.Severity)
			assert.Contains(t, issue.Message, "SK010")
		}
	}
}

// TestIntegration_ValidateWithContext_SK013 validates with registry for redundancy detection (6.2.2)
func TestIntegration_ValidateWithContext_SK013(t *testing.T) {
	t.Parallel()

	registry := NewRegistry()

	skill1 := &SkillMeta{
		Name:             "test-1",
		Description:      "Testing patterns with TDD",
		Triggers:         []string{"test", "tdd", "testing"},
		StructureVersion: "v2",
	}

	skill2 := &SkillMeta{
		Name:             "test-2",
		Description:      "TDD testing patterns",
		Triggers:         []string{"test", "tdd"},
		StructureVersion: "v2",
	}

	registry.skills = append(registry.skills, *skill1, *skill2)

	content1 := `<role>Test expert</role><instructions>Write tests</instructions>`
	content2 := `<role>Testing expert</role><instructions>Do TDD</instructions>`

	v := NewValidator()

	result1 := v.ValidateWithContext(skill1, content1, registry)
	result2 := v.ValidateWithContext(skill2, content2, registry)

	sk013Found := false
	for _, issue := range result1.Issues {
		if issue.Rule == "SK013" {
			sk013Found = true
			assert.Equal(t, SeverityWarning, issue.Severity)
			assert.Contains(t, issue.Message, "SK013")
			assert.Contains(t, issue.Message, "overlaps")
		}
	}
	assert.True(t, sk013Found, "SK013 should detect overlap between test-1 and test-2")

	sk013Found2 := false
	for _, issue := range result2.Issues {
		if issue.Rule == "SK013" {
			sk013Found2 = true
		}
	}
	assert.True(t, sk013Found2, "SK013 should detect overlap for test-2 as well")
}

// TestIntegration_WarningsDoNotBlockValidation verifies warnings don't block validation (6.2.3)
func TestIntegration_WarningsDoNotBlockValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		meta    *SkillMeta
		content string
	}{
		{
			name: "SK010 warning allows valid result",
			meta: &SkillMeta{
				Name:             "test-skill",
				Description:      "Test skill",
				Version:          "1.0.0",
				StructureVersion: "v2",
			},
			content: `<role>test</role><instructions>test</instructions>
<examples><example><input>test</input><output>test</output></example>
<example><input>test2</input><output>test2</output></example>
<example><input>test3</input><output>test3</output></example></examples>`,
		},
		{
			name: "SK011 warning allows valid result",
			meta: &SkillMeta{
				Name:             "test-skill",
				Description:      "Test skill",
				Version:          "1.0.0",
				StructureVersion: "v2",
			},
			content: `<role>test</role><instructions>` + strings.Repeat("test ", 5000) + `</instructions>`,
		},
		{
			name: "SK012 warning allows valid result",
			meta: &SkillMeta{
				Name:             "test-skill",
				Description:      "Test skill. Auto-activates for: testing",
				Triggers:         []string{"testing"},
				StructureVersion: "v2",
			},
			content: `<role>test</role><instructions>test</instructions>`,
		},
		{
			name: "Multiple warnings still valid",
			meta: &SkillMeta{
				Name:             "test-skill",
				Description:      "Test skill. Auto-activates for: testing",
				Triggers:         []string{"testing"},
				Version:          "1.0.0",
				StructureVersion: "v2",
			},
			content: `<role>test</role><instructions>` + strings.Repeat("test ", 5000) + `</instructions>
<examples><example><input>test</input><output>test</output></example>
<example><input>test2</input><output>test2</output></example>
<example><input>test3</input><output>test3</output></example></examples>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := NewValidator()
			result := v.Validate(tt.meta, tt.content)

			assert.True(t, result.Valid, "skill with warnings should be valid")
			assert.True(t, result.ErrorCount() == 0, "should have no errors")
			assert.True(t, result.WarningCount() > 0, "should have warnings")
		})
	}
}

// TestIntegration_OnlyErrorsBlockValidation verifies only errors block validation (6.2.3)
func TestIntegration_OnlyErrorsBlockValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		meta         *SkillMeta
		content      string
		expectValid  bool
		expectErrors bool
		expectWarn   bool
	}{
		{
			name:         "Only warnings - valid",
			meta:         &SkillMeta{Name: "test", Description: "Test", StructureVersion: "v2"},
			content:      `<role>test</role>`,
			expectValid:  true,
			expectErrors: false,
			expectWarn:   true,
		},
		{
			name:         "Errors and warnings - invalid",
			meta:         &SkillMeta{Name: "", Description: "Test", StructureVersion: "v2"},
			content:      `<role>test</role>`,
			expectValid:  false,
			expectErrors: true,
			expectWarn:   true,
		},
		{
			name:         "No issues - valid",
			meta:         &SkillMeta{Name: "test", Description: "Test", StructureVersion: "v1"},
			content:      `---\nname: test\ndescription: Test\n---`,
			expectValid:  true,
			expectErrors: false,
			expectWarn:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := NewValidator()
			result := v.Validate(tt.meta, tt.content)

			assert.Equal(t, tt.expectValid, result.Valid, "validity mismatch")
			assert.Equal(t, tt.expectErrors, result.ErrorCount() > 0, "error presence mismatch")
			assert.Equal(t, tt.expectWarn, result.WarningCount() > 0, "warning presence mismatch")
		})
	}
}

// TestIntegration_ValidateRealSkillFiles validates real skill files from repository (6.2.4)
func TestIntegration_ValidateRealSkillFiles(t *testing.T) {
	t.Parallel()

	realSkills := []struct {
		path     string
		expected string
	}{
		{
			path:     ".claude/skills/ent/go/go-code/SKILL.md",
			expected: "go-code",
		},
	}

	registry := NewRegistry()

	for _, skillInfo := range realSkills {
		t.Run(skillInfo.expected, func(t *testing.T) {
			if _, err := os.Stat(skillInfo.path); os.IsNotExist(err) {
				t.Skipf("skill file not found: %s", skillInfo.path)
				return
			}

			err := registry.RegisterSkill(skillInfo.expected, skillInfo.path)
			if err != nil {
				t.Fatalf("failed to load skill: %v", err)
			}

			result, err := registry.ValidateSkill(skillInfo.expected)
			assert.NoError(t, err, "validation should succeed")
			assert.NotNil(t, result, "result should not be nil")
			assert.NotNil(t, result.Issues, "issues should not be nil")

			rulesFound := make(map[string]bool)
			for _, issue := range result.Issues {
				rulesFound[issue.Rule] = true
			}

			hasWarnings := result.WarningCount() > 0
			hasErrors := result.ErrorCount() > 0
			isValid := result.Valid

			if !hasErrors {
				assert.True(t, isValid, "skill with only warnings should be valid")
			}

			t.Logf("Skill %s: valid=%v, errors=%v, warnings=%v, rules=%v",
				skillInfo.expected, isValid, hasErrors, hasWarnings, rulesFound)
		})
	}
}

// TestIntegration_MultipleSkillsWithRegistry tests validation of multiple skills together
func TestIntegration_MultipleSkillsWithRegistry(t *testing.T) {
	t.Parallel()

	registry := NewRegistry()

	skillA := &SkillMeta{
		Name:             "skill-a",
		Description:      "Go code patterns",
		Triggers:         []string{"go", "golang"},
		StructureVersion: "v2",
	}

	skillB := &SkillMeta{
		Name:             "skill-b",
		Description:      "Testing patterns",
		Triggers:         []string{"test", "testing"},
		StructureVersion: "v2",
	}

	registry.skills = append(registry.skills, *skillA, *skillB)

	v := NewValidator()

	contentA := `<role>Expert in Go</role><instructions>Write Go code</instructions>`
	contentB := `<role>Expert in testing</role><instructions>Write tests</instructions>`

	resultA := v.Validate(skillA, contentA)
	resultB := v.Validate(skillB, contentB)

	assert.NotNil(t, resultA, "resultA should not be nil")
	assert.NotNil(t, resultB, "resultB should not be nil")
	assert.NotNil(t, resultA.Issues, "issues A should not be nil")
	assert.NotNil(t, resultB.Issues, "issues B should not be nil")

	totalIssues := len(resultA.Issues) + len(resultB.Issues)
	t.Logf("Validated %d skills with %d total issues", 2, totalIssues)

	ruleCounts := make(map[string]int)
	for _, issue := range resultA.Issues {
		ruleCounts[issue.Rule]++
	}
	for _, issue := range resultB.Issues {
		ruleCounts[issue.Rule]++
	}

	for rule, count := range ruleCounts {
		t.Logf("Rule %s: %d issues", rule, count)
	}
}

// TestIntegration_SK010DiversityScore tests diversity score calculation
func TestIntegration_SK010DiversityScore(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		content     string
		expectIssue bool
	}{
		{
			name:        "High diversity - no issue",
			expectIssue: false,
			content: `<examples>
<example><input>valid string</input><output>success</output></example>
<example><input>""</input><output>error</output></example>
<example><input>0</input><output>edge case</output></example>
</examples>`,
		},
		{
			name:        "Low diversity - issue",
			expectIssue: true,
			content: `<examples>
<example><input>test string</input><output>test string</output></example>
<example><input>test string 2</input><output>test string 2</output></example>
<example><input>test string 3</input><output>test string 3</output></example>
</examples>`,
		},
		{
			name:        "Too few examples - no issue",
			expectIssue: false,
			content: `<examples>
<example><input>test</input><output>result</output></example>
</examples>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			meta := &SkillMeta{Name: "test", Description: "Test", StructureVersion: "v2"}
			ctx := &ValidationContext{Content: tt.content, Lines: splitLines(tt.content), Meta: meta}

			issues := checkExampleDiversity(ctx)

			if tt.expectIssue {
				assert.True(t, len(issues) > 0, "expected SK010 issue")
				if len(issues) > 0 {
					assert.Equal(t, "SK010", issues[0].Rule)
					assert.Equal(t, SeverityWarning, issues[0].Severity)
				}
			} else {
				for _, issue := range issues {
					assert.NotEqual(t, "SK010", issue.Rule, "unexpected SK010 issue")
				}
			}
		})
	}
}

// TestIntegration_SK011InstructionLength tests instruction length validation
func TestIntegration_SK011InstructionLength(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		content     string
		expectIssue bool
	}{
		{
			name:        "Short instructions - no issue",
			expectIssue: false,
			content:     `<instructions>Short instructions</instructions>`,
		},
		{
			name:        "Medium instructions - no issue",
			expectIssue: false,
			content:     `<instructions>` + strings.Repeat("word ", 2000) + `</instructions>`,
		},
		{
			name:        "Long instructions - issue",
			expectIssue: true,
			content:     `<instructions>` + strings.Repeat("word ", 8000) + `</instructions>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			meta := &SkillMeta{Name: "test", Description: "Test", StructureVersion: "v2"}
			ctx := &ValidationContext{Content: tt.content, Lines: splitLines(tt.content), Meta: meta}

			issues := checkInstructionConcise(ctx)

			if tt.expectIssue {
				assert.True(t, len(issues) > 0, "expected SK011 issue")
				if len(issues) > 0 {
					assert.Equal(t, "SK011", issues[0].Rule)
					assert.Equal(t, SeverityWarning, issues[0].Severity)
					assert.Contains(t, issues[0].Message, "tokens")
				}
			} else {
				for _, issue := range issues {
					assert.NotEqual(t, "SK011", issue.Rule, "unexpected SK011 issue")
				}
			}
		})
	}
}
