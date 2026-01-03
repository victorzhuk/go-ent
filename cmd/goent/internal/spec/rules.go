package spec

import (
	"regexp"
	"strings"
)

// validateRequirementHasScenario checks that each requirement has at least one scenario.
func validateRequirementHasScenario(ctx *ValidationContext) []ValidationIssue {
	var issues []ValidationIssue

	requirements := parseRequirements(ctx.Lines)
	if len(requirements) == 0 {
		return issues
	}

	for reqName, reqLine := range requirements {
		// Find the next requirement or end of file
		endLine := findNextRequirementLine(ctx.Lines, reqLine)

		// Count scenarios between this requirement and the next
		scenarioCount := parseScenariosAfterLine(ctx.Lines, reqLine, endLine)

		if scenarioCount == 0 {
			issues = append(issues, ValidationIssue{
				Severity: SeverityError,
				File:     ctx.CurrentFile,
				Line:     reqLine,
				Message:  "Requirement must have at least one scenario: " + reqName,
				RuleID:   "requirement-needs-scenario",
			})
		}
	}

	return issues
}

// validateScenarioFormat checks that scenarios use the correct #### header format.
func validateScenarioFormat(ctx *ValidationContext) []ValidationIssue {
	var issues []ValidationIssue

	// Check for common incorrect scenario formats
	incorrectPatterns := []struct {
		pattern *regexp.Regexp
		message string
	}{
		{
			pattern: regexp.MustCompile(`(?m)^-\s+\*\*Scenario:`),
			message: "Scenario should use '#### Scenario:' header, not bullet with bold",
		},
		{
			pattern: regexp.MustCompile(`(?m)^\*\*Scenario\*\*:`),
			message: "Scenario should use '#### Scenario:' header, not bold text",
		},
		{
			pattern: regexp.MustCompile(`(?m)^###\s+Scenario:`),
			message: "Scenario should use '####' (4 hashtags), not '###' (3 hashtags)",
		},
		{
			pattern: regexp.MustCompile(`(?m)^#####\s+Scenario:`),
			message: "Scenario should use '####' (4 hashtags), not '#####' (5 hashtags)",
		},
	}

	for _, p := range incorrectPatterns {
		matches := p.pattern.FindAllStringIndex(ctx.Content, -1)
		for _, match := range matches {
			line := findLineAtOffset(ctx.Content, match[0])
			issues = append(issues, ValidationIssue{
				Severity: SeverityError,
				File:     ctx.CurrentFile,
				Line:     line,
				Message:  p.message,
				RuleID:   "scenario-format",
			})
		}
	}

	return issues
}

// validateDeltaOperations checks that delta operations are valid.
func validateDeltaOperations(ctx *ValidationContext) []ValidationIssue {
	var issues []ValidationIssue

	// Valid delta operations
	validOps := []string{"ADDED", "MODIFIED", "REMOVED", "RENAMED"}

	// Check for delta operation headers
	deltaPattern := regexp.MustCompile(`(?m)^##\s+(\w+)\s+Requirements?`)

	for i, line := range ctx.Lines {
		matches := deltaPattern.FindStringSubmatch(line)
		if matches != nil {
			op := matches[1]
			valid := false
			for _, validOp := range validOps {
				if strings.EqualFold(op, validOp) {
					valid = true
					break
				}
			}
			if !valid {
				issues = append(issues, ValidationIssue{
					Severity: SeverityError,
					File:     ctx.CurrentFile,
					Line:     i + 1,
					Message:  "Invalid delta operation: " + op + ". Must be ADDED, MODIFIED, REMOVED, or RENAMED",
					RuleID:   "delta-operation",
				})
			}
		}
	}

	return issues
}

// validateRequirementFormat checks that requirements use the correct ### header format.
func validateRequirementFormat(ctx *ValidationContext) []ValidationIssue {
	var issues []ValidationIssue

	// Check for common incorrect requirement formats
	incorrectPatterns := []struct {
		pattern *regexp.Regexp
		message string
	}{
		{
			pattern: regexp.MustCompile(`(?m)^-\s+\*\*Requirement:`),
			message: "Requirement should use '### Requirement:' header, not bullet with bold",
		},
		{
			pattern: regexp.MustCompile(`(?m)^##\s+Requirement:`),
			message: "Requirement should use '###' (3 hashtags), not '##' (2 hashtags)",
		},
		{
			pattern: regexp.MustCompile(`(?m)^####\s+Requirement:`),
			message: "Requirement should use '###' (3 hashtags), not '####' (4 hashtags)",
		},
	}

	for _, p := range incorrectPatterns {
		matches := p.pattern.FindAllStringIndex(ctx.Content, -1)
		for _, match := range matches {
			line := findLineAtOffset(ctx.Content, match[0])
			issues = append(issues, ValidationIssue{
				Severity: SeverityError,
				File:     ctx.CurrentFile,
				Line:     line,
				Message:  p.message,
				RuleID:   "requirement-format",
			})
		}
	}

	return issues
}

// findLineAtOffset converts a byte offset to a line number.
func findLineAtOffset(content string, offset int) int {
	line := 1
	for i := 0; i < offset && i < len(content); i++ {
		if content[i] == '\n' {
			line++
		}
	}
	return line
}
