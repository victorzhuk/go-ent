package tools

import (
	"context"
	"fmt"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/cmd/goent/internal/spec"
)

type SpecCreateInput struct {
	Type    string `json:"type"`
	ID      string `json:"id"`
	Content string `json:"content"`
}

type SpecUpdateInput struct {
	Type    string `json:"type"`
	ID      string `json:"id"`
	Content string `json:"content"`
}

type SpecDeleteInput struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

func registerCRUD(s *mcp.Server) {
	createTool := &mcp.Tool{
		Name:        "goent_spec_create",
		Description: "Create a new spec, change, or task",
	}
	mcp.AddTool(s, createTool, specCreateHandler)

	updateTool := &mcp.Tool{
		Name:        "goent_spec_update",
		Description: "Update an existing spec, change, or task",
	}
	mcp.AddTool(s, updateTool, specUpdateHandler)

	deleteTool := &mcp.Tool{
		Name:        "goent_spec_delete",
		Description: "Delete a spec, change, or task",
	}
	mcp.AddTool(s, deleteTool, specDeleteHandler)
}

func specCreateHandler(ctx context.Context, req *mcp.CallToolRequest, input SpecCreateInput) (*mcp.CallToolResult, any, error) {
	if input.Type == "" || input.ID == "" || input.Content == "" {
		return nil, nil, fmt.Errorf("type, id, and content are required")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error getting current directory: %v", err)}},
		}, nil, nil
	}

	store := spec.NewStore(cwd)

	exists, err := store.Exists()
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error checking .spec folder: %v", err)}},
		}, nil, nil
	}

	if !exists {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "No openspec folder found. Run goent_spec_init first."}},
		}, nil, nil
	}

	var path string
	switch input.Type {
	case "spec":
		path = fmt.Sprintf("specs/%s/spec.md", input.ID)
	case "change":
		path = fmt.Sprintf("changes/%s/proposal.md", input.ID)
	case "task":
		path = fmt.Sprintf("tasks/%s.md", input.ID)
	default:
		return nil, nil, fmt.Errorf("invalid type: %s", input.Type)
	}

	if err := store.WriteFile(path, input.Content); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error creating %s: %v", path, err)}},
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("✅ Created %s: %s at %s", input.Type, input.ID, path)}},
	}, nil, nil
}

func specUpdateHandler(ctx context.Context, req *mcp.CallToolRequest, input SpecUpdateInput) (*mcp.CallToolResult, any, error) {
	if input.Type == "" || input.ID == "" || input.Content == "" {
		return nil, nil, fmt.Errorf("type, id, and content are required")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error getting current directory: %v", err)}},
		}, nil, nil
	}

	store := spec.NewStore(cwd)

	exists, err := store.Exists()
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error checking .spec folder: %v", err)}},
		}, nil, nil
	}

	if !exists {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "No openspec folder found. Run goent_spec_init first."}},
		}, nil, nil
	}

	var path string
	switch input.Type {
	case "spec":
		path = fmt.Sprintf("specs/%s/spec.md", input.ID)
	case "change":
		path = fmt.Sprintf("changes/%s/proposal.md", input.ID)
	case "task":
		path = fmt.Sprintf("tasks/%s.md", input.ID)
	default:
		return nil, nil, fmt.Errorf("invalid type: %s", input.Type)
	}

	if err := store.WriteFile(path, input.Content); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error updating %s: %v", path, err)}},
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("✅ Updated %s: %s", input.Type, input.ID)}},
	}, nil, nil
}

func specDeleteHandler(ctx context.Context, req *mcp.CallToolRequest, input SpecDeleteInput) (*mcp.CallToolResult, any, error) {
	if input.Type == "" || input.ID == "" {
		return nil, nil, fmt.Errorf("type and id are required")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error getting current directory: %v", err)}},
		}, nil, nil
	}

	store := spec.NewStore(cwd)

	exists, err := store.Exists()
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error checking .spec folder: %v", err)}},
		}, nil, nil
	}

	if !exists {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "No openspec folder found. Run goent_spec_init first."}},
		}, nil, nil
	}

	var err2 error
	switch input.Type {
	case "spec":
		err2 = store.DeleteDir(fmt.Sprintf("specs/%s", input.ID))
	case "change":
		err2 = store.DeleteDir(fmt.Sprintf("changes/%s", input.ID))
	case "task":
		err2 = store.DeleteFile(fmt.Sprintf("tasks/%s.md", input.ID))
	default:
		return nil, nil, fmt.Errorf("invalid type: %s", input.Type)
	}

	if err2 != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error deleting %s: %v", input.Type, err2)}},
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("✅ Deleted %s: %s", input.Type, input.ID)}},
	}, nil, nil
}
