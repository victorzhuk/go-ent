package domain

//nolint:gosec // test file with necessary file operations

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSpecAction_String(t *testing.T) {
	tests := []struct {
		name   string
		action SpecAction
		want   string
	}{
		{"research", SpecActionResearch, "research"},
		{"analyze", SpecActionAnalyze, "analyze"},
		{"proposal", SpecActionProposal, "proposal"},
		{"implement", SpecActionImplement, "implement"},
		{"review", SpecActionReview, "review"},
		{"archive", SpecActionArchive, "archive"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.action.String())
		})
	}
}

func TestSpecAction_Valid(t *testing.T) {
	tests := []struct {
		name   string
		action SpecAction
		want   bool
	}{
		{"valid research", SpecActionResearch, true},
		{"valid analyze", SpecActionAnalyze, true},
		{"valid retrofit", SpecActionRetrofit, true},
		{"valid proposal", SpecActionProposal, true},
		{"valid plan", SpecActionPlan, true},
		{"valid design", SpecActionDesign, true},
		{"valid split", SpecActionSplit, true},
		{"valid implement", SpecActionImplement, true},
		{"valid execute", SpecActionExecute, true},
		{"valid scaffold", SpecActionScaffold, true},
		{"valid review", SpecActionReview, true},
		{"valid verify", SpecActionVerify, true},
		{"valid debug", SpecActionDebug, true},
		{"valid lint", SpecActionLint, true},
		{"valid approve", SpecActionApprove, true},
		{"valid archive", SpecActionArchive, true},
		{"valid status", SpecActionStatus, true},
		{"invalid empty", SpecAction(""), false},
		{"invalid unknown", SpecAction("unknown"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.action.Valid())
		})
	}
}

func TestSpecAction_Phase(t *testing.T) {
	tests := []struct {
		name   string
		action SpecAction
		want   ActionPhase
	}{
		{"research is discovery", SpecActionResearch, ActionPhaseDiscovery},
		{"analyze is discovery", SpecActionAnalyze, ActionPhaseDiscovery},
		{"retrofit is discovery", SpecActionRetrofit, ActionPhaseDiscovery},
		{"proposal is planning", SpecActionProposal, ActionPhasePlanning},
		{"plan is planning", SpecActionPlan, ActionPhasePlanning},
		{"design is planning", SpecActionDesign, ActionPhasePlanning},
		{"split is planning", SpecActionSplit, ActionPhasePlanning},
		{"implement is execution", SpecActionImplement, ActionPhaseExecution},
		{"execute is execution", SpecActionExecute, ActionPhaseExecution},
		{"scaffold is execution", SpecActionScaffold, ActionPhaseExecution},
		{"review is validation", SpecActionReview, ActionPhaseValidation},
		{"verify is validation", SpecActionVerify, ActionPhaseValidation},
		{"debug is validation", SpecActionDebug, ActionPhaseValidation},
		{"lint is validation", SpecActionLint, ActionPhaseValidation},
		{"approve is lifecycle", SpecActionApprove, ActionPhaseLifecycle},
		{"archive is lifecycle", SpecActionArchive, ActionPhaseLifecycle},
		{"status is lifecycle", SpecActionStatus, ActionPhaseLifecycle},
		{"invalid returns empty", SpecAction("invalid"), ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.action.Phase())
		})
	}
}

func TestActionPhase_String(t *testing.T) {
	tests := []struct {
		name  string
		phase ActionPhase
		want  string
	}{
		{"discovery", ActionPhaseDiscovery, "discovery"},
		{"planning", ActionPhasePlanning, "planning"},
		{"execution", ActionPhaseExecution, "execution"},
		{"validation", ActionPhaseValidation, "validation"},
		{"lifecycle", ActionPhaseLifecycle, "lifecycle"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.phase.String())
		})
	}
}

func TestActionPhase_Valid(t *testing.T) {
	tests := []struct {
		name  string
		phase ActionPhase
		want  bool
	}{
		{"valid discovery", ActionPhaseDiscovery, true},
		{"valid planning", ActionPhasePlanning, true},
		{"valid execution", ActionPhaseExecution, true},
		{"valid validation", ActionPhaseValidation, true},
		{"valid lifecycle", ActionPhaseLifecycle, true},
		{"invalid empty", ActionPhase(""), false},
		{"invalid unknown", ActionPhase("unknown"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.phase.Valid())
		})
	}
}

func TestAllActionsHaveValidPhase(t *testing.T) {
	actions := []SpecAction{
		SpecActionResearch, SpecActionAnalyze, SpecActionRetrofit,
		SpecActionProposal, SpecActionPlan, SpecActionDesign, SpecActionSplit,
		SpecActionImplement, SpecActionExecute, SpecActionScaffold,
		SpecActionReview, SpecActionVerify, SpecActionDebug, SpecActionLint,
		SpecActionApprove, SpecActionArchive, SpecActionStatus,
	}

	for _, action := range actions {
		t.Run(string(action), func(t *testing.T) {
			phase := action.Phase()
			assert.NotEmpty(t, phase, "action %s should have a phase", action)
			assert.True(t, phase.Valid(), "phase for action %s should be valid", action)
		})
	}
}
