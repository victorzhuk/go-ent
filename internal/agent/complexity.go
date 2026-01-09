package agent

// ComplexityLevel represents the complexity classification of a task.
type ComplexityLevel int

const (
	ComplexityTrivial ComplexityLevel = iota
	ComplexitySimple
	ComplexityModerate
	ComplexityComplex
	ComplexityArchitectural
)

// String returns the string representation of the complexity level.
func (c ComplexityLevel) String() string {
	switch c {
	case ComplexityTrivial:
		return "trivial"
	case ComplexitySimple:
		return "simple"
	case ComplexityModerate:
		return "moderate"
	case ComplexityComplex:
		return "complex"
	case ComplexityArchitectural:
		return "architectural"
	default:
		return "unknown"
	}
}

// TaskComplexity holds the complexity analysis result.
type TaskComplexity struct {
	Level  ComplexityLevel
	Score  int
	Reason string
}

// Complexity analyzes task complexity.
type Complexity struct{}

// NewComplexity creates a new complexity analyzer.
func NewComplexity() *Complexity {
	return &Complexity{}
}

// Analyze determines the complexity level of a task.
func (c *Complexity) Analyze(task Task) TaskComplexity {
	score := c.calculateScore(task)
	level := c.scoreToLevel(score)

	return TaskComplexity{
		Level:  level,
		Score:  score,
		Reason: c.explainScore(task, score),
	}
}

func (c *Complexity) calculateScore(task Task) int {
	score := 0

	score += c.scoreByType(task.Type)
	score += c.scoreByDescription(task.Description)
	score += c.scoreByFiles(task.Files)

	return score
}

func (c *Complexity) scoreByType(taskType TaskType) int {
	switch taskType {
	case TaskTypeArchitecture:
		return 40
	case TaskTypeFeature:
		return 25
	case TaskTypeRefactor:
		return 20
	case TaskTypeBugFix:
		return 15
	case TaskTypeTest:
		return 10
	case TaskTypeDocumentation:
		return 5
	default:
		return 15
	}
}

func (c *Complexity) scoreByDescription(desc string) int {
	score := 0

	keywords := map[string]int{
		"architecture": 15,
		"design":       15,
		"refactor":     10,
		"implement":    8,
		"integrate":    10,
		"migrate":      12,
		"add":          5,
		"fix":          3,
		"update":       3,
		"create":       5,
	}

	for keyword, points := range keywords {
		if contains(desc, keyword) {
			score += points
		}
	}

	return score
}

func (c *Complexity) scoreByFiles(files []string) int {
	fileCount := len(files)

	switch {
	case fileCount == 0:
		return 0
	case fileCount == 1:
		return 2
	case fileCount <= 3:
		return 5
	case fileCount <= 5:
		return 10
	default:
		return 15
	}
}

func (c *Complexity) scoreToLevel(score int) ComplexityLevel {
	switch {
	case score >= 50:
		return ComplexityArchitectural
	case score >= 35:
		return ComplexityComplex
	case score >= 20:
		return ComplexityModerate
	case score >= 10:
		return ComplexitySimple
	default:
		return ComplexityTrivial
	}
}

func (c *Complexity) explainScore(task Task, score int) string {
	return task.Type.String()
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(s == substr || len(s) >= len(substr) && hasSubstring(s, substr))
}

func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// String returns the string representation of the task type.
func (t TaskType) String() string {
	return string(t)
}
