package skill

import (
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

func TestQualityScorer_scoreStructure(t *testing.T) {
	s := NewQualityScorer()

	tests := []struct {
		name     string
		content  string
		expected float64
	}{
		{
			name:     "all sections present",
			content:  `<role>test</role>\n<instructions>test</instructions>\n<examples>test</examples>`,
			expected: 30.0,
		},
		{
			name:     "role and instructions only",
			content:  `<role>test</role>\n<instructions>test</instructions>`,
			expected: 20.0,
		},
		{
			name:     "role only",
			content:  `<role>test</role>`,
			expected: 10.0,
		},
		{
			name:     "no sections",
			content:  "Some text without tags",
			expected: 0.0,
		},
		{
			name:     "empty content",
			content:  "",
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.scoreStructure(tt.content)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestQualityScorer_scoreContent(t *testing.T) {
	s := NewQualityScorer()

	tests := []struct {
		name     string
		content  string
		expected float64
	}{
		{
			name:     "multiple examples and edge cases",
			content:  `<example>test1</example>\n<example>test2</example>\n<edge_cases>test</edge_cases>`,
			expected: 30.0,
		},
		{
			name:     "two examples only",
			content:  `<example>test1</example>\n<example>test2</example>`,
			expected: 15.0,
		},
		{
			name:     "one example only",
			content:  `<example>test</example>`,
			expected: 10.0,
		},
		{
			name:     "edge cases only",
			content:  `<edge_cases>test</edge_cases>`,
			expected: 15.0,
		},
		{
			name:     "no examples or edge cases",
			content:  "Some content",
			expected: 0.0,
		},
		{
			name:     "many examples and edge cases",
			content:  `<example>1</example>\n<example>2</example>\n<example>3</example>\n<edge_cases>test</edge_cases>`,
			expected: 30.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.scoreContent(tt.content)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestQualityScorer_scoreTriggers(t *testing.T) {
	s := NewQualityScorer()

	tests := []struct {
		name     string
		triggers []string
		expected float64
	}{
		{
			name:     "three triggers",
			triggers: []string{"writing tests", "tdd", "unit tests"},
			expected: 20.0,
		},
		{
			name:     "more than three triggers",
			triggers: []string{"a", "b", "c", "d"},
			expected: 20.0,
		},
		{
			name:     "two triggers",
			triggers: []string{"test", "debug"},
			expected: 13.34,
		},
		{
			name:     "one trigger",
			triggers: []string{"test"},
			expected: 6.67,
		},
		{
			name:     "no triggers",
			triggers: []string{},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := &SkillMeta{Triggers: tt.triggers}
			result := s.scoreTriggers(meta)
			assert.InDelta(t, tt.expected, result, 0.01)
		})
	}
}

func TestQualityScorer_Score(t *testing.T) {
	s := NewQualityScorer()

	tests := []struct {
		name     string
		meta     *SkillMeta
		content  string
		expected float64
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
<examples>
<example>test1</example>
<example>test2</example>
</examples>
<edge_cases>test</edge_cases>`,
			expected: 100.0,
		},
		{
			name: "v1 skill minimal",
			meta: &SkillMeta{
				Name:        "go-code",
				Description: "Go coding patterns",
				Triggers:    []string{"writing tests"},
			},
			content:  "Some content",
			expected: 16.67,
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
			expected: 43.34,
		},
		{
			name:     "empty meta and content",
			meta:     &SkillMeta{},
			content:  "",
			expected: 0.0,
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
			expected: 70.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.Score(tt.meta, tt.content)
			assert.InDelta(t, tt.expected, result, 0.01)
		})
	}
}

func TestQualityScorer_Score_Parallel(t *testing.T) {
	s := NewQualityScorer()

	tests := []struct {
		name     string
		meta     *SkillMeta
		content  string
		expected float64
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
<examples>
<example>ex1</example>
<example>ex2</example>
</examples>
<edge_cases>edge</edge_cases>`,
			expected: 100.0,
		},
		{
			name: "minimal skill",
			meta: &SkillMeta{
				Name:        "minimal",
				Description: "Min",
				Triggers:    []string{"test"},
			},
			content:  "",
			expected: 16.67,
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
			expected: 73.34,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := s.Score(tt.meta, tt.content)
			assert.InDelta(t, tt.expected, result, 0.01)
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

	assert.GreaterOrEqual(t, result, 0.0)
	assert.LessOrEqual(t, result, 100.0)
}
