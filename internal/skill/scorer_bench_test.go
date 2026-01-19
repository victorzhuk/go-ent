package skill

import (
	"testing"
)

// BenchmarkQualityScorer_SingleSkill benchmarks scoring a single skill
func BenchmarkQualityScorer_SingleSkill(b *testing.B) {
	scorer := NewQualityScorer()

	content := `<role>Expert Go developer specializing in clean architecture</role>
<instructions>
Use repository pattern with private models and public entities.
Implement create, find, update, and delete operations.
Ensure proper error handling and context propagation.
</instructions>
<constraints>
- Include clean architecture patterns
- Exclude global state
- Follow SOLID principles
</constraints>
<examples>
<example>
<input>Create a repository</input>
<output>Repository struct with private models</output>
</example>
<example>
<input>Find user by ID</input>
<output>User entity from repository</output>
</example>
<example>
<input>Update user data</input>
<output>Updated entity with error handling</output>
</example>
</examples>
<output_format>
Return repository interface with concrete implementation
</output_format>
<edge_cases>
If skill is empty: ask clarifying questions
If database connection fails: return error with context
If user not found: return domain error
</edge_cases>`

	meta := &SkillMeta{
		Name:        "go-repo",
		Description: "Repository pattern for Go",
		Version:     "1.0.0",
		Tags:        []string{"go", "database"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = scorer.Score(meta, content)
	}
}

// BenchmarkQualityScorer_SmallSkill benchmarks scoring a minimal skill
func BenchmarkQualityScorer_SmallSkill(b *testing.B) {
	scorer := NewQualityScorer()

	content := `<role>Test</role>
<instructions>Test instructions</instructions>
<examples><example><input>Test</input><output>Test</output></example></examples>`

	meta := &SkillMeta{
		Name:        "test-skill",
		Description: "Test description",
		Version:     "1.0.0",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = scorer.Score(meta, content)
	}
}

// BenchmarkQualityScorer_VerboseSkill benchmarks scoring a long skill (>8k tokens)
func BenchmarkQualityScorer_VerboseSkill(b *testing.B) {
	scorer := NewQualityScorer()

	words := make([]string, 700)
	for i := range words {
		words[i] = "word"
	}
	content := `<role>Expert developer</role>
<instructions>` + joinWords(words) + `</instructions>
<constraints>Include all patterns</constraints>
<examples>` + joinWords(words[:300]) + `</examples>`

	meta := &SkillMeta{
		Name:        "verbose-skill",
		Description: "Verbose skill for testing",
		Version:     "1.0.0",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = scorer.Score(meta, content)
	}
}

func joinWords(words []string) string {
	result := ""
	for _, w := range words {
		result += w + " "
	}
	return result
}

// BenchmarkQualityScorer_Batch100 benchmarks scoring 100 skills
func BenchmarkQualityScorer_Batch100(b *testing.B) {
	scorer := NewQualityScorer()

	skills := createMockSkills(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, skill := range skills {
			_ = scorer.Score(skill.Meta, skill.Content)
		}
	}
}

func createMockSkills(count int) []MockSkill {
	skills := make([]MockSkill, count)

	content := `<role>Expert Go developer</role>
<instructions>Implement features</instructions>
<constraints>Follow patterns</constraints>
<examples>
<example><input>Test</input><output>Result</output></example>
</examples>`

	for i := 0; i < count; i++ {
		skills[i] = MockSkill{
			Meta: &SkillMeta{
				Name:        "test-skill",
				Description: "Test description",
				Version:     "1.0.0",
				Tags:        []string{"test"},
			},
			Content: content,
		}
	}

	return skills
}

type MockSkill struct {
	Meta    *SkillMeta
	Content string
}

// Benchmark individual scorers

func BenchmarkQualityScorer_calculateStructureScore(b *testing.B) {
	scorer := NewQualityScorer()
	content := `<role>Test</role>
<instructions>Test</instructions>
<constraints>Test</constraints>
<examples>Test</examples>
<output_format>Test</output_format>
<edge_cases>Test</edge_cases>`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = scorer.calculateStructureScore(content)
	}
}

func BenchmarkQualityScorer_scoreRoleClarity(b *testing.B) {
	scorer := NewQualityScorer()
	content := `<role>Expert Go developer specializing in clean architecture with domain-specific patterns and behavioral guidance</role>`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = scorer.scoreRoleClarity(content)
	}
}

func BenchmarkQualityScorer_scoreInstructions(b *testing.B) {
	scorer := NewQualityScorer()
	content := `<instructions>Implement repository pattern with private models and public entities. Ensure all methods return errors with context.</instructions>`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = scorer.scoreInstructions(content)
	}
}

func BenchmarkQualityScorer_scoreConstraints(b *testing.B) {
	scorer := NewQualityScorer()
	content := `<constraints>- Include clean architecture patterns
- Exclude global state
- Follow SOLID principles</constraints>`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = scorer.scoreConstraints(content)
	}
}

func BenchmarkQualityScorer_scoreExamples(b *testing.B) {
	scorer := NewQualityScorer()
	content := `<examples>
<example><input>Create repository</input><output>Repository struct</output></example>
<example><input>Find user</input><output>User entity</output></example>
<example><input>Update user</input><output>Updated entity</output></example>
</examples>`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = scorer.scoreExamples(content)
	}
}

func BenchmarkQualityScorer_scoreTriggers(b *testing.B) {
	scorer := NewQualityScorer()
	meta := &SkillMeta{
		ExplicitTriggers: []Trigger{
			{Keywords: []string{"go code"}, Weight: 0.8},
			{Patterns: []string{".*go.*code"}, FilePatterns: []string{"*.go"}, Weight: 0.7},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = scorer.scoreTriggers(meta)
	}
}

func BenchmarkQualityScorer_scoreConciseness(b *testing.B) {
	scorer := NewQualityScorer()

	words := make([]string, 3800)
	for i := range words {
		words[i] = "word"
	}
	content := joinWords(words)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = scorer.scoreConciseness(content)
	}
}

func BenchmarkQualityScorer_countTokens(b *testing.B) {
	scorer := NewQualityScorer()

	words := make([]string, 3800)
	for i := range words {
		words[i] = "word"
	}
	content := joinWords(words)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = scorer.countTokens(content)
	}
}
