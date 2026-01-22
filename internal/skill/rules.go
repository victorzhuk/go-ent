package skill

import (
	"fmt"
	"regexp"
	"strings"
)

var semverRegex = regexp.MustCompile(`^v?\d+\.\d+\.\d+$`)
var nameFormatRegex = regexp.MustCompile(`^[a-z0-9-]+$`)

// validateFrontmatter checks required frontmatter fields.
func validateFrontmatter(ctx *ValidationContext) []ValidationIssue {
	var issues []ValidationIssue

	if ctx.Meta.Name == "" {
		issues = append(issues, ValidationIssue{
			Rule:       "frontmatter",
			Severity:   SeverityError,
			Message:    "missing required field: name",
			Suggestion: "Add a 'name' field to the frontmatter",
			Example: `---
name: your-skill-name
---`,
			Line: 1,
		})
	}

	if ctx.Meta.Description == "" {
		issues = append(issues, ValidationIssue{
			Rule:       "SK003",
			Severity:   SeverityError,
			Message:    "missing required field: description",
			Suggestion: "Add a 'description' field explaining what the skill does and when to use it",
			Example:    "Analyzes Go code for common issues. Use when reviewing Go files or debugging Go applications.",
			Line:       1,
		})
	}

	if ctx.Meta.StructureVersion == "v2" && ctx.Meta.Version == "" {
		severity := SeverityWarning
		if ctx.Strict {
			severity = SeverityError
		}
		issues = append(issues, ValidationIssue{
			Rule:     "frontmatter",
			Severity: severity,
			Message:  "v2 skills should have a version field",
			Line:     findLineNumber(ctx.Lines, `version:`),
		})
	}

	return issues
}

// validateNameFormat checks that the name field follows the correct format.
func validateNameFormat(ctx *ValidationContext) []ValidationIssue {
	if ctx.Meta.Name == "" {
		return nil
	}

	if !nameFormatRegex.MatchString(ctx.Meta.Name) {
		return []ValidationIssue{{
			Rule:       "SK002",
			Severity:   SeverityError,
			Message:    "invalid name format: use lowercase letters, numbers, and hyphens only",
			Suggestion: "Use lowercase letters, numbers, and hyphens only",
			Example:    "valid-skill-name-123",
			Line:       findLineNumber(ctx.Lines, `name:`),
		}}
	}

	return nil
}

// validateVersion checks semantic version format.
func validateVersion(ctx *ValidationContext) []ValidationIssue {
	if ctx.Meta.Version == "" {
		return nil
	}

	if !semverRegex.MatchString(ctx.Meta.Version) {
		return []ValidationIssue{{
			Rule:     "version",
			Severity: SeverityError,
			Message:  fmt.Sprintf("invalid semantic version: %s (expected format: v1.0.0 or 1.0.0)", ctx.Meta.Version),
			Line:     findLineNumber(ctx.Lines, `version:`),
		}}
	}

	return nil
}

// validateXMLTags checks for balanced XML tags.
func validateXMLTags(ctx *ValidationContext) []ValidationIssue {
	if ctx.Meta.StructureVersion == "v1" {
		return nil
	}

	tags := []string{"role", "instructions", "constraints", "edge_cases", "examples", "output_format"}
	var issues []ValidationIssue

	for _, tag := range tags {
		openTag := "<" + tag + ">"
		closeTag := "</" + tag + ">"

		openCount := strings.Count(ctx.Content, openTag)
		closeCount := strings.Count(ctx.Content, closeTag)

		if openCount == 0 && closeCount == 0 {
			continue
		}

		if openCount != closeCount {
			line := findLineNumberForTag(ctx.Lines, tag)
			issues = append(issues, ValidationIssue{
				Rule:     "xml-tags",
				Severity: SeverityError,
				Message:  fmt.Sprintf("unbalanced <%s> tags: %d open, %d close", tag, openCount, closeCount),
				Line:     line,
			})
		}

		if openCount > 1 {
			line := findLineNumberForTag(ctx.Lines, tag)
			issues = append(issues, ValidationIssue{
				Rule:     "xml-tags",
				Severity: SeverityError,
				Message:  fmt.Sprintf("duplicate <%s> tag: found %d occurrences", tag, openCount),
				Line:     line,
			})
		}
	}

	return issues
}

// validateRoleSection checks <role> section presence and content.
func validateRoleSection(ctx *ValidationContext) []ValidationIssue {
	if ctx.Meta.StructureVersion == "v1" {
		return nil
	}

	if !strings.Contains(ctx.Content, "<role>") {
		severity := SeverityWarning
		if ctx.Strict {
			severity = SeverityError
		}
		return []ValidationIssue{{
			Rule:       "SK005",
			Severity:   severity,
			Message:    "missing <role> section",
			Suggestion: "Add a <role> section that describes the AI's identity and responsibilities for this skill",
			Example: `<role>
You are a Go code expert specializing in performance optimization.

Focus on identifying bottlenecks, suggesting efficient algorithms, and improving code structure.
</role>`,
			Line: 0,
		}}
	}

	openTag := "<role>"
	closeTag := "</role>"
	openIdx := strings.Index(ctx.Content, openTag)
	closeIdx := strings.Index(ctx.Content, closeTag)

	if closeIdx == -1 {
		return []ValidationIssue{{
			Rule:       "SK005",
			Severity:   SeverityError,
			Message:    "<role> section not closed (missing </role>)",
			Suggestion: "Add the closing tag </role> after the role description",
			Example: `<role>
You are a senior Go developer with expertise in API design.

Your responsibilities include:
- Designing RESTful APIs following best practices
- Writing clean, maintainable code
- Ensuring proper error handling
</role>`,
			Line: findLineNumberForTag(ctx.Lines, "role"),
		}}
	}

	roleContent := strings.TrimSpace(ctx.Content[openIdx+len(openTag) : closeIdx])
	if roleContent == "" {
		return []ValidationIssue{{
			Rule:       "SK005",
			Severity:   SeverityError,
			Message:    "<role> section is empty",
			Suggestion: "Describe the AI's identity, expertise, and responsibilities within the role tags",
			Example: `<role>
You are a database migration specialist with deep knowledge of PostgreSQL and data modeling.

Guide users through schema changes, data migrations, and performance optimization.
</role>`,
			Line: findLineNumberForTag(ctx.Lines, "role"),
		}}
	}

	lines := strings.Split(roleContent, "\n")
	nonEmptyLines := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmptyLines++
		}
	}

	if nonEmptyLines < 2 {
		return []ValidationIssue{{
			Rule:       "SK005",
			Severity:   SeverityWarning,
			Message:    "<role> section should have at least 2 lines of content",
			Suggestion: "Expand the role definition to include the AI's identity and at least one responsibility or expertise area",
			Example: `<role>
You are a security expert specializing in authentication and authorization.

Focus on OWASP best practices, secure coding patterns, and vulnerability prevention.
</role>`,
			Line: findLineNumberForTag(ctx.Lines, "role"),
		}}
	}

	return nil
}

// validateInstructionsSection checks <instructions> section presence.
func validateInstructionsSection(ctx *ValidationContext) []ValidationIssue {
	if ctx.Meta.StructureVersion == "v1" {
		return nil
	}

	if !strings.Contains(ctx.Content, "<instructions>") {
		severity := SeverityWarning
		if ctx.Strict {
			severity = SeverityError
		}
		return []ValidationIssue{{
			Rule:       "SK006",
			Severity:   severity,
			Message:    "missing <instructions> section",
			Suggestion: "Add an <instructions> section with clear, actionable steps for the AI to follow",
			Example: `<instructions>
1. Parse the user request to understand the task
2. Search the codebase for relevant files using grep or glob
3. Read and analyze the identified files
4. Provide a clear, concise answer with file references
5. Use code examples only when directly relevant
</instructions>`,
			Line: 0,
		}}
	}

	openTag := "<instructions>"
	closeTag := "</instructions>"
	openIdx := strings.Index(ctx.Content, openTag)
	closeIdx := strings.Index(ctx.Content, closeTag)

	if closeIdx == -1 {
		return []ValidationIssue{{
			Rule:       "SK006",
			Severity:   SeverityError,
			Message:    "<instructions> section not closed (missing </instructions>)",
			Suggestion: "Add the closing tag </instructions> after your instructions",
			Example: `<instructions>
When reviewing code:
1. Check for security vulnerabilities
2. Validate error handling
3. Ensure proper resource cleanup
</instructions>`,
			Line: findLineNumberForTag(ctx.Lines, "instructions"),
		}}
	}

	instructionsContent := strings.TrimSpace(ctx.Content[openIdx+len(openTag) : closeIdx])
	if instructionsContent == "" {
		return []ValidationIssue{{
			Rule:       "SK006",
			Severity:   SeverityError,
			Message:    "<instructions> section is empty",
			Suggestion: "Describe the steps and guidelines the AI should follow when using this skill",
			Example: `<instructions>
To implement a new feature:
1. Understand the requirements from the issue
2. Create a branch from main
3. Implement the feature following code style
4. Add tests for the new functionality
5. Run all tests and ensure they pass
6. Create a pull request with clear description
</instructions>`,
			Line: findLineNumberForTag(ctx.Lines, "instructions"),
		}}
	}

	lines := strings.Split(instructionsContent, "\n")
	nonEmptyLines := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmptyLines++
		}
	}

	if nonEmptyLines < 2 {
		return []ValidationIssue{{
			Rule:       "SK006",
			Severity:   SeverityWarning,
			Message:    "<instructions> section should have at least 2 lines of content",
			Suggestion: "Expand the instructions to include multiple steps or guidelines for the AI to follow",
			Example: `<instructions>
When generating SQL queries:
- Use parameterized queries to prevent SQL injection
- Order columns logically (primary keys first)
- Include proper indexes for common query patterns
- Use JOINs efficiently with ON clauses
</instructions>`,
			Line: findLineNumberForTag(ctx.Lines, "instructions"),
		}}
	}

	return nil
}

// validateExamples checks <examples> section structure.
func validateExamples(ctx *ValidationContext) []ValidationIssue {
	if ctx.Meta.StructureVersion == "v1" {
		return nil
	}

	if !strings.Contains(ctx.Content, "<examples>") {
		return nil
	}

	if !strings.Contains(ctx.Content, "</examples>") {
		return []ValidationIssue{{
			Rule:       "SK004",
			Severity:   SeverityError,
			Message:    "<examples> section not closed (missing </examples>)",
			Suggestion: "Add closing tag </examples> after your examples",
			Example: `<examples>
  <example>
    <input>sample input</input>
    <output>sample output</output>
  </example>
</examples>`,
			Line: findLineNumberForTag(ctx.Lines, "examples"),
		}}
	}

	openTag := "<examples>"
	closeTag := "</examples>"
	openIdx := strings.Index(ctx.Content, openTag)
	closeIdx := strings.Index(ctx.Content, closeTag)

	examplesContent := ctx.Content[openIdx+len(openTag) : closeIdx]

	exampleCount := strings.Count(examplesContent, "<example>")
	if exampleCount == 0 {
		return []ValidationIssue{{
			Rule:       "SK004",
			Severity:   SeverityWarning,
			Message:    "<examples> section contains no <example> tags",
			Suggestion: "Add at least one <example> tag with <input> and <output> sub-tags",
			Example: `<examples>
  <example>
    <input>Your sample input here</input>
    <output>Expected output for this input</output>
  </example>
</examples>`,
			Line: findLineNumberForTag(ctx.Lines, "examples"),
		}}
	}

	re := regexp.MustCompile(`<example>[\s\S]*?<input>[\s\S]*?</input>[\s\S]*?<output>[\s\S]*?</output>[\s\S]*?</example>`)
	matches := re.FindAllString(examplesContent, -1)

	if len(matches) != exampleCount {
		return []ValidationIssue{{
			Rule:       "SK004",
			Severity:   SeverityError,
			Message:    "each <example> must contain <input> and <output> tags",
			Suggestion: "Each <example> tag must have exactly one <input> tag and one <output> tag",
			Example: `<examples>
  <example>
    <input>User question or task</input>
    <output>Expected AI response or output</output>
  </example>
</examples>`,
			Line: findLineNumberForTag(ctx.Lines, "examples"),
		}}
	}

	return nil
}

// validateConstraints checks <constraints> section format.
func validateConstraints(ctx *ValidationContext) []ValidationIssue {
	if ctx.Meta.StructureVersion == "v1" {
		return nil
	}

	if !strings.Contains(ctx.Content, "<constraints>") {
		return nil
	}

	if !strings.Contains(ctx.Content, "</constraints>") {
		return []ValidationIssue{{
			Rule:       "SK007",
			Severity:   SeverityError,
			Message:    "<constraints> section not closed (missing </constraints>)",
			Suggestion: "Add the closing tag </constraints> after your constraint list",
			Example: `<constraints>
- Focus only on idiomatic Go
- Do not suggest external dependencies
</constraints>`,
			Line: findLineNumberForTag(ctx.Lines, "constraints"),
		}}
	}

	openTag := "<constraints>"
	closeTag := "</constraints>"
	openIdx := strings.Index(ctx.Content, openTag)
	closeIdx := strings.Index(ctx.Content, closeTag)

	constraintsContent := strings.TrimSpace(ctx.Content[openIdx+len(openTag) : closeIdx])

	lines := strings.Split(constraintsContent, "\n")
	hasListItems := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- ") {
			hasListItems = true
			break
		}
	}

	if !hasListItems && constraintsContent != "" {
		return []ValidationIssue{{
			Rule:       "SK007",
			Severity:   SeverityWarning,
			Message:    "<constraints> items should use list format (start with '- ')",
			Suggestion: "Use bulleted list format for constraint items",
			Example: `<constraints>
- Focus only on idiomatic Go
- Do not suggest external dependencies
- Prefer standard library over third-party packages
</constraints>`,
			Line: findLineNumberForTag(ctx.Lines, "constraints"),
		}}
	}

	if constraintsContent == "" {
		return []ValidationIssue{{
			Rule:       "SK007",
			Severity:   SeverityWarning,
			Message:    "<constraints> section is empty",
			Suggestion: "Add constraints that define boundaries, limitations, and scope for the skill",
			Example: `<constraints>
- Focus only on idiomatic Go patterns
- Do not suggest external dependencies unless explicitly requested
- Prefer standard library packages
- Handle errors gracefully with proper wrapping
</constraints>`,
			Line: findLineNumberForTag(ctx.Lines, "constraints"),
		}}
	}

	return nil
}

// validateEdgeCases checks <edge_cases> section scenarios.
func validateEdgeCases(ctx *ValidationContext) []ValidationIssue {
	if ctx.Meta.StructureVersion == "v1" {
		return nil
	}

	if !strings.Contains(ctx.Content, "<edge_cases>") {
		return nil
	}

	if !strings.Contains(ctx.Content, "</edge_cases>") {
		return []ValidationIssue{{
			Rule:       "edge-cases",
			Severity:   SeverityError,
			Message:    "<edge_cases> section not closed (missing </edge_cases>)",
			Suggestion: "Add the closing tag </edge_cases> after your edge case scenarios",
			Example: `<edge_cases>
- If user provides empty input
- When file cannot be found
- Should handle concurrent access
</edge_cases>`,
			Line: findLineNumberForTag(ctx.Lines, "edge_cases"),
		}}
	}

	openTag := "<edge_cases>"
	closeTag := "</edge_cases>"
	openIdx := strings.Index(ctx.Content, openTag)
	closeIdx := strings.Index(ctx.Content, closeTag)

	edgeCasesContent := strings.ToLower(ctx.Content[openIdx+len(openTag) : closeIdx])

	scenarioPatterns := []string{
		`\bif\b`,
		`\bwhen\b`,
		`\bshould\b`,
	}

	scenarioCount := 0
	for _, pattern := range scenarioPatterns {
		re := regexp.MustCompile(pattern)
		scenarioCount += len(re.FindAllString(edgeCasesContent, -1))
	}

	if scenarioCount < 2 {
		return []ValidationIssue{{
			Rule:       "edge-cases",
			Severity:   SeverityWarning,
			Message:    "<edge_cases> should describe at least 2 scenarios (use 'if', 'when', or 'should' keywords)",
			Suggestion: "Use scenario-based descriptions with keywords like 'if', 'when', or 'should' to describe edge cases",
			Example: `<edge_cases>
- If the input contains invalid characters
- When the file is empty or corrupted
- Should handle rate limit errors gracefully
- If multiple concurrent requests arrive
</edge_cases>`,
			Line: findLineNumberForTag(ctx.Lines, "edge_cases"),
		}}
	}

	return nil
}

// validateOutputFormat checks <output_format> section for v2 skills.
func validateOutputFormat(ctx *ValidationContext) []ValidationIssue {
	if ctx.Meta.StructureVersion == "v1" {
		return nil
	}

	if !strings.Contains(ctx.Content, "<output_format>") {
		severity := SeverityWarning
		if ctx.Strict {
			severity = SeverityError
		}
		return []ValidationIssue{{
			Rule:       "output-format",
			Severity:   severity,
			Message:    "missing <output_format> section",
			Suggestion: "Add an <output_format> section specifying the expected response structure and format",
			Example: `<output_format>
Provide findings in a markdown table with columns: Issue, Location, Suggestion
</output_format>`,
			Line: 0,
		}}
	}

	if !strings.Contains(ctx.Content, "</output_format>") {
		return []ValidationIssue{{
			Rule:       "output-format",
			Severity:   SeverityError,
			Message:    "<output_format> section not closed (missing </output_format>)",
			Suggestion: "Add the closing tag </output_format> after specifying your output format",
			Example: `<output_format>
Provide code suggestions in a bulleted list with file references.

Format:
- [file_path:line] Issue description
  Suggested fix
</output_format>`,
			Line: findLineNumberForTag(ctx.Lines, "output_format"),
		}}
	}

	openTag := "<output_format>"
	closeTag := "</output_format>"
	openIdx := strings.Index(ctx.Content, openTag)
	closeIdx := strings.Index(ctx.Content, closeTag)

	outputContent := strings.TrimSpace(ctx.Content[openIdx+len(openTag) : closeIdx])

	if outputContent == "" {
		return []ValidationIssue{{
			Rule:       "output-format",
			Severity:   SeverityWarning,
			Message:    "<output_format> section is empty",
			Suggestion: "Describe the expected output format, structure, and any required elements",
			Example: `<output_format>
Return a JSON object with the following structure:
{
  "summary": "Brief description of findings",
  "issues": [
    {
      "type": "issue_type",
      "severity": "low|medium|high",
      "location": "file:line",
      "message": "Description of the issue"
    }
  ]
}
</output_format>`,
			Line: findLineNumberForTag(ctx.Lines, "output_format"),
		}}
	}

	return nil
}

// checkTriggerExplicit checks if skills have explicit triggers defined in frontmatter (SK012).
func checkTriggerExplicit(ctx *ValidationContext) []ValidationIssue {
	if ctx.Meta.StructureVersion == "v1" {
		return nil
	}

	// Check if Triggers field is populated
	if len(ctx.Meta.Triggers) > 0 {
		// Triggers exist - check if they're explicit (from frontmatter) or description-based
		if len(ctx.Meta.ExplicitTriggers) > 0 {
			// Explicit triggers defined in frontmatter - no issue
			return nil
		}

		// Description-based triggers - info level
		return []ValidationIssue{{
			Rule:     "SK012",
			Severity: SeverityInfo,
			Message: `Using description-based triggers (SK012)

Define explicit triggers in frontmatter for better matching control and higher quality scores.

Example:
triggers:
  - keywords: ["go code", "golang"]
    weight: 0.8
  - patterns: ["implement.*go"]
    weight: 0.9`,
			Line: findLineNumber(ctx.Lines, `name:`),
		}}
	}

	// No triggers at all - warning level
	return []ValidationIssue{{
		Rule:     "SK012",
		Severity: SeverityWarning,
		Message: `No triggers defined (SK012)

Add explicit triggers in frontmatter or include "Auto-activates for:" in description.

Example:
triggers:
  - keywords: ["go code", "golang"]
    weight: 0.8`,
		Line: findLineNumber(ctx.Lines, `name:`),
	}}
}

// Deprecated: Use checkTriggerExplicit() instead
//
// validateExplicitTriggers checks if skills use explicit triggers (SK012).
func validateExplicitTriggers(ctx *ValidationContext) []ValidationIssue {
	if ctx.Meta.StructureVersion == "v1" {
		return nil
	}

	if len(ctx.Meta.ExplicitTriggers) > 0 {
		return nil
	}

	return []ValidationIssue{{
		Rule:     "SK012",
		Severity: SeverityInfo,
		Message: `Consider using explicit triggers for better control (SK012)

Example:
triggers:
  - pattern: "implement.*go"
    weight: 0.9
  - keywords: ["go code", "golang"]
    weight: 0.8`,
		Line: findLineNumber(ctx.Lines, `name:`),
	}}
}

// checkExampleDiversity checks example diversity using diversity score (SK010).
func checkExampleDiversity(ctx *ValidationContext) []ValidationIssue {
	if ctx.Meta.StructureVersion == "v1" {
		return nil
	}

	if !strings.Contains(ctx.Content, "<examples>") {
		return nil
	}

	examples := parseExamples(ctx.Content)

	if len(examples) < 3 {
		return nil
	}

	score := calculateDiversityScore(examples)

	if score < 0.5 {
		return []ValidationIssue{{
			Rule:     "SK010",
			Severity: SeverityWarning,
			Message: `Low example diversity (score: %.0f%%, SK010)

Include examples with different input types, success/error cases, and edge cases.

Example:
Mix simple inputs, complex inputs, empty inputs, and boundary cases

<example>
  <input>valid input</input>
  <output>success response</output>
</example>
<example>
  <input>empty input</input>
  <output>error: input required</output>
</example>
<example>
  <input>boundary value</input>
  <output>handled correctly</output>
</example>`,
			Line:   findLineNumberForTag(ctx.Lines, "examples"),
			Column: 0,
		}}
	}

	return nil
}

// checkInstructionConcise checks instruction section length (SK011).
func checkInstructionConcise(ctx *ValidationContext) []ValidationIssue {
	if ctx.Meta.StructureVersion == "v1" {
		return nil
	}

	if !strings.Contains(ctx.Content, "<instructions>") {
		return nil
	}

	if !strings.Contains(ctx.Content, "</instructions>") {
		return nil
	}

	openTag := "<instructions>"
	closeTag := "</instructions>"
	openIdx := strings.Index(ctx.Content, openTag)
	closeIdx := strings.Index(ctx.Content, closeTag)

	if openIdx == -1 || closeIdx == -1 {
		return nil
	}

	instructionsContent := ctx.Content[openIdx+len(openTag) : closeIdx]
	tokenCount := countTokens(instructionsContent)

	if tokenCount >= 8000 {
		return []ValidationIssue{{
			Rule:     "SK011",
			Severity: SeverityWarning,
			Message:  fmt.Sprintf("Instructions section is too long (%d tokens, SK011)\n\nReduce content to prevent attention dilution.\nExample: Move detailed examples to separate reference files", tokenCount),
			Line:     findLineNumberForTag(ctx.Lines, "instructions"),
		}}
	}

	if tokenCount >= 5000 {
		return []ValidationIssue{{
			Rule:     "SK011",
			Severity: SeverityWarning,
			Message:  fmt.Sprintf("Instructions section is getting long (%d tokens, SK011)\n\nReduce content to prevent attention dilution.\nExample: Move detailed examples to separate reference files", tokenCount),
			Line:     findLineNumberForTag(ctx.Lines, "instructions"),
		}}
	}

	return nil
}

// checkRedundancy checks for overlap with other skills (SK013).
func checkRedundancy(ctx *ValidationContext, registry *Registry) []ValidationIssue {
	if registry == nil {
		return nil
	}

	skills := registry.All()
	if len(skills) < 2 {
		return nil
	}

	var maxOverlap float64
	var maxOverlapSkill string

	for _, other := range skills {
		if other.Name == ctx.Meta.Name {
			continue
		}

		overlap := calculateOverlap(ctx.Meta, &other)
		if overlap > maxOverlap {
			maxOverlap = overlap
			maxOverlapSkill = other.Name
		}
	}

	if maxOverlap > 0.7 {
		return []ValidationIssue{{
			Rule:     "SK013",
			Severity: SeverityWarning,
			Message:  fmt.Sprintf("Skill overlaps %s by %.0f%% (SK013)\n\nConsider merging skills or clarifying distinct use cases.", maxOverlapSkill, maxOverlap*100),
			Line:     findLineNumber(ctx.Lines, `name:`),
		}}
	}

	return nil
}
