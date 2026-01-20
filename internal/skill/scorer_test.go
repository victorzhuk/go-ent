package skill

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQualityScorer_scoreFrontmatter(t *testing.T) {
	s := NewQualityScorer()

	tests := []struct {
		name     string
		meta     *SkillMeta
		expected float64
	}{
		{
			name: "all fields present",
			meta: &SkillMeta{
				Name:        "go-code",
				Description: "Go coding patterns",
				Version:     "1.0.0",
				Tags:        []string{"go", "code"},
			},
			expected: 20.0,
		},
		{
			name: "name and description only",
			meta: &SkillMeta{
				Name:        "go-code",
				Description: "Go coding patterns",
			},
			expected: 10.0,
		},
		{
			name: "name only",
			meta: &SkillMeta{
				Name: "go-code",
			},
			expected: 5.0,
		},
		{
			name:     "no fields",
			meta:     &SkillMeta{},
			expected: 0.0,
		},
		{
			name: "with empty tags",
			meta: &SkillMeta{
				Name:        "go-code",
				Description: "Test",
				Version:     "1.0.0",
				Tags:        []string{},
			},
			expected: 15.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.scoreFrontmatter(tt.meta)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestQualityScorer_calculateStructureScore(t *testing.T) {
	s := NewQualityScorer()

	tests := []struct {
		name     string
		content  string
		expected StructureScore
	}{
		{
			name: "all sections present",
			content: `<role>test</role>
<instructions>test</instructions>
<constraints>test</constraints>
<examples>test</examples>
<output_format>test</output_format>
<edge_cases>test</edge_cases>`,
			expected: StructureScore{
				Total:        20.0,
				Role:         4.0,
				Instructions: 4.0,
				Constraints:  3.0,
				Examples:     3.0,
				OutputFormat: 3.0,
				EdgeCases:    3.0,
			},
		},
		{
			name:    "only role and instructions",
			content: `<role>test</role>\n<instructions>test</instructions>`,
			expected: StructureScore{
				Total:        8.0,
				Role:         4.0,
				Instructions: 4.0,
				Constraints:  0.0,
				Examples:     0.0,
				OutputFormat: 0.0,
				EdgeCases:    0.0,
			},
		},
		{
			name:    "no sections",
			content: `test content`,
			expected: StructureScore{
				Total:        0.0,
				Role:         0.0,
				Instructions: 0.0,
				Constraints:  0.0,
				Examples:     0.0,
				OutputFormat: 0.0,
				EdgeCases:    0.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := s.calculateStructureScore(tt.content)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestQualityScorer_scoreTriggers(t *testing.T) {
	s := NewQualityScorer()

	tests := []struct {
		name     string
		meta     *SkillMeta
		expected float64
	}{
		{
			name: "explicit triggers with weights and diversity",
			meta: &SkillMeta{
				ExplicitTriggers: []Trigger{
					{Keywords: []string{"go code"}, Weight: 0.8},
					{Patterns: []string{".*go.*code"}, FilePatterns: []string{"*.go"}, Weight: 0.7},
				},
			},
			expected: 15.0,
		},
		{
			name: "explicit triggers no weights",
			meta: &SkillMeta{
				ExplicitTriggers: []Trigger{
					{Keywords: []string{"go code"}, Weight: 0.8},
				},
			},
			expected: 13.0,
		},
		{
			name: "legacy description-based triggers",
			meta: &SkillMeta{
				Triggers: []string{"go code", "golang"},
			},
			expected: 5.0,
		},
		{
			name:     "no triggers",
			meta:     &SkillMeta{},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := s.scoreTriggers(tt.meta)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestQualityScorer_Score(t *testing.T) {
	s := NewQualityScorer()

	tests := []struct {
		name              string
		meta              *SkillMeta
		content           string
		expectedTotal     float64
		expectedStructure float64
	}{
		{
			name: "complete v2 skill",
			meta: &SkillMeta{
				Name:        "go-code",
				Description: "Go coding patterns",
				Version:     "1.0.0",
				Tags:        []string{"go", "code"},
				Triggers:    []string{"writing tests", "tdd", "unit tests"},
			},
			content: `<role>test</role>
<instructions>test</instructions>
<constraints>test</constraints>
<examples>
<example>test1</example>
<example>test2</example>
</examples>
<edge_cases>test</edge_cases>`,
			expectedTotal:     63.0,
			expectedStructure: 17.0,
		},
		{
			name: "v1 skill minimal",
			meta: &SkillMeta{
				Name:        "go-code",
				Description: "Go coding patterns",
				Triggers:    []string{"writing tests"},
			},
			content:           "Some content",
			expectedTotal:     30.0,
			expectedStructure: 0.0,
		},
		{
			name: "v1 skill with some structure",
			meta: &SkillMeta{
				Name:        "go-code",
				Description: "Test",
				Version:     "",
				Tags:        nil,
				Triggers:    []string{"test", "debug"},
			},
			content: `<role>test</role>
<instructions>test</instructions>`,
			expectedTotal:     38.0,
			expectedStructure: 8.0,
		},
		{
			name:              "empty meta and content",
			meta:              &SkillMeta{},
			content:           "",
			expectedTotal:     15.0,
			expectedStructure: 0.0,
		},
		{
			name: "partial v2 skill",
			meta: &SkillMeta{
				Name:        "test",
				Description: "test",
				Version:     "",
				Tags:        nil,
				Triggers:    []string{"a", "b", "c"},
			},
			content: `<role>test</role>
<instructions>test</instructions>
<examples>
<example>test</example>
</examples>`,
			expectedTotal:     44.0,
			expectedStructure: 11.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := s.Score(tt.meta, tt.content)
			assert.Equal(t, tt.expectedTotal, result.Total)
			assert.Equal(t, tt.expectedStructure, result.Structure.Total)
			assert.GreaterOrEqual(t, result.Total, 0.0)
			assert.LessOrEqual(t, result.Total, 120.0)
		})
	}
}

func TestQualityScorer_Score_Parallel(t *testing.T) {
	s := NewQualityScorer()

	tests := []struct {
		name          string
		meta          *SkillMeta
		content       string
		expectedTotal float64
	}{
		{
			name: "skill with all sections",
			meta: &SkillMeta{
				Name:        "complete",
				Description: "Complete skill",
				Version:     "2.0.0",
				Tags:        []string{"test"},
				Triggers:    []string{"a", "b", "c"},
			},
			content: `<role>role</role>
<instructions>inst</instructions>
<constraints>constraints</constraints>
<examples>
<example>ex1</example>
<example>ex2</example>
</examples>
<edge_cases>edge</edge_cases>`,
			expectedTotal: 63.0,
		},
		{
			name: "minimal skill",
			meta: &SkillMeta{
				Name:        "minimal",
				Description: "Min",
				Triggers:    []string{"test"},
			},
			content:       "",
			expectedTotal: 30.0,
		},
		{
			name: "medium quality",
			meta: &SkillMeta{
				Name:        "medium",
				Description: "Med",
				Version:     "1.0.0",
				Triggers:    []string{"a", "b"},
			},
			content: `<role>role</role>
<instructions>inst</instructions>
<example>ex</example>
<edge_cases>edge</edge_cases>`,
			expectedTotal: 46.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := s.Score(tt.meta, tt.content)
			assert.Equal(t, tt.expectedTotal, result.Total)
		})
	}
}

func TestNewQualityScorer(t *testing.T) {
	s := NewQualityScorer()
	assert.NotNil(t, s)
	assert.IsType(t, &QualityScorer{}, s)
}

func TestQualityScorer_Score_Range(t *testing.T) {
	s := NewQualityScorer()

	meta := &SkillMeta{
		Name:        "test",
		Description: "test",
		Version:     "1.0.0",
		Tags:        []string{"test"},
		Triggers:    []string{"a", "b", "c"},
	}

	content := `<role>test</role>
<instructions>test</instructions>
<examples>
<example>test1</example>
<example>test2</example>
</examples>
<edge_cases>test</edge_cases>`

	result := s.Score(meta, content)

	assert.GreaterOrEqual(t, result.Total, 0.0)
	assert.LessOrEqual(t, result.Total, 120.0)
}

func TestQualityScorer_scoreRoleClarity(t *testing.T) {
	s := NewQualityScorer()

	tests := []struct {
		name     string
		content  string
		expected float64
	}{
		{
			name:     "expertise and domain only",
			content:  `<role>Expert Go developer specializing in clean architecture and patterns</role>`,
			expected: 5.0,
		},
		{
			name:     "only expertise",
			content:  `<role>Expert Go developer</role>`,
			expected: 5.0,
		},
		{
			name:     "only behavior",
			content:  `<role>Consultant focusing on clean patterns</role>`,
			expected: 3.0,
		},
		{
			name:     "empty role",
			content:  `<role></role>`,
			expected: 0.0,
		},
		{
			name:     "no role section",
			content:  `test content`,
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := s.scoreRoleClarity(tt.content)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestQualityScorer_scoreInstructions(t *testing.T) {
	s := NewQualityScorer()

	tests := []struct {
		name     string
		content  string
		expected float64
	}{
		{
			name: "full instructions with action and specificity",
			content: `<instructions>
Use the repository pattern. Implement create, find, and update operations when handling data access.
Ensure all methods return errors properly.

# Pattern Implementation
Define clear interfaces
Ensure proper error handling
Verify all operations
</instructions>`,
			expected: 9.0,
		},
		{
			name:     "only actions",
			content:  `<instructions>Implement the following patterns. Create and update.</instructions>`,
			expected: 3.0,
		},
		{
			name:     "only specificity",
			content:  `<instructions>For data access patterns, ensure proper error handling when performing operations.</instructions>`,
			expected: 3.0,
		},
		{
			name:     "empty instructions",
			content:  `<instructions></instructions>`,
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := s.scoreInstructions(tt.content)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestQualityScorer_scoreConstraints(t *testing.T) {
	s := NewQualityScorer()

	tests := []struct {
		name     string
		content  string
		expected float64
	}{
		{
			name: "full constraints",
			content: `<constraints>
- Include clean architecture patterns
- Exclude verbose comments
- Follow SOLID principles
- Ensure proper error handling
</constraints>`,
			expected: 8.0,
		},
		{
			name: "only positive rules",
			content: `<constraints>
- Include repository pattern
- Follow DDD principles
</constraints>`,
			expected: 5.0,
		},
		{
			name:     "no constraints",
			content:  `test content`,
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := s.scoreConstraints(tt.content)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestQualityScorer_scoreExamples(t *testing.T) {
	s := NewQualityScorer()

	tests := []struct {
		name     string
		content  string
		expected ExamplesScore
	}{
		{
			name: "optimal 3-5 diverse examples",
			content: `<examples>
<example>
<input>Refactor go code</input>
<output>Updated with graceful shutdown</output>
</example>
<example>
<input>Implement API endpoints</input>
<output>Repository with models</output>
</example>
<example>
<input>Add database config</input>
<output>Config with env loading</output>
</example>
</examples>`,
			expected: ExamplesScore{
				Total:     21.0,
				Count:     10.0,
				Diversity: 8.0,
				EdgeCases: 0.0,
				Format:    3.0,
			},
		},
		{
			name: "minimal 1 example",
			content: `<examples>
<example>
<input>Test input</input>
<output>Test output</output>
</example>
</examples>`,
			expected: ExamplesScore{
				Total:     6.0,
				Count:     3.0,
				Diversity: 0.0,
				EdgeCases: 0.0,
				Format:    3.0,
			},
		},
		{
			name:    "no examples",
			content: `test content`,
			expected: ExamplesScore{
				Total:     0.0,
				Count:     0.0,
				Diversity: 0.0,
				EdgeCases: 0.0,
				Format:    0.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := s.scoreExamples(tt.content)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestQualityScorer_scoreConciseness(t *testing.T) {
	s := NewQualityScorer()

	tests := []struct {
		name     string
		content  string
		expected float64
	}{
		{
			name:     "optimal length <3k tokens",
			content:  strings.Repeat("word ", 2000),
			expected: 15.0,
		},
		{
			name:     "acceptable 3-5k tokens",
			content:  strings.Repeat("word ", 3000),
			expected: 10.0,
		},
		{
			name:     "verbose 5-8k tokens",
			content:  strings.Repeat("word ", 5000),
			expected: 5.0,
		},
		{
			name:     "too verbose >8k tokens",
			content:  strings.Repeat("word ", 7000),
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := s.scoreConciseness(tt.content)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestQualityScorer_Score_Complete(t *testing.T) {
	s := NewQualityScorer()

	meta := &SkillMeta{
		Name:        "test-skill",
		Description: "Test skill description",
		Version:     "1.0.0",
		Tags:        []string{"test"},
	}

	content := `<role>Expert Go developer implementing clean architecture</role>
<instructions>
Use clean patterns. Implement repository pattern.
Ensure proper error handling for all operations.
</instructions>
<constraints>- Include clean code
- Exclude verbose comments
- Follow SOLID principles</constraints>
<examples>
<example>
<input>Test case</input>
<output>Test result</output>
</example>
</examples>
<output_format>Clear output format</output_format>
<edge_cases>Handle edge cases</edge_cases>`

	result := s.Score(meta, content)

	assert.Greater(t, result.Total, 70.0)
	assert.Equal(t, 20.0, result.Structure.Total)
	assert.Equal(t, 0.0, result.Triggers)
	assert.Greater(t, result.Content.Total, 10.0)
	assert.Greater(t, result.Examples.Total, 5.0)
	assert.Greater(t, result.Conciseness, 10.0)
}

func TestParseExamples(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []Example
	}{
		{
			name: "single example",
			content: `<example>
<input>test input</input>
<output>test output</output>
</example>`,
			expected: []Example{
				{Input: "test input", Output: "test output"},
			},
		},
		{
			name: "multiple examples",
			content: `<example>
<input>input1</input>
<output>output1</output>
</example>
<example>
<input>input2</input>
<output>output2</output>
</example>`,
			expected: []Example{
				{Input: "input1", Output: "output1"},
				{Input: "input2", Output: "output2"},
			},
		},
		{
			name:     "empty content",
			content:  "",
			expected: []Example{},
		},
		{
			name: "malformed example (missing output)",
			content: `<example>
<input>input</input>
</example>`,
			expected: []Example{},
		},
		{
			name: "example with whitespace",
			content: `<example>

<input>  test input  </input>

<output>  test output  </output>

</example>`,
			expected: []Example{
				{Input: "test input", Output: "test output"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseExamples(tt.content)
			assert.Equal(t, len(tt.expected), len(result), "count mismatch")
			for i, ex := range result {
				assert.Equal(t, tt.expected[i].Input, ex.Input, "input mismatch at index %d", i)
				assert.Equal(t, tt.expected[i].Output, ex.Output, "output mismatch at index %d", i)
			}
		})
	}
}

func TestCalculateDiversityScore(t *testing.T) {
	tests := []struct {
		name     string
		examples []Example
		expected float64
	}{
		{
			name:     "empty examples",
			examples: []Example{},
			expected: 0.0,
		},
		{
			name: "single example",
			examples: []Example{
				{Input: "test input", Output: "test output"},
			},
			expected: 0.0,
		},
		{
			name: "high variety examples",
			examples: []Example{
				{Input: "success case with struct", Output: "returns valid data"},
				{Input: "error case with string", Output: "returns error message"},
				{Input: "empty null case", Output: "handles zero value"},
				{Input: "api request with map", Output: "json response"},
			},
			expected: 0.8750,
		},
		{
			name: "medium variety examples",
			examples: []Example{
				{Input: "valid string input", Output: "success"},
				{Input: "error case", Output: "invalid"},
			},
			expected: 0.4583,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateDiversityScore(tt.examples)
			assert.InDelta(t, tt.expected, result, 0.01, "score mismatch")
		})
	}
}

func TestCalculateBehaviorVariety(t *testing.T) {
	tests := []struct {
		name     string
		examples []Example
		expected float64
	}{
		{
			name:     "empty examples",
			examples: []Example{},
			expected: 0.0,
		},
		{
			name: "only success cases",
			examples: []Example{
				{Input: "valid input", Output: "success result"},
			},
			expected: 1.0 / 3.0,
		},
		{
			name: "success and error",
			examples: []Example{
				{Input: "valid input", Output: "success"},
				{Input: "invalid input", Output: "error"},
			},
			expected: 2.0 / 3.0,
		},
		{
			name: "all behavior types",
			examples: []Example{
				{Input: "valid input", Output: "success completes"},
				{Input: "invalid input", Output: "error fails"},
				{Input: "empty input", Output: "handles boundary"},
			},
			expected: 1.0,
		},
		{
			name: "edge cases only",
			examples: []Example{
				{Input: "zero value", Output: "handles null"},
				{Input: "negative limit", Output: "boundary check"},
			},
			expected: 1.0 / 3.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateBehaviorVariety(tt.examples)
			assert.InDelta(t, tt.expected, result, 0.01, "score mismatch")
		})
	}
}

func TestCalculateDataTypeVariety(t *testing.T) {
	tests := []struct {
		name     string
		examples []Example
		expected float64
	}{
		{
			name:     "empty examples",
			examples: []Example{},
			expected: 0.0,
		},
		{
			name: "single data type",
			examples: []Example{
				{Input: "test message string", Output: "result"},
			},
			expected: 0.25,
		},
		{
			name: "two data types",
			examples: []Example{
				{Input: "text message", Output: "result"},
				{Input: "int value 42", Output: "number"},
			},
			expected: 0.5,
		},
		{
			name: "four or more data types",
			examples: []Example{
				{Input: "string text", Output: "result"},
				{Input: "int 42", Output: "number"},
				{Input: "struct type { }", Output: "json"},
				{Input: "[] slice array", Output: "list"},
			},
			expected: 1.0,
		},
		{
			name: "api and database types",
			examples: []Example{
				{Input: "http request", Output: "response"},
				{Input: "sql query", Output: "database result"},
				{Input: "func method()", Output: "return"},
				{Input: "true false", Output: "bool"},
			},
			expected: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateDataTypeVariety(tt.examples)
			assert.InDelta(t, tt.expected, result, 0.01, "score mismatch")
		})
	}
}
