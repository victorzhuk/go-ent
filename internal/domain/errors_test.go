package domain

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
		{"ErrRuntimeUnavailable", ErrRuntimeUnavailable, "runtime unavailable"},
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

func TestRuntimeError(t *testing.T) {
	baseErr := errors.New("not available")
	runtimeErr := &RuntimeError{
		Runtime: RuntimeCLI,
		Err:     baseErr,
	}

	assert.Equal(t, "runtime error [cli]: not available", runtimeErr.Error())
	assert.Equal(t, baseErr, runtimeErr.Unwrap())
}

func TestRuntimeError_Wrapping(t *testing.T) {
	baseErr := errors.New("connection failed")
	runtimeErr := &RuntimeError{
		Runtime: RuntimeClaudeCode,
		Err:     baseErr,
	}

	assert.True(t, errors.Is(runtimeErr, baseErr))
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
	baseErr := errors.New("validation failed")
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
	baseErr := errors.New("execution timeout")
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
			err: &RuntimeError{
				Runtime: RuntimeCLI,
				Err:     errors.New("test"),
			},
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

func TestIsRuntimeError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "is runtime error",
			err: &RuntimeError{
				Runtime: RuntimeCLI,
				Err:     errors.New("test"),
			},
			want: true,
		},
		{
			name: "wrapped runtime error",
			err: fmt.Errorf("wrapped: %w", &RuntimeError{
				Runtime: RuntimeClaudeCode,
				Err:     errors.New("test"),
			}),
			want: true,
		},
		{
			name: "not runtime error",
			err:  errors.New("regular error"),
			want: false,
		},
		{
			name: "different domain error",
			err: &AgentError{
				Role: AgentRoleDeveloper,
				Err:  errors.New("test"),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, IsRuntimeError(tt.err))
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
			err: &SkillError{
				Skill: "test-skill",
				Err:   errors.New("test"),
			},
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
			err: &RuntimeError{
				Runtime: RuntimeCLI,
				Err:     errors.New("test"),
			},
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

func TestErrorChaining(t *testing.T) {
	baseErr := errors.New("base error")
	agentErr := &AgentError{
		Role: AgentRoleDeveloper,
		Err:  baseErr,
	}
	wrappedErr := fmt.Errorf("operation failed: %w", agentErr)

	assert.True(t, errors.Is(wrappedErr, baseErr))
	assert.True(t, IsAgentError(wrappedErr))
}

func TestMultipleErrorWrapping(t *testing.T) {
	baseErr := errors.New("connection timeout")
	runtimeErr := &RuntimeError{
		Runtime: RuntimeCLI,
		Err:     baseErr,
	}
	layer1 := fmt.Errorf("layer 1: %w", runtimeErr)
	layer2 := fmt.Errorf("layer 2: %w", layer1)

	assert.True(t, errors.Is(layer2, baseErr))
	assert.True(t, IsRuntimeError(layer2))
}

func TestSentinelErrorWrapping(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		sentinel error
		checker  func(error) bool
	}{
		{
			name:     "wrapped ErrAgentNotFound",
			err:      fmt.Errorf("load config: %w", ErrAgentNotFound),
			sentinel: ErrAgentNotFound,
			checker:  nil,
		},
		{
			name: "AgentError wrapping sentinel",
			err: &AgentError{
				Role: AgentRoleDeveloper,
				Err:  ErrInvalidAgentConfig,
			},
			sentinel: ErrInvalidAgentConfig,
			checker:  IsAgentError,
		},
		{
			name: "RuntimeError wrapping sentinel",
			err: &RuntimeError{
				Runtime: RuntimeCLI,
				Err:     ErrRuntimeUnavailable,
			},
			sentinel: ErrRuntimeUnavailable,
			checker:  IsRuntimeError,
		},
		{
			name: "ActionError wrapping sentinel",
			err: &ActionError{
				Action: SpecActionImplement,
				Err:    ErrInvalidAction,
			},
			sentinel: ErrInvalidAction,
			checker:  IsActionError,
		},
		{
			name: "SkillError wrapping sentinel",
			err: &SkillError{
				Skill: "go-code",
				Err:   ErrSkillNotFound,
			},
			sentinel: ErrSkillNotFound,
			checker:  IsSkillError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.True(t, errors.Is(tt.err, tt.sentinel))
			if tt.checker != nil {
				assert.True(t, tt.checker(tt.err))
			}
		})
	}
}

func TestNestedSentinelWrapping(t *testing.T) {
	baseErr := ErrAgentNotFound
	agentErr := &AgentError{
		Role: AgentRoleArchitect,
		Err:  baseErr,
	}
	layer1 := fmt.Errorf("initialize system: %w", agentErr)
	layer2 := fmt.Errorf("startup failed: %w", layer1)

	assert.True(t, errors.Is(layer2, ErrAgentNotFound))
	assert.True(t, IsAgentError(layer2))

	var ae *AgentError
	assert.True(t, errors.As(layer2, &ae))
	assert.Equal(t, AgentRoleArchitect, ae.Role)
}

func TestCrossTypeWrapping(t *testing.T) {
	skillErr := &SkillError{
		Skill: "go-arch",
		Err:   ErrSkillNotFound,
	}
	actionErr := &ActionError{
		Action: SpecActionDesign,
		Err:    skillErr,
	}

	assert.True(t, errors.Is(actionErr, ErrSkillNotFound))
	assert.True(t, IsActionError(actionErr))
	assert.True(t, IsSkillError(actionErr))

	var ae *ActionError
	assert.True(t, errors.As(actionErr, &ae))
	assert.Equal(t, SpecActionDesign, ae.Action)

	var se *SkillError
	assert.True(t, errors.As(actionErr, &se))
	assert.Equal(t, "go-arch", se.Skill)
}

func TestErrorTypeChecks_NilHandling(t *testing.T) {
	t.Run("IsAgentError with nil", func(t *testing.T) {
		t.Parallel()
		assert.False(t, IsAgentError(nil))
	})

	t.Run("IsRuntimeError with nil", func(t *testing.T) {
		t.Parallel()
		assert.False(t, IsRuntimeError(nil))
	})

	t.Run("IsActionError with nil", func(t *testing.T) {
		t.Parallel()
		assert.False(t, IsActionError(nil))
	})

	t.Run("IsSkillError with nil", func(t *testing.T) {
		t.Parallel()
		assert.False(t, IsSkillError(nil))
	})
}

func TestErrorTypeChecks_DeepWrapping(t *testing.T) {
	agentErr := &AgentError{
		Role: AgentRoleDeveloper,
		Err:  ErrInvalidAgentConfig,
	}
	layer1 := fmt.Errorf("layer 1: %w", agentErr)
	layer2 := fmt.Errorf("layer 2: %w", layer1)
	layer3 := fmt.Errorf("layer 3: %w", layer2)
	layer4 := fmt.Errorf("layer 4: %w", layer3)

	assert.True(t, IsAgentError(layer4))
	assert.True(t, errors.Is(layer4, ErrInvalidAgentConfig))

	var ae *AgentError
	assert.True(t, errors.As(layer4, &ae))
	assert.Equal(t, AgentRoleDeveloper, ae.Role)
}

func TestErrorTypeChecks_CombinedScenarios(t *testing.T) {
	tests := []struct {
		name            string
		err             error
		isAgent         bool
		isRuntime       bool
		isAction        bool
		isSkill         bool
		expectedAgent   *AgentError
		expectedRuntime *RuntimeError
		expectedAction  *ActionError
		expectedSkill   *SkillError
	}{
		{
			name: "pure AgentError",
			err: &AgentError{
				Role: AgentRoleArchitect,
				Err:  errors.New("test"),
			},
			isAgent: true,
			expectedAgent: &AgentError{
				Role: AgentRoleArchitect,
			},
		},
		{
			name: "ActionError wrapping SkillError",
			err: &ActionError{
				Action: SpecActionImplement,
				Err: &SkillError{
					Skill: "go-code",
					Err:   ErrSkillNotFound,
				},
			},
			isAction: true,
			isSkill:  true,
			expectedAction: &ActionError{
				Action: SpecActionImplement,
			},
			expectedSkill: &SkillError{
				Skill: "go-code",
			},
		},
		{
			name: "RuntimeError wrapping AgentError",
			err: &RuntimeError{
				Runtime: RuntimeCLI,
				Err: &AgentError{
					Role: AgentRoleOps,
					Err:  ErrAgentNotFound,
				},
			},
			isRuntime: true,
			isAgent:   true,
			expectedRuntime: &RuntimeError{
				Runtime: RuntimeCLI,
			},
			expectedAgent: &AgentError{
				Role: AgentRoleOps,
			},
		},
		{
			name:    "plain error matches nothing",
			err:     errors.New("plain error"),
			isAgent: false, isRuntime: false, isAction: false, isSkill: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.isAgent, IsAgentError(tt.err), "IsAgentError mismatch")
			assert.Equal(t, tt.isRuntime, IsRuntimeError(tt.err), "IsRuntimeError mismatch")
			assert.Equal(t, tt.isAction, IsActionError(tt.err), "IsActionError mismatch")
			assert.Equal(t, tt.isSkill, IsSkillError(tt.err), "IsSkillError mismatch")

			if tt.expectedAgent != nil {
				var ae *AgentError
				assert.True(t, errors.As(tt.err, &ae))
				assert.Equal(t, tt.expectedAgent.Role, ae.Role)
			}

			if tt.expectedRuntime != nil {
				var re *RuntimeError
				assert.True(t, errors.As(tt.err, &re))
				assert.Equal(t, tt.expectedRuntime.Runtime, re.Runtime)
			}

			if tt.expectedAction != nil {
				var ae *ActionError
				assert.True(t, errors.As(tt.err, &ae))
				assert.Equal(t, tt.expectedAction.Action, ae.Action)
			}

			if tt.expectedSkill != nil {
				var se *SkillError
				assert.True(t, errors.As(tt.err, &se))
				assert.Equal(t, tt.expectedSkill.Skill, se.Skill)
			}
		})
	}
}
