package skill

import (
	"math"
	"regexp"
	"strings"
)

// QualityScore represents the complete quality assessment of a skill with
// component breakdown and overall score (0-100)
type QualityScore struct {
	Total       float64        // Overall score (0-100)
	Structure   StructureScore // Section presence and completeness
	Content     ContentScore   // Role clarity, instruction quality
	Examples    ExamplesScore  // Example count, diversity, format
	Triggers    float64        // Trigger presence and quality (0-15)
	Conciseness float64        // Token count penalty (0-15)
}

// Example represents a single parsed example from <examples> section.
type Example struct {
	Input  string
	Output string
}

// StructureScore evaluates required XML sections and their presence
type StructureScore struct {
	Total        float64 // Overall structure score (0-20)
	Role         float64 // <role> section present (0-4)
	Instructions float64 // <instructions> section present (0-4)
	Constraints  float64 // <constraints> section present (0-3)
	Examples     float64 // <examples> section present (0-3)
	OutputFormat float64 // <output_format> section present (0-3)
	EdgeCases    float64 // <edge_cases> section present (0-3)
}

// ContentScore evaluates content quality and instruction clarity
type ContentScore struct {
	Total        float64 // Overall content score (0-25)
	RoleClarity  float64 // Role expertise, domain, behavior (0-8)
	Instructions float64 // Actionability, specificity, structure (0-9)
	Constraints  float64 // Rule specificity and coverage (0-8)
}

// ExamplesScore evaluates example quality and coverage
type ExamplesScore struct {
	Total     float64 // Overall examples score (0-25)
	Count     float64 // Number of examples (0-10)
	Diversity float64 // Input/behavior variety (0-8)
	EdgeCases float64 // Edge case coverage (0-4)
	Format    float64 // Input/output + XML structure (0-3)
}

type QualityScorer struct{}

func NewQualityScorer() *QualityScorer {
	return &QualityScorer{}
}

// Score calculates complete quality assessment for a skill
func (s *QualityScorer) Score(meta *SkillMeta, content string) *QualityScore {
	result := &QualityScore{}

	frontmatterScore := s.scoreFrontmatter(meta)
	result.Total += frontmatterScore

	structure := s.calculateStructureScore(content)
	result.Structure = structure
	result.Total += structure.Total

	contentScore := s.calculateContentScore(content)
	result.Content = contentScore
	result.Total += contentScore.Total

	examples := s.scoreExamples(content)
	result.Examples = examples
	result.Total += examples.Total

	triggers := s.scoreTriggers(meta)
	result.Triggers = triggers
	result.Total += triggers

	conciseness := s.scoreConciseness(content)
	result.Conciseness = conciseness
	result.Total += conciseness

	return result
}

func (s *QualityScorer) calculateContentScore(content string) ContentScore {
	return ContentScore{
		Total:        math.Min(s.scoreRoleClarity(content)+s.scoreInstructions(content)+s.scoreConstraints(content), 25.0),
		RoleClarity:  s.scoreRoleClarity(content),
		Instructions: s.scoreInstructions(content),
		Constraints:  s.scoreConstraints(content),
	}
}

func (s *QualityScorer) scoreFrontmatter(meta *SkillMeta) float64 {
	score := 0.0

	if meta.Name != "" {
		score += 5.0
	}
	if meta.Description != "" {
		score += 5.0
	}
	if meta.Version != "" {
		score += 5.0
	}
	if len(meta.Tags) > 0 {
		score += 5.0
	}

	return score
}

func (s *QualityScorer) calculateStructureScore(content string) StructureScore {
	score := StructureScore{}

	if strings.Contains(content, "<role>") && strings.Contains(content, "</role>") {
		score.Role = 4.0
	}

	if strings.Contains(content, "<instructions>") && strings.Contains(content, "</instructions>") {
		score.Instructions = 4.0
	}

	if strings.Contains(content, "<constraints>") && strings.Contains(content, "</constraints>") {
		score.Constraints = 3.0
	}

	if strings.Contains(content, "<examples>") && strings.Contains(content, "</examples>") {
		score.Examples = 3.0
	}

	if strings.Contains(content, "<output_format>") && strings.Contains(content, "</output_format>") {
		score.OutputFormat = 3.0
	}

	if strings.Contains(content, "<edge_cases>") && strings.Contains(content, "</edge_cases>") {
		score.EdgeCases = 3.0
	}

	score.Total = score.Role + score.Instructions + score.Constraints +
		score.Examples + score.OutputFormat + score.EdgeCases

	return score
}

func (s *QualityScorer) scoreTriggers(meta *SkillMeta) float64 {
	if len(meta.ExplicitTriggers) > 0 {
		score := 0.0

		score += 10.0

		hasWeights := false
		for _, trigger := range meta.ExplicitTriggers {
			if trigger.Weight > 0 {
				hasWeights = true
				break
			}
		}
		if hasWeights {
			score += 3.0
		}

		hasKeywords := false
		hasPatterns := false
		hasFilePatterns := false

		for _, trigger := range meta.ExplicitTriggers {
			if len(trigger.Keywords) > 0 {
				hasKeywords = true
			}
			if len(trigger.Patterns) > 0 {
				hasPatterns = true
			}
			if len(trigger.FilePatterns) > 0 {
				hasFilePatterns = true
			}
		}

		diversityCount := 0
		if hasKeywords {
			diversityCount++
		}
		if hasPatterns {
			diversityCount++
		}
		if hasFilePatterns {
			diversityCount++
		}

		if diversityCount >= 2 {
			score += 2.0
		}

		return math.Min(score, 15.0)
	}

	if len(meta.Triggers) > 0 {
		return 5.0
	}

	return 0.0
}

func (s *QualityScorer) countTokens(content string) int {
	words := strings.Fields(content)
	wordCount := len(words)

	tokenCount := int(float64(wordCount) * 1.3)

	return tokenCount
}

func (s *QualityScorer) scoreConciseness(content string) float64 {
	tokenCount := s.countTokens(content)

	var score float64
	switch {
	case tokenCount < 3000:
		score = 15.0
	case tokenCount >= 3000 && tokenCount < 5000:
		score = 10.0
	case tokenCount >= 5000 && tokenCount < 8000:
		score = 5.0
	default:
		score = 0.0
	}

	return score
}

// scoreRoleClarity evaluates <role> section content quality
// Analyzes: expertise level, domain specificity, behavioral description
// Max 8 points
func (s *QualityScorer) scoreRoleClarity(content string) float64 {
	openIdx := strings.Index(content, "<role>")
	closeIdx := strings.Index(content, "</role>")

	if openIdx == -1 || closeIdx == -1 {
		return 0.0
	}

	roleContent := strings.TrimSpace(content[openIdx+6 : closeIdx])
	if roleContent == "" {
		return 0.0
	}

	score := 0.0

	expertiseKeywords := []string{"expert", "specialist", "architect", "engineer", "developer"}
	for _, kw := range expertiseKeywords {
		if strings.Contains(strings.ToLower(roleContent), kw) {
			score += 3.0
			break
		}
	}

	domainKeywords := []string{"go", "golang", "python", "rust", "api", "database", "security"}
	for _, kw := range domainKeywords {
		if strings.Contains(strings.ToLower(roleContent), kw) {
			score += 2.0
			break
		}
	}

	behaviorKeywords := []string{"focus", "prioritize", "specialize", "ensure", "implement"}
	for _, kw := range behaviorKeywords {
		if strings.Contains(strings.ToLower(roleContent), kw) {
			score += 3.0
			break
		}
	}

	return math.Min(score, 8.0)
}

// scoreInstructions evaluates <instructions> section quality
// Analyzes: actionability (doable tasks), specificity (clear criteria), structure (organized)
// Max 9 points
func (s *QualityScorer) scoreInstructions(content string) float64 {
	openIdx := strings.Index(content, "<instructions>")
	closeIdx := strings.Index(content, "</instructions>")

	if openIdx == -1 || closeIdx == -1 {
		return 0.0
	}

	instrContent := strings.TrimSpace(content[openIdx+13 : closeIdx])
	if instrContent == "" {
		return 0.0
	}

	score := 0.0

	actionVerbs := []string{"use", "implement", "create", "add", "define", "handle", "check", "verify"}
	hasAction := false
	for _, verb := range actionVerbs {
		if strings.Contains(strings.ToLower(instrContent), verb) {
			hasAction = true
			break
		}
	}
	if hasAction {
		score += 3.0
	}

	specificIndicators := []string{"for", "with", "when", "if", "ensure", "require", "must"}
	hasSpecificity := false
	for _, ind := range specificIndicators {
		if strings.Contains(strings.ToLower(instrContent), ind) {
			hasSpecificity = true
			break
		}
	}
	if hasSpecificity {
		score += 3.0
	}

	lines := strings.Split(instrContent, "\n")
	nonEmptyLines := 0
	headerCount := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			nonEmptyLines++
			if strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "##") {
				headerCount++
			}
		}
	}

	if nonEmptyLines >= 5 && headerCount >= 1 {
		score += 3.0
	}

	return math.Min(score, 9.0)
}

// scoreConstraints evaluates <constraints> section quality
// Analyzes: positive rules (include patterns), negative rules (exclude patterns), specificity
// Max 8 points
func (s *QualityScorer) scoreConstraints(content string) float64 {
	openIdx := strings.Index(content, "<constraints>")
	closeIdx := strings.Index(content, "</constraints>")

	if openIdx == -1 || closeIdx == -1 {
		return 0.0
	}

	constraintsContent := strings.TrimSpace(content[openIdx+12 : closeIdx])
	if constraintsContent == "" {
		return 0.0
	}

	score := 0.0
	lowerContent := strings.ToLower(constraintsContent)

	if strings.Contains(lowerContent, "include") || strings.Contains(lowerContent, "must") {
		score += 3.0
	}

	if strings.Contains(lowerContent, "exclude") || strings.Contains(lowerContent, "don't") ||
		strings.Contains(lowerContent, "avoid") || strings.Contains(lowerContent, "never") {
		score += 3.0
	}

	specificIndicators := []string{"bound to", "follow", "ensure", "verify", "use"}
	for _, ind := range specificIndicators {
		if strings.Contains(lowerContent, ind) {
			score += 2.0
			break
		}
	}

	return math.Min(score, 8.0)
}

func (s *QualityScorer) scoreExamples(content string) ExamplesScore {
	score := ExamplesScore{}

	openIdx := strings.Index(content, "<examples>")
	closeIdx := strings.Index(content, "</examples>")

	if openIdx == -1 || closeIdx == -1 {
		return score
	}

	examplesContent := content[openIdx+10 : closeIdx]

	exampleCount := strings.Count(examplesContent, "<example>")

	switch {
	case exampleCount == 0:
		score.Count = 0.0
	case exampleCount == 1:
		score.Count = 3.0
	case exampleCount == 2:
		score.Count = 6.0
	case exampleCount >= 3 && exampleCount <= 5:
		score.Count = 10.0
	case exampleCount > 5:
		score.Count = 8.0
	}

	if exampleCount > 0 {
		inputTypes := make(map[string]bool)

		re := regexp.MustCompile(`<input>([\s\S]*?)</input>`)
		matches := re.FindAllStringSubmatch(examplesContent, -1)

		for _, match := range matches {
			if len(match) > 1 {
				input := strings.ToLower(match[1])
				if strings.Contains(input, "go") || strings.Contains(input, "code") {
					inputTypes["code"] = true
				}
				if strings.Contains(input, "api") || strings.Contains(input, "http") {
					inputTypes["api"] = true
				}
				if strings.Contains(input, "database") || strings.Contains(input, "sql") {
					inputTypes["database"] = true
				}
				if strings.Contains(input, "config") || strings.Contains(input, "setup") {
					inputTypes["config"] = true
				}
			}
		}

		switch {
		case len(inputTypes) >= 3:
			score.Diversity = 8.0
		case len(inputTypes) >= 2:
			score.Diversity = 5.0
		case len(inputTypes) >= 1:
			score.Diversity = 2.0
		}
	}

	if strings.Contains(examplesContent, "<edge_cases>") {
		edgeCasesContent := strings.ToLower(examplesContent[strings.Index(examplesContent, "<edge_cases>"):strings.Index(examplesContent, "</edge_cases>")])

		edgeCaseIndicators := []string{"empty", "null", "error", "invalid", "timeout", "boundary", "zero", "negative"}
		edgeCaseCount := 0

		for _, indicator := range edgeCaseIndicators {
			if strings.Contains(edgeCasesContent, indicator) {
				edgeCaseCount++
			}
		}

		score.EdgeCases = math.Min(float64(edgeCaseCount)*2.0, 4.0)
	}

	if exampleCount > 0 {
		re := regexp.MustCompile(`<example>[\s\S]*?<input>[\s\S]*?</input>[\s\S]*?<output>[\s\S]*?</output>[\s\S]*?</example>`)
		validExamples := len(re.FindAllString(examplesContent, -1))

		if validExamples == exampleCount {
			score.Format = 3.0
		} else if validExamples > 0 {
			score.Format = 1.5
		}
	}

	score.Total = score.Count + score.Diversity + score.EdgeCases + score.Format

	return score
}

// parseExamples extracts all examples from the examples section.
func parseExamples(content string) []Example {
	re := regexp.MustCompile(`<example>([\s\S]*?)</example>`)
	matches := re.FindAllStringSubmatch(content, -1)

	examples := make([]Example, 0, len(matches))

	for _, match := range matches {
		if len(match) > 1 {
			exampleContent := match[1]

			inputRe := regexp.MustCompile(`<input>([\s\S]*?)</input>`)
			inputMatch := inputRe.FindStringSubmatch(exampleContent)

			outputRe := regexp.MustCompile(`<output>([\s\S]*?)</output>`)
			outputMatch := outputRe.FindStringSubmatch(exampleContent)

			if len(inputMatch) > 1 && len(outputMatch) > 1 {
				examples = append(examples, Example{
					Input:  strings.TrimSpace(inputMatch[1]),
					Output: strings.TrimSpace(outputMatch[1]),
				})
			}
		}
	}

	return examples
}

// calculateDiversityScore calculates diversity score (0.0-1.0) for examples.
// Checks behavior variety (success/error/edge cases) and data type variety.
func calculateDiversityScore(examples []Example) float64 {
	if len(examples) == 0 {
		return 0.0
	}

	behaviorScore := calculateBehaviorVariety(examples)
	dataTypeScore := calculateDataTypeVariety(examples)

	return (behaviorScore + dataTypeScore) / 2.0
}

// calculateBehaviorVariety checks for success, error, and edge case patterns.
func calculateBehaviorVariety(examples []Example) float64 {
	if len(examples) == 0 {
		return 0.0
	}

	hasSuccess := false
	hasError := false
	hasEdgeCase := false

	successKeywords := []string{"success", "correct", "valid", "pass", "work", "complete"}
	errorKeywords := []string{"error", "fail", "invalid", "wrong", "timeout", "reject", "denied"}
	edgeCaseKeywords := []string{"empty", "null", "zero", "negative", "boundary", "limit", "maximum", "minimum"}

	for _, ex := range examples {
		lowerInput := strings.ToLower(ex.Input)
		lowerOutput := strings.ToLower(ex.Output)

		if !hasSuccess {
			for _, kw := range successKeywords {
				if strings.Contains(lowerInput, kw) || strings.Contains(lowerOutput, kw) {
					hasSuccess = true
					break
				}
			}
		}

		if !hasError {
			for _, kw := range errorKeywords {
				if strings.Contains(lowerInput, kw) || strings.Contains(lowerOutput, kw) {
					hasError = true
					break
				}
			}
		}

		if !hasEdgeCase {
			for _, kw := range edgeCaseKeywords {
				if strings.Contains(lowerInput, kw) || strings.Contains(lowerOutput, kw) {
					hasEdgeCase = true
					break
				}
			}
		}
	}

	behaviorTypes := 0
	if hasSuccess {
		behaviorTypes++
	}
	if hasError {
		behaviorTypes++
	}
	if hasEdgeCase {
		behaviorTypes++
	}

	return float64(behaviorTypes) / 3.0
}

// calculateDataTypeVariety checks for different data types in inputs.
func calculateDataTypeVariety(examples []Example) float64 {
	if len(examples) == 0 {
		return 0.0
	}

	dataTypeKeywords := map[string][]string{
		"string":   {"\"", "'", "text", "message", "name", "description", "string"},
		"number":   {"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "float", "decimal"},
		"struct":   {"struct", "{", "}", "type", "json", "object"},
		"slice":    {"[]", "slice", "array", "list"},
		"map":      {"map[", "key", "dictionary"},
		"boolean":  {"true", "false", "bool"},
		"function": {"func", "function", "method", "()"},
		"api":      {"api", "http", "request", "response", "endpoint"},
		"database": {"sql", "query", "database", "db", "table", "insert", "select"},
	}

	foundTypes := make(map[string]bool)

	for _, ex := range examples {
		lowerInput := strings.ToLower(ex.Input)

		for dataType, keywords := range dataTypeKeywords {
			if !foundTypes[dataType] {
				for _, kw := range keywords {
					if strings.Contains(lowerInput, kw) {
						foundTypes[dataType] = true
						break
					}
				}
			}
		}
	}

	typeCount := len(foundTypes)
	if typeCount >= 4 {
		return 1.0
	}
	return float64(typeCount) / 4.0
}
