package generator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInlinePrompts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		prompts  *PromptContent
		agent    *AgentSource
		validate func(t *testing.T, result string)
	}{
		{
			name: "inlines main prompt only",
			prompts: &PromptContent{
				Main:   "You are a helpful assistant.",
				Shared: map[string]string{},
			},
			agent: &AgentSource{
				Prompts: PromptsConfig{},
			},
			validate: func(t *testing.T, result string) {
				assert.Contains(t, result, "You are a helpful assistant.")
			},
		},
		{
			name: "inlines main with shared prompts",
			prompts: &PromptContent{
				Main: "Main prompt content.",
				Shared: map[string]string{
					"coding":  "Coding guidelines here.",
					"testing": "Testing guidelines here.",
				},
			},
			agent: &AgentSource{
				Prompts: PromptsConfig{
					Shared: []string{"coding", "testing"},
				},
			},
			validate: func(t *testing.T, result string) {
				assert.Contains(t, result, "Main prompt content.")
				assert.Contains(t, result, "## Coding")
				assert.Contains(t, result, "Coding guidelines here.")
				assert.Contains(t, result, "## Testing")
				assert.Contains(t, result, "Testing guidelines here.")
			},
		},
		{
			name: "orders shared prompts as specified",
			prompts: &PromptContent{
				Main: "Main.",
				Shared: map[string]string{
					"first":  "First shared.",
					"second": "Second shared.",
					"third":  "Third shared.",
				},
			},
			agent: &AgentSource{
				Prompts: PromptsConfig{
					Shared: []string{"third", "first", "second"},
				},
			},
			validate: func(t *testing.T, result string) {
				thirdIdx := assertIndex(t, result, "## Third")
				firstIdx := assertIndex(t, result, "## First")
				secondIdx := assertIndex(t, result, "## Second")

				assert.Less(t, thirdIdx, firstIdx, "third should come before first")
				assert.Less(t, firstIdx, secondIdx, "first should come before second")
			},
		},
		{
			name: "skips missing shared prompts",
			prompts: &PromptContent{
				Main: "Main.",
				Shared: map[string]string{
					"existing": "This exists.",
				},
			},
			agent: &AgentSource{
				Prompts: PromptsConfig{
					Shared: []string{"existing", "nonexistent"},
				},
			},
			validate: func(t *testing.T, result string) {
				assert.Contains(t, result, "This exists.")
				assert.NotContains(t, result, "nonexistent")
			},
		},
		{
			name: "handles empty main prompt",
			prompts: &PromptContent{
				Main:   "",
				Shared: map[string]string{},
			},
			agent: &AgentSource{
				Prompts: PromptsConfig{},
			},
			validate: func(t *testing.T, result string) {
				assert.NotPanics(t, func() {
					_ = InlinePrompts(&PromptContent{Main: "", Shared: map[string]string{}}, &AgentSource{})
				})
			},
		},
		{
			name: "capitalizes shared prompt headers",
			prompts: &PromptContent{
				Main: "Main.",
				Shared: map[string]string{
					"lowercase": "Lower content.",
					"UPPERCASE": "Upper content.",
					"mixedCase": "Mixed content.",
				},
			},
			agent: &AgentSource{
				Prompts: PromptsConfig{
					Shared: []string{"lowercase", "UPPERCASE", "mixedCase"},
				},
			},
			validate: func(t *testing.T, result string) {
				assert.Contains(t, result, "## Lowercase")
				assert.Contains(t, result, "## UPPERCASE")
				assert.Contains(t, result, "## MixedCase")
			},
		},
		{
			name: "trims whitespace from result",
			prompts: &PromptContent{
				Main:   "Content with spaces.   ",
				Shared: map[string]string{},
			},
			agent: &AgentSource{
				Prompts: PromptsConfig{},
			},
			validate: func(t *testing.T, result string) {
				assert.Equal(t, "Content with spaces.", result)
			},
		},
		{
			name: "handles multiline content",
			prompts: &PromptContent{
				Main: "Main prompt\nwith multiple\nlines.",
				Shared: map[string]string{
					"coding": "Line 1\nLine 2\nLine 3",
				},
			},
			agent: &AgentSource{
				Prompts: PromptsConfig{
					Shared: []string{"coding"},
				},
			},
			validate: func(t *testing.T, result string) {
				assert.Contains(t, result, "Main prompt\nwith multiple\nlines.")
				assert.Contains(t, result, "Line 1\nLine 2\nLine 3")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := InlinePrompts(tt.prompts, tt.agent)

			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}

func TestInlinePrompts_Structure(t *testing.T) {
	t.Parallel()

	prompts := &PromptContent{
		Main: "Main content.",
		Shared: map[string]string{
			"first":  "First shared.",
			"second": "Second shared.",
		},
	}

	agent := &AgentSource{
		Prompts: PromptsConfig{
			Shared: []string{"first", "second"},
		},
	}

	result := InlinePrompts(prompts, agent)

	assert.Contains(t, result, "Main content.")
	assert.Contains(t, result, "## First")
	assert.Contains(t, result, "First shared.")
	assert.Contains(t, result, "## Second")
	assert.Contains(t, result, "Second shared.")
}

func assertIndex(t testing.TB, s, substr string) int {
	t.Helper()
	idx := len(s)
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			idx = i
			break
		}
	}
	return idx
}
