package genspec

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Validator validates generated agent files against tool specs
type Validator struct {
	spec *ToolSpec
}

// NewValidator creates a validator for the given tool
func NewValidator(tool string) (*Validator, error) {
	spec, err := LoadToolSpec(tool)
	if err != nil {
		return nil, fmt.Errorf("load spec: %w", err)
	}

	return &Validator{spec: spec}, nil
}

// ValidateAgent validates an agent markdown file
func (v *Validator) ValidateAgent(path string) ValidationResult {
	result := ValidationResult{File: path}

	// Read file
	content, err := os.ReadFile(path)
	if err != nil {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "",
			Message: fmt.Sprintf("read file: %v", err),
		})
		return result
	}

	// Extract frontmatter
	frontmatter, err := extractFrontmatter(content)
	if err != nil {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "",
			Message: fmt.Sprintf("parse frontmatter: %v", err),
		})
		return result
	}

	// Parse frontmatter as map
	var data map[string]any
	if err := yaml.Unmarshal([]byte(frontmatter), &data); err != nil {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "",
			Message: fmt.Sprintf("parse YAML: %v", err),
		})
		return result
	}

	// Validate against spec
	result.Errors = v.validateFields(data, v.spec.Agent)

	return result
}

// validateFields validates a data map against a spec
func (v *Validator) validateFields(data map[string]any, spec Spec) []ValidationError {
	var errs []ValidationError

	// Check required fields
	for name, field := range spec.Fields {
		value, exists := data[name]

		if field.Required && !exists {
			errs = append(errs, ValidationError{
				Field:   name,
				Message: "required field missing",
			})
			continue
		}

		if !exists {
			continue
		}

		// Validate type and constraints
		if err := v.validateField(name, value, field); err != nil {
			errs = append(errs, *err)
		}
	}

	// Check for unknown fields (strict mode)
	for name := range data {
		if _, known := spec.Fields[name]; !known {
			errs = append(errs, ValidationError{
				Field:   name,
				Message: "unknown field (not in spec)",
			})
		}
	}

	return errs
}

// validateField validates a single field value against its spec
func (v *Validator) validateField(name string, value any, spec FieldSpec) *ValidationError {
	switch spec.Type {
	case "string":
		str, ok := value.(string)
		if !ok {
			return &ValidationError{
				Field:   name,
				Message: fmt.Sprintf("expected string, got %T", value),
			}
		}

		// Check enum values
		if len(spec.Values) > 0 {
			found := false
			for _, allowed := range spec.Values {
				if str == allowed {
					found = true
					break
				}
			}
			if !found {
				return &ValidationError{
					Field:   name,
					Message: fmt.Sprintf("invalid value %q, must be one of: %v", str, spec.Values),
				}
			}
		}

	case "array":
		arr, ok := value.([]any)
		if !ok {
			return &ValidationError{
				Field:   name,
				Message: fmt.Sprintf("expected array, got %T", value),
			}
		}

		// Validate item types if specified
		if spec.ItemType != "" {
			for i, item := range arr {
				if err := v.validateItemType(spec.ItemType, item); err != nil {
					return &ValidationError{
						Field:   fmt.Sprintf("%s[%d]", name, i),
						Message: err.Error(),
					}
				}
			}
		}

	case "object":
		_, ok := value.(map[string]any)
		if !ok {
			return &ValidationError{
				Field:   name,
				Message: fmt.Sprintf("expected object, got %T", value),
			}
		}

	case "boolean":
		_, ok := value.(bool)
		if !ok {
			return &ValidationError{
				Field:   name,
				Message: fmt.Sprintf("expected boolean, got %T", value),
			}
		}

	case "integer":
		_, ok := value.(int)
		if !ok {
			return &ValidationError{
				Field:   name,
				Message: fmt.Sprintf("expected integer, got %T", value),
			}
		}

	case "float":
		_, ok := value.(float64)
		if !ok {
			return &ValidationError{
				Field:   name,
				Message: fmt.Sprintf("expected float, got %T", value),
			}
		}
	}

	return nil
}

// validateItemType validates the type of an array item
func (v *Validator) validateItemType(expectedType string, value any) error {
	switch expectedType {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case "integer":
		if _, ok := value.(int); !ok {
			return fmt.Errorf("expected integer, got %T", value)
		}
	case "float":
		if _, ok := value.(float64); !ok {
			return fmt.Errorf("expected float, got %T", value)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected boolean, got %T", value)
		}
	}
	return nil
}

// extractFrontmatter extracts YAML frontmatter from markdown
func extractFrontmatter(content []byte) (string, error) {
	lines := strings.Split(string(content), "\n")

	if len(lines) < 3 || lines[0] != "---" {
		return "", fmt.Errorf("missing frontmatter delimiter")
	}

	var fmLines []string
	for i := 1; i < len(lines); i++ {
		if lines[i] == "---" {
			return strings.Join(fmLines, "\n"), nil
		}
		fmLines = append(fmLines, lines[i])
	}

	return "", fmt.Errorf("unterminated frontmatter")
}

// ValidateDirectory validates all agent files in a directory
func (v *Validator) ValidateDirectory(dir string) ([]ValidationResult, error) {
	var results []ValidationResult

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		result := v.ValidateAgent(path)
		results = append(results, result)
	}

	return results, nil
}
