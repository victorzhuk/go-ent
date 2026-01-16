package integration

//nolint:gosec // test file with necessary file operations

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/victorzhuk/go-ent/internal/agent"
	"github.com/victorzhuk/go-ent/internal/config"
	"github.com/victorzhuk/go-ent/internal/domain"
	"github.com/victorzhuk/go-ent/internal/skill"
	"github.com/victorzhuk/go-ent/internal/spec"
)

func TestAgentSelector_WithConfig(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create config
	cfg := &config.Config{
		Version: "1.0",
		Agents: config.AgentsConfig{
			Default: domain.AgentRoleDeveloper,
			Roles: map[string]config.AgentRoleConfig{
				"developer": {
					Model:       "sonnet",
					Skills:      []string{"go-code", "go-test"},
					BudgetLimit: 100.0,
				},
				"architect": {
					Model:       "opus",
					Skills:      []string{"go-arch", "go-api"},
					BudgetLimit: 200.0,
				},
			},
		},
		Runtime: config.RuntimeConfig{
			Preferred: domain.RuntimeClaudeCode,
		},
		Budget: config.BudgetConfig{
			Daily:    1000.0,
			Monthly:  10000.0,
			PerTask:  100.0,
			Tracking: true,
		},
		Models: config.ModelsConfig{
			"opus":   "claude-opus-4-5",
			"sonnet": "claude-sonnet-4-5",
			"haiku":  "claude-haiku-4",
		},
	}

	// Save config
	store := spec.NewStore(tmpDir)
	err := store.SaveConfig(cfg)
	require.NoError(t, err)

	// Load config
	loadedCfg, err := store.LoadConfig()
	require.NoError(t, err)

	// Create selector with config
	registry := skill.NewRegistry()
	selector := agent.NewSelector(agent.Config{
		MaxBudget:  int(loadedCfg.Budget.PerTask * 1000), // Convert to tokens
		StrictMode: false,
	}, registry)

	// Test selection
	task := agent.Task{
		Description: "Add new REST endpoint",
		Type:        agent.TaskTypeFeature,
		Action:      domain.SpecActionImplement,
		Phase:       domain.ActionPhaseExecution,
		Files:       []string{"handler.go", "service.go"},
	}

	result, err := selector.Select(context.Background(), task)
	require.NoError(t, err)
	assert.Equal(t, domain.AgentRoleDeveloper, result.Role)
	assert.Equal(t, "sonnet", result.Model)
}

func TestSkillRegistry_WithConfigPaths(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	store := spec.NewStore(tmpDir)

	// Create skills directory structure
	skillsPath := store.SkillsPath()
	require.NoError(t, os.MkdirAll(filepath.Join(skillsPath, "go-code"), 0750))

	// Create a test skill
	skillContent := `---
name: go-code
description: "Go code implementation. Auto-activates for: implement, code, go."
---

# Go Code Skill

Handles Go code implementation tasks.
`
	skillFile := filepath.Join(skillsPath, "go-code", "SKILL.md")
	require.NoError(t, os.WriteFile(skillFile, []byte(skillContent), 0600))

	// Load skills
	registry := skill.NewRegistry()
	err := registry.Load(skillsPath)
	require.NoError(t, err)

	// Verify skill was loaded
	all := registry.All()
	assert.Len(t, all, 1)
	assert.Equal(t, "go-code", all[0].Name)

	// Test skill matching
	ctx := domain.SkillContext{
		Action: domain.SpecActionImplement,
		Agent:  domain.AgentRoleDeveloper,
	}
	matched := registry.MatchForContext(ctx)
	assert.Contains(t, matched, "go-code")
}

func TestWorkflowState_AgentTracking(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	store := spec.NewStore(tmpDir)

	// Initialize spec directory
	err := store.Init(spec.Project{
		Name:        "test-project",
		Module:      "github.com/test/project",
		Description: "Test project",
	})
	require.NoError(t, err)

	// Create workflow state
	state := spec.NewWorkflowState("test-change", "execution")
	state.SetAgent(domain.AgentRoleArchitect)

	// Save workflow
	err = store.SaveWorkflow(state)
	require.NoError(t, err)

	// Load workflow
	loaded, err := store.LoadWorkflow()
	require.NoError(t, err)

	assert.Equal(t, domain.AgentRoleArchitect, loaded.AgentRole)
	assert.Equal(t, "execution", loaded.Phase)
	assert.Equal(t, spec.WorkflowStatusActive, loaded.Status)

	// Update agent
	loaded.SetAgent(domain.AgentRoleDeveloper)
	err = store.SaveWorkflow(loaded)
	require.NoError(t, err)

	// Reload and verify
	reloaded, err := store.LoadWorkflow()
	require.NoError(t, err)
	assert.Equal(t, domain.AgentRoleDeveloper, reloaded.AgentRole)
}

func TestStore_AgentAndSkillPaths(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	store := spec.NewStore(tmpDir)

	// Verify paths are correctly constructed
	agentsPath := store.AgentsPath()
	skillsPath := store.SkillsPath()

	expectedAgents := filepath.Join(tmpDir, "plugins", "go-ent", "agents")
	expectedSkills := filepath.Join(tmpDir, "plugins", "go-ent", "skills")

	assert.Equal(t, expectedAgents, agentsPath)
	assert.Equal(t, expectedSkills, skillsPath)

	// Create directories
	require.NoError(t, os.MkdirAll(agentsPath, 0750))
	require.NoError(t, os.MkdirAll(skillsPath, 0750))

	// Verify they exist
	_, err := os.Stat(agentsPath)
	assert.NoError(t, err)

	_, err = os.Stat(skillsPath)
	assert.NoError(t, err)
}

func TestAgentDelegation_WithConfig(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		Version: "1.0",
		Agents: config.AgentsConfig{
			Default: domain.AgentRoleDeveloper,
			Roles: map[string]config.AgentRoleConfig{
				"architect": {Model: "opus"},
				"senior":    {Model: "sonnet"},
				"developer": {Model: "sonnet"},
				"reviewer":  {Model: "opus"},
			},
			Delegation: config.DelegationConfig{
				Auto:             true,
				ApprovalRequired: []domain.AgentRole{domain.AgentRoleArchitect},
			},
		},
		Runtime: config.RuntimeConfig{
			Preferred: domain.RuntimeClaudeCode,
		},
		Budget: config.BudgetConfig{
			Daily:    1000.0,
			Monthly:  10000.0,
			PerTask:  100.0,
			Tracking: true,
		},
		Models: config.ModelsConfig{
			"opus":   "claude-opus-4-5",
			"sonnet": "claude-sonnet-4-5",
		},
	}

	// Validate config
	err := cfg.Validate()
	require.NoError(t, err)

	// Create delegator
	delegator := agent.NewDelegator()

	// Test delegation for feature task
	task := agent.Task{
		Description: "Design new microservice architecture",
		Type:        agent.TaskTypeFeature,
	}

	chain, err := delegator.GetDelegationChain(task)
	require.NoError(t, err)
	assert.Contains(t, chain, domain.AgentRoleArchitect)
	assert.Contains(t, chain, domain.AgentRoleSenior)
	assert.Contains(t, chain, domain.AgentRoleDeveloper)
	assert.Contains(t, chain, domain.AgentRoleReviewer)

	// Verify architect can delegate to senior
	canDelegate := delegator.CanDelegate(domain.AgentRoleArchitect, domain.AgentRoleSenior)
	assert.True(t, canDelegate)

	// Verify approval required for architect (from config)
	assert.Contains(t, cfg.Agents.Delegation.ApprovalRequired, domain.AgentRoleArchitect)
}

func TestCompleteWorkflow_WithAllComponents(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	store := spec.NewStore(tmpDir)

	// Initialize project
	err := store.Init(spec.Project{
		Name:        "integration-test",
		Module:      "github.com/test/integration",
		Description: "Integration test project",
	})
	require.NoError(t, err)

	// Create config
	cfg := &config.Config{
		Version: "1.0",
		Agents: config.AgentsConfig{
			Default: domain.AgentRoleDeveloper,
			Roles: map[string]config.AgentRoleConfig{
				"developer": {
					Model:       "sonnet",
					Skills:      []string{"go-code"},
					BudgetLimit: 50.0,
				},
			},
		},
		Runtime: config.RuntimeConfig{
			Preferred: domain.RuntimeClaudeCode,
		},
		Budget: config.BudgetConfig{
			Daily:    1000.0,
			Monthly:  10000.0,
			PerTask:  100.0,
			Tracking: true,
		},
		Models: config.ModelsConfig{
			"sonnet": "claude-sonnet-4-5",
		},
	}

	err = store.SaveConfig(cfg)
	require.NoError(t, err)

	// Create skill
	skillsPath := store.SkillsPath()
	require.NoError(t, os.MkdirAll(filepath.Join(skillsPath, "go-code"), 0750))

	skillContent := `---
name: go-code
description: "Go implementation. Auto-activates for: implement, code."
---`
	require.NoError(t, os.WriteFile(
		filepath.Join(skillsPath, "go-code", "SKILL.md"),
		[]byte(skillContent),
		0600,
	))

	// Initialize components
	registry := skill.NewRegistry()
	err = registry.Load(skillsPath)
	require.NoError(t, err)

	selector := agent.NewSelector(agent.Config{
		MaxBudget:  50000, // 50k tokens
		StrictMode: false,
	}, registry)

	// Create workflow
	workflow := spec.NewWorkflowState("test-change", "execution")

	// Select agent for task
	task := agent.Task{
		Description: "Implement user service",
		Type:        agent.TaskTypeFeature,
		Action:      domain.SpecActionImplement,
		Phase:       domain.ActionPhaseExecution,
		Files:       []string{"user.go"},
	}

	result, err := selector.Select(context.Background(), task)
	require.NoError(t, err)

	// Update workflow with selected agent
	workflow.SetAgent(result.Role)

	// Save workflow
	err = store.SaveWorkflow(workflow)
	require.NoError(t, err)

	// Load and verify complete state
	loadedWorkflow, err := store.LoadWorkflow()
	require.NoError(t, err)
	assert.Equal(t, domain.AgentRoleDeveloper, loadedWorkflow.AgentRole)
	assert.Equal(t, spec.WorkflowStatusActive, loadedWorkflow.Status)

	loadedCfg, err := store.LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, domain.AgentRoleDeveloper, loadedCfg.Agents.Default)

	// Verify skill matching works
	skillCtx := domain.SkillContext{
		Action: domain.SpecActionImplement,
		Agent:  result.Role,
	}
	matched := registry.MatchForContext(skillCtx)
	assert.Contains(t, matched, "go-code")
}
