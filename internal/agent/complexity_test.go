package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComplexityLevel_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		level ComplexityLevel
		want  string
	}{
		{"trivial", ComplexityTrivial, "trivial"},
		{"simple", ComplexitySimple, "simple"},
		{"moderate", ComplexityModerate, "moderate"},
		{"complex", ComplexityComplex, "complex"},
		{"architectural", ComplexityArchitectural, "architectural"},
		{"unknown", ComplexityLevel(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.level.String())
		})
	}
}

func TestComplexity_Analyze(t *testing.T) {
	t.Parallel()

	c := NewComplexity()

	tests := []struct {
		name      string
		task      Task
		wantLevel ComplexityLevel
		minScore  int
		maxScore  int
	}{
		{
			name: "trivial documentation task",
			task: Task{
				Description: "update readme",
				Type:        TaskTypeDocumentation,
				Files:       []string{},
			},
			wantLevel: ComplexityTrivial,
			minScore:  0,
			maxScore:  9,
		},
		{
			name: "simple test with one file",
			task: Task{
				Description: "add unit test",
				Type:        TaskTypeTest,
				Files:       []string{"test.go"},
			},
			wantLevel: ComplexitySimple,
			minScore:  10,
			maxScore:  19,
		},
		{
			name: "moderate feature with multiple files",
			task: Task{
				Description: "add new endpoint",
				Type:        TaskTypeFeature,
				Files:       []string{"handler.go", "service.go", "model.go"},
			},
			wantLevel: ComplexityComplex, // feature(25) + add(5) + 3files(5) = 35
			minScore:  35,
			maxScore:  49,
		},
		{
			name: "complex refactor with integration",
			task: Task{
				Description: "refactor and integrate new service",
				Type:        TaskTypeRefactor,
				Files:       []string{"a.go", "b.go", "c.go", "d.go"},
			},
			wantLevel: ComplexityArchitectural, // refactor(20) + refactor(10) + integrate(10) + 4files(10) = 50
			minScore:  50,
			maxScore:  100,
		},
		{
			name: "architectural design task",
			task: Task{
				Description: "design new architecture for microservices",
				Type:        TaskTypeArchitecture,
				Files:       []string{"a.go", "b.go", "c.go", "d.go", "e.go", "f.go"},
			},
			wantLevel: ComplexityArchitectural,
			minScore:  50,
			maxScore:  100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := c.Analyze(tt.task)

			assert.Equal(t, tt.wantLevel, result.Level)
			assert.GreaterOrEqual(t, result.Score, tt.minScore)
			assert.LessOrEqual(t, result.Score, tt.maxScore)
			assert.NotEmpty(t, result.Reason)
		})
	}
}

func TestComplexity_scoreByType(t *testing.T) {
	t.Parallel()

	c := NewComplexity()

	tests := []struct {
		name      string
		taskType  TaskType
		wantScore int
	}{
		{"architecture", TaskTypeArchitecture, 40},
		{"feature", TaskTypeFeature, 25},
		{"refactor", TaskTypeRefactor, 20},
		{"bugfix", TaskTypeBugFix, 15},
		{"test", TaskTypeTest, 10},
		{"documentation", TaskTypeDocumentation, 5},
		{"unknown", TaskType("unknown"), 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.wantScore, c.scoreByType(tt.taskType))
		})
	}
}

func TestComplexity_scoreByDescription(t *testing.T) {
	t.Parallel()

	c := NewComplexity()

	tests := []struct {
		name        string
		description string
		wantScore   int
	}{
		{
			name:        "architecture keyword",
			description: "new architecture for system",
			wantScore:   15,
		},
		{
			name:        "design keyword",
			description: "design system",
			wantScore:   15,
		},
		{
			name:        "multiple keywords",
			description: "implement and integrate new feature",
			wantScore:   18, // implement(8) + integrate(10)
		},
		{
			name:        "migrate keyword",
			description: "migrate database schema",
			wantScore:   12,
		},
		{
			name:        "fix keyword",
			description: "fix bug in handler",
			wantScore:   3,
		},
		{
			name:        "create keyword",
			description: "create new service",
			wantScore:   5,
		},
		{
			name:        "add keyword",
			description: "add validation",
			wantScore:   5,
		},
		{
			name:        "update keyword",
			description: "update configuration",
			wantScore:   3,
		},
		{
			name:        "no keywords",
			description: "some random task",
			wantScore:   0,
		},
		{
			name:        "empty description",
			description: "",
			wantScore:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.wantScore, c.scoreByDescription(tt.description))
		})
	}
}

func TestComplexity_scoreByFiles(t *testing.T) {
	t.Parallel()

	c := NewComplexity()

	tests := []struct {
		name      string
		fileCount int
		wantScore int
	}{
		{"no files", 0, 0},
		{"one file", 1, 2},
		{"two files", 2, 5},
		{"three files", 3, 5},
		{"four files", 4, 10},
		{"five files", 5, 10},
		{"six files", 6, 15},
		{"ten files", 10, 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			files := make([]string, tt.fileCount)
			for i := 0; i < tt.fileCount; i++ {
				files[i] = "file.go"
			}

			assert.Equal(t, tt.wantScore, c.scoreByFiles(files))
		})
	}
}

func TestComplexity_scoreToLevel(t *testing.T) {
	t.Parallel()

	c := NewComplexity()

	tests := []struct {
		name      string
		score     int
		wantLevel ComplexityLevel
	}{
		{"score 0 -> trivial", 0, ComplexityTrivial},
		{"score 5 -> trivial", 5, ComplexityTrivial},
		{"score 9 -> trivial", 9, ComplexityTrivial},
		{"score 10 -> simple", 10, ComplexitySimple},
		{"score 15 -> simple", 15, ComplexitySimple},
		{"score 19 -> simple", 19, ComplexitySimple},
		{"score 20 -> moderate", 20, ComplexityModerate},
		{"score 25 -> moderate", 25, ComplexityModerate},
		{"score 34 -> moderate", 34, ComplexityModerate},
		{"score 35 -> complex", 35, ComplexityComplex},
		{"score 40 -> complex", 40, ComplexityComplex},
		{"score 49 -> complex", 49, ComplexityComplex},
		{"score 50 -> architectural", 50, ComplexityArchitectural},
		{"score 70 -> architectural", 70, ComplexityArchitectural},
		{"score 100 -> architectural", 100, ComplexityArchitectural},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.wantLevel, c.scoreToLevel(tt.score))
		})
	}
}

func TestComplexity_contains(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		s      string
		substr string
		want   bool
	}{
		{"exact match", "test", "test", true},
		{"substring at start", "testing", "test", true},
		{"substring in middle", "this is a test", "is a", true},
		{"substring at end", "hello world", "world", true},
		{"not found", "hello", "bye", false},
		{"empty substring", "hello", "", false},
		{"empty string", "", "test", false},
		{"both empty", "", "", false},
		{"case sensitive", "Hello", "hello", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, contains(tt.s, tt.substr))
		})
	}
}

func TestComplexity_realWorldScenarios(t *testing.T) {
	t.Parallel()

	c := NewComplexity()

	tests := []struct {
		name      string
		task      Task
		wantLevel ComplexityLevel
	}{
		{
			name: "quick typo fix",
			task: Task{
				Description: "fix typo in comment",
				Type:        TaskTypeBugFix,
				Files:       []string{"main.go"},
			},
			wantLevel: ComplexityModerate, // bugfix(15) + fix(3) + 1file(2) = 20
		},
		{
			name: "add endpoint with tests",
			task: Task{
				Description: "add REST endpoint",
				Type:        TaskTypeFeature,
				Files:       []string{"handler.go", "handler_test.go", "service.go"},
			},
			wantLevel: ComplexityComplex, // feature(25) + add(5) + 3files(5) = 35
		},
		{
			name: "database migration",
			task: Task{
				Description: "migrate user table schema",
				Type:        TaskTypeRefactor,
				Files:       []string{"migration.sql", "repo.go", "model.go"},
			},
			wantLevel: ComplexityComplex, // refactor(20) + migrate(12) + 3files(5) = 37
		},
		{
			name: "full system redesign",
			task: Task{
				Description: "design and implement new architecture",
				Type:        TaskTypeArchitecture,
				Files: []string{
					"app.go", "config.go", "handler.go",
					"service.go", "repo.go", "model.go",
				},
			},
			wantLevel: ComplexityArchitectural, // architecture(40) + design(15) + implement(8) + 6files(15) = 78
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := c.Analyze(tt.task)

			assert.Equal(t, tt.wantLevel, result.Level,
				"Task: %s\nExpected: %s (score should be in range for %s)\nGot: %s (score: %d)",
				tt.task.Description, tt.wantLevel, tt.wantLevel, result.Level, result.Score)
		})
	}
}
