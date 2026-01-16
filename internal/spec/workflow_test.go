package spec

//nolint:gosec // test file with necessary file operations

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/victorzhuk/go-ent/internal/domain"
)

func TestNewWorkflowState(t *testing.T) {
	changeID := "test-change"
	phase := "planning"

	state := NewWorkflowState(changeID, phase)

	assert.NotEmpty(t, state.ID)
	assert.Equal(t, changeID, state.ChangeID)
	assert.Equal(t, phase, state.Phase)
	assert.Equal(t, WorkflowStatusActive, state.Status)
	assert.NotNil(t, state.Context)
	assert.Empty(t, state.Context)
	assert.Empty(t, state.AgentRole)
	assert.Empty(t, state.WaitPoint)
	assert.False(t, state.CreatedAt.IsZero())
	assert.False(t, state.UpdatedAt.IsZero())
	assert.Equal(t, state.CreatedAt, state.UpdatedAt)
}

func TestWorkflowState_SetAgent(t *testing.T) {
	state := NewWorkflowState("test", "planning")
	originalUpdated := state.UpdatedAt

	time.Sleep(1 * time.Millisecond)
	state.SetAgent(domain.AgentRoleArchitect)

	assert.Equal(t, domain.AgentRoleArchitect, state.AgentRole)
	assert.True(t, state.UpdatedAt.After(originalUpdated))
}

func TestWorkflowState_SetWaitPoint(t *testing.T) {
	state := NewWorkflowState("test", "planning")
	originalUpdated := state.UpdatedAt

	time.Sleep(1 * time.Millisecond)
	state.SetWaitPoint("user-approval")

	assert.Equal(t, "user-approval", state.WaitPoint)
	assert.Equal(t, WorkflowStatusWaiting, state.Status)
	assert.True(t, state.UpdatedAt.After(originalUpdated))
}

func TestWorkflowState_Approve(t *testing.T) {
	state := NewWorkflowState("test", "planning")
	state.SetWaitPoint("user-approval")
	originalUpdated := state.UpdatedAt

	time.Sleep(1 * time.Millisecond)
	state.Approve()

	assert.Empty(t, state.WaitPoint)
	assert.Equal(t, WorkflowStatusActive, state.Status)
	assert.True(t, state.UpdatedAt.After(originalUpdated))
}

func TestWorkflowState_Complete(t *testing.T) {
	state := NewWorkflowState("test", "planning")
	originalUpdated := state.UpdatedAt

	time.Sleep(1 * time.Millisecond)
	state.Complete()

	assert.Equal(t, WorkflowStatusCompleted, state.Status)
	assert.True(t, state.UpdatedAt.After(originalUpdated))
}

func TestWorkflowState_Cancel(t *testing.T) {
	state := NewWorkflowState("test", "planning")
	originalUpdated := state.UpdatedAt

	time.Sleep(1 * time.Millisecond)
	state.Cancel()

	assert.Equal(t, WorkflowStatusCancelled, state.Status)
	assert.True(t, state.UpdatedAt.After(originalUpdated))
}

func TestWorkflowState_StatusTransitions(t *testing.T) {
	tests := []struct {
		name           string
		setupFn        func(*WorkflowState)
		expectedStatus WorkflowStatus
	}{
		{
			name:           "initial state is active",
			setupFn:        func(s *WorkflowState) {},
			expectedStatus: WorkflowStatusActive,
		},
		{
			name: "active to waiting",
			setupFn: func(s *WorkflowState) {
				s.SetWaitPoint("approval")
			},
			expectedStatus: WorkflowStatusWaiting,
		},
		{
			name: "waiting to active",
			setupFn: func(s *WorkflowState) {
				s.SetWaitPoint("approval")
				s.Approve()
			},
			expectedStatus: WorkflowStatusActive,
		},
		{
			name: "active to completed",
			setupFn: func(s *WorkflowState) {
				s.Complete()
			},
			expectedStatus: WorkflowStatusCompleted,
		},
		{
			name: "active to cancelled",
			setupFn: func(s *WorkflowState) {
				s.Cancel()
			},
			expectedStatus: WorkflowStatusCancelled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewWorkflowState("test", "planning")
			tt.setupFn(state)
			assert.Equal(t, tt.expectedStatus, state.Status)
		})
	}
}

func TestWorkflowState_AgentTracking(t *testing.T) {
	state := NewWorkflowState("test-change", "execution")

	// Initially no agent
	assert.Empty(t, state.AgentRole)

	// Set architect for planning phase
	state.SetAgent(domain.AgentRoleArchitect)
	assert.Equal(t, domain.AgentRoleArchitect, state.AgentRole)

	// Change to developer for implementation
	state.SetAgent(domain.AgentRoleDeveloper)
	assert.Equal(t, domain.AgentRoleDeveloper, state.AgentRole)

	// Change to reviewer for validation
	state.SetAgent(domain.AgentRoleReviewer)
	assert.Equal(t, domain.AgentRoleReviewer, state.AgentRole)
}

func TestStore_WorkflowPath(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)

	// SpecPath() checks for openspec first, falls back to .spec
	// Since neither exists, it will use .spec
	expected := tmpDir + "/.spec/.workflow.yaml"
	assert.Equal(t, expected, store.WorkflowPath())
}

func TestStore_WorkflowPersistence(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)

	// Initialize the spec directory structure
	err := store.Init(Project{
		Name:        "test",
		Module:      "test/module",
		Description: "test",
	})
	require.NoError(t, err)

	// Initially no workflow exists
	assert.False(t, store.WorkflowExists())

	// Create and save workflow
	state := NewWorkflowState("test-change", "planning")
	state.SetAgent(domain.AgentRoleArchitect)
	state.Context["key"] = "value"

	err = store.SaveWorkflow(state)
	require.NoError(t, err)

	// Now workflow exists
	assert.True(t, store.WorkflowExists())

	// Load and verify
	loaded, err := store.LoadWorkflow()
	require.NoError(t, err)

	assert.Equal(t, state.ID, loaded.ID)
	assert.Equal(t, state.ChangeID, loaded.ChangeID)
	assert.Equal(t, state.Phase, loaded.Phase)
	assert.Equal(t, state.AgentRole, loaded.AgentRole)
	assert.Equal(t, state.Status, loaded.Status)
	assert.Equal(t, "value", loaded.Context["key"])
}

func TestStore_WorkflowLoadNonexistent(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)

	_, err := store.LoadWorkflow()
	assert.Error(t, err)
}
