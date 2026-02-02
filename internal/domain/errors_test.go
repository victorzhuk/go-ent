package domain

//nolint:gosec // test file with necessary file operations

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		msg  string
	}{
		{"ErrAgentNotFound", ErrAgentNotFound, "agent not found"},
		{"ErrInvalidAgentConfig", ErrInvalidAgentConfig, "invalid agent config"},
		{"ErrInvalidAction", ErrInvalidAction, "invalid action"},
		{"ErrInvalidStrategy", ErrInvalidStrategy, "invalid strategy"},
		{"ErrSkillNotFound", ErrSkillNotFound, "skill not found"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.EqualError(t, tt.err, tt.msg)
		})
	}
}

func TestAgentError(t *testing.T) {
	baseErr := errors.New("config invalid")
	agentErr := &AgentError{
		Role: AgentRoleDeveloper,
		Err:  baseErr,
	}

	assert.Equal(t, "agent error [developer]: config invalid", agentErr.Error())
	assert.Equal(t, baseErr, agentErr.Unwrap())
}

func TestAgentError_Wrapping(t *testing.T) {
	baseErr := errors.New("base error")
	agentErr := &AgentError{
		Role: AgentRoleArchitect,
		Err:  baseErr,
	}

	assert.True(t, errors.Is(agentErr, baseErr))
}

func TestActionError(t *testing.T) {
	baseErr := errors.New("execution failed")
	actionErr := &ActionError{
		Action: SpecActionImplement,
		Err:    baseErr,
	}

	assert.Equal(t, "action error [implement]: execution failed", actionErr.Error())
	assert.Equal(t, baseErr, actionErr.Unwrap())
}

func TestActionError_Wrapping(t *testing.T) {
	baseErr := errors.New("task failed")
	actionErr := &ActionError{
		Action: SpecActionReview,
		Err:    baseErr,
	}

	assert.True(t, errors.Is(actionErr, baseErr))
}

func TestSkillError(t *testing.T) {
	baseErr := errors.New("not found")
	skillErr := &SkillError{
		Skill: "go-code",
		Err:   baseErr,
	}

	assert.Equal(t, "skill error [go-code]: not found", skillErr.Error())
	assert.Equal(t, baseErr, skillErr.Unwrap())
}

func TestSkillError_Wrapping(t *testing.T) {
	baseErr := errors.New("load failed")
	skillErr := &SkillError{
		Skill: "go-test",
		Err:   baseErr,
	}

	assert.True(t, errors.Is(skillErr, baseErr))
}

func TestIsAgentError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "is agent error",
			err: &AgentError{
				Role: AgentRoleDeveloper,
				Err:  errors.New("test"),
			},
			want: true,
		},
		{
			name: "wrapped agent error",
			err: fmt.Errorf("wrapped: %w", &AgentError{
				Role: AgentRoleArchitect,
				Err:  errors.New("test"),
			}),
			want: true,
		},
		{
			name: "not agent error",
			err:  errors.New("regular error"),
			want: false,
		},
		{
			name: "different domain error",
			err:  &ActionError{Action: SpecActionImplement, Err: errors.New("test")},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, IsAgentError(tt.err))
		})
	}
}

func TestIsActionError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "is action error",
			err: &ActionError{
				Action: SpecActionImplement,
				Err:    errors.New("test"),
			},
			want: true,
		},
		{
			name: "wrapped action error",
			err: fmt.Errorf("wrapped: %w", &ActionError{
				Action: SpecActionReview,
				Err:    errors.New("test"),
			}),
			want: true,
		},
		{
			name: "not action error",
			err:  errors.New("regular error"),
			want: false,
		},
		{
			name: "different domain error",
			err:  &SkillError{Skill: "test", Err: errors.New("test")},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, IsActionError(tt.err))
		})
	}
}

func TestIsSkillError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "is skill error",
			err: &SkillError{
				Skill: "go-code",
				Err:   errors.New("test"),
			},
			want: true,
		},
		{
			name: "wrapped skill error",
			err: fmt.Errorf("wrapped: %w", &SkillError{
				Skill: "go-test",
				Err:   errors.New("test"),
			}),
			want: true,
		},
		{
			name: "not skill error",
			err:  errors.New("regular error"),
			want: false,
		},
		{
			name: "different domain error",
			err:  &AgentError{Role: AgentRoleDeveloper, Err: errors.New("test")},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, IsSkillError(tt.err))
		})
	}
}
