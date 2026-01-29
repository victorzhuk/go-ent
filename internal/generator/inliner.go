package generator

import (
	"fmt"
	"strings"
)

// InlinePrompts combines shared prompts with the main agent prompt
func InlinePrompts(prompts *PromptContent, agent *AgentSource) string {
	var sb strings.Builder

	// Write main agent prompt first
	sb.WriteString(prompts.Main)
	sb.WriteString("\n\n")

	// Append shared prompts in order specified
	for _, name := range agent.Prompts.Shared {
		content, ok := prompts.Shared[name]
		if !ok {
			// This shouldn't happen if LoadPrompts succeeded
			continue
		}

		// Add section header for clarity
		header := strings.ToUpper(name[:1]) + name[1:]
		fmt.Fprintf(&sb, "## %s\n\n", header)
		sb.WriteString(content)
		sb.WriteString("\n\n")
	}

	return strings.TrimSpace(sb.String())
}
