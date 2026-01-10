package agent

import (
	"fmt"

	"github.com/victorzhuk/go-ent/internal/domain"
)

// Delegator handles agent delegation decisions based on task complexity and type.
type Delegator struct {
	rules map[TaskType][]domain.AgentRole
}

// NewDelegator creates a new delegator with standard workflow rules.
func NewDelegator() *Delegator {
	return &Delegator{
		rules: map[TaskType][]domain.AgentRole{
			TaskTypeFeature: {
				domain.AgentRoleArchitect,
				domain.AgentRoleSenior,
				domain.AgentRoleDeveloper,
				domain.AgentRoleReviewer,
			},
			TaskTypeBugFix: {
				domain.AgentRoleSenior,
				domain.AgentRoleDeveloper,
				domain.AgentRoleReviewer,
			},
			TaskTypeRefactor: {
				domain.AgentRoleSenior,
				domain.AgentRoleDeveloper,
				domain.AgentRoleReviewer,
			},
			TaskTypeTest: {
				domain.AgentRoleDeveloper,
				domain.AgentRoleReviewer,
			},
			TaskTypeDocumentation: {
				domain.AgentRoleDeveloper,
			},
			TaskTypeArchitecture: {
				domain.AgentRoleArchitect,
				domain.AgentRoleSenior,
				domain.AgentRoleReviewer,
			},
		},
	}
}

// CanDelegate checks if the 'from' role can delegate to the 'to' role.
func (d *Delegator) CanDelegate(from, to domain.AgentRole) bool {
	delegationMap := map[domain.AgentRole][]domain.AgentRole{
		domain.AgentRoleProduct: {
			domain.AgentRoleArchitect,
			domain.AgentRoleSenior,
			domain.AgentRoleDeveloper,
		},
		domain.AgentRoleArchitect: {
			domain.AgentRoleSenior,
			domain.AgentRoleDeveloper,
		},
		domain.AgentRoleSenior: {
			domain.AgentRoleDeveloper,
			domain.AgentRoleReviewer,
		},
		domain.AgentRoleDeveloper: {
			domain.AgentRoleReviewer,
		},
		domain.AgentRoleReviewer: {},
		domain.AgentRoleOps: {
			domain.AgentRoleSenior,
			domain.AgentRoleDeveloper,
		},
	}

	allowedTargets, exists := delegationMap[from]
	if !exists {
		return false
	}

	for _, target := range allowedTargets {
		if target == to {
			return true
		}
	}

	return false
}

// GetDelegationChain returns the complete workflow chain for a given task.
func (d *Delegator) GetDelegationChain(task Task) ([]domain.AgentRole, error) {
	if err := task.Validate(); err != nil {
		return nil, fmt.Errorf("invalid task: %w", err)
	}

	chain, exists := d.rules[task.Type]
	if !exists {
		return []domain.AgentRole{domain.AgentRoleDeveloper}, nil
	}

	result := make([]domain.AgentRole, len(chain))
	copy(result, chain)
	return result, nil
}

// GetNextAgent returns the next agent in the workflow after the given role.
func (d *Delegator) GetNextAgent(task Task, current domain.AgentRole) (domain.AgentRole, error) {
	chain, err := d.GetDelegationChain(task)
	if err != nil {
		return "", err
	}

	for i, role := range chain {
		if role == current && i+1 < len(chain) {
			return chain[i+1], nil
		}
	}

	return "", fmt.Errorf("no next agent for role %s in task type %s", current, task.Type)
}

// ShouldDelegate determines if delegation is needed based on complexity.
func (d *Delegator) ShouldDelegate(task Task, current domain.AgentRole) bool {
	chain, err := d.GetDelegationChain(task)
	if err != nil {
		return false
	}

	for _, role := range chain {
		if role == current {
			return true
		}
	}

	return false
}
