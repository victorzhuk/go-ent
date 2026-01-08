package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/victorzhuk/go-ent/internal/spec"
)

type WorkflowStartInput struct {
	Path     string                 `json:"path"`
	ChangeID string                 `json:"change_id"`
	Phase    string                 `json:"phase"`
	Context  map[string]interface{} `json:"context,omitempty"`
}

type WorkflowApproveInput struct {
	Path string `json:"path"`
}

type WorkflowStatusInput struct {
	Path string `json:"path"`
}

func registerWorkflow(s *mcp.Server) {
	startTool := &mcp.Tool{
		Name:        "go_ent_workflow_start",
		Description: "Start a guided workflow with state tracking and wait points",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path":      map[string]any{"type": "string", "description": "Path to project directory"},
				"change_id": map[string]any{"type": "string", "description": "Change ID to work on"},
				"phase":     map[string]any{"type": "string", "description": "Workflow phase to execute"},
				"context":   map[string]any{"type": "object", "description": "Additional context for the workflow"},
			},
			"required": []string{"path", "change_id", "phase"},
		},
	}
	mcp.AddTool(s, startTool, workflowStartHandler)

	approveTool := &mcp.Tool{
		Name:        "go_ent_workflow_approve",
		Description: "Approve the current wait point and continue workflow",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{"type": "string", "description": "Path to project directory"},
			},
			"required": []string{"path"},
		},
	}
	mcp.AddTool(s, approveTool, workflowApproveHandler)

	statusTool := &mcp.Tool{
		Name:        "go_ent_workflow_status",
		Description: "Check current workflow status and wait points",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{"type": "string", "description": "Path to project directory"},
			},
			"required": []string{"path"},
		},
	}
	mcp.AddTool(s, statusTool, workflowStatusHandler)
}

func workflowStartHandler(ctx context.Context, req *mcp.CallToolRequest, input WorkflowStartInput) (*mcp.CallToolResult, any, error) {
	if input.Path == "" {
		return nil, nil, fmt.Errorf("path is required")
	}
	if input.ChangeID == "" {
		return nil, nil, fmt.Errorf("change_id is required")
	}
	if input.Phase == "" {
		return nil, nil, fmt.Errorf("phase is required")
	}

	store := spec.NewStore(input.Path)

	if store.WorkflowExists() {
		existing, err := store.LoadWorkflow()
		if err == nil && existing.Status == spec.WorkflowStatusActive {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{
					Text: fmt.Sprintf("Workflow already active for change %s. Use go_ent_workflow_status to check or cancel first.", existing.ChangeID),
				}},
			}, nil, nil
		}
	}

	workflow := spec.NewWorkflowState(input.ChangeID, input.Phase)
	if input.Context != nil {
		workflow.Context = input.Context
	}

	if err := store.SaveWorkflow(workflow); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error starting workflow: %v", err)}},
		}, nil, nil
	}

	msg := fmt.Sprintf("✅ Workflow started\n\n")
	msg += fmt.Sprintf("ID: %s\n", workflow.ID)
	msg += fmt.Sprintf("Change: %s\n", workflow.ChangeID)
	msg += fmt.Sprintf("Phase: %s\n", workflow.Phase)
	msg += fmt.Sprintf("Status: %s\n", workflow.Status)

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
	}, nil, nil
}

func workflowApproveHandler(ctx context.Context, req *mcp.CallToolRequest, input WorkflowApproveInput) (*mcp.CallToolResult, any, error) {
	if input.Path == "" {
		return nil, nil, fmt.Errorf("path is required")
	}

	store := spec.NewStore(input.Path)

	if !store.WorkflowExists() {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "No active workflow found"}},
		}, nil, nil
	}

	workflow, err := store.LoadWorkflow()
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error loading workflow: %v", err)}},
		}, nil, nil
	}

	if workflow.Status != spec.WorkflowStatusWaiting {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{
				Text: fmt.Sprintf("Workflow not waiting for approval (status: %s)", workflow.Status),
			}},
		}, nil, nil
	}

	workflow.Approve()

	if err := store.SaveWorkflow(workflow); err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error approving workflow: %v", err)}},
		}, nil, nil
	}

	msg := fmt.Sprintf("✅ Wait point approved\n\n")
	msg += fmt.Sprintf("Phase: %s\n", workflow.Phase)
	msg += fmt.Sprintf("Status: %s (ready to continue)\n", workflow.Status)

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
	}, nil, nil
}

func workflowStatusHandler(ctx context.Context, req *mcp.CallToolRequest, input WorkflowStatusInput) (*mcp.CallToolResult, any, error) {
	if input.Path == "" {
		return nil, nil, fmt.Errorf("path is required")
	}

	store := spec.NewStore(input.Path)

	if !store.WorkflowExists() {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: "No workflow found"}},
		}, nil, nil
	}

	workflow, err := store.LoadWorkflow()
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error loading workflow: %v", err)}},
		}, nil, nil
	}

	data, err := json.MarshalIndent(workflow, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Error formatting workflow: %v", err)}},
		}, nil, nil
	}

	msg := fmt.Sprintf("Workflow Status:\n\n```json\n%s\n```", string(data))

	if workflow.Status == spec.WorkflowStatusWaiting && workflow.WaitPoint != "" {
		msg += fmt.Sprintf("\n\n⏸️  WAITING: %s", workflow.WaitPoint)
		msg += "\n\nUse go_ent_workflow_approve to continue"
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
	}, nil, nil
}
