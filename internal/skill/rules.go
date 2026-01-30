package skill

import (
	"fmt"
	"regexp"
	"strings"
)

var nameFormatRegexV4 = regexp.MustCompile(`^[a-z][a-z0-9-]{0,63}$`)

// validateFrontmatterV4 checks required v4 frontmatter fields (name, description, triggers).
func validateFrontmatterV4(ctx *ValidationContext) []ValidationIssue {
	var issues []ValidationIssue

	if ctx.Meta.Name == "" {
		issues = append(issues, ValidationIssue{
			Rule:       "frontmatter",
			Severity:   SeverityError,
			Message:    "missing required field: name",
			Suggestion: "Add a 'name' field to frontmatter",
			Example:    "---\nname: skill-name\n---",
			Line:       1,
		})
	}

	if ctx.Meta.Description == "" {
		issues = append(issues, ValidationIssue{
			Rule:       "frontmatter",
			Severity:   SeverityError,
			Message:    "missing required field: description",
			Suggestion: "Add a 'description' field explaining what the skill does",
			Example:    "Error handling patterns for Go",
			Line:       1,
		})
	}

	if len(ctx.Meta.Triggers) == 0 {
		issues = append(issues, ValidationIssue{
			Rule:       "frontmatter",
			Severity:   SeverityError,
			Message:    "missing required field: triggers",
			Suggestion: "Add at least one trigger phrase",
			Example:    "triggers:\n  - error handling",
			Line:       findLineNumber(ctx.Lines, "triggers:"),
		})
	}

	return issues
}

// validateNameFormatV4 checks that name field follows v4 format: ^[a-z][a-z0-9-]{0,63}$.
func validateNameFormatV4(ctx *ValidationContext) []ValidationIssue {
	if ctx.Meta.Name == "" {
		return nil
	}

	if !nameFormatRegexV4.MatchString(ctx.Meta.Name) {
		return []ValidationIssue{{
			Rule:       "name-format",
			Severity:   SeverityError,
			Message:    "invalid name format: must start with lowercase letter, followed by lowercase letters, numbers, or hyphens (max 64 chars)",
			Suggestion: "Use format: ^[a-z][a-z0-9-]{0,63}$",
			Example:    "valid-skill-name",
			Line:       findLineNumber(ctx.Lines, "name:"),
		}}
	}

	return nil
}

// validateRoleSectionV4 checks that ## Role section exists and has appropriate length.
func validateRoleSectionV4(ctx *ValidationContext) []ValidationIssue {
	if ctx.Meta.Role == "" {
		return []ValidationIssue{{
			Rule:       "role-section",
			Severity:   SeverityError,
			Message:    "missing required section: ## Role",
			Suggestion: "Add a ## Role section with persona definition",
			Example:    "## Role\n\nExpert Go engineer with focus on error handling.",
			Line:       findLineNumber(ctx.Lines, "## Role"),
		}}
	}

	roleLen := len(strings.TrimSpace(ctx.Meta.Role))
	if roleLen < 50 {
		return []ValidationIssue{{
			Rule:       "role-section",
			Severity:   SeverityError,
			Message:    fmt.Sprintf("## Role section too short: %d chars (minimum: 50)", roleLen),
			Suggestion: "Expand persona definition to at least 50 characters",
			Line:       findLineNumber(ctx.Lines, "## Role"),
		}}
	}

	if roleLen > 500 {
		return []ValidationIssue{{
			Rule:       "role-section",
			Severity:   SeverityWarning,
			Message:    fmt.Sprintf("## Role section too long: %d chars (recommended: 500)", roleLen),
			Suggestion: "Consider shortening the persona definition",
			Line:       findLineNumber(ctx.Lines, "## Role"),
		}}
	}

	return nil
}

// validateInstructionsSectionV4 checks that ## Instructions section exists and has appropriate content.
func validateInstructionsSectionV4(ctx *ValidationContext) []ValidationIssue {
	if ctx.Meta.Instructions == "" {
		return []ValidationIssue{{
			Rule:       "instructions-section",
			Severity:   SeverityError,
			Message:    "missing required section: ## Instructions",
			Suggestion: "Add a ## Instructions section with patterns and rules",
			Example:    "## Instructions\n\n### Pattern Name\n\nContent here...",
			Line:       findLineNumber(ctx.Lines, "## Instructions"),
		}}
	}

	instrLen := len(strings.TrimSpace(ctx.Meta.Instructions))
	if instrLen < 200 {
		return []ValidationIssue{{
			Rule:       "instructions-section",
			Severity:   SeverityError,
			Message:    fmt.Sprintf("## Instructions section too short: %d chars (minimum: 200)", instrLen),
			Suggestion: "Expand instructions to at least 200 characters",
			Line:       findLineNumber(ctx.Lines, "## Instructions"),
		}}
	}

	if instrLen > 10240 {
		return []ValidationIssue{{
			Rule:       "instructions-section",
			Severity:   SeverityWarning,
			Message:    fmt.Sprintf("## Instructions section too long: %d chars (recommended: 10KB)", instrLen),
			Suggestion: "Consider moving detailed patterns to references/ directory",
			Line:       findLineNumber(ctx.Lines, "## Instructions"),
		}}
	}

	hasSubsection := strings.Contains(ctx.Meta.Instructions, "### ")
	hasCodeBlock := strings.Contains(ctx.Meta.Instructions, "```")

	if !hasSubsection && !hasCodeBlock {
		return []ValidationIssue{{
			Rule:       "instructions-section",
			Severity:   SeverityWarning,
			Message:    "## Instructions section should contain subsections (###) or code blocks (```)",
			Suggestion: "Organize instructions with subsections or include code examples",
			Line:       findLineNumber(ctx.Lines, "## Instructions"),
		}}
	}

	return nil
}

// validateExamplesSectionV4 checks that ## Examples section exists with at least one example.
func validateExamplesSectionV4(ctx *ValidationContext) []ValidationIssue {
	if ctx.Meta.Examples == "" {
		return []ValidationIssue{{
			Rule:       "examples-section",
			Severity:   SeverityError,
			Message:    "missing required section: ## Examples",
			Suggestion: "Add a ## Examples section with Input/Output pairs",
			Example:    "## Examples\n\n### Example 1: Description\n\n**Input**: User request\n\n**Output**: Response",
			Line:       findLineNumber(ctx.Lines, "## Examples"),
		}}
	}

	hasExample := strings.Contains(ctx.Meta.Examples, "### Example")
	if !hasExample {
		return []ValidationIssue{{
			Rule:       "examples-section",
			Severity:   SeverityError,
			Message:    "## Examples section must contain at least one example (### Example N:)",
			Suggestion: "Add at least one example with ### Example 1: heading",
			Line:       findLineNumber(ctx.Lines, "## Examples"),
		}}
	}

	hasInput := strings.Contains(ctx.Meta.Examples, "**Input**")
	hasOutput := strings.Contains(ctx.Meta.Examples, "**Output**")

	if !hasInput || !hasOutput {
		return []ValidationIssue{{
			Rule:       "examples-section",
			Severity:   SeverityError,
			Message:    "Each example must have both **Input** and **Output**",
			Suggestion: "Format examples with: ### Example N: Description\n\n**Input**: ...\n\n**Output**: ...",
			Line:       findLineNumber(ctx.Lines, "## Examples"),
		}}
	}

	return nil
}

// validateReferencesV4 checks that references/ directory structure is valid if it exists.
// Returns warnings only (not errors).
func validateReferencesV4(ctx *ValidationContext) []ValidationIssue {
	if len(ctx.Meta.References) == 0 {
		return nil
	}

	for _, ref := range ctx.Meta.References {
		depth := strings.Count(ref, "/")
		if depth > 1 {
			return []ValidationIssue{{
				Rule:       "references",
				Severity:   SeverityWarning,
				Message:    fmt.Sprintf("reference path too deep: %s (max depth: 1)", ref),
				Suggestion: "Keep references/ directory flat or one level deep",
				Line:       findLineNumber(ctx.Lines, "## References"),
			}}
		}
	}

	return nil
}
