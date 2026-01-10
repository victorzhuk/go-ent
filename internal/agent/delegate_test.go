package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/victorzhuk/go-ent/internal/domain"
)

func TestNewDelegator(t *testing.T) {
	t.Parallel()

	d := NewDelegator()

	assert.NotNil(t, d)
	assert.NotNil(t, d.rules)
	assert.Equal(t, 6, len(d.rules))
}

func TestDelegator_CanDelegate(t *testing.T) {
	t.Parallel()

	d := NewDelegator()

	tests := []struct {
		name string
		from domain.AgentRole
		to   domain.AgentRole
		want bool
	}{
		{
			name: "product can delegate to architect",
			from: domain.AgentRoleProduct,
			to:   domain.AgentRoleArchitect,
			want: true,
		},
		{
			name: "product can delegate to senior",
			from: domain.AgentRoleProduct,
			to:   domain.AgentRoleSenior,
			want: true,
		},
		{
			name: "product can delegate to developer",
			from: domain.AgentRoleProduct,
			to:   domain.AgentRoleDeveloper,
			want: true,
		},
		{
			name: "product cannot delegate to reviewer",
			from: domain.AgentRoleProduct,
			to:   domain.AgentRoleReviewer,
			want: false,
		},
		{
			name: "architect can delegate to senior",
			from: domain.AgentRoleArchitect,
			to:   domain.AgentRoleSenior,
			want: true,
		},
		{
			name: "architect can delegate to developer",
			from: domain.AgentRoleArchitect,
			to:   domain.AgentRoleDeveloper,
			want: true,
		},
		{
			name: "architect cannot delegate to reviewer",
			from: domain.AgentRoleArchitect,
			to:   domain.AgentRoleReviewer,
			want: false,
		},
		{
			name: "senior can delegate to developer",
			from: domain.AgentRoleSenior,
			to:   domain.AgentRoleDeveloper,
			want: true,
		},
		{
			name: "senior can delegate to reviewer",
			from: domain.AgentRoleSenior,
			to:   domain.AgentRoleReviewer,
			want: true,
		},
		{
			name: "senior cannot delegate to architect",
			from: domain.AgentRoleSenior,
			to:   domain.AgentRoleArchitect,
			want: false,
		},
		{
			name: "developer can delegate to reviewer",
			from: domain.AgentRoleDeveloper,
			to:   domain.AgentRoleReviewer,
			want: true,
		},
		{
			name: "developer cannot delegate to senior",
			from: domain.AgentRoleDeveloper,
			to:   domain.AgentRoleSenior,
			want: false,
		},
		{
			name: "reviewer cannot delegate to anyone",
			from: domain.AgentRoleReviewer,
			to:   domain.AgentRoleDeveloper,
			want: false,
		},
		{
			name: "ops can delegate to senior",
			from: domain.AgentRoleOps,
			to:   domain.AgentRoleSenior,
			want: true,
		},
		{
			name: "ops can delegate to developer",
			from: domain.AgentRoleOps,
			to:   domain.AgentRoleDeveloper,
			want: true,
		},
		{
			name: "ops cannot delegate to architect",
			from: domain.AgentRoleOps,
			to:   domain.AgentRoleArchitect,
			want: false,
		},
		{
			name: "invalid from role",
			from: domain.AgentRole("invalid"),
			to:   domain.AgentRoleDeveloper,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, d.CanDelegate(tt.from, tt.to))
		})
	}
}

func TestDelegator_GetDelegationChain(t *testing.T) {
	t.Parallel()

	d := NewDelegator()

	tests := []struct {
		name      string
		task      Task
		wantChain []domain.AgentRole
		wantErr   bool
	}{
		{
			name: "feature task chain",
			task: Task{
				Description: "add new feature",
				Type:        TaskTypeFeature,
			},
			wantChain: []domain.AgentRole{
				domain.AgentRoleArchitect,
				domain.AgentRoleSenior,
				domain.AgentRoleDeveloper,
				domain.AgentRoleReviewer,
			},
			wantErr: false,
		},
		{
			name: "bugfix task chain",
			task: Task{
				Description: "fix bug",
				Type:        TaskTypeBugFix,
			},
			wantChain: []domain.AgentRole{
				domain.AgentRoleSenior,
				domain.AgentRoleDeveloper,
				domain.AgentRoleReviewer,
			},
			wantErr: false,
		},
		{
			name: "refactor task chain",
			task: Task{
				Description: "refactor code",
				Type:        TaskTypeRefactor,
			},
			wantChain: []domain.AgentRole{
				domain.AgentRoleSenior,
				domain.AgentRoleDeveloper,
				domain.AgentRoleReviewer,
			},
			wantErr: false,
		},
		{
			name: "test task chain",
			task: Task{
				Description: "write tests",
				Type:        TaskTypeTest,
			},
			wantChain: []domain.AgentRole{
				domain.AgentRoleDeveloper,
				domain.AgentRoleReviewer,
			},
			wantErr: false,
		},
		{
			name: "documentation task chain",
			task: Task{
				Description: "write docs",
				Type:        TaskTypeDocumentation,
			},
			wantChain: []domain.AgentRole{
				domain.AgentRoleDeveloper,
			},
			wantErr: false,
		},
		{
			name: "architecture task chain",
			task: Task{
				Description: "design system",
				Type:        TaskTypeArchitecture,
			},
			wantChain: []domain.AgentRole{
				domain.AgentRoleArchitect,
				domain.AgentRoleSenior,
				domain.AgentRoleReviewer,
			},
			wantErr: false,
		},
		{
			name: "unknown task type defaults to developer",
			task: Task{
				Description: "unknown task",
				Type:        TaskType("unknown"),
			},
			wantChain: []domain.AgentRole{
				domain.AgentRoleDeveloper,
			},
			wantErr: false,
		},
		{
			name: "invalid task - empty description",
			task: Task{
				Description: "",
				Type:        TaskTypeFeature,
			},
			wantChain: nil,
			wantErr:   true,
		},
		{
			name: "invalid task - empty type",
			task: Task{
				Description: "task",
				Type:        "",
			},
			wantChain: nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			chain, err := d.GetDelegationChain(tt.task)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, chain)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantChain, chain)
			}
		})
	}
}

func TestDelegator_GetNextAgent(t *testing.T) {
	t.Parallel()

	d := NewDelegator()

	tests := []struct {
		name        string
		task        Task
		current     domain.AgentRole
		wantNext    domain.AgentRole
		wantErr     bool
		errContains string
	}{
		{
			name: "feature: architect to senior",
			task: Task{
				Description: "add feature",
				Type:        TaskTypeFeature,
			},
			current:  domain.AgentRoleArchitect,
			wantNext: domain.AgentRoleSenior,
			wantErr:  false,
		},
		{
			name: "feature: senior to developer",
			task: Task{
				Description: "add feature",
				Type:        TaskTypeFeature,
			},
			current:  domain.AgentRoleSenior,
			wantNext: domain.AgentRoleDeveloper,
			wantErr:  false,
		},
		{
			name: "feature: developer to reviewer",
			task: Task{
				Description: "add feature",
				Type:        TaskTypeFeature,
			},
			current:  domain.AgentRoleDeveloper,
			wantNext: domain.AgentRoleReviewer,
			wantErr:  false,
		},
		{
			name: "feature: reviewer is last",
			task: Task{
				Description: "add feature",
				Type:        TaskTypeFeature,
			},
			current:     domain.AgentRoleReviewer,
			wantNext:    "",
			wantErr:     true,
			errContains: "no next agent",
		},
		{
			name: "bugfix: senior to developer",
			task: Task{
				Description: "fix bug",
				Type:        TaskTypeBugFix,
			},
			current:  domain.AgentRoleSenior,
			wantNext: domain.AgentRoleDeveloper,
			wantErr:  false,
		},
		{
			name: "documentation: developer is only",
			task: Task{
				Description: "write docs",
				Type:        TaskTypeDocumentation,
			},
			current:     domain.AgentRoleDeveloper,
			wantNext:    "",
			wantErr:     true,
			errContains: "no next agent",
		},
		{
			name: "current role not in chain",
			task: Task{
				Description: "write docs",
				Type:        TaskTypeDocumentation,
			},
			current:     domain.AgentRoleArchitect,
			wantNext:    "",
			wantErr:     true,
			errContains: "no next agent",
		},
		{
			name: "invalid task",
			task: Task{
				Description: "",
				Type:        TaskTypeFeature,
			},
			current:     domain.AgentRoleArchitect,
			wantNext:    "",
			wantErr:     true,
			errContains: "invalid task",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			next, err := d.GetNextAgent(tt.task, tt.current)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Empty(t, next)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantNext, next)
			}
		})
	}
}

func TestDelegator_ShouldDelegate(t *testing.T) {
	t.Parallel()

	d := NewDelegator()

	tests := []struct {
		name    string
		task    Task
		current domain.AgentRole
		want    bool
	}{
		{
			name: "architect should delegate for feature",
			task: Task{
				Description: "add feature",
				Type:        TaskTypeFeature,
			},
			current: domain.AgentRoleArchitect,
			want:    true,
		},
		{
			name: "developer should delegate for feature",
			task: Task{
				Description: "add feature",
				Type:        TaskTypeFeature,
			},
			current: domain.AgentRoleDeveloper,
			want:    true,
		},
		{
			name: "product should not delegate for feature (not in chain)",
			task: Task{
				Description: "add feature",
				Type:        TaskTypeFeature,
			},
			current: domain.AgentRoleProduct,
			want:    false,
		},
		{
			name: "ops should not delegate for bugfix (not in chain)",
			task: Task{
				Description: "fix bug",
				Type:        TaskTypeBugFix,
			},
			current: domain.AgentRoleOps,
			want:    false,
		},
		{
			name: "developer should delegate for test",
			task: Task{
				Description: "write tests",
				Type:        TaskTypeTest,
			},
			current: domain.AgentRoleDeveloper,
			want:    true,
		},
		{
			name: "invalid task returns false",
			task: Task{
				Description: "",
				Type:        TaskTypeFeature,
			},
			current: domain.AgentRoleArchitect,
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, d.ShouldDelegate(tt.task, tt.current))
		})
	}
}

func TestDelegator_WorkflowIntegration(t *testing.T) {
	t.Parallel()

	d := NewDelegator()

	t.Run("complete feature workflow", func(t *testing.T) {
		t.Parallel()

		task := Task{
			Description: "implement user authentication",
			Type:        TaskTypeFeature,
		}

		chain, err := d.GetDelegationChain(task)
		require.NoError(t, err)
		assert.Equal(t, 4, len(chain))

		current := chain[0]
		assert.Equal(t, domain.AgentRoleArchitect, current)
		assert.True(t, d.ShouldDelegate(task, current))

		next, err := d.GetNextAgent(task, current)
		require.NoError(t, err)
		assert.Equal(t, domain.AgentRoleSenior, next)
		assert.True(t, d.CanDelegate(current, next))

		current = next
		next, err = d.GetNextAgent(task, current)
		require.NoError(t, err)
		assert.Equal(t, domain.AgentRoleDeveloper, next)
		assert.True(t, d.CanDelegate(current, next))

		current = next
		next, err = d.GetNextAgent(task, current)
		require.NoError(t, err)
		assert.Equal(t, domain.AgentRoleReviewer, next)
		assert.True(t, d.CanDelegate(current, next))

		current = next
		_, err = d.GetNextAgent(task, current)
		assert.Error(t, err)
	})

	t.Run("bugfix workflow", func(t *testing.T) {
		t.Parallel()

		task := Task{
			Description: "fix login error",
			Type:        TaskTypeBugFix,
		}

		chain, err := d.GetDelegationChain(task)
		require.NoError(t, err)
		assert.Equal(t, 3, len(chain))
		assert.Equal(t, domain.AgentRoleSenior, chain[0])
		assert.Equal(t, domain.AgentRoleDeveloper, chain[1])
		assert.Equal(t, domain.AgentRoleReviewer, chain[2])
	})

	t.Run("simple documentation workflow", func(t *testing.T) {
		t.Parallel()

		task := Task{
			Description: "update readme",
			Type:        TaskTypeDocumentation,
		}

		chain, err := d.GetDelegationChain(task)
		require.NoError(t, err)
		assert.Equal(t, 1, len(chain))
		assert.Equal(t, domain.AgentRoleDeveloper, chain[0])

		_, err = d.GetNextAgent(task, domain.AgentRoleDeveloper)
		assert.Error(t, err)
	})
}
