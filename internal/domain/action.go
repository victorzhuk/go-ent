package domain

// SpecAction defines the type of action to perform on a specification.
type SpecAction string

// ActionPhase represents the phase of the development lifecycle.
type ActionPhase string

const (
	ActionPhaseDiscovery  ActionPhase = "discovery"
	ActionPhasePlanning   ActionPhase = "planning"
	ActionPhaseExecution  ActionPhase = "execution"
	ActionPhaseValidation ActionPhase = "validation"
	ActionPhaseLifecycle  ActionPhase = "lifecycle"
)

const (
	SpecActionResearch  SpecAction = "research"
	SpecActionAnalyze   SpecAction = "analyze"
	SpecActionRetrofit  SpecAction = "retrofit"
	SpecActionProposal  SpecAction = "proposal"
	SpecActionPlan      SpecAction = "plan"
	SpecActionDesign    SpecAction = "design"
	SpecActionSplit     SpecAction = "split"
	SpecActionImplement SpecAction = "implement"
	SpecActionExecute   SpecAction = "execute"
	SpecActionScaffold  SpecAction = "scaffold"
	SpecActionReview    SpecAction = "review"
	SpecActionVerify    SpecAction = "verify"
	SpecActionDebug     SpecAction = "debug"
	SpecActionLint      SpecAction = "lint"
	SpecActionApprove   SpecAction = "approve"
	SpecActionArchive   SpecAction = "archive"
	SpecActionStatus    SpecAction = "status"
)

// String returns the string representation of the action.
func (a SpecAction) String() string {
	return string(a)
}

// Valid returns true if the action is valid.
func (a SpecAction) Valid() bool {
	switch a {
	case SpecActionResearch, SpecActionAnalyze, SpecActionRetrofit,
		SpecActionProposal, SpecActionPlan, SpecActionDesign, SpecActionSplit,
		SpecActionImplement, SpecActionExecute, SpecActionScaffold,
		SpecActionReview, SpecActionVerify, SpecActionDebug, SpecActionLint,
		SpecActionApprove, SpecActionArchive, SpecActionStatus:
		return true
	default:
		return false
	}
}

// Phase returns the development phase for this action.
func (a SpecAction) Phase() ActionPhase {
	switch a {
	case SpecActionResearch, SpecActionAnalyze, SpecActionRetrofit:
		return ActionPhaseDiscovery
	case SpecActionProposal, SpecActionPlan, SpecActionDesign, SpecActionSplit:
		return ActionPhasePlanning
	case SpecActionImplement, SpecActionExecute, SpecActionScaffold:
		return ActionPhaseExecution
	case SpecActionReview, SpecActionVerify, SpecActionDebug, SpecActionLint:
		return ActionPhaseValidation
	case SpecActionApprove, SpecActionArchive, SpecActionStatus:
		return ActionPhaseLifecycle
	default:
		return ""
	}
}

// String returns the string representation of the action phase.
func (p ActionPhase) String() string {
	return string(p)
}

// Valid returns true if the action phase is valid.
func (p ActionPhase) Valid() bool {
	switch p {
	case ActionPhaseDiscovery, ActionPhasePlanning, ActionPhaseExecution,
		ActionPhaseValidation, ActionPhaseLifecycle:
		return true
	default:
		return false
	}
}
