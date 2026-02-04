package cli

import (
	"strings"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// templateHelpers returns a FuncMap with helper functions for templates
func templateHelpers() template.FuncMap {
	caser := cases.Title(language.English)
	return template.FuncMap{
		"title":     caser.String,
		"upper":     strings.ToUpper,
		"lower":     strings.ToLower,
		"contains":  strings.Contains,
		"hasPrefix": strings.HasPrefix,
		"hasSuffix": strings.HasSuffix,
		"join":      strings.Join,
		"replace":   strings.ReplaceAll,
	}
}

// AgentTemplateData holds data for agent template rendering
type AgentTemplateData struct {
	// Agent metadata
	Name         string
	Description  string
	Role         string
	RoleTitle    string // Human-readable role title
	Complexity   string
	Dependencies []string
	Skills       []string

	// Tool configuration
	AllowedTools       []string
	DisallowedTools    []string
	HasDisallowedTools bool

	// Content sections
	AgentContent  string   // Agent-specific prompt content
	SharedPrompts []string // Shared prompt contents
}

// NewAgentTemplateData creates template data from agent metadata
func NewAgentTemplateData(meta *agentMeta, agentContent string, sharedPrompts []string) *AgentTemplateData {
	data := &AgentTemplateData{
		Name:            meta.Name,
		Description:     meta.Description,
		Role:            meta.Role,
		RoleTitle:       getRoleTitle(meta.Role),
		Complexity:      meta.Complexity,
		Dependencies:    meta.Dependencies,
		Skills:          meta.Skills,
		DisallowedTools: meta.DisallowedTools,
		AgentContent:    agentContent,
		SharedPrompts:   sharedPrompts,
	}

	data.HasDisallowedTools = len(meta.DisallowedTools) > 0

	return data
}

// getRoleTitle returns human-readable role title
func getRoleTitle(role string) string {
	roleMap := map[string]string{
		"execution":  "Implementation",
		"planning":   "Planning",
		"validation": "Validation",
		"research":   "Research",
	}

	if title, ok := roleMap[role]; ok {
		return title
	}
	caser := cases.Title(language.English)
	return caser.String(role)
}
