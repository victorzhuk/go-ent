package domain

// SpecAction defines the type of action to perform on a specification.
type SpecAction string

// ActionPhase represents the phase of the development lifecycle.
type ActionPhase string

// Action phase constants define the development lifecycle stages.
const (
	// ActionPhaseDiscovery represents the research and analysis phase.
	// Focuses on understanding the problem space and gathering information.
	ActionPhaseDiscovery ActionPhase = "discovery"

	// ActionPhasePlanning represents the design and planning phase.
	// Focuses on creating proposals, designs, and breaking down work.
	ActionPhasePlanning ActionPhase = "planning"

	// ActionPhaseExecution represents the implementation phase.
	// Focuses on writing code and building features.
	ActionPhaseExecution ActionPhase = "execution"

	// ActionPhaseValidation represents the quality assurance phase.
	// Focuses on testing, reviewing, and verifying correctness.
	ActionPhaseValidation ActionPhase = "validation"

	// ActionPhaseLifecycle represents the management phase.
	// Focuses on workflow state transitions and status tracking.
	ActionPhaseLifecycle ActionPhase = "lifecycle"
)

// SpecAction constants define the available actions organized by phase.
const (
	// Discovery phase actions
	// SpecActionResearch performs research on a topic or technology.
	SpecActionResearch SpecAction = "research"
	// SpecActionAnalyze analyzes existing code or systems.
	SpecActionAnalyze SpecAction = "analyze"
	// SpecActionRetrofit analyzes existing code to create specifications.
	SpecActionRetrofit SpecAction = "retrofit"

	// Planning phase actions
	// SpecActionProposal creates a new change proposal.
	SpecActionProposal SpecAction = "proposal"
	// SpecActionPlan creates a comprehensive implementation plan.
	SpecActionPlan SpecAction = "plan"
	// SpecActionDesign creates detailed technical designs.
	SpecActionDesign SpecAction = "design"
	// SpecActionSplit breaks down large changes into smaller ones.
	SpecActionSplit SpecAction = "split"

	// Execution phase actions
	// SpecActionImplement executes implementation tasks.
	SpecActionImplement SpecAction = "implement"
	// SpecActionExecute runs a specific task or command.
	SpecActionExecute SpecAction = "execute"
	// SpecActionScaffold generates code from templates.
	SpecActionScaffold SpecAction = "scaffold"

	// Validation phase actions
	// SpecActionReview performs code review.
	SpecActionReview SpecAction = "review"
	// SpecActionVerify checks correctness and completeness.
	SpecActionVerify SpecAction = "verify"
	// SpecActionDebug investigates and fixes issues.
	SpecActionDebug SpecAction = "debug"
	// SpecActionLint runs static analysis and formatting checks.
	SpecActionLint SpecAction = "lint"

	// Lifecycle phase actions
	// SpecActionApprove approves a proposal or change.
	SpecActionApprove SpecAction = "approve"
	// SpecActionArchive archives a completed change.
	SpecActionArchive SpecAction = "archive"
	// SpecActionStatus displays current workflow state.
	SpecActionStatus SpecAction = "status"
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
