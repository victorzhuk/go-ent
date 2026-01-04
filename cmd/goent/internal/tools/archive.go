package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/spec"
)

type ArchiveInput struct {
	Path      string `json:"path"`
	ID        string `json:"id"`
	SkipSpecs bool   `json:"skip_specs,omitempty"`
	DryRun    bool   `json:"dry_run,omitempty"`
}

func registerArchive(s *mcp.Server) {
	tool := &mcp.Tool{
		Name:        "goent_spec_archive",
		Description: "Archive a completed change and optionally merge deltas into specs. Use dry_run to preview changes.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "Path to the project directory containing openspec folder",
				},
				"id": map[string]any{
					"type":        "string",
					"description": "ID of the change to archive",
				},
				"skip_specs": map[string]any{
					"type":        "boolean",
					"description": "Skip merging delta specs into main specs (useful for tooling-only changes)",
					"default":     false,
				},
				"dry_run": map[string]any{
					"type":        "boolean",
					"description": "Preview changes without actually archiving",
					"default":     false,
				},
			},
			"required": []string{"path", "id"},
		},
	}

	mcp.AddTool(s, tool, archiveHandler)
}

func archiveHandler(ctx context.Context, req *mcp.CallToolRequest, input ArchiveInput) (*mcp.CallToolResult, any, error) {
	if input.Path == "" {
		return nil, nil, fmt.Errorf("path is required")
	}
	if input.ID == "" {
		return nil, nil, fmt.Errorf("id is required")
	}

	store := spec.NewStore(input.Path)

	exists, err := store.Exists()
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error checking spec folder: %v", err)}},
		}, nil, nil
	}

	if !exists {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "No openspec folder found. Run goent_spec_init first."}},
		}, nil, nil
	}

	archiver := spec.NewArchiver(store)

	// Validate before archive
	validationResult, err := archiver.ValidateBeforeArchive(input.ID, true)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error validating change: %v", err)}},
		}, nil, nil
	}

	if !validationResult.Valid {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("❌ Cannot archive '%s' - validation failed\n\n", input.ID))
		sb.WriteString("Issues:\n")
		for _, issue := range validationResult.Issues {
			sb.WriteString(fmt.Sprintf("  %s\n", issue.String()))
		}
		sb.WriteString("\nFix the issues and try again.")
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: sb.String()}},
		}, nil, nil
	}

	// Perform archive
	result, err := archiver.Archive(input.ID, input.SkipSpecs, input.DryRun)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error archiving: %v", err)}},
		}, nil, nil
	}

	// Format result
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: formatArchiveResult(result)}},
	}, nil, nil
}

func formatArchiveResult(result *spec.ArchiveResult) string {
	var sb strings.Builder

	if result.DryRun {
		sb.WriteString("🔍 DRY RUN - No changes made\n\n")
	}

	if len(result.Errors) > 0 {
		sb.WriteString("❌ Archive completed with errors:\n")
		for _, err := range result.Errors {
			sb.WriteString(fmt.Sprintf("  - %s\n", err))
		}
		return sb.String()
	}

	if result.DryRun {
		sb.WriteString(fmt.Sprintf("Would archive '%s' to:\n  %s\n\n", result.ChangeID, result.ArchivePath))
	} else {
		sb.WriteString(fmt.Sprintf("✅ Archived '%s' to:\n  %s\n\n", result.ChangeID, result.ArchivePath))
	}

	if len(result.UpdatedSpecs) > 0 {
		if result.DryRun {
			sb.WriteString("Would update specs:\n")
		} else {
			sb.WriteString("Updated specs:\n")
		}
		for _, spec := range result.UpdatedSpecs {
			sb.WriteString(fmt.Sprintf("  - %s\n", spec))
		}
	} else {
		sb.WriteString("No spec updates (skip_specs=true or no deltas)\n")
	}

	return sb.String()
}
