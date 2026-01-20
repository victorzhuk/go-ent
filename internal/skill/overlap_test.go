package skill

import (
	"testing"
)

func TestCalculateOverlap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		skill1   *SkillMeta
		skill2   *SkillMeta
		expected float64
	}{
		{
			name: "identical skills",
			skill1: &SkillMeta{
				Name:        "skill1",
				Description: "Go code writing for modern applications",
				Triggers:    []string{"go code", "golang", "write go"},
			},
			skill2: &SkillMeta{
				Name:        "skill2",
				Description: "Go code writing for modern applications",
				Triggers:    []string{"go code", "golang", "write go"},
			},
			expected: 1.0,
		},
		{
			name: "different skills",
			skill1: &SkillMeta{
				Name:        "go-code",
				Description: "Go code writing for modern applications",
				Triggers:    []string{"go code", "golang"},
			},
			skill2: &SkillMeta{
				Name:        "python-code",
				Description: "Python data science machine learning",
				Triggers:    []string{"python code", "data science"},
			},
			expected: 0.0,
		},
		{
			name: "partial overlap",
			skill1: &SkillMeta{
				Name:        "go-code",
				Description: "Go code writing for modern applications",
				Triggers:    []string{"go code", "golang", "write go"},
			},
			skill2: &SkillMeta{
				Name:        "go-testing",
				Description: "Go testing for applications",
				Triggers:    []string{"go code", "testing", "go tests"},
			},
			expected: 0.41,
		},
		{
			name: "empty triggers",
			skill1: &SkillMeta{
				Name:        "skill1",
				Description: "First unique skill description",
				Triggers:    []string{},
			},
			skill2: &SkillMeta{
				Name:        "skill2",
				Description: "Second totally different text",
				Triggers:    []string{},
			},
			expected: 0.0,
		},
		{
			name: "one skill empty triggers",
			skill1: &SkillMeta{
				Name:        "skill1",
				Description: "First unique skill description",
				Triggers:    []string{"go code"},
			},
			skill2: &SkillMeta{
				Name:        "skill2",
				Description: "Second totally different text",
				Triggers:    []string{},
			},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := calculateOverlap(tt.skill1, tt.skill2)
			if result < tt.expected-0.01 || result > tt.expected+0.01 {
				t.Errorf("calculateOverlap() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCalculateTriggerOverlap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		skill1   *SkillMeta
		skill2   *SkillMeta
		expected float64
	}{
		{
			name: "identical triggers",
			skill1: &SkillMeta{
				Triggers: []string{"go code", "golang", "write go"},
			},
			skill2: &SkillMeta{
				Triggers: []string{"go code", "golang", "write go"},
			},
			expected: 1.0,
		},
		{
			name: "disjoint triggers",
			skill1: &SkillMeta{
				Triggers: []string{"go code", "golang"},
			},
			skill2: &SkillMeta{
				Triggers: []string{"python code", "data science"},
			},
			expected: 0.0,
		},
		{
			name: "partial overlap",
			skill1: &SkillMeta{
				Triggers: []string{"go code", "golang", "write go"},
			},
			skill2: &SkillMeta{
				Triggers: []string{"go code", "testing", "go tests"},
			},
			expected: 0.33,
		},
		{
			name: "case insensitive",
			skill1: &SkillMeta{
				Triggers: []string{"Go Code", "GOLANG"},
			},
			skill2: &SkillMeta{
				Triggers: []string{"go code", "golang"},
			},
			expected: 1.0,
		},
		{
			name: "empty triggers",
			skill1: &SkillMeta{
				Triggers: []string{},
			},
			skill2: &SkillMeta{
				Triggers: []string{},
			},
			expected: 0.0,
		},
		{
			name: "duplicate triggers",
			skill1: &SkillMeta{
				Triggers: []string{"go code", "go code", "golang"},
			},
			skill2: &SkillMeta{
				Triggers: []string{"go code", "golang"},
			},
			expected: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := calculateTriggerOverlap(tt.skill1, tt.skill2)
			if result < tt.expected-0.01 || result > tt.expected+0.01 {
				t.Errorf("calculateTriggerOverlap() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCalculateTextSimilarity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		text1    string
		text2    string
		expected float64
	}{
		{
			name:     "identical text",
			text1:    "Go code writing for modern applications",
			text2:    "Go code writing for modern applications",
			expected: 1.0,
		},
		{
			name:     "disjoint text",
			text1:    "Go code writing modern applications",
			text2:    "Python data science machine learning",
			expected: 0.0,
		},
		{
			name:     "partial overlap",
			text1:    "Go code writing for modern applications",
			text2:    "Go testing for applications and services",
			expected: 0.5,
		},
		{
			name:     "case insensitive",
			text1:    "Go Code Writing",
			text2:    "go code writing",
			expected: 1.0,
		},
		{
			name:     "empty text",
			text1:    "",
			text2:    "",
			expected: 0.0,
		},
		{
			name:     "punctuation ignored",
			text1:    "Go code, writing. For modern!",
			text2:    "Go code writing for modern",
			expected: 1.0,
		},
		{
			name:     "duplicates removed",
			text1:    "Go code writing Go code",
			text2:    "Go code writing",
			expected: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := calculateTextSimilarity(tt.text1, tt.text2)
			if result < tt.expected-0.01 || result > tt.expected+0.01 {
				t.Errorf("calculateTextSimilarity() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestTokenize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		text     string
		expected []string
	}{
		{
			name:     "simple text",
			text:     "Go code writing",
			expected: []string{"go", "code", "writing"},
		},
		{
			name:     "case insensitive",
			text:     "Go Code WRITING",
			expected: []string{"go", "code", "writing"},
		},
		{
			name:     "punctuation removed",
			text:     "Go code, writing. For modern!",
			expected: []string{"go", "code", "writing", "for", "modern"},
		},
		{
			name:     "duplicates removed",
			text:     "Go code writing Go code",
			expected: []string{"go", "code", "writing"},
		},
		{
			name:     "empty string",
			text:     "",
			expected: []string{},
		},
		{
			name:     "special characters",
			text:     "Go (code) \"writing\"",
			expected: []string{"go", "code", "writing"},
		},
		{
			name:     "dashes removed",
			text:     "Go code - writing",
			expected: []string{"go", "code", "writing"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tokenize(tt.text)
			if len(result) != len(tt.expected) {
				t.Errorf("tokenize() length = %d, want %d", len(result), len(tt.expected))
				return
			}
			for i, word := range result {
				if word != tt.expected[i] {
					t.Errorf("tokenize()[%d] = %v, want %v", i, word, tt.expected[i])
				}
			}
		})
	}
}
