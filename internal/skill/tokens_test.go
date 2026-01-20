package skill

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCountTokens(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantWords int
	}{
		{
			name:      "empty string",
			input:     "",
			wantWords: 0,
		},
		{
			name:      "whitespace only",
			input:     "   \t\n  ",
			wantWords: 0,
		},
		{
			name:      "single word",
			input:     "hello",
			wantWords: 1,
		},
		{
			name:      "multiple words",
			input:     "hello world foo bar",
			wantWords: 4,
		},
		{
			name:      "words with multiple spaces",
			input:     "hello    world   foo",
			wantWords: 3,
		},
		{
			name:      "words with tabs and newlines",
			input:     "hello\tworld\nfoo\nbar",
			wantWords: 4,
		},
		{
			name:      "XML tags content",
			input:     "<role>You are a Go developer</role>",
			wantWords: 5,
		},
		{
			name:      "markdown content",
			input:     "# Heading\n\nThis is a paragraph with some **bold** text.",
			wantWords: 10,
		},
		{
			name:      "short skill body",
			input:     "Implement a function to validate user input. Return error if invalid.",
			wantWords: 11,
		},
		{
			name:      "medium skill body",
			input:     strings.Repeat("word ", 100),
			wantWords: 100,
		},
		{
			name:      "long skill body",
			input:     strings.Repeat("word ", 1000),
			wantWords: 1000,
		},
		{
			name: "XML with instructions",
			input: `<instructions>
Create a new user endpoint that accepts JSON input.
Validate required fields: name, email, age.
Return 201 on success, 400 on validation error.</instructions>`,
			wantWords: 24,
		},
		{
			name: "mixed format with examples",
			input: `<examples>
<example>
<input>valid data</input>
<output>success</output>
</example>
<example>
<input>invalid data</input>
<output>error</output>
</example>
</examples>`,
			wantWords: 12,
		},
		{
			name: "complex skill with all sections",
			input: `<role>You are a database expert specializing in PostgreSQL.</role>
<instructions>
Design schemas, optimize queries, handle migrations.
Use prepared statements to prevent SQL injection.
Follow best practices for connection pooling.</instructions>
<constraints>- Never expose credentials
- Always use parameterized queries
- Validate all user input</constraints>`,
			wantWords: 42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := countTokens(tt.input)
			wantTokens := int(float64(tt.wantWords) * 1.3)

			assert.Equal(t, wantTokens, got, "token count should equal words * 1.3")
		})
	}
}

func TestCountTokens_Accuracy(t *testing.T) {
	tests := []struct {
		name         string
		wordCount    int
		tolerancePct float64
	}{
		{
			name:         "short text",
			wordCount:    10,
			tolerancePct: 0.1,
		},
		{
			name:         "medium text",
			wordCount:    100,
			tolerancePct: 0.1,
		},
		{
			name:         "long text",
			wordCount:    1000,
			tolerancePct: 0.1,
		},
		{
			name:         "very long text",
			wordCount:    5000,
			tolerancePct: 0.1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			text := strings.Repeat("word ", tt.wordCount)
			got := countTokens(text)
			expected := int(float64(tt.wordCount) * 1.3)

			minExpected := int(float64(expected) * (1 - tt.tolerancePct))
			maxExpected := int(float64(expected) * (1 + tt.tolerancePct))

			assert.GreaterOrEqual(t, got, minExpected, "token count should be within -%d%% of expected", int(tt.tolerancePct*100))
			assert.LessOrEqual(t, got, maxExpected, "token count should be within +%d%% of expected", int(tt.tolerancePct*100))
		})
	}
}

func TestCountTokens_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{
			name:  "nil-like empty string",
			input: "",
			want:  0,
		},
		{
			name:  "single space",
			input: " ",
			want:  0,
		},
		{
			name:  "multiple spaces",
			input: "     ",
			want:  0,
		},
		{
			name:  "tabs only",
			input: "\t\t\t",
			want:  0,
		},
		{
			name:  "newlines only",
			input: "\n\n\n",
			want:  0,
		},
		{
			name:  "mixed whitespace",
			input: " \t\n \r \t\n ",
			want:  0,
		},
		{
			name:  "single character",
			input: "a",
			want:  1,
		},
		{
			name:  "single number",
			input: "123",
			want:  1,
		},
		{
			name:  "XML tags only (no content)",
			input: "<role></role>",
			want:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := countTokens(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCountTokens_RealWorldSkills(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantWords int
	}{
		{
			name: "simple go-code skill",
			input: `<role>You are a senior Go developer with 10 years of experience writing clean, production-ready code.</role>
<instructions>Follow these guidelines when writing Go code:
- Use interfaces at consumer side
- Return errors, don't panic
- Keep functions small and focused
- Write table-driven tests</instructions>`,
			wantWords: 43,
		},
		{
			name: "comprehensive API design skill",
			input: `<role>You are an API architect specializing in REST and gRPC design.</role>
<instructions>Design APIs following these principles:
1. Use nouns for resources, verbs for actions
2. Provide consistent error responses
3. Use appropriate HTTP status codes
4. Include pagination for list endpoints
5. Version your APIs using URL paths</instructions>
<constraints>- Never expose internal IDs
- Always validate input
- Rate limit all endpoints
- Log all requests for audit</constraints>`,
			wantWords: 68,
		},
		{
			name: "minimal skill",
			input: `<role>Go tester</role>
<instructions>Write tests</instructions>`,
			wantWords: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := countTokens(tt.input)
			wantTokens := int(float64(tt.wantWords) * 1.3)

			assert.Equal(t, wantTokens, got)
		})
	}
}
