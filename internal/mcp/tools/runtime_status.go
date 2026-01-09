package tools

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/config"
	"github.com/victorzhuk/go-ent/internal/domain"
)

type RuntimeStatusInput struct{}

func registerRuntimeStatus(s *mcp.Server) {
	tool := &mcp.Tool{
		Name:        "runtime_status",
		Description: "Get current runtime environment status and configuration",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
	}

	mcp.AddTool(s, tool, runtimeStatusHandler)
}

func runtimeStatusHandler(ctx context.Context, req *mcp.CallToolRequest, input RuntimeStatusInput) (*mcp.CallToolResult, any, error) {
	var sb strings.Builder
	sb.WriteString("# Runtime Status\n\n")

	// Detect current runtime
	currentRuntime := detectCurrentRuntime()
	sb.WriteString(fmt.Sprintf("## Current Runtime: %s\n\n", currentRuntime))

	cap := domain.NewRuntimeCapability(currentRuntime)
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

	// Try to load configuration
	cwd, err := os.Getwd()
	if err == nil {
		cfg, err := config.Load(cwd)
		if err == nil {
			sb.WriteString("## Configuration\n\n")
			sb.WriteString(fmt.Sprintf("**Preferred Runtime**: %s\n", cfg.Runtime.Preferred))
			if len(cfg.Runtime.Fallback) > 0 {
				sb.WriteString("**Fallback Runtimes**: ")
				fallbacks := make([]string, len(cfg.Runtime.Fallback))
				for i, fb := range cfg.Runtime.Fallback {
					fallbacks[i] = string(fb)
				}
				sb.WriteString(strings.Join(fallbacks, ", "))
				sb.WriteString("\n")
			}
			sb.WriteString("\n")
		} else {
			sb.WriteString("## Configuration\n\n")
			sb.WriteString("*No configuration found. Using defaults.*\n\n")
		}
	}

	// Environment info
	sb.WriteString("## Environment\n\n")
	if claudeCodePath := os.Getenv("CLAUDE_CODE_PATH"); claudeCodePath != "" {
		sb.WriteString(fmt.Sprintf("- CLAUDE_CODE_PATH: %s\n", claudeCodePath))
	}
	if openCodePath := os.Getenv("OPEN_CODE_PATH"); openCodePath != "" {
		sb.WriteString(fmt.Sprintf("- OPEN_CODE_PATH: %s\n", openCodePath))
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: sb.String()}},
	}, nil, nil
}

// detectCurrentRuntime attempts to detect which runtime environment is active.
func detectCurrentRuntime() domain.Runtime {
	// Check for Claude Code environment variables
	if os.Getenv("CLAUDE_CODE") != "" || os.Getenv("ANTHROPIC_CLI") != "" {
		return domain.RuntimeClaudeCode
	}

	// Check for OpenCode environment variables
	if os.Getenv("OPEN_CODE") != "" {
		return domain.RuntimeOpenCode
	}

	// Check if running in MCP server mode (stdio communication)
	// If we're being called via MCP, we're likely in an IDE environment
	if os.Getenv("MCP_SERVER") != "" {
		return domain.RuntimeClaudeCode // Default to Claude Code for MCP
	}

	// Default to CLI if no specific runtime detected
	return domain.RuntimeCLI
}
