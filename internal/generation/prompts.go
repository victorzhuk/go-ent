package generation

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

// PromptContext contains variables for prompt template substitution.
type PromptContext struct {
	SpecContent   string
	Requirements  []Requirement
	ExistingCode  string
	ProjectName   string
	Conventions   string
	ComponentType string
}

// Requirement represents a single requirement from the spec.
type Requirement struct {
	Name        string
	Description string
}

// PromptTemplate represents a loaded prompt template.
type PromptTemplate struct {
	Name    string
	Content string
}

// LoadPromptTemplate loads a prompt template from the prompts/ directory.
func LoadPromptTemplate(projectRoot, templateType string) (*PromptTemplate, error) {
	templatePath := filepath.Join(projectRoot, "prompts", templateType+".md")

	content, err := os.ReadFile(templatePath) // #nosec G304 -- controlled file path
	if err != nil {
		// If template doesn't exist, return built-in default
		if os.IsNotExist(err) {
			return getBuiltInPrompt(templateType), nil
		}
		return nil, fmt.Errorf("read prompt template: %w", err)
	}

	return &PromptTemplate{
		Name:    templateType,
		Content: string(content),
	}, nil
}

// Execute executes the prompt template with the given context.
func (pt *PromptTemplate) Execute(ctx PromptContext) (string, error) {
	tmpl, err := template.New(pt.Name).Parse(pt.Content)
	if err != nil {
		return "", fmt.Errorf("parse prompt template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return "", fmt.Errorf("execute prompt template: %w", err)
	}

	return buf.String(), nil
}

// getBuiltInPrompt returns a built-in prompt template for the given type.
func getBuiltInPrompt(templateType string) *PromptTemplate {
	templates := map[string]string{
		"usecase":    usecasePromptTemplate,
		"handler":    handlerPromptTemplate,
		"repository": repositoryPromptTemplate,
	}

	content, ok := templates[templateType]
	if !ok {
		content = genericPromptTemplate
	}

	return &PromptTemplate{
		Name:    templateType,
		Content: content,
	}
}

const usecasePromptTemplate = `# Use Case Implementation

## Context
You are generating a use case implementation for a Go service following clean architecture patterns.

## Spec
{{.SpecContent}}

## Requirements to Implement
{{range .Requirements}}
- **{{.Name}}**: {{.Description}}
{{end}}

## Generated Code So Far
` + "```go" + `
{{.ExistingCode}}
` + "```" + `

## Instructions
1. Implement each requirement as a method on the use case struct
2. Follow clean architecture patterns - use cases orchestrate domain logic
3. Return proper errors with context using fmt.Errorf
4. Add input validation where appropriate
5. Use dependency injection for repositories and services
6. Keep methods focused and single-purpose

## Output Format
Generate only the method implementations. Do not regenerate the struct or constructor.
Use the repository and service interfaces that are already injected.
`

const handlerPromptTemplate = `# Handler Implementation

## Context
You are generating HTTP/API handlers for a Go service.

## Spec
{{.SpecContent}}

## Requirements to Implement
{{range .Requirements}}
- **{{.Name}}**: {{.Description}}
{{end}}

## Generated Code So Far
` + "```go" + `
{{.ExistingCode}}
` + "```" + `

## Instructions
1. Implement each endpoint as a handler method
2. Parse and validate request DTOs
3. Call the appropriate use case method
4. Map domain responses to API DTOs
5. Return proper HTTP status codes
6. Handle errors appropriately (400 for validation, 500 for internal errors)
7. Use JSON for request/response bodies

## Output Format
Generate only the handler method implementations.
`

const repositoryPromptTemplate = `# Repository Implementation

## Context
You are generating repository implementations for data persistence in a Go service.

## Spec
{{.SpecContent}}

## Requirements to Implement
{{range .Requirements}}
- **{{.Name}}**: {{.Description}}
{{end}}

## Generated Code So Far
` + "```go" + `
{{.ExistingCode}}
` + "```" + `

## Instructions
1. Implement each repository method following the repository pattern
2. Use parameterized queries to prevent SQL injection
3. Map between domain entities and database models
4. Return domain entities, not database models
5. Handle errors appropriately (ErrNotFound for missing records)
6. Use transactions for operations that modify multiple records
7. Use context for cancellation and timeouts

## Output Format
Generate only the repository method implementations.
`

const genericPromptTemplate = `# {{.ComponentType}} Implementation

## Context
You are generating a {{.ComponentType}} for a Go service.

## Spec
{{.SpecContent}}

## Requirements to Implement
{{range .Requirements}}
- **{{.Name}}**: {{.Description}}
{{end}}

## Generated Code So Far
` + "```go" + `
{{.ExistingCode}}
` + "```" + `

## Instructions
1. Implement each requirement as appropriate for this component type
2. Follow Go best practices and idioms
3. Use proper error handling with context
4. Keep code simple and focused
5. Add comments only where the intent isn't obvious

## Output Format
Generate the implementation for the marked extension points.
`
