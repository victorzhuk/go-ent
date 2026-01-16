package spec

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ValidationSeverity indicates the severity of a validation issue.
type ValidationSeverity string

const (
	SeverityError   ValidationSeverity = "error"
	SeverityWarning ValidationSeverity = "warning"
)

// ValidationIssue represents a validation error or warning.
type ValidationIssue struct {
	Severity ValidationSeverity
	File     string
	Line     int
	Message  string
	RuleID   string
}

func (v ValidationIssue) String() string {
	loc := v.File
	if v.Line > 0 {
		loc = fmt.Sprintf("%s:%d", v.File, v.Line)
	}
	return fmt.Sprintf("[%s] %s: %s (%s)", v.Severity, loc, v.Message, v.RuleID)
}

// ValidationResult holds the results of validation.
type ValidationResult struct {
	Issues  []ValidationIssue
	Valid   bool
	Summary string
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
	BasePath    string
	CurrentFile string
	Content     string
	Lines       []string
	Strict      bool
}

// ValidationRule is a function that validates content and returns issues.
type ValidationRule func(ctx *ValidationContext) []ValidationIssue

// Validator validates OpenSpec files using a set of rules.
type Validator struct {
	rules []ValidationRule
}

// NewValidator creates a new validator with default rules.
func NewValidator() *Validator {
	return &Validator{
		rules: []ValidationRule{
			validateRequirementHasScenario,
			validateScenarioFormat,
			validateDeltaOperations,
			validateRequirementFormat,
		},
	}
}

// ValidateSpec validates a single spec file.
func (v *Validator) ValidateSpec(path string, strict bool) (*ValidationResult, error) {
	content, err := os.ReadFile(path) // #nosec G304 -- controlled file path
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	ctx := &ValidationContext{
		BasePath:    filepath.Dir(path),
		CurrentFile: path,
		Content:     string(content),
		Lines:       strings.Split(string(content), "\n"),
		Strict:      strict,
	}

	var issues []ValidationIssue
	for _, rule := range v.rules {
		issues = append(issues, rule(ctx)...)
	}

	result := &ValidationResult{
		Issues: issues,
		Valid:  true,
	}

	// In strict mode, warnings also make validation fail
	if strict {
		result.Valid = len(issues) == 0
	} else {
		result.Valid = result.ErrorCount() == 0
	}

	result.Summary = v.buildSummary(result)
	return result, nil
}

// ValidateChange validates all files in a change directory.
func (v *Validator) ValidateChange(changePath string, strict bool) (*ValidationResult, error) {
	var allIssues []ValidationIssue

	// Check required files exist
	proposalPath := filepath.Join(changePath, "proposal.md")
	if _, err := os.Stat(proposalPath); os.IsNotExist(err) {
		allIssues = append(allIssues, ValidationIssue{
			Severity: SeverityError,
			File:     changePath,
			Message:  "Missing proposal.md",
			RuleID:   "change-structure",
		})
	}

	tasksPath := filepath.Join(changePath, "tasks.md")
	if _, err := os.Stat(tasksPath); os.IsNotExist(err) {
		allIssues = append(allIssues, ValidationIssue{
			Severity: SeverityError,
			File:     changePath,
			Message:  "Missing tasks.md",
			RuleID:   "change-structure",
		})
	}

	// Check for spec deltas
	specsDir := filepath.Join(changePath, "specs")
	if info, err := os.Stat(specsDir); err == nil && info.IsDir() {
		err := filepath.Walk(specsDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(path, ".md") {
				result, err := v.ValidateSpec(path, strict)
				if err != nil {
					allIssues = append(allIssues, ValidationIssue{
						Severity: SeverityError,
						File:     path,
						Message:  fmt.Sprintf("Failed to validate: %v", err),
						RuleID:   "file-read",
					})
				} else {
					allIssues = append(allIssues, result.Issues...)
				}
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("walk specs dir: %w", err)
		}
	} else {
		allIssues = append(allIssues, ValidationIssue{
			Severity: SeverityError,
			File:     changePath,
			Message:  "Missing specs/ directory - change must have at least one delta",
			RuleID:   "change-structure",
		})
	}

	result := &ValidationResult{
		Issues: allIssues,
		Valid:  true,
	}

	if strict {
		result.Valid = len(allIssues) == 0
	} else {
		result.Valid = result.ErrorCount() == 0
	}

	result.Summary = v.buildSummary(result)
	return result, nil
}

// ValidateAllSpecs validates all spec files in the specs directory.
func (v *Validator) ValidateAllSpecs(specsPath string, strict bool) (*ValidationResult, error) {
	var allIssues []ValidationIssue

	err := filepath.Walk(specsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".md") {
			result, err := v.ValidateSpec(path, strict)
			if err != nil {
				allIssues = append(allIssues, ValidationIssue{
					Severity: SeverityError,
					File:     path,
					Message:  fmt.Sprintf("Failed to validate: %v", err),
					RuleID:   "file-read",
				})
			} else {
				allIssues = append(allIssues, result.Issues...)
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk specs: %w", err)
	}

	result := &ValidationResult{
		Issues: allIssues,
		Valid:  true,
	}

	if strict {
		result.Valid = len(allIssues) == 0
	} else {
		result.Valid = result.ErrorCount() == 0
	}

	result.Summary = v.buildSummary(result)
	return result, nil
}

func (v *Validator) buildSummary(result *ValidationResult) string {
	errors := result.ErrorCount()
	warnings := result.WarningCount()

	if errors == 0 && warnings == 0 {
		return "✅ Validation passed with no issues"
	}

	var parts []string
	if errors > 0 {
		parts = append(parts, fmt.Sprintf("%d error(s)", errors))
	}
	if warnings > 0 {
		parts = append(parts, fmt.Sprintf("%d warning(s)", warnings))
	}

	status := "✅ Passed"
	if !result.Valid {
		status = "❌ Failed"
	}

	return fmt.Sprintf("%s: %s", status, strings.Join(parts, ", "))
}

// FindLineNumber finds the line number of a pattern in content.
// ParseRequirements extracts requirement names and their line numbers.
func parseRequirements(lines []string) map[string]int {
	result := make(map[string]int)
	re := regexp.MustCompile(`^###\s+Requirement:\s+(.+)$`)
	for i, line := range lines {
		if matches := re.FindStringSubmatch(strings.TrimSpace(line)); matches != nil {
			result[matches[1]] = i + 1
		}
	}
	return result
}

// ParseScenariosAfterLine returns the number of scenarios between start and end lines.
func parseScenariosAfterLine(lines []string, startLine, endLine int) int {
	count := 0
	re := regexp.MustCompile(`^####\s+Scenario:\s+`)
	for i := startLine; i < endLine && i < len(lines); i++ {
		if re.MatchString(strings.TrimSpace(lines[i])) {
			count++
		}
	}
	return count
}

// FindNextRequirementLine finds the next ### Requirement: line after startLine.
func findNextRequirementLine(lines []string, startLine int) int {
	re := regexp.MustCompile(`^###\s+Requirement:\s+`)
	for i := startLine; i < len(lines); i++ {
		if re.MatchString(strings.TrimSpace(lines[i])) {
			return i
		}
	}
	return len(lines)
}

// ReadFileLines reads a file and returns its lines.
func ReadFileLines(path string) ([]string, error) {
	file, err := os.Open(path) // #nosec G304 -- controlled file path
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
