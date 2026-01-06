package domain_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/victorzhuk/go-ent/internal/domain"
)

// ExampleAgentRole demonstrates validating and using agent roles.
func ExampleAgentRole() {
	role := domain.AgentRoleDeveloper

	if !role.Valid() {
		fmt.Printf("invalid role: %s\n", role)
		return
	}

	fmt.Printf("Role: %s\n", role)
	fmt.Printf("Valid: %v\n", role.Valid())

	// Output:
	// Role: developer
	// Valid: true
}

// ExampleAgentRole_validation demonstrates role validation.
func ExampleAgentRole_validation() {
	validRole := domain.AgentRoleArchitect
	invalidRole := domain.AgentRole("invalid")

	fmt.Printf("Valid role '%s': %v\n", validRole, validRole.Valid())
	fmt.Printf("Invalid role '%s': %v\n", invalidRole, invalidRole.Valid())

	// Output:
	// Valid role 'architect': true
	// Invalid role 'invalid': false
}

// ExampleSpecAction demonstrates action classification by phase.
func ExampleSpecAction() {
	action := domain.SpecActionImplement
	phase := action.Phase()

	fmt.Printf("Action: %s\n", action)
	fmt.Printf("Phase: %s\n", phase)

	switch phase {
	case domain.ActionPhaseDiscovery:
		fmt.Println("Route to research agent")
	case domain.ActionPhaseExecution:
		fmt.Println("Route to developer agent")
	}

	// Output:
	// Action: implement
	// Phase: execution
	// Route to developer agent
}

// ExampleSpecAction_Phase demonstrates routing actions by phase.
func ExampleSpecAction_Phase() {
	actions := []domain.SpecAction{
		domain.SpecActionResearch,
		domain.SpecActionProposal,
		domain.SpecActionImplement,
		domain.SpecActionReview,
	}

	for _, action := range actions {
		fmt.Printf("%s -> %s\n", action, action.Phase())
	}

	// Output:
	// research -> discovery
	// proposal -> planning
	// implement -> execution
	// review -> validation
}

// ExampleSkill demonstrates implementing a custom skill.
func ExampleSkill() {
	skill := &customSkill{name: "go-code"}

	ctx := domain.SkillContext{
		Action:  domain.SpecActionImplement,
		Runtime: domain.RuntimeClaudeCode,
		Agent:   domain.AgentRoleDeveloper,
		Metadata: map[string]interface{}{
			"language": "go",
		},
	}

	if skill.CanHandle(ctx) {
		fmt.Printf("Skill '%s' can handle context\n", skill.Name())
	}

	req := domain.SkillRequest{
		Input: "package main",
		Parameters: map[string]interface{}{
			"file": "main.go",
		},
		Context: ctx,
	}

	result, err := skill.Execute(context.Background(), req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Success: %v\n", result.Success)

	// Output:
	// Skill 'go-code' can handle context
	// Success: true
}

// customSkill is an example Skill implementation.
type customSkill struct {
	name string
}

func (s *customSkill) Name() string {
	return s.name
}

func (s *customSkill) Description() string {
	return "Go code implementation skill"
}

func (s *customSkill) CanHandle(ctx domain.SkillContext) bool {
	lang, ok := ctx.Metadata["language"].(string)
	return ok && lang == "go"
}

func (s *customSkill) Execute(ctx context.Context, req domain.SkillRequest) (domain.SkillResult, error) {
	return domain.SkillResult{
		Success: true,
		Output:  "formatted code",
		Metadata: map[string]interface{}{
			"formatted": true,
		},
	}, nil
}

// ExampleAgentError demonstrates error handling with domain errors.
func ExampleAgentError() {
	err := &domain.AgentError{
		Role: domain.AgentRoleDeveloper,
		Err:  domain.ErrInvalidAgentConfig,
	}

	fmt.Println(err.Error())

	// Check error type
	if domain.IsAgentError(err) {
		var ae *domain.AgentError
		if errors.As(err, &ae) {
			fmt.Printf("Agent role: %s\n", ae.Role)
		}
	}

	// Check sentinel error
	if errors.Is(err, domain.ErrInvalidAgentConfig) {
		fmt.Println("Invalid config detected")
	}

	// Output:
	// agent error [developer]: invalid agent config
	// Agent role: developer
	// Invalid config detected
}

// ExampleExecutionContext demonstrates creating execution context.
func ExampleExecutionContext() {
	ctx := &domain.ExecutionContext{
		Runtime:  domain.RuntimeClaudeCode,
		Agent:    domain.AgentRoleDeveloper,
		Strategy: domain.ExecutionStrategySingle,
		ChangeID: "add-feature",
		TaskID:   "1.2",
	}

	fmt.Printf("Runtime: %s\n", ctx.Runtime)
	fmt.Printf("Agent: %s\n", ctx.Agent)
	fmt.Printf("Strategy: %s\n", ctx.Strategy)
	fmt.Printf("Change: %s, Task: %s\n", ctx.ChangeID, ctx.TaskID)

	// Output:
	// Runtime: claude-code
	// Agent: developer
	// Strategy: single
	// Change: add-feature, Task: 1.2
}
