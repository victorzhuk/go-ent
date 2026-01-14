package template

import (
	"fmt"
	"strings"
	"text/template"

	"embed"
)

// TemplateEngine wraps Go's text/template with custom functions.
type TemplateEngine struct {
	fs embed.FS
}

// NewTemplateEngine creates a new template engine.
func NewTemplateEngine(fs embed.FS) *TemplateEngine {
	return &TemplateEngine{fs: fs}
}

// Funcs returns the FuncMap for use with text/template.
func (e *TemplateEngine) Funcs() template.FuncMap {
	return template.FuncMap{
		"include": e.include,
		"if_tool": e.ifTool,
		"model":   e.model,
		"list":    e.list,
		"tools":   e.tools,
	}
}

// include reads a shared section from the embedded filesystem.
// Reads file from plugins/go-ent/agents/prompts/shared/{name}.md
func (e *TemplateEngine) include(name string) string {
	path := fmt.Sprintf("plugins/go-ent/agents/prompts/shared/%s.md", name)
	content, err := e.fs.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(content)
}

// ifTool returns true if the specified tool is in the current tool context.
// For now returns false, can be extended with tool context later.
func (e *TemplateEngine) ifTool(tool string) bool {
	return false
}

// model returns the appropriate model based on category and tool.
// Category: fast, main, heavy
// Tool: claude, opencode
func (e *TemplateEngine) model(category, tool string) string {
	claudeModels := map[string]string{
		"fast":  "haiku",
		"main":  "sonnet",
		"heavy": "opus",
	}

	openCodeModels := map[string]string{
		"fast":  "gpt-4o-mini",
		"main":  "gpt-4",
		"heavy": "o1-preview",
	}

	switch tool {
	case "claude":
		if m, ok := claudeModels[category]; ok {
			return m
		}
		return "sonnet"
	case "opencode":
		if m, ok := openCodeModels[category]; ok {
			return m
		}
		return "gpt-4"
	default:
		return ""
	}
}

// list formats an array as a newline-separated list.
func (e *TemplateEngine) list(array []string) string {
	return strings.Join(array, "\n")
}

// tools formats the tools array for the specified tool format.
// For Claude: returns YAML key-value format
// For OpenCode: returns YAML array format
func (e *TemplateEngine) tools(tools []string, tool string) string {
	if len(tools) == 0 {
		return ""
	}

	switch tool {
	case "claude":
		var result []string
		for _, t := range tools {
			result = append(result, fmt.Sprintf("  - name: %s", t))
		}
		return strings.Join(result, "\n")
	case "opencode":
		var result []string
		for _, t := range tools {
			result = append(result, fmt.Sprintf("  - %s", t))
		}
		return strings.Join(result, "\n")
	default:
		return ""
	}
}
