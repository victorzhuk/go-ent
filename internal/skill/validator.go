package skill

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/victorzhuk/go-ent/internal/skill/domain"
)

// Severity indicates the severity of a validation issue.
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

// ValidationIssue represents a validation error or warning.
type ValidationIssue struct {
	Rule       string
	Severity   Severity
	Message    string
	Suggestion string
	Example    string
	Line       int
	Column     int
}

func (v ValidationIssue) String() string {
	loc := v.Rule
	if v.Line > 0 {
		loc = fmt.Sprintf("%s:%d", v.Rule, v.Line)
	}
	result := fmt.Sprintf("[%s] %s: %s", v.Severity, loc, v.Message)
	if v.Suggestion != "" {
		result += fmt.Sprintf("\n  Suggestion: %s", v.Suggestion)
	}
	if v.Example != "" {
		result += fmt.Sprintf("\n  Example: %s", v.Example)
	}
	return result
}

// ValidationResult holds the results of validation.
type ValidationResult struct {
	Valid  bool
	Issues []ValidationIssue
}

// ErrorCount returns the number of errors.
func (r *ValidationResult) ErrorCount() int {
	count := 0
	for _, issue := range r.Issues {
		if issue.Severity == SeverityError {
			count++
		}
	}
	return count
}

// WarningCount returns the number of warnings.
func (r *ValidationResult) WarningCount() int {
	count := 0
	for _, issue := range r.Issues {
		if issue.Severity == SeverityWarning {
			count++
		}
	}
	return count
}

// ValidationContext provides context for validation rules.
type ValidationContext struct {
	FilePath string
	Content  string
	Lines    []string
	Meta     *domain.Info
	Strict   bool
}

// ValidationRule is a function that validates content and returns issues.
type ValidationRule func(ctx *ValidationContext) []ValidationIssue

// Validator validates skill files using a set of rules.
type Validator struct {
	rules []ValidationRule
}

// NewValidator creates a new validator with default rules.
func NewValidator() *Validator {
	return &Validator{
		rules: []ValidationRule{
			validateFrontmatterV4,
			validateNameFormatV4,
			validateRoleSectionV4,
			validateInstructionsSectionV4,
			validateExamplesSectionV4,
			validateReferencesV4,
		},
	}
}

// Validate validates a skill's metadata and content.
func (v *Validator) Validate(meta *domain.Info, content string) *ValidationResult {
	ctx := &ValidationContext{
		FilePath: meta.FilePath,
		Content:  content,
		Lines:    strings.Split(content, "\n"),
		Meta:     meta,
		Strict:   false,
	}

	var issues []ValidationIssue
	for _, rule := range v.rules {
		issues = append(issues, rule(ctx)...)
	}

	result := &ValidationResult{
		Issues: issues,
		Valid:  true,
	}

	if ctx.Strict {
		result.Valid = len(issues) == 0
	} else {
		result.Valid = result.ErrorCount() == 0
	}

	return result
}

// ValidateStrict validates a skill in strict mode.
func (v *Validator) ValidateStrict(meta *domain.Info, content string) *ValidationResult {
	ctx := &ValidationContext{
		FilePath: meta.FilePath,
		Content:  content,
		Lines:    strings.Split(content, "\n"),
		Meta:     meta,
		Strict:   true,
	}

	var issues []ValidationIssue
	for _, rule := range v.rules {
		issues = append(issues, rule(ctx)...)
	}

	return &ValidationResult{
		Issues: issues,
		Valid:  len(issues) == 0,
	}
}

// findLineNumber finds the line number of a pattern in content.
func findLineNumber(lines []string, pattern string) int {
	re := regexp.MustCompile(pattern)
	for i, line := range lines {
		if re.MatchString(line) {
			return i + 1
		}
	}
	return 0
}

// ValidateWithContext validates a skill with additional context (e.g., registry for redundancy detection).
func (v *Validator) ValidateWithContext(meta *domain.Info, content string, registry *Registry) *ValidationResult {
	return v.Validate(meta, content)
}
