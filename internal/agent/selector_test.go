package agent

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/victorzhuk/go-ent/internal/domain"
)

type mockSkillRegistry struct {
	skills []string
}

func (m *mockSkillRegistry) MatchForContext(ctx domain.SkillContext) []string {
	return m.skills
}

func TestSelector_Select(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		task           Task
		maxBudget      int
		strictMode     bool
		mockSkills     []string
		expectedRole   domain.AgentRole
		expectedModel  string
		expectedSkills []string
		wantErr        bool
	}{
		{
			name: "architectural task selects architect with opus",
			task: Task{
				Description: "Design new microservices architecture",
				Type:        TaskTypeArchitecture,
				Action:      domain.SpecActionPlan,
				Phase:       domain.ActionPhasePlanning,
			},
			mockSkills:     []string{"go-arch", "go-api"},
			expectedRole:   domain.AgentRoleArchitect,
			expectedModel:  "opus",
			expectedSkills: []string{"go-arch", "go-api"},
		},
		{
			name: "complex refactor selects senior with opus",
			task: Task{
				Description: "Refactor database layer with multiple integrations and migrations",
				Type:        TaskTypeRefactor,
				Action:      domain.SpecActionImplement,
				Phase:       domain.ActionPhaseExecution,
				Files:       []string{"repo1.go", "repo2.go", "repo3.go", "repo4.go", "repo5.go", "repo6.go"},
			},
			mockSkills:     []string{"go-code", "go-db"},
			expectedRole:   domain.AgentRoleSenior,
			expectedModel:  "opus",
			expectedSkills: []string{"go-code", "go-db"},
		},
		{
			name: "moderate feature selects developer with sonnet",
			task: Task{
				Description: "Add new REST endpoint with validation",
				Type:        TaskTypeFeature,
				Action:      domain.SpecActionImplement,
				Phase:       domain.ActionPhaseExecution,
				Files:       []string{"handler.go", "service.go"},
			},
			mockSkills:     []string{"go-code", "go-api"},
			expectedRole:   domain.AgentRoleDeveloper,
			expectedModel:  "sonnet",
			expectedSkills: []string{"go-code", "go-api"},
		},
		{
			name: "simple test selects developer with haiku",
			task: Task{
				Description: "Add unit test for existing function",
				Type:        TaskTypeTest,
				Action:      domain.SpecActionImplement,
				Phase:       domain.ActionPhaseExecution,
				Files:       []string{"user_test.go"},
			},
			mockSkills:     []string{"go-test"},
			expectedRole:   domain.AgentRoleDeveloper,
			expectedModel:  "haiku",
			expectedSkills: []string{"go-test"},
		},
		{
			name: "trivial documentation selects developer with haiku",
			task: Task{
				Description: "Fix typo in README",
				Type:        TaskTypeDocumentation,
				Action:      domain.SpecActionImplement,
				Phase:       domain.ActionPhaseExecution,
			},
			mockSkills:     []string{},
			expectedRole:   domain.AgentRoleDeveloper,
			expectedModel:  "haiku",
			expectedSkills: []string{},
		},
		{
			name: "invalid task returns error",
			task: Task{
				Description: "",
				Type:        TaskTypeFeature,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			registry := &mockSkillRegistry{skills: tt.mockSkills}
			selector := NewSelector(Config{
				MaxBudget:  tt.maxBudget,
				StrictMode: tt.strictMode,
			}, registry)

			result, err := selector.Select(context.Background(), tt.task)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedRole, result.Role)
			assert.Equal(t, tt.expectedModel, result.Model)
			assert.Equal(t, tt.expectedSkills, result.Skills)
			assert.NotEmpty(t, result.Reason)
		})
	}
}

func TestSelector_selectRole(t *testing.T) {
	t.Parallel()

	registry := &mockSkillRegistry{}
	selector := NewSelector(Config{}, registry)

	tests := []struct {
		name         string
		complexity   TaskComplexity
		expectedRole domain.AgentRole
	}{
		{
			name:         "architectural complexity",
			complexity:   TaskComplexity{Level: ComplexityArchitectural},
			expectedRole: domain.AgentRoleArchitect,
		},
		{
			name:         "complex task",
			complexity:   TaskComplexity{Level: ComplexityComplex},
			expectedRole: domain.AgentRoleSenior,
		},
		{
			name:         "moderate task",
			complexity:   TaskComplexity{Level: ComplexityModerate},
			expectedRole: domain.AgentRoleDeveloper,
		},
		{
			name:         "simple task",
			complexity:   TaskComplexity{Level: ComplexitySimple},
			expectedRole: domain.AgentRoleDeveloper,
		},
		{
			name:         "trivial task",
			complexity:   TaskComplexity{Level: ComplexityTrivial},
			expectedRole: domain.AgentRoleDeveloper,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			role := selector.selectRole(tt.complexity)
			assert.Equal(t, tt.expectedRole, role)
		})
	}
}

func TestSelector_selectModel(t *testing.T) {
	t.Parallel()

	registry := &mockSkillRegistry{}
	selector := NewSelector(Config{}, registry)

	tests := []struct {
		name          string
		role          domain.AgentRole
		complexity    TaskComplexity
		expectedModel string
	}{
		{
			name:          "architect always uses opus",
			role:          domain.AgentRoleArchitect,
			complexity:    TaskComplexity{Level: ComplexityModerate},
			expectedModel: "opus",
		},
		{
			name:          "senior uses opus for complex tasks",
			role:          domain.AgentRoleSenior,
			complexity:    TaskComplexity{Level: ComplexityComplex},
			expectedModel: "opus",
		},
		{
			name:          "senior uses sonnet for moderate tasks",
			role:          domain.AgentRoleSenior,
			complexity:    TaskComplexity{Level: ComplexityModerate},
			expectedModel: "sonnet",
		},
		{
			name:          "developer uses sonnet for moderate tasks",
			role:          domain.AgentRoleDeveloper,
			complexity:    TaskComplexity{Level: ComplexityModerate},
			expectedModel: "sonnet",
		},
		{
			name:          "developer uses haiku for simple tasks",
			role:          domain.AgentRoleDeveloper,
			complexity:    TaskComplexity{Level: ComplexitySimple},
			expectedModel: "haiku",
		},
		{
			name:          "reviewer always uses opus",
			role:          domain.AgentRoleReviewer,
			complexity:    TaskComplexity{Level: ComplexitySimple},
			expectedModel: "opus",
		},
		{
			name:          "ops uses sonnet",
			role:          domain.AgentRoleOps,
			complexity:    TaskComplexity{Level: ComplexityModerate},
			expectedModel: "sonnet",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			model := selector.selectModel(tt.role, tt.complexity)
			assert.Equal(t, tt.expectedModel, model)
		})
	}
}

func TestSelector_matchSkills(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		task           Task
		role           domain.AgentRole
		mockSkills     []string
		expectedSkills []string
	}{
		{
			name: "matches skills based on action and phase",
			task: Task{
				Action: domain.SpecActionImplement,
				Phase:  domain.ActionPhaseExecution,
				Metadata: map[string]interface{}{
					"language": "go",
				},
			},
			role:           domain.AgentRoleDeveloper,
			mockSkills:     []string{"go-code", "go-test"},
			expectedSkills: []string{"go-code", "go-test"},
		},
		{
			name: "empty when no skills match",
			task: Task{
				Action: domain.SpecActionPlan,
				Phase:  domain.ActionPhasePlanning,
			},
			role:           domain.AgentRoleArchitect,
			mockSkills:     []string{},
			expectedSkills: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			registry := &mockSkillRegistry{skills: tt.mockSkills}
			selector := NewSelector(Config{}, registry)

			skills := selector.matchSkills(context.Background(), tt.task, tt.role)
			assert.Equal(t, tt.expectedSkills, skills)
		})
	}
}

func TestNewSelector(t *testing.T) {
	t.Parallel()

	registry := &mockSkillRegistry{}
	cfg := Config{
		MaxBudget:  10000,
		StrictMode: true,
	}

	selector := NewSelector(cfg, registry)

	assert.NotNil(t, selector)
	assert.NotNil(t, selector.analyzer)
	assert.Equal(t, registry, selector.registry)
	assert.Equal(t, 10000, selector.maxBudget)
	assert.True(t, selector.strictMode)
}

func TestTask_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		task    Task
		wantErr bool
	}{
		{
			name: "valid task",
			task: Task{
				Description: "Add new feature",
				Type:        TaskTypeFeature,
			},
			wantErr: false,
		},
		{
			name: "missing description",
			task: Task{
				Type: TaskTypeFeature,
			},
			wantErr: true,
		},
		{
			name: "missing type",
			task: Task{
				Description: "Add new feature",
			},
			wantErr: true,
		},
		{
			name:    "empty task",
			task:    Task{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.task.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
