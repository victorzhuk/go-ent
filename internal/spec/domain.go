package spec

import (
	_ "github.com/victorzhuk/go-ent/internal/domain" // Domain types for agent system
)

// Domain Boundaries:
//
// This package (internal/spec) defines types for the OpenSpec system:
// - Change and task lifecycle (ChangeStatus, TaskStatus)
// - Delta operations (OpAdded, OpModified, etc.)
// - Project and spec metadata (Project, ListItem)
//
// The internal/domain package defines types for the multi-agent execution system:
// - Agent roles and capabilities (AgentRole, AgentConfig, AgentCapability)
// - Runtime environments (Runtime, RuntimeCapability)
// - Execution semantics (ExecutionStrategy, ExecutionContext, ExecutionResult)
// - Actions and phases (SpecAction, ActionPhase)
// - Skill abstraction (Skill interface and supporting types)
//
// Relationship:
// - Spec domain: WHAT needs to be done (proposals, tasks, changes)
// - Agent domain: WHO does it and HOW it's executed (agents, runtimes, strategies)
// - WorkflowState (in workflow.go) bridges these domains by tracking which agent
//   is executing which spec-defined task

type ChangeStatus string

const (
	StatusDraft    ChangeStatus = "draft"
	StatusActive   ChangeStatus = "active"
	StatusApproved ChangeStatus = "approved"
	StatusArchived ChangeStatus = "archived"
)

type TaskStatus string

const (
	TaskPending    TaskStatus = "pending"
	TaskInProgress TaskStatus = "in_progress"
	TaskCompleted  TaskStatus = "completed"
)

type DeltaOperation string

const (
	OpAdded    DeltaOperation = "ADDED"
	OpModified DeltaOperation = "MODIFIED"
	OpRemoved  DeltaOperation = "REMOVED"
	OpRenamed  DeltaOperation = "RENAMED"
)

type Project struct {
	Name        string            `yaml:"name"`
	Module      string            `yaml:"module"`
	Description string            `yaml:"description"`
	Conventions map[string]string `yaml:"conventions"`
}

type ListItem struct {
	ID          string
	Name        string
	Type        string
	Status      string
	Path        string
	Description string
}
