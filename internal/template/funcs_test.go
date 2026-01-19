package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTemplateEngine(t *testing.T) {
	t.Parallel()

	engine := NewTemplateEngine(testFS)
	assert.NotNil(t, engine)
}

func TestTemplateEngine_Funcs(t *testing.T) {
	t.Parallel()

	engine := NewTemplateEngine(testFS)
	funcs := engine.Funcs()

	assert.NotNil(t, funcs)
	assert.Contains(t, funcs, "include")
	assert.Contains(t, funcs, "if_tool")
	assert.Contains(t, funcs, "model")
	assert.Contains(t, funcs, "list")
	assert.Contains(t, funcs, "tools")
}

func TestIfTool(t *testing.T) {
	t.Parallel()

	engine := NewTemplateEngine(testFS)

	result := engine.ifTool("any-tool")
	assert.False(t, result)
}

func TestModel_Claude(t *testing.T) {
	t.Parallel()

	engine := NewTemplateEngine(testFS)

	tests := []struct {
		name     string
		category string
		want     string
	}{
		{"fast category", "fast", "haiku"},
		{"main category", "main", "sonnet"},
		{"heavy category", "heavy", "opus"},
		{"unknown category defaults to main", "unknown", "sonnet"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := engine.model(tt.category, "claude")
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestModel_OpenCode(t *testing.T) {
	t.Parallel()

	engine := NewTemplateEngine(testFS)

	tests := []struct {
		name     string
		category string
		want     string
	}{
		{"fast category", "fast", "gpt-4o-mini"},
		{"main category", "main", "gpt-4"},
		{"heavy category", "heavy", "o1-preview"},
		{"unknown category defaults to main", "unknown", "gpt-4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := engine.model(tt.category, "opencode")
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestModel_UnknownTool(t *testing.T) {
	t.Parallel()

	engine := NewTemplateEngine(testFS)

	result := engine.model("main", "unknown")
	assert.Empty(t, result)
}

func TestList(t *testing.T) {
	t.Parallel()

	engine := NewTemplateEngine(testFS)

	tests := []struct {
		name  string
		array []string
		want  string
	}{
		{
			name:  "single item",
			array: []string{"item1"},
			want:  "item1",
		},
		{
			name:  "multiple items",
			array: []string{"item1", "item2", "item3"},
			want:  "item1\nitem2\nitem3",
		},
		{
			name:  "empty array",
			array: []string{},
			want:  "",
		},
		{
			name:  "nil array",
			array: nil,
			want:  "",
		},
		{
			name:  "items with newlines",
			array: []string{"line1", "line2"},
			want:  "line1\nline2",
		},
		{
			name:  "items with spaces",
			array: []string{"item one", "item two"},
			want:  "item one\nitem two",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := engine.list(tt.array)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestTools_Claude(t *testing.T) {
	t.Parallel()

	engine := NewTemplateEngine(testFS)

	tests := []struct {
		name  string
		tools []string
		want  string
	}{
		{
			name:  "single tool",
			tools: []string{"bash"},
			want:  "  - name: bash",
		},
		{
			name:  "multiple tools",
			tools: []string{"bash", "read", "write"},
			want:  "  - name: bash\n  - name: read\n  - name: write",
		},
		{
			name:  "empty tools",
			tools: []string{},
			want:  "",
		},
		{
			name:  "nil tools",
			tools: nil,
			want:  "",
		},
		{
			name:  "tools with hyphens",
			tools: []string{"go-test", "run-lint"},
			want:  "  - name: go-test\n  - name: run-lint",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := engine.tools(tt.tools, "claude")
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestTools_OpenCode(t *testing.T) {
	t.Parallel()

	engine := NewTemplateEngine(testFS)

	tests := []struct {
		name  string
		tools []string
		want  string
	}{
		{
			name:  "single tool",
			tools: []string{"bash"},
			want:  "  - bash",
		},
		{
			name:  "multiple tools",
			tools: []string{"bash", "read", "write"},
			want:  "  - bash\n  - read\n  - write",
		},
		{
			name:  "empty tools",
			tools: []string{},
			want:  "",
		},
		{
			name:  "nil tools",
			tools: nil,
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := engine.tools(tt.tools, "opencode")
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestTools_UnknownTool(t *testing.T) {
	t.Parallel()

	engine := NewTemplateEngine(testFS)

	result := engine.tools([]string{"bash", "read"}, "unknown")
	assert.Empty(t, result)
}
