package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/domain"
)

type RuntimeListInput struct{}

func registerRuntimeList(s *mcp.Server) {
	tool := &mcp.Tool{
		Name:        "runtime_list",
		Description: "List available runtime environments and their capabilities",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}

	mcp.AddTool(s, tool, runtimeListHandler)
}

func runtimeListHandler(ctx context.Context, req *mcp.CallToolRequest, input RuntimeListInput) (*mcp.CallToolResult, any, error) {
	runtimes := []domain.Runtime{
		domain.RuntimeClaudeCode,
		domain.RuntimeOpenCode,
		domain.RuntimeCLI,
	}

	var sb strings.Builder
	sb.WriteString("# Available Runtimes\n\n")

	for _, rt := range runtimes {
		cap := domain.NewRuntimeCapability(rt)

		sb.WriteString(fmt.Sprintf("## %s\n\n", rt))
		sb.WriteString(fmt.Sprintf("**Description**: %s\n\n", cap.Description))
		sb.WriteString("**Capabilities**:\n")
		sb.WriteString(fmt.Sprintf("- Interactive: %v\n", cap.SupportsInteractive))
		sb.WriteString(fmt.Sprintf("- Filesystem: %v\n", cap.SupportsFileSystem))
		sb.WriteString(fmt.Sprintf("- Tools: %v\n", cap.SupportsTools))
		sb.WriteString(fmt.Sprintf("- Skills: %v\n", cap.SupportsSkills))

		if cap.MaxConcurrentAgents > 0 {
			sb.WriteString(fmt.Sprintf("- Max Concurrent Agents: %d\n", cap.MaxConcurrentAgents))
		} else {
			sb.WriteString("- Max Concurrent Agents: unlimited\n")
		}
		sb.WriteString("\n")
	}

	sb.WriteString("---\n\n")
	sb.WriteString("**Usage**: Configure preferred runtime in `.go-ent/config.yaml`:\n\n")
	sb.WriteString("```yaml\n")
	sb.WriteString("runtime:\n")
	sb.WriteString("  preferred: claude-code\n")
	sb.WriteString("  fallback:\n")
	sb.WriteString("    - open-code\n")
	sb.WriteString("    - cli\n")
	sb.WriteString("```\n")

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: sb.String()}},
	}, nil, nil
}
