package genspec

// Spec defines validation rules for a tool (agent or skill)
type Spec struct {
	Description string               `yaml:"description"`
	Fields      map[string]FieldSpec `yaml:"fields"`
}

// FieldSpec defines validation rules for a single field
type FieldSpec struct {
	Type        string   `yaml:"type"`        // string, array, object, boolean, integer, float
	Required    bool     `yaml:"required"`    // field must be present
	Values      []string `yaml:"values"`      // allowed enum values (for string type)
	Default     any      `yaml:"default"`     // default value if not specified
	ItemType    string   `yaml:"itemType"`    // type of array items
	Description string   `yaml:"description"` // field documentation
}

// ToolSpec contains agent and skill specs for a tool
type ToolSpec struct {
	Agent Spec `yaml:"agent"`
	Skill Spec `yaml:"skill"`
}

// ValidationError represents a validation failure
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// ValidationResult contains all validation errors for a file
type ValidationResult struct {
	File   string
	Errors []ValidationError
}

// IsValid returns true if there are no errors
func (r ValidationResult) IsValid() bool {
	return len(r.Errors) == 0
}
